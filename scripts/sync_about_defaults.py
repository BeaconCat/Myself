"""把 client/src/about/default-modules.json 同步进 server/internal/config/config.go 的 about.modules 默认值。

默认关于页模块的唯一来源是 default-modules.json；修改后在仓库根目录运行：
    python scripts/sync_about_defaults.py
再跑 `go -C server test ./...` 确认 config 包测试通过。
"""
import json
import os
import sys

root = sys.argv[1] if len(sys.argv) > 1 else os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
mods = json.load(open(f'{root}/client/src/about/default-modules.json', encoding='utf-8'))
p = f'{root}/server/internal/config/config.go'
s = open(p, encoding='utf-8', newline='').read()
nl = '\r\n' if '\r\n' in s else '\n'
s = s.replace('\r\n', '\n')
start = s.index('    "modules": [')
end = s.index('\n  },\n  "backup"')
body = json.dumps(mods, ensure_ascii=False, indent=2)
body = body.replace('`', '\\u0060')  # Go 原始字符串不能含反引号，用 JSON 转义
body = '\n'.join(('    ' + l if i > 0 else l) for i, l in enumerate(body.split('\n')))
s = s[:start] + '    "modules": ' + body + s[end:]
assert '`X-Api' not in s
open(p, 'w', encoding='utf-8', newline='').write(s.replace('\n', nl))
print('ok', nl == '\r\n')
