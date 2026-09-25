<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import StSeg from './studio/StSeg.vue';

/**
 * 评论审核（P4 预留）：界面完整，数据为示例；评论系统上线后接入真实接口。
 * 所有操作按钮禁用并以 tooltip 说明。
 */
const { t } = useI18n();

type Tab = 'pending' | 'approved' | 'spam';
const tab = ref<Tab>('pending');

interface Sample { n: string; c: string; post: string; t: string; text: string; flag?: string }
const PENDING: Sample[] = [
  { n: '林间', c: '#3b7d5a', post: '主题系统指南', t: '12 分钟前', text: '色盘从一个主色推导整套变量这个思路太好了，我试了下自定义一个青色，深色模式下对比度居然也是对的。想问下 primary-ink 是怎么算的？' },
  { n: '阿澈', c: '#b5651d', post: '用 Markdown 写作', t: '1 小时前', text: '代码块能不能加一个一键复制？手机上长按选择很痛苦。' },
  { n: 'Moss', c: '#556b8d', post: '把发文托管给 AI：API 中心', t: '3 小时前', text: '已经用 API 接上了自己的周报脚本，每周五自动发一篇草稿，省了好多事。', flag: 'link' },
  { n: '小满', c: '#9a4f7a', post: '欢迎使用 Myself', t: '昨天', text: '页面好安静，读起来很舒服。宋体标题配黑体正文的组合很少见，但一点也不违和。' },
  { n: 'Echo', c: '#6d6a3a', post: '主题系统指南', t: '昨天', text: '秋天的黄色在浅色背景上会不会太亮？我这边看按钮文字有点吃力。' },
];
const APPROVED: Sample[] = [
  { n: 'Moss', c: '#556b8d', post: '欢迎使用 Myself', t: '3 天前', text: '终于有一个不臃肿的博客引擎了。' },
  { n: '林间', c: '#3b7d5a', post: '用 Markdown 写作', t: '5 天前', text: '目录锚点的滚动高亮做得很细腻。' },
];

const list = computed(() => (tab.value === 'pending' ? PENDING : tab.value === 'approved' ? APPROVED : []));
const tip = computed(() => t('studio.reserved.commentsTip'));
</script>

<template>
  <section class="studio view">
    <div class="st-vh">
      <div>
        <h1>{{ t('studio.comments.title') }}</h1>
        <p>{{ t('studio.comments.desc') }}</p>
      </div>
      <div class="act">
        <StSeg
          v-model="tab"
          :options="[
            { value: 'pending', label: t('studio.comments.pending'), count: PENDING.length },
            { value: 'approved', label: t('studio.comments.approved') },
            { value: 'spam', label: t('studio.comments.spam') },
          ]"
        />
      </div>
    </div>

    <div class="st-note-bar st-rise"><SIcon name="info" />{{ t('studio.reserved.comments') }}</div>

    <div v-if="tab === 'pending'" class="bulk">
      <label class="st-ckrow"><span class="st-ck"><svg viewBox="0 0 24 24"><path d="M5 12.5l4.5 4.5L19 7" /></svg></span>{{ t('studio.comments.selectAll') }}</label>
      <span class="sp" />
      <button type="button" class="st-btn q sm st-tip" :data-tip="tip" disabled><SIcon name="check" :size="18" />{{ t('studio.comments.approveAll') }}</button>
    </div>

    <div v-if="!list.length" class="st-empty">
      <svg class="clean" viewBox="0 0 96 96" width="96" height="96" fill="none" aria-hidden="true">
        <path d="M22 30h52a4 4 0 0 1 4 4v28a4 4 0 0 1-4 4H42l-12 10V66h-8a4 4 0 0 1-4-4V34a4 4 0 0 1 4-4z" stroke="var(--st-ink-4)" stroke-width="1.5" />
        <path d="M38 48l7 7 13-14" stroke="var(--ink)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
      <h4>{{ t('studio.comments.cleanTitle') }}</h4>
      <p>{{ t('studio.comments.cleanSub') }}</p>
    </div>

    <div class="cm-list">
      <div v-for="(c, i) in list" :key="`${tab}-${i}`" class="cm st-rise" :style="{ '--i': i }">
        <span class="st-ck"><svg viewBox="0 0 24 24"><path d="M5 12.5l4.5 4.5L19 7" /></svg></span>
        <span class="av" :style="{ background: c.c }">{{ c.n[0] }}</span>
        <div class="body">
          <div class="who">
            <b>{{ c.n }}</b><span>{{ t('studio.comments.on') }}</span><a>《{{ c.post }}》</a><span>· {{ c.t }}</span>
            <span v-if="c.flag" class="flag">{{ t('studio.comments.hasLink') }}</span>
          </div>
          <p>{{ c.text }}</p>
        </div>
        <div class="acts">
          <template v-if="tab === 'pending'">
            <button type="button" class="st-btn p sm st-tip" :data-tip="tip" disabled><SIcon name="check" :size="18" />{{ t('studio.comments.approve') }}</button>
            <button type="button" class="st-ibtn st-tip" :data-tip="tip" disabled><SIcon name="reply" :size="18" /></button>
            <button type="button" class="st-ibtn st-tip" :data-tip="tip" disabled><SIcon name="flag" :size="18" /></button>
          </template>
          <button v-else type="button" class="st-ibtn st-tip" :data-tip="tip" disabled><SIcon name="reply" :size="18" /></button>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped lang="scss">
.view {
  max-width: 1280px;
  margin: 0 auto;
  padding: 32px 48px 72px;
}

.bulk {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  margin-bottom: 4px;
  font-size: 13.5px;
  color: var(--st-ink-3);

  .st-ckrow { font-size: 14px; }
  .sp { flex: 1; }
}

.cm-list { display: flex; flex-direction: column; }

.cm {
  display: grid;
  grid-template-columns: 18px 40px minmax(0, 1fr) auto;
  gap: 14px;
  padding: 14px 12px;
  border-bottom: 1px solid var(--line);
  align-items: start;
  transition: background var(--dur-fast);

  &:hover { background: linear-gradient(90deg, transparent, var(--well) 12%, var(--well) 88%, transparent); }

  > .st-ck { margin-top: 11px; }

  .av {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    color: #fff;
    font: 600 15px var(--font-serif);
  }

  .body { min-width: 0; }

  .who {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    font-size: 13.5px;
    color: var(--st-ink-3);
    margin-bottom: 4px;

    b { color: var(--st-ink); font-weight: 600; font-size: 15px; }
    a { color: var(--st-ink-2); font-family: var(--font-serif); }
  }

  p { margin: 0; font-size: 15px; line-height: 1.7; color: var(--st-ink); }

  .flag {
    font-size: 12.5px;
    color: color-mix(in oklab, var(--yellow) 60%, var(--st-ink));
    background: color-mix(in oklab, var(--yellow) 14%, var(--paper));
    padding: 1px 7px;
    border-radius: var(--r-xs);
  }

  .acts { display: flex; gap: 6px; opacity: 0.55; transition: opacity var(--dur-fast); }
  &:hover .acts { opacity: 1; }
}

.st-empty .clean { margin-bottom: 0; }

@media (max-width: 1180px) {
  .view { padding: 28px 32px 64px; }
}
</style>
