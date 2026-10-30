package parser

import (
	"testing"
)

func TestTokenize(t *testing.T) {
	input := `== Heading 1 ==
{{Infobox Scientist}}
Regular text line
[[Internal Link]]
<ref>Source 2026</ref>
| table row
`

	tokens := Tokenize(input)
	if len(tokens) == 0 {
		t.Fatalf("expected tokens, got 0")
	}

	foundHeading := false
	foundTemplate := false
	foundLink := false
	foundRef := false
	foundTable := false

	for _, tok := range tokens {
		switch tok.Type {
		case "heading":
			foundHeading = true
		case "template":
			foundTemplate = true
		case "link":
			foundLink = true
		case "ref":
			foundRef = true
		case "table":
			foundTable = true
		}
	}

	if !foundHeading || !foundTemplate || !foundLink || !foundRef || !foundTable {
		t.Errorf("missing token types: heading=%v, template=%v, link=%v, ref=%v, table=%v",
			foundHeading, foundTemplate, foundLink, foundRef, foundTable)
	}
}

