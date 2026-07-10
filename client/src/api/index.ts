/** RESTful API 客户端（/api/v1，开发态经 Vite 代理到后端） */

export interface Post {
  id: number;
  slug: string;
  title: string;
  excerpt: string;
  covers: string[];
  tags: string[];
  createdAt: string;
  updatedAt: string;
  contentMd?: string;
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

export interface ArchiveGroup {
  month: string;
  items: Post[];
}

export interface Note {
  id: number;
  contentMd: string;
  mood: string;
  images: string[];
  createdAt: string;
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

import { readToken } from '../stores/auth';

/** 带管理员 Token 的请求 */
async function authed<T>(path: string, init: RequestInit = {}): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${readToken()}`,
      ...init.headers,
    },
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error((body as { error?: string }).error ?? `api_error_${res.status}`);
  }
  return res.json() as Promise<T>;
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
  archive: () => get<ArchiveGroup[]>('/archive'),
  notes: (params: { page?: number; pageSize?: number } = {}) => {
    const query = new URLSearchParams();
    if (params.page) query.set('page', String(params.page));
    if (params.pageSize) query.set('pageSize', String(params.pageSize));
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
}

export const adminApi = {
  login: async (username: string, password: string): Promise<string> => {
    const res = await fetch(`${BASE}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    });
    if (!res.ok) throw new Error('bad_credentials');
    const data = (await res.json()) as { token: string };
    return data.token;
  },
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
  createNote: (note: { contentMd: string; mood: string; images: string[] }) =>
    authed<{ id: number }>('/admin/notes', { method: 'POST', body: JSON.stringify(note) }),
  deleteNote: (id: number) =>
    authed<{ ok: boolean }>(`/admin/notes/${id}`, { method: 'DELETE' }),
};
