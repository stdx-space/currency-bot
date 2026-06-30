package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stommydx/narwhl/currency-bot/internal/alert"
	"github.com/stommydx/narwhl/currency-bot/internal/discord"
	"github.com/stommydx/narwhl/currency-bot/internal/wise"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run a one-off exchange rate check and alert if conditions are met",
	RunE:  runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(_ *cobra.Command, _ []string) error {
	source := viper.GetString("source")
	target := viper.GetString("target")
	length := viper.GetInt("length")
	webhookURL := viper.GetString("webhook-url")
	mode := alert.Mode(viper.GetString("alert-mode"))
	alertDays := viper.GetInt("alert-days")
	dryRun := viper.GetBool("dry-run")

	if !dryRun && webhookURL == "" {
		return fmt.Errorf("--webhook-url is required (or set CURRENCY_BOT_WEBHOOK_URL); use --dry-run to skip sending")
	}

	log.Printf("Fetching %s→%s rates (last %d days)…", source, target, length)

	client := wise.NewClient()
	rates, err := client.FetchHistory(source, target, length)
	if err != nil {
		return fmt.Errorf("fetch failed: %w", err)
	}

	result := alert.Evaluate(rates, mode, alertDays)

	log.Printf("Current rate: %.6f | hi-threshold (top %d days): %.6f | lo-threshold (bottom %d days): %.6f",
		result.Current, alertDays, result.ThresholdHi, alertDays, result.ThresholdLo)
	log.Printf("IsMaxima: %v | IsMinima: %v | ShouldAlert: %v", result.IsMaxima, result.IsMinima, result.ShouldAlert)

	if !result.ShouldAlert {
		log.Println("No alert condition met, skipping.")
		return nil
	}

	title, description, color := buildMessage(source, target, result, alertDays)

	if dryRun {
		fmt.Printf("[dry-run] %s\n%s\n", title, description)
		return nil
	}

	if err := discord.Send(webhookURL, title, description, color); err != nil {
		return fmt.Errorf("discord send failed: %w", err)
	}
	log.Println("Alert sent.")
	return nil
}

// buildMessage constructs the Discord embed content.
func buildMessage(source, target string, r alert.Result, alertDays int) (title, description string, color int) {
	pair := fmt.Sprintf("%s → %s", source, target)
	color = 0x5865F2 // default Discord blurple

	var tag string
	if r.IsMaxima && r.IsMinima {
		tag = "📊 Notable Rate"
	} else if r.IsMaxima {
		tag = fmt.Sprintf("📈 Near %d-Day High", alertDays)
		color = 0x57F287 // green
	} else if r.IsMinima {
		tag = fmt.Sprintf("📉 Near %d-Day Low", alertDays)
		color = 0xED4245 // red
	} else {
		tag = "💱 Rate Update"
	}

	title = fmt.Sprintf("%s  %s", tag, pair)
	reciprocal := 1.0 / r.Current
	description = fmt.Sprintf(
		"**Current rate:** `%.6f` (%s per %s) / `%.6f` (%s per %s)\n**%d-day high threshold:** `%.6f`\n**%d-day low threshold:** `%.6f`",
		r.Current, target, source,
		reciprocal, source, target,
		alertDays, r.ThresholdHi,
		alertDays, r.ThresholdLo,
	)
	return
}
