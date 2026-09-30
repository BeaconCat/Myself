<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { Book, BookStatus, BookshelfData } from '../../../about/types';
import Select from '../../ui/Select.vue';
import ColorSwatch from '../../ui/ColorSwatch.vue';
import Switch from '../../ui/Switch.vue';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/**
 * 书架：左侧书脊实时预览（颜色 / 字色 / 高度 / 厚度 / 斜放），右侧
 * 书名 · 作者 · 状态（在读时显示进度）/ 一句书摘 / 书脊色 · 字色 · 高度 · 厚度 · 斜放。
 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<BookshelfData>(() => props.mod);
const { t } = useI18n();

/** 状态值沿用存储里的中文取值，显示文案走字典 */
const STATUS: [BookStatus, string][] = [['在读', 'reading'], ['读完', 'done'], ['想读', 'want']];
const statusOptions = computed(() => STATUS.map(([value, k]) => ({ value, label: t(`aboutKit.ed.bookStatus.${k}`) })));

const make = (): Book => ({ title: '', author: '', color: '#1d3a5f', textColor: '#f4efe4', height: 88, width: 34, status: '读完', note: '' });

/** 预览书脊：高度按 120px 书格折算，厚度 0.8 倍 */
const spineStyle = (b: Book) => ({
  background: b.color,
  color: b.textColor || '#f4efe4',
  height: `${((b.height ?? 88) / 100) * 120}px`,
  width: `${(b.width ?? 34) * 0.8}px`,
});
</script>

<template>
  <div class="ed">
    <EdList v-slot="{ item }" :items="d.items" :make="make">
      <div class="bk">
        <div class="shelf" aria-hidden="true">
          <span class="spine" :class="{ lean: item.lean }" :style="spineStyle(item)">{{ item.title.slice(0, 6) }}</span>
        </div>
        <div class="bk-main">
          <div class="bk-r1">
            <input v-model="item.title" class="a-input nm" type="text" :placeholder="t('aboutKit.ed.bookTitle')" :aria-label="t('aboutKit.ed.bookTitle')" />
            <input v-model="item.author" class="a-input" type="text" :placeholder="t('aboutKit.ed.author')" :aria-label="t('aboutKit.ed.author')" />
            <Select v-model="item.status" :options="statusOptions" :aria-label="t('aboutKit.ed.state')" />
            <label class="affix prog" :class="{ off: item.status !== '在读' }" :title="t('aboutKit.ed.progress')">
              <span>{{ t('aboutKit.ed.progress') }}</span>
              <input v-model.number="item.progress" type="number" min="0" max="100" :disabled="item.status !== '在读'" placeholder="0" />
              <span class="aff">%</span>
            </label>
          </div>
          <input v-model="item.note" class="a-input" type="text" :placeholder="t('aboutKit.ed.bookNote')" :aria-label="t('aboutKit.ed.bookNote')" />
          <div class="bk-r3">
            <div class="colors">
              <ColorSwatch v-model="item.color" :label="t('aboutKit.ed.spineColor')" show-hex />
              <ColorSwatch v-model="item.textColor" :label="t('aboutKit.ed.textColor')" fallback="#f4efe4" />
            </div>
            <div class="sl">
              <span class="sl-h"><span>{{ t('aboutKit.ed.height') }}</span><b>{{ item.height ?? 88 }}%</b></span>
              <input v-model.number="item.height" class="range" type="range" min="70" max="100" :aria-label="t('aboutKit.ed.height')" />
            </div>
            <div class="sl">
              <span class="sl-h"><span>{{ t('aboutKit.ed.thickness') }}</span><b>{{ item.width ?? 34 }}px</b></span>
              <input v-model.number="item.width" class="range" type="range" min="22" max="52" :aria-label="t('aboutKit.ed.thickness')" />
            </div>
            <Switch v-model="item.lean" class="lean-sw">{{ t('aboutKit.ed.lean') }}</Switch>
          </div>
        </div>
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.bk { display: flex; gap: 14px; align-items: stretch; }

/* 书格：底部一条隔板，书脊立在上面 */
.shelf {
  flex: none;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  width: 72px;
  padding: 8px 0 6px;
  border-radius: var(--r-sm);
  background: var(--well);
  box-shadow: 0 0 0 1px var(--line) inset, inset 0 -5px 0 var(--line-2);
}

.spine {
  display: flex;
  justify-content: center;
  padding-top: 8px;
  overflow: hidden;
  border-radius: calc(var(--r-xs) * 0.45);
  writing-mode: vertical-rl;
  font: 700 11px var(--font-serif);
  letter-spacing: 0.1em;
  box-shadow: inset 2px 0 0 rgb(255 255 255 / 0.12), inset -3px 0 6px rgb(0 0 0 / 0.25);
  transform-origin: bottom left;
  transition: height var(--dur) var(--ease-out), width var(--dur) var(--ease-out), background-color var(--dur), color var(--dur), transform var(--dur) var(--ease-spring);

  &.lean { transform: rotate(8deg); }
}

.bk-main { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 8px; }

.bk-r1 {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(0, 1fr) minmax(96px, 120px) minmax(110px, 130px);
  gap: 8px;
}
.nm { font-weight: 600; }

.prog {
  b, .aff { font-size: 12px; color: var(--text-2); }
  &.off { opacity: 0.45; }
}

.bk-r3 {
  display: grid;
  grid-template-columns: auto minmax(120px, 1fr) minmax(120px, 1fr) auto;
  align-items: center;
  gap: 8px 18px;
}

.colors { display: flex; gap: 6px; }

/* 滑杆：上方 标签 · 数值 同行，下方滑杆 */
.sl { display: flex; flex-direction: column; gap: 4px; min-width: 0; }
.sl-h {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  font-size: 12px;

  span { font-weight: 600; color: var(--text-2); }
  b { font: 500 12px var(--font-mono); font-variant-numeric: tabular-nums; color: var(--text); }
}

@container ed (max-width: 820px) {
  .bk-r1 { grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); }
  /* 窄时：色块与斜放一行，两根滑杆一行 */
  .bk-r3 { grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); }
  .lean-sw { order: -1; justify-self: end; }
  .colors { order: -2; }
}

@container ed (max-width: 520px) {
  .bk-r1, .bk-r3 { grid-template-columns: minmax(0, 1fr); }
  .shelf { width: 56px; }
}

@media (prefers-reduced-motion: reduce) {
  .spine { transition: none; }
}
</style>
