package entity

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var EMAIL_REGEXP = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type Auth struct {
	AccessToken  string
	RefreshToken string
	CsrfToken    string
	XCsrfToken   string

	Username string
	Email    string
	Password string
	RoleID   uuid.UUID

	Latitude  string
	Longitude string

	UserID    uint
	SessionID string
}

func (e *Auth) GetUsernameAsEmail() string {
	username := e.GetUsername()
	if e.UsernameIsEmailAddress() {
		return username
	}
	// assume user is trying to insert as username and not email so add @pln.co.id
	return username + "@pln.co.id"
}

func (e *Auth) GetUsername() string {
	username := strings.TrimSpace(e.Username)
	if username == "" {
		username = strings.TrimSpace(e.Email)
	}
	isEmail := EMAIL_REGEXP.MatchString(username)
	if isEmail {
		username = strings.ToLower(username)
	}
	return username
}

func (e *Auth) UsernameIsEmailAddress() bool {
	un := e.GetUsername()
	return EMAIL_REGEXP.MatchString(un)
}

type ResetPassword struct {
	Token    string
	Email    string
	Password string
}
