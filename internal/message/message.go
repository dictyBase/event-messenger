package message

import (
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
)

// IssueTracker is the interface for methods related to the Github Issue Tracker.
type IssueTracker interface {
	// CreateIssue creates a new GitHub issue.
	CreateIssue() (*order.Order, error)
}

// NatsSubscriber is an interface to encapsulate the behavior of subscribers.
type NatsSubscriber interface {
	// Start will begin subscription for handling new stock orders.
	Start(string, *order.Order) error
	// Stop will initiate a graceful shutdown of the subscriber connection.
	Stop() error
}
