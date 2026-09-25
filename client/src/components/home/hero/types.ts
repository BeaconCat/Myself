/** 首页 Hero 轮播共享类型 */

export interface HeroItem {
  title: string;
  excerpt: string;
  /** 头图 1–3 张：立体相册逐张轮转，放完切下一条；`css:<kind>` / 无图占位渲染为 CSS 光影封面 */
  covers: string[];
  tag: string;
  /** 发布日期（已格式化，如「7月4日」；可缺省） */
  date?: string;
  /** 阅读全文跳转目标 */
  slug?: string;
}
