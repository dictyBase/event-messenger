package datasource

import (
	"github.com/dictyBase/event-messenger/internal/service"
	grpc "google.golang.org/grpc"
)

type Sources struct {
	AnnoSource  *Annotation
	StockSource *Stock
	PubSource   *Publication
}

func GrpcSources(mc map[string]*grpc.ClientConn) *Sources {
	return &Sources{
		AnnoSource:  &Annotation{Client: service.AnnoClient(mc)},
		StockSource: &Stock{Client: service.StockClient(mc)},
	}
}
