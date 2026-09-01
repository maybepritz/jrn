package engine

import (
	"context"
	"jrn/internal/parser"
	"jrn/pkg/domain"
	"jrn/pkg/storage"
	"testing"
	"time"
)

func TestEngine_OpenDay_InitNewDay_Streak_And_Rollover(t *testing.T) {
	ctx := context.Background()
	mockStore := storage.NewMock()
	engine := New(mockStore)

	// 1. Создаем день 1 (2026-08-28): 2 задачи, 1 выполнена
	day1Doc := &domain.Document{
		Meta: domain.Metadata{
			Date:          "2026-08-28",
			Streak:        1,
			TaskCompleted: 1,
		},
		Tasks: []domain.Task{
			{Title: "Выполненная задача", Done: true, Tags: []string{"habit"}},
			{Title: "Невыполненная задача", Done: false, Tags: []string{"dev"}, Notes: []string{"Заметка 1"}},
		},
	}
	mockStore.Seed("2026-08-28", parser.Serialize(day1Doc))

	// 2. Открываем следующий день (2026-08-29)
	day2Date := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)
	day2Doc, err := engine.OpenDay(ctx, day2Date)
	if err != nil {
		t.Fatalf("OpenDay failed: %v", err)
	}

	// Проверяем Streak: так как 28-го выполнены не все задачи (1 из 2), Streak должен сброситься в 1
	if day2Doc.Meta.Streak != 1 {
		t.Errorf("Streak = %d, хотели 1", day2Doc.Meta.Streak)
	}

	// Проверяем Rollover: невыполненная задача должна перенестись на 29-е число
	if len(day2Doc.Tasks) != 1 {
		t.Fatalf("len(Tasks) = %d, хотели 1 (перенесенную задачу)", len(day2Doc.Tasks))
	}
	if day2Doc.Tasks[0].Title != "Невыполненная задача" {
		t.Errorf("Title = %q, хотели 'Невыполненная задача'", day2Doc.Tasks[0].Title)
	}
	if len(day2Doc.Tasks[0].Notes) != 1 || day2Doc.Tasks[0].Notes[0] != "Заметка 1" {
		t.Errorf("Notes = %v", day2Doc.Tasks[0].Notes)
	}

	// 3. Выполняем перенесенную задачу на 29-е и сохраняем
	day2Doc.Tasks[0].Done = true
	day2Doc.Meta.TaskCompleted = 1
	if err := engine.SaveDay(ctx, day2Doc); err != nil {
		t.Fatalf("SaveDay failed: %v", err)
	}

	// 4. Открываем 3-й день (2026-08-30): так как 29-го выполнены 100% задач, Streak должен стать 2
	day3Date := time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC)
	day3Doc, err := engine.OpenDay(ctx, day3Date)
	if err != nil {
		t.Fatalf("OpenDay day 3 failed: %v", err)
	}

	if day3Doc.Meta.Streak != 2 {
		t.Errorf("Streak = %d, хотели 2", day3Doc.Meta.Streak)
	}
	if len(day3Doc.Tasks) != 0 {
		t.Errorf("len(Tasks) = %d, хотели 0 (все задачи были закрыты)", len(day3Doc.Tasks))
	}
}

func TestEngine_Navigation_And_ListDays(t *testing.T) {
	ctx := context.Background()
	mockStore := storage.NewMock()
	engine := New(mockStore)

	// Засеиваем несколько дней
	doc1 := &domain.Document{Meta: domain.Metadata{Date: "2026-08-20"}}
	doc2 := &domain.Document{Meta: domain.Metadata{Date: "2026-08-25"}}
	doc3 := &domain.Document{Meta: domain.Metadata{Date: "2026-08-29"}}

	mockStore.Seed("2026-08-20", parser.Serialize(doc1))
	mockStore.Seed("2026-08-25", parser.Serialize(doc2))
	mockStore.Seed("2026-08-29", parser.Serialize(doc3))

	// 1. ListAvailableDays
	days, err := engine.ListAvailableDays(ctx)
	if err != nil {
		t.Fatalf("ListAvailableDays failed: %v", err)
	}
	if len(days) != 3 {
		t.Fatalf("len(days) = %d, хотели 3", len(days))
	}

	// 2. OpenPreviousDay от 2026-08-25 -> должен открыть 2026-08-20
	current := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	prevDoc, err := engine.OpenPreviousDay(ctx, current)
	if err != nil {
		t.Fatalf("OpenPreviousDay failed: %v", err)
	}
	if prevDoc.Meta.Date != "2026-08-20" {
		t.Errorf("Date = %q, хотели 2026-08-20", prevDoc.Meta.Date)
	}

	// 3. OpenNextDay от 2026-08-25 -> должен открыть 2026-08-29
	nextDoc, err := engine.OpenNextDay(ctx, current)
	if err != nil {
		t.Fatalf("OpenNextDay failed: %v", err)
	}
	if nextDoc.Meta.Date != "2026-08-29" {
		t.Errorf("Date = %q, хотели 2026-08-29", nextDoc.Meta.Date)
	}
}

func TestEngine_Search_And_Filters(t *testing.T) {
	ctx := context.Background()
	mockStore := storage.NewMock()
	engine := New(mockStore)

	day1 := &domain.Document{
		Meta: domain.Metadata{Date: "2026-08-20"},
		Tasks: []domain.Task{
			{Title: "Купить молоко", Tags: []string{"home"}, Done: true},
			{Title: "Написать лексер", Tags: []string{"dev", "go"}, Attributes: []domain.Attribute{{Key: "prio", Value: "high"}}, Done: false},
		},
	}
	day2 := &domain.Document{
		Meta: domain.Metadata{Date: "2026-08-25"},
		Tasks: []domain.Task{
			{Title: "Починить парсер", Tags: []string{"dev"}, Attributes: []domain.Attribute{{Key: "due", Value: "2026-08-22"}}, Done: false},
		},
	}

	mockStore.Seed("2026-08-20", parser.Serialize(day1))
	mockStore.Seed("2026-08-25", parser.Serialize(day2))

	// 1. Search ("парсер")
	results, err := engine.Search(ctx, "парсер")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 || results[0].Task.Title != "Починить парсер" || results[0].Date != "2026-08-25" {
		t.Errorf("Search results = %v", results)
	}

	// 2. GetTasksByTag ("dev") -> 2 задачи из разных дней
	tagResults, err := engine.GetTasksByTag(ctx, "dev")
	if err != nil {
		t.Fatalf("GetTasksByTag failed: %v", err)
	}
	if len(tagResults) != 2 {
		t.Errorf("GetTasksByTag len = %d, хотели 2", len(tagResults))
	}

	// 3. GetOverdueTasks (на дату 2026-08-29) -> задача с due 2026-08-22
	today := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)
	overdue, err := engine.GetOverdueTasks(ctx, today)
	if err != nil {
		t.Fatalf("GetOverdueTasks failed: %v", err)
	}
	if len(overdue) != 1 || overdue[0].Task.Title != "Починить парсер" {
		t.Errorf("Overdue results = %v", overdue)
	}
}
