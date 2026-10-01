/**
 * 关于页模块 data 结构，与模块注册表和编辑器保持一致。
 * 存储层仍是松散 JSON（AboutModule.data: any），读取时经 migrate.ts 归一为这些形状。
 */

export type Span = 1 | 2 | 3;

export type IconName =
  | 'github' | 'mail' | 'rss' | 'telegram' | 'x' | 'link' | 'weibo' | 'bilibili'
  | 'laptop' | 'monitor' | 'keyboard' | 'camera' | 'code' | 'terminal' | 'figma'
  | 'command' | 'cloud' | 'headphones' | 'phone' | 'pen' | 'book' | 'music';

export interface SocialLink {
  name: string;
  handle?: string;
  url: string;
  icon: IconName | string;
  primary?: boolean;
  /** 显示在名片（身份区）的按钮上，最多 3 个；桌面与移动端共用 */
  card?: boolean;
}

export interface Portrait {
  /** 图片 URL；空 = 站点 logo */
  src: string;
  /** 渐隐方向：left = 左缘向左渐隐（默认） */
  fade: 'left' | 'bottom' | 'none';
  /** 渐隐宽度（占图片宽 / 高的百分比，10–80，默认 50）：从完全透明过渡到不透明的距离 */
  fadeWidth?: number;
  /** 圆角 px；未配置时跟随全局 --r-xl */
  radius?: number;
  /** object-position，如 '50% 30%' */
  focus?: string;
}

export interface ProfileData {
  hello?: string;
  name: string;
  alias?: string;
  /** 一句话；*星号* 包裹的片段高亮 */
  lede: string;
  bio: string;
  kicker?: string;
  status: { doing: string; city: string; tz: number };
  links: SocialLink[];
  portrait: Portrait;
}

export interface ChapterData { no?: string; title: string; subtitle?: string; meta?: string }

export type StatusState = 'online' | 'focus' | 'away';
export interface StatusData {
  state: StatusState;
  activity: string;
  app?: string;
  lastActive?: string;
  device?: string;
  note?: string;
}

export interface SocialsData { items: SocialLink[] }

export interface ContactData {
  title: string;
  text: string;
  email?: string;
  buttonText: string;
  url: string;
  sla?: string;
}

export type StatKey = 'days' | 'posts' | 'notes' | 'tags';
export interface StatsData {
  items: { key: StatKey; label?: string; hint?: string }[];
  showYearProgress: boolean;
}

export interface GithubData { showCommits: boolean; commitCount: number }

export interface LanguagesData { unit?: string; items: { name: string; percent: number; color: string }[] }

export interface SkillbarsData { items: { name: string; level: number; tier?: string }[] }

export interface YearEntry {
  sub?: string;
  nums: [string, number][];
  months: (number | null)[];
  pins: number[];
  highlights: [string, string][];
}
export interface YearData { metric?: string; years: Record<string, YearEntry> }

export interface NowItem { kind: string; text: string; note?: string; progress?: number | null }
export interface NowData { updatedAt: string; items: NowItem[] }

export interface Milestone { date: string; title: string; text?: string; now?: boolean }
export interface MilestonesData { items: Milestone[] }

export interface Project {
  name: string;
  desc: string;
  url: string;
  /** 图片 URL；空时用 scene 光影构图 */
  cover?: string;
  scene?: string;
  lang?: string;
  color?: string;
  stars?: number;
  featured?: boolean;
}
export interface ProjectsData { items: Project[] }

export interface PrinciplesData { items: { title: string; text?: string }[] }

export type BookStatus = '在读' | '读完' | '想读';
export interface Book {
  title: string;
  author: string;
  color: string;
  textColor?: string;
  height?: number;
  width?: number;
  status: BookStatus;
  progress?: number;
  note?: string;
  lean?: boolean;
}
export interface BookshelfData { items: Book[] }

export interface Track { title: string; artist: string; album?: string; duration: number; position?: number }
export interface ListeningData {
  playing: boolean;
  now: Track;
  recent: { title: string; artist: string; at: string; scene?: string; cover?: string }[];
}

export interface QuotesData { interval: number; items: { text: string; from?: string }[] }

/** 格言收尾装饰：落款线 / 地平线 / 引号 / 无 */
export type MottoFlourish = 'line' | 'horizon' | 'quote' | 'none';
export const MOTTO_FLOURISHES: MottoFlourish[] = ['line', 'horizon', 'quote', 'none'];

export interface MottoData { text: string; sign?: string; flourish?: MottoFlourish }

export interface GalleryImage { src: string; scene?: string; title?: string; place?: string; date?: string }
export interface GalleryData { images: GalleryImage[] }

export interface Place { name: string; lon: number; lat: number; year?: string; home?: boolean }
export interface PlacesData { region: 'china'; items: Place[] }

export interface FavoriteItem { name: string; by?: string; year?: number | string; note?: string }
export interface FavoritesData { groups: { title: string; items: FavoriteItem[] }[] }

export interface SkillsData { groups: { title: string; items: string[]; star?: string }[] }

export interface StackItem {
  name: string;
  role: string;
  /** 品牌图标键（about/tech.ts）；空 = 用 glyph 文字徽标。旧数据无此字段时按名称猜测补上 */
  icon?: string;
  glyph?: string;
  color?: string;
  url?: string;
}
export interface StackData { items: StackItem[] }

export interface UsesItem { name: string; desc?: string; icon?: string; tag?: string; url?: string }
export interface UsesData { groups: { title: string; items: UsesItem[] }[] }

export interface FaqData { single: boolean; items: { q: string; a: string }[] }

export interface GuestNote { name: string; color?: string; at: string; text: string; likes: number; reply?: string }
/** 留言墙：只存展示选项；留言是真实评论。requireLogin / total / items 为旧版示例字段，读取时忽略 */
export interface GuestbookData { pageSize: number; requireLogin?: boolean; total?: number; items?: GuestNote[] }
