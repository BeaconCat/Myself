<script setup lang="ts">
import { siteToday } from '../../../utils/date';
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { NowData } from '../../../about/types';
import Combobox from '../../ui/Combobox.vue';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 此刻：更新日期 + 每条 类别 / 内容 / 注释 / 进度（可空） */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<NowData>(() => props.mod);
const { t } = useI18n();
const KINDS = ['在做', '在学', '在读', '在玩', '在听', '在写'];
const KIND_OPTIONS = KINDS.map((k) => ({ value: k, label: k }));

function touch(): void {
  // eslint-disable-next-line vue/no-mutating-props
  d.value.updatedAt = siteToday();
}
</script>

<template>
  <div class="ed">
    <div class="line">
      <label class="flex-in"><span>{{ t('aboutKit.ed.updatedAt') }}</span><input v-model="d.updatedAt" class="a-input" type="date" /></label>
      <button type="button" class="a-btn ghost sm today" @click="touch">{{ t('aboutKit.ed.today') }}</button>
    </div>
    <EdList v-slot="{ item }" :items="d.items" :make="() => ({ kind: '在做', text: '', note: '' })">
      <div class="line">
        <Combobox v-model="item.kind" :options="KIND_OPTIONS" :min-width="140" :aria-label="t('aboutKit.ed.state')" />
        <input v-model="item.text" class="a-input flex-in" type="text" :placeholder="t('aboutKit.ed.text')" />
      </div>
      <div class="line">
        <input v-model="item.note" class="a-input flex-in" type="text" :placeholder="t('aboutKit.ed.note')" />
        <input
          class="a-input w-num"
          type="number"
          min="0"
          max="100"
          :value="item.progress ?? ''"
          :placeholder="t('aboutKit.ed.progress')"
          @change="item.progress = ($event.target as HTMLInputElement).value === '' ? null : Number(($event.target as HTMLInputElement).value)"
        />
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.line { align-items: flex-end; }
.today { margin-bottom: 1px; }
.line > :deep(.ui-cb) { flex: none; width: 104px; }
</style>
