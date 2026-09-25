/**
 * 内容线（文章列表 / 详情 / 随想）共享的纯函数：日期格式、阅读时长、复制。
 * 时间串形如「2026-07-04 16:20:00」（服务端本地时间），只做字符串切分，不走时区换算。
 */

export interface Ymd {
  y: number;
  m: number;
  d: number;
}

export function ymdOf(s: string): Ymd {
  const [y, m, d] = s.slice(0, 10).split('-').map(Number);
  return { y, m, d };
}

/** 2026.07.04 */
export function dotted(s: string): string {
  return s.slice(0, 10).replaceAll('-', '.');
}

/** Date → YYYY-MM-DD（本地） */
export function isoDay(d: Date): string {
  const p = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`;
}

/** 阅读时长：中文 400 字/分钟，英文 220 词/分钟（与移动端一致） */
export function readMinutes(md: string): number {
  const cjk = (md.match(/[一-鿿]/g) ?? []).length;
  const words = (md.replace(/[一-鿿]/g, ' ').match(/[A-Za-z0-9_]+/g) ?? []).length;
  return Math.max(1, Math.round(cjk / 400 + words / 220));
}

/** 正文字数（去掉 Markdown 标记后的可见字符，按十取整） */
export function wordCount(md: string): number {
  const plain = md
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/!\[[^\]]*]\([^)]*\)/g, ' ')
    .replace(/\[([^\]]*)]\([^)]*\)/g, '$1')
    .replace(/[#>*_`|\-\s]/g, '');
  return Math.max(10, Math.round(plain.length / 10) * 10);
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

/** 元素是否已吸顶（上缘贴到 offset 以内） */
export function isStuck(el: HTMLElement | null, offset: number): boolean {
  return !!el && el.getBoundingClientRect().top <= offset + 0.5;
}
