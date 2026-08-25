package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CalendarModel struct {
	SelectedDate time.Time
	ViewYear     int
	ViewMonth    time.Month
	Activity     map[string]int
	Focused      bool
}

func NewCalendar(initialDate time.Time) CalendarModel {
	return CalendarModel{
		SelectedDate: initialDate,
		ViewYear:     initialDate.Year(),
		ViewMonth:    initialDate.Month(),
		Activity:     make(map[string]int),
		Focused:      false,
	}
}

func (c *CalendarModel) SetDate(t time.Time) {
	c.SelectedDate = t
	c.ViewYear = t.Year()
	c.ViewMonth = t.Month()
}

func (c *CalendarModel) MoveSelection(days int) time.Time {
	c.SelectedDate = c.SelectedDate.AddDate(0, 0, days)
	c.ViewYear = c.SelectedDate.Year()
	c.ViewMonth = c.SelectedDate.Month()
	return c.SelectedDate
}

func (c *CalendarModel) MoveMonth(months int) time.Time {
	c.SelectedDate = c.SelectedDate.AddDate(0, months, 0)
	c.ViewYear = c.SelectedDate.Year()
	c.ViewMonth = c.SelectedDate.Month()
	return c.SelectedDate
}

func (c *CalendarModel) Update(msg tea.Msg) (time.Time, bool) {
	dateChanged := false

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "h", "left":
			c.MoveSelection(-1)
			dateChanged = true
		case "l", "right":
			c.MoveSelection(1)
			dateChanged = true
		case "k", "up":
			c.MoveSelection(-7)
			dateChanged = true
		case "j", "down":
			c.MoveSelection(7)
			dateChanged = true
		case "H":
			c.MoveMonth(-1)
			dateChanged = true
		case "L":
			c.MoveMonth(1)
			dateChanged = true
		case "t":
			c.SetDate(time.Now())
			dateChanged = true
		}
	}

	return c.SelectedDate, dateChanged
}

func (c CalendarModel) View(width, height int) string {
	box := styleCard
	titlePrefix := "📅 CALENDAR"
	if c.Focused {
		box = styleCardFocus
		titlePrefix = "▶ 📅 CALENDAR"
	}

	monthTitle := styleTitle.Render(fmt.Sprintf("%s %s %d", titlePrefix, c.ViewMonth.String()[:3], c.ViewYear))
	weekHeader := lipgloss.NewStyle().Foreground(colorTextMuted).Render("Mo Tu We Th Fr Sa Su")

	firstDay := time.Date(c.ViewYear, c.ViewMonth, 1, 0, 0, 0, 0, time.UTC)
	startOffset := int(firstDay.Weekday())
	if startOffset == 0 {
		startOffset = 7
	}
	startOffset--

	daysInMonth := time.Date(c.ViewYear, c.ViewMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()

	var rows []string
	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left, monthTitle, "   ", weekHeader))

	curLine := strings.Repeat("   ", startOffset)
	col := startOffset

	for d := 1; d <= daysInMonth; d++ {
		isSelected := c.SelectedDate.Year() == c.ViewYear &&
			c.SelectedDate.Month() == c.ViewMonth &&
			c.SelectedDate.Day() == d

		cell := fmt.Sprintf("%2d ", d)
		if isSelected {
			cell = lipgloss.NewStyle().
				Background(colorBorderFocus).
				Foreground(lipgloss.Color("#1a1b26")).
				Bold(true).
				Render(fmt.Sprintf("%2d", d)) + " "
		} else if count, ok := c.Activity[fmt.Sprintf("%04d-%02d-%02d", c.ViewYear, int(c.ViewMonth), d)]; ok && count > 0 {
			cell = lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Render(cell)
		} else {
			cell = lipgloss.NewStyle().Foreground(colorTextPrimary).Render(cell)
		}

		curLine += cell
		col++
		if col == 7 {
			rows = append(rows, curLine)
			curLine = ""
			col = 0
		}
	}
	if curLine != "" {
		rows = append(rows, curLine)
	}

	innerH := height - 2
	for len(rows) < innerH {
		rows = append(rows, "")
	}
	if len(rows) > innerH {
		rows = rows[:innerH]
	}

	calMatrix := lipgloss.JoinVertical(lipgloss.Left, rows...)

	if width > 44 {
		sideNav := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(c.SelectedDate.Format("2006-01-02")),
			"",
			lipgloss.NewStyle().Foreground(colorTextMuted).Render("h/l: ±1д"),
			lipgloss.NewStyle().Foreground(colorTextMuted).Render("j/k: ±1нед"),
			lipgloss.NewStyle().Foreground(colorTextMuted).Render("H/L: ±1мес"),
			lipgloss.NewStyle().Foreground(colorTextMuted).Render("t:   сегодня"),
		)
		content := lipgloss.JoinHorizontal(lipgloss.Top, calMatrix, "    ", sideNav)
		return box.Width(width).Height(height).Render(content)
	}

	return box.Width(width).Height(height).Render(calMatrix)
}
