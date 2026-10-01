import { computed, onBeforeUnmount, watch } from 'vue';

/** Refresh management lists shortly after a due time, without polling idle lists. */
export function usePublicationRefresh(items: () => { status?: string; publishAt?: string }[], reload: () => Promise<unknown>) {
  const next = computed(() => Math.min(...items().filter(item => item.status === 'scheduled').map(item => Date.parse(item.publishAt ?? '')).filter(Number.isFinite)));
  let timer = 0, stopped = false;
  function plan(): void {
    window.clearTimeout(timer);
    if (stopped || !Number.isFinite(next.value)) return;
    timer = window.setTimeout(async () => {
      if (!document.hidden && Date.now() >= next.value) await reload().catch(() => undefined);
      plan();
    }, Math.max(1500, Math.min(60_000, next.value - Date.now() + 1500)));
  }
  watch(next, plan, { immediate: true });
  onBeforeUnmount(() => { stopped = true; window.clearTimeout(timer); });
}
