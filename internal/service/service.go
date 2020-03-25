package service

import (
	"fmt"

	"github.com/dictyBase/go-genproto/dictybaseapis/annotation"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/dictyBase/go-genproto/dictybaseapis/user"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

func ClientConn(c *cli.Context, names []string) (map[string]*grpc.ClientConn, error) {
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

func UserClient(mc map[string]*grpc.ClientConn) user.UserServiceClient {
	return user.NewUserServiceClient(mc["user"])
}

func StockClient(mc map[string]*grpc.ClientConn) stock.StockServiceClient {
	return stock.NewStockServiceClient(mc["stock"])
}

func AnnoClient(mc map[string]*grpc.ClientConn) annotation.TaggedAnnotationServiceClient {
	return annotation.NewTaggedAnnotationServiceClient(mc["annotation"])
}
