import { reactive } from 'vue';

/** 关于页轻提示（复制成功、留言提交等），由 AboutModules 渲染 */
export const kitToast = reactive({ text: '', on: false, timer: 0 });

export function toast(text: string): void {
  kitToast.text = text;
  kitToast.on = true;
  window.clearTimeout(kitToast.timer);
  kitToast.timer = window.setTimeout(() => { kitToast.on = false; }, 1800);
}
