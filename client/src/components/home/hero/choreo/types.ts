import type { ChoreoRun, Rect } from './engine';
import type { DeckGeom } from './geom';

/**
 * Hero 编舞契约：左侧文字动效与右侧卡组动效是两个独立注册表，可任意搭配。
 *
 * - 前台 HeroCarousel 按站点配置 hero.textAnim / hero.cardAnim 取用；
 * - 后台「外观 → 首页轮播」通过 HeroMixer.vue 实时预览并保存组合。
 */

/** 注册表条目的展示信息（后台选择器用） */
export interface ChoreoMeta {
  /** 持久化到配置的稳定 id */
  id: string;
  /** 中文名，如「门缝光扫」 */
  name: string;
  /** 一句话理念 */
  desc: string;
  /** 标签：品牌感 / 克制 / 纵深 … */
  tag: string;
  /** 完整切换时长（ms，出场 + 入场） */
  duration: number;
}

/** 文字动效 id（与 TEXT_CHOREOS 保持一致） */
export type TextChoreoId = 'lightscan' | 'dolly' | 'shift' | 'caption' | 'glow';
/** 卡组动效 id（与 CARD_CHOREOS 保持一致） */
export type CardChoreoId = 'hinge' | 'dolly' | 'shared' | 'parallax' | 'door';

export const DEFAULT_TEXT_CHOREO: TextChoreoId = 'lightscan';
export const DEFAULT_CARD_CHOREO: CardChoreoId = 'hinge';

/* ===== 运行时契约（编舞实现与 Hero 组件之间） ===== */


/** 拆行后的一行：ln 行容器（可裁切）> inner（nowrap，可做 mask / 字距） */
export interface LineEls {
  ln: HTMLElement;
  inner: HTMLElement;
  chars: HTMLElement[];
}

/** HeroText 采集的元素 */
export interface TextEls {
  root: HTMLElement;
  tag: HTMLElement;
  title: HTMLElement;
  excerpt: HTMLElement;
  btn: HTMLElement;
  lines: LineEls[];
  exLines: LineEls[];
  /** 标题全部字符（按阅读顺序） */
  chars: HTMLElement[];
}

/** 卡片三层：变换层 / 外观层（圆角裁切、开合）/ 画面层；另有与外观层并列的中性阴影层 */
export interface CardEl {
  el: HTMLElement;
  /**
   * 中性阴影层（外观层的兄弟节点）：阴影不随 clip-path / 开合被裁掉，
   * 编舞只动画它的 opacity——离场开头淡出、入场结尾淡入（见 card/lights.ts）
   */
  shade: HTMLElement;
  sheet: HTMLElement;
  cv: HTMLElement;
  slot: number;
  /** 静止亮度 */
  b: number;
}

/** HeroDeck 采集的元素（cards 按槽位排序，front 在前） */
export interface CardEls extends DeckGeom {
  root: HTMLElement;
  album: HTMLElement;
  /** 卡组环境光 */
  bg: HTMLElement;
  /** 前卡门缝光外溢到地面的品牌光（全站唯一的卡片发光），只动画 opacity */
  spill: HTMLElement;
  /** 外观层静止圆角（px 字符串，随 --r-base） */
  radius: string;
  cards: CardEl[];
  front: CardEl;
  backs: CardEl[];
  /** 条目序号 */
  idx: number;
}

/** 缩略导航（仅声明 chrome: 'rail' 的卡组动效使用） */
export interface NavEls {
  thumb: (i: number) => HTMLElement | null;
  ind: HTMLElement | null;
  indX: (i: number) => number;
  prevIdx: number;
}

/** 常驻光效层 */
export interface FxEls {
  flare: HTMLElement;
  floor: HTMLElement;
  vig: HTMLElement;
}

/** 编舞工具集：ChoreoRun 的注册能力 + 舞台信息 */
export interface ChoreoCtx {
  a: ChoreoRun['a'];
  set: ChoreoRun['set'];
  make: ChoreoRun['make'];
  /** 相对 Hero 舞台的矩形（已抵消外层缩放） */
  rel: (el: Element) => Rect;
  rng: () => number;
  mode: 'light' | 'dark';
  mobile: boolean;
  W: number;
  H: number;
  fx: FxEls;
  nav: NavEls | null;
  /** 光源：新卡组相册中心（文字「门光」按距离点亮用） */
  light: () => Rect;
  /** 读取根节点 CSS 变量的计算值 */
  color: (name: string) => string;
}

export interface TextChoreo {
  meta: ChoreoMeta & { id: TextChoreoId };
  exit: (e: TextEls, t: ChoreoCtx) => void;
  enter: (e: TextEls, t: ChoreoCtx) => void;
}

export interface CardChoreo {
  meta: ChoreoMeta & { id: CardChoreoId };
  /** 需要的常驻 chrome：rail = 卡组下方缩略导航（替代进度胶囊） */
  chrome?: 'rail';
  exit: (e: CardEls, t: ChoreoCtx) => void;
  enter: (e: CardEls, t: ChoreoCtx) => void;
}
