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
  "hero": { "intervalMs": 3000, "count": 4, "pinnedRule": "pinned-first", "textAnim": "lightscan", "cardAnim": "hinge" },
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
      {
        "id": "m-profile",
        "type": "profile",
        "span": 3,
        "variant": "portrait",
        "data": {
          "hello": "你好，我是",
          "name": "BeaconCat",
          "lede": "在书页这一边*写字*，在工作台那一边*写代码*。",
          "bio": "独立开发者，Myself 的作者。白天做后端与工具链，夜里写字、拍照、听唱片。我相信好的工具应该像一扇门：推开就是光，不需要说明书。这个站是我的书页，也是我的工作台。",
          "status": {
            "doing": "Myself v0.9 · 关于页组件库",
            "city": "杭州 · 西湖区",
            "tz": 8
          },
          "links": [
            {
              "name": "GitHub",
              "handle": "@your-github",
              "icon": "github",
              "url": "https://github.com/your-github",
              "primary": true
            },
            {
              "name": "邮件",
              "handle": "hi@example.com",
              "icon": "mail",
              "url": "mailto:hi@example.com"
            },
            {
              "name": "RSS",
              "handle": "/feed",
              "icon": "rss",
              "url": "/feed"
            }
          ],
          "portrait": {
            "src": "",
            "fade": "left",
            "radius": 32,
            "focus": "50% 40%"
          }
        }
      },
      {
        "id": "m-ch-1",
        "type": "chapter",
        "span": 3,
        "data": {
          "no": "",
          "title": "此刻",
          "subtitle": "在做什么、在听什么、这一年写了多少。"
        }
      },
      {
        "id": "m-stats",
        "type": "stats",
        "span": 2,
        "variant": "row",
        "data": {
          "items": [
            {
              "key": "days",
              "label": "运行天数",
              "hint": "自建站起"
            },
            {
              "key": "posts",
              "label": "文章",
              "hint": "长文与教程"
            },
            {
              "key": "notes",
              "label": "随想",
              "hint": "短内容流"
            },
            {
              "key": "tags",
              "label": "标签",
              "hint": "主题索引"
            }
          ],
          "showYearProgress": true
        }
      },
      {
        "id": "m-status",
        "type": "status",
        "span": 1,
        "variant": "card",
        "data": {
          "state": "focus",
          "activity": "在写 Myself 的关于页组件",
          "app": "VS Code · about/AboutModules.vue",
          "lastActive": "",
          "device": "MacBook Pro"
        }
      },
      {
        "id": "m-now",
        "type": "now",
        "span": 1,
        "variant": "list",
        "data": {
          "updatedAt": "2026-09-20",
          "items": [
            {
              "kind": "在做",
              "text": "Myself v0.9：关于页组件库重做",
              "note": "27 个模块，12 栏 bento",
              "progress": 72
            },
            {
              "kind": "在学",
              "text": "Go 的迭代器与 range-over-func",
              "note": "顺手重写了 store 层",
              "progress": 40
            },
            {
              "kind": "在读",
              "text": "《设计中的设计》原研哉",
              "note": "读到「白」那一章",
              "progress": 58
            },
            {
              "kind": "在玩",
              "text": "Outer Wilds，第二次通关",
              "note": "这次不看攻略"
            }
          ]
        }
      },
      {
        "id": "m-github",
        "type": "github",
        "span": 2,
        "variant": "full",
        "title": "代码",
        "data": {
          "showCommits": true,
          "commitCount": 3
        }
      },
      {
        "id": "m-listening",
        "type": "listening",
        "span": 1,
        "variant": "vinyl",
        "data": {
          "playing": true,
          "now": {
            "title": "Everything In Its Right Place",
            "artist": "Radiohead",
            "album": "Kid A · 2000",
            "duration": 251,
            "position": 88
          },
          "recent": [
            {
              "title": "Holocene",
              "artist": "Bon Iver",
              "at": "12 分钟前",
              "scene": "snow"
            },
            {
              "title": "山海",
              "artist": "草东没有派对",
              "at": "38 分钟前",
              "scene": "dusk"
            },
            {
              "title": "Clair de Lune",
              "artist": "Debussy",
              "at": "1 小时前",
              "scene": "sea"
            }
          ]
        }
      },
      {
        "id": "m-languages",
        "type": "languages",
        "span": 1,
        "variant": "bar",
        "data": {
          "unit": "过去一年代码行",
          "items": [
            {
              "name": "Go",
              "percent": 38,
              "color": "#0078ff"
            },
            {
              "name": "TypeScript",
              "percent": 27,
              "color": "#00c853"
            },
            {
              "name": "Vue",
              "percent": 18,
              "color": "#ffb300"
            },
            {
              "name": "SCSS",
              "percent": 9,
              "color": "#ff0032"
            },
            {
              "name": "Markdown",
              "percent": 8,
              "color": "#8a96ab"
            }
          ]
        }
      },
      {
        "id": "m-skillbars",
        "type": "skillbars",
        "span": 1,
        "variant": "ruler",
        "title": "能力刻度",
        "data": {
          "items": [
            {
              "name": "写作",
              "level": 80
            },
            {
              "name": "编程",
              "level": 70
            },
            {
              "name": "英语",
              "level": 65
            },
            {
              "name": "摄影",
              "level": 60
            },
            {
              "name": "设计",
              "level": 55
            }
          ]
        }
      },
      {
        "id": "m-ch-2",
        "type": "chapter",
        "span": 3,
        "data": {
          "no": "",
          "title": "做过的事",
          "subtitle": "四条信条，三个项目，一条时间线。"
        }
      },
      {
        "id": "m-principles",
        "type": "principles",
        "span": 3,
        "variant": "cols",
        "data": {
          "items": [
            {
              "title": "先写下来，再写好。",
              "text": "草稿比完美重要。记录本身，就是意义。"
            },
            {
              "title": "工具要像一扇门。",
              "text": "推开就能用；说明书是设计失败的补丁。"
            },
            {
              "title": "少，但不是空。",
              "text": "删掉噪音，留下一个主角，再把它做到极致。"
            },
            {
              "title": "公开地学习。",
              "text": "把半成品也放上来——路径比结果更值得分享。"
            }
          ]
        }
      },
      {
        "id": "m-projects",
        "type": "projects",
        "span": 2,
        "variant": "feature",
        "data": {
          "items": [
            {
              "name": "Myself",
              "desc": "开源个人博客引擎。Go 单二进制 + SQLite + Vue 3，四季主题，API 中心可让 AI 发文。",
              "url": "https://github.com/BeaconCat/Myself",
              "cover": "",
              "scene": "door",
              "lang": "Go",
              "color": "#0078ff",
              "stars": 4,
              "featured": true
            },
            {
              "name": "door-light",
              "desc": "纯 CSS 光影封面生成器，没有照片也能有好封面。",
              "url": "#",
              "cover": "",
              "scene": "grid",
              "lang": "CSS",
              "color": "#ff0032",
              "stars": 0
            },
            {
              "name": "md-lint-zh",
              "desc": "中文 Markdown 排版检查：空格、标点、引号。",
              "url": "#",
              "cover": "",
              "scene": "page",
              "lang": "TypeScript",
              "color": "#00c853",
              "stars": 0
            }
          ]
        }
      },
      {
        "id": "m-milestones",
        "type": "milestones",
        "span": 1,
        "variant": "vertical",
        "data": {
          "items": [
            {
              "date": "2026.01",
              "title": "Myself 第一行代码",
              "text": "决定不用现成博客，自己造一扇门。"
            },
            {
              "date": "2026.03",
              "title": "四季主题系统",
              "text": "色盘由主色运行时派生，深浅 × 春夏秋冬。"
            },
            {
              "date": "2026.05",
              "title": "后端迁移到 Go",
              "text": "单二进制托管 SPA + API，部署只需一个文件。"
            },
            {
              "date": "2026.07",
              "title": "API 中心上线",
              "text": "APIKey 发文，把写作托管给 AI。"
            },
            {
              "date": "2026.09",
              "title": "关于页组件库 v2",
              "text": "你正在看的这一页。",
              "now": true
            }
          ]
        }
      },
      {
        "id": "m-year",
        "type": "year",
        "span": 3,
        "variant": "chart",
        "data": {
          "metric": "提交",
          "years": {
            "2026": {
              "sub": "1 月 1 日 → 今天",
              "nums": [
                [
                  "文章",
                  4
                ],
                [
                  "随想",
                  4
                ],
                [
                  "提交",
                  326
                ],
                [
                  "读完",
                  7
                ]
              ],
              "months": [
                18,
                26,
                41,
                35,
                52,
                30,
                47,
                39,
                38,
                null,
                null,
                null
              ],
              "pins": [
                0,
                6
              ],
              "highlights": [
                [
                  "一月",
                  "Myself 立项，写下第一行代码"
                ],
                [
                  "七月",
                  "API 中心上线，第一篇 AI 托管发文"
                ],
                [
                  "九月",
                  "关于页组件库 v2 进行中"
                ]
              ]
            },
            "2025": {
              "sub": "建站之前",
              "nums": [
                [
                  "文章",
                  0
                ],
                [
                  "随想",
                  0
                ],
                [
                  "提交",
                  214
                ],
                [
                  "读完",
                  11
                ]
              ],
              "months": [
                12,
                8,
                15,
                20,
                22,
                18,
                9,
                14,
                25,
                30,
                21,
                20
              ],
              "pins": [
                9
              ],
              "highlights": [
                [
                  "四月",
                  "第一台胶片相机"
                ],
                [
                  "十月",
                  "决定要有一个自己的站"
                ],
                [
                  "十二月",
                  "画出了门的 logo 草图"
                ]
              ]
            }
          }
        }
      },
      {
        "id": "m-ch-3",
        "type": "chapter",
        "span": 3,
        "data": {
          "no": "",
          "title": "喜欢的",
          "subtitle": "书、句子、照片，和去过的地方。"
        }
      },
      {
        "id": "m-bookshelf",
        "type": "bookshelf",
        "span": 2,
        "variant": "spine",
        "data": {
          "items": [
            {
              "title": "设计中的设计",
              "author": "原研哉",
              "color": "#e9e4d8",
              "textColor": "#1c2230",
              "height": 92,
              "width": 36,
              "status": "在读",
              "progress": 58,
              "note": "「白」不是空，而是可能性。"
            },
            {
              "title": "看不见的城市",
              "author": "卡尔维诺",
              "color": "#1d3a5f",
              "height": 84,
              "width": 30,
              "status": "读完",
              "note": "每座城市都是马可波罗的威尼斯。"
            },
            {
              "title": "人月神话",
              "author": "布鲁克斯",
              "color": "#8c2f2a",
              "height": 96,
              "width": 40,
              "status": "读完",
              "note": "没有银弹，但有更好的锤子。"
            },
            {
              "title": "禅与摩托车维修艺术",
              "author": "波西格",
              "color": "#2f4a3a",
              "height": 100,
              "width": 42,
              "status": "读完",
              "note": "良质先于主客体。"
            },
            {
              "title": "活着",
              "author": "余华",
              "color": "#c9a14a",
              "textColor": "#1c1a14",
              "height": 78,
              "width": 28,
              "status": "读完",
              "note": "人是为活着本身而活着。"
            },
            {
              "title": "代码大全",
              "author": "McConnell",
              "color": "#23272f",
              "height": 98,
              "width": 48,
              "status": "读完",
              "note": "写给人看的代码，顺便给机器执行。"
            },
            {
              "title": "刻意练习",
              "author": "艾利克森",
              "color": "#d8d2c4",
              "textColor": "#20252f",
              "height": 86,
              "width": 32,
              "status": "读完",
              "note": "舒适区之外一点点。"
            },
            {
              "title": "三体",
              "author": "刘慈欣",
              "color": "#0c0f16",
              "height": 90,
              "width": 44,
              "status": "想读",
              "note": "终于决定重读一遍。",
              "lean": true
            }
          ]
        }
      },
      {
        "id": "m-quotes",
        "type": "quotes",
        "span": 1,
        "variant": "rotator",
        "data": {
          "interval": 6,
          "items": [
            {
              "text": "记录本身，就是意义。",
              "from": "Myself"
            },
            {
              "text": "简单是可靠的先决条件。",
              "from": "Edsger W. Dijkstra"
            },
            {
              "text": "要耐心对待心中一切未解的问题，试着去爱这些问题本身。",
              "from": "里尔克《给青年诗人的信》"
            },
            {
              "text": "设计不只是看起来和感觉起来如何，而是它如何运作。",
              "from": "Steve Jobs"
            }
          ]
        }
      },
      {
        "id": "m-gallery",
        "type": "gallery",
        "span": 3,
        "variant": "mosaic",
        "data": {
          "images": [
            {
              "src": "",
              "scene": "door",
              "title": "门缝",
              "place": "工作室",
              "date": "2026.08.14"
            },
            {
              "src": "",
              "scene": "sea",
              "title": "落日与海",
              "place": "舟山",
              "date": "2026.07.02"
            },
            {
              "src": "",
              "scene": "window",
              "title": "午后的窗",
              "place": "杭州",
              "date": "2026.05.21"
            },
            {
              "src": "",
              "scene": "dusk",
              "title": "黄昏的街",
              "place": "上海",
              "date": "2026.04.09"
            },
            {
              "src": "",
              "scene": "tunnel",
              "title": "隧道尽头",
              "place": "重庆",
              "date": "2026.03.18"
            },
            {
              "src": "",
              "scene": "snow",
              "title": "雪夜路灯",
              "place": "哈尔滨",
              "date": "2026.01.23"
            },
            {
              "src": "",
              "scene": "page",
              "title": "书页",
              "place": "家",
              "date": "2026.02.06"
            }
          ]
        }
      },
      {
        "id": "m-places",
        "type": "places",
        "span": 2,
        "variant": "map",
        "data": {
          "region": "china",
          "items": [
            {
              "name": "杭州",
              "lon": 120.2,
              "lat": 30.3,
              "year": "常住",
              "home": true
            },
            {
              "name": "上海",
              "lon": 121.5,
              "lat": 31.2,
              "year": "2026"
            },
            {
              "name": "北京",
              "lon": 116.4,
              "lat": 39.9,
              "year": "2025"
            },
            {
              "name": "成都",
              "lon": 104.1,
              "lat": 30.7,
              "year": "2025"
            },
            {
              "name": "西安",
              "lon": 108.9,
              "lat": 34.3,
              "year": "2024"
            },
            {
              "name": "大理",
              "lon": 100.2,
              "lat": 25.6,
              "year": "2024"
            },
            {
              "name": "厦门",
              "lon": 118.1,
              "lat": 24.5,
              "year": "2023"
            },
            {
              "name": "哈尔滨",
              "lon": 126.6,
              "lat": 45.8,
              "year": "2026"
            }
          ]
        }
      },
      {
        "id": "m-favorites",
        "type": "favorites",
        "span": 1,
        "variant": "tabs",
        "data": {
          "groups": [
            {
              "title": "电影",
              "items": [
                {
                  "name": "《一一》",
                  "by": "杨德昌",
                  "year": 2000,
                  "note": "我们只能看到一半的事情"
                },
                {
                  "name": "《海街日记》",
                  "by": "是枝裕和",
                  "year": 2015
                },
                {
                  "name": "《星际穿越》",
                  "by": "诺兰",
                  "year": 2014
                },
                {
                  "name": "《千与千寻》",
                  "by": "宫崎骏",
                  "year": 2001
                }
              ]
            },
            {
              "title": "游戏",
              "items": [
                {
                  "name": "Outer Wilds",
                  "by": "Mobius",
                  "year": 2019,
                  "note": "一个只关于好奇心的游戏"
                },
                {
                  "name": "塞尔达传说：旷野之息",
                  "by": "任天堂",
                  "year": 2017
                },
                {
                  "name": "星露谷物语",
                  "by": "ConcernedApe",
                  "year": 2016
                }
              ]
            },
            {
              "title": "音乐",
              "items": [
                {
                  "name": "Kid A",
                  "by": "Radiohead",
                  "year": 2000
                },
                {
                  "name": "For Emma, Forever Ago",
                  "by": "Bon Iver",
                  "year": 2007
                },
                {
                  "name": "醜奴兒",
                  "by": "草东没有派对",
                  "year": 2016
                }
              ]
            }
          ]
        }
      },
      {
        "id": "m-ch-4",
        "type": "chapter",
        "span": 3,
        "data": {
          "no": "",
          "title": "工具与答疑",
          "subtitle": "每天在用的东西，和常被问到的问题。"
        }
      },
      {
        "id": "m-uses",
        "type": "uses",
        "span": 2,
        "variant": "groups",
        "data": {
          "groups": [
            {
              "title": "硬件",
              "items": [
                {
                  "icon": "laptop",
                  "name": "MacBook Pro 14″",
                  "desc": "M3 Pro · 36GB",
                  "tag": "主力"
                },
                {
                  "icon": "monitor",
                  "name": "LG 27UP850",
                  "desc": "4K · 读代码不累"
                },
                {
                  "icon": "keyboard",
                  "name": "Keychron Q1 Pro",
                  "desc": "香草奶昔轴"
                },
                {
                  "icon": "camera",
                  "name": "Sony A7C II",
                  "desc": "+ 35mm F1.8",
                  "tag": "街拍"
                }
              ]
            },
            {
              "title": "软件",
              "items": [
                {
                  "icon": "code",
                  "name": "VS Code",
                  "desc": "Go + Volar",
                  "tag": "每天"
                },
                {
                  "icon": "terminal",
                  "name": "Ghostty",
                  "desc": "终端"
                },
                {
                  "icon": "figma",
                  "name": "Figma",
                  "desc": "画门与光"
                },
                {
                  "icon": "command",
                  "name": "Raycast",
                  "desc": "启动一切"
                }
              ]
            }
          ]
        }
      },
      {
        "id": "m-stack",
        "type": "stack",
        "span": 1,
        "variant": "grid",
        "data": {
          "items": [
            {
              "name": "Go",
              "role": "后端 · 单二进制",
              "glyph": "Go",
              "color": "#00add8"
            },
            {
              "name": "SQLite",
              "role": "存储 · 纯 Go",
              "glyph": "SQ",
              "color": "#4a8fd6"
            },
            {
              "name": "Vue 3",
              "role": "前端框架",
              "glyph": "Vu",
              "color": "#42b883"
            },
            {
              "name": "Vite",
              "role": "构建",
              "glyph": "Vi",
              "color": "#b58cff"
            },
            {
              "name": "TypeScript",
              "role": "类型",
              "glyph": "TS",
              "color": "#3178c6"
            },
            {
              "name": "Markdown",
              "role": "内容规范",
              "glyph": "Md",
              "color": "#8a96ab"
            }
          ]
        }
      },
      {
        "id": "m-skills",
        "type": "skills",
        "span": 1,
        "variant": "keys",
        "data": {
          "groups": [
            {
              "title": "创作",
              "items": [
                "文章",
                "随想",
                "摄影"
              ],
              "star": "文章"
            },
            {
              "title": "工具",
              "items": [
                "Markdown",
                "主题系统",
                "API 中心",
                "Go",
                "Vue"
              ],
              "star": "Go"
            },
            {
              "title": "兴趣",
              "items": [
                "阅读",
                "旅行",
                "音乐",
                "胶片"
              ],
              "star": "胶片"
            }
          ]
        }
      },
      {
        "id": "m-faq",
        "type": "faq",
        "span": 2,
        "variant": "accordion",
        "data": {
          "single": true,
          "items": [
            {
              "q": "为什么叫 Myself？",
              "a": "因为它首先是写给自己的。博客是一扇门：门里是记录，门外才是读者。"
            },
            {
              "q": "这个站是用什么搭的？",
              "a": "前端 Vue 3 + Vite，后端 Go + SQLite，打包成一个单二进制文件。内容全部是 Markdown，可以随时带走。"
            },
            {
              "q": "可以转载文章吗？",
              "a": "欢迎，署名并附上原文链接即可（CC BY-NC 4.0）。商业用途请先发邮件聊聊。"
            },
            {
              "q": "能让 AI 帮我往 Myself 里发文吗？",
              "a": "可以。后台「API 中心」创建 APIKey，请求头带上 \u0060X-Api-Key\u0060，POST 到 \u0060/api/v1/posts\u0060 即可。"
            },
            {
              "q": "随想和文章有什么区别？",
              "a": "文章是长文，有封面和目录；随想是短内容流，一段话加最多九张图，更像是写给明天的便签。"
            }
          ]
        }
      },
      {
        "id": "m-ch-5",
        "type": "chapter",
        "span": 3,
        "data": {
          "no": "",
          "title": "找到我",
          "subtitle": "留言、写信，或者只是路过。"
        }
      },
      {
        "id": "m-guestbook",
        "type": "guestbook",
        "span": 2,
        "variant": "wall",
        "data": {
          "pageSize": 4,
          "requireLogin": false,
          "total": 38,
          "items": [
            {
              "name": "阿澄",
              "color": "#0078ff",
              "at": "2 小时前",
              "text": "门的 logo 太好看了，四季色盘切到秋天那一下很惊艳。",
              "likes": 6,
              "reply": "谢谢！秋天的黄是我最喜欢的一组。"
            },
            {
              "name": "Kite",
              "color": "#00c853",
              "at": "昨天",
              "text": "API 中心那篇写得很清楚，已经让我的脚本每天发一条随想了。",
              "likes": 3
            },
            {
              "name": "林间",
              "color": "#ffb300",
              "at": "3 天前",
              "text": "「先写下来，再写好」——今天也写了一点。",
              "likes": 9
            },
            {
              "name": "夜航船",
              "color": "#ff0032",
              "at": "上周",
              "text": "Outer Wilds 同好！第二次通关不看攻略真的是另一种体验。",
              "likes": 2
            }
          ]
        }
      },
      {
        "id": "m-contact",
        "type": "contact",
        "span": 1,
        "variant": "card",
        "data": {
          "title": "想聊聊？",
          "text": "合作、提问，或者只是打个招呼，都欢迎。",
          "email": "hi@example.com",
          "buttonText": "写邮件",
          "url": "mailto:hi@example.com",
          "sla": "通常 24 小时内回复"
        }
      },
      {
        "id": "m-motto",
        "type": "motto",
        "span": 3,
        "variant": "closing",
        "data": {
          "text": "记录本身，就是意义。",
          "sign": "MYSELF · SINCE 2026"
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
