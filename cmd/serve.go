package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the bot continuously, checking rates on an internal cron schedule",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().String("schedule", "0 9,21 * * *", "cron expression for the check interval")
	viper.BindPFlag("schedule", serveCmd.Flags().Lookup("schedule"))
	viper.SetDefault("schedule", "0 9,21 * * *")

	rootCmd.AddCommand(serveCmd)
}

func runServe(_ *cobra.Command, _ []string) error {
	schedule := viper.GetString("schedule")
	log.Printf("Starting scheduler with cron: %q", schedule)

	s, err := gocron.NewScheduler()
	if err != nil {
		return err
	}

	_, err = s.NewJob(
		gocron.CronJob(schedule, false),
		gocron.NewTask(func() {
			if err := runCheck(nil, nil); err != nil {
				log.Printf("check error: %v", err)
			}
		}),
	)
	if err != nil {
		return err
	}

	s.Start()
	log.Println("Scheduler running. Press Ctrl+C to stop.")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down…")
	return s.Shutdown()
}
