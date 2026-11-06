package parser

import (
	"regexp"
	"strings"
)

// SanitizeHTML neutralizes dangerous raw HTML in parsed article output.
//
// ParseWikitext escapes every markup-derived fragment it generates, but plain
// text lines pass through verbatim so editors can use a small set of raw tags.
// This filter runs last and removes executable content while preserving the
// safe elements the parser emits (a/h1-h3/strong/em/table/details/div/span/
// sup/ol/ul/li/mark...):
//
//   - script, iframe, object, embed, form, style, link, meta, base elements
//   - on* event handler attributes
//   - inline style attributes (overlay/phishing defense)
//   - javascript: URLs in href/src attributes
var (
	// Block elements with content (RE2 has no backreferences, so any listed
	// closing tag ends the match — over-stripping fails closed on hostile input).
	sanitizeBlockRe  = regexp.MustCompile(`(?is)<(script|iframe|object|embed|form|style)\b[^>]*>.*?</(script|iframe|object|embed|form|style)\s*>`)
	sanitizeLoneRe   = regexp.MustCompile(`(?i)<(script|iframe|object|embed|form|style|link|meta|base)\b[^>]*>`)
	sanitizeEventsRe = regexp.MustCompile(`(?i)\s+on[a-z]+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	sanitizeStyleRe  = regexp.MustCompile(`(?i)\s+style\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	sanitizeJSURLRe  = regexp.MustCompile(`(?i)(href|src)\s*=\s*(["']?)\s*javascript:`)
)

// SanitizeHTML strips executable HTML while keeping safe article markup.
func SanitizeHTML(body string) string {
	out := sanitizeBlockRe.ReplaceAllString(body, "")
	out = sanitizeLoneRe.ReplaceAllString(out, "")
	out = sanitizeEventsRe.ReplaceAllString(out, "")
	out = sanitizeStyleRe.ReplaceAllString(out, "")
	out = sanitizeJSURLRe.ReplaceAllString(out, "${1}=${2}#blocked:")
	return strings.TrimSpace(out)
}
