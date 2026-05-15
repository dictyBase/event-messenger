package issue

import (
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
)

// Tracker is an interface for creating new issues
type Tracker interface {
	// CreateIssue creates an issue when a new stock order is received
	CreateIssue(ord *order.Order) error
}
