package wiki

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// NormalizeTitle applies Unicode NFKC normalization, trims, collapses whitespace,
// replaces spaces with underscores, and capitalizes the first rune.
func NormalizeTitle(raw string) string {
	s := norm.NFKC.String(raw)
	s = strings.TrimSpace(s)
	// Collapse runs of whitespace into a single space
	fields := strings.Fields(s)
	s = strings.Join(fields, " ")
	// MediaWiki style: spaces -> underscores
	s = strings.ReplaceAll(s, " ", "_")
	// Capitalize first rune
	runes := []rune(s)
	if len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
		s = string(runes)
	}
	return s
}

// Slugify returns a URL-friendly slug (lowercase, underscores preserved)
func Slugify(title string) string {
	return strings.ToLower(NormalizeTitle(title))
}

// SplitTitleNamespace splits "Namespace:Title" into namespace name and title
func SplitTitleNamespace(raw string) (string, string) {
	if idx := strings.Index(raw, ":"); idx > 0 {
		ns := raw[:idx]
		rest := raw[idx+1:]
		if _, ok := namespaceNameToID[ns]; ok {
			return ns, rest
		}
	}
	return "", raw
}