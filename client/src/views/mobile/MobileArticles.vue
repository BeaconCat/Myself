<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { api, type Post, type Tag } from '../../api';
import LargeTitlePage from '../../components/mobile/LargeTitlePage.vue';
import CoverArt from '../../components/common/CoverArt.vue';
import MIcon from '../../components/mobile/MIcon.vue';
import { monthDay, openSearch } from '../../components/mobile/shell';

/**
 * 移动端文章列表：大标题折叠、标签 chips 吸顶（?tag= 驱动，可被抽屉/搜索深链）、
 * 首篇大封面 + 其余图文行、按月分组、触底续载、下拉刷新。
 */
const props = withDefaults(defineProps<{ tag?: string }>(), { tag: '' });
const { t } = useI18n();
const router = useRouter();

const PAGE_SIZE = 10;
const page = ref<InstanceType<typeof LargeTitlePage> | null>(null);
const posts = ref<Post[]>([]);
const tags = ref<Tag[]>([]);
const total = ref(0);
const allTotal = ref(0);
const latestAt = ref('');
const pageNo = ref(1);
const loading = ref(true);
const loadingMore = ref(false);
const failed = ref(false);
let seq = 0;

const hasMore = computed(() => posts.value.length < total.value);

async function reload(): Promise<void> {
  const my = ++seq;
  failed.value = false;
  try {
    const [list, tagList] = await Promise.all([
      api.posts({ page: 1, pageSize: PAGE_SIZE, tag: props.tag || undefined }),
      tags.value.length ? Promise.resolve(tags.value) : api.tags(),
    ]);
    if (my !== seq) return;
    posts.value = list.items;
    total.value = list.total;
    pageNo.value = 1;
    tags.value = tagList;
    if (!props.tag) {
      allTotal.value = list.total;
      latestAt.value = list.items[0]?.createdAt ?? '';
    } else if (!allTotal.value) {
      const all = await api.posts({ pageSize: 1 });
      allTotal.value = all.total;
      latestAt.value = all.items[0]?.createdAt ?? '';
    }
  } catch {
    if (my === seq) failed.value = true;
  } finally {
    if (my === seq) loading.value = false;
  }
}

async function loadMore(): Promise<void> {
  if (loading.value || loadingMore.value || !hasMore.value) return;
  loadingMore.value = true;
  const my = seq;
  try {
    const list = await api.posts({ page: pageNo.value + 1, pageSize: PAGE_SIZE, tag: props.tag || undefined });
    if (my !== seq) return;
    posts.value = [...posts.value, ...list.items];
    total.value = list.total;
    pageNo.value += 1;
  } finally {
    loadingMore.value = false;
  }
}

async function refresh(): Promise<void> {
  tags.value = [];
  await reload();
}

watch(
  () => props.tag,
  async () => {
    loading.value = true;
    /* 已滚过大标题则停在 chips 吸顶处，否则保持原位 */
    const sc = page.value?.scroller;
    if (sc && sc.scrollTop > 90) sc.scrollTo({ top: 90 });
    await reload();
  },
);

/* 触底哨兵：以页面滚动容器为 root */
const sentinel = ref<HTMLElement | null>(null);
let io: IntersectionObserver | null = null;
onMounted(() => {
  io = new IntersectionObserver((es) => { if (es.some((e) => e.isIntersecting)) void loadMore(); }, {
    root: page.value?.scroller ?? null,
    rootMargin: '400px',
  });
  if (sentinel.value) io.observe(sentinel.value);
  void reload();
});
watch(sentinel, (el, old) => {
  if (old) io?.unobserve(old);
  if (el) io?.observe(el);
});
onBeforeUnmount(() => io?.disconnect());

function selectTag(name: string): void {
  if (name === props.tag) return;
  void router.replace({ name: 'articles', query: name ? { tag: name } : {} });
}

const sub = computed(() => {
  if (!latestAt.value) return allTotal.value ? '' : t('mobile.articles.subEmpty');
  const { m, d } = monthDay(latestAt.value);
  return t('mobile.articles.sub', { n: allTotal.value, m, d });
});

/** 按月分组：首篇作为大封面，后续按月插入分组标题 */
interface Row { kind: 'month' | 'feat' | 'row'; key: string; post?: Post; label?: string; i: number }
const rows = computed<Row[]>(() => {
  const out: Row[] = [];
  let month = '';
  posts.value.forEach((p, i) => {
    const key = p.createdAt.slice(0, 7);
    if (key !== month) {
      month = key;
      const { y, m } = monthDay(p.createdAt);
      out.push({ kind: 'month', key: `m-${key}`, label: t('mobile.articles.month', { y, m }), i });
    }
    out.push({ kind: i === 0 ? 'feat' : 'row', key: `p-${p.id}`, post: p, i });
  });
  return out;
});

function md(s: string): string {
  const { m, d } = monthDay(s);
  return `${m}月${d}日`;
}

function open(p: Post): void {
  void router.push(`/articles/${p.slug}`);
}
</script>

<template>
  <LargeTitlePage ref="page" :title="t('mobile.articles.title')" :sub="sub" :refresh="refresh">
    <template #right>
      <button class="m-icbtn m-tap" :aria-label="t('mobile.tabs.search')" @click="openSearch()"><MIcon name="search" /></button>
    </template>

    <template #extra>
      <div class="chips m-hscroll">
        <button class="chip m-tap" :class="{ on: !tag }" @click="selectTag('')">
          {{ t('mobile.articles.all') }}<em>{{ allTotal || '' }}</em>
        </button>
        <button
          v-for="tg in tags"
          :key="tg.name"
          class="chip m-tap"
          :class="{ on: tg.name === tag }"
          @click="selectTag(tg.name)"
        >
          {{ tg.name }}<em>{{ tg.count }}</em>
        </button>
      </div>
    </template>

    <Transition name="m-swap" mode="out-in">
      <div v-if="loading" key="sk" class="sk">
        <span class="m-sk m-sk-line grp-sk" />
        <div class="m-sk feat-sk" />
        <span class="m-sk m-sk-line" style="width: 36%; height: 10px; margin: 14px 20px 0" />
        <span class="m-sk m-sk-line" style="width: 78%; height: 20px; margin: 10px 20px 0" />
        <span class="m-sk m-sk-line" style="width: 88%; margin: 10px 20px 22px" />
        <div v-for="n in 3" :key="n" class="sk-row">
          <div class="rt">
            <span class="m-sk m-sk-line" style="width: 40%; height: 10px" />
            <span class="m-sk m-sk-line" style="width: 80%; height: 16px" />
            <span class="m-sk m-sk-line" style="width: 92%" />
            <span class="m-sk m-sk-line" style="width: 60%" />
          </div>
          <span class="m-sk thumb" />
        </div>
      </div>

      <div v-else :key="`list-${tag}`" class="al">
        <div v-if="failed" class="m-empty">{{ t('mobile.loadFailed') }}</div>
        <div v-else-if="!posts.length" class="m-empty">{{ t('mobile.articles.empty') }}</div>
        <template v-for="r in rows" :key="r.key">
          <div v-if="r.kind === 'month'" class="grp m-in" :style="{ '--i': Math.min(r.i, 8) }">{{ r.label }}</div>
          <button
            v-else-if="r.kind === 'feat' && r.post"
            class="feat m-tap m-in"
            :style="{ '--i': 1 }"
            @click="open(r.post)"
          >
            <div class="fc"><CoverArt :src="r.post.covers[0]" :seed="r.post.slug" /></div>
            <div class="meta"><em>{{ r.post.tags[0] }}</em> · {{ md(r.post.createdAt) }}</div>
            <h3>{{ r.post.title }}</h3>
            <p>{{ r.post.excerpt }}</p>
          </button>
          <button
            v-else-if="r.post"
            class="m-arow m-in"
            :style="{ '--i': Math.min(r.i + 1, 9) }"
            @click="open(r.post)"
          >
            <div class="rt">
              <div class="meta"><em>{{ r.post.tags[0] }}</em> · {{ md(r.post.createdAt) }}</div>
              <b>{{ r.post.title }}</b>
              <span class="ex">{{ r.post.excerpt }}</span>
            </div>
            <div class="thumb"><CoverArt :src="r.post.covers[0]" :seed="r.post.slug" thumb /></div>
          </button>
        </template>
        <div ref="sentinel" class="sentinel" />
        <div v-if="loadingMore" class="sk-row more-sk">
          <div class="rt">
            <span class="m-sk m-sk-line" style="width: 40%; height: 10px" />
            <span class="m-sk m-sk-line" style="width: 80%; height: 16px" />
          </div>
          <span class="m-sk thumb" />
        </div>
      </div>
    </Transition>
  </LargeTitlePage>
</template>

<style scoped lang="scss">
.chips {
  display: flex;
  gap: 8px;
  padding: 0 20px;
}

.chip {
  flex: none;
  height: 36px;
  padding: 0 15px;
  border-radius: 999px;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: var(--text-2);
  background: var(--m-fill);
  box-shadow: inset 0 0 0 0.5px var(--m-line);
  transition:
    background var(--dur) var(--ease-out),
    color var(--dur),
    box-shadow var(--dur),
    transform var(--dur-fast) var(--ease-spring);

  em {
    font-style: normal;
    font-family: var(--m-font-mono);
    font-size: 11px;
    opacity: 0.6;
  }

  &.on {
    background: var(--text);
    color: var(--bg);
    box-shadow: none;
    font-weight: 500;
  }
}

.al { padding: 0 4px; }

.grp {
  padding: 14px 16px 8px;
  font-size: 12.5px;
  color: var(--m-text-3);
  letter-spacing: 0.06em;
}

.feat {
  display: block;
  width: calc(100% - 32px);
  margin: 4px 16px 18px;
  text-align: left;

  .fc {
    position: relative;
    height: 218px;
    border-radius: 24px;
    overflow: hidden;
    box-shadow: 0 22px 40px -24px rgba(0, 10, 40, 0.8), 0 0 0 0.5px var(--m-line);
  }

  .meta {
    margin-top: 14px;
    font-size: 12.5px;
    color: var(--m-text-3);

    em {
      font-style: normal;
      color: var(--m-ink);
      font-weight: 500;
    }
  }

  h3 {
    margin-top: 6px;
    font-family: var(--font-serif);
    font-size: 23px;
    line-height: 1.38;
    font-weight: 700;
  }

  p {
    margin-top: 6px;
    font-size: 14px;
    line-height: 1.7;
    color: var(--text-2);
  }
}

.m-arow {
  .thumb {
    width: 84px;
    height: 84px;
    border-radius: 18px;
  }

  b { font-size: 16.5px; }

  .ex {
    white-space: normal;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    line-height: 1.55;
  }
}

.sentinel { height: 1px; }

/* 骨架 */
.sk { padding-top: 4px; }

.grp-sk {
  display: block;
  width: 84px;
  height: 10px;
  margin: 16px 20px 12px;
}

.feat-sk {
  margin: 0 20px;
  height: 218px;
  border-radius: 24px;
}

.m-sk-line { display: block; }

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
    width: 84px;
    height: 84px;
    border-radius: 18px;
    flex: none;
  }
}

.more-sk { padding: 14px 16px; }
</style>
