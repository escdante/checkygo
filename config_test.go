package main

import (
	"testing"
)

func TestDataDir_envOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("T_DATA_DIR", dir)
	got := DataDir()
	if got != dir {
		t.Fatalf("want %q got %q", dir, got)
	}
}

func TestLoadTasks_roundtrip(t *testing.T) {
	dir := t.TempDir()
	tasks := []string{"task one", "task two"}
	if err := saveTasks(dir, tasks); err != nil {
		t.Fatal(err)
	}
	got, err := LoadTasks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(tasks) {
		t.Fatalf("want %d tasks got %d", len(tasks), len(got))
	}
	for i := range tasks {
		if got[i] != tasks[i] {
			t.Errorf("[%d] want %q got %q", i, tasks[i], got[i])
		}
	}
}

func TestLoadConfig_defaults(t *testing.T) {
	dir := t.TempDir()
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DayStartHour != 4 {
		t.Fatalf("want DayStartHour=4 got %d", cfg.DayStartHour)
	}
}
