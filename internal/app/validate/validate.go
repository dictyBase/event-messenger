package validate

import (
	"fmt"

	cli "gopkg.in/urfave/cli.v1"
)

// GithubIssueArgs validates that the necessary flags for creating a Github issue are not missing
func GithubIssueArgs(c *cli.Context) error {
	for _, p := range []string{"repository", "owner", "gh-token", "subject", "nats-host",
		"nats-port"} {
		if len(c.String(p)) == 0 {
			return cli.NewExitError(
				fmt.Sprintf("argument %s is missing", p),
				2,
			)
		}
	}
	return nil
}

// SendEmailArgs validates that the necessary flags for sending an email are not missing
func SendEmailArgs(c *cli.Context) error {
	for _, p := range []string{"subject", "nats-host", "nats-port", "gmail-secret", "cache-file"} {
		if len(c.String(p)) == 0 {
			return cli.NewExitError(
				fmt.Sprintf("argument %s is missing", p),
				2,
			)
		}
	}
	return nil
}
