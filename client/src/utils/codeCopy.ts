import { i18n } from '../i18n';
import { copyText } from './clipboard';

/**
 * 代码块一键复制：Markdown 渲染的每个代码块右上角带一个 .code-copy 按钮（见 utils/markdown.ts），
 * 这里在 document 上挂一个点击委托统一处理，文章、随想、移动端通用。复制成功后按钮短暂显示对勾与「已复制」，
 * 失败显示「复制失败」；读屏通过按钮名称的变化得到同样的反馈。
 */
export function installCodeCopy(): void {
  document.addEventListener('click', async (e) => {
    const btn = (e.target as HTMLElement | null)?.closest<HTMLButtonElement>('.code-copy');
    if (!btn) return;
    e.preventDefault();
    e.stopPropagation();
    const code = btn.parentElement?.querySelector('code')?.innerText ?? '';
    const ok = await copyText(code);
    const t = i18n.global.t;
    const label = t('content.article.copyCode');
    btn.dataset.tip = ok ? t('content.article.codeCopied') : t('content.article.copyFailed');
    btn.setAttribute('aria-label', btn.dataset.tip);
    btn.classList.toggle('done', ok);
    btn.classList.add('tipping');
    window.clearTimeout(Number(btn.dataset.timer));
    btn.dataset.timer = String(window.setTimeout(() => {
      btn.classList.remove('done', 'tipping');
      btn.setAttribute('aria-label', label);
    }, 1600));
  });
}
