import { reactive } from 'vue';

/** Studio 轻提示：顶部居中胶囊，可带一个动作按钮 */
export interface Toast {
  id: number;
  msg: string;
  icon: string;
  action?: string;
  fn?: () => void;
  leaving: boolean;
}

export const toasts = reactive<Toast[]>([]);
let seq = 0;

export function dismiss(id: number): void {
  const t = toasts.find((x) => x.id === id);
  if (!t || t.leaving) return;
  t.leaving = true;
  window.setTimeout(() => {
    const i = toasts.findIndex((x) => x.id === id);
    if (i >= 0) toasts.splice(i, 1);
  }, 300);
}

export function toast(
  msg: string,
  opts: { icon?: string; action?: string; fn?: () => void; ms?: number } = {},
): void {
  seq += 1;
  const id = seq;
  toasts.push({ id, msg, icon: opts.icon ?? 'check', action: opts.action, fn: opts.fn, leaving: false });
  window.setTimeout(() => dismiss(id), opts.ms ?? (opts.action ? 4200 : 2600));
  const alive = toasts.filter((x) => !x.leaving);
  if (alive.length > 3) dismiss(alive[0].id);
}
