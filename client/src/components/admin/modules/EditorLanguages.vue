<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { LanguagesData } from '../../../about/types';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';
import { useGithubLanguages } from '../../../about/useGithubLanguages';

/** 语言占比：口径说明 + 名称 / 百分比；配色由前台按主色单色阶梯自动生成；合计不足 100 前台自动归一 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<LanguagesData>(() => props.mod);
const { t } = useI18n();
const automatic = computed(() => d.value.source === 'github');
const { data: remote, loading, error, username, refresh } = useGithubLanguages(automatic);
const STEPS = [88, 62, 44, 30, 20];
const ladder = (i: number) => `color-mix(in oklab, var(--primary) ${STEPS[Math.min(i, STEPS.length - 1)]}%, transparent)`;
const sum = computed(() => d.value.items.reduce((a, b) => a + (Number(b.percent) || 0), 0));
</script>

<template>
  <div class="ed">
    <label><span>{{ t('aboutKit.languages.source') }}</span><select v-model="d.source" class="a-input" :aria-label="t('aboutKit.languages.source')"><option value="manual">{{ t('aboutKit.languages.manual') }}</option><option value="github">{{ t('aboutKit.languages.github') }}</option></select></label>
    <template v-if="automatic">
      <p class="hint">{{ t('aboutKit.languages.scope') }}</p>
      <p class="hint">{{ t('aboutKit.languages.account', { name: username || '—' }) }} <router-link :to="{ name: 'admin-settings', query: { full: '1' }, hash: '#set-github' }">{{ t('aboutKit.languages.settings') }}</router-link></p>
      <p class="hint">{{ t('aboutKit.languages.refreshHint') }}</p>
      <button type="button" class="a-btn ghost sm" :disabled="loading || !username" @click="refresh(true)">{{ t(loading ? 'aboutKit.languages.loading' : 'aboutKit.languages.sync') }}</button>
      <p v-if="error" class="hint" role="status">{{ t(remote ? 'aboutKit.languages.stale' : 'aboutKit.languages.failed') }}</p>
      <template v-if="remote">
        <p class="hint">{{ t('aboutKit.languages.result', { n: remote.repositories, date: new Date(remote.fetchedAt).toLocaleString() }) }}</p>
        <p v-if="!remote.items.length" class="hint">{{ t('aboutKit.languages.empty') }}</p>
        <ul class="synced"><li v-for="item in remote.items" :key="item.name"><span>{{ item.name }}</span><b>{{ Number(item.percent.toFixed(2)) }}%</b></li></ul>
      </template>
    </template>
    <template v-else>
    <label><span>{{ t('aboutKit.ed.unit') }}</span><input v-model="d.unit" class="a-input" type="text" /></label>
    <div class="bar" aria-hidden="true">
      <i v-for="(it, i) in d.items" :key="i" :style="{ flex: Math.max(0, it.percent), background: ladder(i) }" />
    </div>
    <p class="hint">{{ t('aboutKit.ed.sum', { n: sum }) }}{{ t('aboutKit.ed.langColorHint') }}</p>
    <EdList v-slot="{ item }" :items="d.items" :make="() => ({ name: '', percent: 10, color: '#8a96ab' })" compact>
      <div class="line">
        <input v-model="item.name" class="a-input flex-in" type="text" :placeholder="t('aboutKit.ed.itemName')" />
        <input v-model.number="item.percent" class="a-input w-num" type="number" min="0" max="100" />
      </div>
    </EdList>
    </template>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.bar { display: flex; gap: 3px; height: 10px; border-radius: var(--r-pill); overflow: hidden; }
.bar i { display: block; min-width: 2px; border-radius: calc(var(--r-xs) * 0.5); }
.synced { list-style: none; padding: 0; display: grid; gap: 8px; }.synced li { display: flex; justify-content: space-between; gap: 12px; font-size: 14px; }.synced span { overflow-wrap: anywhere; }.synced b { flex: none; }
</style>
