# t — Daily Work Tracker CLI: Design Spec

**Date:** 2026-08-19  
**Author:** Dante  
**Status:** Approved

---

## Purpose

A personal productivity CLI that forces honest output tracking. The core insight: a checkbox you can silently tick is a lie detector you always pass. A checkbox that costs you one sentence of evidence does not.

Designed for daily recurring tasks that reset each cycle. Built for one user, open-sourced.

---

## Guiding Constraints

- **SLC:** Simple, Loveable, Complete. Ship one tight thing.
- No gamification, no streaks, no scores — ratios are data, not rewards.
- No daemon, no scheduler, no background process.
- Append-only log files: human-readable without the tool, `grep`-able, AI-ingestible later.
- Single static binary. No runtime dependency.

---

## Runtime

**Go.** Compiles to a single `.exe` (Windows), macOS binary, Linux binary. Distributed via GitHub Releases and `go install`. Minimal external dependencies — prefer stdlib; allow one package for terminal color/ANSI if needed.

---

## Architecture

### Day Boundary

Every invocation computes the current **day-key** (a `YYYY-MM-DD` string) using local time adjusted by `day_start_hour` (default: `4`). A "day" runs from `04:00` local to `03:59` the next calendar day. This prevents midnight false-rollovers for late-night workers.

```
day_key = local_date  if local_hour >= day_start_hour
        = local_date - 1  otherwise
```

### Lazy Rollover

On every invocation, the binary reads `state.json` and compares its `day_key` against the computed current key. If they differ, it resets `checked` to an empty map and writes the new key. No scheduler required. Works correctly after multi-day gaps.

---

## Storage

All data lives under `~/.local/share/t/` (overridable via `T_DATA_DIR` env var).

```
~/.local/share/t/
  tasks.json          # ordered array of task name strings
  config.json         # user configuration
  state.json          # today's check state (ephemeral, reset on rollover)
  logs/
    2026-08-19.jsonl  # append-only event log, one JSON object per line
    2026-08-18.jsonl
    …
```

### tasks.json

```json
[
  "Ship one Trace feature",
  "Record build-in-public clip",
  "Outreach — 5 prospects",
  "Deep work block (2h)",
  "Review inbox to zero",
  "Read 20 pages"
]
```

Edited directly by the user via `t tasks` (opens `$EDITOR`). Task identity within a day is its 0-based index. **Do not reorder or remove tasks mid-day** (while any are checked) — doing so will cause the checked state to point to wrong items. Safe to edit at the start of a new day before checking anything.

### config.json

```json
{
  "day_start_hour": 4
}
```

### state.json

```json
{
  "day_key": "2026-08-19",
  "checked": {
    "0": "2026-08-19T09:14:23-03:00",
    "3": "2026-08-19T14:40:00-03:00"
  }
}
```

`checked` keys are task indices (as strings). Values are RFC3339 timestamps. Reset to `{}` on rollover; never archived (JSONL is the archive).

### logs/YYYY-MM-DD.jsonl

One JSON object per line, append-only. Three event types:

```jsonl
{"ts":"2026-08-19T09:14:23-03:00","type":"check","task":0,"log":"shipped the dark mode toggle"}
{"ts":"2026-08-19T14:32:00-03:00","type":"free","log":"fixed the double-write bug in session store"}
{"ts":"2026-08-19T15:00:00-03:00","type":"uncheck","task":0}
```

- **`check`** — task checked; `task` is 0-based index; `log` is the required one-line note.
- **`free`** — free-form note not attached to any task.
- **`uncheck`** — task unchecked; no log field.

---

## Commands

### `t` — Board view (default)

Prints today's board. Checked items are dimmed; unchecked items full-brightness. Time shown next to checked items.

```
  TODAY  ·  Wed 19 Aug  ·  3/6

  1  ✓  Ship one Trace feature      09:14
  2  ✓  Record build-in-public clip 11:02
  3  ·  Outreach — 5 prospects
  4  ✓  Deep work block (2h)        14:40
  5  ·  Review inbox to zero
  6  ·  Read 20 pages
```

### `t N` — Check an item

If unchecked: shows task name, prompts for log inline, appends events, reprints board.

```
$ t 3
  Outreach — 5 prospects
  log › sent 5 WhatsApp messages, got 2 replies back
  ✓
```

If already checked: shows current state, confirms before unchecking (default: No).

```
$ t 1
  Ship one Trace feature  ✓  09:14
  uncheck? [y/N]
```

If confirmed: appends `uncheck` event to JSONL, removes entry from `state.json checked`, reprints board.

### `t log "note"` — Free-form log

Records a note not attached to any task. Appends a `free` event to today's JSONL. All arguments after `log` are joined with spaces — no quoting required.

```
$ t log fixed the double-write bug in session store
  ✓ logged 14:32
```

If called with no arguments (`t log`), prompts inline: `log › `.

### `t today` — Verbose board

Board view with log lines shown under each checked task. Free-form logs shown at the bottom, separated by a rule.

```
  TODAY  ·  Wed 19 Aug  ·  3/6

  1  ✓  Ship one Trace feature      09:14
       → shipped the dark mode toggle

  2  ✓  Record build-in-public clip 11:02
       → recorded the session store refactor walkthrough

  3  ·  Outreach — 5 prospects
  4  ✓  Deep work block (2h)        14:40
       → 2h on Trace auth flow, got token refresh working

  5  ·  Review inbox to zero
  6  ·  Read 20 pages

  ────────────────────────
  14:32  fixed the double-write bug in session store
```

### `t history` — 30-day grid

One character per day. Three states. No numbers, no streaks.

```
  last 30 days  (■ all  ▪ partial  · none)

  · ■ ■ ■ ▪ · · ■ ■ ■ ■ ▪ · · ■ ■ ■ ■ ■ · · ■ ■ ▪ ■ ■ ■ · ■
```

A day with no tasks configured shows `·`.

### `t tasks` — Edit task list

Opens `~/.local/share/t/tasks.json` in `$EDITOR`. On Windows, falls back to `notepad` if `$EDITOR` is unset.

### `t config` — Show config

Prints current config values.

```
  day_start_hour  4
  data_dir        C:\Users\estud\.local\share\t
```

---

## Visual Design Principles

- **Checked items:** dimmed (ANSI dim/faint)
- **Unchecked items:** normal weight
- **Check mark `✓`:** green
- **Times:** dim
- **Header `·` separators:** dim
- **The ratio `3/6`:** plain, no color — it is data
- **`NO_COLOR` env:** respected — all ANSI suppressed, ASCII-only output
- **Nerd Fonts not required** — all glyphs are standard Unicode or ASCII fallback

---

## Error Handling

- `tasks.json` missing or empty → print a one-line prompt to run `t tasks` and exit.
- `state.json` corrupted → treat as empty (log a warning, continue).
- `N` out of range → print error, exit 1.
- `log` prompt receives empty input → re-prompt once, then abort check without writing.
- `$EDITOR` unset on non-Windows → fall back to `vi`.

---

## Out of Scope for v1

- `t summary` — AI daily summary (v2: reads today's JSONL, sends to Claude/OpenAI API with user-supplied key, caches result keyed on log count to avoid re-charging)
- `t week` — weekly summary
- Task categories or tags
- Multiple boards / contexts
- Config editing via CLI (edit `config.json` directly for now)

---

## File Structure (Go project)

```
t-cli/
  main.go           # entry point, command dispatch
  board.go          # board rendering
  state.go          # state.json read/write, rollover logic
  log.go            # JSONL append, read
  config.go         # config.json + tasks.json load
  history.go        # 30-day grid computation
  color.go          # ANSI helpers, NO_COLOR detection
  go.mod
  go.sum
```

Each file has one clear purpose. Total estimated: 500–700 lines.

---

## Install

```sh
go install github.com/dante/t@latest
```

Or download the binary from GitHub Releases and add to PATH. No runtime required.
