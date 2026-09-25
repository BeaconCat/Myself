import type { AboutModule, SiteConfig } from '../stores/config';

/**
 * 关于页模块数据迁移：旧 schema → about-kit v2 schema。
 *
 * 规则（读取时即时执行，存储保持原样，后台保存时写回新结构）：
 * - devices          → uses：{ items } 变为 uses.groups[0] = { title: '硬件', items }
 * - milestones       → items[].year → date；旧 text 作为 title（无 title 时）
 * - now              → items: string[] → { kind: '在做', text }；补 updatedAt
 * - gallery          → images: string[] → { src }
 * - favorites        → groups[].items: string → { name }
 * - skills           → 补 star
 * - stack            → 补 glyph（取名称前两个字母）
 * - socials/profile  → icon 'twitter' → 'x'
 * - stats            → data {} → 默认四项
 * - 其余：补齐必需数组/字段，避免渲染器与编辑器判空
 */

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type Any = any;

const arr = <T = Any>(v: unknown): T[] => (Array.isArray(v) ? (v as T[]) : []);
const str = (v: unknown, d = ''): string => (typeof v === 'string' ? v : v == null ? d : String(v));
const num = (v: unknown, d = 0): number => {
  const n = typeof v === 'number' ? v : Number(v);
  return Number.isFinite(n) ? n : d;
};
const obj = (v: unknown): Any => (v && typeof v === 'object' && !Array.isArray(v) ? v : {});

function glyphOf(name: string): string {
  const latin = name.replace(/[^A-Za-z0-9]/g, '');
  if (latin.length >= 2) return latin[0].toUpperCase() + latin[1].toLowerCase();
  return name.slice(0, 2);
}

function icon(v: unknown): string {
  const s = str(v, 'link');
  return s === 'twitter' ? 'x' : s;
}

function links(v: unknown): Any[] {
  return arr(v).map((l) => ({ ...obj(l), name: str(obj(l).name), url: str(obj(l).url), icon: icon(obj(l).icon) }));
}

const DATA: Record<string, (d: Any) => Any> = {
  profile: (d) => ({
    ...d,
    hello: str(d.hello),
    name: str(d.name),
    lede: str(d.lede),
    bio: str(d.bio),
    status: { doing: '', city: '', ...obj(d.status), tz: num(obj(d.status).tz, 8) },
    links: links(d.links),
    portrait: {
      src: '',
      fade: 'left',
      ...obj(d.portrait),
    },
  }),
  chapter: (d) => ({ ...d, title: str(d.title), subtitle: str(d.subtitle) }),
  status: (d) => ({ ...d, state: ['online', 'focus', 'away'].includes(d.state) ? d.state : 'online', activity: str(d.activity) }),
  socials: (d) => ({ ...d, items: links(d.items) }),
  contact: (d) => ({ ...d, title: str(d.title), text: str(d.text), buttonText: str(d.buttonText), url: str(d.url) }),
  stats: (d) => ({
    ...d,
    items: arr(d.items).length
      ? arr(d.items).map((x) => ({ ...obj(x), key: str(obj(x).key, 'posts') }))
      : [{ key: 'days' }, { key: 'posts' }, { key: 'notes' }, { key: 'tags' }],
    showYearProgress: d.showYearProgress !== false,
  }),
  github: (d) => ({ ...d, showCommits: d.showCommits !== false, commitCount: num(d.commitCount, 3) }),
  languages: (d) => ({
    ...d,
    items: arr(d.items).map((x) => ({ ...obj(x), name: str(obj(x).name), percent: num(obj(x).percent), color: str(obj(x).color, '#8a96ab') })),
  }),
  skillbars: (d) => ({ ...d, items: arr(d.items).map((x) => ({ ...obj(x), name: str(obj(x).name), level: num(obj(x).level) })) }),
  year: (d) => {
    const years: Record<string, Any> = {};
    for (const [k, v] of Object.entries(obj(d.years))) {
      const y = obj(v);
      const months = arr(y.months).slice(0, 12);
      while (months.length < 12) months.push(null);
      years[k] = {
        ...y,
        nums: arr(y.nums),
        months: months.map((m) => (m == null || m === '' ? null : num(m))),
        pins: arr(y.pins).map((p) => num(p)),
        highlights: arr(y.highlights ?? y.hl),
      };
    }
    return { ...d, years };
  },
  now: (d) => ({
    ...d,
    updatedAt: str(d.updatedAt),
    items: arr(d.items).map((x) =>
      typeof x === 'string' ? { kind: '在做', text: x } : { ...obj(x), kind: str(obj(x).kind, '在做'), text: str(obj(x).text) },
    ),
  }),
  milestones: (d) => ({
    ...d,
    items: arr(d.items).map((x) => {
      const it = obj(x);
      const date = str(it.date ?? it.year);
      const title = str(it.title ?? it.text);
      const text = it.title != null ? str(it.text) : '';
      const { year: _drop, ...rest } = it;
      void _drop;
      return { ...rest, date, title, text };
    }),
  }),
  projects: (d) => ({ ...d, items: arr(d.items).map((x) => ({ ...obj(x), name: str(obj(x).name), desc: str(obj(x).desc), url: str(obj(x).url) })) }),
  principles: (d) => ({ ...d, items: arr(d.items).map((x) => (typeof x === 'string' ? { title: x } : { ...obj(x), title: str(obj(x).title) })) }),
  bookshelf: (d) => ({
    ...d,
    items: arr(d.items).map((x) => ({ ...obj(x), title: str(obj(x).title), author: str(obj(x).author), color: str(obj(x).color ?? obj(x).c, '#1d3a5f'), status: str(obj(x).status, '读完') })),
  }),
  listening: (d) => ({
    ...d,
    playing: d.playing !== false,
    now: { title: '', artist: '', duration: 240, position: 0, ...obj(d.now) },
    recent: arr(d.recent),
  }),
  quotes: (d) => ({
    ...d,
    interval: num(d.interval, 6) || 6,
    items: arr(d.items).map((x) => (typeof x === 'string' ? { text: x } : { ...obj(x), text: str(obj(x).text) })),
  }),
  motto: (d) => ({ ...d, text: str(d.text) }),
  gallery: (d) => ({
    ...d,
    images: arr(d.images).map((x) => (typeof x === 'string' ? { src: x } : { ...obj(x), src: str(obj(x).src) })),
  }),
  places: (d) => ({
    ...d,
    region: 'china',
    items: arr(d.items).map((x) => ({ ...obj(x), name: str(obj(x).name), lon: num(obj(x).lon, 116.4), lat: num(obj(x).lat, 39.9) })),
  }),
  favorites: (d) => ({
    ...d,
    groups: arr(d.groups).map((g) => ({
      ...obj(g),
      title: str(obj(g).title),
      items: arr(obj(g).items).map((x) => (typeof x === 'string' ? { name: x } : { ...obj(x), name: str(obj(x).name) })),
    })),
  }),
  skills: (d) => ({
    ...d,
    groups: arr(d.groups).map((g) => ({ ...obj(g), title: str(obj(g).title), items: arr(obj(g).items).map((x) => str(x)), star: str(obj(g).star) })),
  }),
  stack: (d) => ({
    ...d,
    items: arr(d.items).map((x) => {
      const it = obj(x);
      const name = str(it.name);
      return { ...it, name, role: str(it.role ?? it.desc), glyph: str(it.glyph) || glyphOf(name) };
    }),
  }),
  uses: (d) => ({
    ...d,
    groups: arr(d.groups).map((g) => ({ ...obj(g), title: str(obj(g).title), items: arr(obj(g).items).map((x) => ({ ...obj(x), name: str(obj(x).name) })) })),
  }),
  faq: (d) => ({ ...d, single: d.single !== false, items: arr(d.items).map((x) => ({ ...obj(x), q: str(obj(x).q), a: str(obj(x).a) })) }),
  guestbook: (d) => ({ ...d, pageSize: num(d.pageSize, 4) || 4, requireLogin: !!d.requireLogin, items: arr(d.items) }),
};

/** 单模块迁移：返回新对象（不修改入参） */
export function migrateModule(mod: AboutModule): AboutModule {
  let type = mod.type;
  let data: Any = JSON.parse(JSON.stringify(obj(mod.data)));
  let variant = mod.variant;

  if (type === 'devices') {
    type = 'uses';
    variant = 'groups';
    data = {
      groups: [{
        title: '硬件',
        items: arr(data.items).map((x) => ({ icon: 'laptop', ...obj(x), name: str(obj(x).name), desc: [obj(x).desc, obj(x).spec].filter(Boolean).join(' · ') })),
      }],
    };
  }

  const norm = DATA[type];
  return { ...mod, type, variant, data: norm ? norm(data) : data };
}

/** 原地迁移（后台编辑器用：直接修改响应式对象，保证编辑器拿到新 schema） */
export function migrateInPlace(mod: AboutModule): void {
  const next = migrateModule(mod);
  if (mod.type !== next.type) mod.type = next.type;
  if (next.variant !== undefined && mod.variant !== next.variant) mod.variant = next.variant;
  // 仅补齐/改写字段，保持 data 对象引用不变（编辑器 v-model 绑在其上）
  if (!mod.data || typeof mod.data !== 'object') mod.data = {};
  for (const k of Object.keys(mod.data)) if (!(k in next.data)) delete mod.data[k];
  Object.assign(mod.data, next.data);
}

type LegacyAbout = Partial<Pick<SiteConfig['about'], 'avatar' | 'name' | 'tagline' | 'bio' | 'motto'>>;

/**
 * 整页迁移：逐个 migrateModule；若列表中没有 profile，
 * 用旧的顶层字段 about.avatar/name/tagline/bio 合成一个 profile 放在最前；
 * 若有 about.motto 且列表中没有 motto 模块，在末尾补一个收尾格言。
 */
export function migrateModules(list: AboutModule[] | undefined, about: LegacyAbout = {}): AboutModule[] {
  const out = arr<AboutModule>(list).filter((m) => m && typeof m.type === 'string').map(migrateModule);
  if (!out.some((m) => m.type === 'profile')) {
    const socials = out.find((m) => m.type === 'socials');
    out.unshift(migrateModule({
      id: 'm-profile-legacy',
      type: 'profile',
      span: 3,
      variant: 'portrait',
      data: {
        hello: '你好，我是',
        name: about.name ?? '',
        lede: about.tagline ?? '',
        bio: about.bio ?? '',
        status: { doing: '', city: '', tz: 8 },
        links: socials ? socials.data.items.slice(0, 3).map((l: Any, i: number) => ({ ...l, primary: i === 0 })) : [],
        portrait: { src: about.avatar ?? '', fade: 'left' },
      },
    }));
  }
  if (about.motto && !out.some((m) => m.type === 'motto')) {
    out.push({ id: 'm-motto-legacy', type: 'motto', span: 3, variant: 'closing', data: { text: about.motto, sign: '' } });
  }
  return out;
}
