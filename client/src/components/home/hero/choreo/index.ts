/**
 * Hero 编舞注册表：文字、卡组、组内切换三个独立注册表，任意搭配。
 * id 与 types.ts 的 TextChoreoId / CardChoreoId / RotateChoreoId 一一对应。
 */
import { fade } from './engine';
import {
  DEFAULT_CARD_CHOREO,
  DEFAULT_ROTATE_CHOREO,
  DEFAULT_TEXT_CHOREO,
  type CardChoreo,
  type CardChoreoId,
  type ChoreoMeta,
  type RotateChoreo,
  type RotateChoreoId,
  type TextChoreo,
  type TextChoreoId,
} from './types';
import { caption } from './text/caption';
import { dolly as textDolly } from './text/dolly';
import { glow } from './text/glow';
import { lightscan } from './text/lightscan';
import { shift } from './text/shift';
import { door } from './card/door';
import { dolly as cardDolly } from './card/dolly';
import { hinge } from './card/hinge';
import { parallax } from './card/parallax';
import { shared } from './card/shared';
import { lift } from './rotate/lift';
import { recede } from './rotate/recede';
import { slide } from './rotate/slide';

const TEXT: Record<TextChoreoId, TextChoreo> = {
  lightscan,
  dolly: textDolly,
  shift,
  caption,
  glow,
};

const CARD: Record<CardChoreoId, CardChoreo> = {
  hinge,
  dolly: cardDolly,
  shared,
  parallax,
  door,
};

const ROTATE: Record<RotateChoreoId, RotateChoreo> = {
  lift,
  recede,
  slide,
};

export const TEXT_CHOREOS: ChoreoMeta[] = Object.values(TEXT).map((c) => c.meta);
export const CARD_CHOREOS: ChoreoMeta[] = Object.values(CARD).map((c) => c.meta);
export const ROTATE_CHOREOS: ChoreoMeta[] = Object.values(ROTATE).map((c) => c.meta);

export function isRotateChoreoId(id: unknown): id is RotateChoreoId {
  return typeof id === 'string' && id in ROTATE;
}

/** 按 id 取组内切换实现；未知 id 回落默认 */
export function getRotateChoreo(id: unknown): RotateChoreo {
  return ROTATE[isRotateChoreoId(id) ? id : DEFAULT_ROTATE_CHOREO];
}

export function isTextChoreoId(id: unknown): id is TextChoreoId {
  return typeof id === 'string' && id in TEXT;
}

export function isCardChoreoId(id: unknown): id is CardChoreoId {
  return typeof id === 'string' && id in CARD;
}

/** 按 id 取文字动效实现；未知 id 回落默认 */
export function getTextChoreo(id: unknown): TextChoreo {
  return TEXT[isTextChoreoId(id) ? id : DEFAULT_TEXT_CHOREO];
}

/** 按 id 取卡组动效实现；未知 id 回落默认 */
export function getCardChoreo(id: unknown): CardChoreo {
  return CARD[isCardChoreoId(id) ? id : DEFAULT_CARD_CHOREO];
}

/** reduced-motion 统一降级：旧层 200ms 淡出 + 新层 260ms 淡入（重叠 120ms，共 ~340ms） */
export const REDUCED_TEXT: TextChoreo = {
  meta: { id: DEFAULT_TEXT_CHOREO, name: '', desc: '', tag: '', duration: 340 },
  exit(e, t) {
    t.a(e.root, fade(1, 0), { dur: 200, ease: 'linear' });
  },
  enter(e, t) {
    t.a(e.root, fade(0, 1), { dur: 260, delay: 80, ease: 'linear' });
  },
};

export const REDUCED_CARD: CardChoreo = {
  meta: { id: DEFAULT_CARD_CHOREO, name: '', desc: '', tag: '', duration: 340 },
  exit(e, t) {
    t.a(e.root, fade(1, 0), { dur: 200, ease: 'linear' });
  },
  enter(e, t) {
    t.a(e.root, fade(0, 1), { dur: 260, delay: 80, ease: 'linear' });
  },
};

export * from './types';
