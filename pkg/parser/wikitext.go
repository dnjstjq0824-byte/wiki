package parser

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

// Heading represents an extracted section heading
type Heading struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
	Slug  string `json:"slug"`
}

// ParseResult holds the parsed HTML, TOC headings, categories, and references
type ParseResult struct {
	HTML       string    `json:"html"`
	Headings   []Heading `json:"headings"`
	Categories []string  `json:"categories"`
	References []string  `json:"references"`
}

var (
	linkRegex     = regexp.MustCompile(`\[\[([^\]|]+)(?:\|([^\]]+))?\]\]`)
	categoryRegex = regexp.MustCompile(`\[\[(?:Category|분류):([^\]]+)\]\]`)
	boldRegex     = regexp.MustCompile(`'''(.*?)'''|\*\*(.*?)\*\*`)
	italicRegex   = regexp.MustCompile(`''(.*?)''|\*(.*?)\*`)
	refRegex      = regexp.MustCompile(`<ref>(.*?)</ref>`)
	h3Regex       = regexp.MustCompile(`(?m)^===\s*(.*?)\s*===$|^###\s*(.*?)$`)
	h2Regex       = regexp.MustCompile(`(?m)^==\s*(.*?)\s*==$|^##\s*(.*?)$`)
	h1Regex       = regexp.MustCompile(`(?m)^=\s*(.*?)\s*=$|^#\s*(.*?)$`)
)

// ParseWikitext transforms wikitext/markdown into sanitized HTML with TOC and categories
func ParseWikitext(input string) ParseResult {
	var categories []string
	var references []string
	var headings []Heading

	// Extract and remove Categories
	text := categoryRegex.ReplaceAllStringFunc(input, func(m string) string {
		sub := categoryRegex.FindStringSubmatch(m)
		if len(sub) > 1 {
			categories = append(categories, strings.TrimSpace(sub[1]))
		}
		return ""
	})

	// Extract References
	text = refRegex.ReplaceAllStringFunc(text, func(m string) string {
		sub := refRegex.FindStringSubmatch(m)
		if len(sub) > 1 {
			refIdx := len(references) + 1
			references = append(references, strings.TrimSpace(sub[1]))
			return fmt.Sprintf(`<sup><a href="#cite_note-%d" id="cite_ref-%d">[%d]</a></sup>`, refIdx, refIdx, refIdx)
		}
		return m
	})

	// Parse Headings
	lines := strings.Split(text, "\n")
	var outLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if m := h3Regex.FindStringSubmatch(trimmed); len(m) > 0 {
			hText := m[1]
			if hText == "" {
				hText = m[2]
			}
			slug := slugify(hText)
			headings = append(headings, Heading{Level: 3, Text: hText, Slug: slug})
			outLines = append(outLines, fmt.Sprintf(`<h3 id="%s">%s</h3>`, slug, html.EscapeString(hText)))
			continue
		}
		if m := h2Regex.FindStringSubmatch(trimmed); len(m) > 0 {
			hText := m[1]
			if hText == "" {
				hText = m[2]
			}
			slug := slugify(hText)
			headings = append(headings, Heading{Level: 2, Text: hText, Slug: slug})
			outLines = append(outLines, fmt.Sprintf(`<h2 id="%s">%s</h2>`, slug, html.EscapeString(hText)))
			continue
		}
		if m := h1Regex.FindStringSubmatch(trimmed); len(m) > 0 {
			hText := m[1]
			if hText == "" {
				hText = m[2]
			}
			slug := slugify(hText)
			headings = append(headings, Heading{Level: 1, Text: hText, Slug: slug})
			outLines = append(outLines, fmt.Sprintf(`<h1 id="%s">%s</h1>`, slug, html.EscapeString(hText)))
			continue
		}
		outLines = append(outLines, line)
	}

	body := strings.Join(outLines, "\n")

	// Parse Internal Links [[Target|Text]]
	body = linkRegex.ReplaceAllStringFunc(body, func(m string) string {
		sub := linkRegex.FindStringSubmatch(m)
		target := strings.TrimSpace(sub[1])
		label := target
		if len(sub) > 2 && sub[2] != "" {
			label = strings.TrimSpace(sub[2])
		}
		return fmt.Sprintf(`<a href="/w/%s" class="wiki-link">%s</a>`, slugify(target), html.EscapeString(label))
	})

	// Bold and Italics
	body = boldRegex.ReplaceAllStringFunc(body, func(m string) string {
		sub := boldRegex.FindStringSubmatch(m)
		val := sub[1]
		if val == "" {
			val = sub[2]
		}
		return fmt.Sprintf(`<strong>%s</strong>`, html.EscapeString(val))
	})

	body = italicRegex.ReplaceAllStringFunc(body, func(m string) string {
		sub := italicRegex.FindStringSubmatch(m)
		val := sub[1]
		if val == "" {
			val = sub[2]
		}
		return fmt.Sprintf(`<em>%s</em>`, html.EscapeString(val))
	})

	// Append Footnotes Section if any
	if len(references) > 0 {
		body += "\n\n<section class=\"wiki-references\"><h2>References</h2><ol>"
		for i, ref := range references {
			idx := i + 1
			body += fmt.Sprintf(`<li id="cite_note-%d"><a href="#cite_ref-%d">^</a> %s</li>`, idx, idx, html.EscapeString(ref))
		}
		body += "</ol></section>"
	}

	return ParseResult{
		HTML:       body,
		Headings:   headings,
		Categories: categories,
		References: references,
	}
}

func slugify(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, " ", "_")
	return s
}
