import { reactive } from 'vue';
import { adminApi, api, type AdminPost, type MediaItem, type Note } from '../../api';

/**
 * 移动后台外壳的共享状态（模块单例）：
 * action sheet、写随想 composer、全屏 sheet 时的舞台后退量、待上传文件、数据刷新信号、灵动岛提示。
 */

export type IslandKind = 'ok' | 'error' | 'info';

interface IslandState {
  seq: number;
  text: string;
  sub: string;
  kind: IslandKind;
  open: boolean;
}

export const island = reactive<IslandState>({ seq: 0, text: '', sub: '', kind: 'ok', open: false });

let islandTimer = 0;
/** 灵动岛提示：黑色胶囊从顶部岛屿展开，2.4s 后收回 */
export function toast(text: string, sub = '', kind: IslandKind = 'ok'): void {
  island.seq += 1;
  island.text = text;
  island.sub = sub;
  island.kind = kind;
  island.open = true;
  window.clearTimeout(islandTimer);
  islandTimer = window.setTimeout(() => {
    island.open = false;
  }, 2400);
}

export const shell = reactive({
  /** 中央 + 的 action sheet */
  actionSheet: false,
  /** 写随想 composer：open + 编辑的随想 id（null 为新建） */
  composer: { open: false, id: null as number | null, seq: 0 },
  /** 0..1：全屏 sheet 时后台页面后退（缩小 + 下沉 + 变暗） */
  stack: 0,
  /** 从 action sheet 选好的待上传文件，素材页接手 */
  pendingUploads: [] as File[],
  /** 数据变更信号：视图 watch 对应计数刷新 */
  bump: { posts: 0, notes: 0, media: 0 },
});

export function openComposer(id: number | null = null): void {
  shell.actionSheet = false;
  shell.composer.id = id;
  shell.composer.seq += 1;
  shell.composer.open = true;
}

export function closeComposer(): void {
  shell.composer.open = false;
}

/* ---------- 数据缓存：切页先显示缓存，后台静默刷新（无缓存时显示骨架屏） ---------- */


export const cache = reactive({
  posts: null as AdminPost[] | null,
  notes: null as Note[] | null,
  notesTotal: 0,
  notesPage: 0,
  media: null as MediaItem[] | null,
});

export async function loadPosts(): Promise<AdminPost[]> {
  const list = await adminApi.posts();
  cache.posts = list;
  return list;
}

const NOTE_PAGE = 50;
export async function loadNotes(): Promise<Note[]> {
  const res = await api.notes({ page: 1, pageSize: NOTE_PAGE, all: true });
  cache.notes = res.items;
  cache.notesTotal = res.total;
  cache.notesPage = 1;
  return res.items;
}

export async function loadMoreNotes(): Promise<void> {
  if (!cache.notes || cache.notes.length >= cache.notesTotal) return;
  const res = await api.notes({ page: cache.notesPage + 1, pageSize: NOTE_PAGE, all: true });
  const seen = new Set(cache.notes.map((n) => n.id));
  cache.notes = [...cache.notes, ...res.items.filter((n) => !seen.has(n.id))];
  cache.notesPage += 1;
  cache.notesTotal = res.total;
}

export async function loadMedia(): Promise<MediaItem[]> {
  const list = await adminApi.media();
  cache.media = list;
  return list;
}
