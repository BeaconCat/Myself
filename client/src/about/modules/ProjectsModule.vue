<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { Project, ProjectsData } from '../types';
import type { ModProps } from './props';
import KitIcon from '../parts/KitIcon.vue';
import ModHead from '../parts/ModHead.vue';
import Scene from '../parts/Scene.vue';

/** 作品（projects）：一大两小 —— 精选项目大封面，其余横向卡；无图时用 CSS 光影构图，悬停封面缓推 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as ProjectsData);
const { t } = useI18n();

const ordered = computed(() => {
  const items = [...d.value.items];
  const fi = items.findIndex((p) => p.featured);
  if (fi > 0) items.unshift(...items.splice(fi, 1));
  return { featured: items[0] as Project | undefined, rest: items.slice(1) };
});
const external = (url: string) => (/^https?:/.test(url) ? '_blank' : undefined);
</script>

<template>
  <ModHead :title="title">{{ t('aboutKit.projects.count', { n: d.items.length }) }}</ModHead>
  <div v-if="ordered.featured" class="pj" :class="{ solo: !ordered.rest.length }">
    <a :href="ordered.featured.url || undefined" :target="external(ordered.featured.url)" rel="noopener" class="big">
      <div class="cvw"><Scene :src="ordered.featured.cover" :scene="ordered.featured.scene" :alt="ordered.featured.name" /></div>
      <div class="bd">
        <div class="tt"><b>{{ ordered.featured.name }}</b><KitIcon name="arrow" :size="15" /></div>
        <p>{{ ordered.featured.desc }}</p>
        <div class="ft">
          <span v-if="ordered.featured.lang" class="lang" :style="{ '--c': ordered.featured.color || 'var(--primary)' }">{{ ordered.featured.lang }}</span>
          <span v-if="ordered.featured.stars != null"><KitIcon name="star" :size="13" />{{ ordered.featured.stars }}</span>
        </div>
      </div>
    </a>
    <div v-if="ordered.rest.length" class="side">
      <a v-for="p in ordered.rest" :key="p.name" :href="p.url || undefined" :target="external(p.url)" rel="noopener">
        <div class="cvw"><Scene :src="p.cover" :scene="p.scene" :alt="p.name" /></div>
        <div class="bd">
          <div class="tt"><b>{{ p.name }}</b><KitIcon name="arrow" :size="15" /></div>
          <p>{{ p.desc }}</p>
          <div class="ft">
            <span v-if="p.lang" class="lang" :style="{ '--c': p.color || 'var(--primary)' }">{{ p.lang }}</span>
            <span v-if="p.stars != null"><KitIcon name="star" :size="13" />{{ p.stars }}</span>
          </div>
        </div>
      </a>
    </div>
  </div>
</template>

<style scoped lang="scss">
.pj {
  display: grid;
  grid-template-columns: 1.25fr 1fr;
  gap: 16px;
  flex: 1;

  &.solo { grid-template-columns: 1fr; }

  a {
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border-radius: var(--r-md);
    background: var(--ak-sunken);
    box-shadow: inset 0 0 0 1px var(--ak-line);
    transition: transform var(--dur) var(--ease-out), box-shadow var(--dur) var(--ease-out), background-color var(--dur);

    &:hover { transform: translateY(-3px); background: var(--fill-2); box-shadow: inset 0 0 0 1px var(--ak-line-2), var(--shadow-card-hover); }
    &:hover .cvw > * { transform: scale(1.04); }
    &:hover .tt svg { color: var(--text); transform: translate(2px, -2px); }
  }
}

.cvw {
  flex: 1;
  min-height: 200px;
  overflow: hidden;

  > * { transition: transform 0.8s var(--ease-out); }
}

.bd { display: flex; flex-direction: column; gap: 6px; flex: 1; padding: 14px 16px 16px; }

.tt {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;

  b { font: 700 16px var(--font-serif); }
  svg { color: var(--ak-text-3); transition: transform var(--dur) var(--ease-spring), color var(--dur); }
}

.bd p { font-size: 13px; line-height: 1.65; color: var(--text-2); }

.ft {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-top: auto;
  padding-top: 8px;
  font: 400 11.5px var(--ak-mono);
  color: var(--ak-text-3);

  span { display: flex; align-items: center; gap: 5px; }
  .lang::before { content: ''; width: 8px; height: 8px; border-radius: 50%; background: var(--c); }
}

.side {
  display: flex;
  flex-direction: column;
  gap: 16px;

  a { flex: 1; flex-direction: row; }
  .cvw { flex: none; width: 42%; min-height: 0; }
  .bd { flex: 1; min-width: 0; }
}

@container (max-width: 560px) { .pj { grid-template-columns: 1fr; } }

@container (max-width: 380px) {
  .side .cvw { width: 30%; }
  .side p { display: none; }
}
</style>
