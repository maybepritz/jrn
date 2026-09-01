# 🤖 AGENTS.md — Developer & AI Agent Reference Manual

> **Purpose:** This document provides essential architectural context, invariants, design decisions, and guidelines for AI Coding Agents (Claude, GPT, Gemini, Cursor, Windsurf, Devin, etc.) working on or integrating with the **JRN Core** codebase.

---

## 🏗️ 1. Project Overview & Architecture

JRN Core is a high-performance, embedded Go library/SDK for daily journaling, task tracking, productivity statistics (`Streak`), and automatic task rollover.

### Module Boundaries

```
task-cli-go/
├── pkg/                             # 🟢 PUBLIC API (Importable by external Go consumers)
│   ├── domain/                      # Core models (Task, Document, Attribute, Metadata), TaskBuilder, Storage interface, errors
│   ├── engine/                      # Service layer (Engine, OpenToday, OpenDay, Search, SaveDay)
│   └── storage/                     # Storage implementations: FS (with ~ resolution) & in-memory Mock
│
├── internal/                        # 🔒 PRIVATE IMPLEMENTATION (Never importable outside)
│   └── parser/                      # Zero-alloc FSM Markdown parser & fastjson pool serializer
│
└── cmd/
    └── jrn-c/                       # CGO C-Shared entrypoint exporting C ABI for non-Go consumers
```

---

## 🛡️ 2. Critical Domain Invariants (NEVER VIOLATE)

When modifying or extending the core, all agents **MUST** preserve these invariants:

1. **Task Title Invariant:**
   * A task title must never be empty or contain only whitespace.
   * Always validate with `strings.TrimSpace(title) == ""` and return `domain.ErrEmptyTitle`.
   * Always store sanitized/trimmed titles (`task.Title = trimmed`).

2. **`TaskCompleted` Statistical Invariant:**
   * `doc.Meta.TaskCompleted` must accurately reflect the number of tasks with `Done == true`.
   * In `ToggleTask(index)`: increment when toggling `false -> true`, decrement when `true -> false`.
   * In `DeleteTask(index)`: if `task.Done == true`, decrement `TaskCompleted`.
   * In `AddTask(task)`: if `task.Done == true`, increment `TaskCompleted`.

3. **Streak Calculation & Time Invariant:**
   * A streak increments (`doc.Meta.Streak = prevStreak + 1`) **ONLY IF**:
     1. `prevTasksCount > 0` (there was at least 1 task on the previous day),
     2. `prevCompleted == prevTasksCount` (all tasks were completed),
     3. `isYesterday == true` (`prevDate.Format("2006-01-02") == date.AddDate(0, 0, -1).Format("2006-01-02")`).
   * Never use direct `time.Time == time.Time` comparison due to monotonic/wall clock and timezone offsets. Always compare calendar strings or `y1 == y2 && m1 == m2 && d1 == d2`.

4. **Rollover Deep Copy Invariant:**
   * When initializing a new day from a previous day (`initNewDay`), uncompleted tasks (`!task.Done`) must be **deep-copied** (copying slices `Tags`, `Attributes`, and `Notes`). Never shallow copy slice references.

5. **Memory Safety & GC Slicing:**
   * When deleting items from slices in Go (`Tasks`, `Notes`, `DailyLog`), always use `slices.Delete(slice, i, i+1)` or explicitly clear the dangling pointer in the backing array to avoid memory leaks.

---

## ⚡ 3. Performance & Parser Guidelines

* The Markdown parser in `internal/parser/parser.go` is an ultra-fast Finite State Machine (FSM).
* It uses **String Arena Substring Indexing** over a single source string and **String Interning** for common tags (`dev`, `go`, `health`, etc.) and attribute keys/values (`prio`, `due`, `high`).
* Do not introduce regex or heavy string allocations in `internal/parser/`.
* Benchmark target: `< 4.0 µs/op` and `< 40 allocs/op` for full markdown parsing.

---

## 🧪 4. Commands for Verification

Agents must run these commands before finalizing any code changes:

### Run All Unit Tests:
```bash
go test -v ./...
```

### Run Performance Benchmarks:
```bash
go test -bench=. -benchmem ./internal/parser/
```

### Build C-Shared DLL:
```bash
# Windows (Requires GCC / MinGW or Zig)
$env:CGO_ENABLED="1"
go build -buildmode=c-shared -o jrn.dll ./cmd/jrn-c
```

---

## 🔌 5. C ABI Functions (for Non-Go Bindings)

`cmd/jrn-c/main.go` exports the following C functions:
* `int JRN_Init(char* path)` — Returns `0` on success, `-1` on failure.
* `char* JRN_OpenTodayJSON()` — Returns heap-allocated JSON string (must free with `JRN_FreeString`).
* `char* JRN_OpenDayJSON(char* dateStr)` — Opens specific date `YYYY-MM-DD`.
* `int JRN_AddTask(char* dateStr, char* taskJSON)` — Adds task.
* `int JRN_ToggleTask(char* dateStr, int index)` — Toggles task state.
* `int JRN_DeleteTask(char* dateStr, int index)` — Deletes task.
* `int JRN_UpdateTaskTitle(char* dateStr, int index, char* title)` — Renames task.
* `int JRN_ReorderTasks(char* dateStr, int from, int to)` — Moves task.
* `int JRN_AddDailyLog(char* dateStr, char* text)` — Appends daily log note.
* `int JRN_DeleteDailyLog(char* dateStr, int index)` — Deletes daily log note.
* `char* JRN_SearchJSON(char* query)` — Global search across all entries.
* `char* JRN_GetTasksByTagJSON(char* tag)` — Filter by tag across archive.
* `char* JRN_GetOverdueTasksJSON(char* todayStr)` — Get overdue tasks.
* `char* JRN_ListAvailableDaysJSON()` — List all recorded dates.
* `void JRN_FreeString(char* str)` — Frees C strings returned by the engine.
