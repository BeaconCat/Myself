package store

import (
	_ "embed"
	"encoding/json"
	"regexp"
	"time"
)

// demoAbout 示例关于页模块，与 client/src/about/demo-modules.json 保持同步。
//
//go:embed demo_about.json
var demoAbout []byte

// demoCoverRe Demo 内容引用的默认封面地址（前端 public/covers）。
var demoCoverRe = regexp.MustCompile(`/covers/[0-9]{2}\.webp`)

// AssetMapper 改写 Demo 内容里的图片地址（初始化时把默认封面导入素材库）；nil 表示原样保留。
type AssetMapper func(url string) string

func (m AssetMapper) apply(urls []string) []string {
	if m == nil {
		return urls
	}
	out := make([]string, len(urls))
	for i, u := range urls {
		out[i] = m(u)
	}
	return out
}

// DemoAboutModules 示例关于页模块（每次返回新副本），其中的默认封面地址经 mapAsset 改写。
func DemoAboutModules(mapAsset AssetMapper) ([]any, error) {
	raw := demoAbout
	if mapAsset != nil {
		raw = demoCoverRe.ReplaceAllFunc(raw, func(u []byte) []byte { return []byte(mapAsset(string(u))) })
	}
	var mods []any
	err := json.Unmarshal(raw, &mods)
	return mods, err
}

type seedComment struct {
	Guest string
	Body  string
	Ago   time.Duration
}

// 留言墙示例（访客留言，已公开）
var seedGuestbook = []seedComment{
	{"青柠", "页面好安静，读起来很舒服。宋体标题配黑体正文，一点也不违和。", 2 * time.Hour},
	{"Kite", "API 中心那篇写得很清楚，我已经让脚本每周五往这里投一篇草稿了。", 26 * time.Hour},
	{"林间", "「先写下来，再写好」，今天也写了一点。", 3 * day},
	{"夜航船", "喜欢那张远帆的封面，像是周末早上的海。", 6 * day},
}

// 默认 Demo 数据：首次启动初始化时可选注入。封面与配图取自前端内置的 /covers/NN.webp（程序生成的抽象封面），
// 时间按「现在」往前倒推，开箱即是一个近期有更新的站点。内容围绕 Myself 自身功能 + 两篇随笔示范阅读排版。

type seedNote struct {
	ContentMd string
	Mood      string
	Images    []string
	Pinned    bool
	Ago       time.Duration
}

type seedPost struct {
	Slug      string
	Title     string
	Excerpt   string
	Tags      []string
	Covers    []string
	Pinned    bool
	Ago       time.Duration
	ContentMd string
}

const day = 24 * time.Hour

func cover(n string) string { return "/covers/" + n + ".webp" }

var seedPosts = []seedPost{
	{
		Slug:    "welcome-to-myself",
		Title:   "欢迎来到 Myself",
		Excerpt: "一个安静的个人站点：文章、随想与关于页，内容皆 Markdown。五分钟把它变成你的。",
		Tags:    []string{"指南"},
		Covers:  []string{cover("05")},
		Pinned:  true,
		Ago:     2 * time.Hour,
		ContentMd: `Myself 是一个开源的个人博客引擎：前端 Vue 3，后端 Go + SQLite，打包后是一个二进制文件。所有内容都以 **Markdown** 存储，随时可以带走。

## 五分钟上手

1. 打开后台「身份」，换上你的头像、名字、签名与格言——关于页、页脚与文章作者栏都会一起更新
2. 在「外观」里挑一组色盘，决定深浅模式与界面风格（简洁 / 卡片）
3. 点左上角「写文章」，写下第一篇；短一点的想法，就发在「随想」里
4. 「关于」页由模块拼成，拖动排序、拖动右缘改宽度，所见即所得

## 这些示例内容

站点里现在的文章、随想和封面都是示例，可以在后台逐条删除，也可以留着当作参考。

> 记录本身，就是意义。`,
	},
	{
		Slug:    "markdown-at-a-glance",
		Title:   "一页看懂 Markdown 在这里的样子",
		Excerpt: "标题、列表、引用、代码、表格与图片——写作页支持快捷输入，也可以随时切到源码。",
		Tags:    []string{"写作", "指南"},
		Covers:  []string{cover("03")},
		Ago:     1*day + 5*time.Hour,
		ContentMd: `写作页是所见即所得的编辑器，底层仍然是 Markdown。输入 ` + "`## `" + ` 变成标题，输入 ` + "`> `" + ` 变成引用。

## 段落与强调

正文使用思源宋体，行距舒展。可以 **加粗**、*倾斜*、~~删除~~，也可以写 ` + "`行内代码`" + `，或者放一个[链接](/articles)。

## 列表

- 无序列表适合罗列
- 有序列表适合步骤
  - 也可以嵌套

- [x] 待办清单
- [ ] 还没做完的事

## 引用

> 好的工具应该像一扇门：推开就是光，不需要说明书。

## 代码

` + "```go" + `
// 一个最小的 HTTP 服务
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello, myself")
	})
	log.Fatal(http.ListenAndServe(":3100", nil))
}
` + "```" + `

` + "```ts" + `
const greet = (name: string): string => ` + "`你好，${name}`" + `;
` + "```" + `

## 表格

| 能力 | 说明 |
| --- | --- |
| 代码高亮 | 离线的 highlight.js |
| 目录 | 自动从二级、三级标题生成 |
| 图片 | 点击放大，支持缩放拖拽 |

## 图片

![叠纸](/covers/03.webp)

写完之后，在右上角的发布面板里检查标题、摘要、封面与标签，再按下发布。`,
	},
	{
		Slug:    "dusk-on-the-horizon",
		Title:   "黄昏的地平线",
		Excerpt: "每天有十几分钟，天空把颜色一层层交出来。我试着把这段时间留给自己。",
		Tags:    []string{"随笔"},
		Covers:  []string{cover("01"), cover("11")},
		Ago:     3*day + 2*time.Hour,
		ContentMd: `傍晚六点多，窗外的楼顶开始发亮。先是一层淡淡的杏色，然后是橘红，最后沉成一种说不清的灰蓝。

这段时间很短，短到来不及拿出相机。后来我索性不拍了，只是站在窗边看着它发生。

## 慢一点

一天里大部分时间都在赶路：赶地铁、赶需求、赶在睡前把消息回完。黄昏是少数不需要做任何事的时刻——太阳落下去这件事，不需要我参与。

> 有些事情，只要在场就够了。

## 记下来

我开始在随想里记一句话：今天的天是什么颜色，风从哪边来，楼下的猫有没有出现。没有什么意义，但过了几个月再翻，会发现那些普通的傍晚一个都没丢。

这大概就是写博客的理由。`,
	},
	{
		Slug:    "themes-palettes-and-styles",
		Title:   "深浅、色盘与界面风格",
		Excerpt: "三个维度互不影响：只选一个主色，整套颜色自动推导；简洁与卡片两种风格随时切换。",
		Tags:    []string{"主题", "指南"},
		Covers:  []string{cover("10")},
		Ago:     5*day + 7*time.Hour,
		ContentMd: `Myself 的外观由三个互不影响的维度组成：

- **模式**：浅色 / 深色
- **色盘**：内置四季四组，后台最多可以自定义十组
- **界面风格**：简洁（透明背景、发丝线分隔）或卡片（高密度卡片与统计条）

## 只选一个主色

每组色盘只需要一个主色。强调色、实底按钮的文字颜色、深浅两套底色，都会按对比度自动推导，亮黄这样的主色也能配上清晰的深色文字。

## 全站圆角

在「外观」里拖动圆角基准，按钮、卡片、输入框与弹层会一起变化，右侧实时预览。

## 访客能做什么

可以允许访客自己切换色盘与风格，也可以锁定为站点默认，只保留深浅切换。`,
	},
	{
		Slug:    "api-center-for-agents",
		Title:   "把发文交给 AI：API 中心",
		Excerpt: "创建一把 Key，外部助手就能通过 REST 接口投稿。默认只能写草稿，由你审阅发布。",
		Tags:    []string{"API", "指南"},
		Covers:  []string{cover("06")},
		Ago:     8*day + 3*time.Hour,
		ContentMd: `后台「API 中心」可以为外部 AI 或脚本创建 Key，通过 ` + "`X-Api-Key`" + ` 请求头调用外部通道。

## 两种权限

| 权限 | 能做什么 |
| --- | --- |
| 仅投稿（默认） | 创建草稿文章，只能读取和修改自己创建的草稿 |
| 全托管 | 读写全部文章与随想，可以直接发布 |

## 投一篇草稿

` + "```bash" + `
curl -X POST https://your-site/api/v1/ext/posts \
  -H "X-Api-Key: myk_xxx" \
  -H "Content-Type: application/json" \
  -d '{"slug":"hello","title":"你好","contentMd":"# 来自助手的第一篇"}'
` + "```" + `

- Key 只在创建时显示一次，服务端只保存哈希
- 可以随时吊销，并能看到最后一次使用的时间
- 草稿会出现在「今天」的待办里，审阅后一键发布`,
	},
	{
		Slug:    "twelve-books-on-the-shelf",
		Title:   "书架上的十二本书",
		Excerpt: "搬了三次家，书越来越少。留下来的这些，每一本都有一个不舍得扔的理由。",
		Tags:    []string{"阅读", "随笔"},
		Covers:  []string{cover("09"), cover("08"), cover("05")},
		Ago:     13*day + 6*time.Hour,
		ContentMd: `每次搬家都会清掉一批书。第一次清掉的是教材，第二次是买来没读的，第三次是读过但不会再读的。

留下来的十二本，大概就是现在的我。

## 反复读的

1. 一本关于写作的小书，每次卡壳都会翻开
2. 一本旅行随笔，读的时候像在别人的窗边坐了一下午
3. 一本讲设计的书，教会我「少即是多」不是口号

## 还没读完的

有两本读到一半。它们留着不是因为好看，而是因为「总有一天」——这四个字，大概是书架上最重的东西。

> 书架不必满，满了反而看不见。`,
	},
}

var seedNotes = []seedNote{
	{ContentMd: "欢迎来到 **Myself**。这里是「随想」：短到装不下一篇文章的念头，支持 Markdown、心情标签和最多九张配图。", Mood: "记录", Pinned: true, Ago: 40 * time.Minute},
	{ContentMd: "今天的天是杏色的，很快就沉成了灰蓝。", Mood: "平静", Images: []string{cover("01")}, Ago: 20 * time.Hour},
	{ContentMd: "配图会按张数自动排成双拼、三拼或九宫格，点开可以全屏缩放拖拽。", Mood: "灵感", Images: []string{cover("05"), cover("08"), cover("02")}, Ago: 2*day + 4*time.Hour},
	{ContentMd: "试试右上角的外观切换：色盘、深浅、界面风格，三个维度随意组合。", Mood: "欢喜", Ago: 4*day + 1*time.Hour},
	{ContentMd: "下雨天适合整理书架，也适合什么都不整理。", Mood: "平静", Images: []string{cover("11"), cover("09")}, Ago: 6*day + 9*time.Hour},
	{ContentMd: "给助手开了一把「仅投稿」的 Key，它写的草稿都在待办里等我审阅。", Mood: "灵感", Ago: 9*day + 2*time.Hour},
	{ContentMd: "关于页的模块可以拖动排序、拖动右缘改宽度。把最想让人看到的放在最前面。", Mood: "记录", Images: []string{cover("06")}, Ago: 11*day + 5*time.Hour},
	{ContentMd: "记录本身，就是意义。", Mood: "记录", Images: []string{cover("05"), cover("04"), cover("03"), cover("07")}, Ago: 15 * day},
}

func stamp(ago time.Duration) string {
	return time.Now().UTC().Add(-ago).Format("2006-01-02 15:04:05")
}

// SeedDemo 注入 Demo 文章、随想与留言墙示例（仅在对应表为空时）；封面与配图地址经 mapAsset 改写。
// 关于页示例模块见 DemoAboutModules。
func (db *DB) SeedDemo(mapAsset AssetMapper) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var n int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM posts`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		for _, p := range seedPosts {
			at := stamp(p.Ago)
			if _, err := tx.Exec(`INSERT INTO posts (slug, title, excerpt, content_md, covers, tags, pinned, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				p.Slug, p.Title, p.Excerpt, p.ContentMd, JSONStrings(mapAsset.apply(p.Covers)), JSONStrings(p.Tags), boolInt(p.Pinned), at, at); err != nil {
				return err
			}
		}
	}
	if err := tx.QueryRow(`SELECT COUNT(*) FROM notes`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		for _, note := range seedNotes {
			if _, err := tx.Exec(`INSERT INTO notes (content_md, mood, images, pinned, created_at) VALUES (?, ?, ?, ?, ?)`,
				note.ContentMd, note.Mood, JSONStrings(mapAsset.apply(note.Images)), boolInt(note.Pinned), stamp(note.Ago)); err != nil {
				return err
			}
		}
	}
	if err := tx.QueryRow(`SELECT COUNT(*) FROM comments`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		for _, c := range seedGuestbook {
			if _, err := tx.Exec(`INSERT INTO comments (target, target_id, guest_name, body, status, created_at) VALUES ('guestbook', 0, ?, ?, 'approved', ?)`,
				c.Guest, c.Body, stamp(c.Ago)); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
