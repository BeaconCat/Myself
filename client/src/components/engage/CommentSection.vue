<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { accountApi, type CommentItem, type CommentTarget, type EngageTarget } from '../../api';
import { useAuthStore } from '../../stores/auth';
import { useConfigStore } from '../../stores/config';
import { useEngageStore } from '../../stores/engage';
import { relTime } from './time';

/**
 * 评论区（文章 / 随想 / 留言墙共用）：一层回复的楼层列表 + 发表框。
 * - 已登录：以账号身份发表；未登录且站点允许匿名：填昵称发表（一律先审）；否则引导登录 / 注册。
 * - 自己待审的评论对自己可见，并标「审核中」。正文纯文本，按行渲染（不解析 HTML / Markdown）。
 * - 蜜罐字段 website 对真人不可见，填了即视为机器。
 */
const props = defineProps<{ target: CommentTarget; commentKey: string; engageId?: number }>();
const { t } = useI18n();
const route = useRoute();
const auth = useAuthStore();
const config = useConfigStore();
const engage = useEngageStore();

const on = computed(() => !!config.cfg.users?.comments?.enabled);
const anonymous = computed(() => !!config.cfg.users?.comments?.anonymous);
const signupOpen = computed(() => config.cfg.users?.readers?.enabled && config.cfg.users?.readers?.signup === 'open');
const canWrite = computed(() => auth.loggedIn || anonymous.value);

const items = ref<CommentItem[]>([]);
const loading = ref(true);
const roots = computed(() => items.value.filter((c) => !c.parentId));
const repliesOf = (id: number) => items.value.filter((c) => c.parentId === id);
const count = computed(() => items.value.filter((c) => !c.pending).length);
/** 刚提交了一条待审评论（访客看不到自己的待审评论）：不再显示「还没有评论」 */
const pendingSent = ref(false);

async function load(): Promise<void> {
  if (!on.value) return;
  loading.value = true;
  try {
    items.value = (await accountApi.comments(props.target, props.commentKey)).items;
  } catch {
    items.value = [];
  } finally {
    loading.value = false;
  }
}

/* ---------- 发表 ---------- */
const body = ref('');
const guestName = ref(localStorage.getItem('myself.guestName') ?? '');
const website = ref('');
const replyTo = ref<CommentItem | null>(null);
const busy = ref(false);
const notice = ref<{ ok: boolean; text: string } | null>(null);
const box = ref<HTMLTextAreaElement | null>(null);

function reply(c: CommentItem): void {
  replyTo.value = c;
  box.value?.focus();
}

const ERR: Record<string, string> = {
  too_many_comments: 'comments.errTooMany',
  invalid_body: 'comments.errBody',
  invalid_name: 'comments.errName',
  login_required: 'comments.errLogin',
  comments_closed: 'comments.errClosed',
};

async function submit(): Promise<void> {
  if (busy.value || !body.value.trim()) return;
  if (!auth.loggedIn && !guestName.value.trim()) {
    notice.value = { ok: false, text: t('comments.errName') };
    return;
  }
  busy.value = true;
  notice.value = null;
  try {
    const res = await accountApi.postComment({
      target: props.target,
      key: props.commentKey,
      body: body.value,
      parentId: replyTo.value?.id,
      guestName: auth.loggedIn ? undefined : guestName.value.trim(),
      website: website.value,
    });
    if (!auth.loggedIn) localStorage.setItem('myself.guestName', guestName.value.trim());
    body.value = '';
    replyTo.value = null;
    notice.value = { ok: true, text: res.pending ? t('comments.pending') : t('comments.posted') };
    if (res.pending) pendingSent.value = true;
    if (!res.pending && props.engageId && props.target !== 'guestbook') {
      engage.bumpComments(props.target as EngageTarget, props.engageId);
    }
    await load();
  } catch (e) {
    notice.value = { ok: false, text: t(ERR[(e as Error).message] ?? 'comments.errGeneric') };
  } finally {
    busy.value = false;
  }
}

const loginLink = computed(() => ({ path: '/account/login', query: { next: route.fullPath } }));
const registerLink = computed(() => ({ path: '/account/register', query: { next: route.fullPath } }));
const initial = (name: string) => [...(name || '?')][0];

onMounted(load);
watch(() => props.commentKey, () => {
  pendingSent.value = false;
  void load();
});
</script>

<template>
  <section v-if="on" id="comments" class="cs">
    <header class="cs-hd">
      <h2>{{ t('comments.title') }}</h2>
      <span v-if="count" class="n">{{ count }}</span>
    </header>

    <!-- 发表框 -->
    <form v-if="canWrite" class="cs-form" @submit.prevent="submit">
      <div v-if="replyTo" class="replying">
        {{ t('comments.replyingTo', { name: replyTo.author.name }) }}
        <button type="button" @click="replyTo = null">{{ t('comments.cancelReply') }}</button>
      </div>
      <input v-if="!auth.loggedIn" v-model="guestName" class="name" maxlength="24" :placeholder="t('comments.guestName')" autocomplete="nickname" />
      <!-- 蜜罐：对真人隐藏 -->
      <input v-model="website" class="hp" tabindex="-1" autocomplete="off" aria-hidden="true" name="website" />
      <textarea
        ref="box"
        v-model="body"
        rows="3"
        maxlength="2000"
        :placeholder="auth.loggedIn ? t('comments.placeholder') : t('comments.placeholderGuest')"
        @keydown.ctrl.enter="submit"
        @keydown.meta.enter="submit"
      />
      <div class="ft">
        <span class="who">
          <template v-if="auth.loggedIn && auth.user">{{ t('comments.as', { name: auth.user.name }) }}</template>
          <template v-else>{{ t('comments.guestHint') }}</template>
        </span>
        <Transition name="nt"><span v-if="notice" class="notice" :class="{ ok: notice.ok }">{{ notice.text }}</span></Transition>
        <button type="submit" class="send" :disabled="busy || !body.trim()">{{ busy ? t('comments.sending') : t('comments.send') }}</button>
      </div>
    </form>
    <div v-else class="cs-login">
      <span>{{ t('comments.loginPrompt') }}</span>
      <router-link :to="loginLink" class="primary">{{ t('comments.login') }}</router-link>
      <router-link v-if="signupOpen" :to="registerLink">{{ t('comments.register') }}</router-link>
    </div>

    <!-- 楼层 -->
    <p v-if="!loading && !items.length && !pendingSent" class="empty">{{ t('comments.empty') }}</p>
    <ol class="cs-list">
      <li v-for="c in roots" :key="c.id" class="c">
        <div class="row">
          <span class="av" :class="c.author.role">
            <img v-if="c.author.avatar" :src="c.author.avatar" alt="" referrerpolicy="no-referrer" />
            <template v-else>{{ initial(c.author.name) }}</template>
          </span>
          <div class="main">
            <div class="meta">
              <b>{{ c.author.name }}</b>
              <span v-if="c.author.role === 'admin'" class="badge">{{ t('comments.owner') }}</span>
              <span v-else-if="c.author.role === 'author'" class="badge soft">{{ t('comments.author') }}</span>
              <time>{{ relTime(c.createdAt) }}</time>
              <span v-if="c.pending" class="badge wait">{{ t('comments.reviewing') }}</span>
            </div>
            <p class="text">{{ c.body }}</p>
            <button v-if="canWrite && !c.pending" type="button" class="rp" @click="reply(c)">{{ t('comments.reply') }}</button>
            <ol v-if="repliesOf(c.id).length" class="replies">
              <li v-for="r in repliesOf(c.id)" :key="r.id" class="row">
                <span class="av sm" :class="r.author.role">
                  <img v-if="r.author.avatar" :src="r.author.avatar" alt="" referrerpolicy="no-referrer" />
                  <template v-else>{{ initial(r.author.name) }}</template>
                </span>
                <div class="main">
                  <div class="meta">
                    <b>{{ r.author.name }}</b>
                    <span v-if="r.author.role === 'admin'" class="badge">{{ t('comments.owner') }}</span>
                    <time>{{ relTime(r.createdAt) }}</time>
                    <span v-if="r.pending" class="badge wait">{{ t('comments.reviewing') }}</span>
                  </div>
                  <p class="text">{{ r.body }}</p>
                </div>
              </li>
            </ol>
          </div>
        </div>
      </li>
    </ol>
  </section>
</template>

<style scoped lang="scss">
.cs {
  margin-top: 48px;
  scroll-margin-top: 96px;
}

.cs-hd {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin-bottom: 18px;

  h2 { margin: 0; font: 700 24px/1.3 var(--font-serif); }
  .n { font: 600 15px var(--font-mono); color: var(--text-3); }
}

/* 发表框：下沉面，聚焦时描边轻染 */
.cs-form {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  border-radius: var(--r-lg);
  background: var(--fill);
  box-shadow: inset 0 0 0 1px var(--line);
  transition: box-shadow var(--dur-fast), background var(--dur-fast);

  &:focus-within { background: var(--bg); box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--ink) 45%, transparent), 0 0 0 3px color-mix(in oklab, var(--ink) 12%, transparent); }

  input.name, textarea {
    width: 100%;
    border: 0;
    outline: 0;
    background: none;
    color: var(--text);
    font: 15px/1.7 var(--font-sans);
  }

  input.name { max-width: 260px; padding: 2px 0 8px; border-bottom: 1px solid var(--line); }
  textarea { resize: vertical; min-height: 72px; }
  .hp { position: absolute; left: -9999px; width: 1px; height: 1px; opacity: 0; }

  .ft { display: flex; align-items: center; gap: 12px; }
  .who { flex: 1; font-size: 12.5px; color: var(--text-3); }
  .notice { font-size: 13px; color: var(--accent-red); }
  .notice.ok { color: var(--ink); }

  .send {
    height: 36px;
    padding: 0 18px;
    border: 0;
    border-radius: var(--r-pill);
    background: var(--solid);
    color: var(--on-solid);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: background var(--dur-fast), opacity var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

    &:hover:not(:disabled) { background: var(--solid-hover, var(--solid)); }
    &:active:not(:disabled) { transform: scale(0.97); }
    &:disabled { opacity: 0.5; cursor: not-allowed; }
  }
}

.replying {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-2);

  button { border: 0; background: none; color: var(--ink); cursor: pointer; font-size: 13px; }
}

.cs-login {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 18px;
  border-radius: var(--r-lg);
  background: var(--fill);
  font-size: 14px;
  color: var(--text-2);

  span { flex: 1; }
  a { padding: 7px 14px; border-radius: var(--r-pill); color: var(--text); box-shadow: inset 0 0 0 1px var(--line-2); font-size: 13.5px; }
  a.primary { background: var(--solid); color: var(--on-solid); box-shadow: none; }
}

.empty { margin: 22px 0 0; font-size: 14px; color: var(--text-3); }

.cs-list, .replies { list-style: none; margin: 0; padding: 0; }
.cs-list { margin-top: 10px; }

.c { padding: 16px 0; border-bottom: 0.5px solid var(--line-2); }
.row { display: flex; gap: 12px; }
.replies { margin-top: 12px; display: flex; flex-direction: column; gap: 12px; }
.main { flex: 1; min-width: 0; }

.av {
  flex: none;
  width: 36px;
  height: 36px;
  display: grid;
  place-items: center;
  overflow: hidden;
  border-radius: 50%;
  background: var(--fill-2, var(--fill));
  color: var(--text-2);
  font: 600 14px var(--font-serif);

  img { width: 100%; height: 100%; object-fit: cover; }
  &.admin { box-shadow: 0 0 0 1.5px var(--ink); }
  &.sm { width: 28px; height: 28px; font-size: 12px; }
}

.meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 13px;
  color: var(--text-3);

  b { font-size: 14.5px; font-weight: 600; color: var(--text); }
}

.badge {
  padding: 1px 7px;
  border-radius: var(--r-xs);
  font-size: 11.5px;
  background: color-mix(in oklab, var(--ink) 12%, transparent);
  color: var(--ink);

  &.soft { background: var(--fill); color: var(--text-2); }
  &.wait { background: color-mix(in oklab, var(--accent-yellow) 16%, transparent); color: color-mix(in oklab, var(--accent-yellow) 55%, var(--text)); }
}

.text { margin: 4px 0 0; font-size: 15px; line-height: 1.75; color: var(--text); white-space: pre-wrap; overflow-wrap: anywhere; }

.rp {
  margin-top: 4px;
  padding: 0;
  border: 0;
  background: none;
  font-size: 12.5px;
  color: var(--text-3);
  cursor: pointer;

  &:hover { color: var(--ink); }
}

.nt-enter-active, .nt-leave-active { transition: opacity var(--dur-fast); }
.nt-enter-from, .nt-leave-to { opacity: 0; }
</style>
