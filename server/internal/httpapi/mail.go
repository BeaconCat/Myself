package httpapi

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"myself/server/internal/config"
)

// errMailOff 未开启或配置不完整。
var errMailOff = errors.New("mail not configured")

// sendLetter 按统一版式发一封系统邮件（HTML + 纯文本备选）。base 为站点地址，用于页脚链接。
func (s *Server) sendLetter(to string, l letter, base string) error {
	htmlBody, text, err := s.renderLetter(l, base, time.Now().Year())
	if err != nil {
		return err
	}
	return sendWith(s.Config.Typed().Mail, to, l.Subject, text, htmlBody)
}

func (s *Server) logMail(err error) {
	log.Printf("[mail] %v", err)
}

// sendWith 发一封 UTF-8 邮件：有 HTML 时为 multipart/alternative（纯文本在前、HTML 在后），否则纯文本。
// security: starttls（默认，587）/ tls（465）/ none（仅内网）。
func sendWith(c config.Mail, to, subject, text, htmlBody string) error {
	if !c.Ready() {
		return errMailOff
	}
	from, err := mail.ParseAddress(c.From)
	if err != nil {
		return fmt.Errorf("invalid from: %w", err)
	}
	rcpt, err := mail.ParseAddress(to)
	if err != nil {
		return fmt.Errorf("invalid to: %w", err)
	}
	addr := net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
	tlsCfg := &tls.Config{ServerName: c.Host, MinVersion: tls.VersionTLS12}

	var conn net.Conn
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	if c.Security == "tls" {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsCfg)
	} else {
		conn, err = dialer.Dial("tcp", addr)
	}
	if err != nil {
		return err
	}
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))
	client, err := smtp.NewClient(conn, c.Host)
	if err != nil {
		conn.Close()
		return err
	}
	defer client.Close()
	if c.Security != "tls" && c.Security != "none" {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return errors.New("server does not support STARTTLS")
		}
		if err := client.StartTLS(tlsCfg); err != nil {
			return err
		}
	}
	if c.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", c.Username, c.Password, c.Host)); err != nil {
			return err
		}
	}
	if err := client.Mail(from.Address); err != nil {
		return err
	}
	if err := client.Rcpt(rcpt.Address); err != nil {
		return err
	}
	wc, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := wc.Write(buildMessage(from, rcpt, subject, text, htmlBody)); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return client.Quit()
}

// buildMessage 组装报文：头部 + 纯文本 / multipart/alternative 正文（各部分 base64）。
func buildMessage(from, rcpt *mail.Address, subject, text, htmlBody string) []byte {
	part := func(ctype, body string) string {
		return "Content-Type: " + ctype + "; charset=UTF-8\r\nContent-Transfer-Encoding: base64\r\n\r\n" +
			wrap76(base64.StdEncoding.EncodeToString([]byte(body)))
	}
	head := []string{
		"From: " + (&mail.Address{Name: from.Name, Address: from.Address}).String(),
		"To: " + rcpt.Address,
		"Subject: " + mime.BEncoding.Encode("UTF-8", subject),
		"Date: " + time.Now().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
	}
	if htmlBody == "" {
		return []byte(strings.Join(head, "\r\n") + "\r\n" + part("text/plain", text))
	}
	boundary := "myself-" + randomFilename("")
	return []byte(strings.Join(head, "\r\n") + "\r\n" +
		"Content-Type: multipart/alternative; boundary=\"" + boundary + "\"\r\n\r\n" +
		"--" + boundary + "\r\n" + part("text/plain", text) + "\r\n" +
		"--" + boundary + "\r\n" + part("text/html", htmlBody) + "\r\n" +
		"--" + boundary + "--\r\n")
}

func wrap76(s string) string {
	var b strings.Builder
	for len(s) > 76 {
		b.WriteString(s[:76] + "\r\n")
		s = s[76:]
	}
	b.WriteString(s)
	return b.String()
}

// POST /admin/mail/test {to} 用当前配置发一封测试信（管理员）
func (s *Server) mailTest(w http.ResponseWriter, r *http.Request) {
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	to := strings.TrimSpace(b.strOr("to"))
	if !emailRe.MatchString(to) {
		writeError(w, http.StatusBadRequest, "invalid_email")
		return
	}
	site := s.Config.Typed().Site.Title
	base := s.siteBase(r)
	if err := s.sendLetter(to, letter{
		Subject:   "「" + site + "」发信测试",
		Preheader: "发信配置可用：注册验证、找回密码与邀请邮件都会以这个样式送达。",
		Title:     "发信配置可用",
		Greeting:  "你好：",
		Lines: []string{
			"这是一封来自「" + site + "」后台的测试邮件。能看到它，说明 SMTP 配置正确。",
			"之后的注册验证、找回密码、邀请与重置链接，都会以这个样式发出。",
		},
		Action: &mailAction{Label: "打开站点", URL: base},
		Note:   "这封邮件由站长在后台「设置 · 邮件」里手动发送。",
	}, base); err != nil {
		// 发信失败的原因只回给管理员，便于排查配置
		writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
