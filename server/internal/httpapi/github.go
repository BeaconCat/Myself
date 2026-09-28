package httpapi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"myself/server/internal/config"
)

/*
GitHub 状态：服务端拉取 GitHub API（避免浏览器 CORS、隐藏票证）。
配置 github.mode = 'manual'（手填数字）| 'api'（自动拉取）；
token 可选（只读 PAT，提升配额并可查私有统计）；refreshMinutes 控制缓存。
*/

type githubStats struct {
	Repos     int `json:"repos"`
	Stars     int `json:"stars"`
	Followers int `json:"followers"`
	Commits   int `json:"commits"`
}

type githubActivity struct {
	Type string `json:"type"`
	Repo string `json:"repo"`
	Text string `json:"text"`
	Time string `json:"time"`
}

type heatDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
	Level int    `json:"level"`
}

type githubData struct {
	Username   string           `json:"username"`
	Stats      githubStats      `json:"stats"`
	Activities []githubActivity `json:"activities"`
	Heatmap    []heatDay        `json:"heatmap"`
	FetchedAt  string           `json:"fetchedAt"`
}

type githubCache struct {
	at   time.Time
	key  string
	data *githubData
}

type syncEntry struct {
	At      string `json:"at"`
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func (s *Server) logSync(ok bool, message string) {
	s.ghMu.Lock()
	defer s.ghMu.Unlock()
	s.ghSyncLog = append([]syncEntry{{At: nowISO(), OK: ok, Message: message}}, s.ghSyncLog...)
	if len(s.ghSyncLog) > 20 {
		s.ghSyncLog = s.ghSyncLog[:20]
	}
}

// githubClient 出站策略：可配代理（github.proxy）与跳过 TLS 校验（github.insecureTls）。
// insecureTls 仅作用于本只读拉取通道——用于本机存在 TLS 注入（安全软件/TUN 代理自签证书）的环境。
func githubClient(cfg config.GitHub) *http.Client {
	transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
	if cfg.InsecureTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // 用户显式开启
	}
	proxy := strings.TrimSpace(cfg.Proxy)
	if proxy == "" {
		proxy = os.Getenv("HTTPS_PROXY")
	}
	if proxy == "" {
		proxy = os.Getenv("HTTP_PROXY")
	}
	if proxy != "" {
		if u, err := url.Parse(proxy); err == nil {
			transport.Proxy = http.ProxyURL(u)
		}
	}
	return &http.Client{Transport: transport, Timeout: 30 * time.Second}
}

type ghFetcher struct {
	client *http.Client
	token  string
}

func (g ghFetcher) do(method, rawURL string, body []byte, accept string) ([]byte, error) {
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, rawURL, rd)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "myself-blog")
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if g.token != "" {
		req.Header.Set("Authorization", "Bearer "+g.token)
	}
	res, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("status_%d", res.StatusCode)
	}
	return data, nil
}

// api 调 REST 接口并解析 JSON。
func (g ghFetcher) api(path string, dst any) error {
	data, err := g.do(http.MethodGet, "https://api.github.com"+path, nil, "application/vnd.github+json")
	if err != nil {
		return fmt.Errorf("github_%s", strings.TrimPrefix(err.Error(), "status_"))
	}
	return json.Unmarshal(data, dst)
}

var contribRe = regexp.MustCompile(`data-date="(\d{4}-\d{2}-\d{2})"[^>]*data-level="(\d)"`)

// fetchCalendar 贡献热力图：有票证 → GraphQL（精确次数）；无票证 → 解析公开贡献页 HTML。
func (g ghFetcher) fetchCalendar(username string) ([]heatDay, error) {
	if g.token != "" {
		payload, _ := json.Marshal(map[string]any{
			"query":     `query($login:String!){user(login:$login){contributionsCollection{contributionCalendar{weeks{contributionDays{date contributionCount contributionLevel}}}}}}`,
			"variables": map[string]string{"login": username},
		})
		data, err := g.do(http.MethodPost, "https://api.github.com/graphql", payload, "")
		if err != nil {
			return nil, fmt.Errorf("github_graphql_%s", strings.TrimPrefix(err.Error(), "status_"))
		}
		var resp struct {
			Data struct {
				User struct {
					ContributionsCollection struct {
						ContributionCalendar struct {
							Weeks []struct {
								ContributionDays []struct {
									Date              string `json:"date"`
									ContributionCount int    `json:"contributionCount"`
									ContributionLevel string `json:"contributionLevel"`
								} `json:"contributionDays"`
							} `json:"weeks"`
						} `json:"contributionCalendar"`
					} `json:"contributionsCollection"`
				} `json:"user"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, err
		}
		levels := map[string]int{"NONE": 0, "FIRST_QUARTILE": 1, "SECOND_QUARTILE": 2, "THIRD_QUARTILE": 3, "FOURTH_QUARTILE": 4}
		days := []heatDay{}
		for _, w := range resp.Data.User.ContributionsCollection.ContributionCalendar.Weeks {
			for _, d := range w.ContributionDays {
				days = append(days, heatDay{Date: d.Date, Count: d.ContributionCount, Level: levels[d.ContributionLevel]})
			}
		}
		return days, nil
	}

	html, err := ghFetcher{client: g.client}.do(http.MethodGet,
		"https://github.com/users/"+url.PathEscape(username)+"/contributions", nil, "")
	if err != nil {
		return nil, fmt.Errorf("github_contrib_%s", strings.TrimPrefix(err.Error(), "status_"))
	}
	days := []heatDay{}
	for _, m := range contribRe.FindAllStringSubmatch(string(html), -1) {
		days = append(days, heatDay{Date: m[1], Level: int(m[2][0] - '0')})
	}
	sort.SliceStable(days, func(i, j int) bool { return days[i].Date < days[j].Date })
	return days, nil
}

func (g ghFetcher) fetchStatus(username string) (*githubData, error) {
	user := struct {
		PublicRepos int `json:"public_repos"`
		Followers   int `json:"followers"`
	}{}
	if err := g.api("/users/"+url.PathEscape(username), &user); err != nil {
		return nil, err
	}
	var repos []struct {
		Stars int `json:"stargazers_count"`
	}
	if err := g.api("/users/"+url.PathEscape(username)+"/repos?per_page=100&sort=pushed", &repos); err != nil {
		return nil, err
	}
	stars := 0
	for _, r := range repos {
		stars += r.Stars
	}

	// 年度提交总数：Commit Search API（匿名可用，配额低；失败回退事件统计）
	year := time.Now().Year()
	commitsThisYear := 0
	var search struct {
		TotalCount int `json:"total_count"`
	}
	q := url.QueryEscape(fmt.Sprintf("author:%s author-date:>=%d-01-01", username, year))
	if err := g.api("/search/commits?q="+q+"&per_page=1", &search); err == nil {
		commitsThisYear = search.TotalCount
	}

	// 最近公开动态（PushEvent 提交）
	activities := []githubActivity{}
	var events []struct {
		Type      string `json:"type"`
		CreatedAt string `json:"created_at"`
		Repo      struct {
			Name string `json:"name"`
		} `json:"repo"`
		Payload struct {
			Commits []struct {
				Message string `json:"message"`
			} `json:"commits"`
		} `json:"payload"`
	}
	if err := g.api("/users/"+url.PathEscape(username)+"/events/public?per_page=30", &events); err == nil {
		eventCommits := 0
		for _, ev := range events {
			if ev.Type != "PushEvent" {
				continue
			}
			if t, err := time.Parse(time.RFC3339, ev.CreatedAt); err == nil && t.Year() == year {
				eventCommits += len(ev.Payload.Commits)
			}
			for _, c := range ev.Payload.Commits {
				if len(activities) >= 5 {
					break
				}
				repo := ev.Repo.Name
				if _, after, ok := strings.Cut(repo, "/"); ok {
					repo = after
				}
				first, _, _ := strings.Cut(c.Message, "\n")
				if r := []rune(first); len(r) > 80 {
					first = string(r[:80])
				}
				activities = append(activities, githubActivity{Type: "commit", Repo: repo, Text: first, Time: ev.CreatedAt})
			}
		}
		if commitsThisYear == 0 {
			commitsThisYear = eventCommits
		}
	}

	// 贡献热力图（失败不致命）
	heatmap, err := g.fetchCalendar(username)
	if err != nil {
		heatmap = []heatDay{}
	}

	repoCount := user.PublicRepos
	if repoCount == 0 {
		repoCount = len(repos)
	}
	return &githubData{
		Username:   username,
		Stats:      githubStats{Repos: repoCount, Stars: stars, Followers: user.Followers, Commits: commitsThisYear},
		Activities: activities,
		Heatmap:    heatmap,
		FetchedAt:  nowISO(),
	}, nil
}

func cacheKey(cfg config.GitHub) string {
	suffix := "anon"
	if cfg.Token != "" {
		suffix = "tok"
	}
	return cfg.Username + ":" + suffix
}

func manualPayload(cfg config.GitHub, fallback string) map[string]any {
	out := map[string]any{
		"mode":       "manual",
		"username":   cfg.Username,
		"stats":      cfg.Stats,
		"activities": []any{},
	}
	if fallback != "" {
		out["fallback"] = fallback
	}
	return out
}

// withData 把 githubData 平铺进响应对象。
func withData(base map[string]any, d *githubData) map[string]any {
	raw, _ := json.Marshal(d)
	var flat map[string]any
	_ = json.Unmarshal(raw, &flat)
	for k, v := range flat {
		base[k] = v
	}
	return base
}

// GET /github-status 公开：api 模式拉取（带缓存），manual 模式回配置数字
func (s *Server) githubStatus(w http.ResponseWriter, _ *http.Request) {
	cfg := s.Config.Typed().GitHub
	if cfg.Mode != "api" {
		writeJSON(w, http.StatusOK, manualPayload(cfg, ""))
		return
	}
	minutes := cfg.RefreshMinutes
	if minutes == 0 {
		minutes = 30
	}
	ttl := time.Duration(max(1, minutes)) * time.Minute
	key := cacheKey(cfg)

	s.ghMu.Lock()
	cached := s.ghCache
	s.ghMu.Unlock()
	if cached.data != nil && cached.key == key && time.Since(cached.at) < ttl {
		writeJSON(w, http.StatusOK, withData(map[string]any{"mode": "api", "cached": true}, cached.data))
		return
	}

	// 最近一分钟内刚失败过：不再向外拉取，直接回退（防止匿名刷新耗尽配额）
	s.ghMu.Lock()
	recentFail := s.ghFailKey == key && time.Since(s.ghFailAt) < time.Minute
	s.ghMu.Unlock()
	if recentFail {
		if cached.data != nil && cached.key == key {
			writeJSON(w, http.StatusOK, withData(map[string]any{"mode": "api", "cached": true, "stale": true}, cached.data))
			return
		}
		writeJSON(w, http.StatusOK, manualPayload(cfg, "fetch_failed"))
		return
	}

	// 缓存过期瞬间的并发请求合并为一次拉取
	result, err, _ := s.ghFlight.Do(key, func() (any, error) {
		return s.fetchAndCache(cfg, key, "拉取成功")
	})
	data, _ := result.(*githubData)
	if err != nil {
		s.ghMu.Lock()
		s.ghFailAt, s.ghFailKey = time.Now(), key
		s.ghMu.Unlock()
		// 拉取失败回退：旧缓存 → 手填数字（错误详情只进同步日志，不对外暴露）
		if cached.data != nil && cached.key == key {
			writeJSON(w, http.StatusOK, withData(map[string]any{"mode": "api", "cached": true, "stale": true}, cached.data))
			return
		}
		writeJSON(w, http.StatusOK, manualPayload(cfg, "fetch_failed"))
		return
	}
	writeJSON(w, http.StatusOK, withData(map[string]any{"mode": "api", "cached": false}, data))
}

// fetchAndCache 拉取并写缓存 + 同步日志；label 用于日志前缀。
func (s *Server) fetchAndCache(cfg config.GitHub, key, label string) (*githubData, error) {
	fetcher := ghFetcher{client: githubClient(cfg), token: cfg.Token}
	data, err := fetcher.fetchStatus(cfg.Username)
	if err != nil {
		s.logSync(false, label+"失败："+err.Error())
		return nil, err
	}
	s.ghMu.Lock()
	s.ghCache = githubCache{at: time.Now(), key: key, data: data}
	s.ghMu.Unlock()
	s.logSync(true, fmt.Sprintf("%s：%d 仓库 / %d stars", label, data.Stats.Repos, data.Stats.Stars))
	return data, nil
}

// POST /admin/github/sync 立即同步（强刷缓存），返回最新数据
func (s *Server) githubSync(w http.ResponseWriter, _ *http.Request) {
	cfg := s.Config.Typed().GitHub
	data, err := s.fetchAndCache(cfg, cacheKey(cfg), "手动同步")
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "data": data})
}

// GET /admin/github/log 同步日志 + 当前缓存预览
func (s *Server) githubLog(w http.ResponseWriter, _ *http.Request) {
	s.ghMu.Lock()
	defer s.ghMu.Unlock()
	var cachedAt *string
	if !s.ghCache.at.IsZero() {
		v := isoTime(s.ghCache.at)
		cachedAt = &v
	}
	log := s.ghSyncLog
	if log == nil {
		log = []syncEntry{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"log":      log,
		"preview":  s.ghCache.data,
		"cachedAt": cachedAt,
	})
}
