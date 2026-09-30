<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { PlacesData } from '../types';
import type { ModProps } from './props';
import CountUp from '../parts/CountUp.vue';
import ModHead from '../parts/ModHead.vue';
import KitIcon from '../parts/KitIcon.vue';
import { ICONS } from '../icons';
import { MAP_VIEWBOX, mapDots, project } from '../geo';

/**
 * 足迹（places）：本地点阵地图 + 城市发光脉冲 + 常住地到各城市的弧线；悬停与列表联动。
 * 常住地用星标（Lucide Star）区分形状而非颜色：任何色盘（包括黄色主题）下都一眼可辨。
 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as PlacesData);
const { t } = useI18n();
const hover = ref<number | null>(null);
const dots = mapDots();

const home = computed(() => d.value.items.find((p) => p.home) ?? d.value.items[0]);

const pts = computed(() => d.value.items.map((p) => {
  const [x, y] = project(p.lon, p.lat);
  return { ...p, x, y };
}));

const routes = computed(() => {
  if (!home.value) return [];
  const [hx, hy] = project(home.value.lon, home.value.lat);
  return pts.value.filter((p) => p !== undefined && p.name !== home.value?.name).map((p) => {
    const mx = (p.x + hx) / 2;
    const my = (p.y + hy) / 2 - Math.hypot(p.x - hx, p.y - hy) * 0.25;
    return `M${hx} ${hy} Q${mx} ${my} ${p.x} ${p.y}`;
  });
});

const since = computed(() => {
  const years = d.value.items.map((p) => Number(String(p.year ?? '').slice(0, 4))).filter((y) => y > 1900);
  return years.length ? Math.min(...years) : null;
});
const span = computed(() => (since.value ? new Date().getFullYear() - since.value + 1 : 0));
</script>

<template>
  <ModHead :title="title"><template v-if="home">{{ t('aboutKit.places.home', { city: home.name }) }}</template></ModHead>
  <div class="pl" @pointerleave="hover = null">
    <div class="pl-map">
      <svg :viewBox="MAP_VIEWBOX" role="img" :aria-label="title">
        <circle v-for="([x, y], i) in dots" :key="i" class="d" :cx="x" :cy="y" r=".3" />
        <path v-for="(r, i) in routes" :key="`r${i}`" class="route" :d="r" />
        <g v-for="(p, i) in pts" :key="`p${i}`">
          <circle class="pulse" :class="{ home: p.home }" :cx="p.x" :cy="p.y" r=".7" />
          <!-- 常住地：实心星标（地图坐标系内嵌 24 网格图标，边长约 2.6 个单位） -->
          <!-- eslint-disable-next-line vue/no-v-html -->
          <svg
            v-if="p.home"
            class="c star"
            viewBox="0 0 24 24"
            :x="p.x - (hover === i ? 1.7 : 1.3)"
            :y="p.y - (hover === i ? 1.7 : 1.3)"
            :width="hover === i ? 3.4 : 2.6"
            :height="hover === i ? 3.4 : 2.6"
            :data-tip="`${p.name}${p.year ? ` · ${p.year}` : ''}`"
            @pointerenter="hover = i"
            v-html="ICONS.star"
          />
          <circle
            v-else
            class="c"
            :cx="p.x"
            :cy="p.y"
            :r="hover === i ? 0.95 : 0.62"
            :data-tip="`${p.name}${p.year ? ` · ${p.year}` : ''}`"
            @pointerenter="hover = i"
          />
        </g>
      </svg>
    </div>
    <div class="pl-side">
      <div class="ak-statbar pl-sum" :style="{ '--n': span ? 2 : 1 }">
        <div class="ak-stat"><b><CountUp :value="d.items.length" /></b><span>{{ t('aboutKit.places.cities') }}</span></div>
        <div v-if="span" class="ak-stat"><b><CountUp :value="span" /></b><span>{{ t('aboutKit.places.span', { y: since }) }}</span></div>
      </div>
      <ul class="pl-list">
        <li v-for="(p, i) in d.items" :key="i" :class="{ on: hover === i, home: p.home }" @pointerenter="hover = i">
          <span class="nm">{{ p.name }}<KitIcon v-if="p.home" class="hs" name="star" :size="13" /></span><small>{{ p.year }}</small>
        </li>
      </ul>
    </div>
  </div>
</template>

<style scoped lang="scss">
.pl { display: grid; grid-template-columns: 1fr; gap: 18px; }

.pl-map svg { display: block; width: 100%; height: auto; overflow: visible; }

.d { fill: var(--ak-text-3); opacity: 0.32; }

.c {
  fill: var(--primary);
  cursor: pointer;
  transition: r var(--dur) var(--ease-spring);
}

.star {
  overflow: visible;
  fill: var(--primary);
  stroke: var(--primary);
  stroke-width: 1.5;
  stroke-linejoin: round;
}

.pulse {
  fill: none;
  stroke: var(--primary);
  stroke-width: 0.18;
  transform-box: fill-box;
  transform-origin: center;
  animation: pl-ring 2.6s var(--ease-out) infinite;
}

@keyframes pl-ring { 0% { transform: scale(0.6); opacity: 1; } 100% { transform: scale(3.2); opacity: 0; } }

.route {
  fill: none;
  stroke: color-mix(in oklab, var(--primary) 50%, transparent);
  stroke-width: 0.16;
  stroke-dasharray: 0.6 0.5;
}

.in .route { animation: pl-flow 30s linear infinite; }

@keyframes pl-flow { to { stroke-dashoffset: -22; } }

.pl-sum { margin-bottom: 10px; }

.pl-list {
  list-style: none;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0 18px;

  li {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    padding: 8px 0;
    border-bottom: 1px solid var(--ak-line);
    font-size: 15px;
    font-weight: 500;
    cursor: default;
    transition: color var(--dur-fast);

    small { font: 400 12.5px var(--ak-mono); color: var(--ak-text-3); }
    &.on { color: var(--ak-ink); }
    .nm { display: inline-flex; align-items: center; gap: 6px; }

    /* 常住地：名字后一枚实心小星，颜色随正文 */
    .hs { fill: currentColor; opacity: 0.75; }
  }
}

@container (min-width: 500px) {
  .pl { grid-template-columns: 1.4fr 1fr; align-items: center; gap: 28px; }
  .pl-list { grid-template-columns: 1fr; }
}
</style>
