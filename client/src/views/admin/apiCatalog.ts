/** 平台全部 REST 接口目录：API 中心文档 + 调试台数据源 */

export type AuthKind = 'none' | 'jwt' | 'apikey';

export interface Endpoint {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE';
  path: string;
  desc: string;
  auth: AuthKind;
  /** 调试台默认请求体 */
  sampleBody?: string;
  /** 路径参数示例替换，如 :id → 1 */
  sample?: Record<string, string>;
}

export interface ApiGroup {
  title: string;
  endpoints: Endpoint[];
}

export const API_CATALOG: ApiGroup[] = [
  {
    title: '公开接口（无需认证）',
    endpoints: [
      { method: 'GET', path: '/api/v1/posts?page=1&pageSize=10&tag=&q=', desc: '已发布文章列表：分页、标签过滤、关键词搜索', auth: 'none' },
      { method: 'GET', path: '/api/v1/posts/:slug', desc: '文章详情（含 Markdown 正文）', auth: 'none', sample: { ':slug': 'welcome-to-myself' } },
      { method: 'GET', path: '/api/v1/tags', desc: '标签及文章计数', auth: 'none' },
      { method: 'GET', path: '/feed', desc: 'RSS 2.0 订阅源（最新 20 篇）', auth: 'none' },
      { method: 'GET', path: '/api/v1/notes?page=1&pageSize=20', desc: '随想信息流', auth: 'none' },
      { method: 'GET', path: '/api/v1/site-config', desc: '站点公开配置（标题/主题/关于等）', auth: 'none' },
      { method: 'GET', path: '/api/v1/img/:from/:to/:label', desc: '本地渐变占位图（hex 颜色对 + 文本）', auth: 'none', sample: { ':from': 'ff0032', ':to': '7a0020', ':label': 'Demo' } },
    ],
  },
  {
    title: '外部通道（X-Api-Key，供 AI / 脚本托管）',
    endpoints: [
      { method: 'GET', path: '/api/v1/ext/posts?status=all', desc: '文章列表（含草稿与定时内容）', auth: 'apikey' },
      { method: 'GET', path: '/api/v1/ext/posts/:id', desc: '文章详情（含正文）', auth: 'apikey', sample: { ':id': '1' } },
      {
        method: 'POST', path: '/api/v1/ext/posts', desc: '创建文章；全托管 Key 可用 scheduled + publishAt 定时发布', auth: 'apikey',
        sampleBody: '{\n  "slug": "hello-from-agent",\n  "title": "来自 Agent 的文章",\n  "excerpt": "一句话摘要",\n  "contentMd": "# 正文\\n\\n用 Markdown 书写。",\n  "tags": ["AI"],\n  "covers": [],\n  "status": "draft"\n}',
      },
      {
        method: 'PUT', path: '/api/v1/ext/posts/:id', desc: '完整更新文章；先读取并保留原字段，省略状态与时间可保留排期', auth: 'apikey', sample: { ':id': '1' },
        sampleBody: '{\n  "title": "更新后的标题",\n  "excerpt": "",\n  "contentMd": "# 更新后的正文",\n  "tags": [],\n  "covers": []\n}',
      },
      { method: 'DELETE', path: '/api/v1/ext/posts/:id', desc: '删除文章', auth: 'apikey', sample: { ':id': '999' } },
      { method: 'GET', path: '/api/v1/ext/notes?pageSize=50', desc: '随想列表（含草稿与定时内容）', auth: 'apikey' },
      { method: 'GET', path: '/api/v1/ext/notes/:id', desc: '随想详情与发布状态', auth: 'apikey', sample: { ':id': '1' } },
      {
        method: 'POST', path: '/api/v1/ext/notes', desc: '创建随想；默认立即公开，请显式选择 draft / published / scheduled', auth: 'apikey',
        sampleBody: '{\n  "contentMd": "来自 Agent 的一条随想",\n  "mood": "AI",\n  "images": [],\n  "status": "draft"\n}',
      },
      {
        method: 'PUT', path: '/api/v1/ext/notes/:id', desc: '完整更新随想；带回正文、心情与配图，省略状态与时间可保留排期', auth: 'apikey', sample: { ':id': '1' },
        sampleBody: '{\n  "contentMd": "更新后的随想",\n  "mood": "AI",\n  "images": []\n}',
      },
      { method: 'DELETE', path: '/api/v1/ext/notes/:id', desc: '删除随想', auth: 'apikey', sample: { ':id': '999' } },
    ],
  },
  {
    title: '管理接口（JWT，控制中心专用）',
    endpoints: [
      { method: 'POST', path: '/api/v1/auth/login', desc: '登录换取 JWT', auth: 'none', sampleBody: '{\n  "username": "admin",\n  "password": "……"\n}' },
      { method: 'PUT', path: '/api/v1/auth/password', desc: '修改密码', auth: 'jwt', sampleBody: '{\n  "oldPassword": "……",\n  "newPassword": "……"\n}' },
      { method: 'GET', path: '/api/v1/admin/posts', desc: '全量文章（含草稿）', auth: 'jwt' },
      { method: 'GET', path: '/api/v1/admin/posts/:id', desc: '文章编辑详情', auth: 'jwt', sample: { ':id': '1' } },
      { method: 'POST', path: '/api/v1/admin/posts', desc: '新建文章', auth: 'jwt' },
      { method: 'PUT', path: '/api/v1/admin/posts/:id', desc: '更新文章', auth: 'jwt', sample: { ':id': '1' } },
      { method: 'DELETE', path: '/api/v1/admin/posts/:id', desc: '删除文章', auth: 'jwt', sample: { ':id': '999' } },
      { method: 'POST', path: '/api/v1/admin/notes', desc: '发布随想', auth: 'jwt' },
      { method: 'PUT', path: '/api/v1/admin/notes/:id', desc: '编辑随想', auth: 'jwt', sample: { ':id': '1' } },
      { method: 'DELETE', path: '/api/v1/admin/notes/:id', desc: '删除随想', auth: 'jwt', sample: { ':id': '999' } },
      { method: 'GET', path: '/api/v1/admin/media', desc: '素材列表', auth: 'jwt' },
      { method: 'POST', path: '/api/v1/admin/media', desc: '上传素材（multipart，字段 files）', auth: 'jwt' },
      { method: 'POST', path: '/api/v1/admin/media/:name/crop', desc: '裁切（原图备份 + 裁切框入库）', auth: 'jwt' },
      { method: 'DELETE', path: '/api/v1/admin/media/:name', desc: '删除素材', auth: 'jwt' },
      { method: 'GET', path: '/api/v1/admin/apikeys', desc: 'APIKey 列表', auth: 'jwt' },
      { method: 'POST', path: '/api/v1/admin/apikeys', desc: '创建 APIKey（明文仅返回一次）', auth: 'jwt', sampleBody: '{\n  "name": "my-agent"\n}' },
      { method: 'DELETE', path: '/api/v1/admin/apikeys/:id', desc: '吊销 APIKey', auth: 'jwt', sample: { ':id': '1' } },
      { method: 'GET', path: '/api/v1/admin/api-logs?page=1&pageSize=20&key=&status=', desc: '外部通道调用日志：分页，按 Key id / 成功（ok）失败（error）筛选', auth: 'jwt' },
      { method: 'GET', path: '/api/v1/admin/users?page=1&pageSize=20&role=all&q=', desc: '用户列表：分页、角色分段（admin/author/reader/disabled）、搜索', auth: 'jwt' },
      { method: 'GET', path: '/api/v1/admin/comments?status=pending&page=1&pageSize=20', desc: '评论审核列表：按状态分页', auth: 'jwt' },
      { method: 'GET', path: '/api/v1/admin/settings', desc: '读取全站配置', auth: 'jwt' },
      { method: 'PUT', path: '/api/v1/admin/settings', desc: '更新配置（深合并）', auth: 'jwt', sampleBody: '{\n  "site": { "title": "Myself" }\n}' },
      { method: 'GET', path: '/api/v1/admin/backups', desc: '备份列表', auth: 'jwt' },
      { method: 'POST', path: '/api/v1/admin/backups', desc: '立即备份（db + 素材打包 zip）', auth: 'jwt' },
      { method: 'GET', path: '/api/v1/admin/quality/scan', desc: '扫描可压缩图片', auth: 'jwt' },
      { method: 'POST', path: '/api/v1/admin/quality/compress', desc: '压缩（PNG→WebP 保 alpha / JPG 重压）', auth: 'jwt', sampleBody: '{\n  "names": [],\n  "quality": 80\n}' },
    ],
  },
];

/** 基于所选 APIKey 生成 agent 托管提示词 */
export function buildAgentPrompt(baseUrl: string, apiKey: string): string {
  const key = apiKey || '<你的 APIKey>';
  return `你是博客站点「Myself」的内容托管助手。通过以下 REST 接口管理站点内容。

## 认证
调用本站 /api/v1/ext/* 时携带请求头（不要将 Key 发送到第三方地址）：
X-Api-Key: ${key}
JSON 写请求同时携带 Content-Type: application/json。
基础地址：${baseUrl}

权限分为 contrib（仅创建文章草稿、读写自己创建的草稿）与 full（管理全部文章和随想，可立即/定时发布）。APIKey 不用于 /admin/* 管理接口，也没有素材上传权限；配图与媒体使用已有可访问 URL，需要上传时请站长先在素材库完成。

## 内容规范
- 一切正文使用 Markdown（文章与随想均支持标题、列表、待办、表格、代码块、引用、删除线、链接、图片与脚注）
- 文章需要 slug（1–80 位小写字母、数字、连字符且唯一）、非空 title、contentMd；excerpt 为纯文本摘要，可用换行保留分行结构（JSON 中用 \\n 编码）；tags 字符串数组；covers 最多 3 个图片 URL
- 新文章默认 status="draft"（草稿，待站长审核）；仅在站长明确要求时使用 "published"
- 随想支持与文章相同的完整 Markdown：contentMd + mood（心情标签）+ images（最多 9 个 URL）
- 注意：新随想省略 status 会立即公开。没有明确发布授权时必须显式传 status="draft"；只有明确要求立即发布才传 "published"。

## Markdown 与脚注
- 使用标准命名脚注：正文写 [^source]，文末写 [^source]: 说明或来源链接。
- 标签在同一篇内容内唯一，不包含空格；同一来源重复引用时复用标签，显示编号按首次引用顺序自动生成。
- 每处引用必须有对应定义；脚注定义统一放在文末，不手写编号、HTML 锚点或返回链接。
- 多段脚注的续行、列表和代码块缩进四个空格，段落之间保留空行；修改内容时保留完整定义块与所有引用。
- 引用真实、可核验的来源，不编造脚注、文献或链接。无来源需求时不必添加脚注。
- API 的 contentMd 与 excerpt 都是 JSON 字符串，换行编码为 \\n；不要把换行转成字面量反斜杠与 n，也不要转成 HTML <br>。

~~~markdown
一个观点[^source]，再次引用同一来源[^source]。

[^source]: [来源标题](https://example.com/source)

    补充说明，作为脚注的第二段。
~~~

## 定时发布与草稿
- 文章与随想均支持 status="draft" / "published" / "scheduled"。
- 只有站长明确要求定时发布时，使用 status="scheduled" 并提供未来的 publishAt；推荐带时区的 RFC3339（如 2099-01-02T09:30:00+08:00）。不带时区的 YYYY-MM-DDTHH:mm:ss 按站点时区解释。
- 先读取 GET ${baseUrl}/api/v1/site-config 的 timezone，再结合当前时间解析“明天上午”等表达；站长指定其他时区时按指定时区换算。示例日期不能直接照抄，时间含义不明确时先澄清。
- 服务器返回规范化的 UTC publishAt。定时内容在到期前不会进入前台、搜索、RSS 或互动；服务重启会补发到期内容。
- 修改排期：先读取详情，再在完整内容字段中加入 status="scheduled" 与新的 publishAt。取消排期：完整内容加 status="draft", publishAt=null；立即发布：完整内容加 status="published", publishAt=null。
- 更新正文但不改变发布安排时，省略 status 和 publishAt；不要把定时内容意外改成草稿或立即发布。
- 外部 API 定时发布只对 full Key 开放；contrib Key 的文章会保留为草稿，且不能操作任何随想。务必检查响应中的 status 和 publishAt，不能把草稿说成已排期，也不能把已排期说成已发布。

## 更新前必须保留完整字段
PUT 不是部分字段 PATCH。只传 status / publishAt 会被拒绝；省略可选内容字段可能清空已有信息。
- 文章：先 GET /ext/posts/{id}，带回 title、excerpt、contentMd、tags、covers，再合并需要修改的字段。slug 不可通过此接口修改。
- 随想：先 GET /ext/notes/{id}，带回 contentMd、mood、images，再合并需要修改的字段。随想正文不能为空。
- 不用列表摘要代替详情正文，也不要把旧 status / publishAt 不加判断地回传；修改正文时省略这两个字段可避免覆盖刚刚到期的发布状态。
- 保存后再 GET 详情核对正文、配图、状态与排期。hidden 为 true 的内容即使已发布仍不会公开，外部 API 不能取消隐藏，需要站长在后台处理。

定时创建文章的 JSON 请求体（替换 slug、内容与时间）：
~~~json
{"slug":"scheduled-post","title":"定时文章","excerpt":"摘要","contentMd":"正文","tags":[],"covers":[],"status":"scheduled","publishAt":"2099-01-02T09:30:00+08:00"}
~~~
定时创建随想的 JSON 请求体（仅 full Key）：
~~~json
{"contentMd":"定时随想正文","mood":"记录","images":[],"status":"scheduled","publishAt":"2099-01-02T09:30:00+08:00"}
~~~
取消文章排期时，PUT 请求体应保留 GET 返回的内容，例如：
~~~json
{"title":"原文章标题","excerpt":"原摘要","contentMd":"完整原正文","tags":["原标签"],"covers":[],"status":"draft","publishAt":null}
~~~
上述原内容仅为占位说明，必须用实际读取值替换；已有封面不得改为空数组。

## 接口
文章：
- GET    ${baseUrl}/api/v1/ext/posts?status=all        列表（含草稿/定时；可筛选 draft / published / scheduled）
- GET    ${baseUrl}/api/v1/ext/posts/{id}              详情（含正文）
- POST   ${baseUrl}/api/v1/ext/posts                   创建 {slug,title,excerpt,contentMd,tags,covers,status,publishAt}
- PUT    ${baseUrl}/api/v1/ext/posts/{id}              更新（同上，slug 不可改）
- DELETE ${baseUrl}/api/v1/ext/posts/{id}              删除

随想：
- GET    ${baseUrl}/api/v1/ext/notes?pageSize=50       最新随想列表（含草稿/定时，pageSize 为 1–100，不支持翻页）
- GET    ${baseUrl}/api/v1/ext/notes/{id}             详情（含发布状态）
- POST   ${baseUrl}/api/v1/ext/notes                   创建 {contentMd,mood,images,status,publishAt}
- PUT    ${baseUrl}/api/v1/ext/notes/{id}              更新
- DELETE ${baseUrl}/api/v1/ext/notes/{id}              删除

只读参考（无需认证）：
- GET ${baseUrl}/api/v1/posts / /api/v1/tags / /api/v1/site-config / /feed

## 行为准则
1. 写作前先 GET 现有内容，避免重复主题与 slug 冲突（409 = slug 已存在）
2. 修改/删除前必须先读取确认目标，谨慎执行删除
3. 保持站点现有文风与标签体系，产出高质量中文内容
4. 401 表示 Key 缺失、无效或已吊销：停止操作并报告，不猜测或索要站长密码。
5. 403 scope_forbidden 表示当前 Key 权限不足：不能通过管理接口绕过；文章任务可在授权范围内保留草稿，随想任务应报告需要 full Key，不擅自改成文章。
6. 400 invalid_publish_time / publish_time_must_be_future 表示时间无效或不在未来，校正格式和时区；invalid_post / invalid_note 时检查完整内容字段，invalid_status 时检查状态值。不把失败请求说成成功。
7. 404 可能是目标不存在或当前 Key 无权读取；先核对 ID 与权限，不自动创建替代内容。POST 超时或断网时先查询是否已创建，避免重复投稿。
8. 成功后报告内容 ID、实际状态；排期需同时说明站点时区中的具体日期时间。删除必须符合站长明确意图。`;
}
