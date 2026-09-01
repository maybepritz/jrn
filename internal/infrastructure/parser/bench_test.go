package parser

import (
	"jrn/internal/domain"
	"testing"
)

var benchDoc = &domain.Document{
	Meta: domain.Metadata{
		Date:          "2026-08-28",
		Streak:        14,
		TaskCompleted: 3,
	},
	DailyLog: []string{
		"Утро началось с долгой отладки парсера. Пришлось полностью переосмыслить подход к работе со строками.",
		"Днем были пары по сетям — разобрали работу протоколов транспортного уровня.",
		"Вечером планирую добить TUI-интерфейс на Bubbletea.",
	},
	Tasks: []domain.Task{
		{Done: false, Title: "Утренняя пробежка 5 км", Tags: []string{"habit", "health"}, Attributes: []domain.Attribute{{Key: "prio", Value: "low"}}, Notes: []string{}},
		{Done: false, Title: "Разобрать лекцию по компьютерным сетям", Tags: []string{"study"}, Attributes: []domain.Attribute{{Key: "prio", Value: "medium"}, {Key: "due", Value: "2026-08-25"}}, Notes: []string{"Выписать основные отличия TCP от UDP", "Понять схему трехстороннего рукопожатия"}},
		{Done: false, Title: "Пофиксить панику с индексами", Tags: []string{"dev", "go"}, Attributes: []domain.Attribute{{Key: "prio", Value: "high"}, {Key: "due", Value: "2026-08-25"}}, Notes: []string{"Отказаться от ручного управления RawLines", "Переписать метод Save"}},
		{Done: false, Title: "Реализовать интерфейс TUI", Tags: []string{"dev", "go"}, Attributes: []domain.Attribute{{Key: "prio", Value: "high"}, {Key: "due", Value: "2026-08-26"}}, Notes: []string{"Подключить обработку нажатий стрелок", "Добавить модальное окно", "Вывести теги и дедлайны"}},
		{Done: true, Title: "Закупить продукты на неделю", Tags: []string{"home"}, Attributes: []domain.Attribute{{Key: "due", Value: "2026-08-26"}}, Notes: []string{"Овсяное молоко", "Свежие томаты", "Куриное филе", "Зерновой кофе"}},
		{Done: false, Title: "Сделать резервную копию", Tags: []string{"backup"}, Attributes: []domain.Attribute{{Key: "prio", Value: "low"}}, Notes: []string{}},
		{Done: false, Title: "Почитать доку по архитектуре CLI", Tags: []string{"reading"}, Attributes: []domain.Attribute{{Key: "prio", Value: "medium"}, {Key: "due", Value: "2026-08-28"}}, Notes: []string{}},
	},
}

var benchData = Serialize(benchDoc)

func BenchmarkParse(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(benchData)))

	for b.Loop() {
		_, err := Parse(benchData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSerialize(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		_ = Serialize(benchDoc)
	}
}

func BenchmarkRoundTrip(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		data := Serialize(benchDoc)
		_, err := Parse(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
