// main.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func main() {
	dataDir := DataDir()
	cfg, err := LoadConfig(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	args := os.Args[1:]

	// These commands work even without tasks.json
	if len(args) > 0 {
		switch args[0] {
		case "tasks":
			if err := runTasks(dataDir); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			return
		case "config":
			runConfig(cfg, dataDir)
			return
		}
	}

	// All other commands require tasks.json
	tasks, err := LoadTasks(dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	dayKey := DayKey(time.Now(), cfg.DayStartHour)
	state, err := LoadState(dataDir, dayKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading state: %v\n", err)
		os.Exit(1)
	}
	// persist rollover if the day changed (LoadState may have reset Checked)
	if err := SaveState(dataDir, state); err != nil {
		fmt.Fprintf(os.Stderr, "error saving state: %v\n", err)
		os.Exit(1)
	}

	// No args → compact board
	if len(args) == 0 {
		RenderBoard(tasks, state, os.Stdout)
		return
	}

	switch args[0] {
	case "today":
		events, err := ReadEvents(dataDir, dayKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading events: %v\n", err)
			os.Exit(1)
		}
		RenderVerboseBoard(tasks, state, events, os.Stdout)

	case "history":
		statuses, err := ComputeHistory(dataDir, len(tasks), 30)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		RenderHistory(statuses, os.Stdout)

	case "log":
		if err := runFreeLog(dataDir, dayKey, args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	default:
		n, parseErr := strconv.Atoi(args[0])
		if parseErr != nil || n < 1 || n > len(tasks) {
			fmt.Fprintf(os.Stderr, "unknown command %q — run `t` to see the board\n", args[0])
			os.Exit(1)
		}
		taskIdx := n - 1 // display is 1-based, storage is 0-based
		if err := runCheck(dataDir, dayKey, taskIdx, tasks, &state); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if err := SaveState(dataDir, state); err != nil {
			fmt.Fprintf(os.Stderr, "error saving state: %v\n", err)
			os.Exit(1)
		}
		RenderBoard(tasks, state, os.Stdout)
	}
}

// runCheck handles `t N`. If the task is unchecked it prompts for a log and
// checks it; if already checked it confirms before unchecking.
func runCheck(dataDir, dayKey string, idx int, tasks []string, state *State) error {
	key := strconv.Itoa(idx)
	ts, alreadyChecked := state.Checked[key]

	if alreadyChecked {
		t, _ := time.Parse(time.RFC3339, ts)
		fmt.Printf("  %s  %s  %s\n", checkMark(), tasks[idx], dim(t.Format("15:04")))
		fmt.Print("  uncheck? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(answer)) != "y" {
			return nil
		}
		delete(state.Checked, key)
		now := time.Now().Format(time.RFC3339)
		return AppendEvent(dataDir, dayKey, Event{Ts: now, Type: EventUncheck, Task: &idx})
	}

	// Unchecked → prompt for log
	fmt.Printf("  %s\n  log %s ", tasks[idx], dim("›"))
	reader := bufio.NewReader(os.Stdin)
	logLine, _ := reader.ReadString('\n')
	logLine = strings.TrimSpace(logLine)
	if logLine == "" {
		fmt.Printf("  log %s ", dim("›"))
		logLine, _ = reader.ReadString('\n')
		logLine = strings.TrimSpace(logLine)
		if logLine == "" {
			fmt.Println("  aborted — no log provided")
			return nil
		}
	}

	nowT := time.Now()
	now := nowT.Format(time.RFC3339)
	state.Checked[key] = now
	fmt.Printf("  %s\n", checkMark())
	return AppendEvent(dataDir, dayKey, Event{Ts: now, Type: EventCheck, Task: &idx, Log: logLine})
}

// runFreeLog handles `t log [text...]`. All args are joined; if no args,
// prompts inline.
func runFreeLog(dataDir, dayKey string, args []string) error {
	var logLine string
	if len(args) > 0 {
		logLine = strings.Join(args, " ")
	} else {
		fmt.Printf("  log %s ", dim("›"))
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		logLine = strings.TrimSpace(line)
		if logLine == "" {
			fmt.Println("  aborted — no log provided")
			return nil
		}
	}
	nowT := time.Now()
	now := nowT.Format(time.RFC3339)
	if err := AppendEvent(dataDir, dayKey, Event{Ts: now, Type: EventFree, Log: logLine}); err != nil {
		return err
	}
	fmt.Printf("  %s logged %s\n", checkMark(), dim(nowT.Format("15:04")))
	return nil
}

// runTasks opens tasks.json in $EDITOR (notepad on Windows, vi elsewhere).
func runTasks(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dataDir, "tasks.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte("[]\n"), 0o644); err != nil {
			return err
		}
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		if runtime.GOOS == "windows" {
			editor = "notepad"
		} else {
			editor = "vi"
		}
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runConfig prints the current configuration to stdout.
func runConfig(cfg Config, dataDir string) {
	fmt.Printf("\n  day_start_hour  %d\n  data_dir        %s\n\n", cfg.DayStartHour, dataDir)
}
