# currency-bot

A Discord bot that fetches historical exchange rates from [Wise](https://wise.com) and sends an alert when the current rate is notable (near a multi-day high or low).

## Features

- Pulls hourly rate history from the Wise public API
- Detects whether the current rate is within the highest or lowest X days of the window
- Sends a formatted Discord embed via webhook
- **Dry-run mode** — prints the alert to stdout without touching Discord
- Two execution modes:
  - **`check`** — one-off run, suitable for an external cron system
  - **`serve`** — built-in cron scheduler, runs the check on a configurable schedule
- Fully configurable via CLI flags, environment variables, or a YAML config file (powered by [Cobra](https://github.com/spf13/cobra) + [Viper](https://github.com/spf13/viper))

## Installation

```bash
go install github.com/stommydx/narwhl/currency-bot@latest
```

Or build from source:

```bash
git clone ...
cd currency-bot
go build -o currency-bot .
```

## Quick Start

```bash
# One-off check, send to Discord
currency-bot check --webhook-url https://discord.com/api/webhooks/...

# Dry run — no Discord message, just print to stdout
currency-bot check --dry-run

# Only alert when rate is near a 7-day high
currency-bot check --alert-mode maxima --alert-days 7 --webhook-url ...

# Run as an internal daemon (checks at 09:00 and 21:00 daily)
currency-bot serve --webhook-url ...
```

## Usage

See [CONFIGURATION.md](CONFIGURATION.md) for the full flag/env/config-file reference.

## Running Tests

```bash
go test ./...
```
