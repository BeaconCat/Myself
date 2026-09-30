import type { MediaItem, MediaKindName } from '../api';

/** 与后端 media_kinds.go 的分类一致；旧后端未下发 kind 时按扩展名推断 */
const EXT: Record<string, MediaKindName> = {};
for (const e of ['png', 'jpg', 'jpeg', 'webp', 'gif']) EXT[e] = 'image';
for (const e of ['mp4', 'm4v', 'webm', 'mov', 'ogv']) EXT[e] = 'video';
for (const e of ['mp3', 'm4a', 'aac', 'ogg', 'oga', 'wav', 'flac', 'opus']) EXT[e] = 'audio';
for (const e of ['zip', '7z', 'rar', 'tar', 'gz', 'tgz', 'bz2', 'xz']) EXT[e] = 'archive';

export function extOf(name: string): string {
  const m = /\.([a-z0-9]{1,10})(?:[?#].*)?$/i.exec(name);
  return m ? m[1].toLowerCase() : '';
}

export function kindOfExt(ext: string): MediaKindName {
  return EXT[ext.toLowerCase()] ?? 'file';
}

export function mediaKind(m: Pick<MediaItem, 'kind' | 'name'>): MediaKindName {
  return m.kind ?? kindOfExt(extOf(m.name));
}

/** 上传控件的 accept：按允许的类型拼接（含 file 时不限制） */
export function acceptFor(kinds: MediaKindName[]): string {
  if (kinds.includes('file')) return '';
  return Object.entries(EXT).filter(([, k]) => kinds.includes(k)).map(([e]) => `.${e}`).join(',');
}
