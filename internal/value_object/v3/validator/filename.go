package validator

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrFilenameEmpty         = errors.New("filename cannot be empty")
	ErrFilenameTooLong       = errors.New("filename exceeds maximum length of 255 characters")
	ErrFilenamePathTraversal = errors.New("path traversal is not allowed")
	ErrFilenameHiddenFile    = errors.New("hidden files not allowed")
	ErrFilenameInvalidChar   = errors.New("contains invalid characters")
)

// IsValidFilename memvalidasi format filename
// Mengembalikan error jika filename tidak valid
func IsValidFilename(filename string) error {
	// Check if the filename is empty
	if filename == "" {
		return ErrFilenameEmpty
	}

	// Check filename length
	if len(filename) > 255 {
		return fmt.Errorf("%w: exceeds maximum length of %d characters", ErrFilenameTooLong, 255)
	}

	// Prevent path traversal
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") || strings.HasPrefix(filename, "..") {
		return ErrFilenamePathTraversal
	}

	// Prevent hidden files (files starting with a dot)
	if strings.HasPrefix(filename, ".") {
		return ErrFilenameHiddenFile
	}

	// Check each character in the filename for invalid characters
	for _, char := range filename {
		if !isAlphaNumericOrSpecial(char) {
			return ErrFilenameInvalidChar
		}
	}
	return nil
}

func isAlphaNumericOrSpecial(char rune) bool {
	// Validate the character (alphanumeric, dash, underscore, or dot)
	switch {
	case 'A' <= char && char <= 'Z':
		return true
	case 'a' <= char && char <= 'z':
		return true
	case '0' <= char && char <= '9':
		return true
	case char == '-' || char == '_':
		return true
	case char == '.':
		return true
	default:
		return false
	}
}
