// Package mailer for managing sending mail functionalities.
package mailer

import "embed"

const (
	FromName            = "GopherSocial"
	maxRetries          = 3
	UserWelcomeTemplate = "user_invitation.html"
)

//go:embed "templates"
var FS embed.FS

type Client interface {
	Send(templateFile, username, email string, data any, isSandbox bool) error
}
