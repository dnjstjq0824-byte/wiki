package parser

import (
	"strings"
	"testing"
)

func TestParseWikitext(t *testing.T) {
	input := `== Introduction ==
This is a '''bold''' statement and an ''italic'' note.
Here is a link to [[Quantum Computing|Quantum]].
Also here is a citation<ref>Nature Journal 2026</ref>.
[[Category:Science]]
[[분류:컴퓨터 과학]]
`

	res := ParseWikitext(input)

	if len(res.Headings) != 1 || res.Headings[0].Text != "Introduction" {
		t.Errorf("Headings mismatch: %+v", res.Headings)
	}

	if len(res.Categories) != 2 || res.Categories[0] != "Science" || res.Categories[1] != "컴퓨터 과학" {
		t.Errorf("Categories mismatch: %+v", res.Categories)
	}

	if len(res.References) != 1 || res.References[0] != "Nature Journal 2026" {
		t.Errorf("References mismatch: %+v", res.References)
	}

	if !strings.Contains(res.HTML, `<strong>bold</strong>`) {
		t.Errorf("Expected bold HTML, got: %s", res.HTML)
	}

	if !strings.Contains(res.HTML, `<a href="/w/quantum_computing" class="wiki-link">Quantum</a>`) {
		t.Errorf("Expected wiki-link HTML, got: %s", res.HTML)
	}

	if !strings.Contains(res.HTML, `Nature Journal 2026`) {
		t.Errorf("Expected reference in HTML, got: %s", res.HTML)
	}
}
