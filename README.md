# 📓 JRN Core (Task & Journal Engine)

**JRN Core** — это высокопроизводительное встраиваемое ядро (SDK) для ведения ежедневных журналов и управления задачами в формате Markdown с поддержкой JSON и расчётом продуктивности (Streak / Rollover).

Разработано по принципам **Clean Architecture** с ультрабыстрым нуль-аллокационным конечным автоматом (FSM) для парсинга и сериализации данных.

---

## ⚡ Особенности

* **Формат хранения:** Прозрачные Markdown-файлы (`.md`) с YAML frontmatter или структурированный JSON.
* **Экстремальная производительность:** Парсер на базе FSM со String Interning и String Arena (**~3.9 мкс / 35 аллокаций**).
* **Умный расчет серии (Streak):** Автоматический пересчет непрерывности дней и прогресса выполнения.
* **Перенос задач (Rollover):** Невыполненные задачи со всеми тегами, атрибутами и заметками автоматически переносятся на новый день.
* **Мультиязычность:** Готов к импорту в Go-проекты, а также компилируется в C-Shared DLL (`.dll`, `.so`, `.dylib`) для Python, Java, C#, C++, Rust, Node.js.

---

## 📁 Архитектура пакетов

```
task-cli-go/
│
├── pkg/                             # 🟢 ПУБЛИЧНЫЙ API (для внешних проектов)
│   ├── domain/                      # Сущности (Task, Document, TaskBuilder), ошибки, интерфейс Storage
│   ├── engine/                      # Сервисный слой (OpenToday, OpenDay, Search, SaveDay)
│   └── storage/                     # Файловая система (с поддержкой ~) и in-memory Mock
│
├── internal/                        # 🔒 ПРИВАТНЫЙ СЛОЙ (скрыт от внешнего импорта)
│   └── parser/                      # FSM Markdown парсер и fastjson сериализатор
│
└── cmd/
    └── jrn-c/                       # CGO точка входа для сборки C-Shared библиотеки (DLL / .so)
```

---

## 🚀 Использование в Go

### 1. Подключение

```bash
go get github.com/maybepritz/task-cli-go
```

*(Для локальной разработки в `go.mod` вашего приложения добавьте `replace jrn => ../task-cli-go`)*.

### 2. Пример кода

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

    // 1. Инициализируем ядро с путем к хранилищу (поддерживается ~)
    core, err := engine.NewDefault("~/.jrn")
    if err != nil {
        log.Fatalf("Ошибка: %v", err)
    }

    // 2. Открываем сегодняшний день (создастся автоматически с расчетом Streak и Rollover)
    doc, err := core.OpenToday(ctx)
    if err != nil {
        log.Fatalf("Ошибка: %v", err)
    }

    // 3. Создаем задачу через fluent builder
    err = doc.NewTask("Изучить архитектуру Go").
        Tag("dev", "study").
        Attr("prio", "high").
        Attr("due", "2026-09-05").
        Note("Почитать про memory layout").
        Add()

    // 4. Переключаем статус задачи (Done <-> !Done с авто-пересчетом счетчика)
    _ = doc.ToggleTask(0)

    // 5. Добавляем запись в дневник дня
    doc.AddDailyLog("Успешно настроили интеграцию ядра.")

    // 6. Сохраняем изменения на диск
    if err := core.SaveDay(ctx, doc); err != nil {
        log.Fatalf("Ошибка сохранения: %v", err)
    }

    // 7. Поиск по всему архиву
    results, _ := core.Search(ctx, "архитектура")
    for _, res := range results {
        fmt.Printf("[%s] %s (Done: %v)\n", res.Date, res.Task.Title, res.Task.Done)
    }
}
```

---

## 🌍 Использование в других языках (C-Shared DLL / .so)

### Сборка библиотеки

Для компиляции нативной библиотеки выполните:

```bash
# Windows (создаст jrn.dll и jrn.h)
$env:CGO_ENABLED="1"
go build -buildmode=c-shared -o jrn.dll ./cmd/jrn-c

# Linux / macOS (создаст jrn.so или jrn.dylib)
CGO_ENABLED=1 go build -buildmode=c-shared -o jrn.so ./cmd/jrn-c
```

---

### 🐍 Python (`ctypes`)

```python
import ctypes
import json

# Загружаем библиотеку
jrn = ctypes.CDLL("./jrn.dll")

# Настройка сигнатур
jrn.JRN_Init.argtypes = [ctypes.c_char_p]
jrn.JRN_Init.restype = ctypes.c_int

jrn.JRN_OpenTodayJSON.restype = ctypes.c_char_p
jrn.JRN_AddTask.argtypes = [ctypes.c_char_p, ctypes.c_char_p]
jrn.JRN_AddTask.restype = ctypes.c_int
jrn.JRN_FreeString.argtypes = [ctypes.c_char_p]

# 1. Инициализация
jrn.JRN_Init(b"~/.jrn")

# 2. Получение данных дня в JSON
raw_json = jrn.JRN_OpenTodayJSON()
doc = json.loads(raw_json.decode("utf-8"))
print(f"Дата: {doc['meta']['date']}, Серия: {doc['meta']['streak']} дней")

# 3. Добавление задачи
task_json = json.dumps({"title": "Задача из Python", "tags": ["py", "demo"]})
jrn.JRN_AddTask(b"2026-09-01", task_json.encode("utf-8"))
```

---

### ☕ Java (JNA)

```java
import com.sun.jna.Library;
import com.sun.jna.Native;
import com.sun.jna.Pointer;

public class JrnExample {
    public interface JrnLib extends Library {
        JrnLib INSTANCE = Native.load("jrn.dll", JrnLib.class);

        int JRN_Init(String path);
        String JRN_OpenTodayJSON();
        int JRN_AddTask(String dateStr, String taskJSON);
        int JRN_ToggleTask(String dateStr, int index);
        String JRN_SearchJSON(String query);
        void JRN_FreeString(Pointer str);
    }

    public static void main(String[] args) {
        JrnLib jrn = JrnLib.INSTANCE;

        jrn.JRN_Init("~/.jrn");
        String jsonStr = jrn.JRN_OpenTodayJSON();
        System.out.println("Сегодняшний день: " + jsonStr);

        jrn.JRN_AddTask("2026-09-01", "{\"title\":\"Задача из Java\",\"tags\":[\"java\"]}");
    }
}
```

---

### 🔷 C# (.NET P/Invoke)

```csharp
using System;
using System.Runtime.InteropServices;

class Program {
    [DllImport("jrn.dll", CallingConvention = CallingConvention.Cdecl)]
    public static extern int JRN_Init(string path);

    [DllImport("jrn.dll", CallingConvention = CallingConvention.Cdecl)]
    public static extern IntPtr JRN_OpenTodayJSON();

    static void Main() {
        JRN_Init("~/.jrn");
        IntPtr ptr = JRN_OpenTodayJSON();
        string json = Marshal.PtrToStringAnsi(ptr);
        Console.WriteLine(json);
    }
}
```

---

### 🦀 Rust (`libloading`)

```rust
use libloading::{Library, Symbol};
use std::ffi::{CStr, CString};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    unsafe {
        let lib = Library::new("jrn.dll")?;

        let init: Symbol<unsafe extern "C" fn(*const i8) -> i32> = lib.get(b"JRN_Init")?;
        let open_today: Symbol<unsafe extern "C" fn() -> *const i8> = lib.get(b"JRN_OpenTodayJSON")?;

        let path = CString::new("~/.jrn")?;
        init(path.as_ptr());

        let json_ptr = open_today();
        let json_str = CStr::from_ptr(json_ptr).to_str()?;
        println!("JSON: {}", json_str);
    }
    Ok(())
}
```

---

## 🧪 Тестирование и Бенчмарки

Запуск всех тестов:
```bash
go test -v ./...
```

Запуск бенчмарков производительности:
```bash
go test -bench=. -benchmem ./internal/parser/
```

Результаты бенчмарка:
```
BenchmarkParse-16          294625       3921 ns/op     427.70 MB/s      35 allocs/op
BenchmarkParseJson-16      221528       5413 ns/op     409.77 MB/s      48 allocs/op
BenchmarkSerialize-16     2617711        458 ns/op    3657.87 MB/s       2 allocs/op
```

---

## 📄 Лицензия

MIT License.
