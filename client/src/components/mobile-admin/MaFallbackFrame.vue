<script setup lang="ts">
/** 未做移动版的后台页：套一层原生式导航（返回 + 标题），内容区独立滚动承载桌面组件。 */
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import MaIcon from './MaIcon.vue';

defineProps<{ title: string }>();
const emit = defineEmits<{ back: [] }>();
const { t } = useI18n();
const y = ref(0);
</script>

<template>
  <div class="ma-fb" :class="{ scrolled: y > 4 }">
    <header class="fb-nav">
      <button class="fb-back tap" :aria-label="t('mobileAdmin.common.back')" @click="emit('back')">
        <MaIcon name="back" :size="22" />
        <span>{{ t('mobileAdmin.tab.me') }}</span>
      </button>
      <b>{{ title }}</b>
    </header>
    <div class="fb-scroll" @scroll.passive="y = ($event.target as HTMLElement).scrollTop">
      <div class="fb-inner"><slot /></div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.ma-fb {
  position: absolute;
  inset: 0;
  background: var(--bg);
  overflow: hidden;
}

.fb-nav {
  position: absolute;
  z-index: 10;
  top: 0;
  left: 0;
  right: 0;
  height: calc(var(--safe-t) + var(--nav-row));
  padding-top: var(--safe-t);
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--glass-2);
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  box-shadow: 0 0.5px 0 transparent;
  transition: box-shadow var(--dur);

  .scrolled & { box-shadow: 0 0.5px 0 var(--line-2); }

  b {
    font-size: 16.5px;
    font-weight: 600;
  }
}

.fb-back {
  position: absolute;
  left: 8px;
  bottom: 4px;
  height: 36px;
  display: flex;
  align-items: center;
  gap: 0;
  padding: 0 8px 0 2px;
  font-size: 16px;
  color: var(--ink);
}

.fb-scroll {
  position: absolute;
  inset: 0;
  overflow-y: auto;
  overflow-x: hidden;
  overscroll-behavior: contain;
  padding-top: calc(var(--safe-t) + var(--nav-row));
}

.fb-inner {
  min-height: 100%;
  padding: 16px 16px calc(var(--safe-b) + 24px);

  /* 桌面组件在窄屏下的兜底：去掉自带的大外边距与最小高度 */
  :deep(> *) {
    min-height: 0 !important;
  }

  :deep(> .view) { width: 100%; min-width: 0; max-width: none; padding: 0; }
  :deep(.st-vh) { flex-wrap: wrap; align-items: flex-start; gap: 12px; }
  :deep(.st-vh > div:first-child) { display: block; flex: 1 1 180px; }
  :deep(.st-vh h1) { font-size: 24px; }
  :deep(.st-vh p) { margin-top: 6px; white-space: normal; overflow-wrap: anywhere; line-height: 1.6; }
  :deep(.st-vh .act) { max-width: 100%; flex-wrap: wrap; }
  :deep(.st-card) { min-width: 0; padding: 16px; }
  :deep(.st-sec-t) { flex-wrap: wrap; align-items: center; gap: 10px; }
  :deep(.st-sec-t h2) { flex: none; font-size: 20px; }
  :deep(.st-field) { min-width: 0; max-width: 100%; }
  :deep(.st-field input), :deep(.st-field select), :deep(.st-field textarea) { min-width: 0; }
  :deep(.seg) { min-width: 0; max-width: 100%; overflow-x: auto; scrollbar-width: thin; }
  :deep(.seg button) { flex-shrink: 0; }
  :deep(.st-stat) { padding: 14px 12px; }
  :deep(.st-stat b) { font-size: 26px; }
  :deep(.st-stat small) { white-space: normal; overflow: visible; line-height: 1.5; }
  :deep(.st-note-bar) { flex-wrap: wrap; }
}
</style>
