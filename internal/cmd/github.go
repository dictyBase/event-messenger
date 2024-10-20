package cmd

import (
	"github.com/dictyBase/event-messenger/internal/app/github"
	"github.com/urfave/cli"
)

func ghRepoFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:     "token, ght",
			Usage:    "Github personal access token file",
			EnvVar:   "GITHUB_TOKEN",
			Required: true,
		},
		cli.StringFlag{
			Name:     "repository, r",
			Usage:    "Github repository",
			EnvVar:   "GITHUB_REPOSITORY",
			Required: true,
		},
		cli.StringFlag{
			Name:     "owner",
			Usage:    "Github repository owner",
			EnvVar:   "GITHUB_OWNER",
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

func priceFlags() []cli.Flag {
	return []cli.Flag{
		cli.IntFlag{
			Name:  "strain-price",
			Usage: "price of individual strain",
			Value: 30,
		},
		cli.IntFlag{
			Name:  "plasmid-price",
			Usage: "price of individual plasmid",
			Value: 15,
		},
	}
}

func GhIssueCmd() cli.Command {
	flags := ghRepoFlags()
	flags = append(flags, ghNatsFlags()...)
	flags = append(flags, serviceFlags()...)
	flags = append(flags, priceFlags()...)
	return cli.Command{
		Name:   "gh-issue",
		Usage:  "creates a github issue when a new stock order comes through",
		Action: github.RunCreateIssue,
		Flags:  flags,
	}
}
