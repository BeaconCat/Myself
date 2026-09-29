import {
  siBilibili,
  siBluesky,
  siDiscord,
  siDouban,
  siFacebook,
  siFigma,
  siGitee,
  siGithub,
  siGitlab,
  siInstagram,
  siJuejin,
  siMastodon,
  siNeteasecloudmusic,
  siPixiv,
  siQq,
  siReddit,
  siSinaweibo,
  siSpotify,
  siSteam,
  siTelegram,
  siThreads,
  siTiktok,
  siTwitch,
  siWechat,
  siX,
  siXiaohongshu,
  siYoutube,
  siZhihu,
  type SimpleIcon,
} from 'simple-icons';

/**
 * 社交平台注册表：品牌图标取自离线安装的 simple-icons（官方 24 网格实心路径 + 品牌色）。
 * 键即 SocialLink.icon 的取值；KitIcon 遇到这些键时渲染实心品牌图标，其余仍走线性图标集。
 * mono = 品牌色为黑 / 近黑，深色模式下不可读，界面上改用正文色。
 */
export interface Brand {
  label: string;
  path: string;
  hex: string;
  mono: boolean;
}

const b = (label: string, icon: SimpleIcon): Brand => {
  const hex = `#${icon.hex}`;
  const n = parseInt(icon.hex, 16);
  const lum = 0.299 * ((n >> 16) & 255) + 0.587 * ((n >> 8) & 255) + 0.114 * (n & 255);
  return { label, path: icon.path, hex, mono: lum < 48 };
};

export const BRANDS: Record<string, Brand> = {
  github: b('GitHub', siGithub),
  bilibili: b('哔哩哔哩', siBilibili),
  youtube: b('YouTube', siYoutube),
  x: b('X', siX),
  weibo: b('微博', siSinaweibo),
  zhihu: b('知乎', siZhihu),
  xiaohongshu: b('小红书', siXiaohongshu),
  douyin: b('抖音', siTiktok),
  wechat: b('微信', siWechat),
  qq: b('QQ', siQq),
  juejin: b('稀土掘金', siJuejin),
  douban: b('豆瓣', siDouban),
  netease: b('网易云音乐', siNeteasecloudmusic),
  telegram: b('Telegram', siTelegram),
  discord: b('Discord', siDiscord),
  instagram: b('Instagram', siInstagram),
  threads: b('Threads', siThreads),
  bluesky: b('Bluesky', siBluesky),
  mastodon: b('Mastodon', siMastodon),
  facebook: b('Facebook', siFacebook),
  reddit: b('Reddit', siReddit),
  twitch: b('Twitch', siTwitch),
  steam: b('Steam', siSteam),
  spotify: b('Spotify', siSpotify),
  pixiv: b('pixiv', siPixiv),
  gitlab: b('GitLab', siGitlab),
  gitee: b('Gitee', siGitee),
  figma: b('Figma', siFigma),
};

/** 名片上最多展示的链接按钮数 */
export const CARD_LINK_MAX = 3;

/**
 * 名片按钮：勾选了「名片」的链接（最多 3 个）；一个都没勾时取前 3 个。
 * 桌面身份区与移动端名片共用，保证两端一致。
 */
export function cardLinks<T extends { url?: string; card?: boolean }>(links: T[] | undefined): T[] {
  const list = (links ?? []).filter((l) => l?.url);
  const picked = list.filter((l) => l.card);
  return (picked.length ? picked : list).slice(0, CARD_LINK_MAX);
}

/** 按钮着色：品牌色（黑色系品牌用正文色） */
export function brandColor(icon: string): string | undefined {
  const br = BRANDS[icon];
  if (!br) return undefined;
  return br.mono ? 'var(--text)' : br.hex;
}
