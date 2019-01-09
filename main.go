package main

import (
	"os"

	"github.com/dictyBase/event-messenger/commands"

	cli "gopkg.in/codegangsta/cli.v1"
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
			Action: commands.CreateIssue,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "gh-token, ght",
					Usage: "github personal access token file, defaults to ~/.credentials/github.json",
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
			},
		},
	}
	app.Run(os.Args)
}
