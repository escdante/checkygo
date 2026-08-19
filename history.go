package main

import (
	"fmt"
	"io"
	"time"
)

type DayStatus int

const (
	DayNone    DayStatus = 0
	DayPartial DayStatus = 1
	DayAll     DayStatus = 2
)

// computeDayStatus returns how complete a day was based on its events.
// A task is net-checked if its last check/uncheck event is a check.
func computeDayStatus(events []Event, totalTasks int) DayStatus {
	last := make(map[int]EventType)
	for _, e := range events {
		if e.Task != nil && (e.Type == EventCheck || e.Type == EventUncheck) {
			last[*e.Task] = e.Type
		}
	}
	netChecked := 0
	for _, typ := range last {
		if typ == EventCheck {
			netChecked++
		}
	}
	switch {
	case netChecked == 0:
		return DayNone
	case netChecked >= totalTasks:
		return DayAll
	default:
		return DayPartial
	}
}

// ComputeHistory returns DayStatus for the last `days` calendar days, oldest first.
func ComputeHistory(dataDir string, totalTasks int, days int) ([]DayStatus, error) {
	statuses := make([]DayStatus, days)
	now := time.Now()
	for i := 0; i < days; i++ {
		day := now.AddDate(0, 0, -(days - 1 - i))
		key := day.Format("2006-01-02")
		events, err := ReadEvents(dataDir, key)
		if err != nil {
			return nil, err
		}
		statuses[i] = computeDayStatus(events, totalTasks)
	}
	return statuses, nil
}

// RenderHistory prints the 30-day completion grid.
func RenderHistory(statuses []DayStatus, w io.Writer) {
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  last %d days  %s\n\n  ", len(statuses), dim("(■ all  ▪ partial  · none)"))
	for _, s := range statuses {
		switch s {
		case DayAll:
			fmt.Fprint(w, "■ ")
		case DayPartial:
			fmt.Fprint(w, "▪ ")
		default:
			fmt.Fprint(w, dim("· "))
		}
	}
	fmt.Fprintln(w)
}
