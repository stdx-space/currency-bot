package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "currency-bot",
	Short: "Discord bot that alerts on notable exchange rates via Wise",
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $HOME/.currency-bot.yaml)")

	// Shared flags available to all sub-commands.
	rootCmd.PersistentFlags().String("source", "CAD", "source currency code")
	rootCmd.PersistentFlags().String("target", "HKD", "target currency code")
	rootCmd.PersistentFlags().Float64("reference-amount", 294, "reference amount in target currency shown in the alert (e.g. 294 displays as '294 CAD ≈ X HKD')")
	rootCmd.PersistentFlags().String("webhook-url", "", "Discord webhook URL (required unless --dry-run)")
	rootCmd.PersistentFlags().String("alert-mode", "always", "alert mode: always | maxima | minima")
	rootCmd.PersistentFlags().Int("alert-days", 3, "X in 'within highest/lowest X days'")
	rootCmd.PersistentFlags().Bool("dry-run", false, "print alert to stdout instead of sending to Discord")

	viper.BindPFlag("source", rootCmd.PersistentFlags().Lookup("source"))
	viper.BindPFlag("target", rootCmd.PersistentFlags().Lookup("target"))
	viper.BindPFlag("length", rootCmd.PersistentFlags().Lookup("length"))
	viper.BindPFlag("webhook-url", rootCmd.PersistentFlags().Lookup("webhook-url"))
	viper.BindPFlag("reference-amount", rootCmd.PersistentFlags().Lookup("reference-amount"))
	viper.BindPFlag("alert-days", rootCmd.PersistentFlags().Lookup("alert-days"))
	viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run"))

	viper.SetDefault("source", "CAD")
	viper.SetDefault("target", "HKD")
	viper.SetDefault("length", 30)
	viper.SetDefault("reference-amount", 294)
	viper.SetDefault("alert-days", 3)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".currency-bot")
	}

	// Allow env vars like CURRENCY_BOT_WEBHOOK_URL to override flags.
	viper.SetEnvPrefix("CURRENCY_BOT")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
