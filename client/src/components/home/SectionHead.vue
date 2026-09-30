<script setup lang="ts">
import Icon from '../../components/ui/Icon.vue';
import { ArrowRight } from 'lucide';
/** 首页分区标题：小字副标题 + 宋体标题；右侧额外操作（slot）+ 文字链接 */
defineProps<{ title: string; sub?: string; loading?: boolean; to?: string; linkText?: string }>();
</script>

<template>
  <header class="sec-h">
    <div class="ttl">
      <!-- 副标题（统计）放在 h2 外：标题的可访问名称只是标题本身 -->
      <small v-if="loading" class="sk sk-line sub-sk" />
      <small v-else-if="sub">{{ sub }}</small>
      <h2>{{ title }}</h2>
    </div>
    <div class="acts">
      <slot />
      <router-link v-if="to && linkText" :to="to" class="link">
        <span>{{ linkText }}</span>
        <Icon :icon="ArrowRight" :stroke="2" />
      </router-link>
    </div>
  </header>
</template>

<style scoped lang="scss">
.sec-h {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: var(--section-head-gap);

  h2 {
    font-family: var(--font-serif);
    font-size: 28px;
    font-weight: 700;
    line-height: 1.3;
  }

  small {
    display: block;
    margin-bottom: 4px;
    font-family: var(--font-sans);
    font-size: 13px;
    font-weight: 400;
    letter-spacing: 0.04em;
    color: var(--text-3);
  }

  .sub-sk {
    width: 160px;
    height: 12px;
    margin-bottom: 10px;
  }
}

.acts {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: none;
}

/* 文字链接：--ink 信号色，悬停下划线从左长出 */
.link {
  margin-left: 6px;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 2px 0;
  border-radius: var(--r-xs);
  font-size: 15px;
  font-weight: 500;
  color: var(--ink);

  span {
    padding-bottom: 2px;
    background: linear-gradient(currentColor, currentColor) 0 100% / 0 1px no-repeat;
    transition: background-size var(--dur) var(--ease-out);
  }

  svg {
    width: 18px;
    height: 18px;
    transition: transform var(--dur) var(--ease-spring);
  }

  &:hover span { background-size: 100% 1px; }
  &:hover svg { transform: translateX(3px); }
  &:active { transform: scale(0.97); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

@media (max-width: 768px) {
  .sec-h h2 { font-size: 24px; }
}
</style>
