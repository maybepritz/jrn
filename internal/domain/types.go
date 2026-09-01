package domain

type Attribute struct {
	Key   string
	Value string
}

type Task struct {
	Done       bool
	Title      string
	Tags       []string
	Attributes []Attribute
	Notes      []string
}

type Document struct {
	Meta     Metadata
	DailyLog []string
	Tasks    []Task
}

type Metadata struct {
	Date          string `yaml:"date"`           // "2026-08-25"
	Streak        int    `yaml:"streak"`         // 14
	TaskCompleted int    `yaml:"task_completed"` // 3
}
