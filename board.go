// board.go
package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// RenderBoard prints the compact daily board.
func RenderBoard(tasks []string, s State, w io.Writer) {
	checked := len(s.Checked)
	total := len(tasks)
	now := time.Now()

	fmt.Fprintf(w, "\n  TODAY  %s  %s  %s  %d/%d\n\n",
		dim("·"), dim(now.Format("Mon 2 Jan")), dim("·"), checked, total)

	for i, task := range tasks {
		key := strconv.Itoa(i)
		if ts, ok := s.Checked[key]; ok {
			t, err := time.Parse(time.RFC3339, ts)
			timeStr := ""
			if err == nil {
				timeStr = t.Format("15:04")
			}
			fmt.Fprintf(w, "  %d  %s  %s  %s\n",
				i+1, checkMark(), dim(task), dim(timeStr))
		} else {
			fmt.Fprintf(w, "  %d  %s  %s\n", i+1, dot(), task)
		}
	}
	fmt.Fprintln(w)
}

// RenderVerboseBoard prints the board with log lines under each checked task
// and free-form events below a separator.
func RenderVerboseBoard(tasks []string, s State, events []Event, w io.Writer) {
	// Most recent check log per task index
	taskLogs := make(map[int]string)
	for _, e := range events {
		if e.Type == EventCheck && e.Task != nil {
			taskLogs[*e.Task] = e.Log
		}
	}

	checked := len(s.Checked)
	total := len(tasks)
	now := time.Now()

	fmt.Fprintf(w, "\n  TODAY  %s  %s  %s  %d/%d\n\n",
		dim("·"), dim(now.Format("Mon 2 Jan")), dim("·"), checked, total)

	for i, task := range tasks {
		key := strconv.Itoa(i)
		if ts, ok := s.Checked[key]; ok {
			t, err := time.Parse(time.RFC3339, ts)
			timeStr := ""
			if err == nil {
				timeStr = t.Format("15:04")
			}
			fmt.Fprintf(w, "  %d  %s  %s  %s\n",
				i+1, checkMark(), dim(task), dim(timeStr))
			if log := taskLogs[i]; log != "" {
				fmt.Fprintf(w, "       %s %s\n", dim("→"), dim(log))
			}
			fmt.Fprintln(w)
		} else {
			fmt.Fprintf(w, "  %d  %s  %s\n", i+1, dot(), task)
		}
	}

	var freeEvents []Event
	for _, e := range events {
		if e.Type == EventFree {
			freeEvents = append(freeEvents, e)
		}
	}
	if len(freeEvents) > 0 {
		fmt.Fprintln(w, "  "+dim(strings.Repeat("─", 24)))
		for _, e := range freeEvents {
			t, err := time.Parse(time.RFC3339, e.Ts)
			timeStr := ""
			if err == nil {
				timeStr = t.Format("15:04")
			}
			fmt.Fprintf(w, "  %s  %s\n", dim(timeStr), e.Log)
		}
		fmt.Fprintln(w)
	}
}
