<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useAuthStore } from '../../stores/auth';
import ThemeSwitcher from '../../components/layout/ThemeSwitcher.vue';

const { t } = useI18n();
const router = useRouter();
const auth = useAuthStore();

const menu = [
  { to: '/admin/posts', key: 'admin.menuPosts' },
  { to: '/admin/notes', key: 'admin.menuNotes' },
  { to: '/admin/apikeys', key: 'admin.menuApi' },
  { to: '/admin/settings', key: 'admin.menuSettings' },
];

function logout(): void {
  auth.logout();
  void router.push('/admin/login');
}
</script>

<template>
  <div class="admin">
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
  background: var(--surface);
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
  padding: 11px 14px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-2);
  transition: all var(--dur-fast);

  &:hover { background: var(--surface-2); color: var(--text); }

  &.on {
    background: rgba(var(--primary-rgb), 0.12);
    color: var(--primary);
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

@media (max-width: 860px) {
  .admin { grid-template-columns: 1fr; }

  .side {
    position: static;
    height: auto;
    flex-direction: row;
    align-items: center;
    gap: 12px;
    border-right: none;
    border-bottom: 1px solid var(--border);

    .brand { margin-bottom: 0; }
    .menu { flex-direction: row; }
    .side-foot { flex-direction: row; margin-left: auto; }
  }

  .content { padding: 20px 16px; }
}
</style>
