package parser

import (
	"strings"
	"testing"
)

func TestSanitizeHTMLStripsExecutableContent(t *testing.T) {
	in := `<p>hello</p><script>alert(1)</script><SCRIPT src=x></SCRIPT>` +
		`<img src=x onerror="alert(2)"><img src='y' ONLOAD='alert(3)'>` +
		`<a href="javascript:alert(4)">click</a><a href='JaVaScRiPt:alert(5)'>x</a>` +
		`<iframe src="https://evil.example"></iframe>` +
		`<div style="position:fixed">overlay</div>` +
		`<form action="/steal"><input></form>`
	out := SanitizeHTML(in)

	for _, banned := range []string{"<script", "onerror", "ONLOAD", "onload", "javascript:", "JaVaScRiPt:", "<iframe", "position:fixed", "<form"} {
		if strings.Contains(out, banned) {
			t.Fatalf("output still contains %q: %s", banned, out)
		}
	}
	if !strings.Contains(out, "<p>hello</p>") {
		t.Fatalf("safe markup must survive: %s", out)
	}
	if !strings.Contains(out, "#blocked:") {
		t.Fatalf("javascript: URL must be neutralized: %s", out)
	}
}

func TestParseWikitextSanitizesStoredXSS(t *testing.T) {
	res := ParseWikitext("= Title =\nHello <img src=x onerror=alert(1)> world\n[[Target|Label]] and '''bold'''.")
	if strings.Contains(res.HTML, "onerror") {
		t.Fatalf("parser output contains event handler: %s", res.HTML)
	}
	for _, want := range []string{`<h1 id="title">Title</h1>`, `<a href="/w/target"`, `>Label</a>`, `<strong>bold</strong>`} {
		if !strings.Contains(res.HTML, want) {
			t.Fatalf("expected %q in: %s", want, res.HTML)
		}
	}
}
