<script setup lang="ts">
/**
 * V2 贴底底栏：概览 / 内容 / 凸起 + / 素材 / 我的。
 * 选中：图标描边加粗，图标着 --ink、文字 --text，不发光，并做一次弹跳；
 * 中央 + 为实底 --solid / --on-solid + 中性紧阴影，打开 action sheet 时旋转成 ×。
 */
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import MaIcon from './MaIcon.vue';
import type { IconName } from './icons';

export type AdminTab = 'today' | 'content' | 'media' | 'me';

const props = defineProps<{ active: AdminTab | null; fabOpen: boolean; hidden?: boolean }>();
const emit = defineEmits<{ select: [tab: AdminTab]; fab: [] }>();
const { t } = useI18n();

const left: { id: AdminTab; icon: IconName }[] = [
  { id: 'today', icon: 'grid' },
  { id: 'content', icon: 'docs' },
];
const right: { id: AdminTab; icon: IconName }[] = [
  { id: 'media', icon: 'image' },
  { id: 'me', icon: 'user' },
];

/** 弹跳只在切换时播一次 */
const popping = ref<AdminTab | null>(null);
watch(
  () => props.active,
  (v) => {
    popping.value = null;
    requestAnimationFrame(() => {
      popping.value = v;
    });
  },
);
</script>

<template>
  <nav class="ma-tabbar" :class="{ hidden }" :aria-hidden="hidden">
    <button
      v-for="it in left"
      :key="it.id"
      class="tb-it"
      :class="{ on: active === it.id, pop: popping === it.id }"
      :aria-current="active === it.id ? 'page' : undefined"
      @click="emit('select', it.id)"
    >
      <MaIcon :name="it.icon" />
      <span>{{ t(`mobileAdmin.tab.${it.id}`) }}</span>
    </button>
    <div class="fab-slot">
      <button
        class="fab tap"
        :class="{ open: fabOpen }"
        :aria-label="t('mobileAdmin.tab.create')"
        :aria-expanded="fabOpen"
        @click="emit('fab')"
      >
        <MaIcon name="plus" :size="26" />
      </button>
    </div>
    <button
      v-for="it in right"
      :key="it.id"
      class="tb-it"
      :class="{ on: active === it.id, pop: popping === it.id }"
      :aria-current="active === it.id ? 'page' : undefined"
      @click="emit('select', it.id)"
    >
      <MaIcon :name="it.icon" />
      <span>{{ t(`mobileAdmin.tab.${it.id}`) }}</span>
    </button>
  </nav>
</template>

<style scoped lang="scss">
.ma-tabbar {
  position: absolute;
  z-index: 46;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  height: calc(var(--tab-h) + var(--safe-b));
  padding: 0 6px var(--safe-b);
  background: var(--glass-2);
  backdrop-filter: blur(26px) saturate(190%);
  -webkit-backdrop-filter: blur(26px) saturate(190%);
  box-shadow: 0 -0.5px 0 var(--line-2);
  transition: transform 0.45s var(--ease-sheet), opacity 0.3s;

  &.hidden {
    transform: translateY(calc(100% + 40px));
    opacity: 0;
    pointer-events: none;
  }
}

.tb-it {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  padding-top: 4px;
  color: var(--text-3);
  transition: color var(--dur) var(--ease-out);

  span {
    font-size: 10.5px;
    font-weight: 500;
  }

  :deep(.ma-ic) { transition: color var(--dur) var(--ease-out); }

  &.on {
    color: var(--text);

    :deep(.ma-ic) { color: var(--ink); }
    :deep(.ma-ic) { stroke-width: 2.1; }
  }

  &.pop :deep(.ma-ic) {
    animation: ma-pop 0.55s var(--ease-spring);
  }

  &:active :deep(.ma-ic) {
    transform: scale(0.88);
    transition: transform 0.08s;
  }
}

@keyframes ma-pop {
  0% { transform: none; }
  30% { transform: scale(0.8); }
  65% { transform: scale(1.16) translateY(-3px); }
  100% { transform: none; }
}

.fab-slot {
  flex: 1;
  display: grid;
  place-items: center;
}

.fab {
  width: 58px;
  height: 58px;
  margin-top: -26px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: var(--on-solid);
  background: var(--solid);
  /* 底色描边把按钮从底栏里「挖」出来 + 中性紧阴影（不发光） */
  box-shadow:
    0 0 0 5px var(--bg),
    var(--btn-shadow),
    0 6px 14px -8px rgb(0 0 0 / 0.35);
  transition: background-color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  &:active { background: var(--solid-hover); }

  :deep(.ma-ic) {
    stroke-width: 2;
    transition: transform 0.45s var(--ease-spring);
  }

  &.open :deep(.ma-ic) {
    transform: rotate(135deg);
  }
}
</style>
