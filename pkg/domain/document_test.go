package domain

import (
	"testing"
	"time"
)

func TestTaskBuilder_Success(t *testing.T) {
	task, err := NewTask("  Изучить Go  ").
		Tag("#dev", "golang").
		Attr("@prio", "high").
		Attr("due", "2026-09-01").
		Note("Почитать про slice headers", "Запустить бенчмарки").
		Done(true).
		Build()

	if err != nil {
		t.Fatalf("неожиданная ошибка билдера: %v", err)
	}

	if task.Title != "Изучить Go" {
		t.Errorf("Title = %q, хотели %q", task.Title, "Изучить Go")
	}
	if len(task.Tags) != 2 || task.Tags[0] != "dev" || task.Tags[1] != "golang" {
		t.Errorf("Tags = %v, хотели [dev, golang]", task.Tags)
	}
	if len(task.Attributes) != 2 || task.Attributes[0].Key != "prio" || task.Attributes[0].Value != "high" {
		t.Errorf("Attributes = %v", task.Attributes)
	}
	if len(task.Notes) != 2 || task.Notes[0] != "Почитать про slice headers" {
		t.Errorf("Notes = %v", task.Notes)
	}
	if !task.Done {
		t.Errorf("Done = false, хотели true")
	}
}

func TestTaskBuilder_EmptyTitle(t *testing.T) {
	_, err := NewTask("   \t  ").
		Tag("test").
		Build()

	if err != ErrEmptyTitle {
		t.Errorf("ожидалась ошибка ErrEmptyTitle, получено: %v", err)
	}
}

func TestTask_Matches_And_HasTag(t *testing.T) {
	task, _ := NewTask("Починить баг в парсере").
		Tag("dev", "backend").
		Attr("prio", "critical").
		Note("Проверить FSM переходы").
		Build()

	if !task.Matches("баг") {
		t.Error("должен найти по слову в Title")
	}
	if !task.Matches("FSM") {
		t.Error("должен найти по слову в Notes")
	}
	if !task.Matches("BACKEND") {
		t.Error("должен найти по тегу (регистронезависимо)")
	}
	if !task.Matches("critical") {
		t.Error("должен найти по значению атрибута")
	}
	if task.Matches("несуществующее") {
		t.Error("не должен найти")
	}

	if !task.HasTag("#dev") || !task.HasTag("DEV") {
		t.Error("HasTag должен находить с # и без с любым регистром")
	}
	if task.HasTag("frontend") {
		t.Error("HasTag не должен находить отсутствующий тег")
	}

	val, found := task.AttrVal("@prio")
	if !found || val != "critical" {
		t.Errorf("AttrVal(@prio) = (%q, %v), хотели ('critical', true)", val, found)
	}
}

func TestDocument_AddTask_And_NewTaskBuilder(t *testing.T) {
	doc := &Document{
		Meta: Metadata{Date: "2026-09-01"},
	}

	// 1. AddTask
	task1, _ := NewTask("Первая задача").Build()
	if err := doc.AddTask(task1); err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}
	if len(doc.Tasks) != 1 {
		t.Fatalf("len(Tasks) = %d, хотели 1", len(doc.Tasks))
	}

	// 2. NewTask builder прямо из документа
	err := doc.NewTask("Вторая задача").
		Tag("home").
		Done(true).
		Add()

	if err != nil {
		t.Fatalf("doc.NewTask.Add failed: %v", err)
	}

	if len(doc.Tasks) != 2 {
		t.Fatalf("len(Tasks) = %d, хотели 2", len(doc.Tasks))
	}
	if doc.Meta.TaskCompleted != 1 {
		t.Errorf("TaskCompleted = %d, хотели 1 (так как вторая Done=true)", doc.Meta.TaskCompleted)
	}
}

func TestDocument_ToggleTask_And_Stats(t *testing.T) {
	doc := &Document{
		Tasks: []Task{
			{Title: "Task 1", Done: false},
			{Title: "Task 2", Done: true},
		},
		Meta: Metadata{TaskCompleted: 1},
	}

	// 1. Включаем Task 1 (false -> true)
	if err := doc.ToggleTask(0); err != nil {
		t.Fatalf("ToggleTask(0) error: %v", err)
	}
	if !doc.Tasks[0].Done || doc.Meta.TaskCompleted != 2 {
		t.Errorf("Task 1 Done = %v, TaskCompleted = %d (хотели true, 2)", doc.Tasks[0].Done, doc.Meta.TaskCompleted)
	}

	// 2. Выключаем Task 2 (true -> false)
	if err := doc.ToggleTask(1); err != nil {
		t.Fatalf("ToggleTask(1) error: %v", err)
	}
	if doc.Tasks[1].Done || doc.Meta.TaskCompleted != 1 {
		t.Errorf("Task 2 Done = %v, TaskCompleted = %d (хотели false, 1)", doc.Tasks[1].Done, doc.Meta.TaskCompleted)
	}

	// 3. Выход за границы
	if err := doc.ToggleTask(99); err != ErrTaskNotFound {
		t.Errorf("ожидалась ErrTaskNotFound, получено: %v", err)
	}
}

func TestDocument_DeleteTask(t *testing.T) {
	doc := &Document{
		Tasks: []Task{
			{Title: "Task 1", Done: true},
			{Title: "Task 2", Done: false},
			{Title: "Task 3", Done: true},
		},
		Meta: Metadata{TaskCompleted: 2},
	}

	// Удаляем выполненную Task 1
	if err := doc.DeleteTask(0); err != nil {
		t.Fatalf("DeleteTask(0) error: %v", err)
	}

	if len(doc.Tasks) != 2 || doc.Tasks[0].Title != "Task 2" {
		t.Errorf("Tasks = %v", doc.Tasks)
	}
	if doc.Meta.TaskCompleted != 1 {
		t.Errorf("TaskCompleted = %d, хотели 1", doc.Meta.TaskCompleted)
	}

	// Ошибка при неверном индексе
	if err := doc.DeleteTask(10); err != ErrTaskNotFound {
		t.Errorf("ожидалась ErrTaskNotFound, получено: %v", err)
	}
}

func TestDocument_UpdateTaskTitle(t *testing.T) {
	doc := &Document{
		Tasks: []Task{{Title: "Старый заголовок"}},
	}

	if err := doc.UpdateTaskTitle(0, "  Новый заголовок  "); err != nil {
		t.Fatalf("UpdateTaskTitle error: %v", err)
	}
	if doc.Tasks[0].Title != "Новый заголовок" {
		t.Errorf("Title = %q, хотели 'Новый заголовок'", doc.Tasks[0].Title)
	}

	if err := doc.UpdateTaskTitle(0, "   "); err != ErrEmptyTitle {
		t.Errorf("ожидалась ErrEmptyTitle, получено: %v", err)
	}
}

func TestDocument_ReorderTasks(t *testing.T) {
	doc := &Document{
		Tasks: []Task{
			{Title: "A"},
			{Title: "B"},
			{Title: "C"},
			{Title: "D"},
		},
	}

	// Перемещаем A (0) на позицию 2 -> B, C, A, D
	if err := doc.ReorderTasks(0, 2); err != nil {
		t.Fatalf("ReorderTasks error: %v", err)
	}
	expected := []string{"B", "C", "A", "D"}
	for i, name := range expected {
		if doc.Tasks[i].Title != name {
			t.Errorf("Tasks[%d] = %q, хотели %q", i, doc.Tasks[i].Title, name)
		}
	}

	// Перемещаем D (3) на позицию 0 -> D, B, C, A
	if err := doc.ReorderTasks(3, 0); err != nil {
		t.Fatalf("ReorderTasks error: %v", err)
	}
	expected2 := []string{"D", "B", "C", "A"}
	for i, name := range expected2 {
		if doc.Tasks[i].Title != name {
			t.Errorf("Tasks[%d] = %q, хотели %q", i, doc.Tasks[i].Title, name)
		}
	}
}

func TestDocument_Notes(t *testing.T) {
	doc := &Document{
		Tasks: []Task{
			{Title: "Task with notes", Notes: []string{"Note 1", "Note 2"}},
		},
	}

	// Добавление
	if err := doc.AddNoteToTask(0, "  Note 3  "); err != nil {
		t.Fatalf("AddNoteToTask error: %v", err)
	}
	if len(doc.Tasks[0].Notes) != 3 || doc.Tasks[0].Notes[2] != "Note 3" {
		t.Errorf("Notes = %v", doc.Tasks[0].Notes)
	}

	// Удаление Note 1
	if err := doc.DeleteNoteFromTask(0, 0); err != nil {
		t.Fatalf("DeleteNoteFromTask error: %v", err)
	}
	if len(doc.Tasks[0].Notes) != 2 || doc.Tasks[0].Notes[0] != "Note 2" {
		t.Errorf("Notes = %v", doc.Tasks[0].Notes)
	}

	// Ошибки
	if err := doc.DeleteNoteFromTask(0, 99); err != ErrNoteNotFound {
		t.Errorf("ожидалась ErrNoteNotFound, получено: %v", err)
	}
}

func TestDocument_DailyLog(t *testing.T) {
	doc := &Document{}

	doc.AddDailyLog("  Запись 1  ")
	doc.AddDailyLog("Запись 2")

	if len(doc.DailyLog) != 2 || doc.DailyLog[0] != "Запись 1" {
		t.Errorf("DailyLog = %v", doc.DailyLog)
	}

	if err := doc.DeleteDailyLog(0); err != nil {
		t.Fatalf("DeleteDailyLog error: %v", err)
	}
	if len(doc.DailyLog) != 1 || doc.DailyLog[0] != "Запись 2" {
		t.Errorf("DailyLog = %v", doc.DailyLog)
	}
}

func TestDocument_Search_And_Filter(t *testing.T) {
	doc := &Document{
		Tasks: []Task{
			{Title: "Купить молоко", Tags: []string{"home"}, Done: false},
			{Title: "Написать тесты", Tags: []string{"dev", "go"}, Attributes: []Attribute{{Key: "prio", Value: "high"}, {Key: "due", Value: "2026-08-01"}}, Done: false},
			{Title: "Починить баг", Tags: []string{"dev"}, Attributes: []Attribute{{Key: "due", Value: "2026-09-30"}}, Done: true},
		},
	}

	// 1. SearchTasks
	results := doc.SearchTasks("молоко")
	if len(results) != 1 || results[0].Title != "Купить молоко" {
		t.Errorf("SearchTasks('молоко') = %v", results)
	}

	// 2. TasksByTag
	devTasks := doc.TasksByTag("dev")
	if len(devTasks) != 2 {
		t.Errorf("TasksByTag('dev') len = %d, хотели 2", len(devTasks))
	}

	// 3. TasksByAttr
	highPrio := doc.TasksByAttr("prio", "high")
	if len(highPrio) != 1 || highPrio[0].Title != "Написать тесты" {
		t.Errorf("TasksByAttr('prio', 'high') = %v", highPrio)
	}

	// 4. OverdueTasks
	today := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)
	overdue := doc.OverdueTasks(today)
	if len(overdue) != 1 || overdue[0].Title != "Написать тесты" {
		t.Errorf("OverdueTasks = %v", overdue)
	}
}
