package domain

import (
	"strings"
)

// TaskBuilder provides a fluent interface for constructing Task instances safely.
type TaskBuilder struct {
	task Task
	err  error
}

// NewTask initializes a new TaskBuilder and validates that the title is not empty.
func NewTask(title string) *TaskBuilder {
	trimmed := strings.TrimSpace(title)
	var err error
	if trimmed == "" {
		err = ErrEmptyTitle
	}
	return &TaskBuilder{
		task: Task{
			Title:      trimmed,
			Tags:       make([]string, 0),
			Attributes: make([]Attribute, 0),
			Notes:      make([]string, 0),
		},
		err: err,
	}
}

// Tag adds one or more tags to the task, trimming leading # and whitespace.
func (b *TaskBuilder) Tag(tags ...string) *TaskBuilder {
	if b.err != nil {
		return b
	}
	for _, tag := range tags {
		trimmed := strings.TrimSpace(strings.TrimPrefix(tag, "#"))
		if trimmed != "" {
			b.task.Tags = append(b.task.Tags, trimmed)
		}
	}
	return b
}

// Attr adds a key-value attribute (e.g., @prio(high), @due(2026-09-01)).
func (b *TaskBuilder) Attr(key, value string) *TaskBuilder {
	if b.err != nil {
		return b
	}
	k := strings.TrimSpace(strings.TrimPrefix(key, "@"))
	v := strings.TrimSpace(value)
	if k != "" {
		b.task.Attributes = append(b.task.Attributes, Attribute{Key: k, Value: v})
	}
	return b
}

// Note appends one or more note lines to the task.
func (b *TaskBuilder) Note(notes ...string) *TaskBuilder {
	if b.err != nil {
		return b
	}
	for _, note := range notes {
		trimmed := strings.TrimSpace(note)
		if trimmed != "" {
			b.task.Notes = append(b.task.Notes, trimmed)
		}
	}
	return b
}

// Done sets the completion state of the task.
func (b *TaskBuilder) Done(done bool) *TaskBuilder {
	b.task.Done = done
	return b
}

// Build finalizes task creation and returns the constructed Task or an error.
func (b *TaskBuilder) Build() (Task, error) {
	if b.err != nil {
		return Task{}, b.err
	}
	return b.task, nil
}

// Matches checks if the task matches the search query across title, notes, tags, or attributes.
func (t *Task) Matches(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return false
	}

	if strings.Contains(strings.ToLower(t.Title), q) {
		return true
	}

	for _, note := range t.Notes {
		if strings.Contains(strings.ToLower(note), q) {
			return true
		}
	}

	for _, tag := range t.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}

	for _, attr := range t.Attributes {
		if strings.Contains(strings.ToLower(attr.Key), q) ||
			strings.Contains(strings.ToLower(attr.Value), q) {
			return true
		}
	}

	return false
}

// HasTag checks if the task has the specified tag (case-insensitive, ignoring optional # prefix).
func (t *Task) HasTag(tag string) bool {
	target := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(tag)), "#")
	for _, cur := range t.Tags {
		if strings.ToLower(cur) == target {
			return true
		}
	}
	return false
}

// AttrVal returns the value of the attribute for the given key and true if found.
func (t *Task) AttrVal(key string) (string, bool) {
	target := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(key)), "@")
	for _, attr := range t.Attributes {
		if strings.ToLower(attr.Key) == target {
			return attr.Value, true
		}
	}
	return "", false
}

// Clone creates a deep copy of the task and its slices to ensure memory safety during rollover.
func (t *Task) Clone() Task {
	return Task{
		Done:       t.Done,
		Title:      t.Title,
		Tags:       append([]string(nil), t.Tags...),
		Attributes: append([]Attribute(nil), t.Attributes...),
		Notes:      append([]string(nil), t.Notes...),
	}
}
