<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { GuestbookData, GuestNote } from '../types';
import type { ModProps } from './props';
import KitIcon from '../parts/KitIcon.vue';
import ModHead from '../parts/ModHead.vue';
import { toast } from '../toast';

/**
 * 留言墙（guestbook）：输入框、留言卡、站长回复、喜欢。
 * 留言后端随 P4 用户/评论系统接入；当前展示配置中的留言，访客提交后本地置顶并提示「审核后公开」。
 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as GuestbookData);
const { t } = useI18n();

const MAX = 120;
const draft = ref('');
const local = ref<(GuestNote & { fresh?: boolean })[]>([]);
const liked = ref<Set<string>>(new Set());

const notes = computed(() => [...local.value, ...d.value.items].slice(0, Math.max(d.value.pageSize, local.value.length + 1)));
const total = computed(() => (d.value.total ?? d.value.items.length) + local.value.length);

function submit(): void {
  const text = draft.value.trim();
  if (!text) return;
  local.value.unshift({ name: t('aboutKit.guestbook.visitor'), color: 'var(--primary)', at: t('aboutKit.guestbook.justNow'), text, likes: 0, fresh: true });
  draft.value = '';
  toast(t('aboutKit.guestbook.submitted'));
}

const keyOf = (n: GuestNote, i: number) => `${i}-${n.name}-${n.at}`;
function like(n: GuestNote, i: number): void {
  const k = keyOf(n, i);
  const next = new Set(liked.value);
  if (next.has(k)) next.delete(k);
  else next.add(k);
  liked.value = next;
}
</script>

<template>
  <ModHead :title="title">{{ t('aboutKit.guestbook.count', { n: total }) }}</ModHead>
  <form class="gb-form" @submit.prevent="submit">
    <span class="av" style="--c: var(--primary)">{{ t('aboutKit.guestbook.visitorShort') }}</span>
    <input v-model="draft" :maxlength="MAX" :placeholder="t('aboutKit.guestbook.placeholder')" :aria-label="t('aboutKit.guestbook.placeholder')" />
    <span class="cnt">{{ draft.length }}/{{ MAX }}</span>
    <button class="ak-btn pri send" type="submit"><KitIcon name="send" :size="15" />{{ t('aboutKit.guestbook.send') }}</button>
  </form>
  <div class="gb-meta">
    <span>{{ d.requireLogin ? t('aboutKit.guestbook.needLogin') : t('aboutKit.guestbook.asVisitor') }}</span>
    <span>{{ t('aboutKit.guestbook.latest') }}</span>
  </div>
  <div class="gb-wall">
    <article v-for="(n, i) in notes" :key="keyOf(n, i)" class="note" :class="{ fresh: (n as { fresh?: boolean }).fresh }">
      <header>
        <span class="av" :style="{ '--c': n.color || 'var(--primary)' }">{{ n.name.slice(0, 1) }}</span>
        <b>{{ n.name }}</b>
        <time>{{ n.at }}</time>
      </header>
      <p>{{ n.text }}</p>
      <div v-if="n.reply" class="re"><b>{{ t('aboutKit.guestbook.owner') }}</b>{{ n.reply }}</div>
      <button class="lk" :class="{ on: liked.has(keyOf(n, i)) }" @click="like(n, i)">
        <KitIcon name="heart" :size="13" /><span>{{ n.likes + (liked.has(keyOf(n, i)) ? 1 : 0) }}</span>
      </button>
    </article>
  </div>
</template>

<style scoped lang="scss">
.gb-form {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px;
  border-radius: 15px;
  background: var(--ak-sunken);
  border: 1px solid var(--ak-line);
  transition: border-color var(--dur), box-shadow var(--dur);

  &:focus-within { border-color: rgba(var(--primary-rgb), 0.55); box-shadow: 0 0 0 4px rgba(var(--primary-rgb), 0.12); }

  .av { width: 34px; height: 34px; }

  input {
    flex: 1;
    min-width: 0;
    background: none;
    border: 0;
    outline: 0;
    font: inherit;
    font-size: 14px;
    color: var(--text);

    &::placeholder { color: var(--ak-text-3); }
  }

  .cnt { font: 400 11px var(--ak-mono); color: var(--ak-text-3); }
  .send { height: 34px; }
}

.gb-meta { display: flex; justify-content: space-between; margin: 10px 2px 18px; font-size: 12px; color: var(--ak-text-3); }

.gb-wall { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }

.note {
  padding: 14px 16px;
  border-radius: 14px;
  background: var(--ak-sunken);
  border: 1px solid var(--ak-line);
  transition: transform var(--dur) var(--ease-out), border-color var(--dur);

  &:hover { border-color: var(--ak-line-2); transform: translateY(-2px); }
  &.fresh { animation: gb-pop 0.6s var(--ease-spring); }

  header { display: flex; align-items: center; gap: 10px; margin-bottom: 8px; }
  header b { font-size: 13px; font-weight: 600; }
  time { margin-left: auto; font: 400 11px var(--ak-mono); color: var(--ak-text-3); }
  p { font-size: 13.5px; line-height: 1.75; color: var(--text-2); overflow-wrap: anywhere; }

  .re {
    margin-top: 10px;
    padding: 8px 10px;
    border-radius: 10px;
    background: var(--ak-surface);
    border: 1px solid var(--ak-line);
    font-size: 12.5px;
    color: var(--text-2);

    b { margin-right: 6px; font-size: 12px; color: var(--ak-ink); }
  }

  .lk {
    display: flex;
    align-items: center;
    gap: 5px;
    margin-top: 10px;
    font: 400 11px var(--ak-mono);
    color: var(--ak-text-3);
    transition: color var(--dur-fast);

    &:hover, &.on { color: var(--ak-red); }
    &.on svg { fill: currentColor; }
  }
}

@keyframes gb-pop { from { opacity: 0; transform: translateY(-14px) scale(0.96); } }

.av {
  display: grid;
  place-items: center;
  flex: none;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  font: 600 12px var(--font-sans);
  color: #fff;
  background: var(--c);
}

@container (max-width: 560px) {
  .gb-wall { grid-template-columns: 1fr; }
  .gb-wall .note:nth-child(n + 4) { display: none; }
  .gb-form .cnt { display: none; }
}
</style>
