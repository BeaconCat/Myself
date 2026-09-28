package store

// 演示种子数据：围绕 Myself 平台自身功能，便于新用户开箱即见示例内容。

type seedNote struct {
	ContentMd string
	Mood      string
	Images    []string
	CreatedAt string
}

type seedPost struct {
	Slug      string
	Title     string
	Excerpt   string
	Tags      []string
	Covers    []string
	CreatedAt string
	ContentMd string
}

func img(from, to, label string) string {
	return "/api/v1/img/" + from + "/" + to + "/" + label
}

var seedNotes = []seedNote{
	{
		ContentMd: "欢迎使用 **Myself**。这是一条「随想」——短内容流，支持 Markdown、心情标签和最多九张配图。",
		Mood:      "欢迎",
		Images:    []string{img("0078ff", "00295c", "Welcome")},
		CreatedAt: "2026-07-10 12:00:00",
	},
	{
		ContentMd: "随想的配图支持双拼、三拼、四宫格直到九宫格，点开可全屏缩放拖拽查看。",
		Mood:      "功能",
		Images: []string{
			img("ff0032", "7a0020", "Grid-1"),
			img("ffb300", "7a5200", "Grid-2"),
			img("00c853", "00512a", "Grid-3"),
			img("0078ff", "00295c", "Grid-4"),
		},
		CreatedAt: "2026-07-09 18:30:00",
	},
	{
		ContentMd: "试试右上角的主题切换：四季色盘 × 深浅模式，切换时有圆形蒙版扩散动画。",
		Mood:      "主题",
		CreatedAt: "2026-07-08 09:15:00",
	},
	{
		ContentMd: "后台「API 中心」可以创建 APIKey，把发文托管给你的 AI 助手——内容统一走 Markdown。",
		Mood:      "API",
		CreatedAt: "2026-07-07 21:40:00",
	},
}

var seedPosts = []seedPost{
	{
		Slug:      "welcome-to-myself",
		Title:     "欢迎使用 Myself",
		Excerpt:   "一个开源的个人博客引擎：Vue 3 + Go + SQLite，内容皆 Markdown，主题随四季流转。",
		Tags:      []string{"指南"},
		CreatedAt: "2026-07-01 10:00:00",
		ContentMd: `# 欢迎使用 Myself

Myself 是一个开源的个人博客引擎，前端 Vue 3 + Vite，后端 Go + SQLite，所有内容以 **Markdown** 统一存储。

## 快速开始

1. 登录后台（` + "`/admin`" + `，默认账号 ` + "`admin / myself-admin`" + `，请立即修改密码）
2. 在「身份」里设置头像、名字、签名与格言，在「外观」里挑选色盘、深浅与界面风格
3. 点左上角「写文章」，发布你的第一篇文章

> 这篇文章本身就是演示数据，可以在后台随时删除。`,
	},
	{
		Slug:      "theme-system-guide",
		Title:     "主题系统指南",
		Excerpt:   "三维正交的主题：深浅模式 × 色盘 × 界面风格；只需选一个主色，整套颜色自动推导。",
		Tags:      []string{"指南", "主题"},
		CreatedAt: "2026-07-02 14:00:00",
		ContentMd: `# 主题系统指南

主题 = **模式**（浅色 / 深色）×**色盘**（最多十组预设）×**界面风格**（简洁 / 卡片），三者互不影响。

## 自定义

后台「外观」中：

- 修改每组色盘的主色，整套界面色由主色自动派生
- 拖拽调整顺序，前 N 组对访客可见；也可以按季节自动切换
- 选择默认界面风格：简洁为透明背景与发丝线，卡片为高密度卡片与统计条
- 调整全站圆角基准，按钮、卡片与输入框一起变化
- 可关闭访客换肤或切换风格，只保留深浅切换

` + "```css" + `
/* 组件只消费变量，换主题零改动 */
color: var(--primary);
` + "```",
	},
	{
		Slug:      "writing-with-markdown",
		Title:     "用 Markdown 写作",
		Excerpt:   "文章与随想共用一条 Markdown 管线：可移植、可版本化、对 AI 友好。",
		Tags:      []string{"指南", "写作"},
		CreatedAt: "2026-07-03 09:30:00",
		ContentMd: `# 用 Markdown 写作

数据库只存 Markdown 原文，渲染发生在展示端。

| 能力 | 说明 |
| --- | --- |
| 代码块 | 支持围栏语法 |
| 表格 | 就像这张 |
| 引用 | 见下方 |

> 内容不锁在任何编辑器里——随时导出，随时迁移。`,
	},
	{
		Slug:      "api-center-for-agents",
		Title:     "把发文托管给 AI：API 中心",
		Excerpt:   "创建 APIKey，让外部 AI 通过 X-Api-Key 认证的 REST 接口替你发布文章与随想。",
		Tags:      []string{"指南", "API"},
		CreatedAt: "2026-07-04 16:20:00",
		ContentMd: `# 把发文托管给 AI：API 中心

后台「API 中心」创建 Key 后，外部程序即可投稿：

` + "```bash" + `
curl -X POST https://your-site/api/v1/ext/posts \
  -H "X-Api-Key: myk_xxx" \
  -H "Content-Type: application/json" \
  -d '{"slug":"hello","title":"你好","contentMd":"# 来自 AI 的第一篇"}'
` + "```" + `

- Key 只在创建时显示一次，哈希入库
- 可随时吊销，带最后使用时间追踪
- 外部投稿默认进草稿箱，人工审核后发布`,
	},
}

// SeedIfEmpty 在 posts / notes 表为空时分别注入演示数据。
func (db *DB) SeedIfEmpty() error {
	if err := db.seedNotesIfEmpty(); err != nil {
		return err
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM posts`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, p := range seedPosts {
		_, err := tx.Exec(`INSERT INTO posts (slug, title, excerpt, content_md, covers, tags, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			p.Slug, p.Title, p.Excerpt, p.ContentMd, JSONStrings(p.Covers), JSONStrings(p.Tags), p.CreatedAt)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (db *DB) seedNotesIfEmpty() error {
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM notes`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, note := range seedNotes {
		_, err := tx.Exec(`INSERT INTO notes (content_md, mood, images, created_at) VALUES (?, ?, ?, ?)`,
			note.ContentMd, note.Mood, JSONStrings(note.Images), note.CreatedAt)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
