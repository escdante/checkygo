package main

import (
	"testing"
	"time"
)

func TestDayKey_afterStartHour(t *testing.T) {
	now := time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC)
	got := DayKey(now, 4)
	if got != "2026-08-19" {
		t.Fatalf("want 2026-08-19 got %s", got)
	}
}

func TestDayKey_beforeStartHour(t *testing.T) {
	// 02:00 local — still "yesterday" when start hour is 4
	now := time.Date(2026, 8, 19, 2, 0, 0, 0, time.UTC)
	got := DayKey(now, 4)
	if got != "2026-08-18" {
		t.Fatalf("want 2026-08-18 got %s", got)
	}
}

func TestDayKey_exactStartHour(t *testing.T) {
	// 04:00 exactly — counts as "today"
	now := time.Date(2026, 8, 19, 4, 0, 0, 0, time.UTC)
	got := DayKey(now, 4)
	if got != "2026-08-19" {
		t.Fatalf("want 2026-08-19 got %s", got)
	}
}

func TestRollover_setsNewKey(t *testing.T) {
	s := State{
		DayKey:  "2026-08-18",
		Checked: map[string]string{"0": "2026-08-18T09:00:00Z"},
	}
	rolled := Rollover(&s, "2026-08-19")
	if !rolled {
		t.Fatal("expected rollover to occur")
	}
	if s.DayKey != "2026-08-19" {
		t.Fatalf("want DayKey=2026-08-19 got %s", s.DayKey)
	}
	if len(s.Checked) != 0 {
		t.Fatalf("expected Checked to be empty after rollover, got %v", s.Checked)
	}
}

func TestRollover_noopWhenSameKey(t *testing.T) {
	s := State{
		DayKey:  "2026-08-19",
		Checked: map[string]string{"0": "2026-08-19T09:00:00Z"},
	}
	rolled := Rollover(&s, "2026-08-19")
	if rolled {
		t.Fatal("expected no rollover")
	}
	if len(s.Checked) != 1 {
		t.Fatal("expected Checked to be unchanged")
	}
}
