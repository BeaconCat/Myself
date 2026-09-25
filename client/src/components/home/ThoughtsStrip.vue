<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { thumbOf, type Note } from '../../api';
import SectionHead from './SectionHead.vue';
import { monthDay, plainText } from './format';

/** 首页「随想」横向预览：卡片横滑（吸附），左右按钮翻页，末尾一张「查看全部」 */
defineProps<{ notes: Note[]; total: number; loading: boolean }>();

const { t } = useI18n();
const router = useRouter();
const scroller = ref<HTMLElement | null>(null);

function md(s: string): string {
  return t('home.md', monthDay(s));
}

function page(dir: -1 | 1): void {
  const el = scroller.value;
  if (!el) return;
  const card = el.firstElementChild as HTMLElement | null;
  const step = (card?.offsetWidth ?? 340) + 20;
  el.scrollBy({ left: dir * step * Math.max(1, Math.floor(el.clientWidth / step) - 1), behavior: 'smooth' });
}

function open(): void {
  void router.push('/thoughts');
}
</script>

<template>
  <section class="thoughts-sec">
    <SectionHead v-reveal :title="t('home.thoughts')" :sub="t('home.thoughtsSub')" to="/thoughts" :link-text="t('home.allThoughts')">
      <button class="ib" :aria-label="t('home.scrollLeft')" @click="page(-1)">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 5l-7 7 7 7" /></svg>
      </button>
      <button class="ib" :aria-label="t('home.scrollRight')" @click="page(1)">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 5l7 7-7 7" /></svg>
      </button>
    </SectionHead>

    <div v-if="loading" class="hscroll" aria-hidden="true">
      <div v-for="n in 4" :key="n" class="tcard sk-card">
        <span class="sk sk-line" style="width: 64px" />
        <span class="sk sk-line" style="width: 92%; margin-top: 20px" />
        <span class="sk sk-line" style="width: 84%; margin-top: 10px" />
        <span class="sk sk-line" style="width: 60%; margin-top: 10px" />
        <div class="ims"><span class="sk" /><span class="sk" /></div>
      </div>
    </div>

    <div v-else ref="scroller" v-reveal class="hscroll rise-stagger">
      <div
        v-for="n in notes"
        :key="n.id"
        class="tcard"
        role="link"
        tabindex="0"
        @click="open"
        @keydown.enter="open"
      >
        <div class="top">
          <span>{{ md(n.createdAt) }}</span>
          <span v-if="n.mood" class="tag">{{ n.mood }}</span>
        </div>
        <p>{{ plainText(n.contentMd) }}</p>
        <div v-if="n.images.length" class="ims">
          <img v-for="src in n.images.slice(0, 3)" :key="src" :src="thumbOf(src)" alt="" loading="lazy" draggable="false" />
          <span v-if="n.images.length > 3" class="more">+{{ n.images.length - 3 }}</span>
        </div>
        <div v-else class="qt">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M5 5h14v10H9l-4 4z" /></svg>
          {{ t('home.plainNote') }}
        </div>
      </div>

      <router-link to="/thoughts" class="tcard all">
        <span class="go">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14M13 6l6 6-6 6" /></svg>
        </span>
        <span class="all-t">{{ t('home.moreThoughts', { n: total }) }}</span>
      </router-link>
    </div>
  </section>
</template>

<style scoped lang="scss">
/* 横滑出血到视口两侧，吸附起点与内容栏左缘对齐 */
.hscroll {
  --bleed: max(24px, calc((100vw - 1200px) / 2));

  display: grid;
  grid-auto-flow: column;
  grid-auto-columns: 340px;
  gap: 20px;
  overflow-x: auto;
  scroll-snap-type: x mandatory;
  scroll-padding-inline: var(--bleed);
  scrollbar-width: none;
  margin-inline: calc(50% - 50vw);
  padding: 6px var(--bleed) 30px;

  &::-webkit-scrollbar { display: none; }
}

.tcard {
  scroll-snap-align: start;
  display: flex;
  flex-direction: column;
  min-height: 248px;
  padding: 22px 22px 18px;
  border-radius: var(--r-lg);
  background: var(--elev);
  box-shadow: var(--shadow-card);
  color: var(--text);
  cursor: pointer;
  transition: transform var(--dur) var(--ease-out), box-shadow var(--dur) var(--ease-out);

  &:hover {
    transform: translateY(-3px);
    box-shadow: var(--shadow-card-hover);
  }

  &:focus-visible { outline: none; box-shadow: var(--shadow-card), var(--focus); }

  .top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 12.5px;
    color: var(--text-3);
  }

  p {
    margin-top: 14px;
    font-size: 15px;
    line-height: 1.8;
    display: -webkit-box;
    -webkit-line-clamp: 4;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
}

:root[data-mode='dark'] .tcard:not(.all) {
  background: linear-gradient(180deg, color-mix(in oklab, var(--surface) 90%, white), var(--surface) 70%);
}

.tag {
  color: var(--text-2);

  &::before {
    content: '#';
    margin-right: 2px;
    color: var(--text-3);
    font-family: var(--font-mono);
    font-size: 0.92em;
  }
}

.ims {
  display: flex;
  gap: 4px;
  margin-top: auto;
  padding-top: 16px;

  img,
  .more,
  .sk {
    width: 56px;
    height: 56px;
    border-radius: var(--r-sm);
    flex: none;
    object-fit: cover;
    background: var(--fill-2);
  }

  .more {
    display: grid;
    place-items: center;
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-2);
  }
}

.qt {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: auto;
  padding-top: 16px;
  font-size: 12.5px;
  color: var(--text-3);

  svg { width: 14px; height: 14px; }
}

/* 末尾「查看全部」：凹陷面，不抬升 */
.tcard.all {
  justify-content: center;
  align-items: center;
  gap: 12px;
  background: var(--fill);
  box-shadow: inset 0 0 0 0.5px var(--line);

  &:hover { background: var(--fill-2); box-shadow: inset 0 0 0 0.5px var(--line-2); }
  &:hover .go svg { transform: translateX(3px); }
}

.go {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: var(--text-2);
  background: var(--fill-2);

  svg {
    width: 18px;
    height: 18px;
    transition: transform var(--dur) var(--ease-spring);
  }
}

.all-t {
  font-size: 14px;
  color: var(--text-2);
}

.sk-card {
  cursor: default;

  .sk-line {
    display: block;
    height: 12px;
  }

  &:hover { transform: none; box-shadow: var(--shadow-card); }
}

.ib {
  width: 32px;
  height: 32px;
  border: 0;
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: var(--text-2);
  background: var(--fill);
  transition: background-color var(--dur-fast), color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  svg { width: 16px; height: 16px; }

  &:hover { color: var(--text); background: var(--fill-2); }
  &:active { transform: scale(0.94); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

@media (max-width: 768px) {
  .hscroll { grid-auto-columns: min(300px, 78vw); }
  .ib { display: none; }
}
</style>
