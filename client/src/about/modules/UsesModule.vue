<script setup lang="ts">
import { safeHref } from '../../utils/safeUrl';
import { computed } from 'vue';
import type { UsesData } from '../types';
import type { ModProps } from './props';
import KitIcon from '../parts/KitIcon.vue';
import ModHead from '../parts/ModHead.vue';

/** 工作台（uses）：硬件 / 软件等分组，每条图标 + 名称 + 规格 + 标签（旧 devices 读取时迁移至此） */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as UsesData);
</script>

<template>
  <ModHead :title="title">/uses</ModHead>
  <div class="us" :class="{ single: d.groups.length < 2 }">
    <div v-for="(g, gi) in d.groups" :key="gi" class="us-g">
      <h5>{{ g.title }}</h5>
      <ul>
        <li v-for="(it, i) in g.items" :key="i">
          <span class="ic"><KitIcon :name="it.icon || 'link'" :size="20" /></span>
          <span class="tx">
            <a v-if="it.url" :href="safeHref(it.url)" target="_blank" rel="noopener noreferrer"><b>{{ it.name }}</b></a>
            <b v-else>{{ it.name }}</b>
            <span v-if="it.desc">{{ it.desc }}</span>
          </span>
          <em v-if="it.tag" class="ak-tag">{{ it.tag }}</em>
        </li>
      </ul>
    </div>
  </div>
</template>

<style scoped lang="scss">
.us { display: grid; grid-template-columns: 1fr 1fr; gap: 10px 28px; }
.us.single { grid-template-columns: 1fr; }

.us-g {
  min-width: 0;

  h5 {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 4px;
    font: 500 12.5px var(--font-sans);
    letter-spacing: 0.04em;
    color: var(--ak-text-3);

    &::after { content: ''; flex: 1; height: 1px; background: var(--ak-line); }
  }

  ul { list-style: none; }

  li {
    display: grid;
    grid-template-columns: 38px minmax(0, 1fr) auto;
    gap: 12px;
    align-items: center;
    padding: 11px 0;
    border-bottom: 1px solid var(--ak-line);

    &:last-child { border-bottom: 0; }
  }

  .ic {
    display: grid;
    place-items: center;
    width: 38px;
    height: 38px;
    border-radius: var(--r-sm);
    background: var(--ak-sunken);
    color: var(--text-2);
    transition: color var(--dur-fast), box-shadow var(--dur-fast);
  }

  li:hover .ic { color: var(--text); background: var(--fill-2); }

  .tx { min-width: 0; }
  b { display: block; font-size: 15px; font-weight: 500; }
  .tx > span { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 13px; color: var(--ak-text-3); }
  a:hover b { color: var(--ak-ink); }

  em { font-size: 13px; }
}

@container (max-width: 560px) { .us { grid-template-columns: 1fr; } }
</style>
