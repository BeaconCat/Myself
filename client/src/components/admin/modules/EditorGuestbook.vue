<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { GuestbookData } from '../../../about/types';
import { useModuleData } from './useModuleData';

/**
 * 留言墙：只有展示选项（首屏显示条数）。留言是真实评论：审核与回复在后台「评论」页，
 * 谁能留言（登录 / 匿名）在「用户」页设置。
 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<GuestbookData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <div class="line">
      <label class="line inl"><span>{{ t('aboutKit.ed.pageSize') }}</span><input v-model.number="d.pageSize" class="a-input w-num" type="number" min="2" max="12" /></label>
    </div>
    <p class="hint">{{ t('aboutKit.ed.guestbookHint') }}</p>
    <router-link class="go" :to="{ name: 'admin-comments' }">{{ t('aboutKit.ed.guestbookGo') }}</router-link>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

label.inl { flex-direction: row; align-items: center; margin-right: 8px; }
.go { font-size: 13px; color: var(--ink); }
</style>
