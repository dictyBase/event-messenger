package client

import (
	"context"

	"github.com/dictyBase/event-messenger/internal/message"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"google.golang.org/grpc"
)

type grpcOrderClient struct {
	client order.OrderServiceClient
}

func NewOrderClient(conn *grpc.ClientConn) message.IssueTracker {
	return &grpcOrderClient{
		client: order.NewOrderServiceClient(conn),
	}
}

func (g *grpcOrderClient) CreateIssue(id string) (*order.Order, error) {
	return g.client.GetOrder(context.Background(), &order.OrderId{Id: id})
}
