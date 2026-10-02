<script setup lang="ts">
import { useConfigStore } from '../../stores/config';
import { scheduleInput } from '../../utils/publication';
import { siteToday } from '../../utils/date';
import Icon from '../../components/ui/Icon.vue';
import { Check } from 'lucide';
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import '../admin/studio/i18n';
import SIcon from '../admin/studio/SIcon.vue';
import StModal from '../admin/studio/StModal.vue';
import PublishTiming from '../../components/admin/PublishTiming.vue';
import LightCover from '../admin/studio/LightCover.vue';

/** 发布确认：左侧「读者将看到」预览卡，右侧发布选项与发布前检查 */
const props = defineProps<{
  open: boolean;
  title: string;
  excerpt: string;
  cover: string;
  seed: number | string;
  tags: string[];
  slug: string;
  minutes: number;
  republish: boolean;
  busy: boolean;
  /** 协作作者：不能置顶、不发随想预告 */
  author?: boolean;
  /** 作者投稿需站长审阅：发布改为「提交审阅」 */
  review?: boolean;
}>();
const publishAt = defineModel<string | null>('publishAt', { default: null });
const scheduleOK = ref(true);
const pinned = defineModel<boolean>('pinned', { default: false });
const announce = defineModel<boolean>('announce', { default: false });
const emit = defineEmits<{ close: []; confirm: [] }>();
const { t } = useI18n();

const today = siteToday();
const config = useConfigStore();
const previewDate = computed(() => publishAt.value === null ? today : scheduleInput(publishAt.value || '', config.cfg.timezone).slice(0, 10) || t('schedule.pending'));
const checks = computed(() => [
  { ok: !!props.title.trim(), label: t('studio.write.ckTitle') },
  { ok: !!props.excerpt.trim(), label: t('studio.write.ckExcerpt') },
  { ok: !!props.cover, label: t('studio.write.ckCover') },
  { ok: props.tags.length > 0, label: t('studio.write.ckTags') },
]);

function onKey(e: KeyboardEvent): void {
  if ((e.ctrlKey || e.metaKey) && e.key === 'Enter' && !props.busy && scheduleOK.value) {
    e.preventDefault();
    emit('confirm');
  }
}
watch(
  () => props.open,
  (v) => (v ? document.addEventListener('keydown', onKey) : document.removeEventListener('keydown', onKey)),
);
onBeforeUnmount(() => document.removeEventListener('keydown', onKey));
</script>

<template>
  <StModal :open="open" panel-class="pub" @close="emit('close')">
    <div class="pub-l">
      <div class="cap">{{ t('studio.write.readersSee') }}</div>
      <div class="pv-card">
        <LightCover class="pcv" :src="cover" :seed="seed" />
        <h4>{{ title || t('studio.untitled') }}</h4>
        <p class="post-excerpt">{{ excerpt || t('studio.posts.noExcerpt') }}</p>
        <div class="meta">
          <span v-if="tags[0]" class="tag">{{ tags[0] }}</span>
          <span class="mono">{{ previewDate }}</span>
          <span>{{ t('studio.write.minutes', { n: minutes }) }}</span>
        </div>
      </div>
      <div class="url"><SIcon name="link" :size="16" />/articles/{{ slug }}</div>
    </div>

    <div class="pub-r">
      <h3>{{ review ? t('studio.write.reviewTitle') : publishAt !== null ? t('schedule.postTitle') : republish ? t('studio.write.pubUpdateTitle') : t('studio.write.pubTitle') }}</h3>

      <PublishTiming v-model="publishAt" :disabled="busy" :allow-schedule="!review && !republish" @valid="scheduleOK = $event" />

      <p v-if="review" class="review-note"><SIcon name="clock" :size="16" />{{ t('studio.write.reviewNote') }}</p>
      <div v-if="!author" class="st-flabel">{{ t('studio.write.after') }}</div>
      <div v-if="!author" class="opts">
        <div class="st-ckrow" role="checkbox" tabindex="0" :aria-checked="pinned" @click.prevent="pinned = !pinned" @keydown.enter.space.prevent="pinned = !pinned">
          <span class="st-ck" :class="{ on: pinned }"><Icon :icon="Check" /></span>
          {{ t('studio.write.optPin') }}
        </div>
        <div v-if="!republish && publishAt === null" class="st-ckrow" role="checkbox" tabindex="0" :aria-checked="announce" @click.prevent="announce = !announce" @keydown.enter.space.prevent="announce = !announce">
          <span class="st-ck" :class="{ on: announce }"><Icon :icon="Check" /></span>
          {{ t('studio.write.optAnnounce') }}
        </div>
      </div>

      <div class="st-flabel">{{ t('studio.write.checklist') }}</div>
      <ul class="checks">
        <li v-for="c in checks" :key="c.label" :class="{ ok: c.ok }">
          <SIcon :name="c.ok ? 'check' : 'minus'" :size="14" />{{ c.label }}<span class="sr-only"> · {{ t(c.ok ? 'studio.write.ckDone' : 'studio.write.ckPending') }}</span>
        </li>
      </ul>

      <div class="ft">
        <span class="kb"><kbd class="st-kbd">Ctrl</kbd><kbd class="st-kbd">Enter</kbd></span>
        <button type="button" class="st-btn g" @click="emit('close')">{{ t('studio.write.moreEdit') }}</button>
        <button type="button" class="st-btn p" :disabled="busy || !title.trim() || !scheduleOK" @click="emit('confirm')">
          <SIcon :name="publishAt !== null ? 'clock' : 'send'" :size="16" />{{ review ? t('studio.write.submitReview') : publishAt !== null ? t('schedule.later') : republish ? t('studio.write.update') : t('studio.write.publish') }}
        </button>
      </div>
    </div>
  </StModal>
</template>

<style scoped lang="scss">
:global(.st-modal.pub) {
  width: min(820px, calc(100vw - 32px));
  display: grid;
  grid-template-columns: 340px 1fr;
  padding: 0;
  max-height: calc(100dvh - 32px);
  overflow: auto;
}

.review-note {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  margin: 4px 0 6px;
  padding: 10px 12px;
  border-radius: var(--r-md);
  background: var(--well);
  font-size: 13px;
  color: var(--st-ink-2);
  line-height: 1.55;

  :deep(svg) { flex: none; margin-top: 2px; }
}

.pub-l {
  padding: 28px;
  background: var(--well);
  display: flex;
  flex-direction: column;

  .cap { font-size: 12px; color: var(--st-ink-3); letter-spacing: 0.12em; margin-bottom: 14px; }

  .url {
    margin-top: auto;
    padding-top: 20px;
    font: 12px var(--font-mono);
    color: var(--st-ink-3);
    display: flex;
    align-items: center;
    gap: 6px;
    word-break: break-all;
  }
}

:global(:root[data-mode='dark'] .st-modal.pub .pub-l) { background: color-mix(in oklab, var(--primary) 4%, #1c1c21); }

.pv-card {
  border-radius: var(--r-md);
  background: var(--paper);
  padding: 8px;
  box-shadow: var(--sh-card-hover);
  transform: rotate(-1.2deg);
  transition: transform var(--dur-slow) var(--ease-spring);

  &:hover { transform: rotate(0) translateY(-3px); }

  .pcv { aspect-ratio: 16 / 10; border-radius: var(--r-sm); }
  h4 { font: 600 17px/1.5 var(--font-serif); margin: 12px 6px 6px; }

  p {
    font-size: 12.5px;
    color: var(--st-ink-3);
    margin: 0 6px 10px;
    line-height: 1.65;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .meta { display: flex; gap: 10px; font-size: 12px; color: var(--st-ink-3); margin: 0 6px 6px; }
  .tag::before { content: '#'; color: var(--st-ink-4); margin-right: 2px; }
}

.pub-r {
  padding: 28px 28px 24px;
  display: flex;
  flex-direction: column;

  h3 { margin-bottom: 22px; }
  :deep(.publish-timing) { margin-bottom: 22px; }
}

.when {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 22px;
  font-size: 13.5px;
  color: var(--st-ink-2);
}

.opts {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 22px;

  .st-ckrow { font-size: 13.5px; }
}

.checks {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px 16px;

  li {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    color: var(--st-ink-3);

    .st-ic {
      width: 18px;
      height: 18px;
      padding: 2px;
      border-radius: 50%;
      background: var(--well-2);
    }

    &.ok { color: var(--st-ink-2); }

    &.ok .st-ic {
      background: color-mix(in oklab, var(--green) 16%, var(--paper));
      color: color-mix(in oklab, var(--green) 70%, var(--st-ink));
    }
  }
}

.ft {
  margin-top: auto;
  padding-top: 26px;

  .kb { margin-right: auto; display: flex; gap: 4px; align-items: center; }
}
</style>
