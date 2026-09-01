package domain

import (
	"strconv"
	"strings"
	"time"
)

// ShouldRepeatOn determines if a recurring task should appear on the given date.
// Supported recurrence rules via @repeat attribute:
//   - @repeat(daily) — triggers every single day
//   - @repeat(weekly:mon,wed,fri) — triggers on specific days of the week
//   - @repeat(biweekly:even:tue) or @repeat(biweekly:odd:mon) — triggers every other week
//   - @repeat(monthly:15) — triggers on a specific day of the month
func (t *Task) ShouldRepeatOn(date time.Time) bool {
	rule, found := t.AttrVal("repeat")
	if !found {
		return false
	}

	rule = strings.ToLower(strings.TrimSpace(rule))

	// 1. Daily recurrence
	if rule == "daily" {
		return true
	}

	// 2. Days of week: weekly:mon,wed,fri
	if strings.HasPrefix(rule, "weekly:") {
		daysStr := strings.TrimPrefix(rule, "weekly:")
		targetDay := shortWeekday(date.Weekday())
		for _, day := range strings.Split(daysStr, ",") {
			if strings.TrimSpace(day) == targetDay {
				return true
			}
		}
		return false
	}

	// 3. Biweekly (even/odd ISO week parity): biweekly:even:tue or biweekly:odd:mon
	if strings.HasPrefix(rule, "biweekly:") {
		parts := strings.Split(strings.TrimPrefix(rule, "biweekly:"), ":")
		if len(parts) != 2 {
			return false
		}
		parity, day := parts[0], parts[1]

		_, weekNum := date.ISOWeek()
		isEvenWeek := weekNum%2 == 0

		if (parity == "even" && !isEvenWeek) || (parity == "odd" && isEvenWeek) {
			return false
		}

		return shortWeekday(date.Weekday()) == day
	}

	// 4. Monthly by day number: monthly:15
	if strings.HasPrefix(rule, "monthly:") {
		dayStr := strings.TrimPrefix(rule, "monthly:")
		dayNum, err := strconv.Atoi(dayStr)
		if err != nil || dayNum < 1 || dayNum > 31 {
			return false
		}
		return date.Day() == dayNum
	}

	return false
}

// shortWeekday converts time.Weekday to a lowercase three-letter string representation.
func shortWeekday(wd time.Weekday) string {
	switch wd {
	case time.Monday:
		return "mon"
	case time.Tuesday:
		return "tue"
	case time.Wednesday:
		return "wed"
	case time.Thursday:
		return "thu"
	case time.Friday:
		return "fri"
	case time.Saturday:
		return "sat"
	case time.Sunday:
		return "sun"
	default:
		return ""
	}
}
