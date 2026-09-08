/** 首页 Hero 轮播共享类型 */

export interface HeroItem {
  title: string;
  excerpt: string;
  /** 头图 1–3 张：立体相册逐张轮转，放完切下一条 */
  covers: string[];
  tag: string;
  /** 阅读全文跳转目标 */
  slug?: string;
}

/** enter：卡片入场动画；idle：静止（slot 换位过渡）；out：退场 */
export type HeroPhase = 'enter' | 'idle' | 'out';
