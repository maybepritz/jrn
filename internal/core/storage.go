package core

import (
	"fmt"
	"os"
	"strings"
)

func (doc *DayDocument) Save() error {
	var sb strings.Builder

	// Записываем фронтматтер
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("date: %s\n", doc.Meta.Date))
	sb.WriteString(fmt.Sprintf("streak: %d\n", doc.Meta.Streak))
	sb.WriteString(fmt.Sprintf("total_completed: %d\n", doc.Meta.TotalCompleted))
	sb.WriteString("---\n\n")

	// Записываем дейлик
	if len(doc.DailyLog) > 0 {
		sb.WriteString("## Daily Log\n")
		for _, line := range doc.DailyLog {
			sb.WriteString(line + "\n")
		}
		sb.WriteString("\n")
	}

	// Записываем таски
	if len(doc.Tasks) > 0 {
		sb.WriteString("## Tasks\n")
		for _, task := range doc.Tasks {
			status := "[ ]"
			if task.Done {
				status = "[x]"
			}
			taskLine := fmt.Sprintf("- %s %s", status, task.Title)
			if len(task.Tags) > 0 {
				var tagStrs []string
				for _, tag := range task.Tags {
					tagStrs = append(tagStrs, "#"+tag)
				}
				taskLine += " " + strings.Join(tagStrs, " ")
			}

			switch task.Priority {
			case PriorityLow:
				taskLine += " @prio(low)"
			case PriorityMedium:
				taskLine += " @prio(medium)"
			case PriorityHigh:
				taskLine += " @prio(high)"
			}

			if !task.DueTime.IsZero() {
				taskLine += fmt.Sprintf(" @due(%s)", task.DueTime.Format("2006-01-02"))
			}

			sb.WriteString(taskLine + "\n")

			for _, note := range task.Notes {
				sb.WriteString(fmt.Sprintf("  - %s\n", note))
			}
		}
	}

	return os.WriteFile(doc.FilePath, []byte(sb.String()), 0644)
}
