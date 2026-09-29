<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { EngageTarget, ReactionKind } from '../../api';
import { useConfigStore } from '../../stores/config';
import { useEngageStore } from '../../stores/engage';
import EngageIcon from './EngageIcon.vue';

/**
 * 互动栏：喜欢 · 评论数 · 回应（灵感 / 会心 / 共鸣，点开小浮层挑选；已有的回应以小胶囊显示计数）· 分享 · 复制正文。
 * 回应按开关 users.reactions 显示；评论数按评论开关显示，点击交给父级（通常是进入详情页）。
 * 分享：触屏且支持系统分享时唤起分享面板，否则复制链接。
 */
const props = defineProps<{ target: EngageTarget; id: number; link: string; text?: string; big?: boolean }>();
const emit = defineEmits<{ comment: [] }>();
const { t } = useI18n();
const config = useConfigStore();
const engage = useEngageStore();

onMounted(() => engage.want(props.target, props.id));

const sum = computed(() => engage.get(props.target, props.id));
const reactionsOn = computed(() => config.cfg.users?.reactions !== false);
const commentsOn = computed(() => !!config.cfg.users?.comments?.enabled);
const EXTRA: ReactionKind[] = ['spark', 'smile', 'resonate'];
const chips = computed(() => EXTRA.filter((k) => (sum.value.reactions[k] ?? 0) > 0));
const mine = (k: ReactionKind) => sum.value.mine.includes(k);

/** 点击后图标弹一下 */
const popped = ref<string>('');
async function toggle(k: ReactionKind): Promise<void> {
  popped.value = k;
  window.setTimeout(() => { if (popped.value === k) popped.value = ''; }, 420);
  picker.value = false;
  try {
    await engage.toggle(props.target, props.id, k);
  } catch { /* 失败已回滚 */ }
}

/* 回应浮层 */
const picker = ref(false);
const root = ref<HTMLElement | null>(null);
function outside(e: MouseEvent): void {
  if (picker.value && !root.value?.contains(e.target as Node)) picker.value = false;
}
onMounted(() => document.addEventListener('mousedown', outside));
onBeforeUnmount(() => document.removeEventListener('mousedown', outside));

/* 分享 / 复制：短暂显示对勾 */
const done = ref<'' | 'share' | 'copy'>('');
function flash(kind: 'share' | 'copy'): void {
  done.value = kind;
  window.setTimeout(() => { if (done.value === kind) done.value = ''; }, 1600);
}
async function share(): Promise<void> {
  const url = new URL(props.link, window.location.origin).href;
  const touch = window.matchMedia('(hover: none)').matches;
  if (touch && navigator.share) {
    try {
      await navigator.share({ url, text: props.text?.slice(0, 80) });
      return;
    } catch { /* 用户取消时退回复制 */ }
  }
  await navigator.clipboard?.writeText(url).catch(() => undefined);
  flash('share');
}
async function copyText(): Promise<void> {
  await navigator.clipboard?.writeText(props.text ?? '').catch(() => undefined);
  flash('copy');
}

const n = (v: number | undefined) => (v && v > 0 ? String(v) : '');
</script>

<template>
  <div ref="root" class="eg" :class="{ big }" @click.stop>
    <button
      v-if="reactionsOn"
      type="button"
      class="b like"
      :class="{ on: mine('like'), pop: popped === 'like' }"
      :aria-pressed="mine('like')"
      :title="t('engage.like')"
      @click="toggle('like')"
    >
      <EngageIcon name="like" :filled="mine('like')" /><span>{{ n(sum.reactions.like) }}</span>
    </button>

    <button v-if="commentsOn" type="button" class="b" :title="t('engage.comment')" @click="emit('comment')">
      <EngageIcon name="comment" /><span>{{ n(sum.comments) }}</span>
    </button>

    <template v-if="reactionsOn">
      <button
        v-for="k in chips"
        :key="k"
        type="button"
        class="chip"
        :class="{ on: mine(k), pop: popped === k }"
        :title="t(`engage.${k}`)"
        @click="toggle(k)"
      >
        <EngageIcon :name="k" :size="15" /><span>{{ sum.reactions[k] }}</span>
      </button>
      <span class="pk">
        <button type="button" class="b" :class="{ on: picker }" :title="t('engage.react')" :aria-expanded="picker" @click="picker = !picker">
          <EngageIcon name="react" />
        </button>
        <Transition name="pk">
          <span v-if="picker" class="pk-pop" role="menu">
            <button v-for="k in EXTRA" :key="k" type="button" role="menuitem" :class="{ on: mine(k) }" @click="toggle(k)">
              <EngageIcon :name="k" :size="20" /><small>{{ t(`engage.${k}`) }}</small>
            </button>
          </span>
        </Transition>
      </span>
    </template>

    <span class="grow" />
    <button type="button" class="b" :title="t('engage.share')" @click="share">
      <EngageIcon :name="done === 'share' ? 'check' : 'share'" />
      <Transition name="tip"><em v-if="done === 'share'">{{ t('engage.linkCopied') }}</em></Transition>
    </button>
    <button v-if="text" type="button" class="b" :title="t('engage.copy')" @click="copyText">
      <EngageIcon :name="done === 'copy' ? 'check' : 'copy'" />
      <Transition name="tip"><em v-if="done === 'copy'">{{ t('engage.textCopied') }}</em></Transition>
    </button>
  </div>
</template>

<style scoped lang="scss">
.eg {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-left: -8px;
  color: var(--text-3);
}

.grow { flex: 1; }

.b {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 32px;
  min-width: 32px;
  padding: 0 8px;
  border: 0;
  border-radius: var(--r-pill);
  background: none;
  color: inherit;
  font: 500 13px var(--font-mono);
  font-variant-numeric: tabular-nums;
  cursor: pointer;
  transition: color var(--dur-fast), background var(--dur-fast);

  &:hover { color: var(--text); background: var(--fill); }
  &.on { color: var(--ink); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }

  span:empty { display: none; }

  em {
    position: absolute;
    right: 0;
    bottom: calc(100% + 6px);
    padding: 4px 9px;
    border-radius: var(--r-sm);
    background: var(--text);
    color: var(--bg);
    font: 500 12px var(--font-sans);
    font-style: normal;
    white-space: nowrap;
    pointer-events: none;
  }
}

/* 喜欢：点亮用品牌红 */
.like.on { color: var(--accent-red); }

.pop :deep(.eg-i) { animation: eg-pop 0.42s var(--ease-spring); }

@keyframes eg-pop {
  0% { transform: scale(1); }
  35% { transform: scale(1.35); }
  100% { transform: scale(1); }
}

.chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 26px;
  padding: 0 9px 0 7px;
  border: 0;
  border-radius: var(--r-pill);
  background: none;
  box-shadow: inset 0 0 0 1px var(--line-2);
  color: var(--text-2);
  font: 500 12.5px var(--font-mono);
  cursor: pointer;
  transition: background var(--dur-fast), color var(--dur-fast), box-shadow var(--dur-fast);

  &:hover { background: var(--fill); }
  &.on { color: var(--ink); background: color-mix(in oklab, var(--ink) 8%, transparent); box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--ink) 40%, transparent); }
}

/* 回应浮层 */
.pk { position: relative; display: inline-flex; }

.pk-pop {
  position: absolute;
  left: 50%;
  bottom: calc(100% + 8px);
  z-index: 5;
  display: flex;
  gap: 2px;
  padding: 5px;
  translate: -50% 0;
  border-radius: var(--r-lg);
  background: var(--surface, var(--bg));
  box-shadow: 0 0 0 1px var(--line-2), 0 12px 28px -12px rgb(0 0 0 / 0.35);

  button {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 3px;
    width: 52px;
    padding: 7px 0 5px;
    border: 0;
    border-radius: var(--r-md);
    background: none;
    color: var(--text-2);
    cursor: pointer;
    transition: background var(--dur-fast), color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

    small { font-size: 11px; color: var(--text-3); }
    &:hover { background: var(--fill); color: var(--text); transform: translateY(-2px); }
    &.on { color: var(--ink); }
  }
}

.pk-enter-active, .pk-leave-active { transition: opacity var(--dur-fast) var(--ease-out), transform var(--dur-fast) var(--ease-spring); }
.pk-enter-from, .pk-leave-to { opacity: 0; transform: translateY(6px) scale(0.92); }
.tip-enter-active, .tip-leave-active { transition: opacity var(--dur-fast), transform var(--dur-fast) var(--ease-out); }
.tip-enter-from, .tip-leave-to { opacity: 0; transform: translateY(4px); }

.big .b { height: 36px; font-size: 14px; }

@media (prefers-reduced-motion: reduce) {
  .pop :deep(.eg-i) { animation: none; }
}
</style>
