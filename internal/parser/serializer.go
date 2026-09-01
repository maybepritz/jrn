package parser

import (
	"bytes"
	"jrn/pkg/domain"
	"strconv"
)

func Serialize(doc *domain.Document) []byte {
	var buf bytes.Buffer
	buf.Grow(1024)

	// фронтматтер
	buf.WriteString("---\ndate: ")
	buf.WriteString(doc.Meta.Date)
	buf.WriteString("\nstreak: ")
	buf.WriteString(strconv.Itoa(doc.Meta.Streak))
	buf.WriteString("\ntask_completed: ")
	buf.WriteString(strconv.Itoa(doc.Meta.TaskCompleted))
	buf.WriteString("\n---\n\n")

	//дейлик
	if len(doc.DailyLog) > 0 {
		buf.WriteString("### Daily Log\n")
		for _, line := range doc.DailyLog {
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
		buf.WriteByte('\n')
	}

	//таски
	buf.WriteString("### Tasks\n")
	for i := range doc.Tasks {
		t := &doc.Tasks[i]

		if t.Done {
			buf.WriteString("- [x] ")
		} else {
			buf.WriteString("- [ ] ")
		}

		buf.WriteString(t.Title)

		for _, tag := range t.Tags {
			buf.WriteString(" #")
			buf.WriteString(tag)
		}

		for _, attr := range t.Attributes {
			buf.WriteString(" @")
			buf.WriteString(attr.Key)
			buf.WriteByte('(')
			buf.WriteString(attr.Value)
			buf.WriteByte(')')
		}

		buf.WriteByte('\n')

		for _, note := range t.Notes {
			buf.WriteString("  - ")
			buf.WriteString(note)
			buf.WriteByte('\n')
		}

	}
	return buf.Bytes()
}

func SerializeJson(doc *domain.Document) []byte {
	if doc == nil {
		return []byte("{}")
	}

	var buf bytes.Buffer
	buf.Grow(1024)

	// meta
	buf.WriteString(`{"meta":{"date":"`)
	buf.WriteString(doc.Meta.Date)
	buf.WriteString(`","streak":`)
	buf.WriteString(strconv.Itoa(doc.Meta.Streak))
	buf.WriteString(`,"task_completed":`)
	buf.WriteString(strconv.Itoa(doc.Meta.TaskCompleted))
	buf.WriteString(`},"daily_log":[`)

	// daily_log
	for i, log := range doc.DailyLog {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteByte('"')
		buf.WriteString(log)
		buf.WriteByte('"')
	}

	// tasks
	buf.WriteString(`],"tasks":[`)
	for i, task := range doc.Tasks {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(`{"done":`)
		if task.Done {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
		buf.WriteString(`,"title":"`)
		buf.WriteString(task.Title)
		buf.WriteByte('"')

		// tags
		buf.WriteString(`,"tags":[`)
		for j, tag := range task.Tags {
			if j > 0 {
				buf.WriteByte(',')
			}
			buf.WriteByte('"')
			buf.WriteString(tag)
			buf.WriteByte('"')
		}
		buf.WriteByte(']')

		// attributes
		buf.WriteString(`,"attributes":[`)
		for j, attr := range task.Attributes {
			if j > 0 {
				buf.WriteByte(',')
			}
			buf.WriteString(`{"key":"`)
			buf.WriteString(attr.Key)
			buf.WriteString(`","value":"`)
			buf.WriteString(attr.Value)
			buf.WriteString(`"}`)
		}
		buf.WriteByte(']')

		// notes
		buf.WriteString(`,"notes":[`)
		for j, note := range task.Notes {
			if j > 0 {
				buf.WriteByte(',')
			}
			buf.WriteByte('"')
			buf.WriteString(note)
			buf.WriteByte('"')
		}
		buf.WriteByte(']')

		buf.WriteByte('}')
	}
	buf.WriteString(`]}`)

	return buf.Bytes()
}

