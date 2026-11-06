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
	linkRegex        = regexp.MustCompile(`\[\[([^\]|]+)(?:\|([^\]]+))?\]\]`)
	categoryRegex    = regexp.MustCompile(`\[\[(?:Category|분류):([^\]]+)\]\]`)
	boldRegex        = regexp.MustCompile(`'''(.*?)'''|\*\*(.*?)\*\*`)
	italicRegex      = regexp.MustCompile(`''(.*?)''|\*(.*?)\*`)
	infoboxRegex     = regexp.MustCompile(`(?s)\{\{Infobox\s*\|?(.*?)\}\}`)
	collapseRegex    = regexp.MustCompile(`\{\{Collapse\|([^|]*)\|([^}]*)\}\}`)
	calloutRegex     = regexp.MustCompile(`(?m)^> \[!(NOTE|WARNING|TIP)\]\s*(.*)$`)
	latexBlockRegex  = regexp.MustCompile(`(?s)\$\$(.*?)\$\$`)
	latexInlineRegex = regexp.MustCompile(`\$(.+?)\$`)
	refRegex         = regexp.MustCompile(`<ref>(.*?)</ref>`)
	h3Regex          = regexp.MustCompile(`(?m)^===\s*(.*?)\s*===$|^###\s*(.*?)$`)
	h2Regex          = regexp.MustCompile(`(?m)^==\s*(.*?)\s*==$|^##\s*(.*?)$`)
	h1Regex          = regexp.MustCompile(`(?m)^=\s*(.*?)\s*=$|^#\s*(.*?)$`)
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

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Table parsing
		if strings.HasPrefix(trimmed, "|") {
			var rows [][]string
			j := i
			for j < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[j]), "|") {
				rows = append(rows, parseTableRow(lines[j]))
				j++
			}
			outLines = append(outLines, renderTable(rows))
			i = j - 1
			continue
		}
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

	// Infobox template
	body = infoboxRegex.ReplaceAllStringFunc(body, func(m string) string {
		sub := infoboxRegex.FindStringSubmatch(m)
		if len(sub) < 2 {
			return m
		}
		content := sub[1]
		lines := strings.Split(content, "\n")
		var sb strings.Builder
		sb.WriteString(`<table class="infobox"><tbody>`)
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l == "" {
				continue
			}
			parts := strings.SplitN(l, "=", 2)
			if len(parts) == 2 {
				k := strings.TrimSpace(parts[0])
				v := strings.TrimSpace(parts[1])
				sb.WriteString(`<tr><th>` + html.EscapeString(k) + `</th><td>` + html.EscapeString(v) + `</td></tr>`)
			}
		}
		sb.WriteString(`</tbody></table>`)
		return sb.String()
	})

	// Collapse boxes
	body = collapseRegex.ReplaceAllStringFunc(body, func(m string) string {
		sub := collapseRegex.FindStringSubmatch(m)
		if len(sub) < 3 {
			return m
		}
		title := strings.TrimSpace(sub[1])
		content := strings.TrimSpace(sub[2])
		return `<details class="collapsible"><summary>` + html.EscapeString(title) + `</summary><div>` + html.EscapeString(content) + `</div></details>`
	})

	// Callouts
	body = calloutRegex.ReplaceAllStringFunc(body, func(m string) string {
		sub := calloutRegex.FindStringSubmatch(m)
		if len(sub) < 3 {
			return m
		}
		kind := strings.ToLower(sub[1])
		text := sub[2]
		return `<div class="callout callout-` + kind + `">` + html.EscapeString(text) + `</div>`
	})
	if len(references) > 0 {
		// LaTeX blocks
		body = latexBlockRegex.ReplaceAllStringFunc(body, func(m string) string {
			sub := latexBlockRegex.FindStringSubmatch(m)
			if len(sub) < 2 {
				return m
			}
			return `<div class="math-block">` + html.EscapeString(sub[1]) + `</div>`
		})
		body = latexInlineRegex.ReplaceAllStringFunc(body, func(m string) string {
			sub := latexInlineRegex.FindStringSubmatch(m)
			if len(sub) < 2 {
				return m
			}
			return `<span class="math-inline">` + html.EscapeString(sub[1]) + `</span>`
		})
		body += "\n\n<section class=\"wiki-references\"><h2>References</h2><ol>"
		for i, ref := range references {
			idx := i + 1
			body += fmt.Sprintf(`<li id="cite_note-%d"><a href="#cite_ref-%d">^</a> %s</li>`, idx, idx, html.EscapeString(ref))
		}
		body += "</ol></section>"
	}

	return ParseResult{
		HTML:       SanitizeHTML(body),
		Headings:   headings,
		Categories: categories,
		References: references,
	}
}

func parseTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	out := []string{}
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}

func isSeparatorRow(row []string) bool {
	if len(row) == 0 {
		return false
	}
	for _, c := range row {
		c = strings.TrimSpace(c)
		c = strings.Trim(c, ":-")
		if c != "" {
			return false
		}
	}
	// ensure at least one dash
	for _, c := range row {
		if strings.Contains(c, "-") {
			return true
		}
	}
	return false
}

func renderTable(rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("<table class=\"wiki-table\"><thead><tr>")
	// determine if second row is separator
	hasHeader := len(rows) > 1 && isSeparatorRow(rows[1])
	headerRow := rows[0]
	if hasHeader {
		for _, c := range headerRow {
			sb.WriteString("<th>" + escapeHTML(c) + "</th>")
		}
		sb.WriteString("</tr></thead><tbody>")
	} else {
		sb.WriteString("</tr></thead><tbody>")
		// first row as data
		sb.WriteString("<tr>")
		for _, c := range headerRow {
			sb.WriteString("<td>" + escapeHTML(c) + "</td>")
		}
		sb.WriteString("</tr>")
	}
	start := 1
	if hasHeader {
		start = 2
	}
	for i := start; i < len(rows); i++ {
		sb.WriteString("<tr>")
		for _, c := range rows[i] {
			sb.WriteString("<td>" + escapeHTML(c) + "</td>")
		}
		sb.WriteString("</tr>")
	}
	sb.WriteString("</tbody></table>")
	return sb.String()
}

func escapeHTML(s string) string {
	return html.EscapeString(s)
}

func slugify(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, " ", "_")
	return s
}
