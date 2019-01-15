package email

import (
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
)

// EmailHandler is an interface for handling emails
type EmailHandler interface {
	// SendEmail sends an email when a new stock order is received
	SendEmail(ord *order.Order) error
}
