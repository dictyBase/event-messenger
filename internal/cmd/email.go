package cmd

import (
	"github.com/dictyBase/event-messenger/internal/app/mailgun"
	"github.com/urfave/cli/v3"
)

func emailParamFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "subject",
			Usage:    "Subject name for nats subscription",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "domain",
			Usage:    "email domain name",
			Sources:  cli.EnvVars("EMAIL_DOMAIN"),
			Required: true,
		},
		&cli.StringFlag{
			Name:     "apiKey",
			Usage:    "mailgun api key for that domain",
			Sources:  cli.EnvVars("MAILGUN_API_KEY"),
			Required: true,
		},
		&cli.StringFlag{
			Name:     "name",
			Usage:    "full name that will be used in the from header",
			Sources:  cli.EnvVars("EMAIL_SENDER_NAME"),
			Required: true,
		},
		&cli.StringFlag{
			Name:     "sender",
			Usage:    "sender including the domain name",
			Sources:  cli.EnvVars("EMAIL_SENDER"),
			Required: true,
		},
		&cli.StringFlag{
			Name:     "cc",
			Usage:    "email address to use as CC for all sent emails",
			Sources:  cli.EnvVars("EMAIL_CC"),
			Required: true,
		},
	}
}

const (
	defaultStrainPrice  = 30
	defaultPlasmidPrice = 15
)

func datasourceFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "publication-api",
			Aliases:  []string{"pub"},
			Usage:    "publication api endpoint",
			Sources:  cli.EnvVars("PUBLICATION_API_ENDPOINT"),
			Required: true,
		},
		&cli.IntFlag{
			Name:  "strain-price",
			Usage: "price of individual strain",
			Value: defaultStrainPrice,
		},
		&cli.IntFlag{
			Name:  "plasmid-price",
			Usage: "price of individual plasmid",
			Value: defaultPlasmidPrice,
		},
	}
}

func SendEmailFlags() *cli.Command {
	flags := emailParamFlags()
	flags = append(flags, datasourceFlags()...)
	flags = append(flags, ghNatsFlags()...)
	flags = append(flags, serviceFlags()...)

	return &cli.Command{
		Name:   "send-email",
		Usage:  "sends an email when a new stock order comes through",
		Action: mailgun.RunSendEmail,
		Flags:  flags,
	}
}
