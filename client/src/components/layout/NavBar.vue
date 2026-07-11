<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import ThemeSwitcher from './ThemeSwitcher.vue';
import { useAuthStore } from '../../stores/auth';

const { t } = useI18n();
const auth = useAuthStore();
const drawerOpen = ref(false);
const writeOpen = ref(false);

const links = [
  { to: '/', key: 'nav.home' },
  { to: '/articles', key: 'nav.articles' },
  { to: '/thoughts', key: 'nav.thoughts' },
  { to: '/about', key: 'nav.about' },
];
</script>

<template>
  <!-- PC：毛玻璃胶囊导航 -->
  <header class="nav-pc">
    <nav class="capsule">
      <router-link
        v-for="link in links"
        :key="link.to"
        :to="link.to"
        class="capsule-link"
        active-class="active"
        :exact-active-class="link.to === '/' ? 'active' : ''"
      >
        {{ t(link.key) }}
      </router-link>
    </nav>
    <ThemeSwitcher class="nav-theme" />

    <!-- 写作入口：文章 / 随想 -->
    <div v-if="auth.loggedIn" class="write-wrap">
      <button class="admin-dot" :title="t('write.menuTitle')" @click="writeOpen = !writeOpen">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 20h9" />
          <path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z" />
        </svg>
      </button>
      <transition name="pop">
        <div v-if="writeOpen" class="write-menu" @click="writeOpen = false">
          <router-link to="/write/post" class="wm-item">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M6 3h9l5 5v13H6zM14 3v6h6" />
            </svg>
            <span>{{ t('write.menuPost') }}</span>
          </router-link>
          <router-link to="/write/note" class="wm-item">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 11.5a8.4 8.4 0 0 1-8.5 8.3 8.8 8.8 0 0 1-3.8-.8L3 21l2-5.2a8 8 0 0 1-1-3.9A8.4 8.4 0 0 1 12.5 3.5 8.4 8.4 0 0 1 21 11.5Z" />
            </svg>
            <span>{{ t('write.menuNote') }}</span>
          </router-link>
        </div>
      </transition>
    </div>

    <!-- 管理员快捷入口（登录态常驻 30 天） -->
    <router-link v-if="auth.loggedIn" to="/admin" class="admin-dot" :title="t('admin.loginTitle')">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="8" r="4" />
        <path d="M4 21c0-4 3.6-6.5 8-6.5s8 2.5 8 6.5" />
      </svg>
    </router-link>
  </header>

  <!-- 移动端：悬浮顶栏 + 汉堡侧栏 -->
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
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          class="drawer-link"
          active-class="active"
          @click="drawerOpen = false"
        >
          {{ t(link.key) }}
        </router-link>
      </nav>

      <!-- 管理员：写作与控制中心入口 -->
      <nav v-if="auth.loggedIn" class="drawer-admin">
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
/* ===== PC 胶囊 ===== */
.nav-pc {
  position: fixed;
  top: 18px;
  left: 0;
  right: 0;
  z-index: 100;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 14px;
  pointer-events: none;

  > * { pointer-events: auto; }
}

.capsule {
  display: flex;
  gap: 2px;
  padding: 5px;
  border-radius: 999px;
  background: var(--glass);
  backdrop-filter: blur(14px) saturate(1.5);
  -webkit-backdrop-filter: blur(14px) saturate(1.5);
  border: 1px solid rgba(var(--primary-rgb), 0.18);
  box-shadow: var(--shadow), inset 0 1px 0 rgba(255, 255, 255, 0.12);
}

.capsule-link {
  font-size: 14px;
  font-weight: 600;
  padding: 8px 18px;
  border-radius: 999px;
  color: var(--text-2);
  transition: all var(--dur-fast) var(--ease-out);

  &:hover { background: var(--surface-2); color: var(--text); }

  &.active {
    background: linear-gradient(180deg, var(--primary), var(--primary-deep));
    color: #fff;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.35);
    box-shadow: 0 2px 10px rgba(var(--primary-rgb), 0.5);
  }
}

/* ===== 移动悬浮顶栏 ===== */
.nav-mobile {
  display: none;
  position: fixed;
  top: 12px;
  left: 12px;
  right: 12px;
  z-index: 100;
  padding: 10px 16px;
  border-radius: var(--radius);
  background: var(--glass);
  backdrop-filter: blur(14px) saturate(1.5);
  -webkit-backdrop-filter: blur(14px) saturate(1.5);
  border: 1px solid rgba(var(--primary-rgb), 0.18);
  box-shadow: var(--shadow);
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
    border-radius: 2px;
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
  border-left: 1px solid var(--border);
  padding: 84px 22px 22px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.drawer-links {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.drawer-link {
  font-size: 16px;
  font-weight: 600;
  padding: 12px 16px;
  border-radius: var(--radius);
  color: var(--text-2);
  transition: all var(--dur-fast);

  &.active {
    background: rgba(var(--primary-rgb), 0.12);
    color: var(--primary);
  }
}

.write-wrap { position: relative; }

.write-menu {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  z-index: 120;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 8px;
  min-width: 150px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  box-shadow: var(--shadow), 0 16px 40px -12px rgba(var(--primary-rgb), 0.25);
}

.wm-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 10px;
  font-size: 13.5px;
  font-weight: 600;
  color: var(--text-2);
  transition: all var(--dur-fast);

  svg { width: 16px; height: 16px; }

  &:hover {
    background: rgba(var(--primary-rgb), 0.1);
    color: var(--primary);
    transform: translateX(2px);
  }
}

.pop-enter-active, .pop-leave-active { transition: opacity var(--dur-fast), transform var(--dur-fast) var(--ease-out); }
.pop-enter-from, .pop-leave-to { opacity: 0; transform: translateY(-8px) scale(0.96); }

.admin-dot {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: var(--text);
  background: var(--glass);
  backdrop-filter: blur(14px) saturate(1.5);
  -webkit-backdrop-filter: blur(14px) saturate(1.5);
  border: 1px solid rgba(var(--primary-rgb), 0.18);
  box-shadow: var(--shadow), inset 0 1px 0 rgba(255, 255, 255, 0.12);
  transition: all var(--dur-fast) var(--ease-out);

  svg { width: 18px; height: 18px; }

  &:hover {
    color: var(--primary);
    border-color: rgba(var(--primary-rgb), 0.5);
    transform: scale(1.08);
    box-shadow: var(--shadow), 0 0 14px rgba(var(--primary-rgb), 0.35);
  }
}

.drawer-admin {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding-top: 14px;
  border-top: 1px solid var(--border);
}

/* 主题条：居中 + 放大触控目标 */
.drawer-theme {
  display: flex;
  justify-content: center;
  margin-top: auto;
  padding-bottom: 8px;

  :deep(.switcher) {
    padding: 12px 18px;
    gap: 12px;
  }

  :deep(.dot) {
    width: 22px;
    height: 22px;
  }

  :deep(.divider) { height: 20px; }

  :deep(.mode-btn) {
    width: 32px;
    height: 32px;
  }

  :deep(.mode-icon) {
    width: 24px;
    height: 24px;
  }
}

.drawer-mask {
  position: fixed;
  inset: 0;
  z-index: 98;
  background: rgba(0, 0, 0, 0.4);
}

/* 过渡 */
.drawer-enter-active, .drawer-leave-active { transition: transform var(--dur) var(--ease-out); }
.drawer-enter-from, .drawer-leave-to { transform: translateX(100%); }
.fade-enter-active, .fade-leave-active { transition: opacity var(--dur-fast); }
.fade-enter-from, .fade-leave-to { opacity: 0; }

@media (max-width: 768px) {
  .nav-pc { display: none; }
  .nav-mobile { display: flex; }
}
</style>
