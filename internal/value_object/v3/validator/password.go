package validator

import (
	"regexp"
	"strings"
)

const (
	Terrible = "zero"
	Weak     = "weak"
	Medium   = "medium"
	Strong   = "strong"
)

var (
	lowercaseRegex = regexp.MustCompile(`[a-z]{1,}`)
	uppercaseRegex = regexp.MustCompile(`[A-Z]`)
	symbolRegex    = regexp.MustCompile(`[^0-9a-zA-Z]`)
	digitRegex     = regexp.MustCompile(`[0-9]`)
)

var (
	asdf = "qwertyuiopasdfghjklzxcvbnm"
)

func PasswordStrengthMeter(v string) string {
	minLength := 8

	if len([]rune(v)) < minLength {
		return Terrible
	}

	if isAsdf(v) || isByStep(v) {
		return Terrible
	}

	score := 0

	if lowercaseRegex.MatchString(v) {
		score++
	}
	if uppercaseRegex.MatchString(v) {
		score++
	}
	if symbolRegex.MatchString(v) {
		score++
	}
	if digitRegex.MatchString(v) {
		score++
	}

	if score < 4 {
		return Terrible
	}

	if len([]rune(v)) == minLength {
		return Weak
	}

	if len([]rune(v)) <= (minLength + 5) {
		return Medium
	}

	return Strong
}

// If the password is in the order on keyboard.
func isAsdf(raw string) bool {
	// s in asdf , or reverse in asdf
	rev := reverse(raw)
	return strings.Contains(asdf, raw) || strings.Contains(asdf, rev)
}

// If the password is alphabet step by step.
func isByStep(raw string) bool {
	r := []rune(raw)
	if len(r) < 2 {
		return false
	}
	delta := r[1] - r[0]

	for i, _ := range r {
		if i == 0 {
			continue
		}
		if r[i]-r[i-1] != delta {
			return false
		}
	}

	return true
}

func reverse(raw string) string {
	r := []rune(raw)

	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
