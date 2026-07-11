<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { api } from '../api';
import type { AboutModule } from '../stores/config';
import { useConfigStore } from '../stores/config';
import { metaOf } from './registry';
import GithubStatusCard from '../components/home/GithubStatusCard.vue';
import ImageViewer from '../components/media/ImageViewer.vue';

/** 关于页模块渲染器：按配置顺序渲染全部模块 */
const props = defineProps<{ modules: AboutModule[] }>();

const config = useConfigStore();

const known = computed(() => props.modules.filter((m) => metaOf(m.type)));

/* stats 模块数据 */
const postCount = ref(0);
const noteCount = ref(0);
const tagCount = ref(0);

const daysRunning = computed(() => {
  const from = new Date(config.cfg.about.foundedAt || '2026-01-01').getTime();
  return Math.max(1, Math.floor((Date.now() - from) / 864e5));
});

const stats = computed(() => [
  { value: daysRunning.value, label: '运行天数' },
  { value: postCount.value, label: '文章' },
  { value: noteCount.value, label: '随想' },
  { value: tagCount.value, label: '标签' },
]);

onMounted(async () => {
  if (!props.modules.some((m) => m.type === 'stats')) return;
  try {
    const [posts, notes, tags] = await Promise.all([
      api.posts({ pageSize: 1 }),
      api.notes({ pageSize: 1 }),
      api.tags(),
    ]);
    postCount.value = posts.total;
    noteCount.value = notes.total;
    tagCount.value = tags.length;
  } catch { /* 后端未启动统计留零 */ }
});

/* gallery 查看器 */
const viewerImages = ref<string[]>([]);
const viewerIndex = ref(0);
const viewerOpen = ref(false);

function openGallery(images: string[], index: number): void {
  viewerImages.value = images;
  viewerIndex.value = index;
  viewerOpen.value = true;
}

/* faq 展开 */
const openFaq = ref<string>('');
</script>

<template>
  <div class="mods">
    <section
      v-for="mod in known"
      :key="mod.id"
      v-reveal
      class="mod"
      :class="`mod-${mod.type}`"
    >
      <!-- 站点数字 -->
      <template v-if="mod.type === 'stats'">
        <div class="stats">
          <div v-for="s in stats" :key="s.label" class="stat">
            <strong>{{ s.value }}</strong>
            <span>{{ s.label }}</span>
          </div>
        </div>
      </template>

      <!-- 格言块 -->
      <template v-else-if="mod.type === 'motto'">
        <blockquote class="motto">{{ mod.data.text }}</blockquote>
      </template>

      <!-- 技能分组 -->
      <template v-else-if="mod.type === 'skills'">
        <h2 class="block-title">技能</h2>
        <div class="skills">
          <div v-for="group in mod.data.groups" :key="group.title" class="skill-group">
            <h3>{{ group.title }}</h3>
            <div class="chips">
              <span v-for="item in group.items" :key="item">{{ item }}</span>
            </div>
          </div>
        </div>
      </template>

      <!-- 技能条形图 -->
      <template v-else-if="mod.type === 'skillbars'">
        <h2 class="block-title">能力图谱</h2>
        <div class="bars card-box">
          <div v-for="item in mod.data.items" :key="item.name" class="bar-row">
            <span class="bar-name">{{ item.name }}</span>
            <div class="bar-track">
              <div class="bar-fill" :style="{ width: `${Math.min(100, Math.max(0, item.level))}%` }" />
            </div>
            <span class="bar-val">{{ item.level }}</span>
          </div>
        </div>
      </template>

      <!-- 语言占比条 -->
      <template v-else-if="mod.type === 'languages'">
        <h2 class="block-title">占比</h2>
        <div class="langs card-box">
          <div class="lang-bar">
            <div
              v-for="item in mod.data.items"
              :key="item.name"
              class="lang-seg"
              :style="{ width: `${item.percent}%`, background: item.color }"
              :title="`${item.name} ${item.percent}%`"
            />
          </div>
          <div class="lang-legend">
            <span v-for="item in mod.data.items" :key="item.name">
              <i :style="{ background: item.color }" />{{ item.name }} {{ item.percent }}%
            </span>
          </div>
        </div>
      </template>

      <!-- 历程 -->
      <template v-else-if="mod.type === 'milestones'">
        <h2 class="block-title">历程</h2>
        <div class="milestones">
          <div v-for="(m, i) in mod.data.items" :key="i" class="milestone">
            <span class="year">{{ m.year }}</span>
            <span class="dot" aria-hidden="true" />
            <p>{{ m.text }}</p>
          </div>
        </div>
      </template>

      <!-- 照片墙 -->
      <template v-else-if="mod.type === 'gallery'">
        <h2 class="block-title">照片墙</h2>
        <div class="gallery">
          <button
            v-for="(src, i) in mod.data.images"
            :key="src"
            class="g-cell"
            @click="openGallery(mod.data.images, Number(i))"
          >
            <img :src="src" loading="lazy" alt="" />
          </button>
        </div>
        <p v-if="!mod.data.images?.length" class="empty">在后台上传照片后展示。</p>
      </template>

      <!-- 语录集 -->
      <template v-else-if="mod.type === 'quotes'">
        <h2 class="block-title">语录</h2>
        <div class="quotes">
          <figure v-for="(q, i) in mod.data.items" :key="i" class="quote card-box">
            <blockquote>{{ q.text }}</blockquote>
            <figcaption v-if="q.from">—— {{ q.from }}</figcaption>
          </figure>
        </div>
      </template>

      <!-- 装备清单 -->
      <template v-else-if="mod.type === 'devices'">
        <h2 class="block-title">装备</h2>
        <div class="devices">
          <div v-for="(d, i) in mod.data.items" :key="i" class="device card-box">
            <strong>{{ d.name }}</strong>
            <span>{{ d.desc }}</span>
          </div>
        </div>
      </template>

      <!-- 喜好清单 -->
      <template v-else-if="mod.type === 'favorites'">
        <h2 class="block-title">喜好</h2>
        <div class="skills">
          <div v-for="group in mod.data.groups" :key="group.title" class="skill-group">
            <h3>{{ group.title }}</h3>
            <div class="chips">
              <span v-for="item in group.items" :key="item">{{ item }}</span>
            </div>
          </div>
        </div>
      </template>

      <!-- FAQ -->
      <template v-else-if="mod.type === 'faq'">
        <h2 class="block-title">问答</h2>
        <div class="faq">
          <div
            v-for="(f, i) in mod.data.items"
            :key="i"
            class="faq-item card-box"
            :class="{ open: openFaq === `${mod.id}-${i}` }"
          >
            <button class="faq-q" @click="openFaq = openFaq === `${mod.id}-${i}` ? '' : `${mod.id}-${i}`">
              <span>{{ f.q }}</span>
              <i aria-hidden="true">+</i>
            </button>
            <p class="faq-a">{{ f.a }}</p>
          </div>
        </div>
      </template>

      <!-- 正在做 -->
      <template v-else-if="mod.type === 'now'">
        <h2 class="block-title">现在</h2>
        <ul class="now card-box">
          <li v-for="(item, i) in mod.data.items" :key="i">
            <span class="pulse" aria-hidden="true" />{{ item }}
          </li>
        </ul>
      </template>

      <!-- GitHub 状态卡 -->
      <template v-else-if="mod.type === 'github'">
        <h2 class="block-title">GitHub</h2>
        <GithubStatusCard />
      </template>

      <!-- 社交链接 -->
      <template v-else-if="mod.type === 'socials'">
        <h2 class="block-title">找到我</h2>
        <div class="links">
          <a
            v-for="l in mod.data.items"
            :key="l.name"
            :href="l.url"
            class="link"
            target="_blank"
            rel="noopener"
          >
            <svg v-if="l.icon === 'github'" viewBox="0 0 16 16" fill="currentColor">
              <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8Z" />
            </svg>
            <svg v-else-if="l.icon === 'mail'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="5" width="18" height="14" rx="2" />
              <path d="m3 7 9 6 9-6" />
            </svg>
            <svg v-else-if="l.icon === 'rss'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round">
              <path d="M4 11a9 9 0 0 1 9 9M4 4a16 16 0 0 1 16 16" />
              <circle cx="5" cy="19" r="1.6" fill="currentColor" stroke="none" />
            </svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M10 14a5 5 0 0 0 7.07 0l2.83-2.83a5 5 0 0 0-7.07-7.07L11.4 5.5" />
              <path d="M14 10a5 5 0 0 0-7.07 0L4.1 12.83a5 5 0 0 0 7.07 7.07l1.42-1.4" />
            </svg>
            <span>{{ l.name }}</span>
          </a>
        </div>
      </template>

      <!-- 技术栈 -->
      <template v-else-if="mod.type === 'stack'">
        <h2 class="block-title">技术栈</h2>
        <div class="stack">
          <div v-for="s in mod.data.items" :key="s.name" class="stack-item">
            <strong>{{ s.name }}</strong>
            <span>{{ s.role }}</span>
          </div>
        </div>
      </template>

      <!-- 联系 CTA -->
      <template v-else-if="mod.type === 'contact'">
        <div class="contact card-box">
          <div class="contact-text">
            <h2>{{ mod.data.title }}</h2>
            <p>{{ mod.data.text }}</p>
          </div>
          <a class="contact-btn" :href="mod.data.url" target="_blank" rel="noopener">
            {{ mod.data.buttonText }}
          </a>
        </div>
      </template>
    </section>

    <ImageViewer
      v-if="viewerOpen"
      :images="viewerImages"
      :start-index="viewerIndex"
      @close="viewerOpen = false"
    />
  </div>
</template>

<style scoped lang="scss">
.mod { margin-bottom: 56px; }

.block-title {
  font-size: 24px;
  margin-bottom: 22px;
  padding-left: 14px;
  border-left: 4px solid var(--primary);
}

.card-box {
  background: var(--surface);
  border: 1px solid var(--border);
  padding: 22px 24px;
}

.empty { color: var(--text-2); font-size: 13px; }

/* 站点数字 */
.stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.stat {
  text-align: center;
  padding: 20px 8px;
  background: var(--surface);
  border: 1px solid var(--border);
  transition: transform var(--dur-fast) var(--ease-out), border-color var(--dur-fast);

  strong {
    display: block;
    font-family: var(--font-serif);
    font-size: 26px;
    background: var(--grad-title);
    background-clip: text;
    -webkit-background-clip: text;
    color: transparent;
  }

  span { display: block; margin-top: 6px; font-size: 12px; color: var(--text-2); }
  &:hover { transform: scale(1.04); border-color: var(--primary); }
}

/* 格言 */
.motto {
  margin: 0 auto;
  padding: 14px 30px;
  width: fit-content;
  max-width: 100%;
  font-family: var(--font-serif);
  font-size: 17px;
  color: var(--text-2);
  border-left: 3px solid rgba(var(--primary-rgb), 0.6);
  border-right: 3px solid rgba(var(--primary-rgb), 0.6);
  background: rgba(var(--primary-rgb), 0.05);
  border-radius: 8px;
}

/* 技能/喜好分组 */
.skills {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.skill-group {
  padding: 20px;
  background: var(--surface);
  border: 1px solid var(--border);

  h3 { font-size: 15px; margin-bottom: 12px; color: var(--primary); }
}

.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;

  span {
    font-size: 12px;
    font-weight: 600;
    padding: 4px 12px;
    background: var(--surface-2);
    transition: all var(--dur-fast) var(--ease-spring);

    &:hover {
      transform: scale(1.08);
      background: rgba(var(--primary-rgb), 0.12);
      color: var(--primary);
    }
  }
}

/* 技能条形图 */
.bars { display: flex; flex-direction: column; gap: 14px; }

.bar-row {
  display: flex;
  align-items: center;
  gap: 14px;
}

.bar-name { width: 90px; font-size: 13.5px; font-weight: 600; flex-shrink: 0; }

.bar-track {
  flex: 1;
  height: 10px;
  border-radius: 999px;
  background: var(--surface-2);
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--primary), var(--primary-deep));
  box-shadow: 0 0 8px rgba(var(--primary-rgb), 0.4);
  transition: width var(--dur-slow) var(--ease-out);
}

.bar-val {
  width: 34px;
  text-align: right;
  font-size: 12px;
  color: var(--text-2);
  font-variant-numeric: tabular-nums;
}

/* 语言占比 */
.lang-bar {
  display: flex;
  height: 16px;
  border-radius: 999px;
  overflow: hidden;
  margin-bottom: 14px;
}

.lang-seg { transition: width var(--dur-slow) var(--ease-out); }

.lang-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  font-size: 12.5px;
  color: var(--text-2);

  i {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 3px;
    margin-right: 6px;
    vertical-align: -1px;
  }
}

/* 历程 */
.milestones { display: flex; flex-direction: column; }

.milestone {
  display: flex;
  align-items: baseline;
  gap: 18px;
  padding: 16px 4px;
  border-bottom: 1px solid var(--border);
  transition: background var(--dur-fast);

  &:hover {
    background: rgba(var(--primary-rgb), 0.04);
    .dot { transform: scale(1.4); box-shadow: 0 0 10px rgba(var(--primary-rgb), 0.6); }
  }

  .year {
    font-family: var(--font-serif);
    font-size: 20px;
    font-weight: 700;
    color: var(--primary);
    flex-shrink: 0;
    width: 72px;
  }

  .dot {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    background: var(--primary);
    flex-shrink: 0;
    align-self: center;
    transition: all var(--dur-fast) var(--ease-spring);
  }

  p { font-size: 14.5px; line-height: 1.8; color: var(--text-2); }
}

/* 照片墙 */
.gallery {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 6px;
}

.g-cell {
  border: none;
  padding: 0;
  background: var(--surface-2);
  aspect-ratio: 1;
  overflow: hidden;
  cursor: zoom-in;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    transition: transform var(--dur) var(--ease-out);
  }

  &:hover img { transform: scale(1.06); }
}

/* 语录 */
.quotes {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 14px;
}

.quote {
  blockquote {
    font-family: var(--font-serif);
    font-size: 15.5px;
    line-height: 1.9;
  }

  figcaption {
    margin-top: 10px;
    font-size: 12.5px;
    color: var(--text-2);
    text-align: right;
  }
}

/* 装备 */
.devices {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
}

.device {
  display: flex;
  flex-direction: column;
  gap: 6px;
  transition: transform var(--dur-fast) var(--ease-out), border-color var(--dur-fast);

  strong { font-size: 15px; }
  span { font-size: 12.5px; color: var(--text-2); }
  &:hover { transform: scale(1.03); border-color: var(--primary); }
}

/* FAQ */
.faq { display: flex; flex-direction: column; gap: 10px; }

.faq-item {
  padding: 0;
  overflow: hidden;

  .faq-q {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    padding: 15px 20px;
    border: none;
    background: none;
    color: var(--text);
    font-size: 14.5px;
    font-weight: 600;
    text-align: left;

    i {
      font-style: normal;
      font-size: 18px;
      color: var(--primary);
      transition: transform var(--dur-fast) var(--ease-spring);
    }
  }

  .faq-a {
    max-height: 0;
    overflow: hidden;
    padding: 0 20px;
    font-size: 14px;
    line-height: 1.8;
    color: var(--text-2);
    transition: max-height var(--dur) var(--ease-out), padding var(--dur) ease;
  }

  &.open {
    .faq-q i { transform: rotate(45deg); }
    .faq-a { max-height: 300px; padding: 0 20px 16px; }
  }
}

/* 正在做 */
.now {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 12px;

  li {
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 14.5px;
  }
}

.pulse {
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background: var(--primary);
  flex-shrink: 0;
  animation: pulse-glow 1.8s ease-in-out infinite;
}

@keyframes pulse-glow {
  0%, 100% { box-shadow: 0 0 0 0 rgba(var(--primary-rgb), 0.5); }
  50% { box-shadow: 0 0 0 6px rgba(var(--primary-rgb), 0); }
}

/* 社交链接 / 技术栈（与旧关于页同款） */
.links { display: flex; gap: 14px; flex-wrap: wrap; }

.link {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 12px 24px;
  background: var(--surface);
  border: 1px solid var(--border);
  font-size: 14px;
  font-weight: 600;
  transition: all var(--dur-fast) var(--ease-out);

  svg { width: 18px; height: 18px; }

  &:hover {
    transform: scale(1.05);
    border-color: var(--primary);
    color: var(--primary);
    box-shadow: 0 6px 20px -6px rgba(var(--primary-rgb), 0.4);
  }
}

.stack {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 12px;
}

.stack-item {
  text-align: center;
  padding: 18px 10px;
  background: var(--surface);
  border: 1px solid var(--border);
  transition: transform var(--dur-fast) var(--ease-out), border-color var(--dur-fast);

  strong { display: block; font-family: var(--font-serif); font-size: 16px; }
  span { display: block; margin-top: 6px; font-size: 12px; color: var(--text-2); }
  &:hover { transform: scale(1.05); border-color: var(--primary); }
}

/* 联系 CTA */
.contact {
  display: flex;
  align-items: center;
  gap: 22px;
  background:
    radial-gradient(400px 160px at 15% 0%, rgba(var(--primary-rgb), 0.1), transparent 70%),
    var(--surface);

  .contact-text {
    flex: 1;

    h2 { font-size: 21px; margin-bottom: 8px; }
    p { font-size: 14px; color: var(--text-2); line-height: 1.8; }
  }
}

.contact-btn {
  flex-shrink: 0;
  padding: 12px 28px;
  font-size: 15px;
  font-weight: 700;
  color: #fff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.35);
  background: linear-gradient(180deg, var(--primary), var(--primary-deep));
  box-shadow: 0 4px 14px rgba(var(--primary-rgb), 0.45);
  transition: transform var(--dur-fast) var(--ease-out), filter var(--dur-fast);

  &:hover { filter: brightness(1.08); transform: scale(1.05); }
}

@media (max-width: 720px) {
  .stats { grid-template-columns: repeat(2, 1fr); }
  .contact { flex-direction: column; text-align: center; }
}
</style>
