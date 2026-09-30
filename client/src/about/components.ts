import type { Component } from 'vue';
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

/** 模块类型 → 渲染组件（前台关于页与后台预览共用） */
export const MODULE_COMPONENTS: Record<string, Component> = {
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
