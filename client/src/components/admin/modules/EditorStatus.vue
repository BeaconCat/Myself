<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { StatusData } from '../../../about/types';
import { useModuleData } from './useModuleData';

/** 在线状态：状态 / 当前活动 / 工具 / 最后活跃（日期时间）/ 设备 / 自定义副题 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<StatusData>(() => props.mod);
const { t } = useI18n();
const STATES = ['online', 'focus', 'away'] as const;

/* datetime-local 需要 'YYYY-MM-DDTHH:mm' */
const toLocal = (iso?: string) => (iso && !Number.isNaN(Date.parse(iso)) ? new Date(Date.parse(iso) - new Date().getTimezoneOffset() * 6e4).toISOString().slice(0, 16) : '');
function setLast(v: string): void {
  // eslint-disable-next-line vue/no-mutating-props
  d.value.lastActive = v ? new Date(v).toISOString() : '';
}
</script>

<template>
  <div class="ed">
    <div class="fld">
      <span>{{ t('aboutKit.ed.state') }}</span>
      <div class="seg">
        <button v-for="s in STATES" :key="s" type="button" :class="{ on: d.state === s }" @click="d.state = s">{{ t(`aboutKit.status.${s}`) }}</button>
      </div>
    </div>
    <div class="grid2">
      <label><span>{{ t('aboutKit.ed.activity') }}</span><input v-model="d.activity" class="a-input" type="text" /></label>
      <label><span>{{ t('aboutKit.ed.app') }}</span><input v-model="d.app" class="a-input" type="text" /></label>
      <label><span>{{ t('aboutKit.status.lastActive') }}</span><input class="a-input" type="datetime-local" :value="toLocal(d.lastActive)" @change="setLast(($event.target as HTMLInputElement).value)" /></label>
      <label><span>{{ t('aboutKit.status.device') }}</span><input v-model="d.device" class="a-input" type="text" /></label>
    </div>
    <label><span>{{ t('aboutKit.ed.statusNote') }}</span><input v-model="d.note" class="a-input" type="text" :placeholder="t(`aboutKit.status.${d.state}Sub`)" /></label>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
