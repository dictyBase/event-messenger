package cmd

import (
	"github.com/dictyBase/event-messenger/internal/app/gmail"
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
			Required: true,
		},
		cli.StringFlag{
			Name:     "apiKey",
			Usage:    "mailgun api key for that domain",
			Required: true,
		},
		cli.StringFlag{
			Name:     "name",
			Usage:    "full name that will be used in the from header",
			Required: true,
		},
		cli.StringFlag{
			Name:     "sender",
			Usage:    "sender including the domain name",
			Required: true,
		},
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
	flags = append(flags, ghNatsFlags()...)
	return cli.Command{
		Name:   "send-email",
		Usage:  "sends an email when a new stock order comes through",
		Action: gmail.RunSendEmail,
		Flags:  flags,
	}
}
