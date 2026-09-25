import { reactive } from 'vue';

/** ⌘K 搜索浮层的开关状态：任意桌面组件可调用 openSearch() 唤起（可带预填关键词） */
export const searchPalette = reactive({ open: false, seed: '' });

export function openSearch(seed = ''): void {
  searchPalette.seed = seed;
  searchPalette.open = true;
}

export function closeSearch(): void {
  searchPalette.open = false;
}
