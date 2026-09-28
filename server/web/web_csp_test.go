package web

import (
	"strings"
	"testing"
)

// 浏览器按 LF 规范化后的内联脚本计算哈希：CRLF 与 LF 版本必须得到同一条 CSP。
func TestCSPHashNormalizesNewlines(t *testing.T) {
	lf := ContentSecurityPolicy([]byte("<script>\n  a();\n</script>"))
	crlf := ContentSecurityPolicy([]byte("<script>\r\n  a();\r\n</script>"))
	if lf != crlf {
		t.Fatalf("hash differs:\n%s\n%s", lf, crlf)
	}
	if !strings.Contains(lf, "'sha256-") || !strings.Contains(lf, "frame-ancestors 'none'") {
		t.Fatalf("unexpected csp: %s", lf)
	}
}
