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
        <button type="button" class="st-btn g st-tip" :data-tip="tip" disabled><SIcon name="download" :size="18" />{{ t('studio.users.export') }}</button>
        <button type="button" class="st-btn p st-tip" :data-tip="tip" disabled><SIcon name="mail" :size="18" />{{ t('studio.users.invite') }}</button>
      </div>
    </div>

    <div class="st-note-bar st-rise"><SIcon name="info" />{{ t('studio.reserved.users') }}</div>

    <div class="st-stats u-stats st-rise" style="--i: 0">
      <div class="st-stat"><b>128<span class="dl up">+14</span></b><small>{{ t('studio.users.sTotal') }}</small></div>
      <div class="st-stat"><b>36</b><small>{{ t('studio.users.sActive') }}</small></div>
      <div class="st-stat"><b>3</b><small>{{ t('studio.users.sAuthors') }}</small></div>
      <div class="st-stat"><b>412<span class="dl up">+38</span></b><small>{{ t('studio.users.sComments') }}</small></div>
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
  max-width: 1280px;
  margin: 0 auto;
  padding: 32px 48px 72px;
}

.u-stats { margin-bottom: 24px; }

.u-stats .dl { margin-left: 6px; }

.st-table {
  .u { display: flex; align-items: center; gap: 12px; }

  .av {
    width: 38px;
    height: 38px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    color: #fff;
    font: 600 13px var(--font-serif);
    overflow: hidden;
    flex: none;

    img { width: 100%; height: 100%; object-fit: cover; }
  }

  b { display: block; font-weight: 600; font-size: 15px; }
  small { color: var(--st-ink-3); font-size: 13px; }
  .mono { font-size: 13px; color: var(--st-ink-3); }
  .act-cell { font-size: 14px; color: var(--st-ink-2); }
  .online { display: inline-flex; align-items: center; gap: 6px; .st-dot { --c: var(--green); } }
  .dash { font-size: 12px; color: var(--st-ink-3); }
}

.role {
  font-size: 13px;
  padding: 4px 11px;
  border-radius: var(--r-pill);
  background: var(--well-2);
  color: var(--st-ink-2);

  &.admin { background: var(--tint); color: var(--ink); }
  &.author { background: color-mix(in oklab, var(--green) 12%, var(--paper)); color: color-mix(in oklab, var(--green) 60%, var(--st-ink)); }
}

@media (max-width: 1180px) {
  .view { padding: 28px 32px 64px; }
  .u-stats { --n: 2; }
  .u-stats .st-stat:nth-child(3) { box-shadow: 0 -1px 0 var(--line-2); }
  .u-stats .st-stat:nth-child(4) { box-shadow: -1px 0 0 var(--line-2), 0 -1px 0 var(--line-2); }
}
</style>
