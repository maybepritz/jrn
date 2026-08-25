package core

import "time"

type Priority int

const (
	PriorityNone Priority = iota
	PriorityLow
	PriorityMedium
	PriorityHigh
)

type Directive struct {
	Key   string // "due", "done", "prio"
	Value string // "2026-08-25", "high"
}

type Task struct {
	ID       int
	Title    string
	Done     bool
	Tags     []string
	Priority Priority
	// Directives  []Directive
	DueTime     time.Time
	CompletedAt time.Time
	Notes       []string

	FilePath string
	// LineNumber int
	// RawLine    string
}

type DayDocument struct {
	FilePath string
	Meta     Metadata
	DailyLog []string
	Tasks    []Task
	// RawLines []string
}

type Metadata struct {
	Date           string `yaml:"date"`            // "2026-08-25"
	Streak         int    `yaml:"streak"`          // 14
	TotalCompleted int    `yaml:"total_completed"` // 3
}
