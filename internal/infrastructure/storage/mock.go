package storage

import (
	"context"
	"jrn/internal/domain"
	"sort"
	"sync"
	"time"
)

// Mock — in-memory реализация domain.Storage для тестов.
type Mock struct {
	mu   sync.RWMutex
	data map[string][]byte // ключ: "2006-01-02"
}

func NewMock() *Mock {
	return &Mock{
		data: make(map[string][]byte),
	}
}

// Seed — загрузить тестовые данные напрямую.
func (m *Mock) Seed(date string, content []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[date] = append([]byte(nil), content...)
}

func (m *Mock) Load(ctx context.Context, date time.Time) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := date.Format("2006-01-02")
	data, ok := m.data[key]
	if !ok {
		return nil, domain.ErrNotFound
	}
	// Возвращаем копию, чтобы тесты не могли мутировать внутреннее состояние
	out := make([]byte, len(data))
	copy(out, data)
	return out, nil
}

func (m *Mock) Save(ctx context.Context, date time.Time, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := date.Format("2006-01-02")
	m.data[key] = append([]byte(nil), data...)
	return nil
}

func (m *Mock) FindPreviousDate(ctx context.Context, before time.Time) (time.Time, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	beforeStr := before.Format("2006-01-02")

	// Собираем и сортируем ключи
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Ищем максимальную дату строго меньше before
	for i := len(keys) - 1; i >= 0; i-- {
		if keys[i] < beforeStr {
			t, err := time.Parse("2006-01-02", keys[i])
			if err != nil {
				return time.Time{}, false, err
			}
			return t, true, nil
		}
	}

	return time.Time{}, false, nil
}

func (m *Mock) FindNextDate(ctx context.Context, after time.Time) (time.Time, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	afterStr := after.Format("2006-01-02")

	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Ищем минимальную дату строго больше after
	for _, key := range keys {
		if key > afterStr {
			t, err := time.Parse("2006-01-02", key)
			if err != nil {
				return time.Time{}, false, err
			}
			return t, true, nil
		}
	}

	return time.Time{}, false, nil
}

func (m *Mock) ListDates(ctx context.Context) ([]time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	dates := make([]time.Time, 0, len(keys))
	for _, key := range keys {
		t, err := time.Parse("2006-01-02", key)
		if err != nil {
			return nil, err
		}
		dates = append(dates, t)
	}

	return dates, nil
}

func (m *Mock) Exists(ctx context.Context, date time.Time) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := date.Format("2006-01-02")
	_, ok := m.data[key]
	return ok, nil
}

// Dates — вернуть все сохранённые даты (для отладки тестов).
func (m *Mock) Dates() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
