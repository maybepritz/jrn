package domain

import (
	"context"
	"time"
)

type Storage interface {
	Load(ctx context.Context, date time.Time) ([]byte, error)
	Save(ctx context.Context, date time.Time, data []byte) error
	FindPreviousDate(ctx context.Context, before time.Time) (time.Time, bool, error)
	FindNextDate(ctx context.Context, after time.Time) (time.Time, bool, error)
	ListDates(ctx context.Context, date time.Time) ([]time.Time, error)
	Exists(ctx context.Context, date time.Time) (bool, error)
}
