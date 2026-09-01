package parser

import (
	"jrn/internal/domain"
	"reflect"
	"testing"
)

// ─── Хелперы ────────────────────────────────────────────────────────────────

func mustParse(t *testing.T, input string) *domain.Document {
	t.Helper()
	doc, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse вернул ошибку: %v\nВход:\n%s", err, input)
	}
	return doc
}

func assertMeta(t *testing.T, doc *domain.Document, date string, streak, completed int) {
	t.Helper()
	if doc.Meta.Date != date {
		t.Errorf("Meta.Date = %q, хотели %q", doc.Meta.Date, date)
	}
	if doc.Meta.Streak != streak {
		t.Errorf("Meta.Streak = %d, хотели %d", doc.Meta.Streak, streak)
	}
	if doc.Meta.TaskCompleted != completed {
		t.Errorf("Meta.TaskCompleted = %d, хотели %d", doc.Meta.TaskCompleted, completed)
	}
}

func assertTaskCount(t *testing.T, doc *domain.Document, want int) {
	t.Helper()
	if got := len(doc.Tasks); got != want {
		t.Fatalf("len(Tasks) = %d, хотели %d", got, want)
	}
}

func assertTask(t *testing.T, task domain.Task, done bool, title string, tags []string, attrs []domain.Attribute, notes []string) {
	t.Helper()
	if task.Done != done {
		t.Errorf("Task.Done = %v, хотели %v (title: %q)", task.Done, done, title)
	}
	if task.Title != title {
		t.Errorf("Task.Title = %q, хотели %q", task.Title, title)
	}
	if tags == nil {
		tags = []string{}
	}
	if task.Tags == nil {
		task.Tags = []string{}
	}
	if !reflect.DeepEqual(task.Tags, tags) {
		t.Errorf("Task.Tags = %v, хотели %v (title: %q)", task.Tags, tags, title)
	}
	if attrs == nil {
		attrs = []domain.Attribute{}
	}
	if task.Attributes == nil {
		task.Attributes = []domain.Attribute{}
	}
	if !reflect.DeepEqual(task.Attributes, attrs) {
		t.Errorf("Task.Attributes = %v, хотели %v (title: %q)", task.Attributes, attrs, title)
	}
	if notes == nil {
		notes = []string{}
	}
	if task.Notes == nil {
		task.Notes = []string{}
	}
	if !reflect.DeepEqual(task.Notes, notes) {
		t.Errorf("Task.Notes = %v, хотели %v (title: %q)", task.Notes, notes, title)
	}
}

// ─── Тесты: Frontmatter ────────────────────────────────────────────────────

func TestParse_MinimalDocument(t *testing.T) {
	input := "---\ndate: 2026-08-29\nstreak: 1\ntask_completed: 0\n---\n\n### Tasks\n"
	doc := mustParse(t, input)
	assertMeta(t, doc, "2026-08-29", 1, 0)
	assertTaskCount(t, doc, 0)
}

func TestParse_FrontmatterValues(t *testing.T) {
	input := "---\ndate: 2026-01-15\nstreak: 42\ntask_completed: 7\n---\n\n### Tasks\n"
	doc := mustParse(t, input)
	assertMeta(t, doc, "2026-01-15", 42, 7)
}

func TestParse_FrontmatterOnlyDate(t *testing.T) {
	// streak и task_completed не указаны — должны быть 0
	input := "---\ndate: 2026-08-29\n---\n\n### Tasks\n"
	doc := mustParse(t, input)
	assertMeta(t, doc, "2026-08-29", 0, 0)
}

// ─── Тесты: Tasks ──────────────────────────────────────────────────────────

func TestParse_SingleTask(t *testing.T) {
	input := "---\ndate: 2026-08-29\nstreak: 1\ntask_completed: 0\n---\n\n### Tasks\n- [ ] Купить молоко\n"
	doc := mustParse(t, input)
	assertTaskCount(t, doc, 1)
	assertTask(t, doc.Tasks[0], false, "Купить молоко", nil, nil, nil)
}

func TestParse_TaskDone(t *testing.T) {
	input := "---\ndate: 2026-08-29\nstreak: 1\ntask_completed: 1\n---\n\n### Tasks\n- [x] Готово\n"
	doc := mustParse(t, input)
	assertTaskCount(t, doc, 1)
	assertTask(t, doc.Tasks[0], true, "Готово", nil, nil, nil)
}

func TestParse_MultipleTasks(t *testing.T) {
	input := "---\ndate: 2026-08-29\nstreak: 1\ntask_completed: 0\n---\n\n### Tasks\n- [ ] Первая\n- [x] Вторая\n- [ ] Третья\n"
	doc := mustParse(t, input)
	assertTaskCount(t, doc, 3)
	assertTask(t, doc.Tasks[0], false, "Первая", nil, nil, nil)
	assertTask(t, doc.Tasks[1], true, "Вторая", nil, nil, nil)
	assertTask(t, doc.Tasks[2], false, "Третья", nil, nil, nil)
}

func TestParse_TaskWithTags(t *testing.T) {
	input := "---\ndate: 2026-08-29\nstreak: 1\ntask_completed: 0\n---\n\n### Tasks\n- [ ] Пробежка #health #habit\n"
	doc := mustParse(t, input)
	assertTaskCount(t, doc, 1)
	assertTask(t, doc.Tasks[0], false, "Пробежка", []string{"health", "habit"}, nil, nil)
}

func TestParse_TaskWithAttributes(t *testing.T) {
	input := "---\ndate: 2026-08-29\nstreak: 1\ntask_completed: 0\n---\n\n### Tasks\n- [ ] Задача @prio(high) @due(2026-09-01)\n"
	doc := mustParse(t, input)
	assertTaskCount(t, doc, 1)
	assertTask(t, doc.Tasks[0], false, "Задача",
		nil,
		[]domain.Attribute{
			{Key: "prio", Value: "high"},
			{Key: "due", Value: "2026-09-01"},
		},
		nil,
	)
}

func TestParse_TaskWithTagsAndAttributes(t *testing.T) {
	input := "---\ndate: 2026-08-29\nstreak: 1\ntask_completed: 0\n---\n\n### Tasks\n- [ ] Фикс бага #dev #go @prio(high) @due(2026-08-30)\n"
	doc := mustParse(t, input)
	assertTaskCount(t, doc, 1)
	assertTask(t, doc.Tasks[0], false, "Фикс бага",
		[]string{"dev", "go"},
		[]domain.Attribute{
			{Key: "prio", Value: "high"},
			{Key: "due", Value: "2026-08-30"},
		},
		nil,
	)
}

func TestParse_TaskWithNotes(t *testing.T) {
	input := "---\ndate: 2026-08-29\nstreak: 1\ntask_completed: 0\n---\n\n### Tasks\n- [ ] Продукты #home\n  - Молоко\n  - Хлеб\n  - Яйца\n"
	doc := mustParse(t, input)
	assertTaskCount(t, doc, 1)
	assertTask(t, doc.Tasks[0], false, "Продукты",
		[]string{"home"},
		nil,
		[]string{"Молоко", "Хлеб", "Яйца"},
	)
}

func TestParse_MultipleTasksWithNotes(t *testing.T) {
	input := `---
date: 2026-08-29
streak: 1
task_completed: 0
---

### Tasks
- [ ] Первая задача #dev
  - Заметка 1
  - Заметка 2
- [x] Вторая задача @prio(low)
  - Единственная заметка
- [ ] Третья без заметок
`
	doc := mustParse(t, input)
	assertTaskCount(t, doc, 3)

	assertTask(t, doc.Tasks[0], false, "Первая задача",
		[]string{"dev"}, nil, []string{"Заметка 1", "Заметка 2"})

	assertTask(t, doc.Tasks[1], true, "Вторая задача",
		nil, []domain.Attribute{{Key: "prio", Value: "low"}}, []string{"Единственная заметка"})

	assertTask(t, doc.Tasks[2], false, "Третья без заметок", nil, nil, nil)
}

// ─── Тесты: Daily Log ──────────────────────────────────────────────────────

func TestParse_DailyLog(t *testing.T) {
	input := `---
date: 2026-08-29
streak: 1
task_completed: 0
---

### Daily Log
Утро началось с кофе.
Днем работал над парсером.

### Tasks
- [ ] Задача
`
	doc := mustParse(t, input)
	if len(doc.DailyLog) != 2 {
		t.Fatalf("len(DailyLog) = %d, хотели 2", len(doc.DailyLog))
	}
	if doc.DailyLog[0] != "Утро началось с кофе." {
		t.Errorf("DailyLog[0] = %q", doc.DailyLog[0])
	}
	if doc.DailyLog[1] != "Днем работал над парсером." {
		t.Errorf("DailyLog[1] = %q", doc.DailyLog[1])
	}
	assertTaskCount(t, doc, 1)
}

func TestParse_NoDailyLog(t *testing.T) {
	input := "---\ndate: 2026-08-29\nstreak: 1\ntask_completed: 0\n---\n\n### Tasks\n- [ ] Задача\n"
	doc := mustParse(t, input)
	if len(doc.DailyLog) != 0 {
		t.Errorf("len(DailyLog) = %d, хотели 0", len(doc.DailyLog))
	}
}

// ─── Тесты: полный документ ─────────────────────────────────────────────────

func TestParse_FullDocument(t *testing.T) {
	input := `---
date: 2026-08-28
streak: 14
task_completed: 0
---

### Daily Log
Утро началось с долгой отладки парсера.
Вечером планирую добить TUI.

### Tasks
- [ ] Утренняя пробежка 5 км #habit #health @prio(low)
- [ ] Разобрать лекцию #study @prio(medium) @due(2026-08-25)
  - Выписать основные отличия TCP от UDP
  - Понять схему трехстороннего рукопожатия
- [x] Закупить продукты #home @due(2026-08-26)
  - Овсяное молоко
  - Свежие томаты
`
	doc := mustParse(t, input)

	assertMeta(t, doc, "2026-08-28", 14, 0)

	if len(doc.DailyLog) != 2 {
		t.Fatalf("len(DailyLog) = %d, хотели 2", len(doc.DailyLog))
	}

	assertTaskCount(t, doc, 3)

	assertTask(t, doc.Tasks[0], false, "Утренняя пробежка 5 км",
		[]string{"habit", "health"},
		[]domain.Attribute{{Key: "prio", Value: "low"}},
		nil)

	assertTask(t, doc.Tasks[1], false, "Разобрать лекцию",
		[]string{"study"},
		[]domain.Attribute{
			{Key: "prio", Value: "medium"},
			{Key: "due", Value: "2026-08-25"},
		},
		[]string{"Выписать основные отличия TCP от UDP", "Понять схему трехстороннего рукопожатия"})

	assertTask(t, doc.Tasks[2], true, "Закупить продукты",
		[]string{"home"},
		[]domain.Attribute{{Key: "due", Value: "2026-08-26"}},
		[]string{"Овсяное молоко", "Свежие томаты"})
}

// ─── Тесты: Round-trip (Parse → Serialize → Parse) ──────────────────────────

func TestRoundTrip_MinimalDocument(t *testing.T) {
	original := &domain.Document{
		Meta: domain.Metadata{
			Date:          "2026-08-29",
			Streak:        5,
			TaskCompleted: 2,
		},
		DailyLog: []string{},
		Tasks: []domain.Task{
			{Done: false, Title: "Задача", Tags: []string{}, Attributes: []domain.Attribute{}, Notes: []string{}},
			{Done: true, Title: "Готово", Tags: []string{}, Attributes: []domain.Attribute{}, Notes: []string{}},
		},
	}

	serialized := Serialize(original)
	parsed, err := Parse(serialized)
	if err != nil {
		t.Fatalf("Round-trip Parse ошибка: %v\nСериализованное:\n%s", err, string(serialized))
	}

	assertMeta(t, parsed, original.Meta.Date, original.Meta.Streak, original.Meta.TaskCompleted)
	assertTaskCount(t, parsed, len(original.Tasks))

	for i, task := range original.Tasks {
		assertTask(t, parsed.Tasks[i], task.Done, task.Title, task.Tags, task.Attributes, task.Notes)
	}
}

func TestRoundTrip_FullDocument(t *testing.T) {
	original := &domain.Document{
		Meta: domain.Metadata{
			Date:          "2026-08-28",
			Streak:        14,
			TaskCompleted: 3,
		},
		DailyLog: []string{
			"Утро началось с кофе.",
			"Работал над проектом весь день.",
		},
		Tasks: []domain.Task{
			{
				Done:       false,
				Title:      "Пробежка",
				Tags:       []string{"health", "habit"},
				Attributes: []domain.Attribute{{Key: "prio", Value: "low"}},
				Notes:      []string{},
			},
			{
				Done:       true,
				Title:      "Фикс бага",
				Tags:       []string{"dev", "go"},
				Attributes: []domain.Attribute{{Key: "prio", Value: "high"}, {Key: "due", Value: "2026-08-28"}},
				Notes:      []string{"Переписал метод Save", "Добавил тесты"},
			},
			{
				Done:       false,
				Title:      "Купить продукты",
				Tags:       []string{"home"},
				Attributes: []domain.Attribute{},
				Notes:      []string{"Молоко", "Хлеб"},
			},
		},
	}

	serialized := Serialize(original)
	parsed, err := Parse(serialized)
	if err != nil {
		t.Fatalf("Round-trip Parse ошибка: %v\nСериализованное:\n%s", err, string(serialized))
	}

	assertMeta(t, parsed, original.Meta.Date, original.Meta.Streak, original.Meta.TaskCompleted)

	if len(parsed.DailyLog) != len(original.DailyLog) {
		t.Fatalf("DailyLog len = %d, хотели %d", len(parsed.DailyLog), len(original.DailyLog))
	}
	for i, line := range original.DailyLog {
		if parsed.DailyLog[i] != line {
			t.Errorf("DailyLog[%d] = %q, хотели %q", i, parsed.DailyLog[i], line)
		}
	}

	assertTaskCount(t, parsed, len(original.Tasks))
	for i, task := range original.Tasks {
		assertTask(t, parsed.Tasks[i], task.Done, task.Title, task.Tags, task.Attributes, task.Notes)
	}
}

// ─── Тесты: ошибки парсера ─────────────────────────────────────────────────

func TestParse_Errors(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"нет frontmatter", "### Tasks\n- [ ] Задача\n"},
		{"незакрытый frontmatter", "---\ndate: 2026-08-29\nstreak: 1\n"},
		{"мусор вместо frontmatter", "hello world\n"},
		{"невалидный ключ в frontmatter", "---\nfoo: bar\n---\n"},
		{"незакрытый чекбокс", "---\ndate: 2026-08-29\n---\n\n### Tasks\n- [  Задача\n"},
		{"незакрытый атрибут", "---\ndate: 2026-08-29\n---\n\n### Tasks\n- [ ] Задача @prio(high\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse([]byte(tc.input))
			if err == nil {
				t.Errorf("ожидалась ошибка для: %q", tc.name)
			}
		})
	}
}

// ─── Тесты: edge-cases ─────────────────────────────────────────────────────

func TestParse_EmptyTasksSection(t *testing.T) {
	input := "---\ndate: 2026-08-29\nstreak: 1\ntask_completed: 0\n---\n\n### Tasks\n"
	doc := mustParse(t, input)
	assertTaskCount(t, doc, 0)
}

func TestParse_TaskWithoutNewlineAtEnd(t *testing.T) {
	input := "---\ndate: 2026-08-29\nstreak: 1\ntask_completed: 0\n---\n\n### Tasks\n- [ ] Без переноса"
	doc := mustParse(t, input)
	assertTaskCount(t, doc, 1)
	assertTask(t, doc.Tasks[0], false, "Без переноса", nil, nil, nil)
}

func TestParse_TaskWithTagNoNewline(t *testing.T) {
	input := "---\ndate: 2026-08-29\nstreak: 1\ntask_completed: 0\n---\n\n### Tasks\n- [ ] Задача #tag"
	doc := mustParse(t, input)
	assertTaskCount(t, doc, 1)
	assertTask(t, doc.Tasks[0], false, "Задача", []string{"tag"}, nil, nil)
}

func TestParse_DailyLogThenTasks(t *testing.T) {
	// Порядок: Daily Log → Tasks
	input := "---\ndate: 2026-08-29\nstreak: 1\ntask_completed: 0\n---\n\n### Daily Log\nТекст записи.\n\n### Tasks\n- [ ] Задача\n"
	doc := mustParse(t, input)
	if len(doc.DailyLog) != 1 {
		t.Fatalf("len(DailyLog) = %d, хотели 1", len(doc.DailyLog))
	}
	assertTaskCount(t, doc, 1)
}

func TestParse_TasksThenDailyLog(t *testing.T) {
	// Порядок: Tasks → Daily Log
	input := "---\ndate: 2026-08-29\nstreak: 1\ntask_completed: 0\n---\n\n### Tasks\n- [ ] Задача\n\n### Daily Log\nТекст записи.\n"
	doc := mustParse(t, input)
	assertTaskCount(t, doc, 1)
	if len(doc.DailyLog) != 1 {
		t.Fatalf("len(DailyLog) = %d, хотели 1", len(doc.DailyLog))
	}
}

func TestParse_UppercaseX(t *testing.T) {
	input := "---\ndate: 2026-08-29\nstreak: 1\ntask_completed: 0\n---\n\n### Tasks\n- [X] Задача с большой X\n"
	doc := mustParse(t, input)
	assertTaskCount(t, doc, 1)
	assertTask(t, doc.Tasks[0], true, "Задача с большой X", nil, nil, nil)
}

func TestParse_LargeStreak(t *testing.T) {
	input := "---\ndate: 2026-08-29\nstreak: 365\ntask_completed: 99\n---\n\n### Tasks\n"
	doc := mustParse(t, input)
	assertMeta(t, doc, "2026-08-29", 365, 99)
}
