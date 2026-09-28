/**
 * 默认封面：scripts/gen_covers.py 程序生成的 12 幅抽象构图（client/public/covers/NN.webp，另有 NN-s.webp 小图）。
 * 无封面的文章、随想占位图与后台示意都从这里按种子稳定取一幅，全站同一套画面语言。
 */
export const DEFAULT_COVERS = ['01', '02', '03', '04', '05', '06', '07', '08', '09', '10', '11', '12'] as const;

function hash(seed: string): number {
  let h = 0;
  for (let i = 0; i < seed.length; i++) h = (h * 31 + seed.charCodeAt(i)) >>> 0;
  return h;
}

export function coverUrl(id: string, small = false): string {
  return `/covers/${id}${small ? '-s' : ''}.webp`;
}

/** 按种子（slug、id 等）稳定取一幅默认封面 */
export function defaultCover(seed: string | number, small = false): string {
  return coverUrl(DEFAULT_COVERS[hash(String(seed)) % DEFAULT_COVERS.length], small);
}

/** 历史 `css:<kind>` / 光影种类名 → 固定的一幅（保证旧数据与示意位画面稳定） */
const KIND_MAP: Record<string, string> = {
  door: '12', slit: '02', beams: '07', season: '10', arcs: '01', page: '03', pages: '03',
  signal: '05', key: '04', grid: '06', band: '07', bands: '07', dawn: '01', night: '02',
  blind: '12', paper: '03', beam: '05', dusk: '01', pane: '12',
};

export function coverForKind(kind: string, small = false): string {
  return coverUrl(KIND_MAP[kind] ?? DEFAULT_COVERS[hash(kind) % DEFAULT_COVERS.length], small);
}
