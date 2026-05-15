package main

import (
	"context"
	"log"
	"os"

	"github.com/dictyBase/event-messenger/internal/cmd"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Version: "1.0.0",
		Name:    "event-messenger",
		Usage:   "Handle events from nats messaging",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "log-format",
				Usage: "format of the logging, either of json or text.",
				Value: "json",
			},
			&cli.StringFlag{
				Name:  "log-level",
				Usage: "log level for the application",
				Value: "error",
			},
		},
		Commands: []*cli.Command{
			cmd.GhIssueCmd(),
			cmd.SendEmailFlags(),
			cmd.WebhookOntoLoadCmd(),
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("error in running command %s", err)
	}
}
