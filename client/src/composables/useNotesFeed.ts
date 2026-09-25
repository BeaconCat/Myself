import { computed, onBeforeUnmount, onMounted, ref, type Ref } from 'vue';
import { api, type Note } from '../api';
import { useLoadingStore } from '../stores/loading';

export type NotesTab = 'posts' | 'media';

/**
 * 随想信息流：帖文 / 媒体双视图 + 搜索防抖 + 时间筛选 + 分段加载（20/页，滚动续载）。
 * 首屏加载期间按住路由揭幕（useLoadingStore.holdRoute）。
 * @param sentinel 触底哨兵元素（页面模板 ref），进入视口 400px 内即续载
 */
export function useNotesFeed(sentinel: Readonly<Ref<HTMLElement | null>>) {
  const notes = ref<Note[]>([]);
  const tab = ref<NotesTab>('posts');
  const keyword = ref('');
  const loading = ref(false);
  const page = ref(1);
  const total = ref(0);
  const PAGE_SIZE = 20;

  /* 时间筛选 */
  const dateFrom = ref('');
  const dateTo = ref('');

  function applyQuickRange(days: number): void {
    const end = new Date();
    const start = new Date(Date.now() - (days - 1) * 864e5);
    const p2 = (n: number) => String(n).padStart(2, '0');
    /* 本地日期（toISOString 是 UTC，东八区凌晨会差一天） */
    const fmt = (d: Date) => `${d.getFullYear()}-${p2(d.getMonth() + 1)}-${p2(d.getDate())}`;
    dateFrom.value = fmt(start);
    dateTo.value = fmt(end);
    void reload();
  }

  function clearDate(): void {
    dateFrom.value = '';
    dateTo.value = '';
    void reload();
  }

  /** 指定起止日期（YYYY-MM-DD，含首尾；from = to 即单日） */
  function setRange(from: string, to: string): void {
    dateFrom.value = from <= to ? from : to;
    dateTo.value = from <= to ? to : from;
    void reload();
  }

  function queryParams(nextPage: number) {
    return {
      page: nextPage,
      pageSize: PAGE_SIZE,
      q: keyword.value || undefined,
      media: tab.value === 'media' || undefined,
      from: dateFrom.value || undefined,
      to: dateTo.value || undefined,
    };
  }

  /** 请求序号：快速连续筛选时只采纳最后一次的结果 */
  let seq = 0;
  /** 至少完成过一次加载（首屏骨架判断用） */
  const loadedOnce = ref(false);

  async function reload(): Promise<void> {
    const my = ++seq;
    loading.value = true;
    try {
      const res = await api.notes(queryParams(1));
      if (my !== seq) return;
      notes.value = res.items;
      total.value = res.total;
      page.value = 1;
    } finally {
      if (my === seq) {
        loading.value = false;
        loadedOnce.value = true;
      }
    }
  }

  const loadingMore = ref(false);
  const hasMore = computed(() => notes.value.length < total.value);

  async function loadMore(): Promise<void> {
    if (loadingMore.value || loading.value || !hasMore.value) return;
    loadingMore.value = true;
    const my = seq;
    try {
      const res = await api.notes(queryParams(page.value + 1));
      if (my !== seq) return;
      notes.value = [...notes.value, ...res.items];
      total.value = res.total;
      page.value += 1;
    } finally {
      loadingMore.value = false;
    }
  }

  /* 触底哨兵 */
  let observer: IntersectionObserver | null = null;

  onMounted(() => {
    observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((e) => e.isIntersecting)) void loadMore();
      },
      { rootMargin: '400px' },
    );
    if (sentinel.value) observer.observe(sentinel.value);
  });

  onBeforeUnmount(() => observer?.disconnect());

  function switchTab(next: NotesTab): void {
    if (tab.value === next) return;
    tab.value = next;
    void reload();
  }

  let debounce = 0;
  function onSearch(): void {
    window.clearTimeout(debounce);
    debounce = window.setTimeout(() => void reload(), 300);
  }

  onMounted(async () => {
    const release = useLoadingStore().holdRoute();
    try {
      await reload();
    } finally {
      release();
    }
  });
  onBeforeUnmount(() => window.clearTimeout(debounce));

  return {
    PAGE_SIZE,
    notes,
    tab,
    keyword,
    loading,
    loadingMore,
    hasMore,
    dateFrom,
    dateTo,
    loadedOnce,
    reload,
    switchTab,
    onSearch,
    applyQuickRange,
    clearDate,
    setRange,
  };
}

/**
 * 随想索引（桌面右栏 / 日期弹层）：按月取当月全部随想，并入「有随想的日子」集合；
 * 另统计心情计数。心情统计取最近 200 条（个人站量级足够，免去专用聚合接口）。
 */
export function useNotesIndex() {
  const days = ref(new Set<string>());
  const moods = ref<{ name: string; count: number }[]>([]);
  const loaded = new Set<string>();

  /** 取某月（YYYY-MM）全部随想的日期并入 days；已取过的月份跳过 */
  async function loadMonth(ym: string): Promise<void> {
    if (loaded.has(ym)) return;
    loaded.add(ym);
    const [y, m] = ym.split('-').map(Number);
    const last = new Date(y, m, 0).getDate();
    const found: string[] = [];
    for (let page = 1; page <= 6; page++) {
      const res = await api
        .notes({ page, pageSize: 50, from: `${ym}-01`, to: `${ym}-${String(last).padStart(2, '0')}` })
        .catch(() => null);
      if (!res) {
        loaded.delete(ym);
        break;
      }
      res.items.forEach((n) => found.push(n.createdAt.slice(0, 10)));
      if (page * 50 >= res.total) break;
    }
    if (found.length) days.value = new Set([...days.value, ...found]);
  }

  async function loadMoods(): Promise<void> {
    const counts = new Map<string, number>();
    for (let page = 1; page <= 4; page++) {
      const res = await api.notes({ page, pageSize: 50 }).catch(() => null);
      if (!res) break;
      res.items.forEach((n) => {
        const k = n.mood.trim();
        if (k) counts.set(k, (counts.get(k) ?? 0) + 1);
      });
      if (page * 50 >= res.total) break;
    }
    moods.value = [...counts].map(([name, count]) => ({ name, count })).sort((a, b) => b.count - a.count);
  }

  return { days, moods, loadMonth, loadMoods };
}
