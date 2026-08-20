package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultModel = "claude-haiku-4-5"
const anthropicBaseURL = "https://api.anthropic.com"

type summaryCache struct {
	EventCount int    `json:"event_count"`
	Summary    string `json:"summary"`
}

func summaryCachePath(dataDir, dayKey string) string {
	return filepath.Join(dataDir, "logs", dayKey+".summary.json")
}

func weeklyCachePath(dataDir, isoWeek string) string {
	return filepath.Join(dataDir, "logs", "week-"+isoWeek+".summary.json")
}

func loadSummaryCache(path string, eventCount int) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	var c summaryCache
	if err := json.Unmarshal(data, &c); err != nil {
		return "", false
	}
	if c.EventCount != eventCount {
		return "", false
	}
	return c.Summary, true
}

func saveSummaryCache(path, summary string, eventCount int) {
	c := summaryCache{EventCount: eventCount, Summary: summary}
	data, _ := json.Marshal(c)
	_ = os.WriteFile(path, data, 0o644)
}

func resolveApiKey(cfg Config) string {
	if k := os.Getenv("ANTHROPIC_API_KEY"); k != "" {
		return k
	}
	if k := os.Getenv("T_API_KEY"); k != "" {
		return k
	}
	return cfg.ApiKey
}

func resolveBaseURL(cfg Config) string {
	if cfg.ApiBaseURL != "" {
		return cfg.ApiBaseURL
	}
	return anthropicBaseURL
}

func resolveModel(cfg Config) string {
	if cfg.ApiModel != "" {
		return cfg.ApiModel
	}
	return defaultModel
}

// callLLM sends a single-turn prompt to the configured API.
// If baseURL is Anthropic, uses the Anthropic Messages format.
// All other URLs (Groq, OpenRouter, local Ollama, etc.) use the
// OpenAI-compatible /chat/completions format.
func callLLM(apiKey, baseURL, model, prompt string) (string, error) {
	isAnthropic := strings.Contains(baseURL, "anthropic.com")

	var endpoint string
	var body []byte
	if isAnthropic {
		endpoint = strings.TrimRight(baseURL, "/") + "/v1/messages"
		body, _ = json.Marshal(map[string]any{
			"model":      model,
			"max_tokens": 1024,
			"messages":   []map[string]string{{"role": "user", "content": prompt}},
		})
	} else {
		endpoint = strings.TrimRight(baseURL, "/") + "/chat/completions"
		body, _ = json.Marshal(map[string]any{
			"model":      model,
			"max_tokens": 1024,
			"messages":   []map[string]string{{"role": "user", "content": prompt}},
		})
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if isAnthropic {
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	} else {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(raw))
	}

	if isAnthropic {
		var result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		}
		if err := json.Unmarshal(raw, &result); err != nil {
			return "", err
		}
		if len(result.Content) == 0 {
			return "", errors.New("empty response")
		}
		return strings.TrimSpace(result.Content[0].Text), nil
	}

	// OpenAI-compatible response
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", errors.New("empty response")
	}
	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

// SummarizeDay generates (or returns cached) a one-paragraph summary of today's work.
func SummarizeDay(cfg Config, dataDir, dayKey string, tasks []string, events []Event) (string, error) {
	apiKey := resolveApiKey(cfg)
	if apiKey == "" {
		return "", errors.New("no API key — set ANTHROPIC_API_KEY or run `t config set api_key <key>`")
	}

	cachePath := summaryCachePath(dataDir, dayKey)
	if cached, ok := loadSummaryCache(cachePath, len(events)); ok {
		return cached, nil
	}

	var sb strings.Builder
	sb.WriteString("Here is my work log for today. Write a single short paragraph (3-5 sentences) summarizing what I accomplished. Be concrete, use plain language, no bullet points.\n\nTasks:\n")
	for i, t := range tasks {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, t))
	}
	sb.WriteString("\nEvents:\n")
	for _, e := range events {
		t, _ := time.Parse(time.RFC3339, e.Ts)
		timeStr := t.Format("15:04")
		switch e.Type {
		case EventCheck:
			idx := 0
			if e.Task != nil {
				idx = *e.Task
			}
			taskName := ""
			if idx < len(tasks) {
				taskName = tasks[idx]
			}
			sb.WriteString(fmt.Sprintf("  [%s] checked: %s — %s\n", timeStr, taskName, e.Log))
		case EventUncheck:
			sb.WriteString(fmt.Sprintf("  [%s] unchecked task %d\n", timeStr, func() int {
				if e.Task != nil {
					return *e.Task + 1
				}
				return 0
			}()))
		case EventFree:
			sb.WriteString(fmt.Sprintf("  [%s] note: %s\n", timeStr, e.Log))
		}
	}

	model := resolveModel(cfg)
	baseURL := resolveBaseURL(cfg)
	summary, err := callLLM(apiKey, baseURL, model, sb.String())
	if err != nil {
		return "", err
	}
	saveSummaryCache(cachePath, summary, len(events))
	return summary, nil
}

// SummarizeWeek generates (or returns cached) a summary of the past 7 days.
func SummarizeWeek(cfg Config, dataDir string, totalTasks int) (string, error) {
	apiKey := resolveApiKey(cfg)
	if apiKey == "" {
		return "", errors.New("no API key — set ANTHROPIC_API_KEY or run `t config set api_key <key>`")
	}

	now := time.Now()
	year, week := now.ISOWeek()
	isoWeek := fmt.Sprintf("%d-W%02d", year, week)
	cachePath := weeklyCachePath(dataDir, isoWeek)

	// count all events this week for cache key
	totalEvents := 0
	var sb strings.Builder
	sb.WriteString("Here are my work logs for the past 7 days. Write a short paragraph (4-6 sentences) summarizing my week: what I focused on, what patterns you notice, and anything worth reflecting on. Plain language, no bullet points.\n\n")

	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		key := day.Format("2006-01-02")
		events, err := ReadEvents(dataDir, key)
		if err != nil {
			continue
		}
		totalEvents += len(events)
		if len(events) == 0 {
			sb.WriteString(fmt.Sprintf("  %s: no activity\n", key))
			continue
		}
		checked := 0
		var notes []string
		for _, e := range events {
			if e.Type == EventCheck {
				checked++
				if e.Log != "" {
					notes = append(notes, e.Log)
				}
			} else if e.Type == EventFree {
				notes = append(notes, e.Log)
			}
		}
		sb.WriteString(fmt.Sprintf("  %s: %d/%d tasks checked", key, checked, totalTasks))
		if len(notes) > 0 {
			sb.WriteString(", logs: " + strings.Join(notes, "; "))
		}
		sb.WriteString("\n")
	}

	if cached, ok := loadSummaryCache(cachePath, totalEvents); ok {
		return cached, nil
	}

	model := resolveModel(cfg)
	baseURL := resolveBaseURL(cfg)
	summary, err := callLLM(apiKey, baseURL, model, sb.String())
	if err != nil {
		return "", err
	}
	saveSummaryCache(cachePath, summary, totalEvents)
	return summary, nil
}
