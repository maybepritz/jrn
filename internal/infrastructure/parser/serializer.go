package parser

import (
	"bytes"
	"jrn/internal/domain"
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
