package validate

import (
	"fmt"

	cli "gopkg.in/urfave/cli.v1"
)

// GithubIssueArgs validates that the necessary flags are not missing
func GithubIssueArgs(c *cli.Context) error {
	for _, p := range []string{"repository", "owner", "gh-token", "nats-host",
		"nats-port", "order-grpc-host", "order-grpc-port"} {
		if len(c.String(p)) == 0 {
			return cli.NewExitError(
				fmt.Sprintf("argument %s is missing", p),
				2,
			)
		}
	}
	return nil
}
