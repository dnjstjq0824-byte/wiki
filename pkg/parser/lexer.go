package parser

import "strings"

// Token represents a lexical token from wikitext
type Token struct {
	Type  string `json:"type"` // text, tag, link, heading, template, ref, table, newline
	Value string `json:"value"`
}

// Tokenize performs a simple lexical scan of wikitext input
func Tokenize(input string) []Token {
	tokens := []Token{}
	// Simple state machine: split by lines and classify
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			tokens = append(tokens, Token{Type: "newline", Value: ""})
			continue
		}
		if strings.HasPrefix(trimmed, "==") && strings.HasSuffix(trimmed, "==") {
			tokens = append(tokens, Token{Type: "heading", Value: trimmed})
			continue
		}
		if strings.HasPrefix(trimmed, "{{") && strings.HasSuffix(trimmed, "}}") {
			tokens = append(tokens, Token{Type: "template", Value: trimmed})
			continue
		}
		if strings.HasPrefix(trimmed, "<ref>") {
			tokens = append(tokens, Token{Type: "ref", Value: trimmed})
			continue
		}
		if strings.HasPrefix(trimmed, "|") {
			tokens = append(tokens, Token{Type: "table", Value: trimmed})
			continue
		}
		// inline links
		if strings.Contains(trimmed, "[[") {
			tokens = append(tokens, Token{Type: "link", Value: trimmed})
			continue
		}
		tokens = append(tokens, Token{Type: "text", Value: trimmed})
	}
	return tokens
}
