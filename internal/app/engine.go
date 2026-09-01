package app

import (
	"context"
	"errors"
	"fmt"
	"jrn/internal/domain"
	"jrn/internal/infrastructure/parser"
	"time"
)

type Engine struct {
	storage domain.Storage
}

func New(storage domain.Storage) *Engine {
	return &Engine{
		storage: storage,
	}
}

func (e *Engine) OpenToday(ctx context.Context) (*domain.Document, error) {
	today := time.Now()
	return e.OpenDay(ctx, today)
}

func (e *Engine) OpenDay(ctx context.Context, date time.Time) (*domain.Document, error) {
	data, err := e.storage.Load(ctx, date)
	if err == nil {
		doc, err := parser.Parse(data)
		if err != nil {
			return nil, fmt.Errorf("ошибка парсинга документа %s: %w", date, err)
		}
		return doc, nil
	}

	if !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("ошибка чтения дня %s: %w", date, err)
	}

	doc, err := e.initNewDay(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации дня %s: %w", date, err)
	}

	if err := e.SaveDay(ctx, doc); err != nil {
		return nil, fmt.Errorf("ошибка сохранения нового дня %s: %w", date, err)
	}

	return doc, nil
}

func (e *Engine) SaveDay(ctx context.Context, doc *domain.Document) error {
	if doc == nil {
		return domain.ErrCorruptedData
	}
	t, err := time.Parse("2006-01-02", doc.Meta.Date)

	if err != nil {
		return err
	}

	raw := parser.Serialize(doc)
	return e.storage.Save(ctx, t, raw)
}

func (e *Engine) initNewDay(ctx context.Context, date time.Time) (*domain.Document, error) {
	doc := &domain.Document{
		Meta: domain.Metadata{
			Date:          date.Format("2006-01-02"),
			Streak:        1,
			TaskCompleted: 0,
		},
		DailyLog: make([]string, 0),
		Tasks:    make([]domain.Task, 0),
	}

	prevDate, found, err := e.storage.FindPreviousDate(ctx, date)
	if err != nil {
		return nil, err
	}

	if !found {
		return doc, nil
	}

	prevData, err := e.storage.Load(ctx, prevDate)
	if err != nil {
		return doc, nil
	}

	prevDoc, err := parser.Parse(prevData)
	if err != nil {
		return doc, nil
	}

	prevCompleted := prevDoc.Meta.TaskCompleted
	prevStreak := prevDoc.Meta.Streak
	prevTasksCount := len(prevDoc.Tasks)

	if prevTasksCount > 0 && prevCompleted == prevTasksCount {
		doc.Meta.Streak = prevStreak + 1
	} else {
		doc.Meta.Streak = 1
	}

	// Rollover
	// for _, task := range prevDoc.Tasks {
	// 	if !task.Done {
	// 		doc.Tasks = append(doc.Tasks, domain.Task{
	// 			Done:       false,
	// 			Title:      task.Title,
	// 			Tags:       append([]string(nil), task.Tags...),
	// 			Attributes: append([]domain.Attribute(nil), task.Attributes...),
	// 			Notes:      append([]string(nil), task.Notes...),
	// 		})
	// 	}
	// }

	return doc, nil
}
