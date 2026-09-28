import { defineStore } from 'pinia';
import { derivePalette } from '../themes/derive';
import type { Palette, PaletteColors } from '../themes/types';
import { useConfigStore } from './config';

export type Mode = 'light' | 'dark';
/** 界面风格：cards = 高密度卡片（默认）；clean = 透明背景简洁版式。与 mode × palette 正交 */
export type UiStyle = 'cards' | 'clean';

const STORAGE_KEY = 'myself.theme';

interface Persisted {
  mode: Mode | '';
  paletteId: string;
  style?: UiStyle | '';
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
  root.style.setProperty('--solid', colors.solid);
  root.style.setProperty('--on-solid', colors.onSolid);
  root.style.setProperty('--ink', colors.ink);
}

function isStyle(v: unknown): v is UiStyle {
  return v === 'cards' || v === 'clean';
}

/** 界面风格写到 :root[data-style]，tokens.scss 按它切换卡片 / 统计条 / 分区间距 token */
function applyStyle(style: UiStyle): void {
  document.documentElement.dataset.style = style;
}

/** 圆角基准（px，0–24）：tokens.scss 按比例派生 --r-xs…--r-xl 与兼容的 --radius / --radius-lg */
export function applyRadius(base: number): void {
  const v = Number.isFinite(base) ? Math.min(24, Math.max(0, base)) : 10;
  document.documentElement.style.setProperty('--r-base', String(v));
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
    style: 'clean' as UiStyle,
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
    allowUserStyle(): boolean {
      return useConfigStore().cfg.theme.allowUserStyle ?? true;
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

      const defStyle: UiStyle = isStyle(cfg.defaultStyle) ? cfg.defaultStyle : 'clean';
      this.style = this.allowUserStyle && isStyle(saved.style) ? saved.style : defStyle;
      applyStyle(this.style);

      apply(this.palette[this.mode], this.mode, this.paletteId);
      applyRadius(useConfigStore().cfg.theme.radius ?? 10);
    },
    setMode(mode: Mode) {
      this.mode = mode;
      this.persistAndApply();
    },
    toggleMode() {
      this.setMode(this.mode === 'light' ? 'dark' : 'light');
    },
    setStyle(style: UiStyle, isAdmin = false) {
      if (!this.allowUserStyle && !isAdmin) return;
      this.style = style;
      this.persistAndApply();
    },
    toggleStyle(isAdmin = false) {
      this.setStyle(this.style === 'cards' ? 'clean' : 'cards', isAdmin);
    },
    setPalette(id: string, isAdmin = false) {
      if (!this.allowUserPalette && !isAdmin) return;
      this.paletteId = id;
      this.persistAndApply();
    },
    persistAndApply() {
      localStorage.setItem(
        STORAGE_KEY,
        JSON.stringify({ mode: this.mode, paletteId: this.paletteId, style: this.style }),
      );
      applyStyle(this.style);
      apply(this.palette[this.mode], this.mode, this.paletteId);
    },
  },
});
