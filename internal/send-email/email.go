package email

import (
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
)

// Handler is an interface for handling emails
type Handler interface {
	// SendEmail sends an email when a new stock order is received
	SendEmail(ord *order.Order) error
}
