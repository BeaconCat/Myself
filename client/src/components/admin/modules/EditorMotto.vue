<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import { MOTTO_FLOURISHES, type MottoData } from '../../../about/types';
import IdentityNote from './IdentityNote.vue';
import { useModuleData } from './useModuleData';

/** 格言：文字与署名来自站点身份；这里只选收尾装饰 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<MottoData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <IdentityNote :text="t('aboutKit.ed.mottoFromIdentity')" />
    <div class="fld">
      <span>{{ t('aboutKit.ed.flourish') }}</span>
      <div class="seg">
        <button
          v-for="f in MOTTO_FLOURISHES"
          :key="f"
          type="button"
          :class="{ on: (d.flourish ?? 'line') === f }"
          @click="d.flourish = f"
        >{{ t(`aboutKit.ed.flourish_${f}`) }}</button>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
