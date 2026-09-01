package domain

import "strings"

type TaskBuilder struct {
	task Task
	err  error
}

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

func (b *TaskBuilder) Tag(tags ...string) *TaskBuilder {
	b.task.Tags = append(b.task.Tags, tags...)
	return b
}

func (b *TaskBuilder) Attr(key, value string) *TaskBuilder {
	b.task.Attributes = append(b.task.Attributes, Attribute{Key: key, Value: value})
	return b
}

func (b *TaskBuilder) Note(notes ...string) *TaskBuilder {
	b.task.Notes = append(b.task.Notes, notes...)
	return b
}

func (b *TaskBuilder) Build() (Task, error) {
	if b.err != nil {
		return Task{}, b.err
	}
	return b.task, nil
}
