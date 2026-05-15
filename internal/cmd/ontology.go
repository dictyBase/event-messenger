package cmd

import (
	"github.com/dictyBase/event-messenger/internal/app/webhook"
	"github.com/urfave/cli/v3"
)

func arangodbFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "arangodb-pass",
			Aliases:  []string{"pass"},
			Sources:  cli.EnvVars("ARANGODB_PASS"),
			Usage:    "arangodb database password",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "arangodb-database",
			Aliases:  []string{"db"},
			Sources:  cli.EnvVars("ARANGODB_DATABASE"),
			Usage:    "arangodb database name",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "arangodb-user",
			Aliases:  []string{"user"},
			Sources:  cli.EnvVars("ARANGODB_USER"),
			Usage:    "arangodb database user",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "arangodb-host",
			Aliases:  []string{"host"},
			Value:    "arangodb",
			Sources:  cli.EnvVars("ARANGODB_SERVICE_HOST"),
			Usage:    "arangodb database host",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "arangodb-port",
			Sources: cli.EnvVars("ARANGODB_SERVICE_PORT"),
			Usage:   "arangodb database port",
			Value:   "8529",
		},
		&cli.BoolFlag{
			Name:  "is-secure",
			Usage: "flag for secured or unsecured arangodb endpoint",
			Value: true,
		},
	}
}

func WebhookOntoLoadCmd() *cli.Command {
	flags := append(
		[]cli.Flag{
			&cli.StringFlag{
				Name:  "term-collection",
				Usage: "arangodb collection for storing ontoloy terms",
				Value: "cvterm",
			},
			&cli.StringFlag{
				Name:  "rel-collection",
				Usage: "arangodb collection for storing cvterm relationships",
				Value: "cvterm_relationship",
			},
			&cli.StringFlag{
				Name:  "cv-collection",
				Usage: "arangodb collection for storing ontology information",
				Value: "cv",
			},
			&cli.StringFlag{
				Name:  "obograph",
				Usage: "arangodb named graph for managing ontology graph",
				Value: "obograph",
			},
			&cli.StringFlag{
				Name:     "token",
				Aliases:  []string{"t"},
				Usage:    "Github personal access token",
				Sources:  cli.EnvVars("GITHUB_TOKEN"),
				Required: true,
			},
			&cli.StringFlag{
				Name:     "port",
				Usage:    "port at which the server will run",
				Required: true,
			},
		},
		arangodbFlags()...,
	)

	return &cli.Command{
		Name:   "start-onto-server",
		Usage:  "starts the webhook server for loading ontologies",
		Flags:  flags,
		Action: webhook.RunOntoServer,
	}
}
