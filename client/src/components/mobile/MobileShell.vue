<script setup lang="ts">
import {
  KeepAlive,
  computed,
  defineAsyncComponent,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
  type Component,
} from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { mobilePublicViews } from '../../router/mobile-public';
import { attachDrag, clamp, rubber } from './gesture';
import { closeDrawer, openSearch, shell, TAB_ORDER, type TabName } from './shell';
import SideDrawer from './SideDrawer.vue';
import TabBarFloating from './TabBarFloating.vue';
import SearchOverlay from './SearchOverlay.vue';
import IslandToast from './IslandToast.vue';
import { useLoadingStore } from '../../stores/loading';
import { skipNextRouteCover } from '../../router';
import './mobile.scss';

/**
 * 移动端前台外壳：抽屉层 + 舞台层（一级 tab 页 KeepAlive + 详情推入层）+ 悬浮胶囊底栏 + 全屏搜索 + 灵动岛。
 *
 * 首屏与路由切换沿用桌面的 AppLoading / RouteLoading（App.vue 渲染；左边缘手势返回不播路由遮罩）。
 *
 * 导航模型（iOS 式）：
 * - tab 之间：方向感知的横向滑动 + 淡入（按底栏顺序判定左右）。
 * - 进入文章：详情层从右推入，底层 tab 页左移 30% 并变暗；返回反向。左边缘右滑跟手返回，按速度/位移决定是否提交。
 * - 左边缘右滑（或点头像）拉出抽屉，舞台缩小后退 + 圆角 + 变暗，跟手并按速度与位置决定开合。
 * 路由仍是唯一真相：外壳只根据当前路由推导「底层 tab」与「详情 slug」，详情离场期间短暂保留以播完动画。
 */
const route = useRoute();
const router = useRouter();

type Lazy = () => Promise<Component>;
const loaders = mobilePublicViews as Record<string, Lazy>;
const TAB_VIEWS = Object.fromEntries(
  TAB_ORDER.filter((n) => loaders[n]).map((n) => [n, defineAsyncComponent(loaders[n])]),
) as Partial<Record<TabName, Component>>;
const ArticleView = loaders.article ? defineAsyncComponent(loaders.article) : null;

const isTab = (n: unknown): n is TabName => TAB_ORDER.includes(n as TabName);
/** 外壳接管的路由；其余（未来新增页面）回落到命名视图 */
const handled = computed(() => (isTab(route.name) && !!TAB_VIEWS[route.name]) || (route.name === 'article' && !!ArticleView));

/* ---------- 底层 tab 与详情层状态 ---------- */
const baseName = ref<TabName>('home');
const tabQueries = ref<Partial<Record<TabName, Record<string, string>>>>({});
const tabTransition = ref('m-tab-none');
const detailSlug = ref('');
const push = ref(0);
const pushDragging = ref(false);
let popTimer = 0;
let firstRoute = true;

const articlesTag = computed(() => tabQueries.value.articles?.tag ?? '');

function queryOf(): Record<string, string> {
  const out: Record<string, string> = {};
  for (const [k, v] of Object.entries(route.query)) if (typeof v === 'string') out[k] = v;
  return out;
}

function pushIn(slug: string, instant: boolean): void {
  window.clearTimeout(popTimer);
  detailSlug.value = slug;
  if (instant) {
    pushDragging.value = true;
    push.value = 1;
    requestAnimationFrame(() => requestAnimationFrame(() => { pushDragging.value = false; }));
    return;
  }
  void nextTick(() => requestAnimationFrame(() => requestAnimationFrame(() => { push.value = 1; })));
}

function pop(): void {
  push.value = 0;
  window.clearTimeout(popTimer);
  popTimer = window.setTimeout(() => {
    if (push.value === 0) detailSlug.value = '';
  }, 560);
}

watch(
  () => route.fullPath,
  () => {
    const name = route.name;
    if (isTab(name)) {
      const popping = !!detailSlug.value;
      if (name !== baseName.value) {
        const dir = TAB_ORDER.indexOf(name) > TAB_ORDER.indexOf(baseName.value);
        tabTransition.value = firstRoute || popping ? 'm-tab-none' : dir ? 'm-tab-fwd' : 'm-tab-back';
      }
      baseName.value = name;
      tabQueries.value = { ...tabQueries.value, [name]: queryOf() };
      shell.tab = name;
      if (popping) pop();
    } else if (name === 'article') {
      const slug = String(route.params.slug ?? '');
      if (!detailSlug.value || push.value < 1) pushIn(slug, firstRoute);
      else detailSlug.value = slug;
    }
    if (isTab(name) || name === 'article') firstRoute = false;
  },
  { immediate: true },
);

watch([detailSlug, push], () => { shell.pushed = !!detailSlug.value && push.value > 0.5; }, { immediate: true });

function selectTab(name: TabName): void {
  if (name === baseName.value && !detailSlug.value) {
    shell.scrollTopSeq += 1;
    return;
  }
  void router.push({ name, query: tabQueries.value[name] ?? {} });
}

/** 详情返回：有站内历史则后退，直接落地详情页则替换为底层 tab */
function goBack(): void {
  const state = window.history.state as { back?: string | null } | null;
  if (state?.back) router.back();
  else void router.replace({ name: baseName.value, query: tabQueries.value[baseName.value] ?? {} });
}

/* ---------- 边缘手势：抽屉 / 返回 ---------- */
const root = ref<HTMLElement | null>(null);
let detach: (() => void) | null = null;
const drawerW = () => Math.min(272, window.innerWidth * 0.7);

onMounted(() => {
  if (!root.value) return;
  detach = attachDrag(root.value, {
    axis: 'x',
    capture: true,
    down: (e) => {
      if (shell.searchOpen) return null;
      if ((e.target as HTMLElement).closest('.m-layer')) return null;
      const x = e.clientX;
      if (detailSlug.value && push.value > 0.5) return x < 28 ? { mode: 'back' as const, p0: 1 } : null;
      if (shell.dp > 0.5) return { mode: 'drawer' as const, p0: 1 };
      if (x < 24) return { mode: 'drawer' as const, p0: 0 };
      return null;
    },
    begin: (c) => {
      if (c.mode === 'back') pushDragging.value = true;
      else shell.drawerDragging = true;
    },
    move: (c, dx) => {
      if (c.mode === 'back') {
        push.value = clamp(1 - dx / window.innerWidth, 0, 1);
        return;
      }
      const W = drawerW();
      let p = c.p0 + dx / W;
      if (p > 1) p = 1 + rubber((p - 1) * W, 40) / W;
      shell.dp = Math.max(0, p);
    },
    end: (c, vx, _vy, dx) => {
      if (c.mode === 'back') {
        pushDragging.value = false;
        if (vx > 0.35 || dx > window.innerWidth * 0.33) {
          push.value = 0;
          // 手势已给出连续反馈：提交后的返回导航不再盖路由遮罩
          skipNextRouteCover();
          goBack();
        } else {
          push.value = 1;
        }
        return;
      }
      shell.drawerDragging = false;
      const p = c.p0 + dx / drawerW();
      shell.dp = vx > 0.35 ? 1 : vx < -0.35 ? 0 : p > 0.5 ? 1 : 0;
    },
  });
});

/* ---------- 首屏 / 路由遮罩（与桌面共用 AppLoading / RouteLoading） ---------- */
const loading = useLoadingStore();
/* 首屏遮罩开始圆形收缩时舞台同步入场（AppLoading 在 bootDone 后 250ms 开始退场） */
let bootTimer = 0;
watch(
  () => loading.bootDone,
  (done) => {
    if (!done || shell.booted) return;
    bootTimer = window.setTimeout(() => { shell.booted = true; }, 250);
  },
  { immediate: true },
);
/** 幕布在屏上：页面错峰入场暂停，揭幕时再播 */
const covered = computed(() => !shell.booted || loading.routeLoading);

/* 抽屉打开时路由变化（标签跳转等）自动收起 */
watch(() => route.fullPath, () => { if (shell.dp) closeDrawer(); });

/* ---------- 文档级设置：viewport-fit、滚动锁定 ---------- */
const html = document.documentElement;
html.classList.add('m-shell');
const viewport = document.querySelector<HTMLMetaElement>('meta[name="viewport"]');
const viewportBefore = viewport?.content ?? '';
if (viewport && !viewportBefore.includes('viewport-fit')) viewport.content = `${viewportBefore}, viewport-fit=cover`;


onMounted(() => {
  /* 预取其余页面分包，切 tab / 进详情时不再等网络 */
  /* 直接预解析异步包装组件本身（__asyncLoader），首次切换即同步渲染，过渡动画作用在真实页面上 */
  window.setTimeout(() => {
    [...Object.values(TAB_VIEWS), ArticleView].forEach((c) => {
      const loader = (c as { __asyncLoader?: () => Promise<unknown> } | null)?.__asyncLoader;
      void loader?.().catch(() => undefined);
    });
  }, 600);
});

onBeforeUnmount(() => {
  detach?.();
  window.clearTimeout(popTimer);
  window.clearTimeout(bootTimer);
  html.classList.remove('m-shell');
  if (viewport) viewport.content = viewportBefore;
  shell.dp = 0;
  shell.searchOpen = false;
});
</script>

<template>
  <div ref="root" class="m-root" :class="{ booted: shell.booted, covered, 'drawer-open': shell.dp > 0.5 }">
    <SideDrawer :class="{ dragging: shell.drawerDragging }" :style="{ '--m-dp': shell.dp }" />

    <div class="stage" :class="{ dragging: shell.drawerDragging }" :style="{ '--m-dp': shell.dp }">
      <!-- 底层：一级 tab 页 -->
      <div class="base" :class="{ dragging: pushDragging }" :style="{ '--m-push': push }">
        <div class="base-in" :class="{ enter: shell.booted }">
          <template v-if="handled">
            <Transition :name="tabTransition">
              <KeepAlive :max="4">
                <component
                  :is="TAB_VIEWS[baseName]"
                  :key="baseName"
                  v-bind="baseName === 'articles' ? { tag: articlesTag } : {}"
                />
              </KeepAlive>
            </Transition>
          </template>
          <router-view v-else v-slot="{ Component: Fallback }" name="mobile">
            <div class="fallback"><component :is="Fallback" /></div>
          </router-view>
        </div>
      </div>
      <div class="base-dim" :class="{ dragging: pushDragging }" :style="{ '--m-push': push }" />

      <!-- 详情层：iOS push -->
      <div
        v-if="detailSlug && ArticleView"
        class="detail"
        :class="{ dragging: pushDragging }"
        :style="{ '--m-push': push }"
      >
        <Transition name="m-swap-detail">
          <component :is="ArticleView" :key="detailSlug" :slug="detailSlug" @back="goBack" />
        </Transition>
      </div>

      <!-- 悬浮胶囊底栏（详情推入时下沉） -->
      <div class="tabwrap" :class="{ dragging: pushDragging }" :style="{ '--m-push': push }">
        <TabBarFloating :active="baseName" @select="selectTab" @search="openSearch()" />
      </div>

      <div
        class="stage-dim"
        :class="{ dragging: shell.drawerDragging }"
        :style="{ '--m-dp': shell.dp }"
        @click="closeDrawer"
      />
    </div>

    <SearchOverlay />
    <IslandToast />
  </div>
</template>

<style scoped lang="scss">
.m-root {
  position: fixed;
  inset: 0;
  overflow: hidden;
  background: var(--m-drawer-bg);
  isolation: isolate;
}

/* 舞台：抽屉拉开时右移 + 缩小后退 + 圆角 */
/* 幕布（首屏 / 路由遮罩）在屏上时，页面错峰入场停在首帧，揭幕后再播 */
.m-root.covered :deep(.m-in) { animation-play-state: paused; }

.stage {
  --m-dw: min(272px, 70vw);
  position: absolute;
  z-index: 2;
  inset: 0;
  overflow: hidden;
  background: var(--bg);
  border-radius: calc(var(--m-dp) * var(--r-xl) * 1.6);
  transform: translateX(calc(var(--m-dp) * var(--m-dw))) scale(calc(1 - var(--m-dp) * 0.15));
  box-shadow:
    0 30px 60px -12px rgba(0, 0, 0, calc(var(--m-dp) * 0.55)),
    0 0 0 0.5px rgba(255, 255, 255, calc(var(--m-dp) * 0.1));
  transition: --m-dp 0.56s var(--m-ease-drawer);

  &.dragging { transition: none; }
}

.stage-dim {
  position: absolute;
  z-index: 60;
  inset: 0;
  background: #000;
  opacity: calc(var(--m-dp) * 0.38);
  pointer-events: none;
  transition: --m-dp 0.56s var(--m-ease-drawer);

  &.dragging { transition: none; }
  .drawer-open & { pointer-events: auto; }
}

html.m-shell[data-mode='light'] .stage-dim { opacity: calc(var(--m-dp) * 0.12); }

/* 底层 tab 页：详情推入时左移 30% */
.base {
  position: absolute;
  inset: 0;
  overflow: hidden;
  transform: translateX(calc(var(--m-push) * -30%));
  transition: --m-push 0.52s var(--m-ease-sheet);

  &.dragging { transition: none; }
}

.base-in {
  position: absolute;
  inset: 0;

  &.enter { animation: app-in 0.7s var(--ease-out) both; }
}

@keyframes app-in {
  from {
    opacity: 0;
    transform: scale(1.04);
    filter: blur(6px);
  }
}

.base-dim {
  position: absolute;
  z-index: 20;
  inset: 0;
  background: #000;
  opacity: calc(var(--m-push) * 0.32);
  pointer-events: none;
  transition: --m-push 0.52s var(--m-ease-sheet);

  &.dragging { transition: none; }
}

.fallback {
  position: absolute;
  inset: 0;
  overflow-y: auto;
  padding-top: calc(var(--m-safe-t) + 20px);
}

/* 详情层：从右推入 */
.detail {
  position: absolute;
  z-index: 30;
  inset: 0;
  overflow: hidden;
  background: var(--bg);
  transform: translateX(calc((1 - var(--m-push)) * 100%));
  box-shadow: -24px 0 50px rgba(0, 0, 0, calc(var(--m-push) * 0.35));
  transition: --m-push 0.52s var(--m-ease-sheet);

  &.dragging { transition: none; }
}

.tabwrap {
  position: absolute;
  z-index: 25;
  left: 0;
  right: 0;
  bottom: 0;
  height: 120px;
  pointer-events: none;
  transform: translateY(calc(var(--m-push) * 150px));
  transition: --m-push 0.52s var(--m-ease-sheet);

  &.dragging { transition: none; }
}

/* tab 切换：方向感知横移 + 淡入（进场自右/左 28%，离场反向 18%） */
.m-tab-fwd-enter-active,
.m-tab-fwd-leave-active,
.m-tab-back-enter-active,
.m-tab-back-leave-active {
  transition: transform 0.46s var(--m-ease-sheet), opacity 0.34s var(--ease-out);
}

.m-tab-fwd-leave-active,
.m-tab-back-leave-active {
  pointer-events: none;
  transition: transform 0.46s var(--m-ease-sheet), opacity 0.42s ease;
}

.m-tab-fwd-enter-from { transform: translateX(28%); opacity: 0; }
.m-tab-fwd-leave-to { transform: translateX(-18%); opacity: 0; }
.m-tab-back-enter-from { transform: translateX(-28%); opacity: 0; }
.m-tab-back-leave-to { transform: translateX(18%); opacity: 0; }

/* 详情内换篇（下一篇）：交叉淡入 + 轻微上浮 */
.m-swap-detail-enter-active { transition: opacity 0.4s var(--ease-out), transform 0.5s var(--ease-out); }
.m-swap-detail-leave-active { transition: opacity 0.2s ease; position: absolute; inset: 0; }
.m-swap-detail-enter-from { opacity: 0; transform: translateY(24px); }
.m-swap-detail-leave-to { opacity: 0; }
</style>
