package gmail

import (
	"encoding/base64"
	"log"

	email "github.com/dictyBase/event-messenger/internal/send-email"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/gmail/v1"
)

type gmailData struct {
	secret  string
	replyto string
	sendto  string
	client  *gmail.Service
	logger  *logrus.Entry
}

// NewEmailSender acts as a constructor for sending an email through Gmail.
func NewEmailSender(secret, replyto, sendto string, client *gmail.Service, logger *logrus.Entry) email.EmailHandler {
	return &gmailData{secret: secret, replyto: replyto, sendto: sendto, client: client, logger: logger}
}

// SendEmail sends an email when a new stock order is received.
func (g *gmailData) SendEmail(ord *order.Order) error {
	var message gmail.Message
	messageStr := []byte(
		"From: 'me'\r\n" +
			"reply-to: " + g.replyto + "\r\n" +
			"To: " + g.sendto + "\r\n" +
			"Subject: New DSC Order \r\n" +
			"\r\n" + ord.Data.Attributes.PurchaseOrderNum + "")

	message.Raw = base64.StdEncoding.EncodeToString(messageStr)
	_, err := g.client.Users.Messages.Send("me", &message).Do()
	if err != nil {
		log.Fatalf("Unable to send gmail message: %v", err)
	}
	g.logger.Infof("successfully sent gmail message to %s", g.sendto)
	return nil
}
