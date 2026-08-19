package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type EventType string

const (
	EventCheck   EventType = "check"
	EventFree    EventType = "free"
	EventUncheck EventType = "uncheck"
)

type Event struct {
	Ts   string    `json:"ts"`
	Type EventType `json:"type"`
	Task *int      `json:"task,omitempty"` // nil for free events
	Log  string    `json:"log,omitempty"`
}

// AppendEvent appends a single event to logs/YYYY-MM-DD.jsonl.
func AppendEvent(dataDir, dayKey string, e Event) error {
	logDir := filepath.Join(dataDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(logDir, dayKey+".jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	return err
}

// ReadEvents reads all events from logs/YYYY-MM-DD.jsonl.
// Returns an empty slice (no error) if the file does not exist.
func ReadEvents(dataDir, dayKey string) ([]Event, error) {
	path := filepath.Join(dataDir, "logs", dayKey+".jsonl")
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var events []Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			continue // skip malformed lines rather than failing
		}
		events = append(events, e)
	}
	return events, scanner.Err()
}
