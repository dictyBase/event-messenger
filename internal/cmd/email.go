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
			Name:   "gmail-secret, gs",
			EnvVar: "GMAIL_CREDENTIALS_FILE",
			Usage:  "location of gmail client secret json file, defaults to ~/.credentials/gmail.json",
		},
		cli.StringFlag{
			Name:     "cache-file, cf",
			EnvVar:   "CACHE_TOKEN_FILE",
			Usage:    "location of cached gmail token file",
			Required: true,
		},
		cli.StringFlag{
			Name:     "reply-to",
			Usage:    "reply-to email address for sent messages",
			Required: true,
		},
		cli.StringFlag{
			Name:     "send-to",
			Usage:    "email address to send messages to",
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
