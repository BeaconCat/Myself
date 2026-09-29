import type { AboutModule, SiteConfig } from '../stores/config';
import type { Portrait, SocialLink } from './types';

/**
 * 站点身份：头像、形象图、名字、签名、自述、状态、链接、格言 —— 全站唯一数据源（存于 about 顶层）。
 * 后台「身份」页编辑；关于页的 profile / motto 模块只存展示选项（kicker、收尾装饰），
 * 内容在渲染时由 injectIdentity 注入。其余位置（首页关于卡、页脚、抽屉、文章作者栏）直接读 about 顶层。
 */
export interface Identity {
  avatar: string;
  name: string;
  /** 别名：小名 / 网名，可选；页脚显示为「名字（别名）」 */
  alias: string;
  hello: string;
  /** 签名 / 一句话；*星号* 包裹的片段在关于页高亮 */
  tagline: string;
  bio: string;
  foundedAt: string;
  motto: string;
  mottoSign: string;
  status: { doing: string; city: string; tz: number };
  links: SocialLink[];
  portrait: Portrait;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type Any = any;

const obj = (v: unknown): Any => (v && typeof v === 'object' && !Array.isArray(v) ? v : {});
const filled = (v: unknown): boolean => (typeof v === 'string' ? v.trim() !== '' : v != null);

/** 去掉 *高亮* 标记，给不渲染高亮的位置用 */
export const plainText = (s: string | undefined): string => (s ?? '').replace(/\*/g, '');

/** 历代出厂默认自述：仍是原样（用户没改过）时换成指向「身份」页的新默认文案 */
const DEFAULT_BIO = '这里是 Myself 的默认介绍。前往后台「身份」写下你自己的故事：你是谁、在做什么、热爱什么。';
const LEGACY_BIOS = [
  '这里是 Myself 的默认介绍。前往后台「设置 → 关于信息」写下你自己的故事：你是谁、在做什么、热爱什么。',
  '这里是 Myself 的默认介绍。前往后台「关于管理」写下你自己的故事：你是谁、在做什么、热爱什么。',
  '这里是 Myself 的默认介绍，可在后台「关于管理」修改。',
];

/** profile 模块里旧版存放的内容字段 → 身份字段 */
const PROFILE_CONTENT: Record<string, keyof Identity> = {
  hello: 'hello',
  name: 'name',
  lede: 'tagline',
  bio: 'bio',
  status: 'status',
  links: 'links',
  portrait: 'portrait',
};

/**
 * 归一站点身份（原地修改 about）：
 * 1. 旧版 profile / motto 模块里存过的内容提升到顶层（模块值优先 —— 那是前台此前实际显示的内容），并从模块里删除；
 * 2. 补齐新增字段的默认值，未改动过的旧版默认自述换成新默认文案。
 * 前台（config store 载入时）与后台（关于 / 身份页载入时）各调用一次，结果一致。
 */
export function normalizeIdentity(about: SiteConfig['about']): void {
  const a = about as Any;
  const mods: AboutModule[] = Array.isArray(a.modules) ? a.modules : [];

  for (const m of mods) {
    if (!m || typeof m !== 'object') continue;
    const d = obj(m.data);
    if (m.type === 'profile') {
      for (const [from, to] of Object.entries(PROFILE_CONTENT)) {
        if (!(from in d)) continue;
        if (filled(d[from]) && !(Array.isArray(d[from]) && !d[from].length)) a[to] = d[from];
        delete d[from];
      }
    } else if (m.type === 'motto') {
      if (filled(d.text)) a.motto = d.text;
      if (filled(d.sign)) a.mottoSign = d.sign;
      delete d.text;
      delete d.sign;
    }
  }

  if (LEGACY_BIOS.includes(a.bio)) a.bio = DEFAULT_BIO;
  a.hello ??= '你好，我是';
  a.alias ??= '';
  a.mottoSign ??= '';
  a.status = { doing: '', city: '', tz: 8, ...obj(a.status) };
  a.status.tz = Number.isFinite(Number(a.status.tz)) ? Number(a.status.tz) : 8;
  a.portrait = { src: '', fade: 'left', ...obj(a.portrait) };
  if (!Array.isArray(a.links)) {
    // 从未配置过链接：借用社交模块的前三项，第一项作主按钮
    const socials = mods.find((m) => m?.type === 'socials');
    const items: Any[] = Array.isArray(socials?.data?.items) ? socials!.data.items : [];
    a.links = items.slice(0, 3).map((l, i) => ({ ...obj(l), primary: i === 0 }));
  }
  a.links = a.links.map((l: Any) => ({ name: '', url: '', icon: 'link', ...obj(l) }));
}

/** 渲染时把身份内容注入 profile / motto 模块（返回新模块，不修改入参） */
export function injectIdentity(mod: AboutModule, about: Partial<Identity>): AboutModule {
  if (mod.type === 'profile') {
    return {
      ...mod,
      data: {
        ...obj(mod.data),
        hello: about.hello ?? '',
        name: about.name ?? '',
        alias: about.alias ?? '',
        lede: about.tagline ?? '',
        bio: about.bio ?? '',
        status: { doing: '', city: '', tz: 8, ...obj(about.status) },
        links: about.links ?? [],
        portrait: { src: '', fade: 'left', ...obj(about.portrait) },
      },
    };
  }
  if (mod.type === 'motto') {
    return { ...mod, data: { ...obj(mod.data), text: about.motto ?? '', sign: about.mottoSign ?? '' } };
  }
  return mod;
}

/** 名字（别名）：页脚、署名等处统一展示 */
export function displayName(about: Partial<Identity>, fallback = ''): string {
  const name = (about.name ?? '').trim() || fallback;
  const alias = (about.alias ?? '').trim();
  return alias && alias !== name ? `${name}（${alias}）` : name;
}
