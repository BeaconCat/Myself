import { reactive } from 'vue';
import { adminApi, api } from '../../../api';
import { revealOrigin, themeReveal } from '../../../utils/themeReveal';

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

/** 主题切换：从触发按钮处圆形扩散，支持连续切换（新蒙版覆盖旧蒙版），不阻塞点击 */
export function themeTransition(e: Event | undefined, run: () => void): void {
  themeReveal(revealOrigin(e), run);
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
