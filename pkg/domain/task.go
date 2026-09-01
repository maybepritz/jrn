package domain

import (
	"strings"
)

type TaskBuilder struct {
	task Task
	err  error
}

// NewTask создает новый билдер для задачи с проверкой обязательного заголовка.
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

// Tag добавляет один или несколько тегов к задаче.
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

// Attr добавляет атрибут ключ-значение (например @prio(high), @due(2026-09-01)).
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

// Note добавляет одну или несколько заметок к задаче.
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

// Done устанавливает статус выполнения задачи.
func (b *TaskBuilder) Done(done bool) *TaskBuilder {
	b.task.Done = done
	return b
}

// Build валидирует и возвращает готовую задачу.
func (b *TaskBuilder) Build() (Task, error) {
	if b.err != nil {
		return Task{}, b.err
	}
	return b.task, nil
}

// Matches проверяет, содержится ли поисковый запрос в заголовке, заметках, тегах или атрибутах задачи.
func (t *Task) Matches(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return false
	}

	// 1. Поиск в заголовке
	if strings.Contains(strings.ToLower(t.Title), q) {
		return true
	}

	// 2. Поиск в заметках
	for _, note := range t.Notes {
		if strings.Contains(strings.ToLower(note), q) {
			return true
		}
	}

	// 3. Поиск в тегах
	for _, tag := range t.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}

	// 4. Поиск в атрибутах
	for _, attr := range t.Attributes {
		if strings.Contains(strings.ToLower(attr.Key), q) ||
			strings.Contains(strings.ToLower(attr.Value), q) {
			return true
		}
	}

	return false
}

// HasTag проверяет наличие указанного тега у задачи (регистронезависимо).
func (t *Task) HasTag(tag string) bool {
	target := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(tag)), "#")
	for _, cur := range t.Tags {
		if strings.ToLower(cur) == target {
			return true
		}
	}
	return false
}

// AttrVal возвращает значение атрибута по ключу и флаг успешности поиска.
func (t *Task) AttrVal(key string) (string, bool) {
	target := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(key)), "@")
	for _, attr := range t.Attributes {
		if strings.ToLower(attr.Key) == target {
			return attr.Value, true
		}
	}
	return "", false
}
