import { reactive } from 'vue';
import { adminApi, api } from '../../../api';

/** 侧栏计数等跨页面共享的轻量状态；页面改动数据后调用 refreshCounts() */
export const studio = reactive({
  posts: 0,
  drafts: 0,
  notes: 0,
});

export async function refreshCounts(): Promise<void> {
  try {
    const [posts, notes] = await Promise.all([adminApi.posts(), api.notes({ pageSize: 1 })]);
    studio.posts = posts.length;
    studio.drafts = posts.filter((p) => p.status === 'draft').length;
    studio.notes = notes.total;
  } catch { /* 计数失败不影响页面 */ }
}

/** 主题切换：View Transition 从点击处圆形扩散；不支持或减弱动效时直接切换 */
export function themeTransition(e: MouseEvent | undefined, run: () => void): void {
  type VT = { ready: Promise<void>; finished: Promise<void> };
  const doc = document as Document & { startViewTransition?: (cb: () => void) => VT };
  const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  if (!doc.startViewTransition || reduced) {
    run();
    return;
  }
  const x = e?.clientX ?? window.innerWidth / 2;
  const y = e?.clientY ?? window.innerHeight / 2;
  const r = Math.hypot(Math.max(x, window.innerWidth - x), Math.max(y, window.innerHeight - y));
  const root = document.documentElement;
  root.classList.add('st-vt');
  const t = doc.startViewTransition(run);
  void t.ready.then(() => {
    root.animate(
      { clipPath: [`circle(0 at ${x}px ${y}px)`, `circle(${r}px at ${x}px ${y}px)`] },
      { duration: 700, easing: 'cubic-bezier(.2,.8,.3,1)', pseudoElement: '::view-transition-new(root)' },
    );
  });
  void t.finished.finally(() => root.classList.remove('st-vt'));
}

/** 浏览器下载 Blob */
export function saveBlob(blob: Blob, name: string): void {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = name;
  a.click();
  window.setTimeout(() => URL.revokeObjectURL(url), 1000);
}

export async function copyText(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    return false;
  }
}
