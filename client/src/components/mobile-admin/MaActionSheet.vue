<script setup lang="ts">
/**
 * 中央 + 的 action sheet：卡片从 + 键处弹性放大升起，三行依次上浮。
 * 「上传资源」行本身是 file input 的 label，保证在用户手势内唤起系统文件选择器。
 */
import { useI18n } from 'vue-i18n';
import MaIcon from './MaIcon.vue';

/** noNote：协作作者不能发随想，隐去该行 */
defineProps<{ open: boolean; noNote?: boolean }>();
const emit = defineEmits<{ close: []; note: []; post: []; files: [files: File[]] }>();
const { t } = useI18n();

function onFiles(e: Event): void {
  const input = e.target as HTMLInputElement;
  const files = Array.from(input.files ?? []);
  input.value = '';
  if (files.length) emit('files', files);
}
</script>

<template>
  <div class="ma-as" :class="{ open }" :aria-hidden="!open">
    <div class="as-card" role="menu">
      <button v-if="!noNote" class="as-row" role="menuitem" @click="emit('note')">
        <span class="as-ic" style="--c: var(--solid); color: var(--on-solid)"><MaIcon name="bubble" /></span>
        <div class="as-t">
          <b>{{ t('mobileAdmin.create.note') }}</b>
          <small>{{ t('mobileAdmin.create.noteSub') }}</small>
        </div>
        <MaIcon name="chev" :size="16" class="chev" />
      </button>
      <button class="as-row" role="menuitem" @click="emit('post')">
        <span class="as-ic" style="--c: #ff7a1a"><MaIcon name="pen" /></span>
        <div class="as-t">
          <b>{{ t('mobileAdmin.create.post') }}</b>
          <small>{{ t('mobileAdmin.create.postSub') }}</small>
        </div>
        <MaIcon name="chev" :size="16" class="chev" />
      </button>
      <label class="as-row" role="menuitem">
        <span class="as-ic" style="--c: #12b76a"><MaIcon name="upload" /></span>
        <div class="as-t">
          <b>{{ t('mobileAdmin.create.upload') }}</b>
          <small>{{ t('mobileAdmin.create.uploadSub') }}</small>
        </div>
        <MaIcon name="chev" :size="16" class="chev" />
        <input type="file" multiple hidden @change="onFiles" />
      </label>
    </div>
  </div>
</template>

<style scoped lang="scss">
.ma-as {
  position: absolute;
  z-index: 45;
  left: 16px;
  right: 16px;
  bottom: calc(var(--tab-h) + var(--safe-b) + 22px);
  visibility: hidden;
  pointer-events: none;
  transition: visibility 0s linear 0.5s;

  &.open {
    visibility: visible;
    pointer-events: auto;
    transition: none;
  }
}

.as-card {
  border-radius: var(--r-xl);
  padding: 8px;
  background: var(--glass-2);
  backdrop-filter: blur(30px) saturate(180%);
  -webkit-backdrop-filter: blur(30px) saturate(180%);
  box-shadow: inset 0 0 0 0.5px var(--glass-line), var(--shadow-pop);
  transform-origin: 50% 125%;
  opacity: 0;
  transform: translateY(40px) scale(0.55);
  transition: transform 0.5s var(--ease-spring), opacity 0.25s var(--ease-out);

  .open & {
    opacity: 1;
    transform: none;
  }
}

.as-row {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 12px;
  border-radius: calc(var(--r-xl) - 8px);
  text-align: left;
  cursor: pointer;
  opacity: 0;
  transform: translateY(14px);
  transition: opacity 0.3s, transform 0.45s var(--ease-spring), background var(--dur-fast);

  .open & {
    opacity: 1;
    transform: none;
  }

  .open &:nth-child(1) { transition-delay: 0.06s; }
  .open &:nth-child(2) { transition-delay: 0.11s; }
  .open &:nth-child(3) { transition-delay: 0.16s; }

  &:active {
    background: var(--fill-2);
    transform: scale(0.98);
    transition-delay: 0s;
  }

  .chev {
    margin-left: auto;
    color: var(--text-3);
  }
}

.as-ic {
  width: 46px;
  height: 46px;
  border-radius: var(--r-md);
  display: grid;
  place-items: center;
  color: #fff;
  background: var(--c);
  box-shadow: inset 0 0 0 0.5px rgb(0 0 0 / 0.1);
  flex: none;
}

.as-t {
  min-width: 0;

  b {
    display: block;
    font-size: 16px;
    font-weight: 600;
  }

  small {
    font-size: 12.5px;
    color: var(--text-3);
  }
}
</style>
