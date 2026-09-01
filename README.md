# jrn

[![Go Reference](https://pkg.go.dev/badge/github.com/maybepritz/task-cli-go.svg)](https://pkg.go.dev/github.com/maybepritz/task-cli-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/maybepritz/task-cli-go)](https://goreportcard.com/report/github.com/maybepritz/task-cli-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Встраиваемая библиотека на Go и C ABI SDK для ежедневного ведения заметок, трекинга задач, подсчёта серий продуктивности (Streak) и автоматического переноса незавершённых дел (Rollover).

Данные хранятся в обычных Markdown-файлах с YAML frontmatter.

---

## Возможности

* **Честный Markdown на диске:** Записи хранятся в читаемых файлах по структуре `YYYY/MM/YYYY-MM-DD.md`.
* **Быстрый парсер на конечном автомате (FSM):** Работа со строками через String Arena и String Interning (~3.9 мкс / 35 аллокаций на документ).
* **Автоматический Rollover и Streak:** Незакрытые задачи автоматически переносятся на следующий день, а серии выполнения рассчитываются только при полном закрытии дня.
* **C-Shared биндинги:** Экспорт функций через C ABI для написания клиентов на Python, Java, C#, Rust или Electron.

---

## Формат данных на диске

Каждый день записывается в стандартный Markdown-файл:

```markdown
---
date: 2026-09-01
streak: 5
task_completed: 2
---

### Daily Log
Обсудили с командой новую архитектуру.
Закончили оптимизацию парсера.

### Tasks
- [x] Утренняя пробежка #health
- [x] Ревью пулл-реквестов #dev @prio(high)
- [ ] Реализовать поиск в TUI #dev @due(2026-09-02)
  - Посмотреть компонент textinput в bubbletea
```

---

## Установка

```bash
go get github.com/maybepritz/task-cli-go
```

*(Для локальной разработки укажите в `go.mod` вашего проекта: `replace jrn => ../task-cli-go`)*.

---

## Быстрый старт (Go)

```go
package main

import (
	"context"
	"fmt"
	"log"

	"jrn/pkg/engine"
)

func main() {
	ctx := context.Background()

	// 1. Инициализация движка (~ разворачивается в домашнюю директорию)
	core, err := engine.NewDefault("~/.jrn")
	if err != nil {
		log.Fatal(err)
	}

	// 2. Открытие дня (создается автоматически с переносом задач и расчетом серии)
	doc, err := core.OpenToday(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Создание задачи через builder
	err = doc.NewTask("Реализовать TUI на Bubbletea").
		Tag("dev", "go").
		Attr("prio", "high").
		Note("Проверить обработку клавиатурных событий").
		Add()

	// 4. Переключение статуса задачи
	_ = doc.ToggleTask(0) // переключает Done и обновляет task_completed

	// 5. Добавление записи в дневник
	doc.AddDailyLog("Переписали слой хранилища.")

	// 6. Сохранение изменений на диск
	if err := core.SaveDay(ctx, doc); err != nil {
		log.Fatal(err)
	}

	// 7. Поиск по всему архиву
	results, _ := core.Search(ctx, "Bubbletea")
	for _, res := range results {
		fmt.Printf("[%s] %s (done: %v)\n", res.Date, res.Task.Title, res.Task.Done)
	}
}
```

---

## Архитектура пакетов

```
pkg/
├── domain/       # Доменные модели (Task, Document, TaskBuilder, ошибки)
├── engine/       # Сервисный слой (OpenToday, OpenDay, Search, SaveDay)
└── storage/      # Драйвер файловой системы (с ~) и in-memory Mock

internal/
└── parser/       # FSM парсер Markdown и fastjson сериализатор

cmd/
└── jrn-c/        # CGO точка входа для компиляции jrn.dll / jrn.so
```

---

## Использование в других языках (C-Shared DLL)

Сборка `jrn.dll` (Windows) или `jrn.so` (Linux/macOS):

```bash
# Windows
$env:CGO_ENABLED="1"
go build -buildmode=c-shared -o jrn.dll ./cmd/jrn-c

# Linux / macOS
CGO_ENABLED=1 go build -buildmode=c-shared -o jrn.so ./cmd/jrn-c
```

### Python (`ctypes`)

```python
import ctypes, json

jrn = ctypes.CDLL("./jrn.dll")
jrn.JRN_Init.argtypes = [ctypes.c_char_p]
jrn.JRN_OpenTodayJSON.restype = ctypes.c_char_p
jrn.JRN_FreeString.argtypes = [ctypes.c_char_p]

jrn.JRN_Init(b"~/.jrn")
raw = jrn.JRN_OpenTodayJSON()
doc = json.loads(raw.decode("utf-8"))

print(f"Дата: {doc['meta']['date']} | Серия: {doc['meta']['streak']}")
```

### Java (JNA)

```java
public interface JrnLib extends Library {
    JrnLib INSTANCE = Native.load("jrn.dll", JrnLib.class);
    int JRN_Init(String path);
    String JRN_OpenTodayJSON();
    int JRN_AddTask(String dateStr, String taskJson);
}

// Использование:
JrnLib.INSTANCE.JRN_Init("~/.jrn");
String json = JrnLib.INSTANCE.JRN_OpenTodayJSON();
```

---

## Бенчмарки

Замеры производительности на AMD Ryzen 7 / Go 1.22:

```
BenchmarkParse-16          294625       3921 ns/op     427.70 MB/s      35 allocs/op
BenchmarkParseJson-16      221528       5413 ns/op     409.77 MB/s      48 allocs/op
BenchmarkSerialize-16     2617711        458 ns/op    3657.87 MB/s       2 allocs/op
```

---

## Лицензия

[MIT](LICENSE)
