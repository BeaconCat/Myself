<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch, type Component } from 'vue';
import type { AboutModule } from '../stores/config';
import { metaOf, spanOf, titleOf, variantOf } from './registry';
import { migrateModules } from './migrate';
import { injectIdentity, type Identity } from './identity';
import { kitToast } from './toast';
import KitIcon from './parts/KitIcon.vue';
import ProfileModule from './modules/ProfileModule.vue';
import ChapterModule from './modules/ChapterModule.vue';
import StatusModule from './modules/StatusModule.vue';
import SocialsModule from './modules/SocialsModule.vue';
import ContactModule from './modules/ContactModule.vue';
import StatsModule from './modules/StatsModule.vue';
import GithubModule from './modules/GithubModule.vue';
import LanguagesModule from './modules/LanguagesModule.vue';
import SkillbarsModule from './modules/SkillbarsModule.vue';
import YearModule from './modules/YearModule.vue';
import NowModule from './modules/NowModule.vue';
import MilestonesModule from './modules/MilestonesModule.vue';
import ProjectsModule from './modules/ProjectsModule.vue';
import PrinciplesModule from './modules/PrinciplesModule.vue';
import BookshelfModule from './modules/BookshelfModule.vue';
import ListeningModule from './modules/ListeningModule.vue';
import QuotesModule from './modules/QuotesModule.vue';
import MottoModule from './modules/MottoModule.vue';
import GalleryModule from './modules/GalleryModule.vue';
import PlacesModule from './modules/PlacesModule.vue';
import FavoritesModule from './modules/FavoritesModule.vue';
import SkillsModule from './modules/SkillsModule.vue';
import StackModule from './modules/StackModule.vue';
import UsesModule from './modules/UsesModule.vue';
import FaqModule from './modules/FaqModule.vue';
import GuestbookModule from './modules/GuestbookModule.vue';
import './kit.scss';

/**
 * 关于页模块分发器（about-kit v2）：
 * 1. migrateModules 归一旧 schema（缺 profile 时补身份区），再由 injectIdentity 注入站点身份内容；
 * 2. 以 chapter 为界切分成若干 12 栏 bento 段落（段内 dense 回填，段间不串位）；
 * 3. 统一外壳（卡片 / 开放排版）、IO reveal + stagger、指针跟随中性高光、共享 tooltip 与轻提示。
 * 主色可读性（亮主色配深字等）由全站 --solid / --on-solid / --ink 统一派生，这里不再单独计算。
 */
const props = defineProps<{
  modules: AboutModule[];
  /** 站点身份：注入 profile / motto 模块的内容 */
  about?: Partial<Identity>;
}>();

const COMPONENTS: Record<string, Component> = {
  profile: ProfileModule,
  chapter: ChapterModule,
  status: StatusModule,
  socials: SocialsModule,
  contact: ContactModule,
  stats: StatsModule,
  github: GithubModule,
  languages: LanguagesModule,
  skillbars: SkillbarsModule,
  year: YearModule,
  now: NowModule,
  milestones: MilestonesModule,
  projects: ProjectsModule,
  principles: PrinciplesModule,
  bookshelf: BookshelfModule,
  listening: ListeningModule,
  quotes: QuotesModule,
  motto: MottoModule,
  gallery: GalleryModule,
  places: PlacesModule,
  favorites: FavoritesModule,
  skills: SkillsModule,
  stack: StackModule,
  uses: UsesModule,
  faq: FaqModule,
  guestbook: GuestbookModule,
};

interface Item {
  mod: AboutModule;
  span: 1 | 2 | 3;
  variant: string;
  title: string;
  chrome: boolean;
  no?: string;
}

const sections = computed<Item[][]>(() => {
  const list = migrateModules(props.modules, props.about)
    .filter((m) => !m.hidden && COMPONENTS[m.type] && metaOf(m.type))
    .map((m) => injectIdentity(m, props.about ?? {}));
  const out: Item[][] = [[]];
  let chapter = 0;
  for (const mod of list) {
    const variant = variantOf(mod);
    const item: Item = {
      mod,
      span: spanOf(mod),
      variant,
      title: titleOf(mod),
      chrome: metaOf(mod.type)!.chrome(variant),
    };
    if (mod.type === 'chapter') {
      chapter += 1;
      item.no = String(chapter).padStart(2, '0');
      if (out[out.length - 1].length) out.push([]);
    }
    out[out.length - 1].push(item);
  }
  return out.filter((s) => s.length);
});

/* ---------- reveal：进入视口依次添加 .in（同一批 80ms 错开） ---------- */
const root = ref<HTMLElement | null>(null);
let io: IntersectionObserver | null = null;
const reduce = () => window.matchMedia('(prefers-reduced-motion: reduce)').matches;

function observe(): void {
  if (!root.value || !io) return;
  root.value.querySelectorAll<HTMLElement>('.rv:not(.in):not([data-obs])').forEach((el) => {
    el.dataset.obs = '1';
    io!.observe(el);
  });
}

onMounted(() => {
  io = new IntersectionObserver((entries) => {
    let k = 0;
    for (const e of entries) {
      if (!e.isIntersecting) continue;
      const el = e.target as HTMLElement;
      el.style.setProperty('--d', reduce() ? '0ms' : `${k++ * 80}ms`);
      el.classList.add('in');
      io?.unobserve(el);
    }
  }, { threshold: 0.12, rootMargin: '0px 0px -6% 0px' });
  observe();
});

watch(sections, () => nextTick(observe));

onBeforeUnmount(() => io?.disconnect());

/* ---------- 指针跟随高光 ---------- */
function onPointerMove(e: PointerEvent): void {
  const card = (e.target as HTMLElement).closest?.('.ak-m.card') as HTMLElement | null;
  if (!card) return;
  const r = card.getBoundingClientRect();
  card.style.setProperty('--mx', `${e.clientX - r.left}px`);
  card.style.setProperty('--my', `${e.clientY - r.top}px`);
}

/* ---------- 共享 tooltip：任意 [data-tip] 元素 ---------- */
const tip = reactive({ on: false, text: '', x: 0, y: 0 });
function onPointerOver(e: PointerEvent): void {
  const el = (e.target as HTMLElement).closest?.('[data-tip]') as HTMLElement | null;
  if (!el) { tip.on = false; return; }
  const r = el.getBoundingClientRect();
  tip.text = el.dataset.tip ?? '';
  tip.x = r.left + r.width / 2;
  tip.y = r.top;
  tip.on = true;
}
</script>

<template>
  <div
    ref="root"
    class="ak"
    @pointermove.passive="onPointerMove"
    @pointerover.passive="onPointerOver"
    @pointerleave="tip.on = false"
  >
    <div v-for="(sec, si) in sections" :key="si" class="ak-sec">
      <section
        v-for="it in sec"
        :key="it.mod.id"
        class="ak-m rv"
        :class="[`m-${it.mod.type}`, { card: it.chrome }]"
        :data-span="it.span"
        :data-type="it.mod.type"
      >
        <div class="ak-body">
          <component
            :is="COMPONENTS[it.mod.type]"
            :mod="it.mod"
            :variant="it.variant"
            :span="it.span"
            :title="it.title"
            :no="it.no"
          />
        </div>
      </section>
    </div>

    <div class="ak-tip" :class="{ on: tip.on }" :style="{ left: `${tip.x}px`, top: `${tip.y}px` }">{{ tip.text }}</div>
    <div class="ak-toast" :class="{ on: kitToast.on }" role="status"><KitIcon name="ok" :size="15" />{{ kitToast.text }}</div>
  </div>
</template>

<style scoped lang="scss">
.ak { position: relative; }

.ak-sec + .ak-sec { margin-top: var(--ak-gap); }

.ak-tip {
  position: fixed;
  z-index: 90;
  padding: 6px 10px;
  border-radius: var(--r-sm);
  white-space: nowrap;
  pointer-events: none;
  font: 500 11.5px var(--ak-mono);
  color: var(--bg);
  background: var(--text);
  opacity: 0;
  transform: translate(-50%, -120%) translateY(4px);
  transition: opacity 0.15s, transform 0.15s;

  &.on { opacity: 1; transform: translate(-50%, -120%); }
}

.ak-toast {
  position: fixed;
  left: 50%;
  bottom: 90px;
  z-index: 95;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border-radius: var(--r-md);
  box-shadow: var(--shadow-pop);
  font-size: 13px;
  color: var(--bg);
  background: var(--text);
  opacity: 0;
  translate: -50% 20px;
  pointer-events: none;
  transition: opacity 0.4s var(--ease-spring), translate 0.4s var(--ease-spring);

  &.on { opacity: 1; translate: -50% 0; }
}
</style>
