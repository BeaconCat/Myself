import { defineStore } from 'pinia';
import { getPalette, palettes } from '../themes/palettes';
import type { PaletteColors } from '../themes/types';

export type Mode = 'light' | 'dark';

const STORAGE_KEY = 'myself.theme';

interface Persisted {
  mode: Mode;
  paletteId: string;
}

function load(): Persisted {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return JSON.parse(raw) as Persisted;
  } catch { /* 忽略损坏数据，走默认 */ }
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
  return { mode: prefersDark ? 'dark' : 'light', paletteId: 'summer' };
}

/** 将色盘变量写入 :root，全站 CSS 即时生效 */
function apply(colors: PaletteColors, mode: Mode, paletteId: string): void {
  const root = document.documentElement;
  root.dataset.mode = mode;
  root.dataset.palette = paletteId;
  root.style.setProperty('--bg', colors.bg);
  root.style.setProperty('--surface', colors.surface);
  root.style.setProperty('--surface-2', colors.surface2);
  root.style.setProperty('--text', colors.text);
  root.style.setProperty('--text-2', colors.text2);
  root.style.setProperty('--border', colors.border);
  root.style.setProperty('--primary', colors.primary);
  root.style.setProperty('--primary-deep', colors.primaryDeep);
  root.style.setProperty('--primary-rgb', colors.primaryRgb);
  root.style.setProperty('--glass', colors.glass);
}

export const useThemeStore = defineStore('theme', {
  state: () => load(),
  getters: {
    palette: (s) => getPalette(s.paletteId),
    allPalettes: () => palettes,
  },
  actions: {
    init() {
      apply(this.palette[this.mode], this.mode, this.paletteId);
    },
    setMode(mode: Mode) {
      this.mode = mode;
      this.persistAndApply();
    },
    toggleMode() {
      this.setMode(this.mode === 'light' ? 'dark' : 'light');
    },
    setPalette(id: string) {
      this.paletteId = id;
      this.persistAndApply();
    },
    persistAndApply() {
      localStorage.setItem(
        STORAGE_KEY,
        JSON.stringify({ mode: this.mode, paletteId: this.paletteId }),
      );
      apply(this.palette[this.mode], this.mode, this.paletteId);
    },
  },
});
