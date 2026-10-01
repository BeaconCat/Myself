package httpapi

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"
	"time"

	"myself/server/internal/store"
)

// 管理员：用户列表 / 统计 / 改角色 / 停用 / 重置链接 / 删除 / 导出，与邀请管理。

type adminUser struct {
	ID            int64  `json:"id"`
	Login         string `json:"login"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Avatar        string `json:"avatar"`
	AvatarPending string `json:"avatarPending,omitempty"`
	Role          string `json:"role"`
	Status        string `json:"status"`
	GitHub        bool   `json:"github"`
	HasPassword   bool   `json:"hasPassword"`
	EmailVerified bool   `json:"emailVerified"`
	CreatedAt     string `json:"createdAt"`
	LastActiveAt  string `json:"lastActiveAt"`
	Comments      int    `json:"comments"`
	Self          bool   `json:"self"`
}

// userFilters 用户表的角色分段（disabled 按状态筛）。
var userFilters = map[string]string{
	"admin":    "role = 'admin'",
	"author":   "role = 'author'",
	"reader":   "role = 'reader'",
	"disabled": "status = 'disabled'",
}

// GET /admin/users?page=&pageSize=&role=all|admin|author|reader|disabled&q=
// → {items, stats, page, pageSize, total, counts, pendingAvatars}。
// 不带 page 时返回全部（兼容旧调用）；counts 为各分段在当前搜索下的人数，pendingAvatars 为全部待审头像（不受分页影响）。
func (s *Server) adminListUsers(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	var search []string
	var args []any
	if q != "" {
		search = append(search, `(name LIKE ? ESCAPE '\' OR login LIKE ? ESCAPE '\' OR COALESCE(email, '') LIKE ? ESCAPE '\')`)
		like := likeArg(q)
		args = append(args, like, like, like)
	}
	whereOf := func(extra string) string {
		conds := search
		if extra != "" {
			conds = append(append([]string{}, search...), extra)
		}
		if len(conds) == 0 {
			return ""
		}
		return "WHERE " + strings.Join(conds, " AND ")
	}
	counts := map[string]int{}
	for _, f := range []string{"all", "admin", "author", "reader", "disabled"} {
		var n int
		if err := s.DB.QueryRow(`SELECT COUNT(*) FROM users `+whereOf(userFilters[f]), args...).Scan(&n); err != nil {
			fail(w, err)
			return
		}
		counts[f] = n
	}
	filter := r.URL.Query().Get("role")
	if _, ok := userFilters[filter]; !ok {
		filter = "all"
	}
	total := counts[filter]
	page, pageSize, limit := 1, total, -1
	if r.URL.Query().Has("page") {
		page = queryInt(r, "page", 1, 1, 1<<30)
		pageSize = queryInt(r, "pageSize", 20, 1, 100)
		limit = pageSize
	}
	users, commentCounts, err := s.DB.QueryUsers(whereOf(userFilters[filter]), args, limit, (page-1)*max(pageSize, 0))
	if err != nil {
		fail(w, err)
		return
	}
	me := userOf(r)
	toAdmin := func(users []store.User, counts map[int64]int) []adminUser {
		out := make([]adminUser, 0, len(users))
		for _, u := range users {
			out = append(out, adminUser{
				ID: u.ID, Login: u.Login, Email: u.Email, Name: u.Name, Avatar: s.avatarOf(u.Role, u.Avatar), AvatarPending: u.AvatarPending,
				Role: u.Role, Status: u.Status,
				GitHub: u.GitHubID != 0, HasPassword: u.PasswordHash != "", EmailVerified: u.EmailVerified,
				CreatedAt: u.CreatedAt, LastActiveAt: u.LastActiveAt, Comments: counts[u.ID], Self: u.ID == me.ID,
			})
		}
		return out
	}
	pending, pendingCounts, err := s.DB.QueryUsers(`WHERE avatar_pending != ''`, nil, 100, 0)
	if err != nil {
		fail(w, err)
		return
	}
	stat := func(q string, args ...any) int {
		var n int
		_ = s.DB.QueryRow(q, args...).Scan(&n)
		return n
	}
	monthAgo := time.Now().UTC().Add(-30 * 24 * time.Hour).Format("2006-01-02 15:04:05")
	writeJSON(w, http.StatusOK, map[string]any{
		"items":          toAdmin(users, commentCounts),
		"page":           page,
		"pageSize":       pageSize,
		"total":          total,
		"counts":         counts,
		"pendingAvatars": toAdmin(pending, pendingCounts),
		"stats": map[string]int{
			"total":         stat(`SELECT COUNT(*) FROM users`),
			"newMonth":      stat(`SELECT COUNT(*) FROM users WHERE created_at >= ?`, monthAgo),
			"activeMonth":   stat(`SELECT COUNT(*) FROM users WHERE last_active_at >= ?`, monthAgo),
			"authors":       stat(`SELECT COUNT(*) FROM users WHERE role = 'author'`),
			"avatars":       stat(`SELECT COUNT(*) FROM users WHERE avatar_pending != ''`),
			"comments":      stat(`SELECT COUNT(*) FROM comments WHERE status = 'approved'`),
			"commentsMonth": stat(`SELECT COUNT(*) FROM comments WHERE status = 'approved' AND created_at >= ?`, monthAgo),
			"pending":       stat(`SELECT COUNT(*) FROM comments WHERE status = 'pending'`),
		},
	})
}

// targetUser 取路径里的用户；不能对自己做角色 / 状态 / 删除操作。
func (s *Server) targetUser(w http.ResponseWriter, r *http.Request) *store.User {
	u, err := s.DB.UserByID(pathID(r))
	if err != nil {
		fail(w, err)
		return nil
	}
	if u == nil {
		writeError(w, http.StatusNotFound, "not_found")
		return nil
	}
	if u.ID == userOf(r).ID {
		writeError(w, http.StatusBadRequest, "cannot_modify_self")
		return nil
	}
	return u
}

// PUT /admin/users/{id} {role?, status?, name?}：改角色 / 停用后该用户令牌立即失效
func (s *Server) adminUpdateUser(w http.ResponseWriter, r *http.Request) {
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	u := s.targetUser(w, r)
	if u == nil {
		return
	}
	role, status, name := u.Role, u.Status, u.Name
	if v := b.strOr("role"); v != "" {
		if v != store.RoleAdmin && v != store.RoleAuthor && v != store.RoleReader {
			writeError(w, http.StatusBadRequest, "invalid_role")
			return
		}
		role = v
	}
	if v := b.strOr("status"); v != "" {
		if v != store.StatusActive && v != store.StatusDisabled {
			writeError(w, http.StatusBadRequest, "invalid_status")
			return
		}
		status = v
	}
	if v := strings.TrimSpace(b.strOr("name")); v != "" {
		name = limitRunes(v, 40)
	}
	if _, err := s.DB.Exec(`UPDATE users SET role = ?, status = ?, name = ? WHERE id = ?`, role, status, name, u.ID); err != nil {
		fail(w, err)
		return
	}
	if role != u.Role || status != u.Status {
		if err := s.Auth.Revoke(u.ID); err != nil {
			fail(w, err)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// DELETE /admin/users/{id}：评论保留但匿名化为「已注销」，其文章归到站长名下
func (s *Server) adminDeleteUser(w http.ResponseWriter, r *http.Request) {
	u := s.targetUser(w, r)
	if u == nil {
		return
	}
	tx, err := s.DB.Begin()
	if err != nil {
		fail(w, err)
		return
	}
	defer tx.Rollback()
	for _, q := range []string{
		`UPDATE comments SET user_id = NULL, guest_name = '已注销用户' WHERE user_id = ?`,
		`UPDATE posts SET author_id = NULL WHERE author_id = ?`,
		`DELETE FROM user_tokens WHERE user_id = ?`,
		`DELETE FROM users WHERE id = ?`,
	} {
		if _, err := tx.Exec(q, u.ID); err != nil {
			fail(w, err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		fail(w, err)
		return
	}
	// 头像文件随账号一起删除
	s.removeAvatarFile(u.Avatar)
	s.removeAvatarFile(u.AvatarPending)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /admin/users/{id}/reset {send?}：生成 24 小时有效的重置链接；已配置发信且 send=true 时同时寄出
func (s *Server) adminResetLink(w http.ResponseWriter, r *http.Request) {
	var b body
	_ = readJSON(w, r, &b)
	u, err := s.DB.UserByID(pathID(r))
	if err != nil {
		fail(w, err)
		return
	}
	if u == nil {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	raw, err := s.Auth.CreateToken("reset", u.ID, "", u.Email, "admin", 24*time.Hour)
	if err != nil {
		fail(w, err)
		return
	}
	link := s.linkBase(r, true) + "/account/reset?token=" + raw
	sent := false
	if b.truthy("send") && u.Email != "" {
		site := s.Config.Typed().Site.Title
		if err := s.sendLetter(u.Email, letter{
			Subject:   "重置你在「" + site + "」的密码",
			Preheader: "站长为你生成了一个重置密码的链接。",
			Title:     "设置新密码",
			Greeting:  u.Name + "，你好：",
			Lines:     []string{"站长为你生成了一个重置密码的链接。点下面的按钮设置新密码，之后用新密码登录即可。"},
			Action:    &mailAction{Label: "设置新密码", URL: link},
			Expire:    "链接 24 小时内有效，只能使用一次。",
			Note:      "如果你并没有向站长请求重置，可以忽略这封邮件，原密码仍然有效。",
		}, s.linkBase(r, true)); err != nil {
			s.logMail(err)
		} else {
			sent = true
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"url": link, "sent": sent})
}

// GET /admin/users/export 导出 CSV（带 BOM，Excel 直接打开不乱码）
func (s *Server) adminExportUsers(w http.ResponseWriter, _ *http.Request) {
	users, counts, err := s.DB.ListUsers()
	if err != nil {
		fail(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="myself-users.csv"`)
	_, _ = w.Write([]byte("\xEF\xBB\xBF"))
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"id", "登录名", "邮箱", "昵称", "角色", "状态", "GitHub", "加入时间", "最近活跃", "评论数"})
	for _, u := range users {
		_ = cw.Write([]string{
			strconv.FormatInt(u.ID, 10), csvSafe(u.Login), csvSafe(u.Email), csvSafe(u.Name), u.Role, u.Status,
			strconv.FormatBool(u.GitHubID != 0), u.CreatedAt, u.LastActiveAt, strconv.Itoa(counts[u.ID]),
		})
	}
	cw.Flush()
}

// csvSafe 防 CSV 公式注入：以 = + - @ 开头的单元格前置单引号。
func csvSafe(v string) string {
	if v != "" && strings.ContainsRune("=+-@\t\r", rune(v[0])) {
		return "'" + v
	}
	return v
}

/* ===== 邀请 ===== */

type inviteInfo struct {
	ID        int64  `json:"id"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	Note      string `json:"note"`
	ExpiresAt string `json:"expiresAt"`
	UsedAt    string `json:"usedAt"`
	UsedBy    string `json:"usedBy"`
	CreatedAt string `json:"createdAt"`
}

// GET /admin/invites 最近的邀请（含已使用）
func (s *Server) adminListInvites(w http.ResponseWriter, _ *http.Request) {
	rows, err := s.DB.Query(`SELECT t.id, t.role, t.email, t.note, t.expires_at, COALESCE(t.used_at, ''), COALESCE(u.name, ''), t.created_at
		FROM user_tokens t LEFT JOIN users u ON u.id = t.used_by
		WHERE t.kind = 'invite' ORDER BY t.id DESC LIMIT 50`)
	if err != nil {
		fail(w, err)
		return
	}
	defer rows.Close()
	out := []inviteInfo{}
	for rows.Next() {
		var i inviteInfo
		if err := rows.Scan(&i.ID, &i.Role, &i.Email, &i.Note, &i.ExpiresAt, &i.UsedAt, &i.UsedBy, &i.CreatedAt); err != nil {
			fail(w, err)
			return
		}
		out = append(out, i)
	}
	writeJSON(w, http.StatusOK, out)
}

// POST /admin/invites {role, email?, days?, note?, send?}：返回一次性邀请链接（只显示这一次）
func (s *Server) adminCreateInvite(w http.ResponseWriter, r *http.Request) {
	var b body
	if err := readJSON(w, r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	role := b.strOr("role")
	if role != store.RoleAuthor && role != store.RoleReader {
		writeError(w, http.StatusBadRequest, "invalid_role")
		return
	}
	email := strings.ToLower(strings.TrimSpace(b.strOr("email")))
	if email != "" && !emailRe.MatchString(email) {
		writeError(w, http.StatusBadRequest, "invalid_email")
		return
	}
	days := clamp(int(b.num("days")), 1, 30)
	if b.num("days") == 0 {
		days = 7
	}
	raw, err := s.Auth.CreateToken("invite", 0, role, email, limitRunes(strings.TrimSpace(b.strOr("note")), 60), time.Duration(days)*24*time.Hour)
	if err != nil {
		fail(w, err)
		return
	}
	link := s.linkBase(r, true) + "/account/join?code=" + raw
	sent := false
	if b.truthy("send") && email != "" {
		site := s.Config.Typed().Site.Title
		roleName := map[string]string{store.RoleAuthor: "协作作者", store.RoleReader: "读者"}[role]
		if err := s.sendLetter(email, letter{
			Subject:   "邀请你加入「" + site + "」",
			Preheader: "站长邀请你以" + roleName + "的身份加入。",
			Title:     "你收到了一份邀请",
			Greeting:  "你好：",
			Lines:     []string{"站长邀请你以" + roleName + "的身份加入「" + site + "」。点下面的按钮，填好昵称和密码即可完成注册。"},
			Action:    &mailAction{Label: "接受邀请", URL: link},
			Expire:    "邀请 " + strconv.Itoa(days) + " 天内有效，只能使用一次。",
			Note:      "如果你不认识发出邀请的站点，忽略这封邮件即可。",
		}, s.linkBase(r, true)); err != nil {
			s.logMail(err)
		} else {
			sent = true
		}
	}
	writeJSON(w, http.StatusCreated, map[string]any{"url": link, "sent": sent, "days": days})
}

// DELETE /admin/invites/{id} 撤销（仅未使用的）
func (s *Server) adminDeleteInvite(w http.ResponseWriter, r *http.Request) {
	res, err := s.DB.Exec(`DELETE FROM user_tokens WHERE id = ? AND kind = 'invite' AND used_at IS NULL`, pathID(r))
	if err != nil {
		fail(w, err)
		return
	}
	okOrNotFound(w, res)
}
