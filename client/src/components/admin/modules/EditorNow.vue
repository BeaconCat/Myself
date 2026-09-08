<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';

/** 正在做：每行一条 */
defineProps<{ mod: AboutModule }>();
const { t } = useI18n();

function parseLines(text: string): string[] {
  return text.split(/\n+/).map((s) => s.trim()).filter(Boolean);
}
</script>

<template>
  <textarea
    class="a-input"
    rows="4"
    :value="mod.data.items.join('\n')"
    :placeholder="t('admin.nowPlaceholder')"
    @change="mod.data.items = parseLines(($event.target as HTMLTextAreaElement).value)"
  />
</template>
