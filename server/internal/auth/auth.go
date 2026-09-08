// Package auth 提供管理员凭据（scrypt）与 JWT 签发/校验。
// 哈希格式与旧 Node 实现兼容：`<salt hex>:<scrypt hex>`，N=16384 r=8 p=1 keyLen=64。
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/scrypt"

	"myself/server/internal/store"
)

const tokenTTL = 30 * 24 * time.Hour

// Service 管理员认证服务。
type Service struct {
	db *store.DB
}

// New 构造认证服务。
func New(db *store.DB) *Service {
	return &Service{db: db}
}

// HashPassword 生成 `salt:hash` 格式的 scrypt 哈希。
func HashPassword(password string) (string, error) {
	saltBytes := make([]byte, 16)
	if _, err := rand.Read(saltBytes); err != nil {
		return "", err
	}
	salt := hex.EncodeToString(saltBytes)
	// Node 端把 salt 十六进制字符串本身作为盐字节，这里保持一致。
	key, err := scrypt.Key([]byte(password), []byte(salt), 16384, 8, 1, 64)
	if err != nil {
		return "", err
	}
	return salt + ":" + hex.EncodeToString(key), nil
}

// VerifyPassword 常量时间比较口令与存储哈希。
func VerifyPassword(password, stored string) bool {
	salt, hashHex, ok := strings.Cut(stored, ":")
	if !ok || salt == "" || hashHex == "" {
		return false
	}
	want, err := hex.DecodeString(hashHex)
	if err != nil {
		return false
	}
	got, err := scrypt.Key([]byte(password), []byte(salt), 16384, 8, 1, len(want))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(got, want) == 1
}

// Init 首次启动初始化 JWT 密钥与默认管理员。
func (s *Service) Init() error {
	secret, err := s.db.GetSetting("jwt_secret")
	if err != nil {
		return err
	}
	if secret == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return err
		}
		if err := s.db.SetSetting("jwt_secret", hex.EncodeToString(b)); err != nil {
			return err
		}
	}
	user, err := s.db.GetSetting("admin_username")
	if err != nil {
		return err
	}
	if user == "" {
		const defaultPassword = "myself-admin"
		hash, err := HashPassword(defaultPassword)
		if err != nil {
			return err
		}
		if err := s.db.SetSetting("admin_username", "admin"); err != nil {
			return err
		}
		if err := s.db.SetSetting("admin_password", hash); err != nil {
			return err
		}
		log.Printf("[myself-server] admin account initialized -> admin / %s (请尽快在后台修改密码)", defaultPassword)
	}
	return nil
}

func (s *Service) secret() ([]byte, error) {
	v, err := s.db.GetSetting("jwt_secret")
	if err != nil {
		return nil, err
	}
	if v == "" {
		return nil, errors.New("jwt secret missing")
	}
	return []byte(v), nil
}

// Login 校验用户名口令，成功返回 JWT。失败返回空串。
func (s *Service) Login(username, password string) (string, error) {
	storedUser, err := s.db.GetSetting("admin_username")
	if err != nil {
		return "", err
	}
	if username != storedUser {
		return "", nil
	}
	storedHash, err := s.db.GetSetting("admin_password")
	if err != nil {
		return "", err
	}
	if !VerifyPassword(password, storedHash) {
		return "", nil
	}
	secret, err := s.secret()
	if err != nil {
		return "", err
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  username,
		"role": "admin",
		"iat":  now.Unix(),
		"exp":  now.Add(tokenTTL).Unix(),
	})
	return token.SignedString(secret)
}

// ChangePassword 校验旧口令后写入新哈希。旧口令错误返回 false。
func (s *Service) ChangePassword(oldPassword, newPassword string) (bool, error) {
	storedHash, err := s.db.GetSetting("admin_password")
	if err != nil {
		return false, err
	}
	if !VerifyPassword(oldPassword, storedHash) {
		return false, nil
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return false, err
	}
	return true, s.db.SetSetting("admin_password", hash)
}

// Verify 校验 Bearer Token；合法返回 claims。
func (s *Service) Verify(authorization string) (jwt.MapClaims, error) {
	raw, ok := strings.CutPrefix(authorization, "Bearer ")
	if !ok || raw == "" {
		return nil, errors.New("missing token")
	}
	secret, err := s.secret()
	if err != nil {
		return nil, err
	}
	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected alg %s", t.Method.Alg())
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}
