package httpapi

import (
	"bytes"
	"strings"
	"testing"
)

func TestAbsolutize(t *testing.T) {
	in := `<p><img src="/covers/03.webp" alt=""> <a href="/articles">全部</a> <a href="//cdn.example/x">x</a> <a href="https://a.b/c">y</a> <a href="#h">z</a></p>`
	want := `<p><img src="https://blog.example/covers/03.webp" alt=""> <a href="https://blog.example/articles">全部</a> <a href="//cdn.example/x">x</a> <a href="https://a.b/c">y</a> <a href="#h">z</a></p>`
	if got := absolutize(in, "https://blog.example"); got != want {
		t.Fatalf("got %s", got)
	}
	if got := absolutize(`<a href="/">home</a>`, "https://blog.example"); got != `<a href="https://blog.example/">home</a>` {
		t.Fatalf("root: %s", got)
	}
}

func TestFeedFootnotes(t *testing.T) {
	source := []byte("观点[^source]，再次引用[^source]。\n\n[^source]: [真实来源](https://example.com/source)\n\n    第二段补充说明。\n")
	for _, id := range []int64{7, 8} {
		var buf bytes.Buffer
		if err := renderFeedMarkdown(source, id, &buf); err != nil {
			t.Fatal(err)
		}
		got := buf.String()
		prefix := "post-7-"
		if id == 8 {
			prefix = "post-8-"
		}
		for _, want := range []string{`id="` + prefix + `fn:1"`, `href="#` + prefix + `fn:1"`, `id="` + prefix + `fnref:1"`, `id="` + prefix + `fnref1:1"`, `href="https://example.com/source"`, "第二段补充说明。"} {
			if !strings.Contains(got, want) {
				t.Fatalf("missing %q in %s", want, got)
			}
		}
		if strings.Count(got, `class="footnote-backref"`) != 2 {
			t.Fatalf("expected return links for both references: %s", got)
		}
	}
}

func TestFeedExcerptNewlines(t *testing.T) {
	if got := feedExcerpt("第一行\r\n<script>第二行</script>\n\n第四行"); got != "第一行<br>\n&lt;script&gt;第二行&lt;/script&gt;<br>\n<br>\n第四行" {
		t.Fatalf("unexpected excerpt %q", got)
	}
}
