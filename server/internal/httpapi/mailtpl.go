package httpapi

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"

	"myself/server/internal/config"
)

// letter 一封系统邮件的内容；渲染成 HTML（带纯文本备选），版式统一。
type letter struct {
	Subject string
	// Preheader 收件箱列表里标题后的预览文字（正文里隐藏）
	Preheader string
	Title     string
	Greeting  string
	Lines     []string
	Action    *mailAction
	// Expire 按钮下方的小字，如「链接 48 小时内有效」
	Expire string
	// Note 分隔线下的补充说明，如「不是你本人操作可忽略」
	Note string
}

type mailAction struct {
	Label string
	URL   string
}

// mailView 模板数据：信件内容 + 站点品牌。
type mailView struct {
	letter
	Site   string
	Base   string
	Accent template.CSS
	Soft   template.CSS
	Year   int
}

var hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// brandAccent 取站点默认色盘的主色（邮件里的按钮与装饰线），不合法时用品牌红。
func (s *Server) brandAccent() string {
	theme := config.Sub(s.Config.Get(), "theme")
	id, _ := theme["defaultPaletteId"].(string)
	presets, _ := theme["presets"].([]any)
	for _, p := range presets {
		m, _ := p.(map[string]any)
		if m["id"] == id {
			if c, _ := m["primary"].(string); hexColorRe.MatchString(c) {
				return c
			}
		}
	}
	return "#ff0032"
}

// softColor 主色按 5% 叠在白底上的实色（邮件客户端对带透明度的颜色支持不一）。
func softColor(hex string) string {
	v, err := strconv.ParseUint(hex[1:], 16, 32)
	if err != nil {
		return "#f6f4f0"
	}
	mix := func(c uint64) uint64 { return (c*5 + 255*95) / 100 }
	return fmt.Sprintf("#%02x%02x%02x", mix(v>>16&0xff), mix(v>>8&0xff), mix(v&0xff))
}

// renderLetter 生成 HTML 与纯文本两个版本。
func (s *Server) renderLetter(l letter, base string, year int) (htmlBody, text string, err error) {
	accent := s.brandAccent()
	v := mailView{
		letter: l,
		Site:   s.Config.Typed().Site.Title,
		Base:   base,
		Accent: template.CSS(accent),
		Soft:   template.CSS(softColor(accent)),
		Year:   year,
	}
	var buf bytes.Buffer
	if err := mailTpl.Execute(&buf, v); err != nil {
		return "", "", err
	}
	return buf.String(), plainLetter(v), nil
}

// plainLetter 纯文本版本：不支持 HTML 的客户端与垃圾邮件过滤器都读它。
func plainLetter(v mailView) string {
	var b strings.Builder
	if v.Greeting != "" {
		b.WriteString(v.Greeting + "\n\n")
	}
	for _, line := range v.Lines {
		b.WriteString(line + "\n\n")
	}
	if v.Action != nil {
		b.WriteString(v.Action.Label + "：\n" + v.Action.URL + "\n\n")
	}
	if v.Expire != "" {
		b.WriteString(v.Expire + "\n\n")
	}
	if v.Note != "" {
		b.WriteString(v.Note + "\n\n")
	}
	b.WriteString("—— " + v.Site)
	if v.Base != "" {
		b.WriteString("  " + v.Base)
	}
	return b.String()
}

// 邮件版式：表格布局 + 行内样式（各家客户端都不可靠地支持 <style> 与 flex）。
// 暖白底、白纸卡片、顶部一道主色线；宋体标题、黑体正文；实底胶囊按钮，下附可复制的原始链接。
var mailTpl = template.Must(template.New("mail").Parse(`<!doctype html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="color-scheme" content="light">
<title>{{.Subject}}</title>
</head>
<body style="margin:0;padding:0;background:#f4f2ee;-webkit-text-size-adjust:100%;">
<div style="display:none;max-height:0;overflow:hidden;opacity:0;color:transparent;">{{.Preheader}}</div>
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="background:#f4f2ee;">
<tr><td align="center" style="padding:40px 16px 32px;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="max-width:520px;">
    <tr><td style="padding:0 6px 18px;font-family:'Noto Serif SC','Source Han Serif SC','Songti SC',STSong,serif;font-size:17px;font-weight:700;letter-spacing:.02em;color:#1d1c1a;">
      <span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:{{.Accent}};margin-right:9px;vertical-align:middle;"></span>{{.Site}}
    </td></tr>
    <tr><td style="background:#ffffff;border-radius:16px;border-top:4px solid {{.Accent}};box-shadow:0 18px 40px -28px rgba(29,28,26,.35);">
      <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
        <tr><td style="padding:36px 36px 8px;font-family:'Noto Serif SC','Source Han Serif SC','Songti SC',STSong,serif;font-size:24px;line-height:1.4;font-weight:700;color:#1d1c1a;">{{.Title}}</td></tr>
        <tr><td style="padding:10px 36px 0;font-family:-apple-system,BlinkMacSystemFont,'PingFang SC','Microsoft YaHei','Noto Sans SC',sans-serif;font-size:15px;line-height:1.85;color:#45423c;">
          {{if .Greeting}}<p style="margin:0 0 12px;">{{.Greeting}}</p>{{end}}
          {{range .Lines}}<p style="margin:0 0 12px;">{{.}}</p>{{end}}
        </td></tr>
        {{if .Action}}
        <tr><td style="padding:14px 36px 4px;">
          <table role="presentation" cellpadding="0" cellspacing="0" border="0"><tr>
            <td align="center" bgcolor="{{.Accent}}" style="border-radius:999px;background:{{.Accent}};">
              <a href="{{.Action.URL}}" target="_blank" style="display:inline-block;padding:13px 30px;font-family:-apple-system,BlinkMacSystemFont,'PingFang SC','Microsoft YaHei',sans-serif;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;border-radius:999px;">{{.Action.Label}}</a>
            </td>
          </tr></table>
        </td></tr>
        <tr><td style="padding:14px 36px 0;font-family:-apple-system,BlinkMacSystemFont,'PingFang SC','Microsoft YaHei',sans-serif;font-size:12.5px;line-height:1.7;color:#8a857c;">
          {{if .Expire}}{{.Expire}}<br>{{end}}按钮无法点击时，复制下面的链接到浏览器打开：
          <div style="margin-top:6px;padding:10px 12px;border-radius:10px;background:{{.Soft}};font-family:'SFMono-Regular',Consolas,Menlo,monospace;font-size:12px;line-height:1.6;color:#45423c;word-break:break-all;">{{.Action.URL}}</div>
        </td></tr>
        {{end}}
        {{if .Note}}
        <tr><td style="padding:24px 36px 0;"><div style="height:1px;background:#ece9e3;line-height:1px;font-size:0;">&nbsp;</div></td></tr>
        <tr><td style="padding:14px 36px 0;font-family:-apple-system,BlinkMacSystemFont,'PingFang SC','Microsoft YaHei',sans-serif;font-size:13px;line-height:1.75;color:#8a857c;">{{.Note}}</td></tr>
        {{end}}
        <tr><td style="height:32px;line-height:32px;font-size:0;">&nbsp;</td></tr>
      </table>
    </td></tr>
    <tr><td align="center" style="padding:20px 12px 0;font-family:-apple-system,BlinkMacSystemFont,'PingFang SC','Microsoft YaHei',sans-serif;font-size:12px;line-height:1.8;color:#a39e94;">
      这是一封系统邮件，请勿直接回复。<br>
      &copy; {{.Year}} {{.Site}}{{if .Base}} · <a href="{{.Base}}" target="_blank" style="color:#a39e94;text-decoration:underline;">{{.Base}}</a>{{end}}
    </td></tr>
  </table>
</td></tr>
</table>
</body>
</html>`))
