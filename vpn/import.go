package vpn

import (
	"fmt"
	"strings"
)

func ParseLinksFromText(input string) ImportResult {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ImportResult{}
	}

	if !strings.Contains(trimmed, "://") {
		if decoded, err := decodeBase64(trimmed); err == nil {
			decodedText := strings.TrimSpace(string(decoded))
			if strings.Contains(decodedText, "://") {
				return ParseLinksFromText(decodedText)
			}
		}
	}

	lines := strings.FieldsFunc(trimmed, func(r rune) bool {
		return r == '\n' || r == '\r'
	})
	result := ImportResult{
		Links:  make([]Link, 0, len(lines)),
		Errors: []error{},
	}

	for idx, line := range lines {
		clean := strings.TrimSpace(line)
		if clean == "" || strings.HasPrefix(clean, "#") {
			continue
		}
		link, err := ParseLink(clean)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("line %d: %w", idx+1, err))
			continue
		}
		result.Links = append(result.Links, link)
	}

	return result
}
