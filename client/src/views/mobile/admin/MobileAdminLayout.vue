<script setup lang="ts">
/**
 * 移动端后台外壳（390–430 宽，iOS 原生质感）：
 * - V2 贴底底栏 + 中央凸起「+」（旋转成 × 并弹出 action sheet）
 * - 路由过渡按方向感知：底栏标签左右滑、二级页右侧推入、编辑器自下而上
 * - 写随想 composer 为全局底部 sheet，拖到全屏时整个后台舞台缩小后退
 * - 灵动岛式轻提示；未做移动版的后台页套原生导航承载桌面组件
 */
import { computed, defineAsyncComponent, onBeforeUnmount, onMounted, ref, watch, type Component } from 'vue';
import { useRoute, useRouter, type RouteLocationNormalizedLoaded } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminChildren } from '../../../router/admin';
import { mobileAdminViews } from '../../../router/mobile-admin';
import MaTabBar, { type AdminTab } from '../../../components/mobile-admin/MaTabBar.vue';
import MaActionSheet from '../../../components/mobile-admin/MaActionSheet.vue';
import MaIsland from '../../../components/mobile-admin/MaIsland.vue';
import MaFallbackFrame from '../../../components/mobile-admin/MaFallbackFrame.vue';
import NoteComposerSheet from '../../../components/mobile-admin/NoteComposerSheet.vue';
import { openComposer, shell } from '../../../components/mobile-admin/state';
import { useAuthStore } from '../../../stores/auth';

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const auth = useAuthStore();
/** 协作作者：只有内容（仅文章）与素材两个标签，+ 里不能发随想 */
const isAuthor = computed(() => auth.role === 'author');
const tabs = computed<AdminTab[] | undefined>(() => (isAuthor.value ? ['content', 'media'] : undefined));

/* ---------- 路由 → 底栏标签 / 视图层级 ---------- */
const TAB_OF: Record<string, AdminTab> = {
  'admin-today': 'today',
  'admin-posts': 'content',
  'admin-notes': 'content',
  'admin-write-note': 'content',
  'admin-media': 'media',
  'admin-settings': 'me',
};
const TAB_ROUTE: Record<AdminTab, string> = {
  today: 'admin-today',
  content: 'admin-posts',
  media: 'admin-media',
  me: 'admin-settings',
};
const TAB_ORDER: AdminTab[] = ['today', 'content', 'media', 'me'];

const nameOf = (r: RouteLocationNormalizedLoaded): string => String(r.name ?? '');
const isSettingsFull = (r: RouteLocationNormalizedLoaded): boolean =>
  nameOf(r) === 'admin-settings' && !!r.query.full;
const isFallback = (r: RouteLocationNormalizedLoaded): boolean =>
  !mobileAdminViews[nameOf(r)] || isSettingsFull(r);

/** 0 = 底栏页；1 = 二级页（推入）；2 = 全屏编辑器（上推） */
function depthOf(r: RouteLocationNormalizedLoaded): number {
  if (nameOf(r) === 'admin-write-post') return 2;
  return isFallback(r) ? 1 : 0;
}

function viewKey(r: RouteLocationNormalizedLoaded): string {
  const n = nameOf(r);
  if (TAB_OF[n] === 'content') return 'content';
  if (n === 'admin-write-post') return `write-${String(r.query.id ?? 'new')}`;
  return isSettingsFull(r) ? 'settings-full' : n;
}

const activeTab = computed<AdminTab | null>(() => (isFallback(route) ? null : TAB_OF[nameOf(route)] ?? null));
const barHidden = computed(() => depthOf(route) > 0);

const transitionName = ref('ma-none');
watch(
  () => route.fullPath,
  () => {
    shell.actionSheet = false;
  },
);
const removeGuard = router.beforeEach((to, from) => {
  if (!to.meta.admin || !from.meta.admin) return;
  const a = depthOf(from as RouteLocationNormalizedLoaded);
  const b = depthOf(to as RouteLocationNormalizedLoaded);
  if (viewKey(from as RouteLocationNormalizedLoaded) === viewKey(to as RouteLocationNormalizedLoaded)) {
    transitionName.value = 'ma-none';
  } else if (b !== a) {
    const kind = Math.max(a, b) === 2 ? 'up' : 'push';
    transitionName.value = `ma-${kind}-${b > a ? 'in' : 'out'}`;
  } else if (a === 0) {
    const ia = TAB_ORDER.indexOf(TAB_OF[String(from.name)] ?? 'today');
    const ib = TAB_ORDER.indexOf(TAB_OF[String(to.name)] ?? 'today');
    transitionName.value = ib >= ia ? 'ma-tab-fwd' : 'ma-tab-back';
  } else {
    transitionName.value = 'ma-push-in';
  }
});

function selectTab(tab: AdminTab): void {
  const target = TAB_ROUTE[tab];
  if (activeTab.value === tab) {
    /* 再次点击当前标签：回到顶部 */
    document.querySelector<HTMLElement>('.ma-view:last-child .scroll')?.scrollTo({ top: 0, behavior: 'smooth' });
    return;
  }
  void router.push({ name: target });
}

function goBack(): void {
  if (window.history.state?.back) router.back();
  else void router.replace({ name: 'admin-settings' });
}

const FALLBACK_TITLES = ['admin-identity', 'admin-about', 'admin-appearance', 'admin-settings', 'admin-apikeys', 'admin-data', 'admin-comments', 'admin-users'];
function fallbackTitle(r: RouteLocationNormalizedLoaded): string {
  const n = nameOf(r);
  return FALLBACK_TITLES.includes(n) ? t(`mobileAdmin.route.${n}`) : t('mobileAdmin.route.other');
}

/** 设置页「完整版」：直接取后台路由表里登记的桌面组件 */
const settingsLoader = adminChildren.find((c) => c.name === 'admin-settings')?.component;
const DesktopSettings = settingsLoader
  ? defineAsyncComponent(settingsLoader as () => Promise<Component>)
  : null;

/* ---------- 中央 + ---------- */
function toggleFab(): void {
  shell.actionSheet = !shell.actionSheet;
}
function createNote(): void {
  shell.actionSheet = false;
  window.setTimeout(() => openComposer(null), 160);
}
function createPost(): void {
  shell.actionSheet = false;
  void router.push({ name: 'admin-write-post' });
}
function uploadFiles(files: File[]): void {
  shell.actionSheet = false;
  shell.pendingUploads = files;
  if (route.name !== 'admin-media') void router.push({ name: 'admin-media' });
}

/* ---------- 舞台后退（全屏 sheet） ---------- */
const stackAnim = ref(true);
function onStack(v: number, animated: boolean): void {
  shell.stack = v;
  stackAnim.value = animated;
}
const stageStyle = computed(() => {
  const s = shell.stack;
  if (!s) return undefined;
  return {
    transform: `translateY(${s * 14}px) scale(${1 - s * 0.075})`,
    borderRadius: `calc(var(--r-xl) * ${Math.min(1, s * 3)})`,
  };
});

/* 随想编辑入口：/admin/write/note?id= 直接打开 composer */
watch(
  () => [route.name, route.query.id] as const,
  ([n, id]) => {
    if (n !== 'admin-write-note') return;
    const num = typeof id === 'string' && id ? Number(id) : NaN;
    openComposer(Number.isFinite(num) ? num : null);
  },
  { immediate: true },
);
function onComposerClosed(): void {
  if (route.name === 'admin-write-note') void router.replace({ name: 'admin-notes' });
}

function onKey(e: KeyboardEvent): void {
  if (e.key === 'Escape' && shell.actionSheet) shell.actionSheet = false;
}
onMounted(() => {
  window.addEventListener('keydown', onKey);
  document.documentElement.classList.add('ma-lock');
});
onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey);
  removeGuard();
  document.documentElement.classList.remove('ma-lock');
  shell.stack = 0;
});
</script>

<template>
  <div class="ma-root">
    <div
      class="ma-stage"
      :class="{ anim: stackAnim, stacked: shell.stack > 0 }"
      :style="stageStyle"
    >
      <router-view v-slot="{ Component, route: r }" name="mobile">
        <transition :name="transitionName">
          <div :key="viewKey(r)" class="ma-view" :class="`d${depthOf(r)}`">
            <MaFallbackFrame v-if="isFallback(r)" :title="fallbackTitle(r)" @back="goBack">
              <component :is="isSettingsFull(r) && DesktopSettings ? DesktopSettings : Component" />
            </MaFallbackFrame>
            <component :is="Component" v-else />
          </div>
        </transition>
      </router-view>

      <div
        class="ma-dim"
        :class="{ on: shell.actionSheet, anim: stackAnim }"
        :style="shell.stack ? { opacity: Math.min(1, shell.stack) * 0.34 } : undefined"
        @click="shell.actionSheet = false"
      />
      <MaActionSheet
        :open="shell.actionSheet"
        :no-note="isAuthor"
        @note="createNote"
        @post="createPost"
        @files="uploadFiles"
      />
      <MaTabBar
        :active="activeTab"
        :tabs="tabs"
        :fab-open="shell.actionSheet"
        :hidden="barHidden"
        @select="selectTab"
        @fab="toggleFab"
      />
    </div>

    <div id="ma-overlay" class="ma-overlay" />
    <NoteComposerSheet
      v-model:open="shell.composer.open"
      :note-id="shell.composer.id"
      :seq="shell.composer.seq"
      @stack="onStack"
      @closed="onComposerClosed"
    />
    <MaIsland />
  </div>
</template>

<style lang="scss">
@use '../../../components/mobile-admin/ma-tokens.scss';

/* 后台移动端期间锁住文档滚动（各页自带滚动容器） */
html.ma-lock,
html.ma-lock body {
  overflow: hidden;
  overscroll-behavior: none;
}
</style>

<style scoped lang="scss">
.ma-root {
  position: fixed;
  inset: 0;
  z-index: 1;
  overflow: hidden;
  background: var(--bg-deep);
  isolation: isolate;
}

.ma-stage {
  position: absolute;
  inset: 0;
  overflow: hidden;
  background: var(--bg);
  transform-origin: 50% 0;

  &.anim {
    transition: transform 0.5s var(--ease-sheet), border-radius 0.5s var(--ease-sheet);
  }
}

.ma-view {
  position: absolute;
  inset: 0;
  background: var(--bg);
}

.ma-overlay {
  position: absolute;
  inset: 0;
  z-index: 70;
  pointer-events: none;
}

.ma-dim {
  position: absolute;
  inset: 0;
  z-index: 44;
  background: #000;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.4s var(--ease-out);

  &.on {
    opacity: 0.55;
    pointer-events: auto;
  }
}

:root[data-mode='light'] .ma-dim.on {
  opacity: 0.22;
}

/* ---------- 路由过渡 ---------- */
$d: 0.46s;

.ma-tab-fwd-enter-active,
.ma-tab-back-enter-active {
  transition: transform $d var(--ease-sheet), opacity 0.3s var(--ease-out) 0.06s;
}

.ma-tab-fwd-leave-active,
.ma-tab-back-leave-active {
  transition: transform 0.3s var(--ease-out), opacity 0.16s linear;
}

.ma-tab-fwd-enter-from { transform: translateX(14%); opacity: 0; }
.ma-tab-fwd-leave-to { transform: translateX(-6%); opacity: 0; }
.ma-tab-back-enter-from { transform: translateX(-14%); opacity: 0; }
.ma-tab-back-leave-to { transform: translateX(6%); opacity: 0; }

/* 推入：新页从右侧整屏滑入，旧页视差左移 + 压暗 */
.ma-push-in-enter-active,
.ma-push-in-leave-active,
.ma-push-out-enter-active,
.ma-push-out-leave-active {
  transition: transform 0.52s var(--ease-sheet), filter 0.52s var(--ease-sheet), box-shadow 0.52s;
}

.ma-push-in-enter-active,
.ma-push-out-leave-active {
  z-index: 3;
  box-shadow: -24px 0 50px rgba(0, 0, 0, 0.35);
}

.ma-push-in-enter-from,
.ma-push-out-leave-to { transform: translateX(100%); }
.ma-push-in-leave-to,
.ma-push-out-enter-from { transform: translateX(-30%); filter: brightness(0.7); }

/* 编辑器：自下而上，底层轻微缩小后退 */
.ma-up-in-enter-active,
.ma-up-in-leave-active,
.ma-up-out-enter-active,
.ma-up-out-leave-active {
  transition: transform 0.55s var(--ease-sheet), border-radius 0.55s var(--ease-sheet), filter 0.55s;
}

.ma-up-in-enter-active,
.ma-up-out-leave-active {
  z-index: 3;
  overflow: hidden;
}

.ma-up-in-enter-from,
.ma-up-out-leave-to { transform: translateY(100%); border-radius: var(--r-xl) var(--r-xl) 0 0; }
.ma-up-in-leave-to,
.ma-up-out-enter-from { transform: scale(0.94); filter: brightness(0.7); border-radius: var(--r-xl); }

.ma-none-enter-active,
.ma-none-leave-active { transition: none; }
.ma-none-leave-active { display: none; }
</style>
