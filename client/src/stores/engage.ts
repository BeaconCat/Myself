import { defineStore } from 'pinia';
import { api, type EngageSummary, type EngageTarget, type ReactionKind } from '../api';

/**
 * 互动摘要缓存（随想 / 文章的回应计数、我点过的回应、评论数）。
 * 组件挂载时 want(target, id) 登记，同一轮事件循环内的登记合并成一次批量请求。
 */
const key = (t: EngageTarget, id: number) => `${t}:${id}`;
const pending: Record<EngageTarget, Set<number>> = { note: new Set(), post: new Set() };
let flushQueued = false;

export const useEngageStore = defineStore('engage', {
  state: () => ({
    map: {} as Record<string, EngageSummary>,
  }),
  actions: {
    get(target: EngageTarget, id: number): EngageSummary {
      return this.map[key(target, id)] ?? { reactions: {}, mine: [], comments: 0 };
    },
    want(target: EngageTarget, id: number) {
      if (this.map[key(target, id)]) return;
      pending[target].add(id);
      if (flushQueued) return;
      flushQueued = true;
      queueMicrotask(() => void this.flush());
    },
    async flush() {
      flushQueued = false;
      for (const target of ['note', 'post'] as EngageTarget[]) {
        const ids = [...pending[target]];
        pending[target].clear();
        for (let i = 0; i < ids.length; i += 100) {
          const chunk = ids.slice(i, i + 100);
          try {
            const res = await api.engage(target, chunk);
            for (const id of chunk) {
              this.map[key(target, id)] = res[String(id)] ?? { reactions: {}, mine: [], comments: 0 };
            }
          } catch { /* 互动摘要失败不影响正文 */ }
        }
      }
    },
    /** 乐观切换：先改本地，失败回滚 */
    async toggle(target: EngageTarget, id: number, kind: ReactionKind) {
      const k = key(target, id);
      const prev = this.get(target, id);
      const had = prev.mine.includes(kind);
      this.map[k] = {
        ...prev,
        reactions: { ...prev.reactions, [kind]: Math.max(0, (prev.reactions[kind] ?? 0) + (had ? -1 : 1)) },
        mine: had ? prev.mine.filter((m) => m !== kind) : [...prev.mine, kind],
      };
      try {
        this.map[k] = await api.react(target, id, kind);
      } catch (e) {
        this.map[k] = prev;
        throw e;
      }
    },
    /** 发表评论后本地 +1（审核中的不计） */
    bumpComments(target: EngageTarget, id: number) {
      const cur = this.get(target, id);
      this.map[key(target, id)] = { ...cur, comments: cur.comments + 1 };
    },
  },
});
