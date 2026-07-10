import type { Palette } from './types';

/**
 * 内置四季色盘：春绿、夏红、秋黄、冬蓝。
 * 深浅两组共用同一荧光主色（浅色不加深），新增自定义色盘仿结构追加即可。
 * 品牌三荧光色为全局常量（见 styles/tokens.scss），不随色盘变化。
 */
export const palettes: Palette[] = [
  {
    id: 'spring',
    nameKey: 'theme.spring',
    light: {
      bg: '#f4faf5', surface: '#ffffff', surface2: '#e9f4ec',
      text: '#121a15', text2: '#5b7263', border: '#dcebe1',
      primary: '#00c853', primaryDeep: '#00a344', primaryRgb: '0,200,83',
      glass: 'rgba(255,255,255,.55)',
    },
    dark: {
      bg: '#0b120d', surface: '#131c15', surface2: '#1a271d',
      text: '#e9f2ec', text2: '#84a08c', border: '#233529',
      primary: '#00c853', primaryDeep: '#00a344', primaryRgb: '0,200,83',
      glass: 'rgba(19,28,21,.5)',
    },
  },
  {
    id: 'summer',
    nameKey: 'theme.summer',
    light: {
      bg: '#fdf5f6', surface: '#ffffff', surface2: '#f9e9ec',
      text: '#221418', text2: '#7a5a63', border: '#f0dbe1',
      primary: '#ff0032', primaryDeep: '#d40029', primaryRgb: '255,0,50',
      glass: 'rgba(255,255,255,.55)',
    },
    dark: {
      bg: '#120b0d', surface: '#1c1215', surface2: '#26181c',
      text: '#f4e9ec', text2: '#a08088', border: '#33222a',
      primary: '#ff0032', primaryDeep: '#d40029', primaryRgb: '255,0,50',
      glass: 'rgba(28,18,21,.5)',
    },
  },
  {
    id: 'autumn',
    nameKey: 'theme.autumn',
    light: {
      bg: '#fdf9f0', surface: '#ffffff', surface2: '#f8f0dd',
      text: '#201a0e', text2: '#7c6f52', border: '#eee3c8',
      primary: '#ffb300', primaryDeep: '#e09600', primaryRgb: '255,179,0',
      glass: 'rgba(255,255,255,.55)',
    },
    dark: {
      bg: '#131009', surface: '#1d1810', surface2: '#282013',
      text: '#f3eddd', text2: '#a89a78', border: '#362c19',
      primary: '#ffb300', primaryDeep: '#e09600', primaryRgb: '255,179,0',
      glass: 'rgba(29,24,16,.5)',
    },
  },
  {
    id: 'winter',
    nameKey: 'theme.winter',
    light: {
      bg: '#f4f8fd', surface: '#ffffff', surface2: '#e8f0fa',
      text: '#121a24', text2: '#54657c', border: '#dde7f2',
      primary: '#0078ff', primaryDeep: '#005fd6', primaryRgb: '0,120,255',
      glass: 'rgba(255,255,255,.55)',
    },
    dark: {
      bg: '#0b0f16', surface: '#131a24', surface2: '#1a2331',
      text: '#e9eef5', text2: '#84919f', border: '#233042',
      primary: '#0078ff', primaryDeep: '#005fd6', primaryRgb: '0,120,255',
      glass: 'rgba(19,26,36,.5)',
    },
  },
];

export function getPalette(id: string): Palette {
  return palettes.find((p) => p.id === id) ?? palettes[0];
}
