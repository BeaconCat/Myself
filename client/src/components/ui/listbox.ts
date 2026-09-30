/**
 * 下拉选择 / 组合框共用：选项类型 + 模糊匹配 + 高亮切片。
 * 匹配规则（不区分大小写）：先找连续子串（越靠前分越高），找不到再按字符顺序做子序列匹配（分数更低）。
 */

export interface UiOption {
  value: string;
  label: string;
  /** 右侧弱化小字（如英文名、经纬度） */
  hint?: string;
  /** 额外参与搜索的关键词（别名、拼音缩写等），不显示 */
  keywords?: string;
  disabled?: boolean;
}

export interface MatchedOption {
  option: UiOption;
  /** label 中命中的字符下标（用于高亮） */
  hits: number[];
  score: number;
}

/** 单串匹配：返回命中下标与分数；不匹配返回 null */
export function fuzzyMatch(text: string, query: string): { hits: number[]; score: number } | null {
  const q = query.trim().toLowerCase();
  if (!q) return { hits: [], score: 0 };
  const s = text.toLowerCase();
  const at = s.indexOf(q);
  if (at >= 0) {
    const hits = Array.from({ length: q.length }, (_, i) => at + i);
    // 完全相等 > 前缀 > 词首 > 其他位置；短文本略优先
    const exact = s === q ? 200 : 0;
    const prefix = at === 0 ? 100 : /[\s\-_./·]/.test(s[at - 1]) ? 60 : 0;
    return { hits, score: 1000 + exact + prefix - at - s.length * 0.1 };
  }
  // 子序列：字符按顺序出现即可，连续命中加分
  const hits: number[] = [];
  let from = 0;
  let run = 0;
  let score = 0;
  for (const ch of q) {
    if (ch === ' ') continue;
    const i = s.indexOf(ch, from);
    if (i < 0) return null;
    run = hits.length && i === hits[hits.length - 1] + 1 ? run + 1 : 0;
    score += 10 + run * 5 - (i - from);
    hits.push(i);
    from = i + 1;
  }
  return { hits, score };
}

/** 过滤并排序选项：label 命中优先，其次 hint / keywords 命中（此时 label 不高亮） */
export function filterOptions(options: UiOption[], query: string, limit = Infinity): MatchedOption[] {
  if (!query.trim()) return options.slice(0, limit).map((option) => ({ option, hits: [], score: 0 }));
  const out: MatchedOption[] = [];
  for (const option of options) {
    const m = fuzzyMatch(option.label, query);
    if (m) {
      out.push({ option, hits: m.hits, score: m.score + 50 });
      continue;
    }
    const extra = fuzzyMatch(`${option.value} ${option.hint ?? ''} ${option.keywords ?? ''}`, query);
    if (extra) out.push({ option, hits: [], score: extra.score });
  }
  return out.sort((a, b) => b.score - a.score).slice(0, limit);
}

/** 把 label 切成 [文本, 是否命中] 片段，模板里逐段渲染（不用 v-html） */
export function highlightParts(text: string, hits: number[]): { text: string; hit: boolean }[] {
  if (!hits.length) return [{ text, hit: false }];
  const set = new Set(hits);
  const parts: { text: string; hit: boolean }[] = [];
  for (let i = 0; i < text.length; i++) {
    const hit = set.has(i);
    const last = parts[parts.length - 1];
    if (last && last.hit === hit) last.text += text[i];
    else parts.push({ text: text[i], hit });
  }
  return parts;
}
