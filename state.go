package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type State struct {
	DayKey  string            `json:"day_key"`
	Checked map[string]string `json:"checked"` // task index (string) → RFC3339 timestamp
}

// DayKey returns the day identifier for now given the configured start hour.
// A "day" runs from startHour:00 to (startHour-1):59 the next calendar day.
func DayKey(now time.Time, startHour int) string {
	if now.Hour() < startHour {
		now = now.AddDate(0, 0, -1)
	}
	return now.Format("2006-01-02")
}

// Rollover resets Checked and updates DayKey if newKey differs.
// Returns true if a rollover occurred.
func Rollover(s *State, newKey string) bool {
	if s.DayKey == newKey {
		return false
	}
	s.DayKey = newKey
	s.Checked = make(map[string]string)
	return true
}

// LoadState reads state.json, applies rollover if needed, and returns the state.
// Returns a fresh state for dayKey if the file is missing or corrupted.
func LoadState(dataDir string, dayKey string) (State, error) {
	fresh := State{DayKey: dayKey, Checked: make(map[string]string)}
	path := filepath.Join(dataDir, "state.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return fresh, nil
	}
	if err != nil {
		return fresh, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return fresh, nil // corrupted — treat as fresh, continue normally
	}
	if s.Checked == nil {
		s.Checked = make(map[string]string)
	}
	Rollover(&s, dayKey)
	return s, nil
}

// SaveState writes state.json.
func SaveState(dataDir string, s State) error {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, "state.json"), data, 0o644)
}
