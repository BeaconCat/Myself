<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import type { ListeningData } from '../types';
import type { ModProps } from './props';
import KitIcon from '../parts/KitIcon.vue';
import ModHead from '../parts/ModHead.vue';
import Scene from '../parts/Scene.vue';
import { useClock } from '../useClock';

/** 最近在听（listening）：黑胶唱片播放时旋转、唱臂落下，进度逐秒走；宽版右侧列出最近曲目 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as ListeningData);
const { t } = useI18n();
const now = useClock();

const playing = ref(d.value.playing);
const pos = ref(Number(d.value.now.position ?? 0));
const dur = computed(() => Math.max(1, Number(d.value.now.duration) || 240));

watch(() => d.value.now.position, (v) => { pos.value = Number(v ?? 0); });
watch(now, () => {
  if (!playing.value) return;
  pos.value = pos.value + 1 > dur.value ? 0 : pos.value + 1;
});

const fmt = (s: number) => `${Math.floor(s / 60)}:${String(Math.floor(s % 60)).padStart(2, '0')}`;
const skip = (delta: number) => { pos.value = Math.max(0, Math.min(dur.value, pos.value + delta)); };
</script>

<template>
  <ModHead :title="title"><span class="ak-dot" :class="{ live: playing }" />{{ playing ? 'LIVE' : 'PAUSED' }}</ModHead>
  <div class="ls" :class="{ playing, wide: d.recent.length > 0 }">
    <div class="ls-main">
      <div class="ls-deck">
        <div class="deck-vinyl">
          <div class="vinyl"><div class="lbl" /></div>
          <span class="arm" />
        </div>
        <div class="ls-info">
          <b>{{ d.now.title }}</b>
          <span>{{ d.now.artist }}</span>
          <small v-if="d.now.album">{{ d.now.album }}</small>
        </div>
      </div>
      <div class="ls-bar">
        <span>{{ fmt(pos) }}</span>
        <span class="tr"><i :style="{ '--w': `${(pos / dur) * 100}%` }" /></span>
        <span>{{ fmt(dur) }}</span>
      </div>
      <div class="ls-ctl">
        <button :aria-label="t('aboutKit.listening.prev')" @click="skip(-15)"><KitIcon name="prev" :size="18" /></button>
        <button class="pp" :aria-label="t('aboutKit.listening.toggle')" @click="playing = !playing"><KitIcon :name="playing ? 'pause' : 'play'" :size="18" /></button>
        <button :aria-label="t('aboutKit.listening.next')" @click="skip(15)"><KitIcon name="next" :size="18" /></button>
      </div>
    </div>
    <ul v-if="d.recent.length" class="ls-recent">
      <li v-for="(r, i) in d.recent" :key="i">
        <span class="cv"><Scene :src="r.cover" :scene="r.scene || '01'" small /></span>
        <span class="nm">{{ r.title }}<small>{{ r.artist }}</small></span>
        <time>{{ r.at }}</time>
      </li>
    </ul>
  </div>
</template>

<style scoped lang="scss">
.ls { display: grid; grid-template-columns: 1fr; gap: 18px; flex: 1; align-content: space-between; }

.ls-deck { display: grid; grid-template-columns: auto minmax(0, 1fr); gap: 18px; align-items: center; }

.deck-vinyl { --s: 104px; position: relative; }

.vinyl {
  position: relative;
  width: var(--s);
  height: var(--s);
  border-radius: 50%;
  background: radial-gradient(circle, #0000 0 17%, #111 17.5%), repeating-radial-gradient(circle, #161616 0 1px, #0c0c0c 1.5px 3px);
  box-shadow: 0 12px 30px -10px rgb(0 0 0 / 0.7), inset 0 0 0 1px rgb(255 255 255 / 0.05);
  animation: ls-spin 3.2s linear infinite paused;

  &::before {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: 50%;
    background: conic-gradient(from 30deg, transparent 0 10%, rgb(255 255 255 / 0.12) 14%, transparent 20% 55%, rgb(255 255 255 / 0.08) 60%, transparent 66%);
  }

  .lbl {
    position: absolute;
    inset: 33%;
    border-radius: 50%;
    background: radial-gradient(circle, var(--bg) 0 8%, transparent 9%), var(--primary);

    &::after { content: ''; position: absolute; inset: 18%; border-radius: 50%; border: 1px solid rgb(255 255 255 / 0.25); }
  }
}

.playing .vinyl { animation-play-state: running; }

@keyframes ls-spin { to { transform: rotate(360deg); } }

.arm {
  position: absolute;
  left: calc(var(--s) - 26px);
  top: -6px;
  z-index: 2;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--ak-text-3);
  box-shadow: 0 0 0 4px var(--ak-sunken);

  &::after {
    content: '';
    position: absolute;
    left: 4px;
    top: 4px;
    width: 3px;
    height: calc(var(--s) * 0.58);
    border-radius: var(--r-pill);
    background: linear-gradient(var(--text-2), var(--ak-text-3));
    transform-origin: top center;
    transform: rotate(8deg);
    transition: transform 0.8s var(--ease-spring);
  }
}

.playing .arm::after { transform: rotate(26deg); }

.ls-info {
  min-width: 0;

  b { display: block; font: 700 19px/1.35 var(--font-serif); }
  span { font-size: 15px; color: var(--text-2); }
  small { display: block; margin-top: 4px; font: 400 12.5px var(--ak-mono); color: var(--ak-text-3); }
}

.ls-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 16px;
  font: 400 12.5px var(--ak-mono);
  color: var(--ak-text-3);

  .tr { position: relative; flex: 1; height: 4px; border-radius: var(--r-pill); background: var(--fill-2); }

  i {
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    width: var(--w, 30%);
    border-radius: inherit;
    background: var(--ink);
    transition: width 1s linear;

    &::after {
      content: '';
      position: absolute;
      right: -4px;
      top: 50%;
      width: 9px;
      height: 9px;
      margin-top: -4.5px;
      border-radius: 50%;
      background: var(--text);
      box-shadow: 0 0 0 3px var(--fill-3);
    }
  }
}

.ls-ctl {
  display: flex;
  gap: 6px;
  margin-top: 12px;

  button {
    display: grid;
    place-items: center;
    width: 38px;
    height: 38px;
    border-radius: var(--r-pill);
    color: var(--text-2);
    background: var(--fill);
    transition: background-color var(--dur-fast), color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

    &:hover { background: var(--fill-2); color: var(--text); }
    &:active { transform: scale(0.94); }
  }

  /* 播放 / 暂停 = 主按钮实底 */
  .pp { background: var(--solid); color: var(--on-solid); box-shadow: var(--btn-shadow); }
  .pp:hover { background: var(--solid-hover); color: var(--on-solid); }
}

.ls-recent {
  list-style: none;
  border-top: 1px solid var(--ak-line);

  li {
    display: grid;
    grid-template-columns: 32px minmax(0, 1fr) auto;
    gap: 12px;
    align-items: center;
    padding: 10px 0;
    border-bottom: 1px solid var(--ak-line);
    font-size: 14px;
    font-weight: 500;

    &:last-child { border-bottom: 0; }
  }

  .cv { display: block; width: 32px; height: 32px; overflow: hidden; border-radius: var(--r-xs); }
  .nm { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .nm small { margin-left: 6px; font-size: 13px; font-weight: 400; color: var(--ak-text-3); }
  time { font: 400 12.5px var(--ak-mono); color: var(--ak-text-3); }
}

@container (min-width: 600px) {
  .ls.wide { grid-template-columns: auto minmax(0, 1fr); gap: 34px; align-items: center; }
  .ls-recent { border-top: 0; border-left: 1px solid var(--ak-line); padding-left: 28px; }
  .deck-vinyl { --s: 150px; }
}

@container (max-width: 260px) { .ls-deck { grid-template-columns: 1fr; justify-items: start; } }
</style>
