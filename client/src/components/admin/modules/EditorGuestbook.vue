<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { GuestbookData } from '../../../about/types';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 留言墙：每页条数 / 是否需要登录；精选留言（姓名 / 颜色 / 时间 / 内容 / 喜欢数 / 站长回复） */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<GuestbookData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <div class="line">
      <label class="line inl"><span>{{ t('aboutKit.ed.pageSize') }}</span><input v-model.number="d.pageSize" class="a-input w-num" type="number" min="2" max="12" /></label>
      <label class="line inl"><span>{{ t('aboutKit.ed.total') }}</span><input v-model.number="d.total" class="a-input w-num" type="number" min="0" /></label>
      <label class="check"><input v-model="d.requireLogin" type="checkbox" />{{ t('aboutKit.ed.requireLogin') }}</label>
    </div>
    <p class="hint">{{ t('aboutKit.ed.guestbookHint') }}</p>
    <EdList v-slot="{ item }" :items="d.items" :make="() => ({ name: '', color: '#0078ff', at: '', text: '', likes: 0, reply: '' })">
      <div class="line">
        <input v-model="item.color" class="color-in" type="color" />
        <input v-model="item.name" class="a-input w-narrow" type="text" :placeholder="t('aboutKit.ed.guestName')" />
        <input v-model="item.at" class="a-input w-narrow" type="text" :placeholder="t('aboutKit.ed.when')" />
        <input v-model.number="item.likes" class="a-input w-num" type="number" min="0" />
      </div>
      <input v-model="item.text" class="a-input" type="text" :placeholder="t('aboutKit.ed.text')" />
      <input v-model="item.reply" class="a-input" type="text" :placeholder="t('aboutKit.ed.reply')" />
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

label.inl { flex-direction: row; align-items: center; margin-right: 8px; }
</style>
