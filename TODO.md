# 📋 TODO: JRN Core Library

## ✅ 1. Завершённая база ядра (Core Foundation)
- [x] **Сериализация и парсинг:**
  - `json:` теги ко всем структурам в `pkg/domain/types.go`
  - FSM Markdown парсер и сериализатор в `internal/parser`
  - JSON парсер на базе `fastjson` с пулом парсеров (`ParseJson` / `SerializeJson`)
  - Оптимизация аллокаций памяти (`String Arena`, `String Interning`, `Pre-allocation`)
- [x] **Доменный слой (`pkg/domain`):**
  - `TaskBuilder` (`NewTask("...").Tag("...").Attr("...").Note("...").Done(true).Build()`)
  - `AddTask`, `ToggleTask` (с пересчетом `TaskCompleted`), `DeleteTask` (с очисткой GC), `UpdateTaskTitle`, `ReorderTasks`
  - Заметки (`AddNoteToTask`, `DeleteNoteFromTask`) и дневник дня (`AddDailyLog`, `DeleteDailyLog`)
  - Локальный поиск: `SearchTasks`, `TasksByTag`, `TasksByAttr`, `OverdueTasks`
  - 100% покрытие тестами в `pkg/domain/document_test.go`
- [x] **Хранилище (`pkg/storage`):**
  - Файловая система `FS` с поддержкой тильды `~` (`os.UserHomeDir`)
  - In-memory `Mock` для тестов
  - Синхронизация `ListDates(ctx)`
- [x] **Сервисный слой (`pkg/engine`):**
  - `OpenToday(ctx)` / `OpenDay(ctx, date)` с защитой от будущих дат (`ErrDateInFuture`)
  - Расчет серии `Streak` (с проверкой непрерывности `isYesterday`)
  - Автоматический перенос незавершенных задач `Rollover` (`!task.Done`)
  - Навигация: `OpenPreviousDay`, `OpenNextDay`, `ListAvailableDays`, `SaveDay`
  - Глобальный поиск: `Search`, `GetTasksByTag`, `GetOverdueTasks`
  - Тесты в `pkg/engine/engine_test.go`
- [x] **C-Shared DLL (`cmd/jrn-c`):**
  - Экспорт полного C ABI (`JRN_Init`, `JRN_OpenTodayJSON`, `JRN_AddTask`, `JRN_ToggleTask`, `JRN_SearchJSON` и др.) для Python, Java, C#, Rust.

---

## 🚀 2. Новые фичи: Расписание, Повторяющиеся задачи и Учёт времени

### А. Повторяющиеся задачи и привычки (`@repeat`)
- [ ] Добавить парсинг правил повторения в `pkg/domain`:
  - `@repeat(daily)` — каждый день
  - `@repeat(weekly:mon,wed,fri)` — по дням недели
  - `@repeat(biweekly:even:tue)` — чётная/нечётная неделя (числитель / знаменатель)
- [ ] Интегрировать генерацию повторяющихся задач в `initNewDay`:
  - Если задача была выполнена в прошлый день и содержит `@repeat`, автоматически пересоздавать её на сегодня в статусе `[ ]`
- [ ] Метод `doc.RecurringTasks() []Task` — выборка всех повторяющихся задач/привычек

---

### Б. Тайм-слоты и расписание на день (`@time`)
- [ ] Добавить парсинг атрибута `@time(09:00-10:30)` или `@time(14:00)`:
  - Выделение времени начала и окончания (`StartTime`, `EndTime`)
- [ ] Реализовать метод `doc.Schedule() []Task`:
  - Сортировка задач на день в строгом хронологическом порядке времени начала
- [ ] Реализовать метод `doc.CurrentTask(now time.Time) (*Task, bool)`:
  - Определение текущей пары/задачи прямо сейчас по системному времени

---

### В. Трекинг времени и оценка нагрузки (`@est`, `@spent`)
- [ ] Парсер строковых интервалов времени (например `"1h 30m"`, `"45m"`, `"2h"`):
  - Хелпер `parseDuration(str string) (time.Duration, error)`
- [ ] Методы расчета нагрузки в `pkg/domain/document.go`:
  - `doc.TotalEstimatedTime() time.Duration` — суммарная плановая нагрузка на день
  - `doc.TotalSpentTime() time.Duration` — суммарное фактически затраченное время
- [ ] Метод для интеграции с Pomodoro/таймером фронтенда:
  - `doc.AddSpentTime(taskIndex int, d time.Duration) error` — добавление/обновление времени задачи

---

### Г. Пакетный импорт расписания из ИИ (Bulk JSON Import)
- [ ] Метод `engine.BulkImportScheduleJSON(ctx, jsonBytes []byte) error`:
  - Приём сгенерированного нейросетью JSON с расписанием на семестр
  - Автоматическая раскладка по соответствующим календарным датам `.md` файлов

---

### Д. Тестирование
- [ ] Написать юнит-тесты для `@repeat` и генерации дней
- [ ] Написать юнит-тесты для `@time` и `doc.Schedule()`
- [ ] Написать юнит-тесты для `@est`/`@spent` и `TotalSpentTime()`
