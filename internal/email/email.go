package email

import (
	"regexp"

	"github.com/Uvic-ECSS/Lockers/internal/env"
	"gopkg.in/gomail.v2"
)

var (
	mailDialier *gomail.Dialer
	HostEmail   string
	uvicEmail   = regexp.MustCompile(`(?i)^[a-z0-9._-]{1,64}@uvic\.ca$`)
)

func Initialize() {
	HostEmail = env.Env("EMAIL_HOST_ADDRESS")

	mailDialier = gomail.NewDialer("smtp.gmail.com",
		587,
		HostEmail,
		env.Env("EMAIL_HOST_PASSWORD"))
}

// NOTE: overides the "From" header, will set it to
// $EMAIL_HOST_ADDRESS email address.
func Send(messages ...*gomail.Message) error {
	for _, msg := range messages {
		msg.SetHeader("From", HostEmail)
	}
	return mailDialier.DialAndSend(messages...)
}

func ValidUVicEmail(email string) bool {
	return uvicEmail.MatchString(email)
}
