<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, type AdminUser, type InviteInfo, type UserFilter, type UserRole, type UserStats } from '../../api';
import { useConfigStore, type UsersConfig } from '../../stores/config';
import { useDialogStore } from '../../stores/dialog';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import StSwitch from './studio/StSwitch.vue';
import StSeg from './studio/StSeg.vue';
import StModal from './studio/StModal.vue';
import PopMenu from './studio/PopMenu.vue';
import StPager from './studio/StPager.vue';
import { copyText, saveBlob } from './studio/state';
import { toast } from './studio/toast';
import { dateText, initial, relTime } from './studio/format';
import type { MenuItem } from './studio/types';

/**
 * 用户：开放设置（总开关 → 读者 / 协作作者 / 评论，各自独立；访客回应不受总开关约束）、
 * 统计、用户表（角色 / 停用 / 重置密码链接 / 删除）、邀请（生成一次性链接，可选发邮件）。
 * 开关改动即时保存。
 */
const { t } = useI18n();
const config = useConfigStore();
const dialog = useDialogStore();

/* ---------- 开放设置 ---------- */
const sw = reactive<UsersConfig>({
  enabled: false,
  readers: { enabled: false, signup: 'open', requireVerify: false },
  authors: { enabled: false, directPublish: false },
  comments: { enabled: true, anonymous: false, moderation: 'first' },
  login: { github: false },
  reactions: true,
});
const mailReady = ref(false);
const loadedSw = ref(false);

let saveTimer = 0;
function saveSwitches(): void {
  if (!loadedSw.value) return;
  window.clearTimeout(saveTimer);
  saveTimer = window.setTimeout(async () => {
    try {
      await adminApi.saveSettings({ users: JSON.parse(JSON.stringify(sw)) });
      void config.load();
      toast(t('studio.users.swSaved'));
    } catch {
      toast(t('studio.saveFailed'), { icon: 'x' });
    }
  }, 350);
}

/** 双向绑定到开关对象的某个字段，改动即保存 */
function bind<T>(get: () => T, set: (v: T) => void) {
  return computed<T>({
    get,
    set: (v) => {
      set(v);
      saveSwitches();
    },
  });
}
const master = bind(() => sw.enabled, (v) => (sw.enabled = v));
const readersOn = bind(() => sw.readers.enabled, (v) => (sw.readers.enabled = v));
const signup = bind(() => sw.readers.signup, (v) => (sw.readers.signup = v));
const requireVerify = bind(() => !!sw.readers.requireVerify, (v) => (sw.readers.requireVerify = v));
const authorsOn = bind(() => sw.authors.enabled, (v) => (sw.authors.enabled = v));
const directPublish = bind(() => !!sw.authors.directPublish, (v) => (sw.authors.directPublish = v));
const commentsOn = bind(() => sw.comments.enabled, (v) => (sw.comments.enabled = v));
const anonymous = bind(() => sw.comments.anonymous, (v) => (sw.comments.anonymous = v));
const moderation = bind(() => sw.comments.moderation, (v) => (sw.comments.moderation = v));
const reactions = bind(() => sw.reactions !== false, (v) => (sw.reactions = v));

const SIGNUP = computed(() => (['open', 'invite', 'closed'] as const).map((v) => ({ value: v, label: t(`studio.users.signup_${v}`) })));
const MODERATION = computed(() => (['all', 'first', 'none'] as const).map((v) => ({ value: v, label: t(`studio.users.mod_${v}`) })));

/* ---------- 用户（服务端分页：角色分段 + 搜索都在服务端完成） ---------- */
/** 当前页的用户 */
const users = ref<AdminUser[]>([]);
const stats = ref<UserStats | null>(null);
const loaded = ref(false);
const busy = ref(false);
type Filter = UserFilter;
const filter = ref<Filter>('all');
const q = ref('');
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
/** 各分段在当前搜索下的人数 */
const counts = ref<Record<Filter, number>>({ all: 0, admin: 0, author: 0, reader: 0, disabled: 0 });
/** 待审头像（不受分页影响） */
const pendingAvatars = ref<AdminUser[]>([]);

const FILTERS = computed(() =>
  (['all', 'admin', 'author', 'reader', 'disabled'] as const).map((v) => ({
    value: v,
    label: t(`studio.users.f_${v}`),
    count: v === 'all' ? counts.value.all : counts.value[v] || undefined,
  })),
);

const inFilter = (u: AdminUser, f: Filter): boolean => (f === 'disabled' ? u.status === 'disabled' : f === 'all' || u.role === f);

let loadSeq = 0;
async function loadUsers(): Promise<void> {
  const seq = ++loadSeq;
  busy.value = true;
  try {
    const needle = q.value.trim();
    const res = await adminApi.usersPage({
      page: page.value,
      pageSize: pageSize.value,
      role: filter.value === 'all' ? undefined : filter.value,
      q: needle || undefined,
    });
    if (seq !== loadSeq) return;
    stats.value = res.stats;
    if (typeof res.total === 'number' && res.counts) {
      users.value = res.items;
      total.value = res.total;
      counts.value = res.counts;
      pendingAvatars.value = res.pendingAvatars ?? [];
    } else {
      // 旧后端一次返回全部：在前端筛选、计数、切页
      const lower = needle.toLowerCase();
      const match = res.items.filter((u) => !lower || `${u.name} ${u.login} ${u.email}`.toLowerCase().includes(lower));
      const list = match.filter((u) => inFilter(u, filter.value));
      counts.value = Object.fromEntries(
        (['all', 'admin', 'author', 'reader', 'disabled'] as const).map((f) => [f, match.filter((u) => inFilter(u, f)).length]),
      ) as Record<Filter, number>;
      total.value = list.length;
      users.value = list.slice((page.value - 1) * pageSize.value, page.value * pageSize.value);
      pendingAvatars.value = res.items.filter((u) => u.avatarPending);
    }
  } catch {
    if (seq === loadSeq) toast(t('studio.loadFailed'), { icon: 'x' });
  } finally {
    if (seq === loadSeq) {
      busy.value = false;
      loaded.value = true;
    }
  }
}

// 换分段 / 搜索回到第 1 页（搜索输入防抖）；翻页 / 换每页条数直接重取
let searchTimer = 0;
watch(q, () => {
  window.clearTimeout(searchTimer);
  searchTimer = window.setTimeout(() => {
    if (page.value !== 1) page.value = 1;
    else void loadUsers();
  }, 250);
});
watch(filter, () => {
  if (page.value !== 1) page.value = 1;
  else void loadUsers();
});
watch([page, pageSize], () => void loadUsers());

async function loadSwitches(): Promise<void> {
  try {
    const s = (await adminApi.settings()) as { users?: Partial<UsersConfig>; mail?: { enabled?: boolean; host?: string } };
    const u = s.users ?? {};
    Object.assign(sw, u, {
      readers: { ...sw.readers, ...u.readers },
      authors: { ...sw.authors, ...u.authors },
      comments: { ...sw.comments, ...u.comments },
      login: { ...sw.login, ...u.login },
    });
    mailReady.value = !!s.mail?.enabled && !!s.mail?.host;
  } catch { /* 读取失败时保持默认展示 */ }
  loadedSw.value = true;
}

onMounted(() => {
  void loadSwitches();
  void loadUsers();
  void loadInvites();
});

const ROLES: UserRole[] = ['admin', 'author', 'reader'];

async function setRole(u: AdminUser, role: UserRole): Promise<void> {
  if (role === u.role) return;
  if (role === 'admin') {
    const ok = await dialog.confirm({
      title: t('studio.users.promoteTitle', { name: u.name }),
      message: t('studio.users.promoteBody'),
      confirmText: t('studio.users.promote'),
      danger: true,
    });
    if (!ok) return;
  }
  try {
    await adminApi.updateUser(u.id, { role });
    u.role = role;
    toast(t('studio.users.roleChanged', { name: u.name, role: t(`studio.users.r_${role}`) }));
    void loadUsers();
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  }
}

async function setActive(u: AdminUser, on: boolean): Promise<void> {
  try {
    await adminApi.updateUser(u.id, { status: on ? 'active' : 'disabled' });
    u.status = on ? 'active' : 'disabled';
    toast(on ? t('studio.users.enabledOne', { name: u.name }) : t('studio.users.disabledOne', { name: u.name }), { icon: on ? 'check' : 'lock' });
    // 「已停用」分段计数随之变化
    void loadUsers();
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  }
}

async function remove(u: AdminUser): Promise<void> {
  const ok = await dialog.confirm({
    title: t('studio.users.deleteTitle', { name: u.name }),
    message: t('studio.users.deleteBody'),
    confirmText: t('studio.delete'),
    danger: true,
  });
  if (!ok) return;
  try {
    await adminApi.deleteUser(u.id);
    toast(t('studio.deleted'));
    void loadUsers();
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  }
}

/* 重置密码链接 */
const reset = reactive({ open: false, user: null as AdminUser | null, url: '', sent: false, busy: false });
async function makeReset(u: AdminUser, send: boolean): Promise<void> {
  Object.assign(reset, { open: true, user: u, url: '', sent: false, busy: true });
  try {
    const res = await adminApi.resetLink(u.id, send);
    reset.url = res.url;
    reset.sent = res.sent;
  } catch {
    reset.open = false;
    toast(t('studio.saveFailed'), { icon: 'x' });
  } finally {
    reset.busy = false;
  }
}

function menuOf(u: AdminUser): MenuItem[] {
  const items: MenuItem[] = [{ icon: 'key', label: t('studio.users.resetLink'), run: () => void makeReset(u, false) }];
  if (u.email && mailReady.value) items.push({ icon: 'mail', label: t('studio.users.resetMail'), run: () => void makeReset(u, true) });
  items.push({ icon: 'trash', label: t('studio.delete'), danger: true, divider: true, run: () => void remove(u) });
  return items;
}

async function copy(url: string): Promise<void> {
  toast((await copyText(url)) ? t('studio.copied') : t('studio.users.copyFailed'), { icon: 'copy' });
}

async function exportCsv(): Promise<void> {
  try {
    saveBlob(await adminApi.exportUsers(), `users-${new Date().toISOString().slice(0, 10)}.csv`);
  } catch {
    toast(t('studio.loadFailed'), { icon: 'x' });
  }
}

/* ---------- 头像审核（待审列表由服务端单独下发，不受用户表分页影响） ---------- */
const reviewing = ref(new Set<number>());
async function review(u: AdminUser, action: 'approve' | 'reject'): Promise<void> {
  if (reviewing.value.has(u.id)) return;
  reviewing.value = new Set(reviewing.value).add(u.id);
  try {
    await adminApi.reviewAvatar(u.id, action);
    // 同步到当前页里的同一用户
    const row = users.value.find((x) => x.id === u.id);
    if (action === 'approve') {
      const next = u.avatarPending || u.avatar;
      if (row) row.avatar = next;
    }
    if (row) row.avatarPending = '';
    pendingAvatars.value = pendingAvatars.value.filter((x) => x.id !== u.id);
    toast(action === 'approve' ? t('studio.users.avatarApproved', { name: u.name }) : t('studio.users.avatarRejected', { name: u.name }), {
      icon: action === 'approve' ? 'check' : 'x',
    });
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  } finally {
    const s = new Set(reviewing.value);
    s.delete(u.id);
    reviewing.value = s;
  }
}

/* ---------- 邀请 ---------- */
const invites = ref<InviteInfo[]>([]);
async function loadInvites(): Promise<void> {
  try {
    invites.value = await adminApi.invites();
  } catch { /* 忽略 */ }
}

const inv = reactive({
  open: false,
  role: 'author' as 'author' | 'reader',
  email: '',
  days: 7,
  note: '',
  send: false,
  busy: false,
  url: '',
  sent: false,
});
function openInvite(): void {
  Object.assign(inv, { open: true, role: sw.authors.enabled || !sw.readers.enabled ? 'author' : 'reader', email: '', days: 7, note: '', send: false, url: '', sent: false });
}
async function createInvite(): Promise<void> {
  if (inv.busy) return;
  inv.busy = true;
  try {
    const res = await adminApi.createInvite({
      role: inv.role,
      email: inv.email.trim() || undefined,
      days: inv.days,
      note: inv.note.trim() || undefined,
      send: inv.send && !!inv.email.trim(),
    });
    inv.url = res.url;
    inv.sent = res.sent;
    void loadInvites();
  } catch (e) {
    toast((e as Error).message === 'invalid_email' ? t('studio.users.badEmail') : t('studio.saveFailed'), { icon: 'x' });
  } finally {
    inv.busy = false;
  }
}

async function revokeInvite(i: InviteInfo): Promise<void> {
  try {
    await adminApi.deleteInvite(i.id);
    invites.value = invites.value.filter((x) => x.id !== i.id);
    toast(t('studio.users.inviteRevoked'));
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  }
}

const inviteState = (i: InviteInfo): 'used' | 'expired' | 'open' => {
  if (i.usedAt) return 'used';
  return new Date(`${i.expiresAt.replace(' ', 'T')}Z`).getTime() < Date.now() ? 'expired' : 'open';
};

/** 无头像时的底色：按名字稳定取色 */
const PALETTE = ['#3b7d5a', '#556b8d', '#b5651d', '#9a4f7a', '#6d6a3a', '#2f6f8f', '#8a4b3c'];
const tint = (s: string): string => PALETTE[[...s].reduce((a, c) => a + c.charCodeAt(0), 0) % PALETTE.length];
const isOnline = (u: AdminUser): boolean => !!u.lastActiveAt && Date.now() - new Date(`${u.lastActiveAt.replace(' ', 'T')}Z`).getTime() < 5 * 60_000;
</script>

<template>
  <section class="studio view">
    <div class="st-vh">
      <div>
        <h1>{{ t('studio.users.title') }}</h1>
        <p>{{ t('studio.users.desc') }}</p>
      </div>
      <div class="act">
        <button type="button" class="st-btn g" @click="exportCsv"><SIcon name="download" :size="18" />{{ t('studio.users.export') }}</button>
        <button type="button" class="st-btn p" :disabled="!sw.enabled" :title="sw.enabled ? '' : t('studio.users.inviteOff')" @click="openInvite">
          <SIcon name="mail" :size="18" />{{ t('studio.users.invite') }}
        </button>
      </div>
    </div>

    <!-- 开放设置 -->
    <section class="st-card sw-card st-rise">
      <div class="sw-row master">
        <div class="tx">
          <b>{{ t('studio.users.master') }}</b>
          <small>{{ sw.enabled ? t('studio.users.masterOn') : t('studio.users.masterOff') }}</small>
        </div>
        <StSwitch v-model="master" :label="t('studio.users.master')" />
      </div>

      <div class="fold" :class="{ open: sw.enabled }" :inert="!sw.enabled">
        <div class="clip">
          <div class="groups">
            <div class="grp">
              <div class="sw-row">
                <div class="tx"><b>{{ t('studio.users.readers') }}</b><small>{{ t('studio.users.readersSub') }}</small></div>
                <StSwitch v-model="readersOn" />
              </div>
              <div class="sub" :class="{ off: !sw.readers.enabled }">
                <div class="line">
                  <span class="st-flabel">{{ t('studio.users.signup') }}</span>
                  <StSeg v-model="signup" :options="SIGNUP" />
                </div>
                <div class="line">
                  <div class="tx">
                    <span>{{ t('studio.users.requireVerify') }}</span>
                    <small v-if="!mailReady">{{ t('studio.users.needMail') }}</small>
                  </div>
                  <StSwitch v-model="requireVerify" :disabled="!mailReady && !sw.readers.requireVerify" />
                </div>
              </div>
            </div>

            <div class="grp">
              <div class="sw-row">
                <div class="tx"><b>{{ t('studio.users.authors') }}</b><small>{{ t('studio.users.authorsSub') }}</small></div>
                <StSwitch v-model="authorsOn" />
              </div>
              <div class="sub" :class="{ off: !sw.authors.enabled }">
                <div class="line">
                  <div class="tx">
                    <span>{{ t('studio.users.directPublish') }}</span>
                    <small>{{ sw.authors.directPublish ? t('studio.users.directOn') : t('studio.users.directOff') }}</small>
                  </div>
                  <StSwitch v-model="directPublish" />
                </div>
              </div>
            </div>

            <div class="grp">
              <div class="sw-row">
                <div class="tx"><b>{{ t('studio.users.comments') }}</b><small>{{ t('studio.users.commentsSub') }}</small></div>
                <StSwitch v-model="commentsOn" />
              </div>
              <div class="sub" :class="{ off: !sw.comments.enabled }">
                <div class="line">
                  <span class="st-flabel">{{ t('studio.users.moderation') }}</span>
                  <StSeg v-model="moderation" :options="MODERATION" />
                </div>
                <div class="line">
                  <div class="tx"><span>{{ t('studio.users.anonymous') }}</span><small>{{ t('studio.users.anonymousSub') }}</small></div>
                  <StSwitch v-model="anonymous" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="sw-row foot">
        <div class="tx"><b>{{ t('studio.users.reactions') }}</b><small>{{ t('studio.users.reactionsSub') }}</small></div>
        <StSwitch v-model="reactions" />
      </div>
    </section>

    <!-- 待审头像 -->
    <section v-if="pendingAvatars.length" class="st-card av-card st-rise" style="--i: 1">
      <div class="st-sec-t"><h2>{{ t('studio.users.avatarReview') }}</h2><span>{{ t('studio.users.avatarReviewSub') }}</span></div>
      <TransitionGroup tag="ul" name="row" class="av-list">
        <li v-for="u in pendingAvatars" :key="u.id">
          <span class="pair">
            <span class="av old" :style="{ background: u.avatar ? undefined : tint(u.name || u.login) }">
              <img v-if="u.avatar" :src="u.avatar" alt="" /><template v-else>{{ initial(u.name || u.login) }}</template>
            </span>
            <SIcon name="arrowR" :size="16" class="to" />
            <span class="av new"><img :src="u.avatarPending" alt="" /></span>
          </span>
          <div class="tx"><b>{{ u.name || u.login }}</b><small class="mono">{{ u.email || `@${u.login}` }}</small></div>
          <button type="button" class="st-btn g sm" :disabled="reviewing.has(u.id)" @click="review(u, 'reject')">{{ t('studio.users.reject') }}</button>
          <button type="button" class="st-btn p sm" :disabled="reviewing.has(u.id)" @click="review(u, 'approve')"><SIcon name="check" :size="15" />{{ t('studio.users.approve') }}</button>
        </li>
      </TransitionGroup>
    </section>

    <!-- 统计 -->
    <div class="st-stats u-stats st-rise" style="--i: 1; --n: 5">
      <div class="st-stat"><b>{{ stats?.total ?? 0 }}<span v-if="stats?.newMonth" class="dl up">+{{ stats.newMonth }}</span></b><small>{{ t('studio.users.sTotal') }}</small></div>
      <div class="st-stat"><b>{{ stats?.activeMonth ?? 0 }}</b><small>{{ t('studio.users.sActive') }}</small></div>
      <div class="st-stat"><b>{{ stats?.authors ?? 0 }}</b><small>{{ t('studio.users.sAuthors') }}</small></div>
      <div class="st-stat"><b>{{ stats?.comments ?? 0 }}<span v-if="stats?.commentsMonth" class="dl up">+{{ stats.commentsMonth }}</span></b><small>{{ t('studio.users.sComments') }}</small></div>
      <router-link class="st-stat" :to="{ name: 'admin-comments' }">
        <b :class="{ warn: stats?.pending }">{{ stats?.pending ?? 0 }}</b><small>{{ t('studio.users.sPending') }}</small>
      </router-link>
    </div>

    <!-- 用户表 -->
    <div class="tools st-rise" style="--i: 2">
      <StSeg v-model="filter" :options="FILTERS" />
      <label class="st-field search"><SIcon name="search" :size="16" /><input v-model="q" :placeholder="t('studio.users.searchPh')" /></label>
    </div>

    <table class="st-table u-table" :class="{ busy }">
      <thead>
        <tr>
          <th>{{ t('studio.users.cUser') }}</th>
          <th>{{ t('studio.users.cRole') }}</th>
          <th>{{ t('studio.users.cJoined') }}</th>
          <th>{{ t('studio.users.cActive') }}</th>
          <th>{{ t('studio.users.cComments') }}</th>
          <th>{{ t('studio.users.cStatus') }}</th>
          <th />
        </tr>
      </thead>
      <TransitionGroup tag="tbody" name="row">
        <tr v-for="u in users" :key="u.id">
          <td>
            <div class="u">
              <span class="av" :style="{ background: u.avatar ? undefined : tint(u.name || u.login) }">
                <img v-if="u.avatar" :src="u.avatar" alt="" /><template v-else>{{ initial(u.name || u.login) }}</template>
              </span>
              <div class="who">
                <b>{{ u.name || u.login }}<span v-if="u.self" class="me">{{ t('studio.users.you') }}</span></b>
                <small>
                  <span class="mono">{{ u.email || `@${u.login}` }}</span>
                  <SIcon v-if="u.github" name="branch" :size="13" class="st-tip" :data-tip="t('studio.users.githubLinked')" />
                  <SIcon v-if="u.email && u.emailVerified" name="check" :size="13" class="st-tip ok" :data-tip="t('studio.users.verified')" />
                </small>
              </div>
            </div>
          </td>
          <td>
            <span v-if="u.self" class="role" :class="u.role">{{ t(`studio.users.r_${u.role}`) }}</span>
            <label v-else class="role pick" :class="u.role">
              {{ t(`studio.users.r_${u.role}`) }}<SIcon name="chevronD" :size="13" />
              <select :value="u.role" :aria-label="t('studio.users.cRole')" @change="setRole(u, ($event.target as HTMLSelectElement).value as UserRole)">
                <option v-for="r in ROLES" :key="r" :value="r">{{ t(`studio.users.r_${r}`) }}</option>
              </select>
            </label>
          </td>
          <td class="mono">{{ dateText(u.createdAt) }}</td>
          <td class="act-cell">
            <span v-if="isOnline(u)" class="online"><i class="st-dot" />{{ t('studio.users.online') }}</span>
            <template v-else>{{ u.lastActiveAt ? relTime(u.lastActiveAt) : t('studio.users.never') }}</template>
          </td>
          <td class="mono">{{ u.comments }}</td>
          <td>
            <span v-if="u.self" class="dash">—</span>
            <span v-else-if="u.status === 'pending'" class="chip pend">{{ t('studio.users.pending') }}</span>
            <StSwitch v-else :model-value="u.status === 'active'" @update:model-value="(v: boolean) => setActive(u, v)" />
          </td>
          <td><PopMenu v-if="!u.self" :items="menuOf(u)" /></td>
        </tr>
      </TransitionGroup>
    </table>
    <p v-if="loaded && !users.length" class="empty">{{ q || filter !== 'all' ? t('studio.users.noMatch') : t('studio.users.empty') }}</p>
    <StPager v-model:page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[20, 50, 100]" :busy="busy" />


    <!-- 邀请记录 -->
    <section v-if="invites.length" class="st-card inv-card st-rise" style="--i: 3">
      <div class="st-sec-t"><h2>{{ t('studio.users.invites') }}</h2><span>{{ t('studio.users.invitesSub') }}</span></div>
      <ul class="inv">
        <li v-for="i in invites" :key="i.id" :class="inviteState(i)">
          <span class="role" :class="i.role">{{ t(`studio.users.r_${i.role}`) }}</span>
          <div class="tx">
            <b>{{ i.email || t('studio.users.anyEmail') }}<em v-if="i.note">{{ i.note }}</em></b>
            <small v-if="inviteState(i) === 'used'">{{ t('studio.users.usedBy', { name: i.usedBy, when: dateText(i.usedAt) }) }}</small>
            <small v-else-if="inviteState(i) === 'expired'">{{ t('studio.users.expired') }}</small>
            <small v-else>{{ t('studio.users.expires', { when: dateText(i.expiresAt) }) }}</small>
          </div>
          <button v-if="inviteState(i) === 'open'" type="button" class="st-btn q sm" @click="revokeInvite(i)">{{ t('studio.users.revoke') }}</button>
        </li>
      </ul>
    </section>

    <!-- 邀请弹窗 -->
    <StModal :open="inv.open" @close="inv.open = false">
      <template v-if="!inv.url">
        <div class="mi"><SIcon name="mail" :size="22" /></div>
        <h3>{{ t('studio.users.inviteTitle') }}</h3>
        <p>{{ t('studio.users.inviteDesc') }}</p>
        <div class="st-flabel">{{ t('studio.users.inviteRole') }}</div>
        <StSeg
          v-model="inv.role"
          :options="[
            { value: 'author', label: t('studio.users.r_author') },
            { value: 'reader', label: t('studio.users.r_reader') },
          ]"
        />
        <div class="st-flabel gap">{{ t('studio.users.inviteEmail') }}<em>{{ t('studio.users.optional') }}</em></div>
        <label class="st-field"><input v-model="inv.email" type="email" placeholder="name@example.com" /></label>
        <div class="two">
          <label>
            <span class="st-flabel gap">{{ t('studio.users.inviteDays') }}</span>
            <span class="st-field"><input v-model.number="inv.days" type="number" min="1" max="30" /></span>
          </label>
          <label>
            <span class="st-flabel gap">{{ t('studio.users.inviteNote') }}<em>{{ t('studio.users.optional') }}</em></span>
            <span class="st-field"><input v-model="inv.note" maxlength="60" /></span>
          </label>
        </div>
        <label v-if="mailReady" class="st-ckrow send" :class="{ dim: !inv.email.trim() }" @click.prevent="inv.send = !inv.send">
          <span class="st-ck" :class="{ on: inv.send && !!inv.email.trim() }"><SIcon name="check" /></span>{{ t('studio.users.inviteSend') }}
        </label>
        <div class="ft">
          <button type="button" class="st-btn g" @click="inv.open = false">{{ t('studio.cancel') }}</button>
          <button type="button" class="st-btn p" :disabled="inv.busy" @click="createInvite">{{ t('studio.users.inviteCreate') }}</button>
        </div>
      </template>
      <template v-else>
        <div class="mi ok"><SIcon name="check" :size="22" /></div>
        <h3>{{ t('studio.users.inviteReady') }}</h3>
        <p>{{ inv.sent ? t('studio.users.inviteSent', { email: inv.email }) : t('studio.users.inviteCopy', { n: inv.days }) }}</p>
        <div class="link-box"><code>{{ inv.url }}</code><button type="button" class="st-btn g sm" @click="copy(inv.url)"><SIcon name="copy" :size="16" />{{ t('studio.copy') }}</button></div>
        <div class="ft"><button type="button" class="st-btn p" @click="inv.open = false">{{ t('studio.users.done') }}</button></div>
      </template>
    </StModal>

    <!-- 重置密码链接 -->
    <StModal :open="reset.open" @close="reset.open = false">
      <div class="mi"><SIcon name="key" :size="22" /></div>
      <h3>{{ t('studio.users.resetTitle', { name: reset.user?.name ?? '' }) }}</h3>
      <p>{{ reset.sent ? t('studio.users.resetSent', { email: reset.user?.email ?? '' }) : t('studio.users.resetDesc') }}</p>
      <div class="link-box">
        <code>{{ reset.busy ? t('studio.loading') : reset.url }}</code>
        <button type="button" class="st-btn g sm" :disabled="reset.busy" @click="copy(reset.url)"><SIcon name="copy" :size="16" />{{ t('studio.copy') }}</button>
      </div>
      <div class="ft"><button type="button" class="st-btn p" @click="reset.open = false">{{ t('studio.users.done') }}</button></div>
    </StModal>
  </section>
</template>

<style scoped lang="scss">
.view {
  max-width: 1280px;
  margin: 0 auto;
  padding: 32px 48px 72px;
}

/* ---------- 开放设置 ---------- */
.sw-card { margin-bottom: 20px; padding: 6px 24px; }

.sw-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 16px 0;

  &.foot { box-shadow: 0 -1px 0 var(--line-2); }
}

.tx {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;

  b { font-size: 15px; font-weight: 600; }
  span { font-size: 14px; }
  small { font-size: 13px; color: var(--st-ink-3); line-height: 1.5; }
}

.master b { font: 600 17px var(--font-serif); }

/* 总开关展开 / 收起：grid 行高过渡，内容常驻 */
.fold {
  display: grid;
  grid-template-rows: 0fr;
  opacity: 0;
  transition: grid-template-rows var(--dur-slow) var(--ease-out), opacity var(--dur) var(--ease-out);

  &.open { grid-template-rows: 1fr; opacity: 1; }
}

.clip { min-height: 0; overflow: hidden; }

.groups {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
  padding: 4px 0 18px;
}

.grp {
  padding: 4px 16px 14px;
  border-radius: var(--r-md);
  background: var(--well);

  .sw-row { padding: 12px 0; }
}

.sub {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-top: 12px;
  box-shadow: 0 -1px 0 var(--line-2);
  transition: opacity var(--dur) var(--ease-out);

  &.off { opacity: 0.45; pointer-events: none; }

  .line {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
  }

  .st-flabel { margin: 0; }
  /* 文案可换行，开关始终留在同一行右侧 */
  .line > .tx { flex: 1 1 160px; }
  .line > .sw { flex: none; }
}

/* ---------- 待审头像 ---------- */
.av-card { margin-bottom: 20px; }

.av-list {
  position: relative;
  list-style: none;
  margin: 6px 0 0;
  padding: 0;

  li {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 12px 0;

    & + li { box-shadow: 0 -1px 0 var(--line-2); }
  }

  .pair { display: flex; align-items: center; gap: 8px; }
  .to { color: var(--st-ink-3); }

  .av {
    width: 44px;
    height: 44px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    overflow: hidden;
    color: #fff;
    font: 600 15px var(--font-serif);
    flex: none;

    img { width: 100%; height: 100%; object-fit: cover; }
    &.old { width: 32px; height: 32px; font-size: 12px; opacity: 0.7; }
    &.new { box-shadow: 0 0 0 2px var(--paper), 0 0 0 3.5px var(--ink); }
  }

  .tx { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
  b { font-weight: 600; font-size: 15px; }
  small { color: var(--st-ink-3); font-size: 12.5px; }
}

/* ---------- 统计 / 工具条 ---------- */
.u-stats { margin-bottom: 22px; }
.u-stats .dl { margin-left: 6px; }
.u-stats a.st-stat { color: inherit; }
.u-stats .warn { color: var(--yellow); }

.tools {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 14px;

  .search { width: 260px; gap: 8px; color: var(--st-ink-3); }
}

/* ---------- 用户表 ---------- */
.st-table {
  .u { display: flex; align-items: center; gap: 12px; }

  .av {
    width: 38px;
    height: 38px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    color: #fff;
    font: 600 13px var(--font-serif);
    overflow: hidden;
    flex: none;

    img { width: 100%; height: 100%; object-fit: cover; }
  }

  .who { min-width: 0; }
  b { display: flex; align-items: center; gap: 8px; font-weight: 600; font-size: 15px; }

  .me {
    font-size: 11.5px;
    font-weight: 500;
    padding: 1px 8px;
    border-radius: var(--r-pill);
    background: var(--tint);
    color: var(--ink);
  }

  small {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--st-ink-3);
    font-size: 13px;

    .ok { color: var(--green); }
  }

  .mono { font-size: 13px; color: var(--st-ink-3); }
  .act-cell { font-size: 14px; color: var(--st-ink-2); }
  .online { display: inline-flex; align-items: center; gap: 6px; .st-dot { --c: var(--green); } }
  .dash { font-size: 12px; color: var(--st-ink-3); }
}

.role {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  padding: 4px 11px;
  border-radius: var(--r-pill);
  background: var(--well-2);
  color: var(--st-ink-2);
  white-space: nowrap;

  &.admin { background: var(--tint); color: var(--ink); }
  &.author { background: color-mix(in oklab, var(--green) 12%, var(--paper)); color: color-mix(in oklab, var(--green) 60%, var(--st-ink)); }

  /* 角色下拉：透明 select 覆盖在胶囊上 */
  &.pick {
    position: relative;
    cursor: pointer;
    transition: filter var(--dur-fast);

    &:hover { filter: brightness(0.96); }
    select { position: absolute; inset: 0; opacity: 0; cursor: pointer; }
  }
}

.chip.pend {
  font-size: 12.5px;
  padding: 3px 10px;
  border-radius: var(--r-pill);
  background: color-mix(in oklab, var(--yellow) 16%, var(--paper));
  color: color-mix(in oklab, var(--yellow) 55%, var(--st-ink));
}

.u-table { transition: opacity var(--dur-fast); }
.u-table.busy { opacity: 0.6; }
.row-enter-active, .row-leave-active { transition: opacity var(--dur) var(--ease-out), transform var(--dur) var(--ease-out); }
.row-enter-from, .row-leave-to { opacity: 0; transform: translateY(6px); }

.empty { padding: 36px 0; text-align: center; color: var(--st-ink-3); font-size: 14px; }

/* ---------- 邀请记录 ---------- */
.inv-card { margin-top: 24px; }

.inv {
  list-style: none;
  margin: 8px 0 0;
  padding: 0;

  li {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 12px 0;

    & + li { box-shadow: 0 -1px 0 var(--line-2); }
    &.used, &.expired { opacity: 0.55; }
  }

  .tx { flex: 1; }

  b {
    display: flex;
    align-items: baseline;
    gap: 10px;
    font-weight: 500;

    em { font-style: normal; font-size: 13px; color: var(--st-ink-3); }
  }
}

/* ---------- 弹窗 ---------- */
.st-flabel.gap { margin-top: 16px; }
.st-flabel em { margin-left: 6px; font-style: normal; color: var(--st-ink-3); font-weight: 400; }
.two { display: grid; grid-template-columns: 120px 1fr; gap: 12px; }
.send { margin-top: 16px; }
.send.dim { opacity: 0.5; pointer-events: none; }

.link-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 10px 10px 14px;
  border-radius: var(--r-md);
  background: var(--well);
  box-shadow: 0 0 0 1px var(--line-2);

  code {
    flex: 1;
    min-width: 0;
    font: 12.5px var(--font-mono);
    word-break: break-all;
    color: var(--st-ink-2);
  }
}

@media (max-width: 1180px) {
  .view { padding: 28px 32px 64px; }
  .groups { grid-template-columns: 1fr; }
  .u-stats { --n: 3; }
  .tools { flex-direction: column; align-items: stretch; .search { width: auto; } }
}

@media (prefers-reduced-motion: reduce) {
  .fold { transition: none; }
}
</style>
