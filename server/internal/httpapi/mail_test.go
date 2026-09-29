package httpapi

import (
	"encoding/base64"
	"net/mail"
	"strings"
	"testing"

	"myself/server/internal/config"
	"myself/server/internal/store"
)

// 邮件版式：用户输入（昵称）经 HTML 转义，按钮与原始链接都在，纯文本版本含链接。
func TestRenderLetter(t *testing.T) {
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	s := &Server{Deps: Deps{Config: config.New(db)}}
	htmlBody, text, err := s.renderLetter(letter{
		Subject:  "验证",
		Title:    "验证你的邮箱",
		Greeting: `<b>坏</b>人，你好：`,
		Lines:    []string{"欢迎加入。"},
		Action:   &mailAction{Label: "验证邮箱", URL: "https://example.com/account/verify?token=abc"},
		Expire:   "48 小时内有效",
	}, "https://example.com", 2026)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(htmlBody, "<b>坏</b>") || !strings.Contains(htmlBody, "&lt;b&gt;坏&lt;/b&gt;") {
		t.Fatal("greeting was not escaped")
	}
	if strings.Count(htmlBody, "https://example.com/account/verify?token=abc") < 2 {
		t.Fatal("button and fallback link should both carry the URL")
	}
	if !strings.Contains(text, "https://example.com/account/verify?token=abc") || !strings.Contains(text, "48 小时内有效") {
		t.Fatalf("plain text missing link or expiry: %q", text)
	}
}

// 报文：有 HTML 时为 multipart/alternative，纯文本在前、HTML 在后，各部分 base64 可解码。
func TestBuildMessageMultipart(t *testing.T) {
	from := &mail.Address{Name: "Myself", Address: "a@example.com"}
	to := &mail.Address{Address: "b@example.com"}
	raw := string(buildMessage(from, to, "主题", "纯文本", "<p>HTML</p>"))
	if !strings.Contains(raw, "multipart/alternative") {
		t.Fatal("not multipart")
	}
	iText, iHTML := strings.Index(raw, "text/plain"), strings.Index(raw, "text/html")
	if iText < 0 || iHTML < 0 || iText > iHTML {
		t.Fatal("plain part must come before html part")
	}
	enc := base64.StdEncoding.EncodeToString([]byte("<p>HTML</p>"))
	if !strings.Contains(raw, enc) {
		t.Fatal("html part not base64 encoded")
	}
	if plain := string(buildMessage(from, to, "主题", "纯文本", "")); strings.Contains(plain, "multipart") {
		t.Fatal("plain-only message should not be multipart")
	}
}

// 圆角跟随站点全局圆角：默认基准 10 → 卡片 20px、链接框 10px；改成 4 → 8px / 4px。
func TestLetterFollowsSiteRadius(t *testing.T) {
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	s := &Server{Deps: Deps{Config: config.New(db)}}
	l := letter{Title: "t", Action: &mailAction{Label: "go", URL: "https://example.com"}}
	h, _, _ := s.renderLetter(l, "", 2026)
	if !strings.Contains(h, "border-radius:20px") || !strings.Contains(h, "border-radius:10px") {
		t.Fatal("default radius not applied")
	}
	if _, err := s.Config.Save(config.Map{"theme": config.Map{"radius": 4}}); err != nil {
		t.Fatal(err)
	}
	h, _, _ = s.renderLetter(l, "", 2026)
	if !strings.Contains(h, "border-radius:8px") || !strings.Contains(h, "border-radius:4px") {
		t.Fatal("custom radius not applied")
	}
}
