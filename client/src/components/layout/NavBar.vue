<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import ThemeSwitcher from './ThemeSwitcher.vue';
import UiIcon from '../ui/UiIcon.vue';
import { useAuthStore } from '../../stores/auth';
import { useConfigStore } from '../../stores/config';

const { t } = useI18n();
const route = useRoute();
const auth = useAuthStore();
const config = useConfigStore();
/** 用户系统开放时，未登录访客看到「登录」入口 */
const showLogin = computed(() => !auth.loggedIn && !!config.cfg.users?.enabled);
const loginTo = computed(() => ({ path: '/account/login', query: route.path.startsWith('/account') ? {} : { next: route.fullPath } }));
const me = computed(() => auth.user);
const meInitial = computed(() => (me.value?.name || me.value?.login || '?').trim()[0]?.toUpperCase() ?? '?');
const drawerOpen = ref(false);
const writeOpen = ref(false);

/* 点击菜单外任意处自动收起 */
const writeWrap = ref<HTMLElement | null>(null);

function onDocPointerDown(e: PointerEvent): void {
  if (!writeOpen.value) return;
  if (!writeWrap.value?.contains(e.target as Node)) writeOpen.value = false;
}

const links = [
  { to: '/', key: 'nav.home' },
  { to: '/articles', key: 'nav.articles' },
  { to: '/thoughts', key: 'nav.thoughts' },
  { to: '/about', key: 'nav.about' },
];

/** 当前所在栏目（文章详情归入「文章」）；-1 = 不在任何栏目，选中块淡出 */
const activeIndex = computed(() => {
  const p = route.path;
  if (p === '/') return 0;
  return links.findIndex((l, i) => i > 0 && (p === l.to || p.startsWith(`${l.to}/`)));
});

/* ===== 选中块 morph：前缘 .30s ease-out，后缘 .46s spring 延迟 .05s ===== */
const capsule = ref<HTMLElement | null>(null);
const itemEls: HTMLElement[] = [];

function setItem(r: unknown, i: number): void {
  const el = (r as { $el?: HTMLElement } | null)?.$el;
  if (el) itemEls[i] = el;
}
const lift = ref({ l: 0, r: 0, dir: 'to-r' as 'to-r' | 'to-l', on: false, instant: true });

function place(instant = false): void {
  const cap = capsule.value;
  const el = itemEls[activeIndex.value];
  if (!cap || !el) {
    lift.value = { ...lift.value, on: false };
    return;
  }
  const l = el.offsetLeft;
  const r = cap.clientWidth - l - el.offsetWidth;
  const wasOn = lift.value.on;
  lift.value = {
    l,
    r,
    dir: l < lift.value.l ? 'to-l' : 'to-r',
    on: true,
    // 从「无选中」进入时不 morph，原地淡入
    instant: instant || !wasOn,
  };
  if (lift.value.instant) {
    requestAnimationFrame(() => requestAnimationFrame(() => { lift.value = { ...lift.value, instant: false }; }));
  }
}

watch(activeIndex, () => void nextTick(() => place()));

let ro: ResizeObserver | null = null;
onMounted(() => {
  document.addEventListener('pointerdown', onDocPointerDown);
  place(true);
  // 字体加载完成 / 容器尺寸变化后重新测量，不做动画
  void document.fonts?.ready.then(() => place(true));
  if (capsule.value && 'ResizeObserver' in window) {
    ro = new ResizeObserver(() => place(true));
    ro.observe(capsule.value);
  }
});
onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', onDocPointerDown);
  ro?.disconnect();
});
</script>

<template>
  <!-- PC：居中毛玻璃胶囊导航 + 外观（色盘 / 深浅）+ 登录态写作 / 后台入口 -->
  <header class="nav-pc">
    <nav ref="capsule" class="capsule" :aria-label="t('shell.mainNav')">
      <span
        class="lift"
        :class="[lift.dir, { on: lift.on, instant: lift.instant }]"
        :style="{ left: `${lift.l}px`, right: `${lift.r}px` }"
        aria-hidden="true"
      />
      <router-link
        v-for="(link, i) in links"
        :key="link.to"
        :ref="(r) => setItem(r, i)"
        :to="link.to"
        class="capsule-link"
        :class="{ on: activeIndex === i }"
        :aria-current="activeIndex === i ? 'page' : undefined"
      >
        {{ t(link.key) }}
      </router-link>
    </nav>
    <ThemeSwitcher class="nav-theme" />

    <!-- 写作入口：文章 / 随想（协作作者只有文章） -->
    <div v-if="auth.staff" ref="writeWrap" class="write-wrap">
      <button
        class="round-btn"
        :class="{ open: writeOpen }"
        :title="t('write.menuTitle')"
        :aria-expanded="writeOpen"
        @click="writeOpen = !writeOpen"
      >
        <UiIcon name="pen" class="s" />
      </button>
      <transition name="pop">
        <div v-if="writeOpen" class="write-menu" @click="writeOpen = false">
          <router-link to="/write/post" class="wm-item">
            <UiIcon name="doc" class="s" />
            <span>{{ t('write.menuPost') }}</span>
          </router-link>
          <router-link v-if="auth.isAdmin" to="/write/note" class="wm-item">
            <UiIcon name="bubble" class="s" />
            <span>{{ t('write.menuNote') }}</span>
          </router-link>
        </div>
      </transition>
    </div>

    <!-- 后台快捷入口（站长 / 作者） -->
    <router-link v-if="auth.staff" to="/admin" class="round-btn" :title="t('admin.loginTitle')">
      <UiIcon name="gear" class="s" />
    </router-link>
    <!-- 我的账号 / 登录 -->
    <router-link v-if="auth.loggedIn" to="/account" class="round-btn me-btn" :title="t('a11y.myAccount')" :aria-label="t('a11y.myAccount')">
      <img v-if="auth.avatar" :src="auth.avatar" alt="" referrerpolicy="no-referrer" />
      <span v-else-if="me">{{ meInitial }}</span>
      <UiIcon v-else name="user" class="s" />
    </router-link>
    <router-link v-else-if="showLogin" :to="loginTo" class="login-pill">{{ t('account.title_login') }}</router-link>
  </header>

  <!-- 窄屏兜底（桌面外壳在 768px 以下几乎不会出现，移动端有独立外壳）：悬浮顶栏 + 汉堡侧栏 -->
  <header class="nav-mobile">
    <span class="site-name">{{ t('common.siteName') }}</span>
    <button
      class="hamburger"
      :class="{ open: drawerOpen }"
      aria-label="菜单"
      @click="drawerOpen = !drawerOpen"
    >
      <span /><span /><span />
    </button>
  </header>

  <transition name="drawer">
    <aside v-if="drawerOpen" class="drawer">
      <nav class="drawer-links">
        <router-link
          v-for="(link, i) in links"
          :key="link.to"
          :to="link.to"
          class="drawer-link"
          :class="{ on: activeIndex === i }"
          @click="drawerOpen = false"
        >
          {{ t(link.key) }}
        </router-link>
      </nav>

      <!-- 管理员：写作与控制中心入口 -->
      <nav v-if="auth.staff" class="drawer-admin">
        <router-link to="/write/post" class="drawer-link" @click="drawerOpen = false">
          {{ t('write.menuPost') }}
        </router-link>
        <router-link to="/write/note" class="drawer-link" @click="drawerOpen = false">
          {{ t('write.menuNote') }}
        </router-link>
        <router-link to="/admin" class="drawer-link" @click="drawerOpen = false">
          {{ t('admin.loginTitle') }}
        </router-link>
      </nav>

      <div class="drawer-theme">
        <ThemeSwitcher />
      </div>
    </aside>
  </transition>
  <transition name="fade">
    <div v-if="drawerOpen" class="drawer-mask" @click="drawerOpen = false" />
  </transition>
</template>

<style scoped lang="scss">
/* ===== 共用：毛玻璃胶囊面（中性描边 + 中性阴影，无主色描边 / 外发光） ===== */
%glass {
  background: color-mix(in oklab, var(--bg) 82%, transparent);
  backdrop-filter: blur(20px) saturate(170%);
  -webkit-backdrop-filter: blur(20px) saturate(170%);
  box-shadow: inset 0 0 0 0.5px var(--line-2), 0 8px 24px -16px rgb(0 0 0 / 0.5);

  :root[data-mode='light'] & {
    background: color-mix(in oklab, var(--bg) 78%, rgb(255 255 255 / 0.5));
    box-shadow: inset 0 0 0 0.5px var(--line-2), 0 1px 2px rgb(16 24 40 / 0.04), 0 10px 28px -18px rgb(16 24 40 / 0.35);
  }

  /* 不支持或未渲染模糊（无 GPU、省电模式）时退回近实底，保证压在正文上的导航文字可读 */
  @supports not ((backdrop-filter: blur(1px)) or (-webkit-backdrop-filter: blur(1px))) {
    background: color-mix(in oklab, var(--bg) 96%, transparent);
    :root[data-mode='light'] & { background: color-mix(in oklab, var(--bg) 94%, #fff); }
  }
}

/* ===== PC ===== */
/* 顶栏圆角随全局 --r-base 走：默认（10）时 44px 高度恰为胶囊，基准调小则变圆角矩形 */
.nav-pc {
  --nav-r: calc(var(--r-base) * 2.2px);
  --nav-r-in: max(2px, calc(var(--nav-r) - 4px));

  position: fixed;
  top: 18px;
  left: 0;
  right: 0;
  z-index: 100;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  pointer-events: none;

  > * { pointer-events: auto; }
}

.capsule {
  @extend %glass;

  position: relative;
  display: flex;
  align-items: center;
  height: 44px;
  padding: 4px;
  border-radius: var(--nav-r);
}

/* 选中块：抬升 + 轻染，在项之间 morph（左右两缘分别过渡，形成拉伸） */
.lift {
  position: absolute;
  top: 4px;
  bottom: 4px;
  border-radius: var(--nav-r-in);
  background: var(--lift);
  box-shadow: var(--lift-shadow);
  opacity: 0;
  pointer-events: none;
  transition:
    right 0.3s var(--ease-out),
    left 0.46s var(--ease-spring) 0.05s,
    opacity var(--dur) var(--ease-out),
    background-color var(--dur),
    box-shadow var(--dur);

  &.to-l {
    transition:
      left 0.3s var(--ease-out),
      right 0.46s var(--ease-spring) 0.05s,
      opacity var(--dur) var(--ease-out),
      background-color var(--dur),
      box-shadow var(--dur);
  }

  &.instant { transition: opacity var(--dur) var(--ease-out); }
  &.on { opacity: 1; }
}

.capsule-link {
  position: relative;
  z-index: 1;
  height: 100%;
  display: inline-flex;
  align-items: center;
  padding: 0 18px 1px;
  border-radius: var(--nav-r-in);
  font-size: 14px;
  font-weight: 500;
  color: var(--text-2);
  transition: color var(--dur) var(--ease-out), transform var(--dur-fast) var(--ease-spring);

  &:hover { color: var(--text); }
  &:active { transform: scale(0.96); }
  &.on { color: var(--lift-fg); }

  &:focus-visible {
    outline: none;
    box-shadow: var(--focus);
  }
}

/* 圆形按钮：写作 / 后台（中性玻璃面，hover 只改明度） */
.round-btn {
  @extend %glass;

  width: 44px;
  height: 44px;
  display: grid;
  place-items: center;
  border: 0;
  border-radius: min(50%, var(--nav-r));
  color: var(--text-2);
  transition: color var(--dur-fast), background-color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  &:hover,
  &.open { color: var(--text); background: color-mix(in oklab, var(--bg) 60%, var(--fill-3)); }
  &:active { transform: scale(0.94); }

  &:focus-visible {
    outline: none;
    box-shadow: var(--focus);
  }
}

/* 我的账号：头像填满圆钮，无头像时显示首字 */
.me-btn {
  overflow: hidden;
  /* 头像内缩一圈，露出与其它圆按钮相同的毛玻璃描边，视觉大小一致；圆角同心收小 */
  padding: 3px;

  img { width: 100%; height: 100%; object-fit: cover; border-radius: min(50%, calc(var(--nav-r) - 3px)); }
  span { font: 600 15px var(--font-serif); color: var(--text); }
}

.login-pill {
  @extend %glass;

  height: 44px;
  display: inline-flex;
  align-items: center;
  padding: 0 18px;
  border-radius: var(--nav-r, var(--r-pill));
  color: var(--text);
  font-size: 14px;
  font-weight: 500;
  transition: background-color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  &:hover { background: color-mix(in oklab, var(--bg) 60%, var(--fill-3)); }
  &:active { transform: scale(0.96); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

.write-wrap { position: relative; }

.write-menu {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  z-index: 120;
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 160px;
  padding: 6px;
  border-radius: var(--r-lg);
  background: color-mix(in oklab, var(--surface) 92%, transparent);
  backdrop-filter: blur(24px) saturate(170%);
  -webkit-backdrop-filter: blur(24px) saturate(170%);
  box-shadow: var(--shadow-pop);
  transform-origin: top right;

  :root[data-mode='light'] & { background: rgb(255 255 255 / 0.92); }
}

.wm-item {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 38px;
  padding: 0 12px;
  border-radius: var(--r-sm);
  font-size: 13.5px;
  font-weight: 500;
  color: var(--text-2);
  transition: background-color var(--dur-fast), color var(--dur-fast);

  &:hover { background: var(--fill-2); color: var(--text); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

.pop-enter-active { transition: opacity 0.2s var(--ease-out), transform 0.35s var(--ease-spring); }
.pop-leave-active { transition: opacity 0.15s ease, transform 0.15s ease; }
.pop-enter-from, .pop-leave-to { opacity: 0; transform: translateY(-6px) scale(0.97); }

/* ===== 窄屏兜底：悬浮顶栏 + 汉堡侧栏 ===== */
.nav-mobile {
  @extend %glass;

  display: none;
  position: fixed;
  top: 12px;
  left: 12px;
  right: 12px;
  z-index: 100;
  padding: 10px 16px;
  border-radius: var(--r-lg);
  justify-content: space-between;
  align-items: center;
}

.site-name {
  font-family: var(--font-serif);
  font-weight: 700;
  font-size: 17px;
}

.hamburger {
  width: 34px;
  height: 34px;
  border: none;
  background: none;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 5px;

  span {
    display: block;
    width: 20px;
    height: 2px;
    border-radius: var(--r-pill);
    background: var(--text);
    transition: transform var(--dur) var(--ease-spring), opacity var(--dur-fast);
  }

  &.open span:nth-child(1) { transform: translateY(7px) rotate(45deg); }
  &.open span:nth-child(2) { opacity: 0; }
  &.open span:nth-child(3) { transform: translateY(-7px) rotate(-45deg); }
}

.drawer {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: min(78vw, 320px);
  z-index: 99;
  background: var(--surface);
  box-shadow: var(--shadow-pop);
  padding: 84px 22px 22px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.drawer-links,
.drawer-admin {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.drawer-admin {
  padding-top: 14px;
  box-shadow: inset 0 0.5px 0 var(--line);
}

.drawer-link {
  font-size: 16px;
  font-weight: 500;
  padding: 12px 16px;
  border-radius: var(--r-md);
  color: var(--text-2);
  transition: background-color var(--dur-fast), color var(--dur-fast);

  &:hover { background: var(--fill); color: var(--text); }

  &.on {
    background: var(--lift);
    box-shadow: var(--lift-shadow);
    color: var(--lift-fg);
  }
}

.drawer-theme {
  display: flex;
  justify-content: center;
  margin-top: auto;
  padding-bottom: 8px;
}

.drawer-mask {
  position: fixed;
  inset: 0;
  z-index: 98;
  background: var(--scrim);
}

.drawer-enter-active, .drawer-leave-active { transition: transform var(--dur) var(--ease-sheet); }
.drawer-enter-from, .drawer-leave-to { transform: translateX(100%); }
.fade-enter-active, .fade-leave-active { transition: opacity var(--dur-fast); }
.fade-enter-from, .fade-leave-to { opacity: 0; }

@media (max-width: 767px) {
  .nav-pc { display: none; }
  .nav-mobile { display: flex; }
}
</style>
