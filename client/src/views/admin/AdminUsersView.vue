<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useConfigStore } from '../../stores/config';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import StSwitch from './studio/StSwitch.vue';

/**
 * 用户（P4 预留）：注册读者与协作作者。界面完整，数据为示例；操作禁用并以 tooltip 说明。
 * 第一行管理员为当前站点身份。
 */
const { t } = useI18n();
const config = useConfigStore();

interface Sample { n: string; e: string; role: 'admin' | 'author' | 'reader'; j: string; a: string; cm: string; c: string; on: boolean }
const USERS = computed<Sample[]>(() => [
  { n: config.cfg.about?.name || 'Myself', e: 'admin', role: 'admin', j: config.cfg.about?.foundedAt || '2026-01-01', a: 'online', cm: '—', c: '#0b1220', on: true },
  { n: '林间', e: 'linjian@mail.cn', role: 'author', j: '2026-03-12', a: '12 分钟前', cm: '38', c: '#3b7d5a', on: true },
  { n: 'Moss', e: 'moss@proton.me', role: 'author', j: '2026-05-02', a: '3 小时前', cm: '21', c: '#556b8d', on: true },
  { n: '阿澈', e: 'ache@qq.com', role: 'reader', j: '2026-07-08', a: '1 小时前', cm: '9', c: '#b5651d', on: true },
  { n: '小满', e: 'xiaoman@163.com', role: 'reader', j: '2026-08-19', a: '昨天', cm: '4', c: '#9a4f7a', on: true },
  { n: 'Echo', e: 'echo@outlook.com', role: 'reader', j: '2026-09-01', a: '昨天', cm: '2', c: '#6d6a3a', on: false },
]);
const tip = computed(() => t('studio.reserved.usersTip'));
const avatar = computed(() => config.cfg.about?.avatar || '/favicon-64.png');
</script>

<template>
  <section class="studio view">
    <div class="st-vh">
      <div>
        <h1>{{ t('studio.users.title') }}</h1>
        <p>{{ t('studio.users.desc') }}</p>
      </div>
      <div class="act">
        <button type="button" class="st-btn g st-tip" :data-tip="tip" disabled><SIcon name="download" :size="16" />{{ t('studio.users.export') }}</button>
        <button type="button" class="st-btn p st-tip" :data-tip="tip" disabled><SIcon name="mail" :size="16" />{{ t('studio.users.invite') }}</button>
      </div>
    </div>

    <div class="st-note-bar st-rise"><SIcon name="info" />{{ t('studio.reserved.users') }}</div>

    <div class="u-stats">
      <div class="st-rise" style="--i: 0"><b class="mono">128</b><small>{{ t('studio.users.sTotal') }}</small><span class="up mono">+14</span></div>
      <div class="st-rise" style="--i: 1"><b class="mono">36</b><small>{{ t('studio.users.sActive') }}</small></div>
      <div class="st-rise" style="--i: 2"><b class="mono">3</b><small>{{ t('studio.users.sAuthors') }}</small></div>
      <div class="st-rise" style="--i: 3"><b class="mono">412</b><small>{{ t('studio.users.sComments') }}</small><span class="up mono">+38</span></div>
    </div>

    <table class="st-table">
      <thead>
        <tr>
          <th>{{ t('studio.users.cUser') }}</th>
          <th>{{ t('studio.users.cRole') }}</th>
          <th>{{ t('studio.users.cJoined') }}</th>
          <th>{{ t('studio.users.cActive') }}</th>
          <th>{{ t('studio.users.cComments') }}</th>
          <th>{{ t('studio.users.cStatus') }}</th>
          <th />
        </tr>
      </thead>
      <tbody>
        <tr v-for="(u, i) in USERS" :key="u.e" class="st-rise" :style="{ '--i': i + 2 }">
          <td>
            <div class="u">
              <span class="av" :style="{ background: u.c }"><img v-if="u.role === 'admin'" :src="avatar" alt="" /><template v-else>{{ u.n[0] }}</template></span>
              <div><b>{{ u.n }}</b><small>{{ u.role === 'admin' ? t('studio.users.you') : u.e }}</small></div>
            </div>
          </td>
          <td><span class="role" :class="u.role">{{ t(`studio.users.r_${u.role}`) }}</span></td>
          <td class="mono">{{ u.j }}</td>
          <td class="act-cell">
            <span v-if="u.a === 'online'" class="online"><i class="st-dot" />{{ t('studio.users.online') }}</span>
            <template v-else>{{ u.a }}</template>
          </td>
          <td class="mono">{{ u.cm }}</td>
          <td>
            <span v-if="u.role === 'admin'" class="dash">—</span>
            <span v-else class="st-tip" :data-tip="tip"><StSwitch :model-value="u.on" disabled /></span>
          </td>
          <td><button type="button" class="st-ibtn st-tip" :data-tip="tip" disabled><SIcon name="more" /></button></td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<style scoped lang="scss">
.view {
  max-width: 1120px;
  margin: 0 auto;
  padding: 52px 64px 96px;
}

.u-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
  margin-bottom: 36px;

  > div { padding: 18px 20px; border-radius: var(--r-md); background: var(--well); }
  b { display: block; font-size: 28px; line-height: 1.2; font-weight: 500; letter-spacing: -0.03em; }
  small { font-size: 12.5px; color: var(--st-ink-3); }
  .up { font-size: 12px; color: color-mix(in oklab, var(--green) 70%, var(--st-ink)); margin-left: 6px; }
}

.st-table {
  .u { display: flex; align-items: center; gap: 12px; }

  .av {
    width: 34px;
    height: 34px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    color: #fff;
    font: 600 13px var(--font-serif);
    overflow: hidden;
    flex: none;

    img { width: 100%; height: 100%; object-fit: cover; }
  }

  b { display: block; font-weight: 500; }
  small { color: var(--st-ink-3); font-size: 12px; }
  .mono { font-size: 12px; color: var(--st-ink-3); }
  .act-cell { font-size: 13px; color: var(--st-ink-2); }
  .online { display: inline-flex; align-items: center; gap: 6px; .st-dot { --c: var(--green); } }
  .dash { font-size: 12px; color: var(--st-ink-3); }
}

.role {
  font-size: 12px;
  padding: 3px 9px;
  border-radius: var(--r-xs);
  background: var(--well-2);
  color: var(--st-ink-2);

  &.admin { background: var(--tint); color: var(--ink); }
  &.author { background: color-mix(in oklab, var(--green) 12%, var(--paper)); color: color-mix(in oklab, var(--green) 60%, var(--st-ink)); }
}

@media (max-width: 1180px) {
  .view { padding: 40px 36px 80px; }
  .u-stats { grid-template-columns: repeat(2, 1fr); }
}
</style>
