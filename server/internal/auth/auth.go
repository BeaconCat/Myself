// Package auth 提供用户凭据（scrypt）、首次启动初始化、JWT 签发/校验与一次性令牌（验证邮箱、重置密码、邀请）。
// 哈希格式与旧 Node 实现兼容：`<salt hex>:<scrypt hex>`，N=16384 r=8 p=1 keyLen=64。
//
// 安全约定：
//   - 不内置默认管理员；首次启动凭启动日志里的一次性初始化码创建管理员（users 表 role=admin）。
//   - 旧库的单管理员（settings 里的 admin_*）在 Init 时迁移成 users 第一行；仍用历史默认口令的标记必须改密。
//   - JWT：sub = 用户 id，tv = 该用户的令牌版本；改密 / 停用 / 改角色即递增，旧令牌全部失效；有效期 7 天。
//   - 一次性令牌只存 SHA-256 哈希，过期或使用后失效。
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
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

var (
	// ErrAlreadySetup 已完成初始化时再次调用 Setup。
	ErrAlreadySetup = errors.New("already set up")
	// ErrBadSetupCode 初始化码不正确。
	ErrBadSetupCode = errors.New("bad setup code")
	// ErrPending 账号待邮箱验证。
	ErrPending = errors.New("account pending verification")
	// ErrDisabled 账号已停用。
	ErrDisabled = errors.New("account disabled")
	// ErrBadToken 一次性令牌无效或已过期。
	ErrBadToken = errors.New("invalid or expired token")
)

// Service 认证服务。
type Service struct {
	// TTL 登录有效期（后台「登录保持时长」）；nil 用默认 7 天
	TTL       func() time.Duration
	db        *store.DB
	mu        sync.Mutex
	setupCode string
	// dummyHash 用户不存在时也跑一遍 scrypt，消除时序差。
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

// RandomHex n 字节随机数的十六进制串。
func RandomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashToken 一次性令牌 / API Key 的存储哈希。
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// Init 启动时：准备 JWT 密钥；迁移旧版单管理员；未初始化时生成一次性初始化码并打印到日志。
func (s *Service) Init() error {
	secret, err := s.db.GetSetting("jwt_secret")
	if err != nil {
		return err
	}
	if secret == "" {
		v, err := RandomHex(32)
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
	if err := s.migrateLegacyAdmin(); err != nil {
		return err
	}
	need, err := s.NeedsSetup()
	if err != nil {
		return err
	}
	if need {
		code, err := RandomHex(4)
		if err != nil {
			return err
		}
		s.setupCode = strings.ToUpper(code)
		log.Printf("[myself-server] 站点尚未初始化：打开 /setup，初始化码 %s（仅本次启动有效）", s.setupCode)
	}
	return nil
}

// migrateLegacyAdmin 旧版把唯一管理员存在 settings（admin_username / admin_password / token_version / admin_must_change）：
// users 表里还没有管理员时迁移成一行，然后清掉旧键。
func (s *Service) migrateLegacyAdmin() error {
	user, err := s.db.GetSetting("admin_username")
	if err != nil || user == "" {
		return err
	}
	if n, err := s.db.CountUsers(store.RoleAdmin); err != nil || n > 0 {
		return err
	}
	hash, _ := s.db.GetSetting("admin_password")
	tv, _ := s.db.GetSetting("token_version")
	mustChange, _ := s.db.GetSetting("admin_must_change")
	if VerifyPassword(legacyDefaultPassword, hash) {
		mustChange = "1"
		log.Print("[myself-server] 管理员仍在使用历史默认口令：下次登录后必须先修改密码")
	}
	id, err := s.db.CreateUser(store.NewUser{Login: user, Name: user, Role: store.RoleAdmin, Status: store.StatusActive, PasswordHash: hash})
	if err != nil {
		return err
	}
	n, _ := strconv.Atoi(tv)
	if _, err := s.db.Exec(`UPDATE users SET token_version = ?, must_change = ? WHERE id = ?`, n, mustChange == "1", id); err != nil {
		return err
	}
	for _, k := range []string{"admin_username", "admin_password", "token_version", "admin_must_change"} {
		if _, err := s.db.Exec(`DELETE FROM settings WHERE key = ?`, k); err != nil {
			return err
		}
	}
	log.Printf("[myself-server] 管理员 %s 已迁移到用户表", user)
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
	n, err := s.db.CountUsers(store.RoleAdmin)
	return n == 0, err
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

// Setup 首次初始化：校验初始化码后创建管理员，返回其 id。并发调用只有一个成功。
func (s *Service) Setup(code, login, password, name string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	need, err := s.NeedsSetup()
	if err != nil {
		return 0, err
	}
	if !need || s.setupCode == "" {
		return 0, ErrAlreadySetup
	}
	if subtle.ConstantTimeCompare([]byte(strings.ToUpper(strings.TrimSpace(code))), []byte(s.setupCode)) != 1 {
		return 0, ErrBadSetupCode
	}
	hash, err := HashPassword(password)
	if err != nil {
		return 0, err
	}
	if name == "" {
		name = login
	}
	id, err := s.db.CreateUser(store.NewUser{Login: login, Name: name, Role: store.RoleAdmin, Status: store.StatusActive, PasswordHash: hash})
	if err != nil {
		return 0, err
	}
	s.setupCode = ""
	return id, nil
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

// Issue 为用户签发 JWT。
// ttl 当前登录有效期（由站点配置提供，未设置时 7 天）。
func (s *Service) ttl() time.Duration {
	if s.TTL != nil {
		if d := s.TTL(); d > 0 {
			return d
		}
	}
	return tokenTTL
}

func (s *Service) Issue(u *store.User) (string, error) {
	secret, err := s.secret()
	if err != nil {
		return "", err
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  strconv.FormatInt(u.ID, 10),
		"role": u.Role,
		"tv":   u.TokenVersion,
		"iat":  now.Unix(),
		"exp":  now.Add(s.ttl()).Unix(),
	})
	return token.SignedString(secret)
}

// Authenticate 校验登录名（或邮箱）与口令。失败返回 (nil, nil)；账号待验证 / 停用返回对应错误。
// 用户不存在或没有口令（仅 GitHub 登录）时同样计算一次哈希，消除时序差。
func (s *Service) Authenticate(identifier, password string) (*store.User, error) {
	u, err := s.db.UserByLogin(identifier)
	if err != nil {
		return nil, err
	}
	hash := s.dummyHash
	if u != nil && u.PasswordHash != "" {
		hash = u.PasswordHash
	}
	if !VerifyPassword(password, hash) || u == nil || u.PasswordHash == "" {
		return nil, nil
	}
	switch u.Status {
	case store.StatusPending:
		return u, ErrPending
	case store.StatusDisabled:
		return u, ErrDisabled
	}
	return u, nil
}

// VerifyToken 校验令牌：固定 HS256、必须含 exp；用户存在、状态正常、令牌版本一致。
func (s *Service) VerifyToken(raw string) (*store.User, error) {
	if raw == "" {
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
	sub, _ := claims["sub"].(string)
	id, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return nil, errors.New("bad subject")
	}
	u, err := s.db.UserByID(id)
	if err != nil {
		return nil, err
	}
	if u == nil || u.Status != store.StatusActive {
		return nil, errors.New("user unavailable")
	}
	if tv, _ := claims["tv"].(float64); int(tv) != u.TokenVersion {
		return nil, errors.New("token revoked")
	}
	return u, nil
}

// Revoke 递增令牌版本：该用户所有已签发令牌立即失效。
func (s *Service) Revoke(userID int64) error {
	_, err := s.db.Exec(`UPDATE users SET token_version = token_version + 1 WHERE id = ?`, userID)
	return err
}

// SetPassword 直接设置口令（重置链接 / 管理员），吊销旧令牌并解除强制改密。
func (s *Service) SetPassword(userID int64, password string) error {
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE users SET password_hash = ?, must_change = 0, token_version = token_version + 1 WHERE id = ?`, hash, userID)
	return err
}

// ChangePassword 校验旧口令后改密（吊销其它会话），返回当前会话的新令牌；旧口令错误返回空串。
// 仅 GitHub 登录、尚无口令的账号可不填旧口令直接设置。
func (s *Service) ChangePassword(userID int64, oldPassword, newPassword string) (string, error) {
	u, err := s.db.UserByID(userID)
	if err != nil || u == nil {
		return "", err
	}
	if u.PasswordHash != "" && !VerifyPassword(oldPassword, u.PasswordHash) {
		return "", nil
	}
	if err := s.SetPassword(userID, newPassword); err != nil {
		return "", err
	}
	if u, err = s.db.UserByID(userID); err != nil {
		return "", err
	}
	return s.Issue(u)
}

/* ===== 一次性令牌：verify（验证邮箱）/ reset（重置密码）/ invite（邀请） ===== */

// Token 一次性令牌记录。
type Token struct {
	ID     int64
	Kind   string
	UserID int64
	Role   string
	Email  string
	Note   string
}

// CreateToken 生成一次性令牌，返回明文（只出现这一次）。
func (s *Service) CreateToken(kind string, userID int64, role, email, note string, ttl time.Duration) (string, error) {
	raw, err := RandomHex(24)
	if err != nil {
		return "", err
	}
	var uid any
	if userID > 0 {
		uid = userID
	}
	_, err = s.db.Exec(`INSERT INTO user_tokens (kind, token_hash, user_id, role, email, note, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, kind, HashToken(raw), uid, role, email, note,
		time.Now().UTC().Add(ttl).Format("2006-01-02 15:04:05"))
	if err != nil {
		return "", err
	}
	return raw, nil
}

// PeekToken 读取未使用、未过期的令牌（不消耗）。
func (s *Service) PeekToken(kind, raw string) (*Token, error) {
	var t Token
	var uid sql.NullInt64
	err := s.db.QueryRow(`SELECT id, kind, user_id, role, email, note FROM user_tokens
		WHERE kind = ? AND token_hash = ? AND used_at IS NULL AND expires_at > datetime('now')`,
		kind, HashToken(strings.TrimSpace(raw))).Scan(&t.ID, &t.Kind, &uid, &t.Role, &t.Email, &t.Note)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrBadToken
	}
	if err != nil {
		return nil, err
	}
	t.UserID = uid.Int64
	return &t, nil
}

// ConsumeToken 原子地标记令牌已使用（并发只成功一次）。
func (s *Service) ConsumeToken(t *Token, usedBy int64) error {
	var by any
	if usedBy > 0 {
		by = usedBy
	}
	res, err := s.db.Exec(`UPDATE user_tokens SET used_at = datetime('now'), used_by = ? WHERE id = ? AND used_at IS NULL`, by, t.ID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrBadToken
	}
	return nil
}
