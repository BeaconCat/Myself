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
  padding: 8px 14px calc(var(--safe-b) + 24px);
  overflow-x: hidden;

  /* 桌面组件在窄屏下的兜底：去掉自带的大外边距与最小高度 */
  :deep(> *) {
    min-height: 0 !important;
  }
}
</style>
