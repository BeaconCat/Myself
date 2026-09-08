package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSPAFallback(t *testing.T) {
	h := Handler()
	for _, path := range []string{"/", "/articles/x", "/admin/login", "/index.html", "/assets/nope.js"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code == http.StatusMovedPermanently {
			t.Fatalf("%s: unexpected redirect to %s", path, rec.Header().Get("Location"))
		}
		t.Logf("%s -> %d %s", path, rec.Code, rec.Header().Get("Content-Type"))
	}
}
