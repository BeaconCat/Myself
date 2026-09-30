/**
 * 复制文本（全站统一入口）：优先 Clipboard API；非安全上下文（http 访问、局域网 IP）里 navigator.clipboard
 * 不存在或被拒，回退到隐藏 textarea + execCommand('copy')。复制后把焦点还给原来的元素。
 * 返回是否成功，调用方据此给出「已复制」或手动复制的提示。
 */
export async function copyText(text: string): Promise<boolean> {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
      return true;
    }
  } catch { /* 回退 */ }
  const active = document.activeElement as HTMLElement | null;
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
  active?.focus?.({ preventScroll: true });
  return ok;
}
