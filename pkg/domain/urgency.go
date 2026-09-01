package domain

import (
	"strconv"
	"strings"
	"time"
)

// UrgencyConfig defines the weights used to calculate the urgency score of tasks.
type UrgencyConfig struct {
	OverdueWeight     float64
	DueTodayWeight    float64
	DueTomorrowWeight float64
	DueThisWeekWeight float64

	PrioHighWeight float64
	PrioMedWeight  float64
	PrioLowWeight  float64

	TimeNowWeight    float64
	TimeSoonWeight   float64
	TimePassedWeight float64
}

// DefaultUrgencyConfig returns the default balanced urgency scoring weights.
func DefaultUrgencyConfig() UrgencyConfig {
	return UrgencyConfig{
		OverdueWeight:     12.0,
		DueTodayWeight:    10.0,
		DueTomorrowWeight: 6.0,
		DueThisWeekWeight: 3.0,

		PrioHighWeight: 4.0,
		PrioMedWeight:  2.0,
		PrioLowWeight:  0.5,

		TimeNowWeight:    5.0,
		TimeSoonWeight:   3.5,
		TimePassedWeight: 4.0,
	}
}

// Urgency calculates the urgency coefficient of a task relative to the provided point in time.
// Completed tasks always return 0.0. Higher scores indicate higher priority and urgency.
func (t *Task) Urgency(now time.Time, cfg ...UrgencyConfig) float64 {
	if t.Done {
		return 0.0
	}

	config := DefaultUrgencyConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	urgency := 1.0

	// 1. Evaluate due dates (@due)
	if dueStr, found := t.AttrVal("due"); found {
		if dueDate, err := time.Parse("2006-01-02", dueStr); err == nil {
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			dueDay := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, now.Location())
			daysDiff := dueDay.Sub(today).Hours() / 24

			switch {
			case daysDiff < 0:
				urgency += config.OverdueWeight
			case daysDiff == 0:
				urgency += config.DueTodayWeight
			case daysDiff == 1:
				urgency += config.DueTomorrowWeight
			case daysDiff <= 7:
				urgency += config.DueThisWeekWeight
			}
		}
	}

	// 2. Evaluate priority levels (@prio)
	if prio, found := t.AttrVal("prio"); found {
		switch strings.ToLower(prio) {
		case "high", "h", "1":
			urgency += config.PrioHighWeight
		case "med", "medium", "m", "2":
			urgency += config.PrioMedWeight
		case "low", "l", "3":
			urgency += config.PrioLowWeight
		}
	}

	// 3. Evaluate time slots (@time)
	if timeStr, found := t.AttrVal("time"); found {
		startTimeStr := strings.Split(timeStr, "-")[0]
		parts := strings.Split(strings.TrimSpace(startTimeStr), ":")
		if len(parts) == 2 {
			if h, errH := strconv.Atoi(parts[0]); errH == nil {
				if m, errM := strconv.Atoi(parts[1]); errM == nil && h >= 0 && h < 24 && m >= 0 && m < 60 {
					slotTime := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, now.Location())
					diff := slotTime.Sub(now).Minutes()

					switch {
					case diff <= 0 && diff >= -45:
						urgency += config.TimeNowWeight
					case diff > 0 && diff <= 60:
						urgency += config.TimeSoonWeight
					case diff < -45:
						urgency += config.TimePassedWeight
					}
				}
			}
		}
	}

	return urgency
}
