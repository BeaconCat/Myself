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
  createdAt: string;
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
}

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
  githubStatus: <T>() => get<T>('/github-status'),
  notes: (params: {
    page?: number; pageSize?: number; q?: string; media?: boolean;
    from?: string; to?: string;
  } = {}) => {
    const query = new URLSearchParams();
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
    publicPost<{ ok: boolean; mustChange?: boolean }>('/auth/login', { username, password }),
  setupStatus: () => get<{ needsSetup: boolean }>('/setup'),
  setup: (payload: SetupPayload) => publicPost<{ ok: boolean }>('/setup', payload),
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
  media: () => authed<MediaItem[]>('/admin/media'),
  uploadMedia: async (files: File[]): Promise<MediaItem[]> => {
    const form = new FormData();
    for (const f of files) form.append('files', f);
    const res = await fetch(`${BASE}/admin/media`, {
      method: 'POST',
      credentials: 'same-origin',
      headers: ADMIN_HEADERS,
      body: form,
    });
    if (res.status === 401) {
      kickToLogin();
      throw new Error('unauthorized');
    }
    if (!res.ok) throw new Error(`api_error_${res.status}`);
    return res.json() as Promise<MediaItem[]>;
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
  comments: (status: CommentStatus) => authed<AdminComment[]>(`/admin/comments?status=${status}`),
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
}

export interface AdminUser {
  id: number;
  login: string;
  email: string;
  name: string;
  avatar: string;
  role: UserRole;
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
}

export interface AdminComment extends CommentItem {
  status: CommentStatus;
  target: CommentTarget;
  targetTitle: string;
  targetLink: string;
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
  verify: (token: string) => accountCall<{ ok: boolean; user: SessionUser }>('POST', '/auth/verify', { token }),
  forgot: (email: string) => accountCall<{ ok: boolean }>('POST', '/auth/forgot', { email }),
  resetInfo: (token: string) => accountCall<{ email: string }>('GET', `/auth/reset?token=${encodeURIComponent(token)}`),
  reset: (token: string, password: string) =>
    accountCall<{ ok: boolean; user: SessionUser }>('POST', '/auth/reset', { token, password }),
  invite: (code: string) => accountCall<{ role: 'author' | 'reader'; email: string }>('GET', `/auth/invite?code=${encodeURIComponent(code)}`),
  updateMe: (name: string, avatar: string) => accountCall<{ ok: boolean }>('PUT', '/me', { name, avatar }),
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
    accountCall<{ enabled: boolean; items: CommentItem[] }>('GET', `/comments?target=${target}&key=${encodeURIComponent(key)}`),
  postComment: (body: { target: CommentTarget; key: string; body: string; parentId?: number; guestName?: string; website?: string }) =>
    accountCall<{ ok: boolean; id: number; pending: boolean }>('POST', '/comments', body),
};

export interface MediaItem {
  name: string;
  url: string;
  /** 最长边 480px 的 webp 缩略图，按需生成 */
  thumb: string;
  size: number;
  hasOriginal: boolean;
  crop: { left: number; top: number; width: number; height: number } | null;
  createdAt: string;
}

export interface BackupInfo {
  name: string;
  size: number;
  createdAt: string;
}

export interface QualityItem {
  name: string;
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
}

/** 站内素材 URL → 缩略图 URL；非站内素材（占位图/外链）原样返回 */
export function thumbOf(url: string): string {
  if (!url.startsWith('/uploads/') || url.startsWith('/uploads/thumbs/')) return url;
  return `/uploads/thumbs/${url.slice('/uploads/'.length)}.webp`;
}
