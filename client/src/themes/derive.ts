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

/** 由主色对生成整套色盘：背景/表面按主色轻染，文字随明度走 */
export function derivePalette(preset: ThemePreset): Palette {
  const { primary, primaryDeep } = preset;
  const rgb = hexToRgb(primary).join(',');

  const light: PaletteColors = {
    bg: mix('#f7f8fa', primary, 0.04),
    surface: '#ffffff',
    surface2: mix('#eef0f4', primary, 0.06),
    text: mix('#17181c', primary, 0.08),
    text2: mix('#5c6270', primary, 0.14),
    border: mix('#e2e5ec', primary, 0.1),
    primary,
    primaryDeep,
    primaryRgb: rgb,
    glass: 'rgba(255,255,255,.55)',
  };

  const darkBg = mix('#0e0f13', primary, 0.05);
  const dark: PaletteColors = {
    bg: darkBg,
    surface: mix('#17181e', primary, 0.06),
    surface2: mix('#1f2129', primary, 0.07),
    text: mix('#eceef2', primary, 0.04),
    text2: mix('#8a90a0', primary, 0.1),
    border: mix('#262933', primary, 0.1),
    primary,
    primaryDeep,
    primaryRgb: rgb,
    glass: 'rgba(23,24,30,.5)',
  };

  return { id: preset.id, nameKey: preset.name, light, dark };
}
