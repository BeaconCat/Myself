<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { Book, BookStatus, BookshelfData } from '../../../about/types';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 书架：书名 / 作者 / 状态 / 进度 / 书脊色与字色 / 高度与厚度 / 斜放 / 一句书摘 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<BookshelfData>(() => props.mod);
const { t } = useI18n();
const STATUS: BookStatus[] = ['在读', '读完', '想读'];
const make = (): Book => ({ title: '', author: '', color: '#1d3a5f', textColor: '#f4efe4', height: 88, width: 34, status: '读完', note: '' });
</script>

<template>
  <div class="ed">
    <EdList v-slot="{ item }" :items="d.items" :make="make">
      <div class="line">
        <span class="spine" :style="{ background: item.color, color: item.textColor || '#f4efe4', height: `${(item.height ?? 88) * 0.5}px`, width: `${(item.width ?? 34) * 0.6}px` }">{{ item.title.slice(0, 4) }}</span>
        <div class="grid3 flex-in">
          <input v-model="item.title" class="a-input" type="text" :placeholder="t('aboutKit.ed.bookTitle')" />
          <input v-model="item.author" class="a-input" type="text" :placeholder="t('aboutKit.ed.author')" />
          <div class="line">
            <select v-model="item.status" class="a-input flex-in">
              <option v-for="s in STATUS" :key="s" :value="s">{{ s }}</option>
            </select>
            <input v-if="item.status === '在读'" v-model.number="item.progress" class="a-input w-num" type="number" min="0" max="100" :placeholder="t('aboutKit.ed.progress')" />
          </div>
          <input v-model="item.note" class="a-input span2" type="text" :placeholder="t('aboutKit.ed.bookNote')" />
          <div class="line">
            <input v-model="item.color" class="color-in" type="color" :title="t('aboutKit.ed.spineColor')" />
            <input v-model="item.textColor" class="color-in" type="color" :title="t('aboutKit.ed.textColor')" />
            <label class="check"><input v-model="item.lean" type="checkbox" />{{ t('aboutKit.ed.lean') }}</label>
          </div>
          <label class="span-all dims">
            <span>{{ t('aboutKit.ed.height') }} {{ item.height ?? 88 }}% · {{ t('aboutKit.ed.thickness') }} {{ item.width ?? 34 }}px</span>
            <div class="line">
              <input v-model.number="item.height" class="range" type="range" min="70" max="100" />
              <input v-model.number="item.width" class="range" type="range" min="22" max="52" />
            </div>
          </label>
        </div>
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.line { align-items: flex-start; }

.spine {
  flex-shrink: 0;
  align-self: flex-end;
  display: flex;
  justify-content: center;
  padding-top: 6px;
  border-radius: calc(var(--r-xs) * 0.45);
  writing-mode: vertical-rl;
  font: 700 10px var(--font-serif);
  letter-spacing: 0.1em;
  box-shadow: inset 2px 0 0 rgb(255 255 255 / 0.12), inset -3px 0 6px rgb(0 0 0 / 0.25);
  transition: height var(--dur), width var(--dur), background var(--dur);
}
</style>
