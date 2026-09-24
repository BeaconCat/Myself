/** 单个模式（light/dark）下的色盘变量集 */
export interface PaletteColors {
  bg: string;
  surface: string;
  surface2: string;
  text: string;
  text2: string;
  border: string;
  primary: string;
  primaryDeep: string;
  /** "r,g,b" 便于 rgba() 组合 */
  primaryRgb: string;
  glass: string;
  /** 主按钮实底（已保证与 onSolid 对比 ≥ 4.5） */
  solid: string;
  /** 主按钮文字色 */
  onSolid: string;
  /** 链接 / 信号文字色（对背景 ≥ 4.5） */
  ink: string;
}

/** 一套季节/自定义色盘：含深浅两组 */
export interface Palette {
  id: string;
  /** i18n key，如 theme.spring */
  nameKey: string;
  light: PaletteColors;
  dark: PaletteColors;
}
