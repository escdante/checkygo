package main

import (
	"testing"
	"time"
)

func TestAppendAndReadEvents(t *testing.T) {
	dir := t.TempDir()
	dayKey := "2026-08-19"
	ts := time.Date(2026, 8, 19, 9, 14, 23, 0, time.UTC).Format(time.RFC3339)
	taskIdx := 0

	if err := AppendEvent(dir, dayKey, Event{Ts: ts, Type: EventCheck, Task: &taskIdx, Log: "shipped the feature"}); err != nil {
		t.Fatalf("AppendEvent check: %v", err)
	}
	if err := AppendEvent(dir, dayKey, Event{Ts: ts, Type: EventFree, Log: "extra note"}); err != nil {
		t.Fatalf("AppendEvent free: %v", err)
	}

	got, err := ReadEvents(dir, dayKey)
	if err != nil {
		t.Fatalf("ReadEvents: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 events got %d", len(got))
	}
	if got[0].Type != EventCheck || *got[0].Task != 0 || got[0].Log != "shipped the feature" {
		t.Errorf("unexpected event[0]: %+v", got[0])
	}
	if got[1].Type != EventFree || got[1].Log != "extra note" {
		t.Errorf("unexpected event[1]: %+v", got[1])
	}
}

func TestReadEvents_missingFile(t *testing.T) {
	dir := t.TempDir()
	events, err := ReadEvents(dir, "2026-08-01")
	if err != nil {
		t.Fatalf("expected no error for missing log file, got: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected empty slice for missing file, got %d events", len(events))
	}
}
