package cmd

import (
	"github.com/dictyBase/event-messenger/internal/app/mailgun"
	"github.com/urfave/cli"
)

func emailParamFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:     "subject",
			Usage:    "Subject name for nats subscription",
			Required: true,
		},
		cli.StringFlag{
			Name:     "domain",
			Usage:    "email domain name",
			EnvVar:   "EMAIL_DOMAIN",
			Required: true,
		},
		cli.StringFlag{
			Name:     "apiKey",
			Usage:    "mailgun api key for that domain",
			EnvVar:   "MAILGUN_API_KEY",
			Required: true,
		},
		cli.StringFlag{
			Name:     "name",
			Usage:    "full name that will be used in the from header",
			EnvVar:   "EMAIL_SENDER_NAME",
			Required: true,
		},
		cli.StringFlag{
			Name:     "sender",
			Usage:    "sender including the domain name",
			EnvVar:   "EMAIL_SENDER",
			Required: true,
		},
	}
}

func datasourceFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:     "publication-api, pub",
			Usage:    "publication api endpoint",
			Required: true,
		},
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

func SendEmailFlags() cli.Command {
	flags := emailParamFlags()
	flags = append(flags, datasourceFlags()...)
	flags = append(flags, ghNatsFlags()...)
	flags = append(flags, serviceFlags()...)
	return cli.Command{
		Name:   "send-email",
		Usage:  "sends an email when a new stock order comes through",
		Action: mailgun.RunSendEmail,
		Flags:  flags,
	}
}
