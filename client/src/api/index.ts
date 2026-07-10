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

const BASE = '/api/v1';

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`);
  if (!res.ok) throw new Error(`api_error_${res.status}`);
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
