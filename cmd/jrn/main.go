package main

import (
	"context"
	"fmt"
	"jrn/internal/infrastructure/storage"
	"path/filepath"
	"time"
)

func main() {
	ctx := context.Background()

	// 1. Хранилище
	storage, err := storage.New(filepath.Join(".", "data"))
	if err != nil {
		panic(err)
	}

	date, find, err := storage.FindNextDate(ctx, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Previous date: %s, found: %v\n", date.Format("2006-01-02"), find)

	// 2. Движок
	// engine := app.New(storage)

	// // 3. Открываем сегодняшний день
	// doc, err := engine.OpenToday(ctx)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Printf("День: %s | Серия: %d дней | Задач: %d\n", doc.Meta.Date, doc.Meta.Streak, len(doc.Tasks))
}
