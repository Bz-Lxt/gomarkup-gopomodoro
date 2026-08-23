package validate

type FieldRule struct {
	Field string
	Min   int
	Max   int
	Kind  string
}

var Rules = []FieldRule{
	{"email", 5, 120, "email"},
	{"password", 8, 72, "password"},
	{"display_name", 1, 40, "string"},
	{"project.name", 1, 80, "string"},
	{"project.description", 0, 400, "string"},
	{"milestone.title", 1, 120, "string"},
	{"milestone.baseline_points", 0, 10000, "int"},
	{"task.title", 1, 160, "string"},
	{"task.estimated_pomodoros", 1, 99, "int"},
	{"kanban_column", 0, 0, "enum"},
	{"granularity", 0, 0, "enum"},
}

func Rule(field string) (FieldRule, bool) {
	for _, r := range Rules {
		if r.Field == field {
			return r, true
		}
	}
	return FieldRule{}, false
}
