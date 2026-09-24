<script setup lang="ts">
/* eslint-disable vue/no-mutating-props */
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { ProfileData } from '../../../about/types';
import { SOCIAL_ICONS } from '../../../about/icons';
import CoverUploader from '../CoverUploader.vue';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 身份区：名字 / 一句话 / 自述 / 状态行 / 社交入口 / 形象图（图片、渐隐方向、圆角、焦点）+ 实时预览 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<ProfileData>(() => props.mod);
const { t } = useI18n();

const portraitList = computed<string[]>({
  get: () => (d.value.portrait.src ? [d.value.portrait.src] : []),
  set: (v) => { d.value.portrait.src = v[0] ?? ''; },
});

/** focus = 'X% Y%' 拆成两个滑杆 */
const focus = computed(() => {
  const [x, y] = (d.value.portrait.focus || '50% 40%').split(/\s+/).map((s) => parseFloat(s));
  return { x: Number.isFinite(x) ? x : 50, y: Number.isFinite(y) ? y : 40 };
});
function setFocus(axis: 'x' | 'y', v: number): void {
  const f = { ...focus.value, [axis]: v };
  d.value.portrait.focus = `${f.x}% ${f.y}%`;
}

const FADES = ['left', 'bottom', 'none'] as const;
const previewMask = computed(() => ({
  left: 'linear-gradient(to right, transparent, rgb(0 0 0 / 0.3) 17%, rgb(0 0 0 / 0.88) 35%, #000 42%)',
  bottom: 'linear-gradient(to top, transparent, rgb(0 0 0 / 0.3) 17%, rgb(0 0 0 / 0.88) 35%, #000 42%)',
  none: 'none',
})[d.value.portrait.fade] ?? 'none');
</script>

<template>
  <div class="ed">
    <div class="grid3">
      <label><span>{{ t('aboutKit.ed.hello') }}</span><input v-model="d.hello" class="a-input" type="text" /></label>
      <label class="span2"><span>{{ t('aboutKit.ed.name') }}</span><input v-model="d.name" class="a-input" type="text" /></label>
    </div>
    <label><span>{{ t('aboutKit.ed.lede') }}</span><input v-model="d.lede" class="a-input" type="text" :placeholder="t('aboutKit.ed.ledeHint')" /></label>
    <label><span>{{ t('aboutKit.ed.bio') }}</span><textarea v-model="d.bio" class="a-input" rows="3" /></label>
    <label><span>{{ t('aboutKit.ed.kicker') }}</span><input v-model="d.kicker" class="a-input" type="text" :placeholder="t('aboutKit.ed.kickerHint')" /></label>

    <div class="sub-title">{{ t('aboutKit.ed.statusRow') }}</div>
    <div class="grid3">
      <label><span>{{ t('aboutKit.doing') }}</span><input v-model="d.status.doing" class="a-input" type="text" /></label>
      <label><span>{{ t('aboutKit.city') }}</span><input v-model="d.status.city" class="a-input" type="text" /></label>
      <label><span>{{ t('aboutKit.ed.tz') }}</span><input v-model.number="d.status.tz" class="a-input" type="number" min="-12" max="14" step="0.5" /></label>
    </div>

    <div class="sub-title">{{ t('aboutKit.ed.portrait') }}</div>
    <div class="pt">
      <div class="pt-form">
        <CoverUploader v-model="portraitList" :max="1" />
        <p class="hint">{{ t('aboutKit.ed.portraitHint') }}</p>
        <div class="fld">
          <span>{{ t('aboutKit.ed.fade') }}</span>
          <div class="seg">
            <button v-for="f in FADES" :key="f" type="button" :class="{ on: d.portrait.fade === f }" @click="d.portrait.fade = f">{{ t(`aboutKit.ed.fade_${f}`) }}</button>
          </div>
        </div>
        <label>
          <span>{{ t('aboutKit.ed.radius') }} · {{ d.portrait.radius }}px</span>
          <input v-model.number="d.portrait.radius" class="range" type="range" min="0" max="80" />
        </label>
        <div class="grid2">
          <label><span>{{ t('aboutKit.ed.focusX') }} · {{ focus.x }}%</span><input class="range" type="range" min="0" max="100" :value="focus.x" @input="setFocus('x', Number(($event.target as HTMLInputElement).value))" /></label>
          <label><span>{{ t('aboutKit.ed.focusY') }} · {{ focus.y }}%</span><input class="range" type="range" min="0" max="100" :value="focus.y" @input="setFocus('y', Number(($event.target as HTMLInputElement).value))" /></label>
        </div>
      </div>
      <figure class="pt-preview" :class="{ logo: !d.portrait.src }">
        <span
          class="pt-frame"
          :style="{ borderRadius: `${d.portrait.radius}px`, maskImage: previewMask, WebkitMaskImage: previewMask }"
        >
          <img :src="d.portrait.src || '/favicon-256.png'" alt="" :style="{ objectPosition: d.portrait.focus || '50% 40%' }" />
        </span>
        <figcaption>{{ t('aboutKit.ed.preview') }}</figcaption>
      </figure>
    </div>

    <div class="sub-title">{{ t('aboutKit.ed.links') }}</div>
    <EdList v-slot="{ item }" :items="d.links" :make="() => ({ name: '', handle: '', url: '', icon: 'link' })">
      <div class="grid4">
        <input v-model="item.name" class="a-input" type="text" :placeholder="t('aboutKit.ed.linkName')" />
        <input v-model="item.handle" class="a-input" type="text" :placeholder="t('aboutKit.ed.handle')" />
        <input v-model="item.url" class="a-input" type="text" placeholder="https://…" />
        <div class="line">
          <select v-model="item.icon" class="a-input flex-in">
            <option v-for="ic in SOCIAL_ICONS" :key="ic" :value="ic">{{ ic }}</option>
          </select>
          <label class="check"><input v-model="item.primary" type="checkbox" />{{ t('aboutKit.ed.primary') }}</label>
        </div>
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.pt { display: grid; grid-template-columns: minmax(0, 1fr) 200px; gap: 18px; align-items: start; }
.pt-form { display: flex; flex-direction: column; gap: 12px; min-width: 0; }

.pt-preview {
  margin: 0;
  padding: 14px;
  border-radius: 14px;
  background: var(--bg);
  box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--text) 8%, transparent);

  figcaption { margin-top: 8px; font-size: 11.5px; text-align: center; color: var(--text-2); }
}

.pt-frame {
  display: block;
  position: relative;
  aspect-ratio: 440 / 540;
  overflow: hidden;
  transition: border-radius var(--dur);

  img { width: 100%; height: 100%; display: block; object-fit: cover; }
}

.pt-preview.logo .pt-frame {
  background: radial-gradient(60% 50% at 58% 62%, rgba(var(--primary-rgb), 0.28), transparent 70%), radial-gradient(120% 90% at 60% 40%, #0b1528, #050b17 70%);

  img { position: absolute; inset: 0 0 0 18%; margin: auto; width: 62%; height: auto; aspect-ratio: 1; object-fit: contain; }
}

@media (max-width: 720px) { .pt { grid-template-columns: 1fr; } .pt-preview { max-width: 200px; } }
</style>
