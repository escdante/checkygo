package main

import "testing"

func TestComputeDayStatus_allChecked(t *testing.T) {
	i0, i1, i2 := 0, 1, 2
	events := []Event{
		{Type: EventCheck, Task: &i0},
		{Type: EventCheck, Task: &i1},
		{Type: EventCheck, Task: &i2},
	}
	if got := computeDayStatus(events, 3); got != DayAll {
		t.Errorf("want DayAll got %v", got)
	}
}

func TestComputeDayStatus_partial(t *testing.T) {
	i0 := 0
	events := []Event{
		{Type: EventCheck, Task: &i0},
	}
	if got := computeDayStatus(events, 3); got != DayPartial {
		t.Errorf("want DayPartial got %v", got)
	}
}

func TestComputeDayStatus_checkThenUncheck(t *testing.T) {
	i0 := 0
	events := []Event{
		{Type: EventCheck, Task: &i0},
		{Type: EventUncheck, Task: &i0},
	}
	if got := computeDayStatus(events, 3); got != DayNone {
		t.Errorf("want DayNone after uncheck got %v", got)
	}
}

func TestComputeDayStatus_empty(t *testing.T) {
	if got := computeDayStatus(nil, 3); got != DayNone {
		t.Errorf("want DayNone for empty events got %v", got)
	}
}
