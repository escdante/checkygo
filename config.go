package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	DayStartHour int `json:"day_start_hour"`
}

// DataDir returns the data directory: T_DATA_DIR env var takes precedence,
// then %LOCALAPPDATA%\t on Windows, then ~/.local/share/t elsewhere.
func DataDir() string {
	if d := os.Getenv("T_DATA_DIR"); d != "" {
		return d
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	if runtime.GOOS == "windows" {
		local := os.Getenv("LOCALAPPDATA")
		if local == "" {
			local = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(local, "t")
	}
	return filepath.Join(home, ".local", "share", "t")
}

// LoadConfig reads config.json; returns defaults (DayStartHour=4) if the file
// does not exist.
func LoadConfig(dataDir string) (Config, error) {
	cfg := Config{DayStartHour: 4}
	path := filepath.Join(dataDir, "config.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// LoadTasks reads tasks.json and returns the ordered task list.
// Returns a descriptive error if the file is missing or empty.
func LoadTasks(dataDir string) ([]string, error) {
	path := filepath.Join(dataDir, "tasks.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("no tasks configured — run `t tasks` to add your daily tasks")
	}
	if err != nil {
		return nil, err
	}
	var tasks []string
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, errors.New("no tasks configured — run `t tasks` to add your daily tasks")
	}
	return tasks, nil
}

// saveTasks writes the task list to tasks.json. Used by runTasks after the
// editor closes (and by tests).
func saveTasks(dataDir string, tasks []string) error {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, "tasks.json"), data, 0o644)
}
