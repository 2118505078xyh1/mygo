package richtext

import "testing"

// Reproduction for issue #2095: HTML() writes an empty paragraph as
// <p><br></p>, and ParseHTML must read that back as ONE empty paragraph.
func TestEmptyParagraphRoundTrip(t *testing.T) {
	in := TextOf(
		Paragraph{Runs: []Run{{Text: "a"}}},
		Paragraph{},
		Paragraph{Runs: []Run{{Text: "b"}}},
	)
	html := in.HTML()
	out := ParseHTML(html)
	got := out.String()
	want := in.String()
	if got != want {
		t.Fatalf("ParseHTML(HTML()) = %q, want %q\nHTML: %s", got, want, html)
	}
	n := 0
	for range out.Paragraphs() {
		n++
	}
	if n != 3 {
		t.Fatalf("got %d paragraphs, want 3 (HTML: %s)", n, html)
	}

	// A <br> that ends a block adds no line, as in browsers.
	if s := ParseHTML("<p>a<br></p>").String(); s != "a" {
		t.Errorf("<p>a<br></p> = %q, want %q", s, "a")
	}
	// A <br> with content after it still splits the paragraph.
	if s := ParseHTML("<p>a<br>b</p>").String(); s != "a\nb" {
		t.Errorf("<p>a<br>b</p> = %q, want %q", s, "a\nb")
	}
	// A trailing <br> inside an inline wrapper is still a block-ending one.
	if s := ParseHTML("<p>a<span><br></span></p>").String(); s != "a" {
		t.Errorf("<p>a<span><br></span></p> = %q, want %q", s, "a")
	}
}
