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
      { method: 'GET', path: '/api/v1/ext/posts?status=all', desc: '文章列表（含草稿）', auth: 'apikey' },
      { method: 'GET', path: '/api/v1/ext/posts/:id', desc: '文章详情（含正文）', auth: 'apikey', sample: { ':id': '1' } },
      {
        method: 'POST', path: '/api/v1/ext/posts', desc: '创建文章（默认草稿，需人工审核发布；status 可显式 published）', auth: 'apikey',
        sampleBody: '{\n  "slug": "hello-from-agent",\n  "title": "来自 Agent 的文章",\n  "excerpt": "一句话摘要",\n  "contentMd": "# 正文\\n\\n用 Markdown 书写。",\n  "tags": ["AI"],\n  "covers": [],\n  "status": "draft"\n}',
      },
      {
        method: 'PUT', path: '/api/v1/ext/posts/:id', desc: '更新文章', auth: 'apikey', sample: { ':id': '1' },
        sampleBody: '{\n  "title": "更新后的标题",\n  "excerpt": "",\n  "contentMd": "# 更新后的正文",\n  "tags": [],\n  "covers": [],\n  "status": "draft"\n}',
      },
      { method: 'DELETE', path: '/api/v1/ext/posts/:id', desc: '删除文章', auth: 'apikey', sample: { ':id': '999' } },
      { method: 'GET', path: '/api/v1/ext/notes?pageSize=50', desc: '随想列表', auth: 'apikey' },
      {
        method: 'POST', path: '/api/v1/ext/notes', desc: '发布随想（Markdown + 心情 + 最多 9 图）', auth: 'apikey',
        sampleBody: '{\n  "contentMd": "来自 Agent 的一条随想",\n  "mood": "AI",\n  "images": []\n}',
      },
      {
        method: 'PUT', path: '/api/v1/ext/notes/:id', desc: '更新随想', auth: 'apikey', sample: { ':id': '1' },
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
所有请求携带请求头：
X-Api-Key: ${key}
基础地址：${baseUrl}

## 内容规范
- 一切正文使用 Markdown（标题/列表/表格/代码块/引用均支持）
- 文章需要 slug（小写字母数字连字符）、title、contentMd；excerpt 一句话摘要；tags 字符串数组；covers 最多 3 个图片 URL
- 新文章默认 status="draft"（草稿，待站长审核）；仅在站长明确要求时使用 "published"
- 随想是短内容：contentMd + mood（心情标签）+ images（最多 9 个 URL）

## 接口
文章：
- GET    ${baseUrl}/api/v1/ext/posts?status=all        列表（含草稿）
- GET    ${baseUrl}/api/v1/ext/posts/{id}              详情（含正文）
- POST   ${baseUrl}/api/v1/ext/posts                   创建 {slug,title,excerpt,contentMd,tags,covers,status}
- PUT    ${baseUrl}/api/v1/ext/posts/{id}              更新（同上，slug 不可改）
- DELETE ${baseUrl}/api/v1/ext/posts/{id}              删除

随想：
- GET    ${baseUrl}/api/v1/ext/notes?pageSize=50       列表
- POST   ${baseUrl}/api/v1/ext/notes                   发布 {contentMd,mood,images}
- PUT    ${baseUrl}/api/v1/ext/notes/{id}              更新
- DELETE ${baseUrl}/api/v1/ext/notes/{id}              删除

只读参考（无需认证）：
- GET ${baseUrl}/api/v1/posts / /api/v1/tags / /api/v1/site-config / /feed

## 行为准则
1. 写作前先 GET 现有内容，避免重复主题与 slug 冲突（409 = slug 已存在）
2. 修改/删除前必须先读取确认目标，谨慎执行删除
3. 保持站点现有文风与标签体系，产出高质量中文内容
4. 收到 401 表示 Key 已被吊销，停止操作并报告
5. 收到 403 scope_forbidden 表示当前 Key 为「仅投稿」权限：只能创建文章草稿、只能读写自己创建的草稿，不能发布、不能操作随想；
   此时不要重试，改为写成草稿并提醒站长审阅`;
}
