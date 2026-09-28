package httpapi

import (
	"bytes"
	"encoding/xml"
	"net/http"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"

	"myself/server/internal/config"
	"myself/server/internal/store"
)

// RSS 2.0：最新 20 篇已发布文章，正文 Markdown 渲染为 HTML 放入 content:encoded。

var feedMarkdown = goldmark.New(goldmark.WithExtensions(extension.GFM))

type rssDoc struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Content string     `xml:"xmlns:content,attr"`
	Atom    string     `xml:"xmlns:atom,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language"`
	AtomLink    atomLink  `xml:"atom:link"`
	Items       []rssItem `xml:"item"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type rssItem struct {
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	GUID        string   `xml:"guid"`
	PubDate     string   `xml:"pubDate"`
	Description string   `xml:"description"`
	Categories  []string `xml:"category"`
	Content     cdata    `xml:"content:encoded"`
}

type cdata struct {
	Text string `xml:",cdata"`
}

// siteBase 站点根地址：优先后台配置的 site.url（初始化时填写），未配置时才按请求推断。
func (s *Server) siteBase(r *http.Request) string {
	if u, _ := config.Sub(s.Config.Get(), "site")["url"].(string); strings.HasPrefix(u, "http") {
		return strings.TrimRight(u, "/")
	}
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// GET /feed RSS 订阅源
func (s *Server) rssFeed(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.QueryPosts(`WHERE status = 'published' ORDER BY created_at DESC LIMIT 20`)
	if err != nil {
		fail(w, err)
		return
	}
	cfg := s.Config.Typed()
	base := s.siteBase(r)
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		loc = time.UTC
	}
	items := make([]rssItem, 0, len(rows))
	for _, row := range rows {
		var buf bytes.Buffer
		if err := feedMarkdown.Convert([]byte(row.ContentMd), &buf); err != nil {
			buf.Reset()
			buf.WriteString(row.Excerpt)
		}
		link := base + "/articles/" + row.Slug
		items = append(items, rssItem{
			Title:       row.Title,
			Link:        link,
			GUID:        link,
			PubDate:     rfc1123(row.CreatedAt, loc),
			Description: row.Excerpt,
			Categories:  store.ParseStrings(row.Tags),
			Content:     cdata{Text: buf.String()},
		})
	}
	feed := rssDoc{
		Version: "2.0",
		Content: "http://purl.org/rss/1.0/modules/content/",
		Atom:    "http://www.w3.org/2005/Atom",
		Channel: rssChannel{
			Title:       cfg.Site.Title,
			Link:        base,
			Description: cfg.Site.Subtitle,
			Language:    "zh-CN",
			AtomLink:    atomLink{Href: base + "/feed", Rel: "self", Type: "application/rss+xml"},
			Items:       items,
		},
	}
	out, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		fail(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=600")
	_, _ = w.Write([]byte(xml.Header))
	_, _ = w.Write(out)
}

// rfc1123 把库内 'YYYY-MM-DD HH:MM:SS'（UTC）转成 RSS 要求的 RFC1123Z。
func rfc1123(sqliteTime string, loc *time.Location) string {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", sqliteTime, time.UTC)
	if err != nil {
		return ""
	}
	return t.In(loc).Format(time.RFC1123Z)
}
