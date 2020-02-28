package main

import (
	"github.com/dictyBase/event-messenger/internal/app/github"
	"github.com/dictyBase/event-messenger/internal/app/validate"
	"github.com/urfave/cli"
)

func ghRepoFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:     "gh-token, ght",
			Usage:    "Github personal access token file",
			Required: true,
		},
		cli.StringFlag{
			Name:     "repository, r",
			Usage:    "Github repository",
			Required: true,
		},
		cli.StringFlag{
			Name:     "owner",
			Usage:    "Github repository owner",
			Required: true,
		},
		cli.StringFlag{
			Name:     "subject",
			Usage:    "Subject name for nats subscription",
			Required: true,
		},
	}
}

func ghNatsFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:     "nats-host",
			EnvVar:   "NATS_SERVICE_HOST",
			Usage:    "nats messaging server host",
			Required: true,
		},
		cli.StringFlag{
			Name:     "nats-port",
			EnvVar:   "NATS_SERVICE_PORT",
			Usage:    "nats messaging server port",
			Required: true,
		},
	}
}

func ghIssueCmd() cli.Command {
	flags := ghRepoFlags()
	flags = append(flags, ghNatsFlags()...)
	return cli.Command{
		Name:   "gh-issue",
		Usage:  "creates a github issue when a new stock order comes through",
		Action: github.RunCreateIssue,
		Before: validate.GithubIssueArgs,
		Flags:  flags,
	}
}
