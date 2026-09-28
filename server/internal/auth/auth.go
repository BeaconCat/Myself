// Package auth 提供管理员凭据（scrypt）、首次启动初始化与 JWT 签发/校验。
// 哈希格式与旧 Node 实现兼容：`<salt hex>:<scrypt hex>`，N=16384 r=8 p=1 keyLen=64。
//
// 安全约定：
//   - 不再内置默认管理员；首次启动进入初始化流程，凭启动日志里的一次性初始化码创建管理员。
//   - 旧库若仍使用历史默认口令，标记 admin_must_change，登录后除改密外的管理接口一律拒绝。
//   - JWT 带 tv（token_version）：改密即递增，旧令牌全部失效；有效期 7 天，必须含 exp。
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/scrypt"

	"myself/server/internal/store"
)

const (
	tokenTTL = 7 * 24 * time.Hour
	// legacyDefaultPassword 历史版本自动创建的默认口令，仅用于识别需强制改密的旧库。
	legacyDefaultPassword = "myself-admin"
	// MinPasswordLen 口令最短长度。
	MinPasswordLen = 8
)

// ErrAlreadySetup 已完成初始化时再次调用 Setup。
var ErrAlreadySetup = errors.New("already set up")

// ErrBadSetupCode 初始化码不正确。
var ErrBadSetupCode = errors.New("bad setup code")

// Service 管理员认证服务。
type Service struct {
	db        *store.DB
	mu        sync.Mutex
	setupCode string
	// dummyHash 用户名不存在时也跑一遍 scrypt，消除时序差。
	dummyHash string
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

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Init 启动时准备 JWT 密钥；未初始化时生成一次性初始化码并打印到日志；
// 已初始化但仍是历史默认口令的旧库标记为必须改密。
func (s *Service) Init() error {
	secret, err := s.db.GetSetting("jwt_secret")
	if err != nil {
		return err
	}
	if secret == "" {
		v, err := randomHex(32)
		if err != nil {
			return err
		}
		if err := s.db.SetSetting("jwt_secret", v); err != nil {
			return err
		}
	}
	if s.dummyHash, err = HashPassword("myself-timing-equalizer"); err != nil {
		return err
	}
	need, err := s.NeedsSetup()
	if err != nil {
		return err
	}
	if need {
		code, err := randomHex(4)
		if err != nil {
			return err
		}
		s.setupCode = strings.ToUpper(code)
		log.Printf("[myself-server] 站点尚未初始化：打开 /setup，初始化码 %s（仅本次启动有效）", s.setupCode)
		return nil
	}
	hash, err := s.db.GetSetting("admin_password")
	if err != nil {
		return err
	}
	if VerifyPassword(legacyDefaultPassword, hash) {
		log.Print("[myself-server] 管理员仍在使用历史默认口令：下次登录后必须先修改密码")
		return s.db.SetSetting("admin_must_change", "1")
	}
	return nil
}

// SetupCode 本次启动的初始化码（已初始化时为空）。供测试与命令行工具读取；HTTP 层不外泄。
func (s *Service) SetupCode() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.setupCode
}

// NeedsSetup 尚未创建管理员。
func (s *Service) NeedsSetup() (bool, error) {
	user, err := s.db.GetSetting("admin_username")
	return user == "", err
}

// CheckSetupCode 仅校验初始化码（初始化向导第一步即时反馈）。已初始化时返回 ErrAlreadySetup。
func (s *Service) CheckSetupCode(code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.setupCode == "" {
		return ErrAlreadySetup
	}
	if subtle.ConstantTimeCompare([]byte(strings.ToUpper(strings.TrimSpace(code))), []byte(s.setupCode)) != 1 {
		return ErrBadSetupCode
	}
	return nil
}

// Setup 首次初始化：校验初始化码后创建管理员。并发调用只有一个成功。
func (s *Service) Setup(code, username, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	need, err := s.NeedsSetup()
	if err != nil {
		return err
	}
	if !need || s.setupCode == "" {
		return ErrAlreadySetup
	}
	if subtle.ConstantTimeCompare([]byte(strings.ToUpper(strings.TrimSpace(code))), []byte(s.setupCode)) != 1 {
		return ErrBadSetupCode
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	if err := s.db.SetSetting("admin_password", hash); err != nil {
		return err
	}
	if err := s.db.SetSetting("admin_username", username); err != nil {
		return err
	}
	s.setupCode = ""
	return nil
}

// MustChange 旧库仍在用历史默认口令，需先改密。
func (s *Service) MustChange() bool {
	v, _ := s.db.GetSetting("admin_must_change")
	return v == "1"
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

func (s *Service) tokenVersion() int {
	v, _ := s.db.GetSetting("token_version")
	n, _ := strconv.Atoi(v)
	return n
}

// Issue 为当前管理员签发 JWT。
func (s *Service) Issue() (string, error) {
	user, err := s.db.GetSetting("admin_username")
	if err != nil {
		return "", err
	}
	secret, err := s.secret()
	if err != nil {
		return "", err
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user,
		"role": "admin",
		"tv":   s.tokenVersion(),
		"iat":  now.Unix(),
		"exp":  now.Add(tokenTTL).Unix(),
	})
	return token.SignedString(secret)
}

// Login 校验用户名口令，成功返回 JWT。失败返回空串。用户名不存在时同样计算一次哈希。
func (s *Service) Login(username, password string) (string, error) {
	storedUser, err := s.db.GetSetting("admin_username")
	if err != nil {
		return "", err
	}
	storedHash, err := s.db.GetSetting("admin_password")
	if err != nil {
		return "", err
	}
	userOK := storedUser != "" && subtle.ConstantTimeCompare([]byte(username), []byte(storedUser)) == 1
	if !userOK {
		storedHash = s.dummyHash
	}
	if !VerifyPassword(password, storedHash) || !userOK {
		return "", nil
	}
	return s.Issue()
}

// ChangePassword 校验旧口令后写入新哈希，递增令牌版本（其它会话全部失效）并解除强制改密，
// 返回当前会话的新令牌。旧口令错误返回空串。
func (s *Service) ChangePassword(oldPassword, newPassword string) (string, error) {
	storedHash, err := s.db.GetSetting("admin_password")
	if err != nil {
		return "", err
	}
	if !VerifyPassword(oldPassword, storedHash) {
		return "", nil
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return "", err
	}
	if err := s.db.SetSetting("admin_password", hash); err != nil {
		return "", err
	}
	if err := s.db.SetSetting("token_version", strconv.Itoa(s.tokenVersion()+1)); err != nil {
		return "", err
	}
	if err := s.db.SetSetting("admin_must_change", ""); err != nil {
		return "", err
	}
	return s.Issue()
}

// Verify 校验 Bearer Token：固定 HS256、必须含 exp，sub 与令牌版本需与当前管理员一致。
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
	_, err = jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) { return secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return nil, err
	}
	user, err := s.db.GetSetting("admin_username")
	if err != nil {
		return nil, err
	}
	if sub, _ := claims["sub"].(string); user == "" || sub != user {
		return nil, errors.New("subject mismatch")
	}
	if tv, _ := claims["tv"].(float64); int(tv) != s.tokenVersion() {
		return nil, errors.New("token revoked")
	}
	return claims, nil
}
