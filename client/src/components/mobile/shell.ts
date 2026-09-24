import { reactive } from 'vue';

/**
 * 移动端外壳的共享状态（模块级单例）。
 * 页面通过它与外壳通信：底栏收缩、打开抽屉/搜索、灵动岛提示、点当前 tab 回顶。
 */
export type TabName = 'home' | 'articles' | 'thoughts' | 'about';

export const TAB_ORDER: TabName[] = ['home', 'articles', 'thoughts', 'about'];

export const shell = reactive({
  /** 抽屉进度 0..1（拖拽时为连续值） */
  dp: 0,
  /** 抽屉拖拽中：关闭过渡，跟手 */
  drawerDragging: false,
  /** 底栏下滑收缩 */
  barMin: false,
  /** 详情页已推入（底栏隐藏、首页轮播暂停） */
  pushed: false,
  searchOpen: false,
  searchSeed: '',
  /** 当前一级 tab */
  tab: 'home' as TabName,
  /** 重复点当前 tab：页面监听后滚回顶部 */
  scrollTopSeq: 0,
  /** 品牌 loading 结束（页面入场动画等它） */
  booted: false,
  toast: { text: '', sub: '', seq: 0 },
});

export function openDrawer(): void {
  shell.dp = 1;
}

export function closeDrawer(): void {
  shell.dp = 0;
}

export function openSearch(seed = ''): void {
  shell.searchSeed = seed;
  shell.searchOpen = true;
}

/** 灵动岛轻提示 */
export function toast(text: string, sub = ''): void {
  shell.toast = { text, sub, seq: shell.toast.seq + 1 };
}

/** 复制文本：优先 Clipboard API，非安全上下文回退 execCommand */
export async function copyText(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    const ta = document.createElement('textarea');
    ta.value = text;
    ta.setAttribute('readonly', '');
    ta.style.cssText = 'position:fixed;left:-9999px;top:0;opacity:0';
    document.body.appendChild(ta);
    ta.select();
    let ok = false;
    try {
      ok = document.execCommand('copy');
    } catch {
      ok = false;
    }
    ta.remove();
    return ok;
  }
}

/** 阅读时长估算：中文按 400 字/分钟，英文单词按 220 词/分钟 */
export function readMinutes(md: string): number {
  const cjk = (md.match(/[一-鿿]/g) ?? []).length;
  const words = (md.replace(/[一-鿿]/g, ' ').match(/[A-Za-z0-9_]+/g) ?? []).length;
  return Math.max(1, Math.round(cjk / 400 + words / 220));
}

/** 「2026-07-04 16:20:00」→ 月/日 */
export function monthDay(s: string): { y: number; m: number; d: number } {
  const [y, m, d] = s.slice(0, 10).split('-').map(Number);
  return { y, m, d };
}
