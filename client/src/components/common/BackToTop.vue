<script setup lang="ts">
import Icon from '../../components/ui/Icon.vue';
import { ArrowUp } from 'lucide';
import { onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

/** 回到顶部：滚过一屏后在右下角浮现，点击平滑回顶（尊重减弱动效） */
const { t } = useI18n();
const show = ref(false);

function onScroll(): void {
  show.value = window.scrollY > window.innerHeight * 0.9;
}

function toTop(): void {
  const reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  window.scrollTo({ top: 0, behavior: reduce ? 'auto' : 'smooth' });
}

onMounted(() => {
  window.addEventListener('scroll', onScroll, { passive: true });
  onScroll();
});
onBeforeUnmount(() => window.removeEventListener('scroll', onScroll));
</script>

<template>
  <Transition name="btt">
    <button v-if="show" type="button" class="btt" :aria-label="t('noteDetail.toTop')" :title="t('noteDetail.toTop')" @click="toTop">
      <Icon :icon="ArrowUp" :size="20" />
    </button>
  </Transition>
</template>

<style scoped lang="scss">
/* 与顶栏同一中性玻璃面 */
.btt {
  position: fixed;
  right: 28px;
  bottom: 28px;
  z-index: 40;
  width: 46px;
  height: 46px;
  display: grid;
  place-items: center;
  border: 0;
  border-radius: 50%;
  color: var(--text);
  background: color-mix(in oklab, var(--bg) 97%, transparent);
  backdrop-filter: blur(20px) saturate(170%);
  -webkit-backdrop-filter: blur(20px) saturate(170%);
  box-shadow: inset 0 0 0 0.5px var(--line-2), 0 10px 26px -14px rgb(0 0 0 / 0.45);
  cursor: pointer;
  transition: transform var(--dur-fast) var(--ease-spring), background var(--dur-fast);

  &:hover { transform: translateY(-3px); }
  &:active { transform: scale(0.94); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }

  svg { transition: transform var(--dur) var(--ease-spring); }
  &:hover svg { transform: translateY(-2px); }
}

.btt-enter-active, .btt-leave-active { transition: opacity var(--dur) var(--ease-out), transform var(--dur) var(--ease-spring); }
.btt-enter-from, .btt-leave-to { opacity: 0; transform: translateY(12px) scale(0.9); }
</style>
