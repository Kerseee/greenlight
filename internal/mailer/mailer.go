package mailer

import (
	"bytes"
	"embed"
	"text/template"
	"time"

	"github.com/go-mail/mail/v2"
)

//go:embed "templates"
var templateFS embed.FS

const dialerTimeout = 5 * time.Second

// Mailer is a struct that sends emails.
type Mailer struct {
	dialer *mail.Dialer // used to connect to a SMTP server
	sender string       // email address of sender
}

// New creates and returns a mailer.
func New(host string, port int, username, password, sender string) Mailer {
	// Create a dialer.
	dialer := mail.NewDialer(host, port, username, password)
	// Set the timeout of dialer
	dialer.Timeout = dialerTimeout

	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// Send sends an email to receiver with data and template combined.
func (m Mailer) Send(receiver, templateFile string, data interface{}) error {
	// Parse the template.
	tmpl, err := template.ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// Execute the subject template.
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// Execute the plainBody template.
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// Execute the htmlBody template.
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	// Set the messages.
	msg := mail.NewMessage()
	headers := map[string][]string{
		"From":    {m.sender},
		"To":      {receiver},
		"Subject": {subject.String()},
	}
	msg.SetHeaders(headers)
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// Sends the email to the SMTP server.
	err = m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}

	return nil
}
