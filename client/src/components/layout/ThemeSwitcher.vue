<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useThemeStore } from '../../stores/theme';
import { useAuthStore } from '../../stores/auth';
import { adminApi } from '../../api';
import { circularReveal } from '../../utils/circularReveal';

const { t } = useI18n();
const theme = useThemeStore();
const auth = useAuthStore();

/** 关闭访客换肤只影响访客：管理员永远保留色盘条 */
const showPalettes = computed(() => theme.allowUserPalette || auth.loggedIn);

function origin(e: MouseEvent) {
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
  return { x: rect.left + rect.width / 2, y: rect.top + rect.height / 2 };
}

/** 管理员切换 = 顺手写入全局默认主题 */
function persistAsDefault(patch: Record<string, string>): void {
  if (!auth.loggedIn) return;
  adminApi.saveSettings({ theme: patch }).catch(() => undefined);
}

function onPalette(e: MouseEvent, id: string): void {
  if (id === theme.paletteId) return;
  circularReveal(origin(e), () => theme.setPalette(id, auth.loggedIn), 'expand');
  persistAsDefault({ defaultPaletteId: id });
}

function onMode(e: MouseEvent): void {
  const toLight = theme.mode === 'dark';
  circularReveal(origin(e), () => theme.toggleMode(), toLight ? 'expand' : 'contract');
  persistAsDefault({ defaultMode: toLight ? 'light' : 'dark' });
}
</script>

<template>
  <!-- 仅深浅切换时收成正圆按钮 -->
  <div class="switcher" :class="{ solo: !showPalettes }">
    <template v-if="showPalettes">
      <button
        v-for="p in theme.allPalettes"
        :key="p.id"
        class="dot"
        :class="{ active: theme.paletteId === p.id }"
        :style="{ background: p[theme.mode].primary }"
        :title="p.nameKey"
        @click="onPalette($event, p.id)"
      />
      <span class="divider" />
    </template>
    <button
      class="mode-btn"
      :title="theme.mode === 'light' ? t('theme.dark') : t('theme.light')"
      @click="onMode($event)"
    >
      <svg v-if="theme.mode === 'light'" class="mode-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
        <circle cx="12" cy="12" r="4.2" fill="currentColor" stroke="none" />
        <path d="M12 2.5v2.6M12 18.9v2.6M2.5 12h2.6M18.9 12h2.6M5.3 5.3l1.85 1.85M16.85 16.85l1.85 1.85M18.7 5.3l-1.85 1.85M7.15 16.85L5.3 18.7" />
      </svg>
      <svg v-else class="mode-icon" viewBox="0 0 24 24" fill="currentColor">
        <path d="M20.6 14.3A8.7 8.7 0 0 1 9.7 3.4a0.5 0.5 0 0 0-.65-.62A9.8 9.8 0 1 0 21.2 14.95a0.5 0.5 0 0 0-.6-.65Z" />
      </svg>
    </button>
  </div>
</template>

<style scoped lang="scss">
.switcher {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 999px;
  background: var(--glass);
  backdrop-filter: blur(14px) saturate(1.5);
  -webkit-backdrop-filter: blur(14px) saturate(1.5);
  border: 1px solid rgba(var(--primary-rgb), 0.18);
  box-shadow: var(--shadow), inset 0 1px 0 rgba(255, 255, 255, 0.12);

  /* 正圆模式：只剩深浅切换 */
  &.solo {
    width: 40px;
    height: 40px;
    padding: 0;
    justify-content: center;
  }
}

.dot {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 2px solid transparent;
  transition: transform var(--dur-fast) var(--ease-spring), border-color var(--dur-fast), box-shadow var(--dur-fast);

  &:hover { transform: scale(1.2); }

  &.active {
    border-color: var(--surface);
    transform: scale(1.25);
    box-shadow: 0 0 0 1.5px rgba(var(--primary-rgb), 0.7), 0 0 10px rgba(var(--primary-rgb), 0.5);
  }
}

.divider {
  width: 1px;
  height: 16px;
  background: var(--border);
}

.mode-btn {
  border: none;
  background: none;
  width: 26px;
  height: 26px;
  display: grid;
  place-items: center;
  color: var(--text);
  transition: transform var(--dur-fast) var(--ease-spring);

  &:hover { transform: rotate(18deg) scale(1.12); }
}

.mode-icon {
  width: 19px;
  height: 19px;
}
</style>
