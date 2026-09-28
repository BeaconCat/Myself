<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import { MOTTO_FLOURISHES, type MottoData } from '../../../about/types';
import { useModuleData } from './useModuleData';

/** 格言：一句话 + 署名小字 + 收尾装饰（印章时可填印文） */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<MottoData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <div class="grid3">
      <label class="span2"><span>{{ t('aboutKit.ed.motto') }}</span><input v-model="d.text" class="a-input" type="text" /></label>
      <label><span>{{ t('aboutKit.ed.sign') }}</span><input v-model="d.sign" class="a-input" type="text" placeholder="NAME · SINCE 2026" /></label>
    </div>
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
    <label v-if="d.flourish === 'seal'" class="fld">
      <span>{{ t('aboutKit.ed.seal') }}</span>
      <input v-model="d.seal" class="a-input" style="max-width: 260px" type="text" maxlength="4" :placeholder="t('aboutKit.ed.sealHint')" />
    </label>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
