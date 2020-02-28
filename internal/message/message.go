package message

import (
	issue "github.com/dictyBase/event-messenger/internal/issue-tracker"
	email "github.com/dictyBase/event-messenger/internal/send-email"
)

// Shutdown interface is for handling stopping of connection
type Shutdown interface {
	// Stop will initiate a graceful shutdown of the subscriber connection.
	Stop() error
}

// GithubSubscriber is an interface to encapsulate the behavior of Github subscribers.
type GithubSubscriber interface {
	// Start will begin the subscription for creating Github issues for new stock orders.
	Start(string, issue.IssueTracker) error
}

// GmailSubscriber is an interface to encapsulate the behavior of Gmail subscribers.
type GmailSubscriber interface {
	// Start will begin the subscription for sending an email when a new stock order is received.
	Start(string, email.EmailHandler) error
}
