<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { ProfileData } from '../../../about/types';
import IdentityNote from './IdentityNote.vue';
import { useModuleData } from './useModuleData';

/** 身份区：内容（名字 / 签名 / 自述 / 状态 / 链接 / 形象图）来自站点身份；这里只设顶部小字 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<ProfileData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <IdentityNote :text="t('aboutKit.ed.profileFromIdentity')" />
    <label><span>{{ t('aboutKit.ed.kicker') }}</span><input v-model="d.kicker" class="a-input" type="text" :placeholder="t('aboutKit.ed.kickerHint')" /></label>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
