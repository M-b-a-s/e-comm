package email

import (
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
)

type Sender struct {
	client *resend.Client
	from   string
}

func NewSender() *Sender {
	apiKey := os.Getenv("RESEND_API_KEY")
	return &Sender{
		client: resend.NewClient(apiKey),
		from:   "Audiophile <onboarding@resend.dev>",
	}
}

func (s *Sender) SendOTP(toEmail, code string) error {
	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{toEmail},
		Subject: "Your verification code",
		Html:    fmt.Sprintf(`<p>Your verification code is: <strong>%s</strong></p><p>This code expires in 10 minutes.</p>`, code),
		Text:    fmt.Sprintf("Your verification code is: %s\nThis code expires in 10 minutes.", code),
	}

	_, err := s.client.Emails.Send(params)
	return err
}
