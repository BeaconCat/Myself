package httpapi

import (
	"regexp"
	"strconv"
	"strings"
)

// parseFrontMatter 拆出 front matter 与正文。支持 YAML（---）与 TOML（+++）两种围栏，
// 以及 Hexo 允许的「省略开头 ---」写法。只解析顶层的标量 / 列表字段（导入需要的都在顶层），嵌套表忽略。
func parseFrontMatter(src string) (map[string]any, string) {
	src = strings.TrimPrefix(strings.ReplaceAll(src, "\r\n", "\n"), "\ufeff")
	lines := strings.Split(src, "\n")
	if len(lines) == 0 {
		return map[string]any{}, src
	}
	fence := strings.TrimSpace(lines[0])
	start := 1
	if fence != "---" && fence != "+++" {
		// Hexo：front matter 可以没有开头的 ---，以第一行 `key: value` 起、`---` 止
		if !yamlKeyRe.MatchString(lines[0]) {
			return map[string]any{}, src
		}
		fence, start = "---", 0
	}
	for i := start; i < len(lines) && i < 400; i++ {
		t := strings.TrimSpace(lines[i])
		if t == fence || (fence == "---" && t == "...") {
			head := lines[start:i]
			body := strings.Join(lines[i+1:], "\n")
			if fence == "+++" {
				return parseTOML(head), body
			}
			return parseYAML(head), body
		}
	}
	return map[string]any{}, src
}

var (
	yamlKeyRe  = regexp.MustCompile(`^([A-Za-z_][\w-]*)\s*:\s*(.*)$`)
	tomlKeyRe  = regexp.MustCompile(`^([A-Za-z_][\w-]*)\s*=\s*(.*)$`)
	listItemRe = regexp.MustCompile(`^\s*-\s+(.*)$`)
)

func parseYAML(lines []string) map[string]any {
	out := map[string]any{}
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		m := yamlKeyRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		key, raw := strings.ToLower(m[1]), strings.TrimSpace(stripComment(m[2]))
		switch {
		case raw == "" || raw == "|" || raw == ">" || raw == "|-" || raw == ">-":
			// 下方缩进块：列表项或多行文本
			var items []any
			var text []string
			for i+1 < len(lines) && (strings.HasPrefix(lines[i+1], " ") || strings.HasPrefix(lines[i+1], "\t") || strings.TrimSpace(lines[i+1]) == "" || listItemRe.MatchString(lines[i+1])) {
				i++
				if lm := listItemRe.FindStringSubmatch(lines[i]); lm != nil && raw == "" {
					items = append(items, scalar(stripComment(lm[1])))
				} else if s := strings.TrimSpace(lines[i]); s != "" {
					text = append(text, s)
				}
				if i+1 < len(lines) && yamlKeyRe.MatchString(lines[i+1]) {
					break
				}
			}
			if len(items) > 0 {
				out[key] = items
			} else if len(text) > 0 && raw != "" {
				sep := " "
				if strings.HasPrefix(raw, "|") {
					sep = "\n"
				}
				out[key] = strings.Join(text, sep)
			}
		case strings.HasPrefix(raw, "["):
			out[key] = inlineList(raw)
		default:
			out[key] = scalar(raw)
		}
	}
	return out
}

func parseTOML(lines []string) map[string]any {
	out := map[string]any{}
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "[") {
			break // 之后是子表，导入用不到
		}
		m := tomlKeyRe.FindStringSubmatch(t)
		if m == nil {
			continue
		}
		key, raw := strings.ToLower(m[1]), strings.TrimSpace(stripComment(m[2]))
		if strings.HasPrefix(raw, "[") {
			out[key] = inlineList(raw)
		} else {
			out[key] = scalar(raw)
		}
	}
	return out
}

// stripComment 去掉行尾 # 注释（引号内的 # 保留）
func stripComment(s string) string {
	inQ := rune(0)
	for i, c := range s {
		switch {
		case inQ != 0 && c == inQ:
			inQ = 0
		case inQ == 0 && (c == '"' || c == '\''):
			inQ = c
		case inQ == 0 && c == '#' && (i == 0 || s[i-1] == ' ' || s[i-1] == '\t'):
			return strings.TrimSpace(s[:i])
		}
	}
	return s
}

func inlineList(raw string) []any {
	raw = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(raw, "["), "]"))
	if raw == "" {
		return nil
	}
	var out []any
	var cur strings.Builder
	inQ := rune(0)
	for _, c := range raw {
		switch {
		case inQ != 0 && c == inQ:
			inQ = 0
			cur.WriteRune(c)
		case inQ == 0 && (c == '"' || c == '\''):
			inQ = c
			cur.WriteRune(c)
		case inQ == 0 && c == ',':
			if s := strings.TrimSpace(cur.String()); s != "" {
				out = append(out, scalar(s))
			}
			cur.Reset()
		default:
			cur.WriteRune(c)
		}
	}
	if s := strings.TrimSpace(cur.String()); s != "" {
		out = append(out, scalar(s))
	}
	return out
}

func scalar(raw string) any {
	raw = strings.TrimSpace(raw)
	if len(raw) >= 2 && (raw[0] == '"' && raw[len(raw)-1] == '"') {
		if s, err := strconv.Unquote(raw); err == nil {
			return s
		}
		return raw[1 : len(raw)-1]
	}
	if len(raw) >= 2 && raw[0] == '\'' && raw[len(raw)-1] == '\'' {
		return strings.ReplaceAll(raw[1:len(raw)-1], "''", "'")
	}
	switch strings.ToLower(raw) {
	case "true", "yes", "on":
		return true
	case "false", "no", "off":
		return false
	case "null", "~":
		return nil
	}
	return raw
}

// fmString 取字符串字段（多个候选键按顺序取第一个非空）
func fmString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		switch v := m[k].(type) {
		case string:
			if s := strings.TrimSpace(v); s != "" {
				return s
			}
		case []any:
			for _, x := range v {
				if s, ok := x.(string); ok && strings.TrimSpace(s) != "" {
					return strings.TrimSpace(s)
				}
			}
		}
	}
	return ""
}

// fmStrings 取列表字段；字符串按逗号 / 空格分隔（Hexo 允许 `tags: a, b`）
func fmStrings(m map[string]any, keys ...string) []string {
	var out []string
	seen := map[string]bool{}
	push := func(s string) {
		s = strings.TrimSpace(s)
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	for _, k := range keys {
		switch v := m[k].(type) {
		case string:
			for _, p := range strings.FieldsFunc(v, func(r rune) bool { return r == ',' || r == '，' }) {
				push(p)
			}
		case []any:
			for _, x := range v {
				switch y := x.(type) {
				case string:
					push(y)
				case []any: // Hexo 分类的层级写法 [[a, b]]
					for _, z := range y {
						if s, ok := z.(string); ok {
							push(s)
						}
					}
				}
			}
		}
	}
	return out
}

func fmBool(m map[string]any, key string) (bool, bool) {
	v, ok := m[key].(bool)
	return v, ok
}
