package domain

type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Task struct {
	Done       bool        `json:"done"`
	Title      string      `json:"title"`
	Tags       []string    `json:"tags"`
	Attributes []Attribute `json:"attributes"`
	Notes      []string    `json:"notes"`
}

type Document struct {
	Meta     Metadata `json:"meta"`
	DailyLog []string `json:"daily_log"`
	Tasks    []Task   `json:"tasks"`
}

type Metadata struct {
	Date          string `yaml:"date" json:"date"`
	Streak        int    `yaml:"streak" json:"streak"`
	TaskCompleted int    `yaml:"task_completed" json:"task_completed"`
}
