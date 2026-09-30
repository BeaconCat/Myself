<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { Check, MessageSquareDashed } from 'lucide';
import { accountApi, adminApi, type AdminComment, type CommentStatus } from '../../api';
import { useConfigStore } from '../../stores/config';
import { useDialogStore } from '../../stores/dialog';
import { useAuthStore } from '../../stores/auth';
import Icon from '../../components/ui/Icon.vue';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import EmptyArt from './studio/EmptyArt.vue';
import StSeg from './studio/StSeg.vue';
import { toast } from './studio/toast';
import { initial, plainText, relTime } from './studio/format';

/**
 * 评论审核：待审 / 已通过 / 垃圾 三栏；单条通过、回复（以站长身份，待审的先通过再回复）、
 * 撤回、标为垃圾、删除；多选批量处理。
 */
const { t } = useI18n();
const config = useConfigStore();
const dialog = useDialogStore();
const auth = useAuthStore();
/** 协作作者看到的是自己文章下的评论 */
const descKey = computed(() => (auth.isAdmin ? 'studio.comments.desc' : 'studio.comments.descAuthor'));

type Tab = CommentStatus;
const tab = ref<Tab>('pending');
const lists = reactive<Record<Tab, AdminComment[] | null>>({ pending: null, approved: null, spam: null });
const list = computed(() => lists[tab.value] ?? []);
const loading = computed(() => lists[tab.value] === null);
const commentsOn = computed(() => !!config.cfg.users?.comments.enabled);

async function load(which: Tab): Promise<void> {
  try {
    lists[which] = await adminApi.comments(which);
  } catch {
    lists[which] = [];
    toast(t('studio.loadFailed'), { icon: 'x' });
  }
}
onMounted(() => void Promise.all((['pending', 'approved', 'spam'] as Tab[]).map(load)));

const TABS = computed(() =>
  (['pending', 'approved', 'spam'] as Tab[]).map((v) => ({
    value: v,
    label: t(`studio.comments.${v}`),
    count: lists[v]?.length || undefined,
  })),
);

/* ---------- 选择 ---------- */
const selected = ref(new Set<number>());
const allOn = computed(() => list.value.length > 0 && list.value.every((c) => selected.value.has(c.id)));
function toggle(id: number): void {
  const s = new Set(selected.value);
  if (s.has(id)) s.delete(id);
  else s.add(id);
  selected.value = s;
}
function toggleAll(): void {
  selected.value = allOn.value ? new Set() : new Set(list.value.map((c) => c.id));
}
function setTab(v: Tab): void {
  tab.value = v;
  selected.value = new Set();
  replying.value = null;
}

/* ---------- 操作 ---------- */
/** 把若干条评论从当前栏移到目标栏（本地即时更新，失败时重新拉取） */
function move(ids: number[], to: Tab | null): void {
  const from = tab.value;
  const moved = (lists[from] ?? []).filter((c) => ids.includes(c.id));
  lists[from] = (lists[from] ?? []).filter((c) => !ids.includes(c.id));
  if (to && lists[to]) lists[to] = [...moved.map((c) => ({ ...c, status: to })), ...lists[to]!].sort((a, b) => b.id - a.id);
  const s = new Set(selected.value);
  ids.forEach((id) => s.delete(id));
  selected.value = s;
}

async function setStatus(c: AdminComment, status: CommentStatus): Promise<boolean> {
  try {
    await adminApi.setCommentStatus(c.id, status);
    move([c.id], status);
    toast(t(`studio.comments.moved_${status}`), { icon: status === 'spam' ? 'flag' : 'check' });
    return true;
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
    return false;
  }
}

async function remove(c: AdminComment): Promise<void> {
  const ok = await dialog.confirm({ title: t('studio.comments.deleteTitle'), message: t('studio.comments.deleteBody'), confirmText: t('studio.delete'), danger: true });
  if (!ok) return;
  try {
    await adminApi.deleteComment(c.id);
    move([c.id], null);
    toast(t('studio.deleted'));
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  }
}

async function bulk(action: CommentStatus | 'delete'): Promise<void> {
  const ids = [...selected.value];
  if (!ids.length) return;
  if (action === 'delete') {
    const ok = await dialog.confirm({
      title: t('studio.comments.bulkDeleteTitle', { n: ids.length }),
      message: t('studio.comments.deleteBody'),
      confirmText: t('studio.delete'),
      danger: true,
    });
    if (!ok) return;
  }
  try {
    await adminApi.batchComments(ids, action === 'delete' ? { delete: true } : { status: action });
    move(ids, action === 'delete' ? null : action);
    toast(t('studio.comments.bulkDone', { n: ids.length }));
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
    void load(tab.value);
  }
}

/* ---------- 回复（以站长身份） ---------- */
const replying = ref<number | null>(null);
const replyText = ref('');
const replyBusy = ref(false);
const replyEl = ref<HTMLTextAreaElement[]>([]);

async function openReply(c: AdminComment): Promise<void> {
  replying.value = replying.value === c.id ? null : c.id;
  replyText.value = '';
  await nextTick();
  replyEl.value[0]?.focus();
}

async function sendReply(c: AdminComment): Promise<void> {
  const body = replyText.value.trim();
  if (!body || replyBusy.value) return;
  replyBusy.value = true;
  try {
    // 只能回复已公开的评论：待审的先通过
    if (c.status === 'pending' && !(await setStatus(c, 'approved'))) return;
    await accountApi.postComment({ target: c.target, key: c.targetKey, body, parentId: c.id });
    replying.value = null;
    replyText.value = '';
    toast(t('studio.comments.replied'), { icon: 'reply' });
    void load('approved');
  } catch (e) {
    toast((e as Error).message === 'comments_closed' ? t('studio.comments.closed') : t('studio.saveFailed'), { icon: 'x' });
  } finally {
    replyBusy.value = false;
  }
}

function onReplyKey(e: KeyboardEvent, c: AdminComment): void {
  if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
    e.preventDefault();
    void sendReply(c);
  }
}

const PALETTE = ['#3b7d5a', '#556b8d', '#b5651d', '#9a4f7a', '#6d6a3a', '#2f6f8f', '#8a4b3c'];
const tint = (s: string): string => PALETTE[[...s].reduce((a, ch) => a + ch.charCodeAt(0), 0) % PALETTE.length];
</script>

<template>
  <section class="studio view">
    <div class="st-vh">
      <div>
        <h1>{{ t('studio.comments.title') }}</h1>
        <p>{{ t(descKey) }}</p>
      </div>
      <div class="act">
        <StSeg :model-value="tab" :options="TABS" @update:model-value="setTab" />
      </div>
    </div>

    <div v-if="!commentsOn" class="st-note-bar st-rise">
      <SIcon name="info" />{{ t('studio.comments.off') }}
      <router-link v-if="auth.isAdmin" class="st-link" :to="{ name: 'admin-users' }">{{ t('studio.comments.goSwitch') }}</router-link>
    </div>

    <div v-if="list.length" class="bulk">
      <label class="st-ckrow" @click.prevent="toggleAll">
        <span class="st-ck" :class="{ on: allOn }"><Icon :icon="Check" /></span>{{ t('studio.comments.selectAll') }}
      </label>
      <span v-if="selected.size" class="sel">{{ t('studio.comments.selected', { n: selected.size }) }}</span>
      <span class="sp" />
      <template v-if="selected.size">
        <button v-if="tab !== 'approved'" type="button" class="st-btn q sm" @click="bulk('approved')"><SIcon name="check" :size="16" />{{ t('studio.comments.approve') }}</button>
        <button v-if="tab !== 'spam'" type="button" class="st-btn q sm" @click="bulk('spam')"><SIcon name="flag" :size="16" />{{ t('studio.comments.markSpam') }}</button>
        <button type="button" class="st-btn q sm danger" @click="bulk('delete')"><SIcon name="trash" :size="16" />{{ t('studio.delete') }}</button>
      </template>
    </div>

    <div v-if="!loading && !list.length" class="st-empty">
      <EmptyArt :icon="MessageSquareDashed" />
      <h4>{{ t(`studio.comments.empty_${tab}`) }}</h4>
      <p>{{ t(`studio.comments.empty_${tab}Sub`) }}</p>
    </div>

    <TransitionGroup tag="div" name="cm" class="cm-list">
      <div v-for="c in list" :key="c.id" class="cm" :class="{ sel: selected.has(c.id) }">
        <span class="st-ck" :class="{ on: selected.has(c.id) }" role="checkbox" :aria-checked="selected.has(c.id)" @click="toggle(c.id)"><Icon :icon="Check" /></span>
        <span class="av" :style="{ background: c.author.avatar ? undefined : tint(c.author.name) }">
          <img v-if="c.author.avatar" :src="c.author.avatar" alt="" /><template v-else>{{ initial(c.author.name) }}</template>
        </span>
        <div class="body">
          <div class="who">
            <b>{{ c.author.name }}</b>
            <span class="role" :class="c.author.role">{{ t(`studio.comments.role_${c.author.role}`) }}</span>
            <span>{{ c.parentId ? t('studio.comments.repliedIn') : t('studio.comments.on') }}</span>
            <a :href="c.targetLink" target="_blank" rel="noopener">{{ c.target === 'guestbook' ? t('studio.comments.guestbook') : `《${plainText(c.targetTitle) || t('studio.untitled')}》` }}</a>
            <span>· {{ relTime(c.createdAt) }}</span>
            <span v-if="c.hasLink" class="flag">{{ t('studio.comments.hasLink') }}</span>
          </div>
          <p>{{ c.body }}</p>

          <div class="reply-fold" :class="{ open: replying === c.id }">
            <div class="clip">
              <div class="reply">
                <textarea
                  v-if="replying === c.id"
                  ref="replyEl"
                  v-model="replyText"
                  rows="3"
                  maxlength="2000"
                  :placeholder="t('studio.comments.replyPh', { name: c.author.name })"
                  @keydown="onReplyKey($event, c)"
                />
                <div class="rf">
                  <small>{{ c.status === 'pending' ? t('studio.comments.replyApproves') : t('studio.comments.replyHint') }}</small>
                  <button type="button" class="st-btn g sm" @click="replying = null">{{ t('studio.cancel') }}</button>
                  <button type="button" class="st-btn p sm" :disabled="replyBusy || !replyText.trim()" @click="sendReply(c)">
                    <SIcon name="send" :size="15" />{{ t('studio.comments.send') }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="acts">
          <button v-if="c.status === 'pending'" type="button" class="st-btn p sm" @click="setStatus(c, 'approved')"><SIcon name="check" :size="16" />{{ t('studio.comments.approve') }}</button>
          <button v-if="c.status === 'spam'" type="button" class="st-btn g sm" @click="setStatus(c, 'approved')">{{ t('studio.comments.notSpam') }}</button>
          <button v-if="c.status !== 'spam'" type="button" class="st-ibtn" :class="{ on: replying === c.id }" :title="t('studio.comments.reply')" @click="openReply(c)"><SIcon name="reply" :size="18" /></button>
          <button v-if="c.status === 'approved'" type="button" class="st-ibtn" :title="t('studio.comments.unpublish')" @click="setStatus(c, 'pending')"><SIcon name="eyeOff" :size="18" /></button>
          <button v-if="c.status !== 'spam'" type="button" class="st-ibtn" :title="t('studio.comments.markSpam')" @click="setStatus(c, 'spam')"><SIcon name="flag" :size="18" /></button>
          <button type="button" class="st-ibtn" :title="t('studio.delete')" @click="remove(c)"><SIcon name="trash" :size="18" /></button>
        </div>
      </div>
    </TransitionGroup>
  </section>
</template>

<style scoped lang="scss">
.view {
  max-width: 1280px;
  margin: 0 auto;
  padding: 32px 48px 72px;
}

.st-note-bar .st-link { margin-left: auto; }

.bulk {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 44px;
  padding: 6px 12px;
  margin-bottom: 4px;
  font-size: 13.5px;
  color: var(--st-ink-3);

  .st-ckrow { font-size: 14px; cursor: pointer; }
  .sel { color: var(--st-ink-2); }
  .sp { flex: 1; }
  .danger { color: var(--red); }
}

.cm-list { position: relative; display: flex; flex-direction: column; }

.cm {
  display: grid;
  grid-template-columns: 18px 40px minmax(0, 1fr) auto;
  gap: 14px;
  padding: 14px 12px;
  border-bottom: 1px solid var(--line);
  align-items: start;
  transition: background var(--dur-fast);

  &:hover, &.sel { background: linear-gradient(90deg, transparent, var(--well) 12%, var(--well) 88%, transparent); }

  > .st-ck { margin-top: 11px; cursor: pointer; }

  .av {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    overflow: hidden;
    color: #fff;
    font: 600 15px var(--font-serif);

    img { width: 100%; height: 100%; object-fit: cover; }
  }

  .body { min-width: 0; }

  .who {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    font-size: 13.5px;
    color: var(--st-ink-3);
    margin-bottom: 4px;

    b { color: var(--st-ink); font-weight: 600; font-size: 15px; }
    a { color: var(--st-ink-2); font-family: var(--font-serif); }
    a:hover { color: var(--ink); }
  }

  .role {
    font-size: 11.5px;
    padding: 1px 7px;
    border-radius: var(--r-pill);
    background: var(--well-2);
    color: var(--st-ink-3);

    &.admin { background: var(--tint); color: var(--ink); }
    &.author { background: color-mix(in oklab, var(--green) 12%, var(--paper)); color: color-mix(in oklab, var(--green) 60%, var(--st-ink)); }
  }

  p { margin: 0; font-size: 15px; line-height: 1.7; color: var(--st-ink); white-space: pre-wrap; overflow-wrap: anywhere; }

  .flag {
    font-size: 12.5px;
    color: color-mix(in oklab, var(--yellow) 60%, var(--st-ink));
    background: color-mix(in oklab, var(--yellow) 14%, var(--paper));
    padding: 1px 7px;
    border-radius: var(--r-xs);
  }

  .acts { display: flex; align-items: center; gap: 6px; opacity: 0.6; transition: opacity var(--dur-fast); }
  &:hover .acts, &.sel .acts { opacity: 1; }
  .st-ibtn.on { background: var(--tint); color: var(--ink); }
}

/* 回复框：grid 行高展开 / 收起 */
.reply-fold {
  display: grid;
  grid-template-rows: 0fr;
  transition: grid-template-rows var(--dur) var(--ease-out);

  &.open { grid-template-rows: 1fr; }
}

.clip { min-height: 0; overflow: hidden; }

.reply {
  padding-top: 10px;

  textarea {
    width: 100%;
    resize: vertical;
    padding: 10px 12px;
    border: 0;
    border-radius: var(--r-md);
    background: var(--well);
    box-shadow: 0 0 0 1px var(--line-2);
    color: var(--st-ink);
    font: 14.5px/1.7 var(--font-sans);
    outline: none;
    transition: box-shadow var(--dur-fast);

    &:focus { box-shadow: 0 0 0 1px var(--ink), 0 0 0 4px var(--tint); }
  }

  .rf {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 8px;

    small { flex: 1; font-size: 12.5px; color: var(--st-ink-3); }
  }
}

.cm-enter-active, .cm-leave-active { transition: opacity var(--dur) var(--ease-out), transform var(--dur) var(--ease-out); }
.cm-enter-from { opacity: 0; transform: translateY(6px); }
.cm-leave-to { opacity: 0; transform: translateX(24px); }
.cm-leave-active { position: absolute; left: 0; right: 0; }
.cm-move { transition: transform var(--dur) var(--ease-out); }

@media (max-width: 1180px) {
  .view { padding: 28px 32px 64px; }
}

@media (prefers-reduced-motion: reduce) {
  .reply-fold, .cm-enter-active, .cm-leave-active, .cm-move { transition: none; }
}
</style>
