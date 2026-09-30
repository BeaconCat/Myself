package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// keyed 用服务端私密值（jwt_secret）按用途派生的 HMAC，十六进制输出。
// 用于访客 Cookie 签名与评论者 IP 的不可逆标识：没有密钥无法伪造，也无法靠枚举 IPv4 空间反推。
func (s *Server) keyed(purpose, msg string) string {
	secret, _ := s.DB.GetSetting("jwt_secret")
	m := hmac.New(sha256.New, []byte(purpose+"\x00"+secret))
	m.Write([]byte(msg))
	return hex.EncodeToString(m.Sum(nil))
}
