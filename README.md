<div align="center">

# checkygo

**a daily work tracker that makes you prove it**

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-slate?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey?style=flat-square)]()

</div>

---

> A checkbox you can silently tick is a lie detector you always pass.
> A checkbox that costs you one sentence of evidence does not.

`checkygo` (command: `t`) is a minimal CLI for people who want an honest record of what they actually did each day — not just what they meant to do. Every task check requires a one-line log. No log, no check.

---

<!-- Screenshot placeholder — replace with a real terminal recording -->
<div align="center">
  <img src="docs/assets/demo.png" alt="checkygo board screenshot" width="600"/>
  <br/>
  <sub><i>screenshot coming soon</i></sub>
</div>

---

## philosophy

- **No gamification.** `3/6` is data, not a score. No streaks, no flames, no "great job."
- **No daemon.** No background process. Every invocation is self-contained. Works after a three-day gap.
- **Your logs are yours.** Everything is stored as plain JSON and append-only JSONL. Readable with `grep`. Ingestible by anything.
- **One rule.** Check a task → write one line about what you actually did.

---

## install

**Via `go install`** (names the binary `checkygo`):
```sh
go install github.com/escdante/checkygo@latest
```

**Build as `t`** (recommended — single character, fast):
```sh
git clone https://github.com/escdante/checkygo
cd checkygo
make build        # produces ./t (or ./t.exe on Windows)
```

Then move the binary somewhere on your `PATH`:
```sh
# macOS / Linux
mv t ~/.local/bin/t

# Windows (PowerShell)
Move-Item t.exe "$env:USERPROFILE\.local\bin\t.exe"
```

**Requirements:** Go 1.23+, no other dependencies.

---

## quick start

```sh
# 1. add your daily recurring tasks
t tasks
# opens your $EDITOR with an empty JSON array — fill it in:
# ["Ship one feature", "Record clip", "Outreach", "Deep work (2h)"]

# 2. view today's board
t

# 3. check a task off (requires a log line)
t 2

# 4. log something that doesn't fit a task
t log fixed the double-write bug in session store

# 5. see everything you did today
t today

# 6. look back at the last 30 days
t history
```

---

## commands

| Command | Description |
|---------|-------------|
| `t` | Print today's board — task list with check status and times |
| `t <N>` | Check task N (1-based). Prompts for a one-line log. Already checked → asks to uncheck. |
| `t log [text]` | Append a free-form note not tied to any task. Args joined; prompts inline if none. |
| `t today` | Verbose board — log lines under each checked task, free notes at the bottom |
| `t history` | 30-day completion grid: `■` all done · `▪` partial · `·` nothing |
| `t tasks` | Open `tasks.json` in `$EDITOR` to edit your recurring task list |
| `t config` | Show current configuration (day start hour, data directory) |

### check flow

```
$ t 3
  Outreach — 5 prospects
  log › sent 5 WhatsApp messages, 2 replies already
  ✓
```

Already checked:
```
$ t 3
  ✓  Outreach — 5 prospects  14:32
  uncheck? [y/N]
```

### board view

```
  TODAY  ·  Wed 19 Aug  ·  3/6

  1  ✓  Ship one Trace feature      09:14
  2  ✓  Record build-in-public clip 11:02
  3  ·  Outreach — 5 prospects
  4  ✓  Deep work block (2h)        14:40
  5  ·  Review inbox to zero
  6  ·  Read 20 pages
```

### verbose view (`t today`)

```
  TODAY  ·  Wed 19 Aug  ·  3/6

  1  ✓  Ship one Trace feature      09:14
       → shipped the dark mode toggle

  2  ✓  Record build-in-public clip 11:02
       → recorded the session store refactor walkthrough

  3  ·  Outreach — 5 prospects
  4  ✓  Deep work block (2h)        14:40
       → 2h on auth flow, got token refresh working

  5  ·  Review inbox to zero
  6  ·  Read 20 pages

  ────────────────────────
  16:22  fixed the double-write bug in session store
```

### history grid (`t history`)

```
  last 30 days  (■ all  ▪ partial  · none)

  · ■ ■ ■ ▪ · · ■ ■ ■ ■ ▪ · · ■ ■ ■ ■ ■ · · ■ ■ ▪ ■ ■ ■ · ■
```

---

## data & storage

Everything is stored locally. No account, no sync, no cloud.

| Path | Format | Contents |
|------|--------|----------|
| `{data_dir}/tasks.json` | JSON array | Your recurring daily task list |
| `{data_dir}/config.json` | JSON object | Configuration (day start hour) |
| `{data_dir}/state.json` | JSON object | Today's check state (resets each day) |
| `{data_dir}/logs/YYYY-MM-DD.jsonl` | JSONL | Append-only event log — one object per line |

**Default data directory:**

| Platform | Path |
|----------|------|
| Windows | `%LOCALAPPDATA%\t` |
| macOS / Linux | `~/.local/share/t` |

Override with the `T_DATA_DIR` environment variable.

**Log format** (append-only, human-readable, grep-friendly):

```jsonl
{"ts":"2026-08-19T09:14:23-03:00","type":"check","task":0,"log":"shipped the dark mode toggle"}
{"ts":"2026-08-19T16:22:00-03:00","type":"free","log":"fixed the double-write bug"}
{"ts":"2026-08-19T18:00:00-03:00","type":"uncheck","task":0}
```

Three event types: `check`, `uncheck`, `free`.

---

## configuration

Edit `config.json` in your data directory directly:

```json
{
  "day_start_hour": 4
}
```

| Key | Default | Description |
|-----|---------|-------------|
| `day_start_hour` | `4` | Hour your "day" starts. Work past midnight without losing the day. A 2am check counts as yesterday if your day starts at 4am. |

---

## build

```sh
make build    # build for current platform → ./t (or ./t.exe)
make install  # go install (binary named checkygo)
make test     # run test suite
make cross    # build for Windows, macOS (Intel + ARM), Linux
```

**Languages & tools used:**

| Layer | Technology |
|-------|-----------|
| Binary | Go 1.23+ |
| Data storage | JSON · JSONL |
| Build | GNU Make |
| Logging format | RFC 3339 timestamps |

---

## roadmap

- [ ] `t summary` — AI-generated daily summary via your own API key (Claude / OpenAI). Reads today's JSONL, sends to API once, caches result keyed on log count so it never re-charges for the same day.
- [ ] `t week` — weekly summary that consumes cached daily summaries, not raw logs.

The log format is already structured for both — no schema changes when v2 ships.

---

## important note on tasks.json

**Do not reorder or remove tasks mid-day** (while any are checked). Task identity within a day is its 0-based index. Safe to edit at the start of a new day before checking anything.

---

## license

MIT — do whatever you want with it.

---

<div align="center">
  <sub>built for personal use · open sourced because why not</sub>
</div>
