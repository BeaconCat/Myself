import type { Component } from 'vue';
import EditorProfile from './EditorProfile.vue';
import EditorChapter from './EditorChapter.vue';
import EditorStatus from './EditorStatus.vue';
import EditorSocials from './EditorSocials.vue';
import EditorContact from './EditorContact.vue';
import EditorStats from './EditorStats.vue';
import EditorGithub from './EditorGithub.vue';
import EditorLanguages from './EditorLanguages.vue';
import EditorSkillbars from './EditorSkillbars.vue';
import EditorYear from './EditorYear.vue';
import EditorNow from './EditorNow.vue';
import EditorMilestones from './EditorMilestones.vue';
import EditorProjects from './EditorProjects.vue';
import EditorPrinciples from './EditorPrinciples.vue';
import EditorBookshelf from './EditorBookshelf.vue';
import EditorListening from './EditorListening.vue';
import EditorQuotes from './EditorQuotes.vue';
import EditorMotto from './EditorMotto.vue';
import EditorGallery from './EditorGallery.vue';
import EditorPlaces from './EditorPlaces.vue';
import EditorFavorites from './EditorFavorites.vue';
import EditorSkills from './EditorSkills.vue';
import EditorStack from './EditorStack.vue';
import EditorUses from './EditorUses.vue';
import EditorFaq from './EditorFaq.vue';
import EditorGuestbook from './EditorGuestbook.vue';

export { default as ModuleFrame } from './ModuleFrame.vue';
export { default as ModulePreview } from './ModulePreview.vue';

/**
 * 模块类型 → 行内编辑器组件映射（与 about/registry.ts 的 type 一一对应）。
 * 每个编辑器 props 为 `{ mod: AboutModule }`，直接修改 mod.data（进入时自动 migrateInPlace）。
 * span / variant / title / hidden 由外层 ModuleFrame 统一编辑。
 * devices（旧）→ EditorUses：编辑器 setup 时会把它原地迁移为 uses。
 */
export const MODULE_EDITORS: Record<string, Component> = {
  profile: EditorProfile,
  chapter: EditorChapter,
  status: EditorStatus,
  socials: EditorSocials,
  contact: EditorContact,
  stats: EditorStats,
  github: EditorGithub,
  languages: EditorLanguages,
  skillbars: EditorSkillbars,
  year: EditorYear,
  now: EditorNow,
  milestones: EditorMilestones,
  projects: EditorProjects,
  principles: EditorPrinciples,
  bookshelf: EditorBookshelf,
  listening: EditorListening,
  quotes: EditorQuotes,
  motto: EditorMotto,
  gallery: EditorGallery,
  places: EditorPlaces,
  favorites: EditorFavorites,
  skills: EditorSkills,
  stack: EditorStack,
  uses: EditorUses,
  devices: EditorUses,
  faq: EditorFaq,
  guestbook: EditorGuestbook,
};
