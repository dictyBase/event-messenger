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
