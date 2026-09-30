import { nextTick, reactive } from 'vue';
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
    const [posts, notes] = await Promise.all([adminApi.posts(), api.notes({ pageSize: 1, all: true })]);
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

/**
 * 未保存检测用的稳定序列化：数字字符串与数字视为相等、undefined 字段忽略，
 * 避免表单控件把 "10" 规范成 10 等无意义差异触发「离开确认」。
 */
export function stableJson(v: unknown): string {
  return JSON.stringify(v, (_k, val: unknown) => {
    if (typeof val === 'string' && val.trim() !== '' && !Number.isNaN(Number(val)) && /^-?\d+(\.\d+)?$/.test(val.trim())) {
      return Number(val);
    }
    return val;
  });
}

/** 等子组件挂载并完成初始化（v-model 规范化、默认值回填）后再拍快照 */
export async function settle(): Promise<void> {
  await nextTick();
  await new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r())));
  await nextTick();
}
