/** RESTful API 客户端（/api/v1，开发态经 Vite 代理到后端） */

export interface Post {
  id: number;
  slug: string;
  title: string;
  excerpt: string;
  covers: string[];
  tags: string[];
  pinned: boolean;
  createdAt: string;
  updatedAt: string;
  contentMd?: string;
  /** 协作作者署名；站长本人的文章为空（用站点身份） */
  author?: { id: number; name: string; avatar: string };
}

export interface HeroFeed {
  intervalMs: number;
  items: Post[];
}

export interface PostList {
  items: Post[];
  page: number;
  pageSize: number;
  total: number;
}

export interface Tag {
  name: string;
  count: number;
}

export interface Note {
  id: number;
  contentMd: string;
  mood: string;
  images: string[];
  pinned: boolean;
  /** 已隐藏（仅后台列表里出现） */
  hidden?: boolean;
  createdAt: string;
}

export type EngageTarget = 'note' | 'post';
/** 可回应的对象：随想 / 文章，以及已公开的评论（留言墙的喜欢） */
export type ReactionTarget = EngageTarget | 'comment';
/** 回应种类：喜欢 / 灵感 / 会心 / 共鸣（线性图标，不用 emoji） */
export type ReactionKind = 'like' | 'spark' | 'smile' | 'resonate';
export const REACTION_KINDS: ReactionKind[] = ['like', 'spark', 'smile', 'resonate'];

export interface EngageSummary {
  reactions: Partial<Record<ReactionKind, number>>;
  mine: ReactionKind[];
  comments: number;
}

export interface NoteDraft {
  contentMd: string;
  mood: string;
  images: string[];
  pinned: boolean;
}

export interface NoteList {
  items: Note[];
  page: number;
  pageSize: number;
  total: number;
}

export interface AdminPost extends Post {
  status: 'published' | 'draft';
  /** 已隐藏：前台不可见 */
  hidden?: boolean;
}

/** 批量操作：隐藏 / 取消隐藏 / 删除 */
export type BatchAction = 'hide' | 'show' | 'delete';

const BASE = '/api/v1';

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`);
  if (!res.ok) throw new Error(`api_error_${res.status}`);
  return res.json() as Promise<T>;
}

import { clearSessionHint } from '../stores/auth';

/**
 * 管理员请求：凭 HttpOnly 会话 Cookie 认证（同源 fetch 自动携带，脚本读不到令牌）；
 * 所有请求带 X-Requested-With，服务端据此放行写操作（CSRF 防护）。
 */
const ADMIN_HEADERS = { 'X-Requested-With': 'myself' };

/** 会话失效：清除登录提示并踢回登录页 */
function kickToLogin(): void {
  clearSessionHint();
  if (!window.location.pathname.startsWith('/admin/login')) {
    window.location.assign('/admin/login');
  }
}

/** 仍在使用历史默认口令：只允许改密，送去改密页 */
function kickToChange(): void {
  if (!window.location.pathname.startsWith('/setup')) window.location.assign('/setup?change=1');
}

/** 带管理员 Token 的请求 */
async function authed<T>(path: string, init: RequestInit = {}): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    credentials: 'same-origin',
    headers: {
      'Content-Type': 'application/json',
      ...ADMIN_HEADERS,
      ...init.headers,
    },
  });
  if (res.status === 401) {
    kickToLogin();
    throw new Error('unauthorized');
  }
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    const code = (body as { error?: string }).error ?? `api_error_${res.status}`;
    if (code === 'must_change_password') kickToChange();
    throw new Error(code);
  }
  return res.json() as Promise<T>;
}

/** 未登录的 POST（登录 / 初始化）：错误码原样抛出（bad_credentials / too_many_attempts / bad_setup_code …） */
async function publicPost<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: 'POST',
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/json', ...ADMIN_HEADERS },
    body: JSON.stringify(body),
  });
  const data = (await res.json().catch(() => ({}))) as T & { error?: string };
  if (!res.ok) throw new Error(data.error ?? `api_error_${res.status}`);
  return data;
}

export interface SetupPayload {
  code: string;
  username: string;
  password: string;
  site: { title: string; url?: string };
  identity: { name: string; tagline?: string; motto?: string };
  demo: boolean;
}

export const api = {
  posts: (params: { page?: number; pageSize?: number; tag?: string; q?: string } = {}) => {
    const query = new URLSearchParams();
    if (params.page) query.set('page', String(params.page));
    if (params.pageSize) query.set('pageSize', String(params.pageSize));
    if (params.tag) query.set('tag', params.tag);
    if (params.q) query.set('q', params.q);
    const qs = query.toString();
    return get<PostList>(`/posts${qs ? `?${qs}` : ''}`);
  },
  post: (slug: string) => get<Post>(`/posts/${encodeURIComponent(slug)}`),
  tags: () => get<Tag[]>('/tags'),
  hero: () => get<HeroFeed>('/hero'),
  siteConfig: <T>() => get<T>('/site-config'),
  /** 单条随想 + 相邻（older = 更早一条，newer = 更新一条；0 表示没有） */
  note: (id: number) => get<{ note: Note; older: number; newer: number }>(`/notes/${id}`),
  /** 批量互动摘要：回应计数、当前访客已点、已公开评论数 */
  engage: (target: EngageTarget, ids: number[]) =>
    fetch(`${BASE}/engage?target=${target}&ids=${ids.join(',')}`, { credentials: 'same-origin' })
      .then((r) => (r.ok ? r.json() : {})) as Promise<Record<string, EngageSummary>>,
  /** 切换一个回应（已点则取消） */
  react: async (target: ReactionTarget, id: number, kind: ReactionKind): Promise<EngageSummary> => {
    const res = await fetch(`${BASE}/reactions`, {
      method: 'POST',
      credentials: 'same-origin',
      headers: { 'Content-Type': 'application/json', 'X-Requested-With': 'myself' },
      body: JSON.stringify({ target, id, kind }),
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error((data as { error?: string }).error ?? `api_error_${res.status}`);
    return data as EngageSummary;
  },
  githubStatus: <T>() => get<T>('/github-status'),
  /** 压缩包目录（公开）；非 zip 415、损坏 422 */
  archive: (name: string) => get<ArchiveListing>(`/archive/${encodeURIComponent(name)}`),
  notes: (params: {
    page?: number; pageSize?: number; q?: string; media?: boolean;
    from?: string; to?: string;
    /** 后台：连同隐藏的一起列出（需站长会话） */
    all?: boolean;
  } = {}) => {
    const query = new URLSearchParams();
    if (params.all) query.set('all', '1');
    if (params.page) query.set('page', String(params.page));
    if (params.pageSize) query.set('pageSize', String(params.pageSize));
    if (params.q) query.set('q', params.q);
    if (params.media) query.set('media', '1');
    if (params.from) query.set('from', params.from);
    if (params.to) query.set('to', params.to);
    const qs = query.toString();
    return get<NoteList>(`/notes${qs ? `?${qs}` : ''}`);
  },
};

export interface PostDraft {
  slug: string;
  title: string;
  excerpt: string;
  contentMd: string;
  covers: string[];
  tags: string[];
  status: 'published' | 'draft';
  pinned: boolean;
}

export const adminApi = {
  /** 登录：成功后服务端写入会话 Cookie；mustChange = 仍在用历史默认口令，需先改密 */
  login: (username: string, password: string) =>
    publicPost<{ ok: boolean; mustChange?: boolean; user: SessionUser }>('/auth/login', { username, password }),
  setupStatus: () => get<{ needsSetup: boolean }>('/setup'),
  setup: (payload: SetupPayload) => publicPost<{ ok: boolean; user: SessionUser }>('/setup', payload),
  verifySetupCode: (code: string) => publicPost<{ ok: boolean }>('/setup/verify', { code }),
  /** 改密：其它会话全部失效，当前会话由服务端换发新 Cookie */
  changePassword: (oldPassword: string, newPassword: string) =>
    authed<{ ok: boolean }>('/auth/password', {
      method: 'PUT',
      body: JSON.stringify({ oldPassword, newPassword }),
    }),
  posts: () => authed<AdminPost[]>('/admin/posts'),
  post: (id: number) => authed<AdminPost>(`/admin/posts/${id}`),
  createPost: (draft: PostDraft) =>
    authed<{ id: number }>('/admin/posts', { method: 'POST', body: JSON.stringify(draft) }),
  updatePost: (id: number, draft: PostDraft) =>
    authed<{ ok: boolean }>(`/admin/posts/${id}`, { method: 'PUT', body: JSON.stringify(draft) }),
  batchPosts: (ids: number[], action: BatchAction) =>
    authed<{ affected: number }>('/admin/posts/batch', { method: 'POST', body: JSON.stringify({ ids, action }) }),
  batchNotes: (ids: number[], action: BatchAction) =>
    authed<{ affected: number }>('/admin/notes/batch', { method: 'POST', body: JSON.stringify({ ids, action }) }),
  deletePost: (id: number) =>
    authed<{ ok: boolean }>(`/admin/posts/${id}`, { method: 'DELETE' }),
  createNote: (note: NoteDraft) =>
    authed<{ id: number }>('/admin/notes', { method: 'POST', body: JSON.stringify(note) }),
  updateNote: (id: number, note: NoteDraft) =>
    authed<{ ok: boolean }>(`/admin/notes/${id}`, { method: 'PUT', body: JSON.stringify(note) }),
  deleteNote: (id: number) =>
    authed<{ ok: boolean }>(`/admin/notes/${id}`, { method: 'DELETE' }),
  apiKeys: () => authed<ApiKeyInfo[]>('/admin/apikeys'),
  createApiKey: (name: string, scope: ApiKeyScope = 'contrib') =>
    authed<ApiKeyInfo & { key: string }>('/admin/apikeys', {
      method: 'POST',
      body: JSON.stringify({ name, scope }),
    }),
  deleteApiKey: (id: number) =>
    authed<{ ok: boolean }>(`/admin/apikeys/${id}`, { method: 'DELETE' }),
  /** 外部通道调用日志（分页；key 为 Key id，status 按成功 / 失败筛） */
  apiLogs: (p: { page: number; pageSize: number; key?: number; status?: 'ok' | 'error' }) =>
    authed<ApiLogPage>(`/admin/api-logs?${pageQuery(p)}`),
  media: () => authed<MediaItem[]>('/admin/media'),
  /**
   * 上传素材（带查重）：浏览器先算每个文件的 SHA-256 问服务端，已存在的直接复用（duplicate = true）、不再上传；
   * 其余正常上传（服务端写入时还会再查一次）。返回顺序与传入一致。
   */
  /** 上传素材（先按哈希查重）；onProgress 给出 0–1 的上传进度（大文件用） */
  uploadMedia: async (files: File[], onProgress?: (p: number) => void): Promise<MediaItem[]> => {
    const hashes = await Promise.all(files.map(sha256Hex));
    let known: Record<string, MediaItem> = {};
    const asked = hashes.filter((h): h is string => !!h);
    if (asked.length) {
      known = await authed<Record<string, MediaItem>>('/admin/media/lookup', { method: 'POST', body: JSON.stringify({ hashes: asked }) })
        .catch(() => ({}));
    }
    const fresh = files.filter((_, i) => !(hashes[i] && known[hashes[i]!]));
    let uploaded: MediaItem[] = [];
    if (fresh.length) {
      const form = new FormData();
      for (const f of fresh) form.append('files', f);
      // XHR 才有上传进度；fetch 的请求体进度浏览器尚未普遍支持
      const { status, body } = await new Promise<{ status: number; body: string }>((resolve, reject) => {
        const xhr = new XMLHttpRequest();
        xhr.open('POST', `${BASE}/admin/media`);
        xhr.withCredentials = true;
        for (const [k, v] of Object.entries(ADMIN_HEADERS)) xhr.setRequestHeader(k, v);
        xhr.upload.onprogress = (e) => { if (e.lengthComputable) onProgress?.(e.loaded / e.total); };
        xhr.onload = () => resolve({ status: xhr.status, body: xhr.responseText });
        xhr.onerror = () => reject(new Error('network'));
        xhr.send(form);
      });
      if (status === 401) {
        kickToLogin();
        throw new Error('unauthorized');
      }
      if (status < 200 || status >= 300) {
        let code = '';
        try { code = (JSON.parse(body) as { error?: string }).error ?? ''; } catch { /* 非 JSON */ }
        throw new Error(code || `api_error_${status}`);
      }
      uploaded = JSON.parse(body) as MediaItem[];
    }
    const out: MediaItem[] = [];
    for (let i = 0; i < files.length; i += 1) {
      const hit = hashes[i] ? known[hashes[i]!] : undefined;
      const next = hit ?? uploaded.shift();
      if (next) out.push(next);
    }
    return out;
  },
  /** 原图二进制（重裁 UI 用，取 blob） */
  mediaOriginal: async (name: string): Promise<Blob> => {
    const res = await fetch(`${BASE}/admin/media/${encodeURIComponent(name)}/original`, {
      credentials: 'same-origin',
      headers: ADMIN_HEADERS,
    });
    if (res.status === 401) {
      kickToLogin();
      throw new Error('unauthorized');
    }
    if (!res.ok) throw new Error(`api_error_${res.status}`);
    return res.blob();
  },
  cropMedia: (name: string, rect: { left: number; top: number; width: number; height: number }) =>
    authed<MediaItem>(`/admin/media/${encodeURIComponent(name)}/crop`, {
      method: 'POST',
      body: JSON.stringify(rect),
    }),
  deleteMedia: (name: string) =>
    authed<{ ok: boolean }>(`/admin/media/${encodeURIComponent(name)}`, { method: 'DELETE' }),
  /** 重命名：只改显示名，文件与引用地址不变；空串恢复为文件名 */
  renameMedia: (name: string, title: string) =>
    authed<MediaItem>(`/admin/media/${encodeURIComponent(name)}`, { method: 'PUT', body: JSON.stringify({ title }) }),
  /** 回退压缩：恢复压缩前的文件（PNG 转 WebP 的改回原名，站内引用同步） */
  revertMedia: (names: string[]) =>
    authed<{ items: MediaItem[]; failed: number }>('/admin/media/revert', { method: 'POST', body: JSON.stringify({ names }) }),
  deleteMediaBatch: (names: string[]) =>
    authed<{ deleted: number }>('/admin/media/delete', { method: 'POST', body: JSON.stringify({ names }) }),
  /** 打包下载：返回 zip（按显示名命名） */
  zipMedia: async (names: string[]): Promise<Blob> => {
    const res = await fetch(`${BASE}/admin/media/zip`, {
      method: 'POST',
      credentials: 'same-origin',
      headers: { ...ADMIN_HEADERS, 'Content-Type': 'application/json' },
      body: JSON.stringify({ names }),
    });
    if (!res.ok) throw new Error(`zip ${res.status}`);
    return res.blob();
  },
  backups: () => authed<BackupInfo[]>('/admin/backups'),
  createBackup: () => authed<{ name: string }>('/admin/backups', { method: 'POST' }),
  deleteBackup: (name: string) =>
    authed<{ ok: boolean }>(`/admin/backups/${encodeURIComponent(name)}`, { method: 'DELETE' }),
  downloadBackup: async (name: string): Promise<Blob> => {
    const res = await fetch(`${BASE}/admin/backups/${encodeURIComponent(name)}`, {
      credentials: 'same-origin',
      headers: ADMIN_HEADERS,
    });
    if (!res.ok) throw new Error(`api_error_${res.status}`);
    return res.blob();
  },
  qualityScan: () => authed<QualityItem[]>('/admin/quality/scan'),
  /** 发起后台压缩任务：立即返回任务 id，用 qualityJob 轮询进度 */
  qualityCompress: (names: string[], quality: number) =>
    authed<{ id: string; total: number }>('/admin/quality/compress', {
      method: 'POST',
      body: JSON.stringify({ names, quality }),
    }),
  qualityJob: (id: string) => authed<CompressJob>(`/admin/quality/jobs/${id}`),
  githubSync: () =>
    authed<{ ok: boolean; data?: unknown }>('/admin/github/sync', { method: 'POST' }),
  githubLog: () =>
    authed<{
      log: { at: string; ok: boolean; message: string }[];
      preview: unknown;
      cachedAt: string | null;
    }>('/admin/github/log'),
  settings: () => authed<Record<string, unknown>>('/admin/settings'),
  saveSettings: (patch: Record<string, unknown>) =>
    authed<Record<string, unknown>>('/admin/settings', {
      method: 'PUT',
      body: JSON.stringify(patch),
    }),

  /* ----- 用户 / 邀请 / 评论审核 / 发信 ----- */
  users: () => authed<{ items: AdminUser[]; stats: UserStats }>('/admin/users'),
  /** 用户分页：role 分段 + 搜索（名字 / 账号 / 邮箱）在服务端完成 */
  usersPage: (p: { page: number; pageSize: number; role?: UserFilter; q?: string }) =>
    authed<UserPage>(`/admin/users?${pageQuery(p)}`),
  updateUser: (id: number, patch: { role?: UserRole; status?: 'active' | 'disabled'; name?: string }) =>
    authed<{ ok: boolean }>(`/admin/users/${id}`, { method: 'PUT', body: JSON.stringify(patch) }),
  deleteUser: (id: number) => authed<{ ok: boolean }>(`/admin/users/${id}`, { method: 'DELETE' }),
  resetLink: (id: number, send = false) =>
    authed<{ url: string; sent: boolean }>(`/admin/users/${id}/reset`, { method: 'POST', body: JSON.stringify({ send }) }),
  exportUsers: async (): Promise<Blob> => {
    const res = await fetch(`${BASE}/admin/users/export`, { credentials: 'same-origin', headers: ADMIN_HEADERS });
    if (!res.ok) throw new Error(`api_error_${res.status}`);
    return res.blob();
  },
  invites: () => authed<InviteInfo[]>('/admin/invites'),
  createInvite: (body: { role: 'author' | 'reader'; email?: string; days?: number; note?: string; send?: boolean }) =>
    authed<{ url: string; sent: boolean; days: number }>('/admin/invites', { method: 'POST', body: JSON.stringify(body) }),
  deleteInvite: (id: number) => authed<{ ok: boolean }>(`/admin/invites/${id}`, { method: 'DELETE' }),
  /** 审核用户待审头像 */
  reviewAvatar: (id: number, action: 'approve' | 'reject') =>
    authed<{ ok: boolean }>(`/admin/users/${id}/avatar`, { method: 'PUT', body: JSON.stringify({ action }) }),
  comments: (status: CommentStatus) => authed<AdminComment[]>(`/admin/comments?status=${status}`),
  /** 评论分页：counts 为三栏各自总数 */
  commentsPage: (p: { status: CommentStatus; page: number; pageSize: number }) =>
    authed<Paged<AdminComment> & { counts: Record<CommentStatus, number> }>(`/admin/comments?${pageQuery(p)}`),
  setCommentStatus: (id: number, status: CommentStatus) =>
    authed<{ ok: boolean }>(`/admin/comments/${id}`, { method: 'PUT', body: JSON.stringify({ status }) }),
  deleteComment: (id: number) => authed<{ ok: boolean }>(`/admin/comments/${id}`, { method: 'DELETE' }),
  batchComments: (ids: number[], action: { status?: CommentStatus; delete?: boolean }) =>
    authed<{ ok: boolean }>('/admin/comments/batch', { method: 'POST', body: JSON.stringify({ ids, ...action }) }),
  mailTest: (to: string) => authed<{ ok: boolean; error?: string }>('/admin/mail/test', { method: 'POST', body: JSON.stringify({ to }) }),
};

/* ===== 账号（前台读者 / 作者 / 管理员共用）与评论 ===== */

export type UserRole = 'admin' | 'author' | 'reader';

export interface SessionUser {
  id: number;
  login: string;
  email: string;
  name: string;
  avatar: string;
  role: UserRole;
  hasPassword: boolean;
  github: boolean;
  emailVerified: boolean;
  mustChange: boolean;
  /** 站长正在使用「身份」里的头像（没有单独设置） */
  avatarDefault?: boolean;
  /** 待站长审核的新头像 */
  avatarPending?: string;
  /** 登录名冷却中：下次可修改的时间（UTC，空 = 现在就能改） */
  loginNextChange?: string;
  /** 待确认的新邮箱（已发确认邮件，点链接后才换绑） */
  emailPending?: string;
}

export interface AdminUser {
  id: number;
  login: string;
  email: string;
  name: string;
  avatar: string;
  role: UserRole;
  /** 待审核的新头像 */
  avatarPending?: string;
  status: 'active' | 'pending' | 'disabled';
  github: boolean;
  hasPassword: boolean;
  emailVerified: boolean;
  createdAt: string;
  lastActiveAt: string;
  comments: number;
  self: boolean;
}

export interface UserStats {
  total: number;
  newMonth: number;
  activeMonth: number;
  authors: number;
  comments: number;
  commentsMonth: number;
  pending: number;
  /** 待审头像数 */
  avatars?: number;
}

export interface InviteInfo {
  id: number;
  role: 'author' | 'reader';
  email: string;
  note: string;
  expiresAt: string;
  usedAt: string;
  usedBy: string;
  createdAt: string;
}

export type CommentStatus = 'pending' | 'approved' | 'spam';
export type CommentTarget = 'post' | 'note' | 'guestbook';

export interface CommentAuthor {
  id?: number;
  name: string;
  avatar: string;
  role: UserRole | 'guest';
}

export interface CommentItem {
  id: number;
  parentId?: number;
  body: string;
  createdAt: string;
  author: CommentAuthor;
  pending?: boolean;
  /** 喜欢数与当前访客是否已点（访客回应关闭时缺省） */
  likes?: number;
  liked?: boolean;
}

export interface AdminComment extends CommentItem {
  status: CommentStatus;
  target: CommentTarget;
  targetTitle: string;
  targetLink: string;
  /** 发表评论用的 key（文章 slug / 随想 id / 留言墙为空），后台回复时带回 */
  targetKey: string;
  hasLink: boolean;
  ipHash: string;
}

/** 前台账号接口：同源 Cookie 会话；写操作带 CSRF 头 */
async function accountCall<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method,
    credentials: 'same-origin',
    headers: body === undefined ? ADMIN_HEADERS : { 'Content-Type': 'application/json', ...ADMIN_HEADERS },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const data = (await res.json().catch(() => ({}))) as T & { error?: string };
  if (!res.ok) throw new Error(data.error ?? `api_error_${res.status}`);
  return data;
}

export const accountApi = {
  session: () => accountCall<{ loggedIn: boolean; mustChange?: boolean; user?: SessionUser }>('GET', '/auth/session'),
  login: (username: string, password: string) =>
    accountCall<{ ok: boolean; mustChange?: boolean; user: SessionUser }>('POST', '/auth/login', { username, password }),
  register: (body: { email: string; name: string; password: string; invite?: string; website?: string }) =>
    accountCall<{ ok: boolean; pending?: boolean; user?: SessionUser }>('POST', '/auth/register', body),
  /** 邮箱验证链接：注册验证与换绑邮箱共用；emailChanged = 这是一次换绑 */
  verify: (token: string) => accountCall<{ ok: boolean; emailChanged?: boolean; user: SessionUser }>('POST', '/auth/verify', { token }),
  forgot: (email: string) => accountCall<{ ok: boolean }>('POST', '/auth/forgot', { email }),
  resetInfo: (token: string) => accountCall<{ email: string }>('GET', `/auth/reset?token=${encodeURIComponent(token)}`),
  reset: (token: string, password: string) =>
    accountCall<{ ok: boolean; user: SessionUser }>('POST', '/auth/reset', { token, password }),
  invite: (code: string) => accountCall<{ role: 'author' | 'reader'; email: string }>('GET', `/auth/invite?code=${encodeURIComponent(code)}`),
  updateMe: (name: string) => accountCall<{ ok: boolean; user: SessionUser }>('PUT', '/me', { name }),
  /** 上传已裁切的头像：站长直接生效，其他人进入待审 */
  uploadAvatar: async (blob: Blob): Promise<{ ok: boolean; user: SessionUser }> => {
    const form = new FormData();
    form.append('file', blob, blob.type === 'image/webp' ? 'avatar.webp' : 'avatar.png');
    const res = await fetch(`${BASE}/me/avatar`, { method: 'POST', credentials: 'same-origin', headers: ADMIN_HEADERS, body: form });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error((data as { error?: string }).error ?? `api_error_${res.status}`);
    return data as { ok: boolean; user: SessionUser };
  },
  /** 移除头像（连同待审的）；站长恢复为身份头像 */
  deleteAvatar: (pendingOnly = false) =>
    accountCall<{ ok: boolean; user: SessionUser }>('DELETE', pendingOnly ? '/me/avatar?pending=1' : '/me/avatar'),
  changeLogin: (login: string) => accountCall<{ ok: boolean; user: SessionUser }>('PUT', '/me/login', { login }),
  /** 添加 / 更换邮箱：向新邮箱发确认邮件，点链接后才换绑 */
  changeEmail: (email: string, password: string) =>
    accountCall<{ ok: boolean; user: SessionUser }>('PUT', '/me/email', { email, password }),
  cancelEmailChange: () => accountCall<{ ok: boolean; user: SessionUser }>('DELETE', '/me/email/pending'),
  changePassword: (oldPassword: string, newPassword: string) =>
    accountCall<{ ok: boolean }>('PUT', '/auth/password', { oldPassword, newPassword }),
  logout: () => accountCall<{ ok: boolean }>('POST', '/auth/logout'),
  /** GitHub 登录 / 绑定入口（整页跳转） */
  githubUrl: (opts: { mode?: 'login' | 'link'; invite?: string; next?: string } = {}) => {
    const q = new URLSearchParams();
    if (opts.mode) q.set('mode', opts.mode);
    if (opts.invite) q.set('invite', opts.invite);
    if (opts.next) q.set('next', opts.next);
    return `${BASE}/auth/github/start?${q.toString()}`;
  },
  comments: (target: CommentTarget, key: string) =>
    accountCall<{ enabled: boolean; items: CommentItem[]; reactions?: boolean }>('GET', `/comments?target=${target}&key=${encodeURIComponent(key)}`),
  postComment: (body: { target: CommentTarget; key: string; body: string; parentId?: number; guestName?: string; website?: string }) =>
    accountCall<{ ok: boolean; id: number; pending: boolean }>('POST', '/comments', body),
};

/** 文件内容的 SHA-256（十六进制）；非安全上下文（http 且非 localhost）没有 crypto.subtle 时返回 null */
async function sha256Hex(file: Blob): Promise<string | null> {
  // 大文件（视频等）不在浏览器里整读算哈希，交给服务端查重
  if (!globalThis.crypto?.subtle || file.size > 64 << 20) return null;
  try {
    const digest = await crypto.subtle.digest('SHA-256', await file.arrayBuffer());
    return [...new Uint8Array(digest)].map((b) => b.toString(16).padStart(2, '0')).join('');
  } catch {
    return null;
  }
}

/** 素材类型：图片 / 视频 / 音频 / 压缩包 / 其他文件 */
export type MediaKindName = 'image' | 'video' | 'audio' | 'archive' | 'file';

/** 压缩包目录（在线预览）：只读中央目录，最多 5000 条 */
export interface ArchiveListing {
  name: string;
  title: string;
  size: number;
  entries: { name: string; size: number; compressed: number; dir: boolean; modified: string }[];
  total: number;
  truncated: boolean;
}

export interface MediaItem {
  /** 类型（旧后端缺省时由扩展名推断，见 utils/mediaKind.ts） */
  kind?: MediaKindName;
  ext?: string;
  mime?: string;
  name: string;
  /** 显示名：上传时的原文件名或重命名后的名字；为空时显示 name */
  title?: string;
  url: string;
  /** 最长边 480px 的 webp 缩略图，按需生成 */
  thumb: string;
  size: number;
  hasOriginal: boolean;
  crop: { left: number; top: number; width: number; height: number } | null;
  createdAt: string;
  /** 与已有素材内容相同：没有新存，返回的是已有的那张 */
  duplicate?: boolean;
  /** 压缩记录（可回退）；未压缩为 null */
  compressed?: MediaCompression | null;
}

export interface BackupInfo {
  name: string;
  size: number;
  createdAt: string;
}

/** 压缩记录：压缩前的文件名（PNG 转 WebP 时不同）与体积 */
export interface MediaCompression {
  from: string;
  before: number;
  at: string;
}

export interface QualityItem {
  name: string;
  title?: string;
  compressed?: MediaCompression | null;
  url: string;
  size: number;
  format: string;
  width: number;
  height: number;
  hasAlpha: boolean;
  compressible: boolean;
}

export interface CompressJob {
  id: string;
  total: number;
  done: number;
  running: boolean;
  current: string;
  results: CompressResult[];
  startedAt: string;
  endedAt?: string;
}

export interface CompressResult {
  name: string;
  newName?: string;
  before?: number;
  after?: number;
  error?: string;
}

/** full = 全托管；contrib = 仅投稿（只能写草稿、只能碰自己创建的草稿） */
export type ApiKeyScope = 'full' | 'contrib';

export interface ApiKeyInfo {
  id: number;
  name: string;
  prefix: string;
  scope: ApiKeyScope;
  lastUsedAt: string | null;
  createdAt: string;
  /** 调用日志保留窗口内的调用数 / 失败数（旧后端没有） */
  calls?: number;
  errors?: number;
}

/** 通用分页响应 */
export interface Paged<T> {
  items: T[];
  page: number;
  pageSize: number;
  total: number;
}

/** 分页 / 筛选参数 → 查询串（跳过空值） */
function pageQuery(p: Record<string, string | number | undefined>): string {
  const q = new URLSearchParams();
  for (const [k, v] of Object.entries(p)) if (v !== undefined && v !== '') q.set(k, String(v));
  return q.toString();
}

/** 一条 API 调用记录；keyId 为空 = Key 缺失或无效 */
export interface ApiLogEntry {
  id: number;
  at: string;
  keyId: number | null;
  keyName: string;
  keyPrefix: string;
  method: string;
  path: string;
  status: number;
  ms: number;
  ip: string;
  ua: string;
}

export interface ApiLogPage extends Paged<ApiLogEntry> {
  counts: { all: number; ok: number; error: number };
}

export type UserFilter = 'all' | UserRole | 'disabled';

export interface UserPage extends Paged<AdminUser> {
  stats: UserStats;
  counts: Record<UserFilter, number>;
  /** 全部待审头像（不受分页影响） */
  pendingAvatars: AdminUser[];
}

/** 站内素材 URL → 缩略图 URL；非站内素材（占位图/外链）原样返回 */
export function thumbOf(url: string): string {
  if (!url.startsWith('/uploads/') || url.startsWith('/uploads/thumbs/')) return url;
  return `/uploads/thumbs/${url.slice('/uploads/'.length)}.webp`;
}
