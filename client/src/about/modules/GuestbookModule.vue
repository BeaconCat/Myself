<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';
import { accountApi, type CommentItem } from '../../api';
import { useAuthStore } from '../../stores/auth';
import { useConfigStore } from '../../stores/config';
import { relTime } from '../../components/engage/time';
import type { GuestbookData } from '../types';
import type { ModProps } from './props';
import KitIcon from '../parts/KitIcon.vue';
import ModHead from '../parts/ModHead.vue';
import { toast } from '../toast';

/**
 * 留言墙（guestbook）：真实留言（评论系统 target=guestbook）。输入框、留言卡、站长回复（取第一条站长的回复）。
 * 发表规则跟随「用户」页：已登录直接留言；开了匿名时访客填昵称即可；否则引导登录。审核中的留言只有自己看得到。
 * 模块只存展示选项（每页条数）；Demo 数据在初始化时写入为真实留言。
 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as GuestbookData);
const { t } = useI18n();
const route = useRoute();
const auth = useAuthStore();
const config = useConfigStore();

const MAX = 120;
const draft = ref('');
const guestName = ref('');
const website = ref('');
const busy = ref(false);
const enabled = ref(true);
const items = ref<CommentItem[]>([]);
const loaded = ref(false);
const expanded = ref(false);
const fresh = ref<number | null>(null);

const anonymous = computed(() => !!config.cfg.users?.comments.anonymous);
const canWrite = computed(() => enabled.value && (auth.loggedIn || anonymous.value));

async function load(): Promise<void> {
  try {
    const res = await accountApi.comments('guestbook', '');
    enabled.value = res.enabled;
    items.value = res.items;
  } catch {
    enabled.value = false;
  }
  loaded.value = true;
}
onMounted(load);

/** 访客刚发、仍在审核的留言：服务端只给登录用户回显自己的待审留言，访客的在本地保留（刷新即消失） */
const mine = ref<CommentItem[]>([]);
const roots = computed(() => {
  const known = new Set(items.value.map((c) => c.id));
  return [...mine.value.filter((c) => !known.has(c.id)), ...items.value.filter((c) => !c.parentId)];
});
const replyOf = (id: number): CommentItem | undefined => items.value.find((c) => c.parentId === id && c.author.role === 'admin');
const pageSize = computed(() => Math.max(2, Number(d.value.pageSize) || 4));
const shown = computed(() => (expanded.value ? roots.value : roots.value.slice(0, pageSize.value)));

async function submit(): Promise<void> {
  const body = draft.value.trim();
  if (!body || busy.value || !canWrite.value) return;
  if (!auth.loggedIn && !guestName.value.trim()) {
    toast(t('aboutKit.guestbook.needName'));
    return;
  }
  busy.value = true;
  try {
    const res = await accountApi.postComment({ target: 'guestbook', key: '', body, guestName: auth.loggedIn ? undefined : guestName.value.trim(), website: website.value });
    draft.value = '';
    if (res.pending && !auth.loggedIn) {
      const now = new Date().toISOString().slice(0, 19).replace('T', ' ');
      mine.value.unshift({ id: res.id, body, createdAt: now, pending: true, author: { name: guestName.value.trim(), avatar: '', role: 'guest' } });
    }
    await load();
    fresh.value = res.id;
    toast(res.pending ? t('aboutKit.guestbook.submitted') : t('aboutKit.guestbook.posted'));
  } catch (e) {
    const code = (e as Error).message;
    toast(code === 'too_many_comments' ? t('comments.errTooMany') : code === 'invalid_name' ? t('aboutKit.guestbook.needName') : t('comments.errGeneric'));
  } finally {
    busy.value = false;
  }
}

const PALETTE = ['#0078ff', '#00c853', '#ffb300', '#ff0032', '#8b5cf6', '#0ea5a4', '#f97316'];
const tint = (s: string): string => PALETTE[[...s].reduce((a, ch) => a + ch.charCodeAt(0), 0) % PALETTE.length];
const loginTo = computed(() => ({ path: '/account/login', query: { next: route.fullPath } }));
</script>

<template>
  <ModHead :title="title">{{ t('aboutKit.guestbook.count', { n: roots.filter((c) => !c.pending).length }) }}</ModHead>

  <form v-if="canWrite" class="gb-form" @submit.prevent="submit">
    <span class="av me">
      <img v-if="auth.user?.avatar" :src="auth.user.avatar" alt="" referrerpolicy="no-referrer" />
      <template v-else>{{ (auth.user?.name || t('aboutKit.guestbook.visitorShort')).slice(0, 1) }}</template>
    </span>
    <input v-if="!auth.loggedIn" v-model="guestName" class="nm" maxlength="24" :placeholder="t('aboutKit.guestbook.namePh')" autocomplete="nickname" />
    <input v-model="draft" :maxlength="MAX" :placeholder="t('aboutKit.guestbook.placeholder')" :aria-label="t('aboutKit.guestbook.placeholder')" />
    <!-- 蜜罐：对真人隐藏 -->
    <input v-model="website" class="hp" tabindex="-1" autocomplete="off" aria-hidden="true" name="website" />
    <span class="cnt">{{ draft.length }}/{{ MAX }}</span>
    <button class="ak-btn pri send" type="submit" :disabled="busy || !draft.trim()"><KitIcon name="send" :size="16" />{{ t('aboutKit.guestbook.send') }}</button>
  </form>
  <div v-else-if="loaded" class="gb-closed">
    <template v-if="enabled">
      <span>{{ t('aboutKit.guestbook.needLogin') }}</span>
      <router-link class="ak-btn pri" :to="loginTo">{{ t('comments.login') }}</router-link>
    </template>
    <span v-else>{{ t('aboutKit.guestbook.closed') }}</span>
  </div>

  <div v-if="canWrite" class="gb-meta">
    <span>{{ auth.loggedIn ? t('aboutKit.guestbook.asUser', { name: auth.user?.name ?? '' }) : t('aboutKit.guestbook.asVisitor') }}</span>
    <span>{{ t('aboutKit.guestbook.latest') }}</span>
  </div>

  <p v-if="loaded && !roots.length" class="gb-empty">{{ t('aboutKit.guestbook.empty') }}</p>
  <TransitionGroup tag="div" name="gb" class="gb-wall">
    <article v-for="n in shown" :key="n.id" class="note" :class="{ fresh: fresh === n.id }">
      <header>
        <span class="av" :style="{ '--c': tint(n.author.name) }">
          <img v-if="n.author.avatar" :src="n.author.avatar" alt="" referrerpolicy="no-referrer" />
          <template v-else>{{ n.author.name.slice(0, 1) }}</template>
        </span>
        <b>{{ n.author.name }}</b>
        <span v-if="n.pending" class="wait">{{ t('comments.reviewing') }}</span>
        <time>{{ relTime(n.createdAt) }}</time>
      </header>
      <p>{{ n.body }}</p>
      <div v-if="replyOf(n.id)" class="re"><b>{{ t('aboutKit.guestbook.owner') }}</b>{{ replyOf(n.id)!.body }}</div>
    </article>
  </TransitionGroup>
  <button v-if="roots.length > pageSize" type="button" class="gb-more" @click="expanded = !expanded">
    {{ expanded ? t('aboutKit.guestbook.less') : t('aboutKit.guestbook.more', { n: roots.length - pageSize }) }}
  </button>
</template>

<style scoped lang="scss">
.gb-form {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px;
  border-radius: var(--r-pill);
  background: var(--ak-sunken);
  box-shadow: inset 0 0 0 1px var(--ak-line);
  transition: background-color var(--dur-fast), box-shadow var(--dur-fast);

  &:hover { box-shadow: inset 0 0 0 1px var(--ak-line-2); }
  /* 聚焦：焦点环（--ink）是唯一允许的彩色环 */
  &:focus-within { background: var(--elev); box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--ink) 70%, transparent), 0 0 0 3px color-mix(in oklab, var(--ink) 18%, transparent); }

  .av { width: 36px; height: 36px; }

  input {
    flex: 1;
    min-width: 0;
    background: none;
    border: 0;
    outline: 0;
    font: inherit;
    font-size: 15px;
    color: var(--text);

    &::placeholder { color: var(--ak-text-3); }
  }

  input.nm {
    flex: 0 0 96px;
    padding-right: 10px;
    border-right: 1px solid var(--ak-line);
  }

  .hp { position: absolute; left: -9999px; width: 1px; height: 1px; opacity: 0; flex: none; }
  .cnt { font: 400 12px var(--ak-mono); color: var(--ak-text-3); }
  .send { height: 36px; }
  .send:disabled { opacity: 0.55; cursor: not-allowed; }
}

.gb-closed {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 10px 10px 18px;
  border-radius: var(--r-pill);
  background: var(--ak-sunken);
  font-size: 14px;
  color: var(--ak-text-3);

  .ak-btn { height: 36px; }
}

.gb-meta { display: flex; justify-content: space-between; margin: 10px 2px 16px; font-size: 13px; color: var(--ak-text-3); }
.gb-closed + .gb-empty, .gb-closed + .gb-wall { margin-top: 16px; }
.gb-empty { margin: 16px 2px 0; font-size: 14px; color: var(--ak-text-3); }

/* 留言卡：默认两列；通栏宽度足够时四列一排（避免 3 + 1 的孤卡） */
.gb-wall { position: relative; display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }

@container (min-width: 1000px) { .gb-wall { grid-template-columns: repeat(4, minmax(0, 1fr)); } }

.note {
  display: flex;
  flex-direction: column;
  padding: 14px 16px;
  border-radius: var(--r-md);
  background: var(--ak-sunken);
  transition: transform var(--dur) var(--ease-out), background-color var(--dur);

  &:hover { background: var(--fill-2); transform: translateY(-2px); }
  &.fresh { animation: gb-pop 0.6s var(--ease-spring); }

  header { display: flex; align-items: center; gap: 10px; margin-bottom: 8px; min-width: 0; }
  header b { font-size: 14.5px; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  time { margin-left: auto; flex: none; font: 400 12.5px var(--ak-mono); color: var(--ak-text-3); }
  p { font-size: 15px; line-height: 1.7; color: var(--text); overflow-wrap: anywhere; white-space: pre-wrap; }

  .wait {
    flex: none;
    padding: 1px 7px;
    border-radius: var(--r-pill);
    background: color-mix(in oklab, var(--accent-yellow) 16%, transparent);
    color: color-mix(in oklab, var(--accent-yellow) 70%, var(--text));
    font-size: 11.5px;
  }

  .re {
    margin-top: 10px;
    padding: 8px 10px;
    border-radius: var(--r-sm);
    background: var(--ak-surface);
    box-shadow: inset 0 0 0 1px var(--ak-line);
    font-size: 13.5px;
    color: var(--text-2);

    b { margin-right: 6px; font-size: 13px; color: var(--ak-ink); }
  }
}

.gb-more {
  display: block;
  margin: 14px auto 0;
  padding: 6px 14px;
  border-radius: var(--r-pill);
  font-size: 13.5px;
  color: var(--ak-text-3);
  transition: color var(--dur-fast), background-color var(--dur-fast);

  &:hover { color: var(--text); background: var(--ak-sunken); }
}

.gb-enter-active, .gb-leave-active { transition: opacity var(--dur) var(--ease-out), transform var(--dur) var(--ease-out); }
.gb-enter-from, .gb-leave-to { opacity: 0; transform: translateY(8px); }

@keyframes gb-pop { from { opacity: 0; transform: translateY(-14px) scale(0.96); } }

.av {
  display: grid;
  place-items: center;
  flex: none;
  overflow: hidden;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  font: 600 13px var(--font-sans);
  color: #fff;
  background: var(--c);

  img { width: 100%; height: 100%; object-fit: cover; }
  &.me { color: var(--on-solid); background: var(--solid); }
}

@container (max-width: 560px) {
  .gb-wall { grid-template-columns: 1fr; }
  .gb-form .cnt { display: none; }
  .gb-form input.nm { flex-basis: 72px; }
}
</style>
