package main

import (
	"os"

	"github.com/dictyBase/event-messenger/internal/app/validate"
	"github.com/dictyBase/event-messenger/internal/issue-tracker/github"

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
			Action: github.CreateIssue,
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
					Name:   "order-grpc-host",
					EnvVar: "ORDER_API_SERVICE_HOST",
					Usage:  "grpc host address for order service",
				},
				cli.StringFlag{
					Name:   "order-grpc-port",
					EnvVar: "ORDER_API_SERVICE_PORT",
					Usage:  "grpc port for order service",
				},
			},
		},
	}
	app.Run(os.Args)
}
