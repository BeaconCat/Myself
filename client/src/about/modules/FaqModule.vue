<script setup lang="ts">
import type { AboutModule } from '../../stores/config';

/** 问答 FAQ：展开态由分发器统一持有（全页同一时刻仅展开一条） */
defineProps<{ mod: AboutModule }>();
const open = defineModel<string>('open', { required: true });
</script>

<template>
  <h2 class="block-title">问答</h2>
  <div class="faq">
    <div
      v-for="(f, i) in mod.data.items"
      :key="i"
      class="faq-item card-box"
      :class="{ open: open === `${mod.id}-${i}` }"
    >
      <button class="faq-q" @click="open = open === `${mod.id}-${i}` ? '' : `${mod.id}-${i}`">
        <span>{{ f.q }}</span>
        <i aria-hidden="true">+</i>
      </button>
      <p class="faq-a">{{ f.a }}</p>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use './shared';

.faq { display: flex; flex-direction: column; gap: 10px; }

.faq-item {
  padding: 0;
  overflow: hidden;

  .faq-q {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    padding: 15px 20px;
    border: none;
    background: none;
    color: var(--text);
    font-size: 14.5px;
    font-weight: 600;
    text-align: left;

    i {
      font-style: normal;
      font-size: 18px;
      color: var(--primary);
      transition: transform var(--dur-fast) var(--ease-spring);
    }
  }

  .faq-a {
    max-height: 0;
    overflow: hidden;
    padding: 0 20px;
    font-size: 14px;
    line-height: 1.8;
    color: var(--text-2);
    transition: max-height var(--dur) var(--ease-out), padding var(--dur) ease;
  }

  &.open {
    .faq-q i { transform: rotate(45deg); }
    .faq-a { max-height: 300px; padding: 0 20px 16px; }
  }
}
</style>
