package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"myself/server/internal/config"
)

type languageTransport func(*http.Request) (*http.Response, error)

func (f languageTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func languageResponse(value any, status int) *http.Response {
	data, _ := json.Marshal(value)
	return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(string(data)))}
}

func TestGithubLanguagesPaginatesAndWeightsPublicOriginals(t *testing.T) {
	repo := func(name string, fork, private bool, owner string) map[string]any {
		return map[string]any{"name": name, "fork": fork, "private": private, "owner": map[string]string{"login": owner}}
	}
	pageOne := []any{repo("first", false, false, "owner"), repo("secret", false, true, "owner"), repo("elsewhere", false, false, "another")}
	for len(pageOne) < 100 {
		pageOne = append(pageOne, repo(fmt.Sprint(len(pageOne)), true, false, "owner"))
	}
	requests := []string{}
	g := ghFetcher{client: &http.Client{Transport: languageTransport(func(r *http.Request) (*http.Response, error) {
		requests = append(requests, r.URL.RequestURI())
		switch r.URL.Path {
		case "/users/Owner/repos":
			if r.URL.Query().Get("page") == "1" {
				return languageResponse(pageOne, 200), nil
			}
			return languageResponse([]any{repo("second", false, false, "OWNER"), repo("empty", false, false, "Owner")}, 200), nil
		case "/repos/Owner/first/languages":
			return languageResponse(map[string]int{"Go": 100, "Vue": 100}, 200), nil
		case "/repos/Owner/second/languages":
			return languageResponse(map[string]int{"Go": 800}, 200), nil
		case "/repos/Owner/empty/languages":
			return languageResponse(map[string]int{}, 200), nil
		default:
			t.Fatalf("Unexpected repository fetched: %s", r.URL)
			return nil, nil
		}
	})}}
	data, err := g.fetchLanguages("Owner")
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) != 5 || data.Repositories != 3 || data.TotalBytes != 1000 || len(data.Items) != 2 {
		t.Fatalf("Incomplete aggregation: %+v, %v", data, requests)
	}
	if data.Items[0].Name != "Go" || data.Items[0].Percent != 90 || data.Items[1].Percent != 10 {
		t.Fatalf("Incorrect byte weighting: %+v", data.Items)
	}
}

func TestGithubLanguagesNeverReturnsPartialResults(t *testing.T) {
	for _, failure := range []string{"languages", "page"} {
		t.Run(failure, func(t *testing.T) {
			g := ghFetcher{client: &http.Client{Transport: languageTransport(func(r *http.Request) (*http.Response, error) {
				if strings.Contains(r.URL.Path, "/languages") || r.URL.Query().Get("page") == "2" {
					return languageResponse(map[string]string{"message": "rate limited"}, 403), nil
				}
				repos := []any{map[string]any{"name": "first", "fork": failure == "page", "owner": map[string]string{"login": "owner"}}}
				if failure == "page" {
					for len(repos) < 100 {
						repos = append(repos, repos[0])
					}
				}
				return languageResponse(repos, 200), nil
			})}}
			data, err := g.fetchLanguages("owner")
			if err == nil || data != nil {
				t.Fatalf("Partial response accepted: %+v, %v", data, err)
			}
		})
	}
}

func TestGithubLanguagesEmptyAndCancellation(t *testing.T) {
	g := ghFetcher{client: &http.Client{Transport: languageTransport(func(r *http.Request) (*http.Response, error) {
		if err := r.Context().Err(); err != nil {
			return nil, err
		}
		return languageResponse([]any{}, 200), nil
	})}}
	data, err := g.fetchLanguages("owner")
	if err != nil || data.Items == nil || len(data.Items) != 0 || data.TotalBytes != 0 {
		t.Fatalf("Empty account: %+v, %v", data, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g.ctx = ctx
	if _, err := g.fetchLanguages("owner"); err == nil {
		t.Fatal("Request ignored cancellation")
	}
}

func TestGithubLanguagesCacheAndFailureBackoff(t *testing.T) {
	cfg := config.GitHub{Username: "owner", RefreshMinutes: 30}
	key := languageKey(cfg)
	old := &languageData{Username: "owner", Items: []languageShare{{Name: "Go", Bytes: 10, Percent: 100}}}
	s := &Server{ghLanguages: languageCache{key: key, at: time.Now(), data: old}}
	data, stale, err := s.readLanguages(cfg, false)
	if err != nil || stale || data != old {
		t.Fatalf("Fresh cache not reused: %v %v", stale, err)
	}
	s.ghLanguages.failedAt = time.Now()
	data, stale, err = s.readLanguages(cfg, true)
	if err == nil || !stale || data != old {
		t.Fatalf("Last complete snapshot lost after failed sync: %v %v", stale, err)
	}
	s.ghLanguages.data = nil
	data, stale, err = s.readLanguages(cfg, false)
	if err == nil || stale || data != nil {
		t.Fatalf("Failed first fetch fabricated data: %v %v", stale, err)
	}
	cfg.Token = "changed-token"
	if languageKey(cfg) == key {
		t.Fatal("Token change did not invalidate cache")
	}
}
