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
