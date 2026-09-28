<script setup lang="ts">
import { useI18n } from 'vue-i18n';

/** 身份区 / 格言模块编辑器顶部提示：内容来自站点身份，一键跳到「身份」页 */
defineProps<{ text: string }>();
const { t } = useI18n();
</script>

<template>
  <div class="idn">
    <p>{{ text }}</p>
    <router-link :to="{ name: 'admin-identity' }" class="idn-go">{{ t('aboutKit.ed.goIdentity') }}</router-link>
  </div>
</template>

<style scoped lang="scss">
.idn {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 12px 14px;
  border-radius: var(--r-md);
  background: var(--tint, var(--fill));
  box-shadow: inset 0 0 0 1px var(--line);

  p { margin: 0; font-size: 13px; line-height: 1.6; color: var(--text-2); }
}

.idn-go {
  flex: none;
  padding: 6px 12px;
  border-radius: var(--r-sm);
  font-size: 13px;
  font-weight: 500;
  color: var(--on-solid);
  background: var(--solid);
  transition: background var(--dur-fast);

  &:hover { background: var(--solid-hover, var(--solid)); }
}
</style>
