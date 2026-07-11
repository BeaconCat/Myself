<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useAuthStore } from '../../stores/auth';
import ThemeSwitcher from '../../components/layout/ThemeSwitcher.vue';

const { t } = useI18n();
const router = useRouter();
const auth = useAuthStore();
const drawerOpen = ref(false);

const menu = [
  { to: '/admin/posts', key: 'admin.menuPosts' },
  { to: '/admin/notes', key: 'admin.menuNotes' },
  { to: '/admin/media', key: 'admin.menuMedia' },
  { to: '/admin/about', key: 'admin.menuAbout' },
  { to: '/admin/apikeys', key: 'admin.menuApi' },
  { to: '/admin/data', key: 'admin.menuData' },
  { to: '/admin/quality', key: 'admin.menuQuality' },
  { to: '/admin/settings', key: 'admin.menuSettings' },
];

function logout(): void {
  auth.logout();
  void router.push('/admin/login');
}
</script>

<template>
  <div class="admin">
    <!-- 移动端顶栏 -->
    <header class="m-top">
      <router-link to="/" class="brand">
        <img src="/favicon-64.png" alt="" draggable="false" />
        <span>Myself</span>
      </router-link>
      <button
        class="hamburger"
        :class="{ open: drawerOpen }"
        aria-label="菜单"
        @click="drawerOpen = !drawerOpen"
      >
        <span /><span /><span />
      </button>
    </header>

    <!-- 移动端抽屉 -->
    <transition name="drawer">
      <aside v-if="drawerOpen" class="m-drawer">
        <nav class="menu">
          <router-link
            v-for="m in menu"
            :key="m.to"
            :to="m.to"
            class="menu-item"
            active-class="on"
            @click="drawerOpen = false"
          >{{ t(m.key) }}</router-link>
        </nav>
        <div class="side-foot">
          <ThemeSwitcher />
          <button class="logout" @click="logout">{{ t('admin.logout') }}</button>
        </div>
      </aside>
    </transition>
    <transition name="fade">
      <div v-if="drawerOpen" class="m-mask" @click="drawerOpen = false" />
    </transition>

    <aside class="side">
      <router-link to="/" class="brand">
        <img src="/favicon-64.png" alt="" draggable="false" />
        <span>Myself</span>
      </router-link>

      <nav class="menu">
        <router-link
          v-for="m in menu"
          :key="m.to"
          :to="m.to"
          class="menu-item"
          active-class="on"
        >{{ t(m.key) }}</router-link>
      </nav>

      <div class="side-foot">
        <ThemeSwitcher />
        <button class="logout" @click="logout">{{ t('admin.logout') }}</button>
      </div>
    </aside>

    <main class="content">
      <router-view v-slot="{ Component }">
        <transition name="page" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
  </div>
</template>

<style scoped lang="scss">
.admin {
  display: grid;
  grid-template-columns: 220px 1fr;
  min-height: 100vh;
}

.side {
  position: sticky;
  top: 0;
  height: 100vh;
  display: flex;
  flex-direction: column;
  padding: 22px 16px;
  background:
    radial-gradient(400px 200px at 0% 0%, rgba(var(--primary-rgb), 0.07), transparent 70%),
    var(--surface);
  border-right: 1px solid var(--border);
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 10px;
  margin-bottom: 26px;

  img { width: 30px; height: 30px; }

  span {
    font-family: var(--font-serif);
    font-weight: 700;
    font-size: 18px;
  }
}

.menu {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}

.menu-item {
  position: relative;
  padding: 11px 14px 11px 18px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-2);
  transition: all var(--dur-fast);

  /* 左缘主色指示条 */
  &::before {
    content: '';
    position: absolute;
    left: 6px;
    top: 50%;
    transform: translateY(-50%) scaleY(0);
    width: 3px;
    height: 18px;
    border-radius: 3px;
    background: linear-gradient(180deg, var(--primary), var(--primary-deep));
    transition: transform var(--dur-fast) var(--ease-out);
  }

  &:hover { background: var(--surface-2); color: var(--text); transform: translateX(2px); }

  &.on {
    background: rgba(var(--primary-rgb), 0.1);
    color: var(--primary);

    &::before { transform: translateY(-50%) scaleY(1); }
  }
}

.side-foot {
  display: flex;
  flex-direction: column;
  gap: 12px;
  align-items: stretch;

  :deep(.switcher) { justify-content: center; }
}

.logout {
  padding: 10px;
  border-radius: 10px;
  border: 1px solid var(--border);
  background: none;
  color: var(--text-2);
  font-size: 13px;
  font-weight: 600;
  transition: all var(--dur-fast);

  &:hover { border-color: var(--accent-red); color: var(--accent-red); }
}

.content {
  padding: 34px 38px;
  min-width: 0;
}

/* 移动端顶栏 + 抽屉（默认隐藏） */
.m-top {
  display: none;
  position: sticky;
  top: 0;
  z-index: 90;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: var(--glass);
  backdrop-filter: blur(14px) saturate(1.4);
  -webkit-backdrop-filter: blur(14px) saturate(1.4);
  border-bottom: 1px solid var(--border);

  .brand { margin-bottom: 0; padding: 0; }
}

.hamburger {
  width: 36px;
  height: 36px;
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

.m-drawer {
  display: none;
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: min(78vw, 300px);
  z-index: 99;
  background: var(--surface);
  border-left: 1px solid var(--border);
  padding: 74px 18px 20px;
  flex-direction: column;
  gap: 20px;
}

.m-mask {
  display: none;
  position: fixed;
  inset: 0;
  z-index: 95;
  background: rgba(0, 0, 0, 0.4);
}

.drawer-enter-active, .drawer-leave-active { transition: transform var(--dur) var(--ease-out); }
.drawer-enter-from, .drawer-leave-to { transform: translateX(100%); }
.fade-enter-active, .fade-leave-active { transition: opacity var(--dur-fast); }
.fade-enter-from, .fade-leave-to { opacity: 0; }

@media (max-width: 860px) {
  .admin { grid-template-columns: 1fr; }

  .side { display: none; }
  .m-top { display: flex; }
  .m-drawer { display: flex; }
  .m-mask { display: block; }

  .m-drawer .menu { flex: 1; }
  .m-drawer .side-foot { align-items: stretch; }

  .content { padding: 18px 14px 40px; }
}
</style>
