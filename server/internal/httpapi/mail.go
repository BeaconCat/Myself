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

// sendMail 以纯文本发一封信（UTF-8）。security: starttls（默认，587）/ tls（465）/ none（仅内网）。
func (s *Server) sendMail(to, subject, text string) error {
	return sendWith(s.Config.Typed().Mail, to, subject, text)
}

func (s *Server) logMail(err error) {
	log.Printf("[mail] %v", err)
}

func sendWith(c config.Mail, to, subject, text string) error {
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
	msg := strings.Join([]string{
		"From: " + (&mail.Address{Name: from.Name, Address: from.Address}).String(),
		"To: " + rcpt.Address,
		"Subject: " + mime.BEncoding.Encode("UTF-8", subject),
		"Date: " + time.Now().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: base64",
		"",
		wrap76(base64.StdEncoding.EncodeToString([]byte(text))),
	}, "\r\n")
	if _, err := wc.Write([]byte(msg)); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return client.Quit()
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
	if err := s.sendMail(to, "「"+site+"」发信测试", "这是一封测试邮件：发信配置可用。\n\n—— "+site); err != nil {
		// 发信失败的原因只回给管理员，便于排查配置
		writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
