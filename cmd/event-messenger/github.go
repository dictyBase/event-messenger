package main

import (
	"github.com/dictyBase/event-messenger/internal/app/github"
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

func serviceFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:     "stock-grpc-host",
			EnvVar:   "STOCK_API_SERVICE_HOST",
			Usage:    "stock grpc host",
			Required: true,
		},
		cli.StringFlag{
			Name:     "stock-grpc-port",
			EnvVar:   "STOCK_API_SERVICE_PORT",
			Usage:    "stock grpc port",
			Required: true,
		},
		cli.StringFlag{
			Name:     "annotation-grpc-host",
			EnvVar:   "ANNOTATION_API_SERVICE_HOST",
			Usage:    "annotation grpc host",
			Required: true,
		},
		cli.StringFlag{
			Name:     "annotation-grpc-port",
			EnvVar:   "ANNOTATION_API_SERVICE_PORT",
			Usage:    "annotation grpc port",
			Required: true,
		},
		cli.StringFlag{
			Name:     "user-grpc-host",
			EnvVar:   "USER_API_SERVICE_HOST",
			Usage:    "user grpc host",
			Required: true,
		},
		cli.StringFlag{
			Name:     "user-grpc-port",
			EnvVar:   "USER_API_SERVICE_PORT",
			Usage:    "user grpc port",
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
	flags = append(flags, serviceFlags()...)
	return cli.Command{
		Name:   "gh-issue",
		Usage:  "creates a github issue when a new stock order comes through",
		Action: github.RunCreateIssue,
		Flags:  flags,
	}
}
