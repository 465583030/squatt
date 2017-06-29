package topic

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	errInvalidLength           = errors.New("invalid length")
	errInvalidUTF8             = errors.New("invalid utf-8")
	errWildcardNotAllowed      = errors.New("wildcard not allowed")
	errInvalidWildcardLocation = errors.New("wildcard not allowed at this location")
)

// Validate a topic
func Validate(topic string, allowWildcard bool) error {
	if len(topic) < 1 {
		return errInvalidLength
	}
	if !utf8.ValidString(topic) || strings.ContainsRune(topic, '\U00000000') {
		return errInvalidUTF8
	}
	parts := strings.Split(topic, "/")
	for i, part := range parts {
		if strings.ContainsAny(part, "#+") {
			if !allowWildcard {
				return errWildcardNotAllowed
			}
			if part == "#" {
				if i != len(parts)-1 {
					return errInvalidWildcardLocation
				}
				continue
			}
			if part != "+" {
				return errInvalidWildcardLocation
			}
		}
	}
	return nil
}
