package httpx

// CodeCatalog is the single source used by docs/API.md and integration tests.
var CodeCatalog = []struct {
	Code    string
	Status  int
	Meaning string
}{
	{"E_VALIDATION", 400, "请求参数校验失败"},
	{"E_UNAUTHORIZED", 401, "未登录或令牌无效"},
	{"E_FORBIDDEN", 403, "无权访问该资源"},
	{"E_NOT_FOUND", 404, "资源不存在"},
	{"E_CONFLICT", 409, "资源冲突"},
	{"E_INVALID_TRANSITION", 409, "非法的番茄钟状态迁移"},
	{"E_SESSION_BUSY", 409, "已有进行中的番茄钟会话"},
	{"E_OPTIMISTIC_LOCK", 409, "会话版本冲突"},
	{"E_INTERNAL", 500, "服务器内部错误"},
}

func LookupCode(code string) int {
	for _, c := range CodeCatalog {
		if c.Code == code {
			return c.Status
		}
	}
	return 500
}
