package engine

import (
	"context"
	"errors"
	"fmt"
	"jrn/internal/parser"
	"jrn/pkg/domain"
	"jrn/pkg/storage"
	"time"
)

type Engine struct {
	storage domain.Storage
}

// New создает новый экземпляр Engine с переданным хранилищем.
func New(store domain.Storage) *Engine {
	return &Engine{
		storage: store,
	}
}

// NewDefault создает экземпляр Engine со стандартным файловым хранилищем по указанному пути.
func NewDefault(dir string) (*Engine, error) {
	store, err := storage.New(dir)
	if err != nil {
		return nil, err
	}
	return New(store), nil
}

// OpenToday открывает или инициализирует документ на текущий день.
func (e *Engine) OpenToday(ctx context.Context) (*domain.Document, error) {
	today := time.Now()
	return e.OpenDay(ctx, today)
}

// OpenDay открывает существующий день либо инициализирует новый с расчетом Streak и переносом задач.
func (e *Engine) OpenDay(ctx context.Context, date time.Time) (*domain.Document, error) {
	data, err := e.storage.Load(ctx, date)
	if err == nil {
		doc, err := parser.Parse(data)
		if err != nil {
			return nil, fmt.Errorf("ошибка парсинга документа %s: %w", date.Format("2006-01-02"), err)
		}
		return doc, nil
	}

	if !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("ошибка чтения дня %s: %w", date.Format("2006-01-02"), err)
	}

	todayStr := time.Now().Format("2006-01-02")
	if date.Format("2006-01-02") > todayStr {
		return nil, domain.ErrDateInFuture
	}

	doc, err := e.initNewDay(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации дня %s: %w", date.Format("2006-01-02"), err)
	}

	if err := e.SaveDay(ctx, doc); err != nil {
		return nil, fmt.Errorf("ошибка сохранения нового дня %s: %w", date.Format("2006-01-02"), err)
	}

	return doc, nil
}

// SaveDay сохраняет документ дня на диск.
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
	isYesterday := prevDate.Format("2006-01-02") == date.AddDate(0, 0, -1).Format("2006-01-02")

	if prevTasksCount > 0 && prevCompleted == prevTasksCount && isYesterday {
		doc.Meta.Streak = prevStreak + 1
	} else {
		doc.Meta.Streak = 1
	}

	for _, task := range prevDoc.Tasks {
		// Rollover
		if !task.Done {
			newTask := task.Clone()
			newTask.Done = false
			doc.Tasks = append(doc.Tasks, newTask)
			continue
		}

		// привычка
		if task.ShouldRepeatOn(date) {
			newTask := task.Clone()
			newTask.Done = false
			doc.Tasks = append(doc.Tasks, newTask)
		}
	}

	return doc, nil
}

// OpenPreviousDay находит и открывает предыдущую сохраненную запись.
func (e *Engine) OpenPreviousDay(ctx context.Context, currentDate time.Time) (*domain.Document, error) {
	prevDate, found, err := e.storage.FindPreviousDate(ctx, currentDate)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, domain.ErrNotFound
	}

	return e.OpenDay(ctx, prevDate)
}

// OpenNextDay находит и открывает следующий сохраненный день.
func (e *Engine) OpenNextDay(ctx context.Context, currentDate time.Time) (*domain.Document, error) {
	nextDate, found, err := e.storage.FindNextDate(ctx, currentDate)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, domain.ErrNotFound
	}

	return e.OpenDay(ctx, nextDate)
}

// ListAvailableDays возвращает все сохраненные даты в хранилище.
func (e *Engine) ListAvailableDays(ctx context.Context) ([]time.Time, error) {
	return e.storage.ListDates(ctx)
}

type SearchResult struct {
	Date string      `json:"date"`
	Task domain.Task `json:"task"`
}

// Search выполняет глобальный полнотекстовый поиск задач по всем дням в архиве.
func (e *Engine) Search(ctx context.Context, query string) ([]SearchResult, error) {
	dates, err := e.storage.ListDates(ctx)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, d := range dates {
		doc, err := e.OpenDay(ctx, d)
		if err != nil {
			continue
		}
		for _, task := range doc.SearchTasks(query) {
			results = append(results, SearchResult{
				Date: doc.Meta.Date,
				Task: task,
			})
		}
	}
	return results, nil
}

// GetTasksByTag ищет задачи с указанным тегом по всему архиву.
func (e *Engine) GetTasksByTag(ctx context.Context, tag string) ([]SearchResult, error) {
	dates, err := e.storage.ListDates(ctx)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, d := range dates {
		doc, err := e.OpenDay(ctx, d)
		if err != nil {
			continue
		}
		for _, task := range doc.TasksByTag(tag) {
			results = append(results, SearchResult{
				Date: doc.Meta.Date,
				Task: task,
			})
		}
	}
	return results, nil
}

// GetOverdueTasks находит все незавершенные просроченные задачи относительно указанной даты.
func (e *Engine) GetOverdueTasks(ctx context.Context, today time.Time) ([]SearchResult, error) {
	dates, err := e.storage.ListDates(ctx)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, d := range dates {
		doc, err := e.OpenDay(ctx, d)
		if err != nil {
			continue
		}
		for _, task := range doc.OverdueTasks(today) {
			results = append(results, SearchResult{
				Date: doc.Meta.Date,
				Task: task,
			})
		}
	}
	return results, nil
}
