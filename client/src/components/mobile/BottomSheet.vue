<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, watch } from 'vue';
import MIcon from './MIcon.vue';
import { attachDrag, clamp, rubber } from './gesture';

/**
 * 底部 sheet：可拖拽 detent（半屏 / 全屏）+ 越界橡皮筋 + 按投影位置吸附，快速下甩关闭。
 * detents 省略时按内容自适应一档（最高 86% 视口）。
 */
const props = withDefaults(
  defineProps<{ open: boolean; title?: string; meta?: string; detents?: number[] }>(),
  { title: '', meta: '', detents: undefined },
);
const emit = defineEmits<{ 'update:open': [v: boolean] }>();

const sheet = ref<HTMLElement | null>(null);
const content = ref<HTMLElement | null>(null);
const visible = ref(false);
const h = ref(0);
const anim = ref(false);
const bounce = ref(false);
const stops = ref<number[]>([]);

function measure(): number[] {
  const vh = window.innerHeight;
  if (props.detents?.length) return props.detents.map((d) => Math.min(d, vh * 0.92));
  const natural = (content.value?.scrollHeight ?? 300) + 92;
  return [Math.min(natural, vh * 0.86)];
}

const max = () => stops.value[stops.value.length - 1] ?? 0;

function setH(v: number, animate: boolean): void {
  const top = max();
  h.value = v > top ? top + rubber(v - top, 50) : Math.max(0, v);
  anim.value = animate;
  bounce.value = animate;
}

watch(
  () => props.open,
  async (on) => {
    if (on) {
      visible.value = true;
      h.value = 0;
      anim.value = false;
      await nextTick();
      stops.value = measure();
      requestAnimationFrame(() => requestAnimationFrame(() => setH(stops.value[0], true)));
    } else if (visible.value) {
      anim.value = true;
      bounce.value = false;
      h.value = 0;
      window.setTimeout(() => { if (!props.open) visible.value = false; }, 520);
    }
  },
);

function close(): void {
  emit('update:open', false);
}

const handles = ref<HTMLElement[]>([]);
const detach: Array<() => void> = [];

watch(visible, async (on) => {
  if (!on) return;
  await nextTick();
  detach.splice(0).forEach((f) => f());
  const els = sheet.value?.querySelectorAll<HTMLElement>('[data-grab]') ?? [];
  els.forEach((el) =>
    detach.push(
      attachDrag(el, {
        axis: 'y',
        down: () => ({ h0: h.value }),
        begin: () => { anim.value = false; },
        move: (c, _dx, dy) => setH(c.h0 - dy, false),
        end: (_c, _vx, vy) => {
          const proj = h.value - vy * 180;
          const ds = stops.value;
          if (proj < ds[0] * 0.55 || vy > 1.4) {
            close();
            return;
          }
          const target = ds.reduce((a, b) => (Math.abs(b - proj) < Math.abs(a - proj) ? b : a));
          setH(clamp(target, 0, max()), true);
        },
      }),
    ),
  );
});

onBeforeUnmount(() => detach.forEach((f) => f()));
defineExpose({ handles });
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="m-layer sheet-layer">
      <div
        class="scrim"
        :class="{ anim }"
        :style="{ opacity: stops[0] ? clamp(h / stops[0], 0, 1) : 0 }"
        @click="close"
      />
      <section
        ref="sheet"
        class="sheet"
        :class="{ anim, bounce }"
        :style="{ height: `${max() + 100}px`, transform: `translateY(${max() + 100 - h - 100}px)` }"
        role="dialog"
        aria-modal="true"
      >
        <div class="grab" data-grab><i /></div>
        <div class="head" data-grab>
          <b>{{ title }}</b><small v-if="meta">{{ meta }}</small>
          <button class="close m-tap" aria-label="close" @click="close"><MIcon name="close" /></button>
        </div>
        <div ref="content" class="content" :style="{ maxHeight: `${Math.max(0, max() - 92)}px` }">
          <slot :close="close" />
        </div>
      </section>
    </div>
  </Teleport>
</template>

<style scoped lang="scss">
.sheet-layer {
  position: fixed;
  inset: 0;
  z-index: 110;
}

.scrim {
  position: absolute;
  inset: 0;
  background: var(--m-scrim);

  &.anim { transition: opacity 0.5s var(--m-ease-sheet); }
}

.sheet {
  position: absolute;
  left: 0;
  right: 0;
  bottom: -100px;
  display: flex;
  flex-direction: column;
  border-radius: var(--r-xl) var(--r-xl) 0 0;
  background: var(--elev);
  box-shadow: 0 -10px 40px -10px rgba(0, 0, 0, 0.4), inset 0 0.5px 0 var(--m-glass-line);
  will-change: transform;
  padding-bottom: 100px;

  &.anim { transition: transform 0.55s var(--m-ease-sheet); }
  &.anim.bounce { transition: transform 0.6s cubic-bezier(0.2, 0.9, 0.3, 1.12); }
}

.grab {
  height: 22px;
  display: grid;
  place-items: center;
  flex: none;
  touch-action: none;

  i {
    width: 38px;
    height: 5px;
    border-radius: 999px;
    background: var(--line-2);
  }
}

.head {
  display: flex;
  align-items: center;
  padding: 2px 20px 12px;
  flex: none;
  touch-action: none;

  b {
    font-size: 17px;
    font-weight: 600;
  }

  small {
    margin-left: 10px;
    font-size: 13px;
    color: var(--text-3);
  }

  .close {
    margin-left: auto;
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background: var(--fill-2);

    .m-ic {
      width: 16px;
      height: 16px;
      stroke-width: 2;
    }
  }
}

.content {
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-width: none;
  padding-bottom: calc(var(--m-safe-b) + 12px);

  &::-webkit-scrollbar { display: none; }
}
</style>
