CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL,
    timezone TEXT NOT NULL DEFAULT 'Asia/Shanghai',
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_projects_user ON projects(user_id);

CREATE TABLE IF NOT EXISTS milestones (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    start_date DATE NOT NULL,
    due_date DATE NOT NULL,
    baseline_points INTEGER NOT NULL CHECK (baseline_points >= 0),
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CHECK (due_date >= start_date)
);
CREATE INDEX IF NOT EXISTS idx_milestones_project ON milestones(project_id);

CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    milestone_id UUID REFERENCES milestones(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    estimated_pomodoros INTEGER NOT NULL CHECK (estimated_pomodoros >= 1),
    consumed_pomodoros INTEGER NOT NULL DEFAULT 0 CHECK (consumed_pomodoros >= 0),
    kanban_column TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_milestone ON tasks(milestone_id);

CREATE TABLE IF NOT EXISTS pomodoro_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    state TEXT NOT NULL,
    focus_duration_ms BIGINT NOT NULL,
    started_at TIMESTAMPTZ,
    paused_at TIMESTAMPTZ,
    paused_accumulated_ms BIGINT NOT NULL DEFAULT 0,
    expected_end_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    abort_reason TEXT NOT NULL DEFAULT '',
    resume_token TEXT NOT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_pomo_user_state ON pomodoro_sessions(user_id, state);
CREATE INDEX IF NOT EXISTS idx_pomo_resume ON pomodoro_sessions(resume_token);

CREATE TABLE IF NOT EXISTS burndown_points (
    id UUID PRIMARY KEY,
    milestone_id UUID NOT NULL REFERENCES milestones(id) ON DELETE CASCADE,
    recorded_at TIMESTAMPTZ NOT NULL,
    remaining_points INTEGER NOT NULL,
    ideal_points DOUBLE PRECISION NOT NULL,
    event_type TEXT NOT NULL,
    event_id TEXT NOT NULL,
    UNIQUE (event_id)
);
CREATE INDEX IF NOT EXISTS idx_bd_milestone_time ON burndown_points(milestone_id, recorded_at);

CREATE TABLE IF NOT EXISTS scope_change_logs (
    id UUID PRIMARY KEY,
    milestone_id UUID NOT NULL REFERENCES milestones(id) ON DELETE CASCADE,
    delta_points INTEGER NOT NULL,
    reason TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS processed_events (
    event_id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL
);
