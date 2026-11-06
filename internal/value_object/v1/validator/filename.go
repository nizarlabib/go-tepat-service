package validator

import (
	"fmt"
	"net/http"
	"strings"

	"go-starter-service/internal/value_object/v1/constant"
	"go-starter-service/internal/value_object/v1/handling"
)

func IsValidFilename(filename string) error {
	// Check if the filename is empty
	if filename == "" {
		return handling.NewHttpError(nil, http.StatusBadRequest, "invalid filename: filename cannot be empty", constant.ERR_UNKNOWN)
	}

	// Check filename length
	if len(filename) > 255 {
		return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("invalid filename: exceeds maximum length of %d characters", 255), constant.ERR_UNKNOWN)
	}

	// Prevent path traversal
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") || strings.HasPrefix(filename, "..") {
		return handling.NewHttpError(nil, http.StatusBadRequest, "invalid filename: path traversal is not allowed", constant.ERR_UNKNOWN)
	}

	// Prevent hidden files (files starting with a dot)
	if strings.HasPrefix(filename, ".") {
		return handling.NewHttpError(nil, http.StatusBadRequest, "invalid filename: hidden files not allowed", constant.ERR_UNKNOWN)
	}

	// Check each character in the filename for invalid characters
	for _, char := range filename {
		if !isAlphaNumericOrSpecial(char) {
			return handling.NewHttpError(nil, http.StatusBadRequest, "invalid filename: contains invalid characters", constant.ERR_UNKNOWN)
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
