// Package config 管理站点配置：settings 表 key = 'site_config'，整体 JSON。
// 配置以松散 map 存储（前端自由扩展字段），服务端只对自身用到的键做类型读取。
package config

import (
	"encoding/json"
	"log"
	"maps"
	"strings"

	"myself/server/internal/store"
)

// Map 是配置对象的通用表示。
type Map = map[string]any

const defaultJSON = `{
  "site": {
    "title": "Myself",
    "subtitle": "个人博客",
    "listEndText": "—— 到底啦 ——"
  },
  "loading": { "bootText": "Myself", "routeText": "加载中" },
  "theme": {
    "defaultPaletteId": "summer",
    "defaultMode": "dark",
    "autoSwitch": "off",
    "allowUserPalette": true,
    "displayCount": 4,
    "presets": [
      { "id": "spring", "name": "春 · 新绿", "primary": "#00c853", "primaryDeep": "#00a344" },
      { "id": "summer", "name": "夏 · 炽红", "primary": "#ff0032", "primaryDeep": "#d40029" },
      { "id": "autumn", "name": "秋 · 暖阳", "primary": "#ffb300", "primaryDeep": "#e09600" },
      { "id": "winter", "name": "冬 · 霜蓝", "primary": "#0078ff", "primaryDeep": "#005fd6" }
    ]
  },
  "hero": { "intervalMs": 3000, "count": 4, "pinnedRule": "pinned-first" },
  "thoughts": { "subtitle": "碎片化的想法、心情与瞬间，短到装不下一篇文章。" },
  "covers": { "expandMs": 10000 },
  "timezone": "Asia/Shanghai",
  "github": {
    "username": "your-github",
    "mode": "manual",
    "token": "",
    "refreshMinutes": 30,
    "proxy": "",
    "insecureTls": false,
    "stats": { "repos": 0, "stars": 0, "followers": 0, "commits": 0 }
  },
  "about": {
    "avatar": "",
    "name": "Myself",
    "tagline": "开源个人博客引擎",
    "bio": "这里是 Myself 的默认介绍。前往后台「关于管理」写下你自己的故事：你是谁、在做什么、热爱什么。",
    "skills": ["写作", "摄影", "编程"],
    "foundedAt": "2026-01-01",
    "motto": "记录本身，就是意义。",
    "modules": [
      { "id": "m-stats", "type": "stats", "data": {} },
      {
        "id": "m-skills",
        "type": "skills",
        "data": {
          "groups": [
            { "title": "创作", "items": ["文章", "随想", "摄影"] },
            { "title": "工具", "items": ["Markdown", "主题系统", "API 中心"] },
            { "title": "兴趣", "items": ["阅读", "旅行", "音乐"] }
          ]
        }
      },
      {
        "id": "m-milestones",
        "type": "milestones",
        "data": {
          "items": [
            { "year": "01", "text": "在后台「设置」里换上你的名字、简介与主题色" },
            { "year": "02", "text": "发布第一篇文章，或用随想记录此刻" },
            { "year": "03", "text": "创建 APIKey，把日常发文托管给你的 AI 助手" }
          ]
        }
      },
      { "id": "m-github", "type": "github", "data": {} },
      {
        "id": "m-socials",
        "type": "socials",
        "data": {
          "items": [
            { "name": "GitHub", "url": "https://github.com/your-github", "icon": "github" },
            { "name": "Email", "url": "mailto:hi@example.com", "icon": "mail" },
            { "name": "RSS", "url": "/feed", "icon": "rss" }
          ]
        }
      },
      {
        "id": "m-stack",
        "type": "stack",
        "data": {
          "items": [
            { "name": "Vue 3", "role": "前端框架" },
            { "name": "Vite", "role": "构建工具" },
            { "name": "Go", "role": "API 服务" },
            { "name": "SQLite", "role": "数据存储" },
            { "name": "Markdown", "role": "内容规范" }
          ]
        }
      }
    ]
  },
  "backup": { "autoHours": 0 }
}`

// Default 返回默认配置的全新副本。
func Default() Map {
	var m Map
	if err := json.Unmarshal([]byte(defaultJSON), &m); err != nil {
		panic("config: default json invalid: " + err.Error())
	}
	return m
}

// DeepMerge 深合并：对象递归，数组/标量整体替换。
func DeepMerge(base, patch any) any {
	bm, bok := base.(Map)
	pm, pok := patch.(Map)
	if !bok || !pok {
		return patch
	}
	out := make(Map, len(bm)+len(pm))
	maps.Copy(out, bm)
	for k, v := range pm {
		out[k] = DeepMerge(out[k], v)
	}
	return out
}

// migrateAboutModules 旧版扁平 about 字段 → 模块化区块。
func migrateAboutModules(cfg Map) Map {
	about, _ := cfg["about"].(Map)
	if about == nil {
		about = Map{}
		cfg["about"] = about
	}
	if mods, ok := about["modules"].([]any); ok && len(mods) > 0 {
		return cfg
	}
	defaults, _ := Default()["about"].(Map)["modules"].([]any)
	for _, raw := range defaults {
		mod, _ := raw.(Map)
		data, _ := mod["data"].(Map)
		switch mod["type"] {
		case "skills":
			if v, ok := about["skillGroups"].([]any); ok {
				data["groups"] = v
			}
		case "milestones":
			if v, ok := about["milestones"].([]any); ok {
				data["items"] = v
			}
		case "socials":
			if v, ok := about["socials"].([]any); ok {
				data["items"] = v
			}
		}
	}
	about["modules"] = defaults
	return cfg
}

// Service 读写站点配置。
type Service struct {
	db *store.DB
}

// New 构造配置服务。
func New(db *store.DB) *Service {
	return &Service{db: db}
}

// Get 返回默认值与库内配置深合并后的完整配置。
func (s *Service) Get() Map {
	raw, err := s.db.GetSetting("site_config")
	if err != nil {
		log.Printf("[config] read: %v", err)
	}
	if strings.TrimSpace(raw) == "" {
		return Default()
	}
	var stored Map
	if err := json.Unmarshal([]byte(raw), &stored); err != nil {
		return Default()
	}
	merged, _ := DeepMerge(Default(), stored).(Map)
	return migrateAboutModules(merged)
}

// Save 深合并补丁并持久化，返回合并结果。
func (s *Service) Save(patch Map) (Map, error) {
	merged, _ := DeepMerge(s.Get(), patch).(Map)
	b, err := json.Marshal(merged)
	if err != nil {
		return nil, err
	}
	return merged, s.db.SetSetting("site_config", string(b))
}

// Sub 取子对象；不存在返回空 map。
func Sub(m Map, key string) Map {
	v, _ := m[key].(Map)
	if v == nil {
		return Map{}
	}
	return v
}

// Str 取字符串字段。
func Str(m Map, key string) string {
	v, _ := m[key].(string)
	return v
}

// Num 取数字字段（JSON number → float64）；非数字返回 0。
func Num(m Map, key string) float64 {
	switch v := m[key].(type) {
	case float64:
		return v
	case string:
		var f float64
		if err := json.Unmarshal([]byte(v), &f); err == nil {
			return f
		}
	}
	return 0
}

// Bool 取布尔字段。
func Bool(m Map, key string) bool {
	v, _ := m[key].(bool)
	return v
}

/* ===== 服务端消费的强类型视图 ===== */

// Hero 首页轮播规则。
type Hero struct {
	IntervalMs int    `json:"intervalMs"`
	Count      int    `json:"count"`
	PinnedRule string `json:"pinnedRule"`
}

// GitHub 状态拉取配置。
type GitHub struct {
	Username       string `json:"username"`
	Mode           string `json:"mode"`
	Token          string `json:"token"`
	RefreshMinutes int    `json:"refreshMinutes"`
	Proxy          string `json:"proxy"`
	InsecureTLS    bool   `json:"insecureTls"`
	Stats          any    `json:"stats"`
}

// Site 站点基础信息。
type Site struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

// Typed 是后端逻辑用到的配置子集；前端专用字段留在 Map 里透传。
type Typed struct {
	Site     Site   `json:"site"`
	Hero     Hero   `json:"hero"`
	GitHub   GitHub `json:"github"`
	Timezone string `json:"timezone"`
	Backup   struct {
		AutoHours float64 `json:"autoHours"`
	} `json:"backup"`
}

// Typed 返回强类型配置（经 JSON 往返，容忍前端写入的字符串数字）。
func (s *Service) Typed() Typed {
	return ToTyped(s.Get())
}

// ToTyped Map → Typed。数字字段若被存成字符串则按 0 处理，调用方需兜底默认值。
func ToTyped(m Map) Typed {
	var t Typed
	raw, _ := json.Marshal(m)
	_ = json.Unmarshal(raw, &t)
	return t
}
