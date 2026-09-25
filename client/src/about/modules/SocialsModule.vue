<script setup lang="ts">
import { computed } from 'vue';
import type { SocialsData } from '../types';
import type { ModProps } from './props';
import KitIcon from '../parts/KitIcon.vue';
import ModHead from '../parts/ModHead.vue';

/** 社交（socials）：列表 = 图标 + 名称 + mono handle，悬停箭头右上飞出；pills = 按钮行（primary 为实底主按钮） */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as SocialsData);
const external = (url: string) => (/^https?:/.test(url) ? '_blank' : undefined);
</script>

<template>
  <ModHead :title="title" />
  <div v-if="variant === 'pills'" class="so-pills">
    <a v-for="l in d.items" :key="l.name + l.url" :href="l.url" class="ak-btn" :class="{ pri: l.primary }" :target="external(l.url)" rel="noopener">
      <KitIcon :name="l.icon" />{{ l.name }}
    </a>
  </div>
  <ul v-else class="so">
    <li v-for="l in d.items" :key="l.name + l.url">
      <a :href="l.url" :target="external(l.url)" rel="noopener">
        <span class="ic"><KitIcon :name="l.icon" /></span>
        <span><b>{{ l.name }}</b><span>{{ l.handle || l.url.replace(/^(https?:\/\/|mailto:)/, '') }}</span></span>
        <span class="arr"><KitIcon name="arrow" /></span>
      </a>
    </li>
  </ul>
</template>

<style scoped lang="scss">
.so { list-style: none; }

.so a {
  display: grid;
  grid-template-columns: 36px minmax(0, 1fr) auto;
  gap: 12px;
  align-items: center;
  padding: 11px 10px;
  margin: 0 -10px;
  border-radius: var(--r-md);
  transition: background var(--dur-fast);

  &:hover { background: var(--ak-sunken); }

  b { display: block; font-size: 14px; font-weight: 500; line-height: 1.3; }
  > span > span { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font: 400 11.5px var(--ak-mono); color: var(--ak-text-3); }
}

.ic {
  display: grid;
  place-items: center;
  width: 36px;
  height: 36px;
  border-radius: var(--r-sm);
  box-shadow: inset 0 0 0 1px var(--ak-line-2);
  color: var(--text-2);
  transition: color var(--dur), background-color var(--dur);
}

.so a:hover .ic { color: var(--text); background: var(--fill); }

.arr { color: var(--ak-text-3); transition: transform var(--dur) var(--ease-spring), color var(--dur); }
.so a:hover .arr { transform: translate(3px, -3px); color: var(--text); }

.so-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;

  a { height: 40px; padding: 0 18px 0 15px; }
}
</style>
