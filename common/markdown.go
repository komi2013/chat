package common

import (
	"html"
	"net/url"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var (
	userContentPolicy = bluemonday.UGCPolicy()
	customLinkPattern = regexp.MustCompile(`「([^」]*)」（([^）]*)）`)
	markdownLinkPattern = regexp.MustCompile(`(!?\[([^\]]*)\]\()([^)]*)(\))`)
)

// SanitizeRichText removes unsafe HTML and unsafe link destinations from
// Markdown/custom chat text before it is persisted. Rendering clients still
// sanitize their generated HTML as a separate defense.
func SanitizeRichText(source string) string {
	clean := userContentPolicy.Sanitize(source)
	clean = customLinkPattern.ReplaceAllStringFunc(clean, func(match string) string {
		parts := customLinkPattern.FindStringSubmatch(match)
		if len(parts) != 3 || safeRichTextURL(parts[2]) {
			return match
		}
		return parts[1]
	})
	return markdownLinkPattern.ReplaceAllStringFunc(clean, func(match string) string {
		parts := markdownLinkPattern.FindStringSubmatch(match)
		if len(parts) != 5 || safeRichTextURL(parts[3]) {
			return match
		}
		return parts[2]
	})
}

func safeRichTextURL(raw string) bool {
	value := strings.TrimSpace(html.UnescapeString(raw))
	parsed, err := url.Parse(value)
	if err != nil || value == "" || strings.HasPrefix(value, "//") {
		return false
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https", "mailto", "tel":
		return true
	case "":
		return true
	default:
		return false
	}
}
