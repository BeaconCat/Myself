<script setup lang="ts">
import type { AboutModule } from '../../stores/config';

/** 联系我 CTA */
defineProps<{ mod: AboutModule }>();
</script>

<template>
  <div class="contact card-box">
    <div class="contact-text">
      <h2>{{ mod.data.title }}</h2>
      <p>{{ mod.data.text }}</p>
    </div>
    <a class="contact-btn" :href="mod.data.url" target="_blank" rel="noopener">
      {{ mod.data.buttonText }}
    </a>
  </div>
</template>

<style scoped lang="scss">
@use './shared';

.contact {
  display: flex;
  align-items: center;
  gap: 22px;
  background:
    radial-gradient(400px 160px at 15% 0%, rgba(var(--primary-rgb), 0.1), transparent 70%),
    var(--surface);

  .contact-text {
    flex: 1;

    h2 { font-size: 21px; margin-bottom: 8px; }
    p { font-size: 14px; color: var(--text-2); line-height: 1.8; }
  }
}

.contact-btn {
  flex-shrink: 0;
  padding: 12px 28px;
  font-size: 15px;
  font-weight: 700;
  color: #fff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.35);
  background: linear-gradient(180deg, var(--primary), var(--primary-deep));
  box-shadow: 0 4px 14px rgba(var(--primary-rgb), 0.45);
  transition: transform var(--dur-fast) var(--ease-out), filter var(--dur-fast);

  &:hover { filter: brightness(1.08); transform: scale(1.05); }
}

@media (max-width: 720px) {
  .contact { flex-direction: column; text-align: center; }
}
</style>
