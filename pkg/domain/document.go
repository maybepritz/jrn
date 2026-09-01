package domain

import (
	"slices"
	"strings"
	"time"
)

// DocumentTaskBuilder позволяет собирать задачу и сразу добавлять её в документ.
type DocumentTaskBuilder struct {
	doc     *Document
	builder *TaskBuilder
}

// NewTask начинает построение новой задачи для документа.
func (d *Document) NewTask(title string) *DocumentTaskBuilder {
	return &DocumentTaskBuilder{
		doc:     d,
		builder: NewTask(title),
	}
}

func (b *DocumentTaskBuilder) Tag(tags ...string) *DocumentTaskBuilder {
	b.builder.Tag(tags...)
	return b
}

func (b *DocumentTaskBuilder) Attr(key, value string) *DocumentTaskBuilder {
	b.builder.Attr(key, value)
	return b
}

func (b *DocumentTaskBuilder) Note(notes ...string) *DocumentTaskBuilder {
	b.builder.Note(notes...)
	return b
}

func (b *DocumentTaskBuilder) Done(done bool) *DocumentTaskBuilder {
	b.builder.Done(done)
	return b
}

// Add финализирует создание задачи и добавляет её в документ.
func (b *DocumentTaskBuilder) Add() error {
	task, err := b.builder.Build()
	if err != nil {
		return err
	}
	return b.doc.AddTask(task)
}

// AddTask валидирует и добавляет задачу в конец списка задач документа.
func (d *Document) AddTask(task Task) error {
	trimmed := strings.TrimSpace(task.Title)
	if trimmed == "" {
		return ErrEmptyTitle
	}
	task.Title = trimmed

	if task.Tags == nil {
		task.Tags = make([]string, 0)
	}
	if task.Attributes == nil {
		task.Attributes = make([]Attribute, 0)
	}
	if task.Notes == nil {
		task.Notes = make([]string, 0)
	}

	d.Tasks = append(d.Tasks, task)

	if task.Done {
		d.Meta.TaskCompleted++
	}

	return nil
}

// ToggleTask переключает статус выполнения задачи (Done <-> !Done) и обновляет счетчик.
func (d *Document) ToggleTask(index int) error {
	if index < 0 || index >= len(d.Tasks) {
		return ErrTaskNotFound
	}

	d.Tasks[index].Done = !d.Tasks[index].Done

	if d.Tasks[index].Done {
		d.Meta.TaskCompleted++
	} else if d.Meta.TaskCompleted > 0 {
		d.Meta.TaskCompleted--
	}

	return nil
}

// DeleteTask удаляет задачу по индексу с сохранением ссылочной целостности и пересчетом статистики.
func (d *Document) DeleteTask(index int) error {
	if index < 0 || index >= len(d.Tasks) {
		return ErrTaskNotFound
	}

	if d.Tasks[index].Done && d.Meta.TaskCompleted > 0 {
		d.Meta.TaskCompleted--
	}

	d.Tasks = slices.Delete(d.Tasks, index, index+1)
	return nil
}

// UpdateTaskTitle обновляет заголовок задачи с проверкой на пустоту.
func (d *Document) UpdateTaskTitle(index int, title string) error {
	if index < 0 || index >= len(d.Tasks) {
		return ErrTaskNotFound
	}

	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return ErrEmptyTitle
	}

	d.Tasks[index].Title = trimmed
	return nil
}

// ReorderTasks перемещает задачу с позиции fromIndex на позицию toIndex.
func (d *Document) ReorderTasks(fromIndex, toIndex int) error {
	if fromIndex < 0 || fromIndex >= len(d.Tasks) ||
		toIndex < 0 || toIndex >= len(d.Tasks) {
		return ErrTaskNotFound
	}

	if fromIndex == toIndex {
		return nil
	}

	target := d.Tasks[fromIndex]

	if fromIndex < toIndex {
		copy(d.Tasks[fromIndex:toIndex], d.Tasks[fromIndex+1:toIndex+1])
	} else {
		copy(d.Tasks[toIndex+1:fromIndex+1], d.Tasks[toIndex:fromIndex])
	}

	d.Tasks[toIndex] = target
	return nil
}

// AddNoteToTask добавляет заметку к конкретной задаче.
func (d *Document) AddNoteToTask(taskIndex int, note string) error {
	if taskIndex < 0 || taskIndex >= len(d.Tasks) {
		return ErrTaskNotFound
	}

	trimmed := strings.TrimSpace(note)
	if trimmed == "" {
		return nil
	}

	d.Tasks[taskIndex].Notes = append(d.Tasks[taskIndex].Notes, trimmed)
	return nil
}

// DeleteNoteFromTask удаляет заметку из задачи по индексу.
func (d *Document) DeleteNoteFromTask(taskIndex, noteIndex int) error {
	if taskIndex < 0 || taskIndex >= len(d.Tasks) {
		return ErrTaskNotFound
	}

	notes := d.Tasks[taskIndex].Notes
	if noteIndex < 0 || noteIndex >= len(notes) {
		return ErrNoteNotFound
	}

	d.Tasks[taskIndex].Notes = slices.Delete(notes, noteIndex, noteIndex+1)
	return nil
}

// AddDailyLog добавляет строку в секцию дневника дня.
func (d *Document) AddDailyLog(text string) {
	trimmed := strings.TrimSpace(text)
	if trimmed != "" {
		d.DailyLog = append(d.DailyLog, trimmed)
	}
}

// DeleteDailyLog удаляет строку дневника по индексу.
func (d *Document) DeleteDailyLog(index int) error {
	if index < 0 || index >= len(d.DailyLog) {
		return ErrNotFound
	}

	d.DailyLog = slices.Delete(d.DailyLog, index, index+1)
	return nil
}

// RecalculateStats пересчитывает счетчик выполненных задач на основе текущего состояния.
func (d *Document) RecalculateStats() {
	count := 0
	for _, t := range d.Tasks {
		if t.Done {
			count++
		}
	}
	d.Meta.TaskCompleted = count
}

// SearchTasks ищет задачи, содержащие указанный текст (в названии, заметках, тегах или атрибутах).
func (d *Document) SearchTasks(query string) []Task {
	var result []Task
	for _, task := range d.Tasks {
		if task.Matches(query) {
			result = append(result, task)
		}
	}
	return result
}

// TasksByTag возвращает список задач, содержащих указанный тег.
func (d *Document) TasksByTag(tag string) []Task {
	var result []Task
	for _, task := range d.Tasks {
		if task.HasTag(tag) {
			result = append(result, task)
		}
	}
	return result
}

// TasksByAttr возвращает задачи, содержащие указанный атрибут (и опционально совпадение по значению).
func (d *Document) TasksByAttr(key, value string) []Task {
	var result []Task
	targetVal := strings.ToLower(strings.TrimSpace(value))

	for _, task := range d.Tasks {
		if val, found := task.AttrVal(key); found {
			if targetVal == "" || strings.ToLower(val) == targetVal {
				result = append(result, task)
			}
		}
	}
	return result
}

// OverdueTasks возвращает незавершенные задачи, срок выполнения которых (атрибут due) истек.
func (d *Document) OverdueTasks(today time.Time) []Task {
	todayStr := today.Format("2006-01-02")
	var result []Task

	for _, task := range d.Tasks {
		if task.Done {
			continue
		}
		if due, found := task.AttrVal("due"); found {
			if due < todayStr {
				result = append(result, task)
			}
		}
	}
	return result
}
