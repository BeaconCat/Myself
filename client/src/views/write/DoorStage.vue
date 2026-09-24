<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import '../admin/studio/i18n';
import SIcon from '../admin/studio/SIcon.vue';
import LightCover from '../admin/studio/LightCover.vue';

/**
 * 发布完成：「门开了」——门扇打开、光洒出、文章卡片从门里走出来。
 * 时间线 open(350ms) → out(1500ms) → fin(2100ms)；减弱动效时直接到终态。
 */
const props = defineProps<{
  open: boolean;
  title: string;
  cover: string;
  seed: number | string;
  url: string;
  meta: string;
}>();
const emit = defineEmits<{ view: []; close: [] }>();
const { t } = useI18n();

const phase = ref<'' | 'open' | 'out' | 'fin'>('');
const shown = ref(false);
let timers: number[] = [];

const DUST = Array.from({ length: 18 }, (_, i) => ({
  x: `${(Math.sin(i * 7.1) * 110).toFixed(0)}px`,
  dl: `${(i * 0.19).toFixed(2)}s`,
  top: `${(Math.cos(i * 3.3) * 20).toFixed(0)}px`,
}));

function clear(): void {
  timers.forEach((id) => window.clearTimeout(id));
  timers = [];
}

function onKey(e: KeyboardEvent): void {
  if (e.key === 'Escape') emit('close');
}

watch(
  () => props.open,
  (v) => {
    clear();
    if (v) {
      shown.value = true;
      phase.value = '';
      const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
      const T = reduced ? [0, 0, 0] : [350, 1500, 2100];
      timers.push(window.setTimeout(() => (phase.value = 'open'), T[0]));
      timers.push(window.setTimeout(() => (phase.value = 'out'), T[1]));
      timers.push(window.setTimeout(() => (phase.value = 'fin'), T[2]));
      document.addEventListener('keydown', onKey);
    } else {
      document.removeEventListener('keydown', onKey);
      timers.push(window.setTimeout(() => {
        shown.value = false;
        phase.value = '';
      }, 500));
    }
  },
);

onBeforeUnmount(() => {
  clear();
  document.removeEventListener('keydown', onKey);
});

const shortTitle = () => props.title.split(/[：:]/)[0];
</script>

<template>
  <Teleport to="body">
    <div
      v-if="shown"
      class="studio stage"
      :class="{
        on: open,
        open: ['open', 'out', 'fin'].includes(phase),
        out: ['out', 'fin'].includes(phase),
        fin: phase === 'fin',
      }"
    >
      <div class="bloom" />
      <div class="scene">
        <div class="door">
          <div class="floor"><i /></div>
          <div class="hole" />
          <div class="leaf l" />
          <div class="leaf r" />
        </div>
        <div class="dust">
          <i v-for="(d, i) in DUST" :key="i" :style="{ '--x': d.x, '--dl': d.dl, top: d.top }" />
        </div>
        <div class="fly">
          <LightCover class="fcv" :src="cover" :seed="seed" />
          <h4>{{ title }}</h4>
          <small>{{ meta }}</small>
        </div>
        <div class="done-t">
          <h2>{{ t('studio.write.doorTitle') }}</h2>
          <p>{{ t('studio.write.doorSub', { title: shortTitle() }) }}</p>
          <div><span class="url"><SIcon name="link" :size="16" />{{ url }}</span></div>
          <div class="row">
            <button type="button" class="st-btn p lg" @click="emit('view')">{{ t('studio.write.viewPost') }}</button>
            <button type="button" class="st-btn w lg" @click="emit('close')">{{ t('studio.write.backToday') }}</button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped lang="scss">
.stage {
  position: fixed;
  inset: 0;
  z-index: 90;
  display: grid;
  place-items: center;
  overflow: hidden;
  pointer-events: none;
  opacity: 0;
  background: radial-gradient(70% 60% at 50% 72%, color-mix(in oklab, var(--primary) 16%, #0d1526), #070a12 72%);
  transition: opacity 0.5s var(--ease-out);

  &.on { opacity: 1; pointer-events: auto; }
}

.bloom {
  position: absolute;
  left: 50%;
  top: 52%;
  width: 1400px;
  height: 1000px;
  transform: translate(-50%, -50%) scale(0.2);
  border-radius: 50%;
  background: radial-gradient(closest-side, color-mix(in oklab, var(--primary) 30%, transparent), transparent);
  opacity: 0;
}

.scene {
  position: relative;
  width: 560px;
  height: 560px;
  perspective: 1100px;
}

.door {
  position: absolute;
  left: 50%;
  top: 70px;
  width: 200px;
  height: 300px;
  margin-left: -100px;
  transform-style: preserve-3d;
  perspective: 900px;
  transition: transform 1.2s var(--ease-out), opacity 1s var(--ease-out);

  .hole {
    position: absolute;
    inset: 0;
    border-radius: 6px;
    background: linear-gradient(#fff, color-mix(in oklab, var(--primary) 18%, #fff));
    box-shadow: 0 0 60px 10px color-mix(in oklab, var(--primary) 50%, transparent), 0 0 160px 40px color-mix(in oklab, var(--primary) 30%, transparent);
    opacity: 0;
    transform: scale(0.96);

    &::after {
      content: '';
      position: absolute;
      inset: 0;
      border-radius: inherit;
      background: linear-gradient(color-mix(in oklab, var(--primary) 35%, #fff), color-mix(in oklab, var(--primary) 75%, #0d1526));
      opacity: 0;
      transition: opacity 0.9s var(--ease-out);
    }
  }

  .leaf {
    position: absolute;
    top: 0;
    height: 100%;
    width: 50%;
    transition: transform 1.1s cubic-bezier(0.5, 0, 0.2, 1);

    &.l {
      left: 0;
      transform-origin: left center;
      background: linear-gradient(95deg, #fdfdff 60%, #dfe3ec);
      border-radius: 12px 3px 3px 12px;
      box-shadow: inset -1px 0 0 rgba(0, 0, 0, 0.06);
    }

    &.r {
      right: 0;
      transform-origin: right center;
      background: linear-gradient(165deg, color-mix(in oklab, var(--primary) 70%, #fff), var(--primary) 40%, var(--primary-deep));
      border-radius: 3px 12px 12px 3px;
    }
  }

  .floor {
    position: absolute;
    left: -180px;
    right: -180px;
    top: 100%;
    height: 260px;
    transform-origin: top;
    transform: scaleY(0);
    filter: blur(10px);
    opacity: 0;

    i {
      position: absolute;
      inset: 0;
      clip-path: polygon(calc(50% - 100px) 0, calc(50% + 100px) 0, 100% 100%, 0 100%);
      background: linear-gradient(rgba(255, 255, 255, 0.95), color-mix(in oklab, var(--primary) 45%, transparent) 50%, transparent 95%);
    }
  }
}

.stage.open {
  .door .leaf.l { transform: rotateY(-64deg); }
  .door .leaf.r { transform: rotateY(64deg); }
  .door .hole { opacity: 1; transform: none; transition: opacity 0.6s ease-out 0.15s, transform 0.9s var(--ease-out); }
  .door .floor { transform: scaleY(1); opacity: 0.9; transition: transform 1.1s var(--ease-out) 0.25s, opacity 0.8s ease-out 0.25s; }
  .bloom { opacity: 1; transform: translate(-50%, -50%) scale(1); transition: transform 1.6s var(--ease-out) 0.2s, opacity 1.2s ease-out 0.2s; }
  .dust i { animation: dust 3.4s ease-out infinite; animation-delay: var(--dl); }
}

.stage.out {
  .door { transform: translateY(-58px) scale(1.22); }
  .door .hole::after { opacity: 0.92; }
  .door .hole { box-shadow: 0 0 90px 20px color-mix(in oklab, var(--primary) 55%, transparent), 0 0 220px 60px color-mix(in oklab, var(--primary) 30%, transparent); }
  .fly { transform: translateZ(0) rotate(-2deg); opacity: 1; }
}

.dust {
  position: absolute;
  left: 50%;
  top: 370px;
  width: 1px;
  height: 1px;

  i {
    position: absolute;
    width: 3px;
    height: 3px;
    border-radius: 50%;
    background: #fff;
    opacity: 0;
    box-shadow: 0 0 6px #fff;
  }
}

@keyframes dust {
  0% { opacity: 0; transform: translate(var(--x), 0); }
  20% { opacity: 0.9; }
  100% { opacity: 0; transform: translate(calc(var(--x) * 1.6), -260px); }
}

.fly {
  position: absolute;
  left: 50%;
  top: 150px;
  width: 262px;
  margin-left: -131px;
  border-radius: 16px;
  background: #fff;
  padding: 8px;
  color: #1e1c19;
  box-shadow: 0 40px 80px -20px rgba(0, 0, 0, 0.7), 0 0 0 1px rgba(255, 255, 255, 0.1);
  transform: translateZ(-600px) scale(0.5);
  opacity: 0;
  transition: transform 1.1s var(--ease-spring), opacity 0.5s ease-out;

  .fcv { aspect-ratio: 16 / 10; border-radius: 11px; }
  h4 { font: 600 15.5px/1.5 var(--font-serif); margin: 10px 6px 3px; }
  small { display: block; font-size: 12px; color: #8c877e; margin: 0 6px 6px; }
}

.done-t {
  position: absolute;
  left: -200px;
  right: -200px;
  top: 432px;
  text-align: center;
  color: #fff;

  > * { opacity: 0; transform: translateY(12px); transition: all 0.7s var(--ease-out); }

  h2 { font: 700 34px/1.3 var(--font-serif); margin: 0 0 8px; letter-spacing: 0.06em; transition-delay: 0.1s; }
  p { margin: 0 0 24px; font-size: 14px; color: rgba(255, 255, 255, 0.62); transition-delay: 0.2s; }

  .url {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font: 12.5px var(--font-mono);
    color: rgba(255, 255, 255, 0.8);
    padding: 6px 12px;
    border-radius: 9px;
    background: rgba(255, 255, 255, 0.08);
    margin-bottom: 22px;
    box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.08) inset;
  }

  > div:nth-of-type(1) { transition-delay: 0.25s; }

  .row { display: flex; gap: 10px; justify-content: center; transition-delay: 0.3s; }
}

.stage.fin .done-t > * { opacity: 1; transform: none; }
</style>
