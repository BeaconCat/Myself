import type { Palette } from './types';

/**
 * 内置四季色盘。新增自定义色盘：仿照结构追加对象即可。
 * 品牌三荧光色为全局常量（见 styles/tokens.scss），不随色盘变化。
 */
export const palettes: Palette[] = [
  {
    id: 'spring',
    nameKey: 'theme.spring',
    light: {
      bg: '#fdf6f7', surface: '#ffffff', surface2: '#f9eaee',
      text: '#221418', text2: '#7a5a63', border: '#f0dbe1',
      primary: '#ff0032', primaryDeep: '#d40029', primaryRgb: '255,0,50',
      glass: 'rgba(255,255,255,.55)',
    },
    dark: {
      bg: '#120b0d', surface: '#1c1215', surface2: '#26181c',
      text: '#f4e9ec', text2: '#a08088', border: '#33222a',
      primary: '#ff0032', primaryDeep: '#c70027', primaryRgb: '255,0,50',
      glass: 'rgba(28,18,21,.5)',
    },
  },
  {
    id: 'summer',
    nameKey: 'theme.summer',
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
  {
    id: 'autumn',
    nameKey: 'theme.autumn',
    light: {
      bg: '#fdf9f0', surface: '#ffffff', surface2: '#f8f0dd',
      text: '#201a0e', text2: '#7c6f52', border: '#eee3c8',
      primary: '#c78800', primaryDeep: '#a06d00', primaryRgb: '199,136,0',
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
      bg: '#f6f8f9', surface: '#ffffff', surface2: '#eceff2',
      text: '#161a1d', text2: '#5b6670', border: '#dfe4e9',
      primary: '#2b6cb0', primaryDeep: '#1e5590', primaryRgb: '43,108,176',
      glass: 'rgba(255,255,255,.55)',
    },
    dark: {
      bg: '#0d0f12', surface: '#15181c', surface2: '#1d2126',
      text: '#e9edf0', text2: '#8b949c', border: '#252b31',
      primary: '#4d9fff', primaryDeep: '#2b7fe0', primaryRgb: '77,159,255',
      glass: 'rgba(21,24,28,.5)',
    },
  },
];

export function getPalette(id: string): Palette {
  return palettes.find((p) => p.id === id) ?? palettes[0];
}
