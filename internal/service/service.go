package service

import (
	"fmt"

	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

func ServiceClients(c *cli.Context, names []string) (map[string]*grpc.ClientConn, error) {
	mc := make(map[string]*grpc.ClientConn)
	for _, n := range names {
		host := fmt.Sprintf("%s-grpc-host", n)
		port := fmt.Sprintf("%s-grpc-port", n)
		conn, err := grpc.Dial(
			fmt.Sprintf("%s:%s", c.String(host), c.String(port)),
			grpc.WithInsecure(),
		)
		if err != nil {
			return mc, fmt.Errorf("error in connecting to %s service %s", n, err)
		}
		mc[n] = conn
	}
	return mc, nil
}
