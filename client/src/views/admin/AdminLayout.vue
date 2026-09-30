<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { canEnterAdmin, staffHome, useAuthStore } from '../../stores/auth';
import { useConfigStore } from '../../stores/config';
import { useThemeStore } from '../../stores/theme';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import ToastHost from './studio/ToastHost.vue';
import { refreshCounts, studio, themeTransition } from './studio/state';
import '@fontsource/noto-serif-sc/400.css';
import '@fontsource/noto-serif-sc/600.css';

/**
 * 桌面后台外壳 · Studio：浮动纸面侧栏 + 主区大纸面。
 * 写文章（admin-write-post）进入沉浸模式：侧栏退场、纸面铺满。
 */
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const auth = useAuthStore();
const config = useConfigStore();
const theme = useThemeStore();

interface NavItem {
  name: string;
  icon: string;
  key: string;
  count?: () => number;
  hot?: boolean;
}

const GROUPS: { label?: string; items: NavItem[] }[] = [
  {
    items: [
      { name: 'admin-today', icon: 'sun', key: 'today' },
      { name: 'admin-posts', icon: 'doc', key: 'posts', count: () => studio.posts },
      { name: 'admin-notes', icon: 'feather', key: 'notes', count: () => studio.notes },
      { name: 'admin-media', icon: 'image', key: 'media' },
      { name: 'admin-identity', icon: 'user', key: 'identity' },
      { name: 'admin-about', icon: 'layers', key: 'about' },
    ],
  },
  {
    label: 'site',
    items: [
      { name: 'admin-appearance', icon: 'palette', key: 'appearance' },
      { name: 'admin-settings', icon: 'settings', key: 'settings' },
      { name: 'admin-apikeys', icon: 'key', key: 'apikeys' },
      { name: 'admin-data', icon: 'archive', key: 'data' },
    ],
  },
  {
    label: 'readers',
    items: [
      { name: 'admin-comments', icon: 'message', key: 'comments' },
      { name: 'admin-users', icon: 'users', key: 'users' },
    ],
  },
];

/** 按角色裁剪：作者只看到自己的写作页，空分组整组隐去 */
const groups = computed(() =>
  GROUPS.map((g) => ({ ...g, items: g.items.filter((it) => canEnterAdmin(auth.role, it.name)) })).filter((g) => g.items.length),
);
const home = computed(() => staffHome(auth.role ?? 'admin'));

/** 写作页归属到对应列表高亮 */
const ALIAS: Record<string, string> = {
  'admin-write-post': 'admin-posts',
  'admin-write-note': 'admin-notes',
};
const active = computed(() => {
  const n = String(route.name ?? '');
  return ALIAS[n] ?? n;
});
const immersive = computed(() => route.name === 'admin-write-post');

/* ===== 侧栏选中块：抬升 + 轻染，在项之间 morph（前缘快、后缘慢） ===== */
const navEl = ref<HTMLElement | null>(null);
const ind = ref({ top: 0, bottom: 0, down: true, show: false });
/** 尺寸变化引起的重定位不播 morph，直接落位 */
const snap = ref(false);

function moveInd(): void {
  const nav = navEl.value;
  const a = nav?.querySelector<HTMLElement>(`a[data-name="${active.value}"]`);
  if (!nav || !a) {
    ind.value.show = false;
    return;
  }
  const top = a.offsetTop;
  ind.value = {
    top,
    bottom: nav.clientHeight - top - a.offsetHeight,
    down: top >= ind.value.top,
    show: true,
  };
}
watch(active, () => void nextTick(moveInd));

/*
 * 选中块用 top + bottom 定位，bottom 依赖导航总高度。进出沉浸写作时侧栏收起/展开、
 * 计数加载、字体就绪都会改变导航高度，而当前项可能不变（写作页归属「文章」），
 * 所以要在导航尺寸变化时无动画地重新落位，否则会残留旧的 bottom 把块拉长。
 */
function reflowInd(): void {
  snap.value = true;
  moveInd();
  requestAnimationFrame(() => requestAnimationFrame(() => (snap.value = false)));
}
let navRo: ResizeObserver | null = null;
watch(navEl, (el) => {
  navRo?.disconnect();
  if (!el) return;
  navRo = new ResizeObserver(reflowInd);
  navRo.observe(el);
});
watch(() => route.name, (n, o) => {
  if (n === 'admin-write-post' || o === 'admin-write-post') void nextTick(reflowInd);
});

/* ===== 主题：深浅 + 色盘（View Transition 圆形扩散） ===== */
const palettes = computed(() =>
  theme.allPalettes.map((p) => ({ id: p.id, color: p.light.primary, name: p.nameKey })),
);

function setMode(mode: 'light' | 'dark', e: MouseEvent): void {
  if (theme.mode === mode) return;
  themeTransition(e, () => theme.setMode(mode));
}

function setPalette(id: string, e: MouseEvent): void {
  if (theme.paletteId === id) return;
  themeTransition(e, () => theme.setPalette(id, true));
}

/* ===== 账号 ===== */
/* 当前登录用户；站长未设头像 / 名字时回落到站点身份 */
const displayName = computed(() => auth.user?.name || config.cfg.about?.name || 'Myself');
const avatar = computed(() => auth.user?.avatar || config.cfg.about?.avatar || config.cfg.site.logo || '/favicon-64.png');
const roleLabel = computed(() => t(`studio.users.r_${auth.role ?? 'admin'}`));

async function logout(): Promise<void> {
  await auth.logout();
  void router.push('/admin/login');
}

/* ===== 页面切换：离场与入场同时进行（不播路由遮罩） ===== */
const scroller = ref<HTMLElement | null>(null);

/**
 * 离场页脱离文档流叠放在原视觉位置（绝对定位，按当前滚动量上移），
 * 入场页从顶部正常排版，二者交叉播放。Vue 对同一 Transition 内的换页先触发 leave 再触发 enter，
 * 所以这里读到的滚动量仍是旧页的。
 */
function pinLeaving(el: Element): void {
  const node = el as HTMLElement;
  const top = node.offsetTop - (scroller.value?.scrollTop ?? 0);
  node.style.position = 'absolute';
  node.style.top = `${top}px`;
  node.style.left = '0';
  node.style.right = '0';
}

/** 入场页：回到顶部 */
function resetScroll(): void {
  if (scroller.value) scroller.value.scrollTop = 0;
}

/* 键盘：N 写随想（不在输入状态时） */
function onKey(e: KeyboardEvent): void {
  const el = document.activeElement as HTMLElement | null;
  const typing = !!el && (/INPUT|TEXTAREA|SELECT/.test(el.tagName) || el.isContentEditable);
  if (typing || e.ctrlKey || e.metaKey || e.altKey) return;
  if (e.key === 'n' && !immersive.value && auth.isAdmin) {
    e.preventDefault();
    if (route.name === 'admin-today') window.dispatchEvent(new CustomEvent('studio:compose'));
    else void router.push({ name: 'admin-write-note' });
  }
}

onMounted(() => {
  document.documentElement.dataset.studio = '';
  void refreshCounts();
  void nextTick(moveInd);
  void document.fonts?.ready.then(moveInd);
  window.addEventListener('resize', moveInd);
  window.addEventListener('keydown', onKey);
});

onBeforeUnmount(() => {
  delete document.documentElement.dataset.studio;
  window.removeEventListener('resize', moveInd);
  window.removeEventListener('keydown', onKey);
  navRo?.disconnect();
});
</script>

<template>
  <div class="studio app" :class="{ immersive }">
    <aside class="side" :aria-hidden="immersive">
      <router-link :to="home" class="brand">
        <img :src="config.cfg.site.logo || '/favicon-64.png'" alt="" draggable="false" />
        <div>
          <b>{{ config.cfg.site.title || 'Myself' }}</b>
          <small>{{ t('studio.brandSub') }}</small>
        </div>
      </router-link>

      <router-link :to="{ name: 'admin-write-post' }" class="st-btn p write-btn">
        <SIcon name="pen" :size="16" />{{ t('studio.nav.write') }}
      </router-link>

      <nav ref="navEl" class="nav">
        <div
          class="nav-ind"
          :class="{ show: ind.show, up: !ind.down, snap }"
          :style="{ top: `${ind.top}px`, bottom: `${ind.bottom}px` }"
        />
        <template v-for="(g, gi) in groups" :key="gi">
          <div v-if="g.label" class="nav-label">{{ t(`studio.nav.${g.label}`) }}</div>
          <router-link
            v-for="item in g.items"
            :key="item.name"
            :to="{ name: item.name }"
            :data-name="item.name"
            :class="{ on: active === item.name }"
          >
            <SIcon :name="item.icon" />
            <span>{{ t(`studio.nav.${item.key}`) }}</span>
            <span v-if="item.count && item.count()" class="cnt mono">{{ item.count() }}</span>
          </router-link>
        </template>
      </nav>

      <div class="side-foot">
        <div class="me">
          <router-link class="who-link" to="/account" :title="t('studio.account')">
            <span class="avatar"><img :src="avatar" alt="" /></span>
            <div class="who">
              <b>{{ displayName }}</b>
              <small>{{ roleLabel }}</small>
            </div>
          </router-link>
          <router-link class="ext" to="/" :title="t('studio.viewSite')">
            <SIcon name="external" :size="16" />
          </router-link>
          <button type="button" class="ext" :title="t('studio.logout')" @click="logout">
            <SIcon name="logout" :size="16" />
          </button>
        </div>
        <div class="theme-row">
          <div class="mode-tg" :class="{ dark: theme.mode === 'dark' }">
            <span class="k" />
            <button type="button" :class="{ on: theme.mode === 'light' }" :title="t('studio.light')" @click="setMode('light', $event)">
              <SIcon name="sun" :size="16" />
            </button>
            <button type="button" :class="{ on: theme.mode === 'dark' }" :title="t('studio.dark')" @click="setMode('dark', $event)">
              <SIcon name="moon" :size="16" />
            </button>
          </div>
          <div class="seasons">
            <button
              v-for="p in palettes.slice(0, 5)"
              :key="p.id"
              type="button"
              :class="{ on: theme.paletteId === p.id }"
              :style="{ '--c': p.color }"
              :title="p.name"
              @click="setPalette(p.id, $event)"
            ><i /></button>
          </div>
        </div>
      </div>
    </aside>

    <main class="paper" :data-view="String(route.name ?? '')">
      <div class="today-glow" :class="{ on: route.name === 'admin-today' }" />
      <div ref="scroller" class="scroll">
        <router-view v-slot="{ Component, route: r }">
          <transition name="st-view" @before-leave="pinLeaving" @before-enter="resetScroll">
            <component :is="Component" :key="String(r.name)" />
          </transition>
        </router-view>
      </div>
    </main>

    <ToastHost />
  </div>
</template>

<style scoped lang="scss">
.app {
  display: grid;
  grid-template-columns: 248px minmax(0, 1fr);
  gap: 14px;
  padding: 14px;
  height: 100vh;
  overflow: hidden;
  background: var(--desk-glow), var(--desk);
  font-size: 15px;
  line-height: 1.6;
  transition:
    grid-template-columns var(--dur-slow) var(--ease-out),
    gap var(--dur-slow) var(--ease-out),
    padding var(--dur-slow) var(--ease-out),
    background-color var(--dur-slow) var(--ease-out);

  &.immersive {
    grid-template-columns: 0 minmax(0, 1fr);
    gap: 0;
    padding: 0;

    .side {
      opacity: 0;
      transform: translateX(-24px) scale(0.98);
      pointer-events: none;
    }

    .paper { border-radius: 0; box-shadow: none; }
  }
}

/* ---------- 侧栏 ---------- */
.side {
  position: relative;
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
  border-radius: var(--r-lg);
  background: var(--side);
  box-shadow: var(--sh-side);
  backdrop-filter: blur(20px) saturate(1.2);
  padding: 18px 12px 12px;
  overflow: hidden;
  transition: opacity var(--dur) var(--ease-out), transform var(--dur-slow) var(--ease-out);
}

/* 退出沉浸写作：侧栏列宽先展开一段再显形，避免窄列里文字折行的中间帧（进入沉浸时立即淡出） */
.app:not(.immersive) .side { transition-delay: 0.22s, 0s; }

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 2px 8px 18px;

  img {
    width: 30px;
    height: 30px;
    border-radius: var(--r-xs);
    box-shadow: 0 4px 10px -4px rgba(10, 20, 40, 0.5);
  }

  b {
    display: block;
    font: 700 17px/1 var(--font-serif);
    letter-spacing: 0.02em;
  }

  small {
    display: block;
    font-size: 11.5px;
    color: var(--st-ink-3);
    margin-top: 4px;
    letter-spacing: 0.08em;
  }
}

.write-btn {
  width: 100%;
  height: 40px;
  margin-bottom: 18px;
  flex: none;
}

.nav {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 2px;
  overflow: auto;
  flex: 1;
  min-height: 0;
  margin: 0 -4px;
  padding: 0 4px;

  a {
    position: relative;
    z-index: 1;
    display: flex;
    align-items: center;
    gap: 11px;
    height: 38px;
    padding: 0 12px;
    border-radius: var(--r-sm);
    color: var(--st-ink-2);
    font-size: 14px;
    flex: none;
    transition: color var(--dur-fast), background var(--dur-fast);

    .st-ic {
      color: var(--st-ink-3);
      transition: color var(--dur-fast), transform var(--dur) var(--ease-spring);
    }

    &:hover { color: var(--st-ink); }
    &:hover:not(.on) { background: var(--hover); }
    &:hover .st-ic { transform: translateY(-1px); }

    /* 当前项：抬升块承载位置（不挂信号点），图标着信号色 */
    &.on {
      color: var(--lift-fg);
      font-weight: 500;

      .st-ic { color: var(--ink); }
    }
  }

  .cnt {
    margin-left: auto;
    font-size: 11.5px;
    color: var(--st-ink-3);
    min-width: 20px;
    height: 20px;
    display: grid;
    place-items: center;
    border-radius: var(--r-sm);
    padding: 0 6px;
  }
}

.nav-ind.snap { transition: none !important; }

.nav-ind {
  position: absolute;
  left: 4px;
  right: 4px;
  border-radius: var(--r-sm);
  background: var(--lift);
  box-shadow: var(--lift-shadow);
  pointer-events: none;
  opacity: 0;
  /* 向下：下缘领先（.30s ease-out），上缘拖尾（.46s spring 延迟 .05s）；向上反之 */
  transition:
    bottom 0.3s var(--ease-out),
    top 0.46s var(--ease-spring) 0.05s,
    background-color var(--dur),
    opacity var(--dur-fast);

  &.up {
    transition:
      top 0.3s var(--ease-out),
      bottom 0.46s var(--ease-spring) 0.05s,
      background-color var(--dur),
      opacity var(--dur-fast);
  }

  &.show { opacity: 1; }
}

.nav-label {
  font-size: 11.5px;
  color: var(--st-ink-4);
  letter-spacing: 0.14em;
  padding: 18px 12px 6px;
  font-weight: 500;
  flex: none;
}

.side-foot {
  border-top: 1px solid var(--line);
  margin: 10px -12px 0;
  padding: 12px 12px 0;
  flex: none;
}

.me {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 4px 10px 6px;

  .who-link {
    display: flex;
    align-items: center;
    gap: 8px;
    flex: 1;
    min-width: 0;
    margin: -4px;
    padding: 4px;
    border-radius: var(--r-md);
    color: inherit;
    transition: background var(--dur-fast);

    &:hover { background: var(--hover); }
  }

  .who { min-width: 0; flex: 1; }

  b {
    display: block;
    font-size: 13.5px;
    font-weight: 500;
    line-height: 1.2;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  small { font-size: 12px; color: var(--st-ink-3); }
}

.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  flex: none;
  overflow: hidden;
  background: linear-gradient(140deg, #1b2a4a, #0b1220);
  box-shadow: 0 0 0 2px var(--paper), 0 0 0 3px var(--line-2);

  img { width: 100%; height: 100%; object-fit: cover; display: block; }
}

.ext {
  color: var(--st-ink-3);
  width: 30px;
  height: 30px;
  display: grid;
  place-items: center;
  border-radius: var(--r-sm);
  flex: none;
  transition: all var(--dur-fast);

  &:hover { background: var(--hover); color: var(--st-ink); }
}

.theme-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 4px 2px;
}

.mode-tg {
  position: relative;
  display: flex;
  background: var(--well);
  border-radius: var(--r-sm);
  padding: 3px;
  box-shadow: 0 0 0 1px var(--line) inset;

  button {
    position: relative;
    z-index: 1;
    width: 30px;
    height: 26px;
    display: grid;
    place-items: center;
    color: var(--st-ink-3);
    border-radius: var(--r-xs);
    transition: color var(--dur);

    &.on { color: var(--lift-fg); }
    &.on .st-ic { color: var(--ink); }
  }

  .k {
    position: absolute;
    top: 3px;
    left: 3px;
    width: 30px;
    height: 26px;
    border-radius: var(--r-xs);
    background: var(--lift);
    box-shadow: var(--lift-shadow);
    transition: transform var(--dur) var(--ease-spring), background-color var(--dur);
  }

  &.dark .k { transform: translateX(30px); }
}

.seasons {
  display: flex;
  gap: 2px;

  button {
    width: 26px;
    height: 26px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    position: relative;
    transition: background-color var(--dur-fast), box-shadow var(--dur-fast);

    i {
      width: 14px;
      height: 14px;
      border-radius: 50%;
      background: var(--c);
      box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.08) inset;
      transition: transform var(--dur) var(--ease-spring);
    }

    &::after {
      content: '';
      position: absolute;
      inset: 3px;
      border-radius: 50%;
      box-shadow: 0 0 0 1.5px var(--st-ink);
      opacity: 0;
      transform: scale(0.6);
      transition: all var(--dur) var(--ease-spring);
    }

    &:hover i { transform: scale(1.12); }

    /* 色盘选中：抬升底 + 中性细环（不用色盘自身颜色描边/发光） */
    &.on {
      background: var(--lift);
      box-shadow: var(--lift-shadow);

      &::after { opacity: 1; transform: scale(1); }
      i { transform: scale(0.72); }
    }
  }
}

/* ---------- 纸面 ---------- */
.paper {
  position: relative;
  min-width: 0;
  min-height: 0;
  border-radius: var(--r-lg);
  background: var(--paper);
  box-shadow: var(--sh-paper);
  overflow: hidden;
  transition:
    background-color var(--dur-slow) var(--ease-out),
    border-radius var(--dur-slow) var(--ease-out),
    box-shadow var(--dur-slow) var(--ease-out);
}

.scroll {
  position: relative;
  height: 100%;
  overflow: auto;
  overscroll-behavior: contain;
}

/* 「今天」右上角的窗光：品牌光影母题（门缝/窗格光），属于允许发光的品牌装饰，非控件 */
.today-glow {
  position: absolute;
  right: 0;
  top: 0;
  width: 520px;
  height: 420px;
  pointer-events: none;
  background:
    radial-gradient(60% 60% at 90% 0%, color-mix(in oklab, var(--primary) 14%, transparent), transparent 70%),
    radial-gradient(40% 40% at 100% 10%, color-mix(in oklab, var(--yellow) 12%, transparent), transparent 70%);
  opacity: 0;
  transition: opacity var(--dur-slow);

  &.on { opacity: 1; }

  &::after {
    content: '';
    position: absolute;
    inset: 0;
    background: repeating-linear-gradient(118deg, transparent 0 44px, color-mix(in oklab, var(--yellow) 16%, transparent) 44px 78px);
    mask: radial-gradient(75% 80% at 100% 0, #000, transparent 72%);
    filter: blur(5px);
  }
}

:root[data-mode='dark'] .today-glow::after { opacity: 0.45; }

/* ---------- 页面过渡：淡入 + 轻微上移 ---------- */
.st-view-enter-active { animation: view-in var(--dur-slow) var(--ease-out) both; }
/* 离场页叠放在上层淡出（pinLeaving 已设绝对定位），入场页同时在下层升起 */
.st-view-leave-active {
  z-index: 2;
  pointer-events: none;
  animation: view-out var(--dur) var(--ease-out) both;
}
/* 沉浸写作页离场：纸面轻缩淡出，与列表入场、侧栏回位同时进行 */
.editor.st-view-leave-active { animation-name: view-out-zoom; }

@keyframes view-in { from { opacity: 0; transform: translateY(14px); filter: blur(4px); } }
@keyframes view-out { to { opacity: 0; transform: translateY(-6px); filter: blur(3px); } }
@keyframes view-out-zoom { to { opacity: 0; transform: scale(0.97); } }

@media (max-width: 860px) {
  .app { grid-template-columns: 1fr; }
  .side { display: none; }
}
</style>
