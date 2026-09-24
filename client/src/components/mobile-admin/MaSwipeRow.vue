<script setup lang="ts">
/**
 * 左滑操作行：跟手位移，右侧按钮随拉出比例放大显形；速度/位置决定展开或收回；
 * 越界橡皮筋；同一时刻只展开一行（点别处自动收起）。
 */
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { bindDrag, clamp, haptic, rubber } from './gesture';
import MaIcon from './MaIcon.vue';
import type { IconName } from './icons';

export interface SwipeAction {
  id: string;
  label: string;
  icon: IconName;
  color: string;
}

const props = defineProps<{ actions: SwipeAction[] }>();
const emit = defineEmits<{ action: [id: string]; tap: [] }>();

const ACTION_W = 74;
const total = computed(() => props.actions.length * ACTION_W);
const x = ref(0);
const dragging = ref(false);
const el = ref<HTMLElement | null>(null);
const progress = computed(() => clamp(-x.value / total.value, 0, 1.2));

const self = (): void => close();
function close(): void {
  x.value = 0;
  if (openRow === self) openRow = null;
}

let unbind: (() => void) | null = null;
onMounted(() => {
  if (!el.value) return;
  unbind = bindDrag(el.value, {
    axis: 'x',
    down: () => ({ x0: x.value }),
    begin: () => {
      if (openRow && openRow !== self) openRow();
      dragging.value = true;
    },
    move: (c, dx) => {
      let v = c.x0 + dx;
      if (v > 0) v = rubber(v, 40);
      else if (v < -total.value) v = -total.value - rubber(-total.value - v, 50);
      x.value = v;
    },
    end: (_c, vx) => {
      dragging.value = false;
      const open = vx < -0.3 || (vx <= 0.3 && x.value < -total.value / 2);
      if (open) {
        if (x.value > -total.value + 4) haptic(8);
        x.value = -total.value;
        openRow = self;
      } else close();
    },
  });
});
onBeforeUnmount(() => {
  unbind?.();
  if (openRow === self) openRow = null;
});

function onTap(): void {
  if (x.value !== 0) {
    close();
    return;
  }
  if (openRow) {
    openRow();
    return;
  }
  emit('tap');
}

function act(id: string): void {
  close();
  emit('action', id);
}
</script>

<script lang="ts">
/** 当前展开的行（全局唯一） */
let openRow: (() => void) | null = null;
export function closeOpenSwipeRow(): void {
  openRow?.();
}
</script>

<template>
  <div class="ma-swipe" :class="{ open: x < 0 }">
    <div class="acts" :style="{ width: `${Math.max(total, -x)}px` }">
      <button
        v-for="(a, k) in actions"
        :key="a.id"
        class="act"
        :style="{
          '--c': a.color,
          transform: `scale(${0.6 + 0.4 * clamp(progress * actions.length - (actions.length - 1 - k) * 0.5, 0, 1)})`,
          opacity: clamp(progress * 1.6, 0, 1),
        }"
        :tabindex="x < 0 ? 0 : -1"
        @click="act(a.id)"
      >
        <span class="bub"><MaIcon :name="a.icon" :size="20" /></span>
        <small>{{ a.label }}</small>
      </button>
    </div>
    <div
      ref="el"
      class="front"
      :class="{ anim: !dragging }"
      :style="{ transform: `translateX(${x}px)` }"
      @click="onTap"
    >
      <slot />
    </div>
  </div>
</template>

<style scoped lang="scss">
.ma-swipe {
  position: relative;
  overflow: hidden;
}

.front {
  position: relative;
  z-index: 1;
  background: var(--bg);
  touch-action: pan-y;

  &.anim {
    transition: transform 0.45s var(--ease-spring);
  }
}

.acts {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  display: flex;
  justify-content: flex-end;
  align-items: center;
  padding-right: 8px;
}

.act {
  width: 66px;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 5px;
  color: var(--text-2);
  transform-origin: center;

  .bub {
    width: 42px;
    height: 42px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    color: #fff;
    background: var(--c);
    box-shadow: 0 8px 16px -8px var(--c);
    transition: transform var(--dur-fast) var(--ease-spring);
  }

  &:active .bub {
    transform: scale(0.9);
  }

  small {
    font-size: 11.5px;
  }
}
</style>
