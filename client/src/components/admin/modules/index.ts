import type { Component } from 'vue';
import EditorInfo from './EditorInfo.vue';
import EditorMotto from './EditorMotto.vue';
import EditorGroups from './EditorGroups.vue';
import EditorSkillbars from './EditorSkillbars.vue';
import EditorLanguages from './EditorLanguages.vue';
import EditorMilestones from './EditorMilestones.vue';
import EditorGallery from './EditorGallery.vue';
import EditorQuotes from './EditorQuotes.vue';
import EditorTiles from './EditorTiles.vue';
import EditorFaq from './EditorFaq.vue';
import EditorNow from './EditorNow.vue';
import EditorSocials from './EditorSocials.vue';
import EditorContact from './EditorContact.vue';

/**
 * 模块类型 → 行内编辑器组件映射（与 about/registry.ts 的 type 一一对应）。
 * 每个编辑器接收 `mod: AboutModule`，直接修改 mod.data。
 */
export const MODULE_EDITORS: Record<string, Component> = {
  stats: EditorInfo,
  github: EditorInfo,
  motto: EditorMotto,
  skills: EditorGroups,
  favorites: EditorGroups,
  skillbars: EditorSkillbars,
  languages: EditorLanguages,
  milestones: EditorMilestones,
  gallery: EditorGallery,
  quotes: EditorQuotes,
  devices: EditorTiles,
  stack: EditorTiles,
  faq: EditorFaq,
  now: EditorNow,
  socials: EditorSocials,
  contact: EditorContact,
};
