/**
 * 代码块一键复制：Markdown 渲染的每个代码块右上角带一个 .code-copy 按钮（见 utils/markdown.ts），
 * 这里在 document 上挂一个点击委托统一处理，文章、随想、移动端通用。复制成功后按钮短暂显示对勾。
 */
export function installCodeCopy(): void {
  document.addEventListener('click', async (e) => {
    const btn = (e.target as HTMLElement | null)?.closest<HTMLButtonElement>('.code-copy');
    if (!btn) return;
    e.preventDefault();
    e.stopPropagation();
    const code = btn.parentElement?.querySelector('code')?.innerText ?? '';
    try {
      await navigator.clipboard.writeText(code);
    } catch {
      // 非安全上下文等无剪贴板权限时：选区兜底
      const range = document.createRange();
      const el = btn.parentElement?.querySelector('code');
      if (!el) return;
      range.selectNodeContents(el);
      const sel = window.getSelection();
      sel?.removeAllRanges();
      sel?.addRange(range);
      document.execCommand('copy');
      sel?.removeAllRanges();
    }
    btn.classList.add('done');
    window.setTimeout(() => btn.classList.remove('done'), 1600);
  });
}
