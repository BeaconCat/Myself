<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../api';
import { useConfigStore, type SiteConfig } from '../../stores/config';
import { useThemeStore } from '../../stores/theme';
import SectionSite from './settings/SectionSite.vue';
import SectionLoading from './settings/SectionLoading.vue';
import SectionTheme from './settings/SectionTheme.vue';
import SectionHero from './settings/SectionHero.vue';
import SectionTimezone from './settings/SectionTimezone.vue';
import SectionGithub from './settings/SectionGithub.vue';
import SectionUsers from './settings/SectionUsers.vue';

/** 设置页外壳：加载 / 保存 cfg，各分区组件直接编辑同一 reactive 对象 */
const { t } = useI18n();
const configStore = useConfigStore();

const cfg = reactive<SiteConfig>(JSON.parse(JSON.stringify(configStore.cfg)));
const message = ref('');
const busy = ref(false);

/** 锚点子导航 */
const SECTIONS = [
  { id: 'sec-platform', key: 'admin.secPlatform' },
  { id: 'sec-loading', key: 'admin.secLoading' },
  { id: 'sec-theme', key: 'admin.secTheme' },
  { id: 'sec-hero', key: 'admin.secHero' },
  { id: 'sec-timezone', key: 'admin.secTimezone' },
  { id: 'sec-github', key: 'admin.secGithub' },
  { id: 'sec-users', key: 'admin.secUsers' },
];

async function load(): Promise<void> {
  const remote = await adminApi.settings() as unknown as SiteConfig;
  Object.assign(cfg, JSON.parse(JSON.stringify(remote)));
}

async function save(): Promise<void> {
  if (busy.value) return;
  busy.value = true;
  message.value = '';
  try {
    cfg.theme.displayCount = Math.min(4, Math.max(1, cfg.theme.displayCount));
    await adminApi.saveSettings(cfg as unknown as Record<string, unknown>);
    await configStore.load();
    useThemeStore().init();
    message.value = t('admin.saved');
  } catch {
    message.value = t('admin.saveFailed');
  } finally {
    busy.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="settings">
    <header class="head">
      <h1 class="page-h">{{ t('admin.menuSettings') }}</h1>
      <div class="actions">
        <span v-if="message" class="msg">{{ message }}</span>
        <button class="btn primary" :disabled="busy" @click="save">{{ t('admin.saveAll') }}</button>
      </div>
    </header>

    <!-- 锚点子导航 -->
    <nav class="subnav">
      <a v-for="s in SECTIONS" :key="s.id" :href="`#${s.id}`">{{ t(s.key) }}</a>
    </nav>

    <SectionSite :cfg="cfg" />
    <SectionLoading :cfg="cfg" />
    <SectionTheme :cfg="cfg" />
    <SectionHero :cfg="cfg" />
    <SectionTimezone :cfg="cfg" />
    <SectionGithub :cfg="cfg" />
    <SectionUsers />
  </div>
</template>

<style scoped lang="scss">
.settings {
  max-width: 860px;
}

.head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 22px;
}

.page-h { font-size: 26px; }

.actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.msg {
  font-size: 13px;
  color: var(--primary);
}

/* 锚点子导航 */
.subnav {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 10px 12px;
  margin-bottom: 16px;
  border-radius: var(--radius);
  background: var(--glass);
  backdrop-filter: blur(14px) saturate(1.4);
  -webkit-backdrop-filter: blur(14px) saturate(1.4);
  border: 1px solid var(--border);

  a {
    font-size: 12.5px;
    font-weight: 600;
    padding: 6px 14px;
    border-radius: 999px;
    color: var(--text-2);
    transition: all var(--dur-fast);

    &:hover {
      background: rgba(var(--primary-rgb), 0.1);
      color: var(--primary);
    }
  }
}

/* 头部保存按钮（分区内按钮样式见 settings-shared.scss） */
.btn {
  padding: 10px 24px;
  border: 1px solid transparent;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 700;
  transition: all var(--dur-fast) var(--ease-out);

  &.primary {
    color: #fff;
    text-shadow: 0 1px 3px rgba(0, 0, 0, 0.35);
    background: linear-gradient(180deg, var(--primary), var(--primary-deep));
    box-shadow: 0 4px 14px rgba(var(--primary-rgb), 0.4);

    &:hover:not(:disabled) { filter: brightness(1.08); }
    &:disabled { opacity: 0.55; }
  }
}
</style>
