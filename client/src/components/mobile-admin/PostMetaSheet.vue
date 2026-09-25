<script setup lang="ts">
/** 文章元信息 sheet：封面（1–3 张，即时上传）/ slug / 标签 / 摘要 / 状态 / 置顶。 */
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, api, thumbOf, type PostDraft, type Tag } from '../../api';
import MaSheet from './MaSheet.vue';
import MaIcon from './MaIcon.vue';
import MaRing from './MaRing.vue';
import MaSegmented from './MaSegmented.vue';
import MaSwitch from './MaSwitch.vue';
import { shell, toast } from './state';
import { SLUG_RE, toSlug } from './format';

defineProps<{ open: boolean }>();
const emit = defineEmits<{ 'update:open': [v: boolean] }>();
const draft = defineModel<PostDraft>('draft', { required: true });
const { t } = useI18n();

const MAX_COVERS = 3;
const uploading = ref(0);
const tagInput = ref('');
const allTags = ref<Tag[]>([]);

onMounted(() => {
  api.tags().then((list) => (allTags.value = list)).catch(() => undefined);
});

const slugOk = computed(() => !draft.value.slug || SLUG_RE.test(draft.value.slug));
const suggestions = computed(() =>
  allTags.value
    .map((x) => x.name)
    .filter((n) => !draft.value.tags.includes(n) && (!tagInput.value || n.includes(tagInput.value.trim())))
    .slice(0, 8),
);

async function onCover(e: Event): Promise<void> {
  const input = e.target as HTMLInputElement;
  const room = MAX_COVERS - draft.value.covers.length;
  const files = Array.from(input.files ?? []).slice(0, Math.max(0, room));
  input.value = '';
  if (!files.length) return;
  uploading.value += files.length;
  try {
    const items = await adminApi.uploadMedia(files);
    draft.value.covers = [...draft.value.covers, ...items.map((i) => i.url)];
    shell.bump.media += 1;
  } catch {
    toast(t('mobileAdmin.common.uploadFailed'), '', 'error');
  } finally {
    uploading.value -= files.length;
  }
}

function removeCover(i: number): void {
  draft.value.covers = draft.value.covers.filter((_, k) => k !== i);
}

function makeFirst(i: number): void {
  if (!i) return;
  const list = [...draft.value.covers];
  const [c] = list.splice(i, 1);
  draft.value.covers = [c, ...list];
}

function addTag(name = tagInput.value): void {
  const v = name.trim().replace(/^#/, '');
  tagInput.value = '';
  if (!v || draft.value.tags.includes(v)) return;
  draft.value.tags = [...draft.value.tags, v];
}

function onTagKey(e: KeyboardEvent): void {
  if (e.key === 'Enter' || e.key === ',' || e.key === '，') {
    e.preventDefault();
    addTag();
  } else if (e.key === 'Backspace' && !tagInput.value && draft.value.tags.length) {
    draft.value.tags = draft.value.tags.slice(0, -1);
  }
}

function autoSlug(): void {
  draft.value.slug = toSlug(draft.value.title);
}

const statusIdx = computed({
  get: () => (draft.value.status === 'published' ? 1 : 0),
  set: (v: number) => (draft.value.status = v ? 'published' : 'draft'),
});
</script>

<template>
  <MaSheet
    :open="open"
    :detents="['half', 'full']"
    :label="t('mobileAdmin.post.meta')"
    @update:open="emit('update:open', $event)"
  >
    <template #head="{ close }">
      <div class="mh">
        <b>{{ t('mobileAdmin.post.meta') }}</b>
        <button class="done tap" @click="close">{{ t('mobileAdmin.common.done') }}</button>
      </div>
    </template>

    <div class="mb">
      <h5>{{ t('mobileAdmin.post.status') }}</h5>
      <MaSegmented
        v-model="statusIdx"
        class="seg"
        :items="[{ label: t('mobileAdmin.content.draft') }, { label: t('mobileAdmin.content.published') }]"
      />

      <h5>{{ t('mobileAdmin.post.covers') }}<small>{{ draft.covers.length }}/{{ MAX_COVERS }}</small></h5>
      <div class="covers">
        <div v-for="(c, i) in draft.covers" :key="c" class="cv" :class="{ first: i === 0 }" @click="makeFirst(i)">
          <img :src="thumbOf(c)" alt="" draggable="false" />
          <span v-if="i === 0" class="tag">{{ t('mobileAdmin.post.mainCover') }}</span>
          <button class="x" :aria-label="t('mobileAdmin.common.remove')" @click.stop="removeCover(i)"><MaIcon name="close" :size="11" /></button>
        </div>
        <div v-for="k in uploading" :key="`u${k}`" class="cv up"><MaRing indeterminate :size="28" /></div>
        <label v-if="draft.covers.length + uploading < MAX_COVERS" class="cv add tap">
          <MaIcon name="image" :size="22" />
          <small>{{ t('mobileAdmin.post.addCover') }}</small>
          <input type="file" accept="image/*" multiple hidden @change="onCover" />
        </label>
      </div>

      <h5>{{ t('mobileAdmin.post.slug') }}</h5>
      <div class="slug" :class="{ bad: !slugOk }">
        <span class="pre">/articles/</span>
        <input
          v-model="draft.slug"
          :placeholder="t('mobileAdmin.post.slugHint')"
          autocapitalize="off"
          autocomplete="off"
          spellcheck="false"
        />
        <button class="wand tap" :aria-label="t('mobileAdmin.post.autoSlug')" @click="autoSlug"><MaIcon name="wand" :size="18" /></button>
      </div>
      <p v-if="!slugOk" class="err">{{ t('mobileAdmin.post.slugInvalid') }}</p>

      <h5>{{ t('mobileAdmin.post.tags') }}</h5>
      <div class="tags ma-field">
        <button v-for="tg in draft.tags" :key="tg" class="tg tap" @click="draft.tags = draft.tags.filter((x) => x !== tg)">
          <span class="hs">#</span>{{ tg }}<MaIcon name="close" :size="11" />
        </button>
        <input
          v-model="tagInput"
          :placeholder="draft.tags.length ? '' : t('mobileAdmin.post.tagHint')"
          enterkeyhint="done"
          @keydown="onTagKey"
          @blur="addTag()"
        />
      </div>
      <div v-if="suggestions.length" class="sugg">
        <button v-for="s in suggestions" :key="s" class="chip tap" @mousedown.prevent @click="addTag(s)"><span class="hs">#</span>{{ s }}</button>
      </div>

      <h5>{{ t('mobileAdmin.post.excerpt') }}</h5>
      <textarea v-model="draft.excerpt" class="ma-field ex" rows="3" :placeholder="t('mobileAdmin.post.excerptHint')" />

      <div class="ma-list opts">
        <div class="ma-li">
          <span class="lic" style="--c: #ff9500"><MaIcon name="pin" :size="17" /></span>
          <span class="lt">{{ t('mobileAdmin.post.pin') }}<small class="d">{{ t('mobileAdmin.post.pinHint') }}</small></span>
          <MaSwitch v-model="draft.pinned" :label="t('mobileAdmin.post.pin')" />
        </div>
      </div>
    </div>
  </MaSheet>
</template>

<style scoped lang="scss">
.mh {
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  padding: 0 16px 12px;

  b {
    font-size: 16.5px;
    font-weight: 600;
  }

  .done {
    position: absolute;
    right: 16px;
    top: -4px;
    font-size: 16px;
    font-weight: 600;
    color: var(--ink);
    padding: 6px 4px;
  }
}

.mb {
  padding: 0 18px;

  h5 {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    font-size: 12.5px;
    font-weight: 500;
    color: var(--text-3);
    letter-spacing: 0.06em;
    margin: 20px 2px 10px;

    &:first-child { margin-top: 4px; }

    small {
      font-family: var(--font-mono);
      letter-spacing: 0;
    }
  }

  .seg { margin: 0; }
}

.covers {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.cv {
  position: relative;
  aspect-ratio: 4 / 3;
  border-radius: var(--r-md);
  overflow: hidden;
  background: var(--fill);
  display: grid;
  place-items: center;
  animation: ma-cv-in 0.45s var(--ease-spring) backwards;

  img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  &.first { box-shadow: 0 0 0 2px var(--ink); }

  .tag {
    position: absolute;
    left: 6px;
    bottom: 6px;
    padding: 2px 7px;
    border-radius: 999px;
    font-size: 10.5px;
    color: #fff;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(8px);
  }

  .x {
    position: absolute;
    top: 5px;
    right: 5px;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background: rgba(0, 0, 0, 0.55);
    color: #fff;

    :deep(.ma-ic) { stroke-width: 2.4; }
  }

  &.up {
    --ring-bg: var(--fill-2);
    --ring-fg: var(--ink);
  }

  &.add {
    gap: 2px;
    align-content: center;
    color: var(--text-3);
    box-shadow: inset 0 0 0 1px var(--line-2);
    cursor: pointer;

    small { font-size: 11.5px; }
  }
}

@keyframes ma-cv-in {
  from { opacity: 0; transform: scale(0.8); }
}

.slug {
  display: flex;
  align-items: center;
  height: 46px;
  padding: 0 6px 0 14px;
  border-radius: var(--r-md);
  background: var(--fill);
  box-shadow: inset 0 0 0 0.5px var(--line);

  &:focus-within { box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--ink) 55%, transparent); }
  &.bad { box-shadow: inset 0 0 0 1px var(--accent-red); }

  .pre {
    font-family: var(--font-mono);
    font-size: 13px;
    color: var(--text-3);
  }

  input {
    flex: 1;
    min-width: 0;
    font-family: var(--font-mono);
    font-size: 15px;
    caret-color: var(--ink);

    &::placeholder { color: var(--text-3); }
  }

  .wand {
    width: 36px;
    height: 36px;
    border-radius: var(--r-sm);
    display: grid;
    place-items: center;
    color: var(--ink);
  }
}

.err {
  margin: 6px 4px 0;
  font-size: 12px;
  color: var(--accent-red);
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 8px 10px;
  min-height: 46px;

  input {
    flex: 1;
    min-width: 100px;
    font-size: 15px;
    height: 30px;
    padding: 0 4px;

    &::placeholder { color: var(--text-3); }
  }
}

.tg {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 30px;
  padding: 0 10px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text);
  background: var(--elev);
  box-shadow: inset 0 0 0 1px var(--line-2);
  animation: ma-cv-in 0.35s var(--ease-spring);
}

.hs { color: var(--text-3); margin-right: -3px; }

.sugg {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin-top: 10px;

  .chip {
    height: 30px;
    font-size: 13px;
  }
}

.ex {
  resize: none;
  line-height: 1.6;
  font-size: 15px;
}

.opts {
  margin-top: 22px;
  background: var(--fill);

  .lt {
    display: flex;
    flex-direction: column;
    padding: 10px 0;
  }

  .d { font-size: 12px; }
}
</style>
