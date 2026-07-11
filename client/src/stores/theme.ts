import { defineStore } from 'pinia';
import { derivePalette } from '../themes/derive';
import type { Palette, PaletteColors } from '../themes/types';
import { useConfigStore } from './config';

export type Mode = 'light' | 'dark';

const STORAGE_KEY = 'myself.theme';

interface Persisted {
  mode: Mode | '';
  paletteId: string;
}

function loadPersisted(): Persisted {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return JSON.parse(raw) as Persisted;
  } catch { /* 忽略损坏数据 */ }
  return { mode: '', paletteId: '' };
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

/** 按月份取季节色盘 id */
function seasonPaletteId(): string {
  const month = new Date().getMonth() + 1;
  if (month >= 3 && month <= 5) return 'spring';
  if (month >= 6 && month <= 8) return 'summer';
  if (month >= 9 && month <= 11) return 'autumn';
  return 'winter';
}

export const useThemeStore = defineStore('theme', {
  state: () => ({
    mode: 'dark' as Mode,
    paletteId: 'summer',
  }),
  getters: {
    /** 访客可见色盘（配置驱动，主题色对派生整套） */
    allPalettes(): Palette[] {
      return useConfigStore().visiblePresets.map(derivePalette);
    },
    palette(): Palette {
      return (
        this.allPalettes.find((p) => p.id === this.paletteId) ?? this.allPalettes[0]
      );
    },
    allowUserPalette(): boolean {
      return useConfigStore().cfg.theme.allowUserPalette;
    },
  },
  actions: {
    /** 站点配置加载后调用：默认值/季节自动切换/用户偏好合并 */
    init() {
      const cfg = useConfigStore().cfg.theme;
      const saved = loadPersisted();

      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      this.mode = saved.mode || cfg.defaultMode || (prefersDark ? 'dark' : 'light');

      let paletteId = cfg.autoSwitch === 'season' ? seasonPaletteId() : '';
      if (!paletteId) paletteId = (cfg.allowUserPalette && saved.paletteId) || cfg.defaultPaletteId;
      if (!cfg.allowUserPalette) paletteId = cfg.defaultPaletteId;
      this.paletteId = this.allPalettes.some((p) => p.id === paletteId)
        ? paletteId
        : this.allPalettes[0]?.id ?? 'summer';

      apply(this.palette[this.mode], this.mode, this.paletteId);
    },
    setMode(mode: Mode) {
      this.mode = mode;
      this.persistAndApply();
    },
    toggleMode() {
      this.setMode(this.mode === 'light' ? 'dark' : 'light');
    },
    setPalette(id: string, isAdmin = false) {
      if (!this.allowUserPalette && !isAdmin) return;
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
