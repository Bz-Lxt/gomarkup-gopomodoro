"""API smoke — mock/offline, expected cost ¥0."""
import json
import os
import urllib.error
import urllib.request

BASE = os.environ.get("API_BASE", "http://backend:8080/api/v1")
EMAIL = "geek@gopomodoro.dev"
PASSWORD = "pomodoro123"


def call(method, path, body=None, token=None, expect=200):
    data = None if body is None else json.dumps(body).encode()
    req = urllib.request.Request(BASE + path, data=data, method=method)
    req.add_header("Content-Type", "application/json")
    if token:
        req.add_header("Authorization", "Bearer " + token)
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            raw = resp.read().decode()
            code = resp.status
    except urllib.error.HTTPError as e:
        raw = e.read().decode()
        code = e.code
    payload = json.loads(raw) if raw else {}
    if code != expect:
        raise SystemExit(f"FAIL {method} {path} -> {code} {payload} want {expect}")
    return payload


def main():
    h = call("GET", "/health")
    assert h.get("ok") is True, h
    print("[PASS] Health Check")

    login = call("POST", "/auth/login", {"email": EMAIL, "password": PASSWORD})
    token = login["data"]["auth"]["token"]
    print("[PASS] Auth")

    projects = call("GET", "/projects", token=token)["data"]
    assert projects, "seed project missing"
    pid = projects[0]["id"]
    ms = call("GET", f"/projects/{pid}/milestones", token=token)["data"]
    assert len(ms) >= 2
    tasks = call("GET", f"/projects/{pid}/tasks", token=token)["data"]
    assert len(tasks) >= 12
    print("[PASS] Seed CRUD")

    todo = next(t for t in tasks if t["kanban_column"] != "done")
    started = call("POST", "/pomodoros", {"task_id": todo["id"]}, token=token, expect=201)
    sid = started["data"]["session"]["id"]
    assert started["data"]["session"]["state"] == "running"
    print("[PASS] Pomodoro start")

    busy = call("POST", "/pomodoros", {"task_id": todo["id"]}, token=token, expect=409)
    assert busy["error"]["code"] == "E_SESSION_BUSY"
    print("[PASS] Session busy")

    paused = call("POST", f"/pomodoros/{sid}/pause", {}, token=token)
    assert paused["data"]["session"]["state"] == "paused"

    # idle→completed is illegal: abort first then try tick on aborted
    aborted = call("POST", f"/pomodoros/{sid}/abort", {"reason": "user"}, token=token)
    assert aborted["data"]["session"]["state"] == "aborted"
    illegal = call("POST", f"/pomodoros/{sid}/resume", {}, token=token, expect=409)
    assert illegal["error"]["code"] == "E_INVALID_TRANSITION"
    print("[PASS] Invalid transition")

    bd = call("GET", f"/milestones/{ms[0]['id']}/burndown", token=token)
    assert "ideal" in bd["data"] and "actual" in bd["data"]
    print("[PASS] Burndown chart")
    print("[PASS] Mock API Response")


if __name__ == "__main__":
    main()
