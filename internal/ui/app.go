package ui

import (
	"fmt"
	"strings"
	"time"

	"jrn/internal/core"
	"jrn/internal/parser"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Палитра Tokyo Night
var (
	colorBorderDim   = lipgloss.Color("#2f354a")
	colorBorderFocus = lipgloss.Color("#7aa2f7")

	colorTextPrimary = lipgloss.Color("#c0caf5")
	colorTextMuted   = lipgloss.Color("#565f89")
	colorAccent      = lipgloss.Color("#7dcfff")
	colorGreen       = lipgloss.Color("#9ece6a")
	colorYellow      = lipgloss.Color("#e0af68")
	colorMagenta     = lipgloss.Color("#bb9af7")
	colorBgTag       = lipgloss.Color("#1f2335")

	styleCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorderDim).
			Padding(0, 1)

	styleCardFocus = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorderFocus).
			Padding(0, 1)

	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent)

	styleTagPill = lipgloss.NewStyle().
			Foreground(colorMagenta).
			Background(colorBgTag).
			Padding(0, 1)

	styleActiveRow = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#282e44"))
)

type paneType int

const (
	paneTasks paneType = iota
	paneInspector
	paneCalendar
	paneTags
)

type focusField int

const (
	fieldTitle focusField = iota
	fieldTag
	fieldDue
)

type Model struct {
	Doc        *core.DayDocument
	Cursor     int
	ActivePane paneType
	Width      int
	Height     int

	TagCursor   int
	SelectedTag string

	InputMode  bool
	FocusedIdx focusField
	Inputs     []textinput.Model

	NoteMode  bool
	NoteInput textinput.Model

	Cal CalendarModel
}

func InitialModel(doc *core.DayDocument) Model {
	inputs := make([]textinput.Model, 3)
	inputs[fieldTitle] = textinput.New()
	inputs[fieldTitle].Placeholder = "Название задачи..."
	inputs[fieldTitle].Focus()

	inputs[fieldTag] = textinput.New()
	inputs[fieldTag].Placeholder = "study dev"

	inputs[fieldDue] = textinput.New()
	inputs[fieldDue].Placeholder = "2026-08-25"

	noteIn := textinput.New()
	noteIn.Placeholder = "Заметка к задаче..."
	noteIn.Focus()

	parsedDate, err := time.Parse("2006-01-02", doc.Meta.Date)
	if err != nil {
		parsedDate = time.Now()
	}

	return Model{
		Doc:         doc,
		Cursor:      0,
		ActivePane:  paneTasks,
		TagCursor:   0,
		SelectedTag: "all",
		Inputs:      inputs,
		FocusedIdx:  fieldTitle,
		NoteInput:   noteIn,
		Cal:         NewCalendar(parsedDate),
		Width:       110,
		Height:      32,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.NoteMode {
			switch msg.Type {
			case tea.KeyEsc:
				m.NoteMode = false
				m.NoteInput.Reset()
				return m, nil
			case tea.KeyEnter:
				val := strings.TrimSpace(m.NoteInput.Value())
				if val != "" && len(m.filteredTasks()) > 0 {
					filtered := m.filteredTasks()
					taskIdx := filtered[m.Cursor].origIdx
					m.Doc.Tasks[taskIdx].Notes = append(m.Doc.Tasks[taskIdx].Notes, val)
					_ = m.Doc.Save()
				}
				m.NoteMode = false
				m.NoteInput.Reset()
				return m, nil
			}
			var cmd tea.Cmd
			m.NoteInput, cmd = m.NoteInput.Update(msg)
			return m, cmd
		}

		if m.InputMode {
			switch msg.Type {
			case tea.KeyEsc:
				m.InputMode = false
				m.resetInputs()
				return m, nil
			case tea.KeyTab, tea.KeyShiftTab:
				if msg.Type == tea.KeyTab {
					m.FocusedIdx = (m.FocusedIdx + 1) % 3
				} else {
					m.FocusedIdx = (m.FocusedIdx + 2) % 3
				}
				for i := range m.Inputs {
					if focusField(i) == m.FocusedIdx {
						m.Inputs[i].Focus()
					} else {
						m.Inputs[i].Blur()
					}
				}
				return m, nil
			case tea.KeyEnter:
				title := strings.TrimSpace(m.Inputs[fieldTitle].Value())
				if title != "" {
					newTask := core.Task{Title: title, Done: false}
					rawTag := strings.TrimSpace(m.Inputs[fieldTag].Value())
					if rawTag != "" {
						for _, tag := range strings.Fields(rawTag) {
							newTask.Tags = append(newTask.Tags, strings.TrimPrefix(tag, "#"))
						}
					}
					rawDue := strings.TrimSpace(m.Inputs[fieldDue].Value())
					if t, err := time.Parse("2006-01-02", rawDue); err == nil {
						newTask.DueTime = t
					}
					_ = m.Doc.AddTask(newTask)
				}
				m.InputMode = false
				m.resetInputs()
				return m, nil
			}
			var cmd tea.Cmd
			m.Inputs[m.FocusedIdx], cmd = m.Inputs[m.FocusedIdx].Update(msg)
			return m, cmd
		}

		// Быстрый выбор окон через [1-4] и Tab
		switch msg.String() {
		case "1":
			m.ActivePane = paneTasks
			m.syncCalendarFocus()
			return m, nil
		case "2":
			m.ActivePane = paneInspector
			m.syncCalendarFocus()
			return m, nil
		case "3":
			m.ActivePane = paneCalendar
			m.syncCalendarFocus()
			return m, nil
		case "4":
			m.ActivePane = paneTags
			m.syncCalendarFocus()
			return m, nil
		case "tab":
			m.ActivePane = (m.ActivePane + 1) % 4
			m.syncCalendarFocus()
			return m, nil
		case "shift+tab":
			m.ActivePane = (m.ActivePane + 3) % 4
			m.syncCalendarFocus()
			return m, nil
		case "q", "ctrl+c":
			return m, tea.Quit
		}

		switch m.ActivePane {
		case paneTasks:
			tasks := m.filteredTasks()
			switch msg.String() {
			case "up", "k":
				if m.Cursor > 0 {
					m.Cursor--
				}
			case "down", "j":
				if m.Cursor < len(tasks)-1 {
					m.Cursor++
				}
			case " ":
				if len(tasks) > 0 {
					origIdx := tasks[m.Cursor].origIdx
					_ = m.Doc.ToggleTaskStatus(origIdx)
				}
			case "d":
				if len(tasks) > 0 {
					origIdx := tasks[m.Cursor].origIdx
					_ = m.Doc.RemoveTask(origIdx)
					if m.Cursor >= len(m.filteredTasks()) && m.Cursor > 0 {
						m.Cursor--
					}
				}
			case "a":
				m.InputMode = true
				m.FocusedIdx = fieldTitle
				m.Inputs[fieldTitle].Focus()
				return m, textinput.Blink
			case "n":
				if len(tasks) > 0 {
					m.NoteMode = true
					m.NoteInput.Focus()
					return m, textinput.Blink
				}
			}

		case paneTags:
			tagList := m.getAllTagsList()
			switch msg.String() {
			case "left", "h":
				if m.TagCursor > 0 {
					m.TagCursor--
				}
			case "right", "l":
				if m.TagCursor < len(tagList)-1 {
					m.TagCursor++
				}
			case "enter", " ":
				if len(tagList) > 0 {
					m.SelectedTag = tagList[m.TagCursor].name
					m.Cursor = 0
				}
			}

		case paneCalendar:
			newDate, changed := m.Cal.Update(msg)
			if changed {
				m.loadDayFile(newDate.Format("2006-01-02"))
			}
		}
	}

	return m, nil
}

func (m *Model) syncCalendarFocus() {
	m.Cal.Focused = (m.ActivePane == paneCalendar)
}

func (m *Model) loadDayFile(dateStr string) {
	fileName := dateStr + ".md"
	doc, err := parser.ParseMarkdown(fileName)
	if err != nil {
		doc = &core.DayDocument{
			FilePath: fileName,
			Meta: core.Metadata{
				Date:   dateStr,
				Streak: m.Doc.Meta.Streak,
			},
			Tasks:    []core.Task{},
			DailyLog: []string{},
		}
	}
	m.Doc = doc
	m.Cursor = 0
	m.SelectedTag = "all"
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err == nil {
		m.Cal.SetDate(parsedDate)
	}
}

func (m *Model) resetInputs() {
	for i := range m.Inputs {
		m.Inputs[i].Reset()
		m.Inputs[i].Blur()
	}
}

type filteredTaskItem struct {
	task    core.Task
	origIdx int
}

func (m Model) filteredTasks() []filteredTaskItem {
	var list []filteredTaskItem
	for idx, t := range m.Doc.Tasks {
		if m.SelectedTag == "all" {
			list = append(list, filteredTaskItem{task: t, origIdx: idx})
			continue
		}
		for _, tag := range t.Tags {
			if strings.EqualFold(tag, m.SelectedTag) {
				list = append(list, filteredTaskItem{task: t, origIdx: idx})
				break
			}
		}
	}
	return list
}

type tagItem struct {
	name  string
	count int
}

func (m Model) getAllTagsList() []tagItem {
	counts := make(map[string]int)
	for _, t := range m.Doc.Tasks {
		for _, tag := range t.Tags {
			counts[tag]++
		}
	}
	list := []tagItem{{name: "all", count: len(m.Doc.Tasks)}}
	for k, v := range counts {
		list = append(list, tagItem{name: k, count: v})
	}
	return list
}

func (m Model) View() string {
	if m.InputMode {
		return m.renderModal()
	}
	if m.NoteMode {
		return m.renderNoteModal()
	}

	totalW := m.Width - 4
	if totalW < 88 {
		totalW = 88
	}
	if totalW > 120 {
		totalW = 120
	}

	totalH := m.Height - 3
	if totalH < 24 {
		totalH = 24
	}

	leftW := (totalW / 2) - 1
	rightW := totalW - leftW - 2

	// 1. Верхний ярус (H=8): HUD метрик + Календарь
	topH := 8
	topCard1 := m.renderDashboardCard(leftW, topH)
	topCard2 := m.Cal.View(rightW, topH)
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, topCard1, topCard2)

	// 2. Центральный ярус: Задачи + Инспектор
	middleH := totalH - topH - 6
	if middleH < 9 {
		middleH = 9
	}

	leftCol := m.renderTaskList(leftW, middleH)
	rightCol := m.renderInspector(rightW, middleH)
	middleRow := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)

	// 3. Нижний ярус: Лента тегов
	tagsBar := m.renderTagRibbon(totalW)

	// 4. Статус-бар
	helpBar := m.renderHelpBar()

	uiContent := lipgloss.JoinVertical(
		lipgloss.Left,
		topRow,
		middleRow,
		tagsBar,
		helpBar,
	)

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, uiContent)
}

func (m Model) renderDashboardCard(width, height int) string {
	box := styleCard

	title := styleTitle.Render("⚡ DEVJOURNAL")
	datePart := lipgloss.NewStyle().Foreground(colorTextMuted).Render("│ " + m.Doc.Meta.Date)
	topLine := lipgloss.JoinHorizontal(lipgloss.Left, title, " ", datePart)

	streakBadge := lipgloss.NewStyle().
		Background(colorYellow).
		Foreground(lipgloss.Color("#1a1b26")).
		Bold(true).
		Padding(0, 1).
		Render(fmt.Sprintf("🔥 %dD", m.Doc.Meta.Streak))

	totalTasks := len(m.Doc.Tasks)
	completedTasks := 0
	for _, t := range m.Doc.Tasks {
		if t.Done {
			completedTasks++
		}
	}
	doneBadge := lipgloss.NewStyle().
		Background(colorGreen).
		Foreground(lipgloss.Color("#1a1b26")).
		Bold(true).
		Padding(0, 1).
		Render(fmt.Sprintf("✓ %d/%d", completedTasks, totalTasks))

	badgeLine := lipgloss.JoinHorizontal(lipgloss.Left, streakBadge, "  ", doneBadge)

	percent := 0
	if totalTasks > 0 {
		percent = (completedTasks * 100) / totalTasks
	}

	barW := width - 18
	if barW < 6 {
		barW = 6
	}
	filled := (percent * barW) / 100
	if filled > barW {
		filled = barW
	}
	bar := lipgloss.NewStyle().Foreground(colorGreen).Render(strings.Repeat("▰", filled)) +
		lipgloss.NewStyle().Foreground(colorBorderDim).Render(strings.Repeat("▱", barW-filled))
	progressLine := fmt.Sprintf("Progress: [%s] %2d%%", bar, percent)

	h1 := lipgloss.NewStyle().Foreground(colorTextMuted).Render("Mon ") + lipgloss.NewStyle().Foreground(colorGreen).Render("░ ░ ▒ █ ░ ▒ █ ░ █ ▒")
	h2 := lipgloss.NewStyle().Foreground(colorTextMuted).Render("Wed ") + lipgloss.NewStyle().Foreground(colorGreen).Render("░ ▒ █ ░ ░ █ ▒ ░ ░ ▒")

	var rows []string
	rows = append(rows, topLine, badgeLine, progressLine, "", h1, h2)

	innerH := height - 2
	for len(rows) < innerH {
		rows = append(rows, "")
	}
	if len(rows) > innerH {
		rows = rows[:innerH]
	}

	return box.Width(width).Height(height).Render(strings.Join(rows, "\n"))
}

func (m Model) renderTaskList(width, height int) string {
	box := styleCard
	titleText := "[1] TODAY'S TASKS"
	if m.ActivePane == paneTasks {
		box = styleCardFocus
		titleText = "▶ [1] TODAY'S TASKS"
	}

	tasks := m.filteredTasks()
	title := styleTitle.Render(fmt.Sprintf("%s (%d)", titleText, len(tasks)))
	lines := []string{title, ""}

	if len(tasks) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorTextMuted).Italic(true).Render("(нет подходящих задач)"))
	}

	contentW := width - 4
	for i, it := range tasks {
		t := it.task
		cursor := "  "
		if i == m.Cursor && m.ActivePane == paneTasks {
			cursor = "❯ "
		}

		status := lipgloss.NewStyle().Foreground(colorTextMuted).Render("○ ")
		if t.Done {
			status = lipgloss.NewStyle().Foreground(colorGreen).Render("● ")
		}

		tagStr := ""
		if len(t.Tags) > 0 {
			tagStr = " " + styleTagPill.Render("#"+t.Tags[0])
		}

		rowRaw := fmt.Sprintf("%s%s%s%s", cursor, status, t.Title, tagStr)
		rowStyle := lipgloss.NewStyle().MaxWidth(contentW).Inline(true)
		if t.Done {
			rowStyle = rowStyle.Foreground(colorTextMuted).Strikethrough(true)
		} else {
			rowStyle = rowStyle.Foreground(colorTextPrimary)
		}

		rowRendered := rowStyle.Render(rowRaw)
		if i == m.Cursor && m.ActivePane == paneTasks {
			rowRendered = styleActiveRow.Width(contentW).Render(rowRendered)
		}
		lines = append(lines, rowRendered)
	}

	innerH := height - 2
	for len(lines) < innerH {
		lines = append(lines, "")
	}
	if len(lines) > innerH {
		lines = lines[:innerH]
	}

	return box.Width(width).Height(height).Render(strings.Join(lines, "\n"))
}

func (m Model) renderInspector(width, height int) string {
	box := styleCard
	titleText := "[2] DETAILS & DAILY LOG"
	if m.ActivePane == paneInspector {
		box = styleCardFocus
		titleText = "▶ [2] DETAILS & DAILY LOG"
	}

	title := styleTitle.Render(titleText)
	lines := []string{title, ""}
	contentW := width - 4

	tasks := m.filteredTasks()
	if len(tasks) > 0 && m.Cursor < len(tasks) {
		task := tasks[m.Cursor].task

		statusText := "● COMPLETED"
		statusColor := colorGreen
		if !task.Done {
			statusText = "○ IN PROGRESS"
			statusColor = colorYellow
		}
		statusBadge := lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(statusText)

		lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(colorTextPrimary).MaxWidth(contentW).Render(task.Title))
		lines = append(lines, statusBadge)

		if len(task.Notes) > 0 {
			lines = append(lines, "", lipgloss.NewStyle().Foreground(colorAccent).Render("Notes:"))
			for _, note := range task.Notes {
				lines = append(lines, lipgloss.NewStyle().Foreground(colorTextPrimary).MaxWidth(contentW).Render(" • "+note))
			}
		}
	}

	if len(m.Doc.DailyLog) > 0 {
		lines = append(lines, "", lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("Daily Log:"))
		for _, logLine := range m.Doc.DailyLog {
			lines = append(lines, lipgloss.NewStyle().Foreground(colorTextPrimary).MaxWidth(contentW).Render(" • "+logLine))
		}
	}

	innerH := height - 2
	for len(lines) < innerH {
		lines = append(lines, "")
	}
	if len(lines) > innerH {
		lines = lines[:innerH]
	}

	return box.Width(width).Height(height).Render(strings.Join(lines, "\n"))
}

func (m Model) renderTagRibbon(width int) string {
	box := styleCard
	prefix := "🏷️  TAGS: "
	if m.ActivePane == paneTags {
		box = styleCardFocus
		prefix = "▶ 🏷️  TAGS [h/l]: "
	}

	tags := m.getAllTagsList()
	var pills []string

	for idx, it := range tags {
		style := styleTagPill.Copy()
		if it.name == m.SelectedTag {
			style = style.Background(colorBorderFocus).Foreground(lipgloss.Color("#1a1b26")).Bold(true)
		}
		if m.ActivePane == paneTags && idx == m.TagCursor {
			style = style.Border(lipgloss.NormalBorder(), false, false, true, false).BorderForeground(colorAccent)
		}
		pills = append(pills, style.Render(fmt.Sprintf("#%s (%d)", it.name, it.count)))
	}

	content := prefix + strings.Join(pills, " ")
	return box.Width(width).Render(content)
}

func (m Model) renderHelpBar() string {
	keys := []struct{ k, desc string }{
		{"1-4", "Окна"},
		{"Tab", "Смена"},
		{"j/k", "Навигация"},
		{"Space", "Выполнить"},
		{"a", "Создать"},
		{"n", "Заметка"},
		{"d", "Удалить"},
		{"q", "Выход"},
	}
	var parts []string
	for _, it := range keys {
		k := lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(it.k)
		d := lipgloss.NewStyle().Foreground(colorTextMuted).Render(it.desc)
		parts = append(parts, k+" "+d)
	}
	return lipgloss.NewStyle().Padding(0, 1).Render(strings.Join(parts, "  •  "))
}

func (m Model) renderModal() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(colorMagenta).Render("✨ СОЗДАНИЕ ЗАДАЧИ")
	f1 := lipgloss.JoinVertical(lipgloss.Left, lipgloss.NewStyle().Foreground(colorTextMuted).Render("Название:"), m.Inputs[fieldTitle].View())
	f2 := lipgloss.JoinVertical(lipgloss.Left, lipgloss.NewStyle().Foreground(colorTextMuted).Render("Теги:"), m.Inputs[fieldTag].View())
	f3 := lipgloss.JoinVertical(lipgloss.Left, lipgloss.NewStyle().Foreground(colorTextMuted).Render("Срок:"), m.Inputs[fieldDue].View())
	hints := lipgloss.NewStyle().Foreground(colorTextMuted).Render("[Enter] Сохранить  •  [Tab] Далее  •  [Esc] Отмена")

	form := lipgloss.JoinVertical(lipgloss.Left, title, "", f1, "", f2, "", f3, "", hints)
	dialog := styleCardFocus.Padding(1, 2).Width(52).Render(form)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, dialog)
}

func (m Model) renderNoteModal() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(colorAccent).Render("📝 ДОБАВИТЬ ЗАМЕТКУ")
	inputBox := m.NoteInput.View()
	hints := lipgloss.NewStyle().Foreground(colorTextMuted).Render("[Enter] Сохранить  •  [Esc] Отмена")

	form := lipgloss.JoinVertical(lipgloss.Left, title, "", inputBox, "", hints)
	dialog := styleCardFocus.Padding(1, 2).Width(52).Render(form)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, dialog)
}
