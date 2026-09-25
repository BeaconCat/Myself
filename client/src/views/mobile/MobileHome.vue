<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { api, thumbOf, type Note, type Post } from '../../api';
import { useConfigStore } from '../../stores/config';
import LargeTitlePage from '../../components/mobile/LargeTitlePage.vue';
import CoverCarousel from '../../components/mobile/CoverCarousel.vue';
import CoverArt from '../../components/common/CoverArt.vue';
import MIcon from '../../components/mobile/MIcon.vue';
import { copyText, monthDay, shell, toast } from '../../components/mobile/shell';

/** 移动端首页：封面卡横滑轮播 + 最新文章 + 随想横滑预览 + 页脚订阅。 */
const { t } = useI18n();
const router = useRouter();
const config = useConfigStore();

const hero = ref<Post[]>([]);
const latest = ref<Post[]>([]);
const notes = ref<Note[]>([]);
const loading = ref(true);
const failed = ref(false);

const eyebrow = computed(() => {
  const now = new Date();
  const week = new Intl.DateTimeFormat('zh-CN', { weekday: 'long', timeZone: config.cfg.timezone }).format(now);
  return `${week} · ${now.getMonth() + 1} 月 ${now.getDate()} 日`;
});

async function load(): Promise<void> {
  failed.value = false;
  try {
    const [feed, list, feedNotes] = await Promise.all([
      api.hero(),
      api.posts({ pageSize: 4 }),
      api.notes({ pageSize: 6 }),
    ]);
    hero.value = feed.items.length ? feed.items : list.items;
    latest.value = list.items;
    notes.value = feedNotes.items;
  } catch {
    failed.value = true;
  } finally {
    loading.value = false;
  }
}

onMounted(load);

function plain(md: string): string {
  return md.replace(/!\[[^\]]*]\([^)]*\)/g, '').replace(/[#>*_`~[\]]/g, '').replace(/\(([^)]*)\)/g, '').trim();
}

function md(s: string): string {
  const { m, d } = monthDay(s);
  return `${m}月${d}日`;
}

function openPost(slug: string): void {
  void router.push(`/articles/${slug}`);
}

async function copyRss(): Promise<void> {
  const ok = await copyText(`${window.location.origin}/feed`);
  toast(ok ? t('mobile.rssCopied') : t('mobile.copyFailed'), ok ? '/feed' : '');
}

const carouselPaused = computed(() => shell.dp > 0.1 || shell.pushed || shell.tab !== 'home' || shell.searchOpen);
</script>

<template>
  <LargeTitlePage :title="config.cfg.site.title" :eyebrow="eyebrow" :refresh="load">
    <template #right>
      <button class="m-icbtn m-tap" :aria-label="t('mobile.rss')" @click="copyRss"><MIcon name="rss" /></button>
    </template>

    <Transition name="m-swap" mode="out-in">
      <!-- 骨架：与最终布局同几何 -->
      <div v-if="loading" key="sk" class="sk">
        <div class="m-sk sk-card" />
        <div class="sk-dots"><span class="m-sk" /><span class="m-sk" /><span class="m-sk" /></div>
        <div class="sec-h"><span class="m-sk m-sk-line" style="width: 96px; height: 20px" /></div>
        <div v-for="n in 3" :key="n" class="sk-row">
          <div class="rt">
            <span class="m-sk m-sk-line" style="width: 40%; height: 10px" />
            <span class="m-sk m-sk-line" style="width: 82%; height: 16px" />
            <span class="m-sk m-sk-line" style="width: 64%" />
          </div>
          <span class="m-sk thumb" />
        </div>
      </div>

      <div v-else key="ok">
        <div v-if="failed && !hero.length" class="m-empty">{{ t('mobile.loadFailed') }}</div>
        <CoverCarousel v-if="hero.length" class="m-in" :posts="hero" :paused="carouselPaused" @open="openPost" />

        <section v-if="latest.length" class="sec">
          <div class="sec-h m-in" style="--i: 2">
            <h2>{{ t('mobile.home.latest') }}</h2>
            <button class="more m-tap" @click="router.push('/articles')">{{ t('mobile.home.all') }}<MIcon name="chev" class="xs" /></button>
          </div>
          <div class="rows">
            <button
              v-for="(p, i) in latest"
              :key="p.id"
              class="m-arow m-in"
              :style="{ '--i': i + 3 }"
              @click="openPost(p.slug)"
            >
              <div class="rt">
                <div class="meta"><em>{{ p.tags[0] }}</em> · {{ md(p.createdAt) }}</div>
                <b>{{ p.title }}</b>
                <span class="ex">{{ p.excerpt }}</span>
              </div>
              <div class="thumb"><CoverArt :src="p.covers[0]" :seed="p.slug" thumb /></div>
            </button>
          </div>
        </section>

        <section v-if="notes.length" class="sec">
          <div class="sec-h m-in" style="--i: 7">
            <h2>{{ t('mobile.home.thoughts') }}</h2>
            <button class="more m-tap" @click="router.push('/thoughts')">{{ t('mobile.home.all') }}<MIcon name="chev" class="xs" /></button>
          </div>
          <div class="tp m-hscroll">
            <button
              v-for="(n, i) in notes"
              :key="n.id"
              class="tcard m-tap m-in"
              :style="{ '--i': i + 8 }"
              @click="router.push('/thoughts')"
            >
              <p>{{ plain(n.contentMd) }}</p>
              <div v-if="n.images.length" class="strip">
                <img v-for="src in n.images.slice(0, 3)" :key="src" :src="thumbOf(src)" alt="" loading="lazy" draggable="false" />
                <span v-if="n.images.length > 3" class="more">+{{ n.images.length - 3 }}</span>
              </div>
              <footer><em v-if="n.mood" class="m-mood">{{ n.mood }}</em><span>{{ md(n.createdAt) }}</span></footer>
            </button>
          </div>
        </section>

        <footer class="foot m-in" style="--i: 12">
          <img class="mk" src="/favicon-256.png" alt="" draggable="false" />
          <div>
            <b>{{ config.cfg.site.title }}</b>
            <small>{{ config.cfg.about.motto || config.cfg.site.subtitle }}</small>
          </div>
          <button class="rss m-tap" @click="copyRss"><MIcon name="rss" class="xs" />{{ t('mobile.subscribe') }}</button>
        </footer>
      </div>
    </Transition>
  </LargeTitlePage>
</template>

<style scoped lang="scss">
.sec { margin-top: 34px; }

.sec-h {
  display: flex;
  align-items: baseline;
  padding: 0 20px;
  margin-bottom: 6px;

  h2 {
    font-family: var(--font-serif);
    font-size: 22px;
    font-weight: 700;
  }

  .more {
    margin-left: auto;
    display: flex;
    align-items: center;
    gap: 2px;
    font-size: 14px;
    color: var(--ink);
  }
}

.rows { padding: 0 4px; }

.tp {
  display: flex;
  gap: 12px;
  padding: 8px 20px 18px;
  scroll-snap-type: x mandatory;
  scroll-padding: 0 20px;
}

.tcard {
  scroll-snap-align: start;
  flex: none;
  width: min(268px, 72vw);
  padding: 16px;
  border-radius: var(--r-xl);
  text-align: left;
  background: var(--elev);
  box-shadow: var(--shadow-card);
  display: flex;
  flex-direction: column;
  gap: 12px;

  p {
    font-size: 14.5px;
    line-height: 1.7;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
    min-height: 73px;
  }

  .strip {
    display: flex;
    gap: 4px;

    img,
    .more {
      width: 52px;
      height: 52px;
      border-radius: var(--r-sm);
      object-fit: cover;
      background: var(--fill-3);
    }

    .more {
      display: grid;
      place-items: center;
      font-size: 13px;
      color: var(--text-2);
      font-family: var(--m-font-mono);
    }
  }

  footer {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: var(--text-3);
    margin-top: auto;
  }
}

.foot {
  margin: 34px 20px 0;
  padding-top: 26px;
  border-top: 0.5px solid var(--line);
  display: flex;
  align-items: center;
  gap: 12px;

  .mk {
    width: 34px;
    height: 34px;
    border-radius: var(--r-sm);
    flex: none;
  }

  b {
    font-family: var(--font-serif);
    font-size: 15px;
    display: block;
  }

  small {
    font-size: 12px;
    color: var(--text-3);
  }

  .rss {
    margin-left: auto;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 12px;
    border-radius: 999px;
    background: var(--fill-2);
    font-size: 13px;
    color: var(--text-2);
    flex: none;
  }
}

/* 骨架 */
.sk-card {
  margin: 4px 20px 0;
  height: min(calc((100vw - 40px) * 1.3), 62vh);
  border-radius: var(--r-xl);
}

.sk-dots {
  display: flex;
  justify-content: center;
  gap: 6px;
  margin: 16px 0 34px;

  span { width: 6px; height: 6px; border-radius: 999px; }
  span:first-child { width: 30px; }
}

.sk-row {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 20px;

  .rt {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .thumb {
    width: 68px;
    height: 68px;
    border-radius: var(--r-lg);
    flex: none;
  }
}

.sk .sec-h { margin-bottom: 8px; }
</style>
