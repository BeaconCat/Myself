import {
  ArrowUpRight, BookOpen, Check, ChevronLeft, ChevronRight, Cloud, Code, Command, Copy, Headphones, Heart, Keyboard,
  Laptop, Link, Mail, Monitor, Music, Pause, PenLine, Play, RefreshCw, Rss, Send, SkipBack, SkipForward, Smartphone,
  SquareTerminal, Star, X, Camera,
} from 'lucide';
import { iconMarkup, iconSet } from '../utils/lucide';

/** 实心版（播放器按钮）：Lucide 线稿加填充 */
const solid = (markup: string): string => markup.replace(/\/>/g, ' fill="currentColor"/>');

/**
 * 关于页通用图标：名字 → Lucide 标准线性图标（SVG 内部片段），经 KitIcon 渲染。
 * 社交平台（github / bilibili / x …）不在这里，由 about/brands.ts 的品牌图标承担。
 * 名称同时是 socials / uses 数据里 icon 字段的取值。
 */
export const ICONS: Record<string, string> = {
  ...iconSet({
    mail: Mail,
    rss: Rss,
    link: Link,
    arrow: ArrowUpRight,
    star: Star,
    copy: Copy,
    ok: Check,
    left: ChevronLeft,
    right: ChevronRight,
    close: X,
    heart: Heart,
    send: Send,
    laptop: Laptop,
    keyboard: Keyboard,
    camera: Camera,
    monitor: Monitor,
    code: Code,
    pen: PenLine,
    command: Command,
    cloud: Cloud,
    terminal: SquareTerminal,
    headphones: Headphones,
    phone: Smartphone,
    book: BookOpen,
    music: Music,
    cycle: RefreshCw,
  }),
  play: solid(iconMarkup(Play)),
  pause: solid(iconMarkup(Pause)),
  prev: solid(iconMarkup(SkipBack)),
  next: solid(iconMarkup(SkipForward)),
};

/** 可供后台选择的图标名（社交 / 工作台） */
export const SOCIAL_ICONS = [
  'github', 'bilibili', 'youtube', 'x', 'weibo', 'zhihu', 'xiaohongshu', 'douyin', 'wechat', 'qq',
  'juejin', 'douban', 'netease', 'telegram', 'discord', 'instagram', 'threads', 'bluesky', 'mastodon',
  'facebook', 'reddit', 'twitch', 'steam', 'spotify', 'pixiv', 'gitlab', 'gitee',
  'mail', 'rss', 'link',
] as const;
export const USES_ICONS = ['laptop', 'monitor', 'keyboard', 'camera', 'headphones', 'phone', 'code', 'terminal', 'figma', 'command', 'cloud', 'pen', 'book', 'music', 'link'] as const;

/** 无真实照片时的默认图：12 幅默认封面（client/public/covers/NN.webp），后台下拉按名字选 */
export const SCENES = ['01', '02', '03', '04', '05', '06', '07', '08', '09', '10', '11', '12'] as const;
export const SCENE_LABELS: Record<string, string> = {
  '01': '地平线', '02': '月相', '03': '叠纸', '04': '拱窗', '05': '潮汐', '06': '格物',
  '07': '光斑', '08': '山影', '09': '书脊', '10': '圆舞', '11': '雨线', '12': '远帆',
};
/** 旧版 CSS 光影场景名 → 相近的默认图（读取旧数据时兼容） */
export const LEGACY_SCENES: Record<string, string> = {
  door: '05', window: '04', sea: '12', tunnel: '02', dusk: '01', snow: '08', page: '03', grid: '06',
};
