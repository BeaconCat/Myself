package httpapi

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"myself/server/internal/config"
)

type languageShare struct {
	Name    string  `json:"name"`
	Bytes   int64   `json:"bytes"`
	Percent float64 `json:"percent"`
}

type languageData struct {
	Username     string          `json:"username"`
	Items        []languageShare `json:"items"`
	Repositories int             `json:"repositories"`
	TotalBytes   int64           `json:"totalBytes"`
	FetchedAt    string          `json:"fetchedAt"`
}

type languageCache struct {
	key          string
	at, failedAt time.Time
	data         *languageData
}

// Use complete byte totals, not an average of per-repository percentages.
// A failed page or language request never replaces the last complete snapshot.
func (g ghFetcher) fetchLanguages(username string) (*languageData, error) {
	out := &languageData{Username: username, Items: []languageShare{}}
	totals := map[string]int64{}
	complete := false
	for page := 1; page <= 100; page++ {
		var repos []struct {
			Name    string `json:"name"`
			Fork    bool   `json:"fork"`
			Private bool   `json:"private"`
			Owner   struct {
				Login string `json:"login"`
			} `json:"owner"`
		}
		endpoint := fmt.Sprintf("/users/%s/repos?type=owner&sort=full_name&per_page=100&page=%d", url.PathEscape(username), page)
		if err := g.api(endpoint, &repos); err != nil {
			return nil, err
		}
		for _, repo := range repos {
			if repo.Fork || repo.Private || !strings.EqualFold(repo.Owner.Login, username) {
				continue
			}
			var languages map[string]int64
			if err := g.api("/repos/"+url.PathEscape(username)+"/"+url.PathEscape(repo.Name)+"/languages", &languages); err != nil {
				return nil, err
			}
			out.Repositories++
			for name, bytes := range languages {
				if bytes > 0 {
					totals[name] += bytes
					out.TotalBytes += bytes
				}
			}
		}
		if len(repos) < 100 {
			complete = true
			break
		}
	}
	if !complete {
		return nil, fmt.Errorf("github_repository_limit")
	}
	for name, bytes := range totals {
		out.Items = append(out.Items, languageShare{Name: name, Bytes: bytes, Percent: float64(bytes) * 100 / float64(out.TotalBytes)})
	}
	sort.Slice(out.Items, func(i, j int) bool {
		if out.Items[i].Bytes == out.Items[j].Bytes {
			return out.Items[i].Name < out.Items[j].Name
		}
		return out.Items[i].Bytes > out.Items[j].Bytes
	})
	out.FetchedAt = nowISO()
	return out, nil
}

func languageKey(cfg config.GitHub) string {
	return fmt.Sprintf("languages:%x", sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(cfg.Username))+"\x00"+cfg.Token+"\x00"+cfg.Proxy+fmt.Sprint(cfg.InsecureTLS))))
}

func (s *Server) readLanguages(cfg config.GitHub, force bool) (*languageData, bool, error) {
	key := languageKey(cfg)
	// Sharing the existing singleflight group coalesces public and admin fetches.
	result, err, _ := s.ghFlight.Do(key, func() (any, error) {
		s.ghMu.Lock()
		cached := s.ghLanguages
		s.ghMu.Unlock()
		if cached.key == key {
			if time.Since(cached.failedAt) < time.Minute {
				return nil, fmt.Errorf("github_retry_later")
			}
			minutes := cfg.RefreshMinutes
			if minutes == 0 {
				minutes = 30
			}
			if !force && cached.data != nil && time.Since(cached.at) < time.Duration(max(1, minutes))*time.Minute {
				return cached.data, nil
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		data, fetchErr := (ghFetcher{client: githubClient(cfg), token: cfg.Token, ctx: ctx}).fetchLanguages(strings.TrimSpace(cfg.Username))
		s.ghMu.Lock()
		if fetchErr == nil {
			s.ghLanguages = languageCache{key: key, at: time.Now(), data: data}
		} else {
			if s.ghLanguages.key != key {
				s.ghLanguages = languageCache{key: key}
			}
			s.ghLanguages.failedAt = time.Now()
		}
		s.ghMu.Unlock()
		if fetchErr != nil {
			s.logSync(false, "语言同步失败："+fetchErr.Error())
		} else {
			s.logSync(true, fmt.Sprintf("语言同步成功：%d 个公开原创仓库 / %d 种语言", data.Repositories, len(data.Items)))
		}
		return data, fetchErr
	})
	if err == nil {
		return result.(*languageData), false, nil
	}
	s.ghMu.Lock()
	cached := s.ghLanguages
	s.ghMu.Unlock()
	if cached.key == key && cached.data != nil {
		return cached.data, true, err
	}
	return nil, false, err
}

func (s *Server) serveLanguages(w http.ResponseWriter, force bool) {
	cfg := s.Config.Typed().GitHub
	if strings.TrimSpace(cfg.Username) == "" {
		writeError(w, http.StatusBadRequest, "github_username_required")
		return
	}
	data, stale, err := s.readLanguages(cfg, force)
	if data == nil {
		message := "github_languages_unavailable"
		if force && err != nil {
			message = err.Error()
		}
		writeError(w, http.StatusBadGateway, message)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		*languageData
		Stale bool `json:"stale"`
	}{data, stale})
}

func (s *Server) githubLanguages(w http.ResponseWriter, _ *http.Request) { s.serveLanguages(w, false) }
func (s *Server) githubLanguagesSync(w http.ResponseWriter, _ *http.Request) {
	s.serveLanguages(w, true)
}
