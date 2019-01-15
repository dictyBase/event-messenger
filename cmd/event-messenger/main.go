package main

import (
	"os"

	"github.com/dictyBase/event-messenger/internal/app/app/github"
	"github.com/dictyBase/event-messenger/internal/app/app/gmail"
	"github.com/dictyBase/event-messenger/internal/app/validate"

	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Version = "1.0.0"
	app.Name = "event-messenger"
	app.Usage = "Handle events from nats messaging"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log-format",
			Usage: "format of the logging out, either of json or text.",
			Value: "json",
		},
		cli.StringFlag{
			Name:  "log-level",
			Usage: "log level for the application",
			Value: "error",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "gh-issue",
			Usage:  "creates a github issue when a new stock order comes through",
			Action: github.RunCreateIssue,
			Before: validate.GithubIssueArgs,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "gh-token, ght",
					Usage: "Github personal access token file",
				},
				cli.StringFlag{
					Name:  "repository, r",
					Usage: "Github repository",
				},
				cli.StringFlag{
					Name:  "owner",
					Usage: "Github repository owner",
				},
				cli.StringFlag{
					Name:  "subject",
					Usage: "Subject name for nats subscription",
				},
				cli.StringFlag{
					Name:   "nats-host",
					EnvVar: "NATS_SERVICE_HOST",
					Usage:  "nats messaging server host",
				},
				cli.StringFlag{
					Name:   "nats-port",
					EnvVar: "NATS_SERVICE_PORT",
					Usage:  "nats messaging server port",
				},
			},
		},
		{
			Name:   "send-email",
			Usage:  "sends an email when a new stock order comes through",
			Action: gmail.RunSendEmail,
			Before: validate.SendEmailArgs,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "subject",
					Usage: "Subject name for nats subscription",
				},
				cli.StringFlag{
					Name:   "nats-host",
					EnvVar: "NATS_SERVICE_HOST",
					Usage:  "nats messaging server host",
				},
				cli.StringFlag{
					Name:   "nats-port",
					EnvVar: "NATS_SERVICE_PORT",
					Usage:  "nats messaging server port",
				},
				cli.StringFlag{
					Name:   "gmail-secret, gs",
					EnvVar: "GMAIL_CREDENTIALS_FILE",
					Usage:  "location of gmail client secret json file, defaults to ~/.credentials/gmail.json",
				},
				cli.StringFlag{
					Name:   "cache-file, cf",
					EnvVar: "CACHE_TOKEN_FILE",
					Usage:  "location of cached gmail token file",
				},
				cli.StringFlag{
					Name:  "reply-to",
					Usage: "reply-to email address for sent messages",
				},
				cli.StringFlag{
					Name:  "send-to",
					Usage: "email address to send messages to",
				},
			},
		},
	}
	app.Run(os.Args)
}
