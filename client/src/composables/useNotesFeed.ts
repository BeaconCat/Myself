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
    const fmt = (d: Date) => d.toISOString().slice(0, 10);
    dateFrom.value = fmt(start);
    dateTo.value = fmt(end);
    void reload();
  }

  function clearDate(): void {
    dateFrom.value = '';
    dateTo.value = '';
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

  async function reload(): Promise<void> {
    loading.value = true;
    try {
      const res = await api.notes(queryParams(1));
      notes.value = res.items;
      total.value = res.total;
      page.value = 1;
    } finally {
      loading.value = false;
    }
  }

  const loadingMore = ref(false);
  const hasMore = computed(() => notes.value.length < total.value);

  async function loadMore(): Promise<void> {
    if (loadingMore.value || loading.value || !hasMore.value) return;
    loadingMore.value = true;
    try {
      const res = await api.notes(queryParams(page.value + 1));
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
    reload,
    switchTab,
    onSearch,
    applyQuickRange,
    clearDate,
  };
}
