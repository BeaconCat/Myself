package httpapi

import (
	"bytes"
	"encoding/xml"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"

	"myself/server/internal/config"
	"myself/server/internal/store"
)

// RSS 2.0：最新 20 篇已发布文章，正文 Markdown 渲染为 HTML 放入 content:encoded。

var feedMarkdown = goldmark.New(goldmark.WithExtensions(extension.GFM))

// rootRelAttr 匹配站内根相对地址的 src / href（"/x"，不含协议相对的 "//x"）。
var rootRelAttr = regexp.MustCompile(`(\s(?:src|href)=")/([^/"][^"]*)?"`)

// absolutize 把正文里的站内根相对链接与图片改成带站点地址的绝对地址：阅读器不在本站域名下，相对地址会失效。
func absolutize(html, base string) string {
	if base == "" {
		return html
	}
	return rootRelAttr.ReplaceAllString(html, `${1}`+base+`/${2}"`)
}

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
	if secureRequest(r) {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// linkBase 邮件等「离站」链接的站点根地址：只用后台配置的 site.url，绝不信任请求里的 Host
// （否则攻击者伪造 Host 发起找回密码，站长收到的真实邮件就会把重置令牌送到攻击者域名）。
// trusted=true 表示请求来自已登录的站长（例如生成重置 / 邀请链接），这时才允许按请求推断兜底。
// 返回空串表示未配置站点地址，调用方应拒绝发出链接。
func (s *Server) linkBase(r *http.Request, trusted bool) string {
	if u, _ := config.Sub(s.Config.Get(), "site")["url"].(string); strings.HasPrefix(u, "http") {
		return strings.TrimRight(u, "/")
	}
	if trusted {
		return s.siteBase(r)
	}
	return ""
}

// rememberSiteURL 站点地址还没配置时，用站长浏览器的 Origin 记下来（初始化、站长登录时调用）。
// Origin 由浏览器设置、页面脚本改不了；调用前请求已通过口令 / 初始化码校验，且 Origin 与 Host 一致。
func (s *Server) rememberSiteURL(r *http.Request) {
	if u, _ := config.Sub(s.Config.Get(), "site")["url"].(string); strings.HasPrefix(u, "http") {
		return
	}
	origin := r.Header.Get("Origin")
	if !strings.HasPrefix(origin, "http") || !sameOrigin(origin, r.Host) {
		return
	}
	_, _ = s.Config.Save(config.Map{"site": config.Map{"url": strings.TrimRight(origin, "/")}})
}

// GET /feed RSS 订阅源
func (s *Server) rssFeed(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.QueryPosts(`WHERE ` + store.PublicPost + ` ORDER BY created_at DESC LIMIT 20`)
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
			Content:     cdata{Text: absolutize(buf.String(), base)},
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
