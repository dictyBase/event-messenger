package message

import issue "github.com/dictyBase/event-messenger/internal/issue-tracker"

// Subscriber is an interface to encapsulate the behavior of subscribers.
type Subscriber interface {
	// Start will begin subscription for creating Github issues for new stock orders.
	Start(string, issue.IssueTracker) error
	// Stop will initiate a graceful shutdown of the subscriber connection.
	Stop() error
}
