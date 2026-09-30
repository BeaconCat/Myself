<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import Select from '../../ui/Select.vue';
import { LEGACY_SCENES, SCENES, SCENE_LABELS } from '../../../about/icons';
import { coverUrl } from '../../../utils/defaultCovers';

/**
 * 光影（默认封面）选择：下拉里每项带一张小图。旧版 CSS 场景名 / 空值按 Scene.vue 同样的规则
 * 显示为对应的默认图（未改动时不写回，数据保持原样）。
 */
const props = defineProps<{ modelValue?: string; disabled?: boolean }>();
const emit = defineEmits<{ 'update:modelValue': [value: string] }>();
const { t } = useI18n();

const options = SCENES.map((s) => ({ value: s, label: SCENE_LABELS[s], hint: s }));
const value = computed(() => {
  const s = props.modelValue ?? '';
  return (SCENES as readonly string[]).includes(s) ? s : LEGACY_SCENES[s] ?? '05';
});
</script>

<template>
  <Select
    :model-value="value"
    :options="options"
    :disabled="disabled"
    :min-width="200"
    :aria-label="t('aboutKit.ed.scene')"
    :title="disabled ? t('aboutKit.ed.sceneHint') : t('aboutKit.ed.scene')"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <template #icon="{ option }"><img class="ed-scene-th" :src="coverUrl(option.value, true)" alt="" /></template>
  </Select>
</template>

<style>
.ed-scene-th {
  display: block;
  width: 26px;
  height: 18px;
  border-radius: calc(var(--r-xs) * 0.6);
  object-fit: cover;
  box-shadow: 0 0 0 1px var(--line);
}
</style>
