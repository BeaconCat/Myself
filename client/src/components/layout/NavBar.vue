<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import ThemeSwitcher from './ThemeSwitcher.vue';

const { t } = useI18n();
const drawerOpen = ref(false);

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
      <ThemeSwitcher />
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
