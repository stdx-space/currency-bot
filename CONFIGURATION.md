# Configuration Reference

currency-bot resolves configuration in this priority order (highest wins):

1. CLI flag
2. Environment variable
3. Config file (`~/.currency-bot.yaml` or `--config <path>`)
4. Built-in default

## Flags & Environment Variables

### Global (all commands)

| Flag | Env var | Default | Description |
|---|---|---|---|
| `--source` | `CURRENCY_BOT_SOURCE` | `CAD` | Source currency code |
| `--target` | `CURRENCY_BOT_TARGET` | `HKD` | Target currency code |
| `--reference-amount` | `CURRENCY_BOT_REFERENCE_AMOUNT` | `294` | Reference amount in target currency shown in the alert (e.g. `294` displays as `294 CAD ≈ X HKD`) |
| `--webhook-url` | `CURRENCY_BOT_WEBHOOK_URL` | _(none)_ | Discord webhook URL. Required unless `--dry-run` is set |
| `--alert-mode` | `CURRENCY_BOT_ALERT_MODE` | `always` | When to send an alert. See [Alert Modes](#alert-modes) |
| `--alert-days` | `CURRENCY_BOT_ALERT_DAYS` | `3` | X in "within highest/lowest X days" |
| `--dry-run` | `CURRENCY_BOT_DRY_RUN` | `false` | Print the alert to stdout instead of sending to Discord |
| `--config` | _(n/a)_ | `~/.currency-bot.yaml` | Path to a YAML config file |

### `serve` command only

| Flag | Env var | Default | Description |
|---|---|---|---|
| `--schedule` | `CURRENCY_BOT_SCHEDULE` | `0 9,21 * * *` | Cron expression for the check interval (5-field, standard cron syntax) |

## Alert Modes

| Mode | Behaviour |
|---|---|
| `always` | Always send an alert regardless of where the rate sits in the window |
| `maxima` | Only alert when the current rate is within the top `--alert-days` days of the window |
| `minima` | Only alert when the current rate is within the bottom `--alert-days` days of the window |

The threshold is derived from the **daily maximum rate** for each day in the history window. For example, with `--alert-mode maxima --alert-days 3` and a 30-day window, an alert fires only if the current rate is ≥ the 3rd-highest daily rate observed over the past 30 days.

## Config File

Place a YAML file at `~/.currency-bot.yaml` (or pass `--config <path>`):

```yaml
source: CAD
target: HKD
reference-amount: 294
webhook-url: "https://discord.com/api/webhooks/YOUR_ID/YOUR_TOKEN"
alert-mode: maxima
alert-days: 7
```

For the `serve` command you can also set:

```yaml
schedule: "0 9,21 * * *"
```

## Examples

```bash
# Always alert, CAD→HKD, default 30-day window
currency-bot check --webhook-url https://discord.com/api/webhooks/...

# Alert only near 7-day highs
currency-bot check \
  --alert-mode maxima \
  --alert-days 7 \
  --webhook-url https://discord.com/api/webhooks/...

# Use environment variables (good for containers/CI)
export CURRENCY_BOT_WEBHOOK_URL=https://discord.com/api/webhooks/...
export CURRENCY_BOT_ALERT_MODE=maxima
currency-bot check

# Dry run — inspect output without pinging Discord
currency-bot check --dry-run --alert-mode minima --alert-days 5

# Internal scheduler, check at 08:00 and 20:00 UTC
currency-bot serve --schedule "0 8,20 * * *" --webhook-url ...
```
