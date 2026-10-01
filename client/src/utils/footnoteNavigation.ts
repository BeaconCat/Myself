import type { Router } from 'vue-router';

export function scrollToFootnote(hash: string): Promise<false> {
  const id = hash.slice(1);
  return new Promise(resolve => {
    const deadline = performance.now() + 6000;
    const find = () => {
      const target = document.getElementById(id);
      if (target) {
        target.scrollIntoView({ behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth', block: 'start' });
        target.focus({ preventScroll: true });
        resolve(false);
      } else if (performance.now() >= deadline) resolve(false);
      else requestAnimationFrame(find);
    };
    find();
  });
}

export function installFootnoteNavigation(router: Router): void {
  document.addEventListener('click', event => {
    if (event.button !== 0 || event.ctrlKey || event.metaKey || event.shiftKey || event.altKey) return;
    const anchor = (event.target as Element | null)?.closest<HTMLAnchorElement>('a[data-footnote-link]');
    const hash = anchor?.getAttribute('href') ?? '';
    if (!hash.startsWith('#myself-fn-') || !document.getElementById(hash.slice(1))) return;
    event.preventDefault();
    if (router.currentRoute.value.hash === hash) void scrollToFootnote(hash);
    else void router.push({ hash });
  });
}
