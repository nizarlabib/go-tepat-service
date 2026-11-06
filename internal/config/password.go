package config

import "time"

const (
	PasswordHistoryLimit = 15                             // New password will be validated with the latest user's password histories
	PasswordDuration     = time.Duration(720) * time.Hour // password will be expired in 30 days
	PasswordEntropy      = 50
)
