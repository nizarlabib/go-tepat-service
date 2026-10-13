package utils

import (
	"reflect"
	"regexp"
	"strings"
	"unicode"
)

func RedactStr(input string) string {
	length := len(input)

	// Handle empty or very short strings
	if length == 0 {
		return ""
	}
	if length <= 3 {
		return "***"
	}

	// For OTP-like numbers (6 digits)
	if length == 6 && isNumeric(input) {
		return input[:2] + "***" + input[5:]
	}

	// For URLs, preserve domain but redact path
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		if idx := strings.Index(input[8:], "/"); idx > 0 {
			domain := input[:8+idx]
			return domain + "/***"
		}
	}

	// For emails, show first char of username and domain structure
	if strings.Contains(input, "@") {
		parts := strings.Split(input, "@")
		if len(parts) == 2 && len(parts[0]) > 0 && len(parts[1]) > 0 {
			username := parts[0]
			domain := parts[1]

			// Take first character of username plus redaction
			usernameRedacted := RedactStr(username)

			// For domain, keep the TLD portion (like .com, .org)
			domainParts := strings.Split(domain, ".")
			if len(domainParts) >= 2 {
				tld := "." + domainParts[len(domainParts)-1]
				domainRedacted := "***" + tld
				return usernameRedacted + "@" + domainRedacted
			}

			return usernameRedacted + "@" + "***"
		}
	}

	// For other strings, show some characters at start and end
	visibleChars := min(max(length/4, 2), 4)

	return input[:visibleChars] + "***" + input[length-visibleChars:]
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func Substr(input string, start int, length int) string {
	asRunes := []rune(input)
	if start >= len(asRunes) {
		return ""
	}
	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}
	return string(asRunes[start : start+length])
}

func StrToOnlyNum(str string) string {
	re := regexp.MustCompile(`[^0-9]`)
	return re.ReplaceAllString(str, "")
}

func CleanSpaces(str string) string {
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(str, " ")
}

func ToSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func GetFieldName(field reflect.StructField, useJsonTag bool) string {
	if useJsonTag {
		tag := field.Tag.Get("json")
		if tag != "" && tag != "-" {
			name := strings.Split(tag, ",")[0]
			if name != "" {
				return name
			}
		}
		return ToSnakeCase(field.Name)
	}
	return field.Name
}
