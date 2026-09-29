<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import MIcon from './MIcon.vue';
import { shell, TAB_ORDER, type TabName } from './shell';

/**
 * V1 悬浮玻璃胶囊底栏：四个 tab + 独立搜索圆钮。
 * 选中胶囊 = 抬升 + 轻染（--lift / --lift-shadow / --lift-fg），不挂信号点；
 * morph：前缘先走（快）、后缘跟上（慢 + 延迟），形成拉伸的液滴感；
 * 页面下滑时收缩为纯图标，上滑恢复；详情推入时整体下沉隐藏。
 */
const props = defineProps<{ active: TabName }>();
const emit = defineEmits<{ select: [tab: TabName]; search: [] }>();
const { t } = useI18n();

const ICONS: Record<TabName, string> = { home: 'home', articles: 'book', thoughts: 'bubble', about: 'user' };

const index = computed(() => Math.max(0, TAB_ORDER.indexOf(props.active)));
const toLeft = ref(false);
const popped = ref<TabName | ''>('');

watch(index, (now, before) => {
  toLeft.value = now < before;
  popped.value = props.active;
  window.setTimeout(() => { if (popped.value === props.active) popped.value = ''; }, 520);
});
</script>

<template>
  <div class="tb1" :class="{ min: shell.barMin }">
    <nav class="cap" :class="{ 'to-l': toLeft }" :style="{ '--i': index, '--n': TAB_ORDER.length }">
      <span class="pill" aria-hidden="true" />
      <button
        v-for="name in TAB_ORDER"
        :key="name"
        class="it"
        :class="{ on: name === active, pop: popped === name }"
        :aria-current="name === active ? 'page' : undefined"
        @click="emit('select', name)"
      >
        <MIcon :name="ICONS[name]" />
        <span>{{ t(`mobile.tabs.${name}`) }}</span>
      </button>
    </nav>
    <button class="search m-tap" :aria-label="t('mobile.tabs.search')" @click="emit('search')">
      <MIcon name="search" />
    </button>
  </div>
</template>

<style scoped lang="scss">
.tb1 {
  position: absolute;
  left: 0;
  right: 0;
  bottom: max(14px, calc(var(--m-safe-b) - 6px));
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  pointer-events: none;

  > * { pointer-events: auto; }
}

.cap {
  --pad: 5px;
  position: relative;
  display: grid;
  grid-template-columns: repeat(var(--n), 1fr);
  width: min(298px, calc(100vw - 104px));
  height: 64px;
  padding: var(--pad);
  border-radius: 999px;
  background: var(--m-glass);
  backdrop-filter: blur(24px) saturate(190%);
  -webkit-backdrop-filter: blur(24px) saturate(190%);
  box-shadow: inset 0 0 0 0.5px var(--m-glass-line), inset 0 1px 0 var(--m-glass-hi), var(--m-bar-shadow);
  transition: width 0.5s var(--ease-spring), height 0.5s var(--ease-spring);
}

.pill {
  position: absolute;
  top: var(--pad);
  bottom: var(--pad);
  border-radius: 999px;
  left: calc(var(--pad) + (100% - 2 * var(--pad)) * var(--i) / var(--n));
  right: calc(var(--pad) + (100% - 2 * var(--pad)) * (var(--n) - 1 - var(--i)) / var(--n));
  background: var(--lift);
  box-shadow: var(--lift-shadow);
  transition: left 0.46s var(--ease-spring) 0.05s, right 0.3s var(--ease-out), background-color var(--dur), box-shadow var(--dur);
}

.cap.to-l .pill {
  transition: left 0.3s var(--ease-out), right 0.46s var(--ease-spring) 0.05s, background-color var(--dur), box-shadow var(--dur);
}

.it {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  border-radius: 999px;
  color: var(--text-2);
  /* 镂空细节描边色 ≈ 抬升块的不透明近似 */
  --cut: var(--lift);
  transition: color var(--dur) var(--ease-out), transform var(--dur-fast) var(--ease-spring);

  &:active { transform: scale(0.92); }

  span {
    font-size: 10.5px;
    font-weight: 500;
    line-height: 1;
    max-height: 12px;
    overflow: hidden;
    transition: max-height 0.4s var(--ease-out), opacity 0.3s;
  }

  &.on {
    color: var(--lift-fg);

    :deep(.m-ic) { stroke-width: 2.1; }
  }

  &.pop :deep(.m-ic) { animation: pop 0.5s var(--ease-spring); }
}

@keyframes pop {
  0% { transform: scale(1); }
  25% { transform: scale(0.78) translateY(1px); }
  60% { transform: scale(1.14) translateY(-2px); }
  100% { transform: none; }
}

html.m-shell[data-mode='dark'] .it {
  --cut: color-mix(in oklab, var(--primary) 16%, color-mix(in oklab, var(--surface) 88%, white));
}

.search {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: var(--text);
  background: var(--m-glass);
  backdrop-filter: blur(24px) saturate(190%);
  -webkit-backdrop-filter: blur(24px) saturate(190%);
  box-shadow: inset 0 0 0 0.5px var(--m-glass-line), inset 0 1px 0 var(--m-glass-hi), var(--m-bar-shadow);
  transition: width 0.5s var(--ease-spring), height 0.5s var(--ease-spring), transform var(--dur-fast) var(--ease-spring);
}

/* 下滑收缩：只剩图标 */
.tb1.min {
  .cap { width: min(216px, calc(100vw - 100px)); height: 50px; }
  .it span { max-height: 0; opacity: 0; }
  .search { width: 50px; height: 50px; }
}
</style>
