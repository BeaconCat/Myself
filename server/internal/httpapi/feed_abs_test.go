package httpapi

import "testing"

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
