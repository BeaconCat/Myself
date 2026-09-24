import { onBeforeUnmount, onMounted, ref, type Ref } from 'vue';

/**
 * 全页共享的秒级时钟：多个模块订阅同一个 interval（引用计数），卸载后自动停止。
 */
const now = ref(Date.now());
let timer = 0;
let users = 0;

export function useClock(): Ref<number> {
  onMounted(() => {
    users += 1;
    if (users === 1) {
      now.value = Date.now();
      timer = window.setInterval(() => { now.value = Date.now(); }, 1000);
    }
  });
  onBeforeUnmount(() => {
    users -= 1;
    if (users <= 0) { users = 0; window.clearInterval(timer); }
  });
  return now;
}

/** 指定 UTC 偏移（小时）下的时间各分量 */
export function zoned(ts: number, tz: number): { h: number; m: number; s: number } {
  const d = new Date(ts + tz * 3600_000);
  return { h: d.getUTCHours(), m: d.getUTCMinutes(), s: d.getUTCSeconds() };
}

const PART_KEYS = ['late', 'late', 'late', 'late', 'late', 'dawn', 'dawn', 'dawn', 'morning', 'morning', 'morning', 'morning', 'noon', 'afternoon', 'afternoon', 'afternoon', 'afternoon', 'dusk', 'dusk', 'evening', 'evening', 'evening', 'evening', 'late'];
/** 时段文案键（aboutKit.dayPart.*：清晨 / 午后 / 夜里…） */
export function dayPartKey(hour: number): string {
  return `aboutKit.dayPart.${PART_KEYS[hour] ?? 'late'}`;
}

export const pad2 = (n: number): string => String(n).padStart(2, '0');

type T = (key: string, named?: Record<string, unknown>) => string;

/** 相对时间：ISO 日期 → 今天 / 昨天 / N 天前…；非日期字符串原样返回 */
export function ago(input: string | undefined, t: T, nowTs = Date.now()): string {
  if (!input) return '';
  const ts = new Date(input).getTime();
  if (!Number.isFinite(ts)) return input;
  const n = Math.floor((nowTs - ts) / 864e5);
  if (n <= 0) return t('aboutKit.ago.today');
  if (n === 1) return t('aboutKit.ago.yesterday');
  if (n < 30) return t('aboutKit.ago.days', { n });
  if (n < 365) return t('aboutKit.ago.months', { n: Math.floor(n / 30) });
  return t('aboutKit.ago.years', { n: Math.floor(n / 365) });
}
