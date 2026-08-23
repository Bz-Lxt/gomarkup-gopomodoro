package validate

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"gopomodoro/internal/httpx"
	"gopomodoro/internal/model"
)

func Fail(field, msg string) error {
	return httpx.ErrValidation.WithDetails(map[string]any{"field": field, "reason": msg})
}

func RequiredString(field, v string, min, max int) (string, error) {
	v, problem := requiredStringProblem(field, v, min, max)
	return v, problem
}

func requiredStringProblem(field, v string, min, max int) (string, *httpx.AppError) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", httpx.ErrValidation.WithDetails(map[string]any{"field": field, "reason": "必填"})
	}
	n := utf8.RuneCountInString(v)
	if n < min || n > max {
		return "", httpx.ErrValidation.WithDetails(map[string]any{"field": field, "reason": fmt.Sprintf("长度须在 %d–%d 之间", min, max)})
	}
	return v, nil
}

func Email(v string) (string, error) {
	v = strings.ToLower(strings.TrimSpace(v))
	if _, err := mail.ParseAddress(v); err != nil {
		return "", Fail("email", "邮箱格式无效")
	}
	return v, nil
}

func Password(v string) (string, error) {
	if utf8.RuneCountInString(v) < 8 || utf8.RuneCountInString(v) > 72 {
		return "", Fail("password", "密码长度须在 8–72 之间")
	}
	return v, nil
}

func PositiveInt(field string, v int, max int) (int, error) {
	if v < 1 {
		return 0, Fail(field, "必须为正整数")
	}
	if max > 0 && v > max {
		return 0, Fail(field, fmt.Sprintf("不能超过 %d", max))
	}
	return v, nil
}

func NonNegInt(field string, v int, max int) (int, error) {
	if v < 0 {
		return 0, Fail(field, "不能为负数")
	}
	if max > 0 && v > max {
		return 0, Fail(field, fmt.Sprintf("不能超过 %d", max))
	}
	return v, nil
}

func Column(v string) (model.KanbanColumn, error) {
	col := model.KanbanColumn(strings.TrimSpace(v))
	if !col.Valid() {
		return "", Fail("kanban_column", "必须是 backlog/todo/in_progress/done")
	}
	return col, nil
}

func Granularity(v string) (string, error) {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return "day", nil
	}
	if v != "day" && v != "week" {
		return "", Fail("granularity", "必须是 day 或 week")
	}
	return v, nil
}
