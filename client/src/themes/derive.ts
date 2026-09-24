import type { Palette, PaletteColors } from './types';
import type { ThemePreset } from '../stores/config';

function hexToRgb(hex: string): [number, number, number] {
  const value = hex.replace('#', '');
  return [
    parseInt(value.slice(0, 2), 16),
    parseInt(value.slice(2, 4), 16),
    parseInt(value.slice(4, 6), 16),
  ];
}

function mix(a: string, b: string, ratio: number): string {
  const ra = hexToRgb(a);
  const rb = hexToRgb(b);
  const channel = (i: number) => Math.round(ra[i] + (rb[i] - ra[i]) * ratio);
  return `#${[0, 1, 2].map((i) => channel(i).toString(16).padStart(2, '0')).join('')}`;
}

/** WCAG 相对亮度 */
function luminance(hex: string): number {
  const [r, g, b] = hexToRgb(hex).map((v) => {
    const c = v / 255;
    return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function contrast(a: string, b: string): number {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (hi + 0.05) / (lo + 0.05);
}

/** 把 color 逐步推向 toward，直到与 against 的对比度 ≥ target */
function pushToContrast(color: string, toward: string, against: string, target: number): string {
  let out = color;
  for (let t = 0; t <= 1 && contrast(out, against) < target; t += 0.02) out = mix(color, toward, t);
  return out;
}

const DARK_TEXT = '#14161b';

/**
 * 主按钮实底与字色：亮主色（春绿、秋黄）配深字；其余配白字，
 * 并把底色向黑推到白字对比 ≥ 4.5。
 */
function solidPair(primary: string): { solid: string; onSolid: string } {
  if (luminance(primary) > 0.3) return { solid: primary, onSolid: DARK_TEXT };
  return { solid: pushToContrast(primary, '#000000', '#ffffff', 4.5), onSolid: '#ffffff' };
}

/**
 * 由主色对生成整套色盘：背景/表面按主色轻染（2–3%，避免偏粉偏橄榄），文字随明度走；
 * 同时派生 solid / onSolid（主按钮）与 ink（链接与信号文字，对背景 ≥ 4.5）。
 */
export function derivePalette(preset: ThemePreset): Palette {
  const { primary, primaryDeep } = preset;
  const rgb = hexToRgb(primary).join(',');
  const pair = solidPair(primary);

  const lightBg = mix('#f7f8fa', primary, 0.025);
  const light: PaletteColors = {
    bg: lightBg,
    surface: '#ffffff',
    surface2: mix('#eef0f4', primary, 0.035),
    text: mix('#17181c', primary, 0.05),
    text2: mix('#5c6270', primary, 0.08),
    border: mix('#e2e5ec', primary, 0.06),
    primary,
    primaryDeep,
    primaryRgb: rgb,
    glass: 'rgba(255,255,255,.55)',
    ...pair,
    ink: pushToContrast(primary, '#000000', lightBg, 4.5),
  };

  const darkBg = mix('#0e0f13', primary, 0.03);
  const dark: PaletteColors = {
    bg: darkBg,
    surface: mix('#17181e', primary, 0.035),
    surface2: mix('#1f2129', primary, 0.04),
    text: mix('#eceef2', primary, 0.03),
    text2: mix('#8a90a0', primary, 0.06),
    border: mix('#262933', primary, 0.06),
    primary,
    primaryDeep,
    primaryRgb: rgb,
    glass: 'rgba(23,24,30,.5)',
    ...pair,
    ink: pushToContrast(primary, '#ffffff', darkBg, 4.5),
  };

  return { id: preset.id, nameKey: preset.name, light, dark };
}
