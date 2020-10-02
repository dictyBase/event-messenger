package cmd

import (
	arangoflag "github.com/dictyBase/arangomanager/command/flag"
	"github.com/dictyBase/event-messenger/internal/app/webhook"
	"github.com/urfave/cli"
)

func WebhookOntoLoadCmd() cli.Command {
	flags := append(
		[]cli.Flag{
			cli.StringFlag{
				Name:  "term-collection",
				Usage: "arangodb collection for storing ontoloy terms",
				Value: "cvterm",
			},
			cli.StringFlag{
				Name:  "rel-collection",
				Usage: "arangodb collection for storing cvterm relationships",
				Value: "cvterm_relationship",
			},
			cli.StringFlag{
				Name:  "cv-collection",
				Usage: "arangodb collection for storing ontology information",
				Value: "cv",
			},
			cli.StringFlag{
				Name:  "obograph",
				Usage: "arangodb named graph for managing ontology graph",
				Value: "obograph",
			},
			cli.StringFlag{
				Name:     "token, t",
				Usage:    "Github personal access token",
				EnvVar:   "GITHUB_TOKEN",
				Required: true,
			},
		},
		arangoflag.ArangodbFlags()...,
	)
	return cli.Command{
		Name:   "start-onto-server",
		Usage:  "starts the webhook server for loading ontologies",
		Flags:  flags,
		Action: webhook.RunOntoServer,
	}
}
