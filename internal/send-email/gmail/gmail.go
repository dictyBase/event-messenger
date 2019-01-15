package gmail

import (
	email "github.com/dictyBase/event-messenger/internal/send-email"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/sirupsen/logrus"
)

type gmailData struct {
	secret string
	logger *logrus.Entry
	// add gmail client?
}

// NewEmailSender acts as a constructor for sending an email through Gmail
func NewEmailSender(secret string, logger *logrus.Entry) email.EmailHandler {
	return &gmailData{secret: secret, logger: logger}
}

func (g *gmailData) SendEmail(ord *order.Order) error {

}
