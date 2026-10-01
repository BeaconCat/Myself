<script setup lang="ts">
import { computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Clock } from 'lucide';
import Icon from '../ui/Icon.vue';
import StSeg from '../../views/admin/studio/StSeg.vue';
import { useConfigStore } from '../../stores/config';
import { useNow } from '../../composables/useNow';
import { scheduleInput, scheduleValid, scheduleLabel } from '../../utils/publication';

const props = withDefaults(defineProps<{ disabled?: boolean; allowSchedule?: boolean }>(), { allowSchedule: true });
// null = publish now; an empty string remains an unfinished scheduled choice.
const value = defineModel<string | null>({ default: null });
const emit = defineEmits<{ valid: [valid: boolean] }>();
const { t } = useI18n();
const config = useConfigStore();
const now = useNow();
const zone = computed(() => config.cfg.timezone);
const scheduled = computed(() => value.value !== null);
const valid = computed(() => scheduleValid(value.value, zone.value, now.value));
watch(valid, v => emit('valid', v), { immediate: true });
function later(hours = 1): void {
  const date = new Date(now.value + hours * 3600_000);
  date.setUTCSeconds(0, 0);
  value.value = scheduleInput(date.toISOString(), zone.value);
}
function tomorrow(): void {
  const date = new Date(`${scheduleInput(new Date(now.value).toISOString(), zone.value).slice(0, 10)}T00:00:00Z`);
  date.setUTCDate(date.getUTCDate() + 1);
  value.value = `${date.toISOString().slice(0, 10)}T09:00:00`;
}
const mode = computed({ get: () => scheduled.value ? 'scheduled' : 'now', set: v => { if (v === 'now') value.value = null; else later(); } });
const input = computed({ get: () => scheduleInput(value.value ?? '', zone.value), set: v => { value.value = v; } });
const min = computed(() => scheduleInput(new Date(now.value + 1000).toISOString(), zone.value));
</script>

<template>
  <section class="publish-timing studio" :inert="disabled">
    <div class="timing-head"><span><Icon :icon="Clock" :size="15" />{{ t('schedule.when') }}</span>
      <StSeg v-if="allowSchedule" v-model="mode" :label="t('schedule.when')"
        :options="[{ value: 'now', label: t('schedule.now') }, { value: 'scheduled', label: t('schedule.later') }]" />
      <small v-else>{{ t('schedule.now') }}</small>
    </div>
    <div class="timing-reveal" :class="{ on: scheduled && props.allowSchedule }" :inert="!scheduled || !allowSchedule" :aria-hidden="!scheduled || !allowSchedule">
      <div><div class="timing-fields">
        <label class="st-field" :class="{ invalid: !valid }"><input v-model="input" type="datetime-local" step="1" :min="min" :disabled="disabled" :aria-label="t('schedule.dateTime')" :aria-invalid="!valid" /></label>
        <div class="presets"><button type="button" :disabled="disabled" @click="later()">{{ t('schedule.inHour') }}</button><button type="button" :disabled="disabled" @click="tomorrow">{{ t('schedule.tomorrow') }}</button></div>
        <p class="zone">{{ t('schedule.timezone', { zone }) }}</p>
        <p v-if="!valid" class="invalid" role="status">{{ t('schedule.invalid') }}</p>
        <p v-else class="summary">{{ t('schedule.summary', { when: scheduleLabel(value ?? '', zone) }) }}</p>
      </div></div>
    </div>
  </section>
</template>

<style scoped lang="scss">
.publish-timing { min-width: 0; }
.timing-head { display: flex; justify-content: space-between; align-items: center; gap: 12px; }
.timing-head > span { display: inline-flex; align-items: center; gap: 7px; font-size: 13px; color: var(--st-ink-2); }
.timing-head > small { color: var(--st-ink-3); }
.timing-reveal { display: grid; grid-template-rows: 0fr; opacity: 0; transition: grid-template-rows .3s var(--ease-out), opacity .2s; }
.timing-reveal.on { grid-template-rows: 1fr; opacity: 1; }
.timing-reveal > div { min-height: 0; overflow: hidden; }
.timing-fields { padding: 12px; margin-top: 10px; border-radius: var(--r-md); background: var(--well); }
.st-field input { width: 100%; min-width: 0; font: 13px/1.6 var(--font-mono); }
.presets { display: flex; gap: 8px; margin-top: 10px; }
.presets button { padding: 5px 9px; border: 1px solid var(--line-2); border-radius: var(--r-pill); font-size: 12px; color: var(--st-ink-2); transition: background var(--dur-fast), transform var(--dur-fast); }
.presets button:hover { background: var(--paper); }
.presets button:active { transform: scale(.96); }
p { margin: 8px 0 0; font-size: 12px; line-height: 1.6; }
.zone { color: var(--st-ink-3); }
.summary { color: var(--st-ink-2); font-variant-numeric: tabular-nums; }
.invalid { color: var(--red); }
.st-field.invalid { box-shadow: 0 0 0 1px var(--red) inset; }
@media (prefers-reduced-motion: reduce) { .timing-reveal, .presets button { transition: none; } }
</style>
