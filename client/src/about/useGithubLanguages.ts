import { computed, ref, watch, type Ref } from 'vue';
import { adminApi, api, type GithubLanguages } from '../api';
import { useConfigStore } from '../stores/config';

export function useGithubLanguages(enabled: Ref<boolean>) {
  const config = useConfigStore();
  const data = ref<GithubLanguages | null>(null);
  const loading = ref(false);
  const error = ref(false);
  const username = computed(() => config.cfg.github.username.trim());
  let request = 0;
  async function refresh(force = false): Promise<void> {
    const id = ++request;
    if (!enabled.value || !username.value) { loading.value = false; return; }
    loading.value = true;
    error.value = false;
    try {
      const next = await (force ? adminApi.githubLanguagesSync() : api.githubLanguages());
      if (id === request) { data.value = next; error.value = next.stale; }
    } catch { if (id === request) error.value = true; }
    finally { if (id === request) loading.value = false; }
  }
  watch([enabled, username], () => { data.value = null; void refresh(); }, { immediate: true });
  return { data, loading, error, username, refresh };
}
