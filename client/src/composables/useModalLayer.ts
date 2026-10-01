import { watch, type Ref } from 'vue';

const layers: HTMLElement[] = [];
const blocked = new WeakMap<HTMLElement, { count: number; original: boolean }>();

/** Keep modal focus and background isolation correct even when dialogs are nested. */
export function useModalLayer(panel: Ref<HTMLElement | null>): () => boolean {
  const isTop = () => !!panel.value && layers.at(-1) === panel.value;
  watch(panel, (el, _, cleanup) => {
    if (!el) return;
    const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    let portal = el;
    while (portal.parentElement && portal.parentElement !== document.body) portal = portal.parentElement;
    const siblings = [...document.body.children].filter((node): node is HTMLElement =>
      node instanceof HTMLElement && node !== portal && !node.matches('script, style, link, [aria-live], [role="status"], [role="alert"]'));
    for (const node of siblings) {
      const state = blocked.get(node) ?? { count: 0, original: node.inert };
      state.count++;
      blocked.set(node, state);
      node.inert = true;
    }
    layers.push(el);
    const portals = () => [...new Set([...el.querySelectorAll('[aria-controls]')].flatMap(anchor => {
      return (anchor.getAttribute('aria-controls') ?? '').split(/\s+/).flatMap(id => {
        let node = document.getElementById(id);
        if (!node || el.contains(node)) return [];
        while (node.parentElement && node.parentElement !== document.body) node = node.parentElement;
        return [node];
      });
    }))];
    const inside = (node: Node) => el.contains(node) || portals().some(portal => portal.contains(node));
    const controls = () => [el, ...portals()].flatMap(root => [...root.querySelectorAll<HTMLElement>('button, a[href], input, textarea, select, [tabindex]')])
      .filter(node => node.tabIndex >= 0 && !node.matches(':disabled') && !node.closest('[inert]') && node.getClientRects().length > 0);
    const focusFirst = () => (controls()[0] ?? el).focus({ preventScroll: true });
    const contain = (event: FocusEvent) => {
      if (isTop() && event.target instanceof Node && !inside(event.target)) focusFirst();
    };
    const onTab = (event: KeyboardEvent) => {
      if (event.key !== 'Tab' || !isTop()) return;
      const items = controls(), first = items[0], last = items.at(-1);
      if (!first || (event.shiftKey && (document.activeElement === first || document.activeElement === el))) {
        event.preventDefault(); (last ?? el).focus({ preventScroll: true });
      } else if (!event.shiftKey && (document.activeElement === last || document.activeElement === el)) {
        event.preventDefault(); first.focus({ preventScroll: true });
      }
    };
    document.addEventListener('focusin', contain);
    document.addEventListener('keydown', onTab);
    if (!el.contains(document.activeElement)) focusFirst();
    cleanup(() => {
      const wasTop = layers.at(-1) === el;
      layers.splice(layers.indexOf(el), 1);
      document.removeEventListener('focusin', contain);
      document.removeEventListener('keydown', onTab);
      for (const node of siblings) {
        const state = blocked.get(node)!;
        if (--state.count === 0) { node.inert = state.original; blocked.delete(node); }
      }
      if (wasTop && previous?.isConnected && !previous.closest('[inert]')) previous.focus({ preventScroll: true });
    });
  }, { flush: 'post' });
  return isTop;
}
