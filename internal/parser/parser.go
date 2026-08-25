package parser

import (
	"bufio"
	"jrn/internal/core"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	StateFrontmatter  = iota // шапку
	StateMetadataDone        // ищем секции
	StateDailyLog            // дейлик
	StateTaskList            // список задач
)

func ParseMarkdown(filePath string) (*core.DayDocument, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	doc := &core.DayDocument{FilePath: filePath}
	scanner := bufio.NewScanner(file)
	state := StateFrontmatter

	isOpen := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		switch state {
		case StateFrontmatter:
			if trimmed == "---" {
				if !isOpen {
					isOpen = true
					continue
				} else {
					state = StateMetadataDone
					continue
				}
			}
			if isOpen {
				parseFrontmatterKeyValue(trimmed, &doc.Meta)
			}

		case StateMetadataDone:
			if strings.EqualFold(trimmed, "## Daily Log") {
				state = StateDailyLog
				continue
			}

			if strings.EqualFold(trimmed, "## Tasks") {
				state = StateTaskList
				continue
			}

		case StateDailyLog:
			if strings.HasPrefix(trimmed, "## ") {
				if strings.EqualFold(trimmed, "## Tasks") {
					state = StateTaskList
				}
				continue
			}

			if trimmed != "" {
				doc.DailyLog = append(doc.DailyLog, trimmed)
			}

		case StateTaskList:
			if strings.HasPrefix(trimmed, "- [ ]") || strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") {
				done := trimmed[3] == 'x' || trimmed[3] == 'X'
				body := strings.TrimSpace(trimmed[5:])
				task := core.Task{
					Done: done,
				}
				extractInlineTokens(body, &task)
				doc.Tasks = append(doc.Tasks, task)
				continue
			}

			if len(doc.Tasks) > 0 && strings.HasPrefix(line, "  -") {
				if trimmed != "" {
					noteText := strings.TrimPrefix(trimmed, "- ")
					lastIdx := len(doc.Tasks) - 1
					doc.Tasks[lastIdx].Notes = append(doc.Tasks[lastIdx].Notes, noteText)
				}
			}
		}
	}

	return doc, scanner.Err()
}

func parseFrontmatterKeyValue(line string, meta *core.Metadata) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	switch key {
	case "date":
		meta.Date = value
	case "streak":
		if streak, err := strconv.Atoi(value); err == nil {
			meta.Streak = streak
		}
	case "total_completed":
		if total, err := strconv.Atoi(value); err == nil {
			meta.TotalCompleted = total
		}
	}
}

func extractInlineTokens(body string, task *core.Task) {
	words := strings.Fields(body)
	var titleWords []string

	for _, word := range words {
		//теги #work #personal
		if strings.HasPrefix(word, "#") {
			task.Tags = append(task.Tags, word[1:])
		}

		//приоритет @low @medium @high
		if strings.HasPrefix(word, "@prio(") && strings.HasSuffix(word, ")") {
			prioStr := word[6 : len(word)-1]

			switch prioStr {
			case "low", "l":
				task.Priority = core.PriorityLow
			case "medium", "med", "m":
				task.Priority = core.PriorityMedium
			case "high", "h":
				task.Priority = core.PriorityHigh
			}
			continue
		}

		//дедлайн @due(2026-08-25)
		if strings.HasPrefix(word, "@due(") && strings.HasSuffix(word, ")") {
			dateStr := word[5 : len(word)-1]
			if t, err := time.Parse("2006-01-02", dateStr); err == nil {
				task.DueTime = t // Присваивание по значению
			}
		}

		if !strings.HasPrefix(word, "@") && !strings.HasPrefix(word, "#") {
			titleWords = append(titleWords, word)
		}

	}

	task.Title = strings.Join(titleWords, " ")
}
