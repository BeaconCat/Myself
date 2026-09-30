<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { onBeforeRouteLeave } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../api';
import { useConfigStore, type AboutModule, type SiteConfig } from '../../stores/config';
import { useDialogStore } from '../../stores/dialog';
import { normalizeIdentity } from '../../about/identity';
import { migrateModules } from '../../about/migrate';
import { SCENES, SCENE_LABELS, SOCIAL_ICONS } from '../../about/icons';
import { BRANDS, CARD_LINK_MAX } from '../../about/brands';
import type { SocialLink } from '../../about/types';
import AboutModules from '../../about/AboutModules.vue';
import KitIcon from '../../about/parts/KitIcon.vue';
import { settle, stableJson } from './studio/state';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import StSwitch from './studio/StSwitch.vue';
import MediaPicker from './studio/MediaPicker.vue';
import { toast } from './studio/toast';

/**
 * 身份：站点身份的唯一编辑入口 —— 头像、形象图、名片头图、名字、签名、自述、状态、链接、格言。
 * 关于页身份区与收尾格言、首页关于卡、页脚、移动端抽屉、文章作者栏都读这里。
 * 右侧实时预览按比例缩放渲染真实的关于页身份区与格言（沿用关于页里这两个模块的展示选项）。
 */
const { t } = useI18n();
const config = useConfigStore();
const dialog = useDialogStore();

type AboutCfg = SiteConfig['about'];
const about = reactive<AboutCfg>(JSON.parse(JSON.stringify(config.cfg.about)));
normalizeIdentity(about);
const snapshot = ref('');
const loaded = ref(false);
const busy = ref(false);

/** 网站名称（site.title）：与「设置 · 站点标题」是同一个字段，这里一并编辑 */
/** 网站名称与站点 logo 属于 site，和身份一起在这里编辑 */
const site = reactive({ title: config.cfg.site.title, logo: config.cfg.site.logo ?? '' });
const state = () => stableJson([about, site.title, site.logo]);
const dirty = computed(() => loaded.value && state() !== snapshot.value);

async function load(): Promise<void> {
  try {
    const remote = (await adminApi.settings()) as unknown as SiteConfig;
    Object.assign(about, JSON.parse(JSON.stringify(remote.about ?? {})));
    site.title = remote.site?.title ?? site.title;
    site.logo = remote.site?.logo ?? '';
  } catch {
    toast(t('studio.loadFailed'), { icon: 'x' });
  }
  normalizeIdentity(about);
  if (!Array.isArray(about.modules)) about.modules = [];
  await settle();
  snapshot.value = state();
  loaded.value = true;
}

async function save(): Promise<void> {
  if (busy.value || !dirty.value) return;
  busy.value = true;
  try {
    await adminApi.saveSettings({ about: JSON.parse(JSON.stringify(about)), site: { title: site.title.trim() || 'Myself', logo: site.logo } });
    snapshot.value = state();
    await config.load();
    toast(t('studio.identity.saved'));
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  } finally {
    busy.value = false;
  }
}

/* ---------- 图片上传（头像 / 形象图） ---------- */
const uploading = ref<'' | 'avatar' | 'portrait' | 'banner' | 'logo'>('');
const logoInput = ref<HTMLInputElement | null>(null);

/* ---------- 从素材库选择 ---------- */
type PickTarget = '' | 'logo' | 'avatar' | 'portrait' | 'banner';
const picking = ref<PickTarget>('');
const pickCurrent = computed(() => {
  switch (picking.value) {
    case 'logo': return site.logo;
    case 'avatar': return about.avatar;
    case 'portrait': return about.portrait.src;
    case 'banner': return about.banner.src;
    default: return '';
  }
});
function onPicked(url: string): void {
  if (picking.value === 'logo') site.logo = url;
  else if (picking.value === 'avatar') about.avatar = url;
  else if (picking.value === 'portrait') about.portrait.src = url;
  else if (picking.value === 'banner') about.banner.src = url;
}
const bannerInput = ref<HTMLInputElement | null>(null);
const avatarInput = ref<HTMLInputElement | null>(null);
const portraitInput = ref<HTMLInputElement | null>(null);

async function upload(e: Event, target: 'avatar' | 'portrait' | 'banner' | 'logo'): Promise<void> {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = '';
  if (!file || uploading.value) return;
  uploading.value = target;
  try {
    const [up] = await adminApi.uploadMedia([file]);
    if (!up) return;
    if (target === 'avatar') about.avatar = up.url;
    else if (target === 'banner') about.banner.src = up.url;
    else if (target === 'logo') site.logo = up.url;
    else about.portrait.src = up.url;
  } catch {
    toast(t('studio.identity.uploadFailed'), { icon: 'x' });
  } finally {
    uploading.value = '';
  }
}

/* ---------- 形象图：渐隐 / 圆角 / 焦点 ---------- */
const FADES = ['left', 'bottom', 'none'] as const;

const radiusAuto = computed<boolean>({
  get: () => typeof about.portrait.radius !== 'number',
  set: (auto) => {
    if (auto) delete about.portrait.radius;
    else about.portrait.radius = 24;
  },
});

const focus = computed(() => {
  const [x, y] = (about.portrait.focus || '50% 40%').split(/\s+/).map((s) => parseFloat(s));
  return { x: Number.isFinite(x) ? x : 50, y: Number.isFinite(y) ? y : 40 };
});
function setFocus(axis: 'x' | 'y', v: number): void {
  const f = { ...focus.value, [axis]: v };
  about.portrait.focus = `${f.x}% ${f.y}%`;
}

/* ---------- 链接 ---------- */
function addLink(): void {
  about.links.push({ name: '', handle: '', url: '', icon: 'link', primary: about.links.length === 0 });
}
function moveLink(i: number, d: -1 | 1): void {
  const j = i + d;
  if (j < 0 || j >= about.links.length) return;
  const [l] = about.links.splice(i, 1);
  about.links.splice(j, 0, l);
}
function removeLink(i: number): void {
  about.links.splice(i, 1);
}
/** 平台名：品牌取注册表里的名字，通用图标走字典 */
const iconLabel = (ic: string): string => BRANDS[ic]?.label ?? t(`studio.identity.ic_${ic}`);

/** 名片按钮：最多 3 个 */
function toggleCard(l: SocialLink): void {
  if (!l.card && about.links.filter((x) => x.card).length >= CARD_LINK_MAX) {
    toast(t('studio.identity.cardMax', { n: CARD_LINK_MAX }), { icon: 'info' });
    return;
  }
  l.card = !l.card;
}

/* ---------- 名片头图：默认封面（按名字选）或上传 ---------- */
const coverSrc = (id: string): string => `/covers/${id}.webp`;
const bannerSrc = computed(() => about.banner.src || coverSrc('05'));

/** 主按钮只允许一个 */
function setPrimary(l: SocialLink): void {
  const on = !l.primary;
  about.links.forEach((x) => { x.primary = false; });
  l.primary = on;
}

/* ---------- 实时预览：真实模块按比例缩放 ---------- */
const PREVIEW_W = 1180;
const previewBox = ref<HTMLElement | null>(null);
const previewInner = ref<HTMLElement | null>(null);
const scale = ref(0.36);
const previewH = ref(360);
let ro: ResizeObserver | null = null;

function measure(): void {
  const box = previewBox.value;
  const inner = previewInner.value;
  if (!box || !inner) return;
  scale.value = box.clientWidth / PREVIEW_W;
  previewH.value = Math.ceil(inner.scrollHeight * scale.value);
}

/** 预览沿用关于页里身份区 / 格言模块的展示选项（顶部小字、收尾装饰） */
const previewModules = computed<AboutModule[]>(() => {
  const mods = migrateModules(about.modules, about);
  const profile = mods.find((m) => m.type === 'profile');
  const motto = mods.find((m) => m.type === 'motto');
  return [
    { ...(profile ?? { id: 'pv-profile', type: 'profile', data: {} }), span: 3, variant: 'portrait', hidden: false },
    ...(about.motto ? [{ ...(motto ?? { id: 'pv-motto', type: 'motto', data: {} }), span: 3 as const, variant: 'closing', hidden: false }] : []),
  ];
});

watch(previewInner, (el) => {
  ro?.disconnect();
  if (!el || !previewBox.value) return;
  ro = new ResizeObserver(measure);
  ro.observe(el);
  ro.observe(previewBox.value);
});

/* ---------- 离开确认 / 快捷键 ---------- */
onBeforeRouteLeave(async () => {
  if (!dirty.value) return true;
  return dialog.confirm({
    title: t('studio.identity.leaveTitle'),
    message: t('studio.identity.leaveBody'),
    confirmText: t('studio.write.leave'),
    danger: true,
  });
});

function onKey(e: KeyboardEvent): void {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') {
    e.preventDefault();
    void save();
  }
}

onMounted(() => {
  void load();
  window.addEventListener('keydown', onKey);
});
onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey);
  ro?.disconnect();
});
</script>

<template>
  <section class="studio view">
    <div class="st-vh">
      <div>
        <h1>{{ t('studio.identity.title') }}</h1>
        <p>{{ t('studio.identity.desc') }}</p>
      </div>
      <div class="act">
        <router-link class="st-btn g" :to="{ name: 'admin-about' }"><SIcon name="layers" :size="18" />{{ t('studio.identity.toAbout') }}</router-link>
        <button type="button" class="st-btn" :class="dirty ? 'p' : 'g'" :disabled="busy || !dirty" @click="save">
          <SIcon name="check" :size="18" />{{ dirty ? t('studio.save') : t('studio.saved') }}
        </button>
      </div>
    </div>

    <div class="layout">
      <div class="form">
        <!-- 形象：头像 + 形象图 -->
        <section class="st-card st-rise" style="--i: 0">
          <div class="st-sec-t"><h2>{{ t('studio.identity.look') }}</h2><span>{{ t('studio.identity.lookSub') }}</span></div>
          <div class="look">
            <!-- 站点 logo：后台品牌、登录页、浏览器标签图标，以及头像 / 形象图未上传时的回退 -->
            <div class="av-block">
              <button type="button" class="av logo" :class="{ busy: uploading === 'logo' }" :title="t('studio.identity.logoPick')" @click="logoInput?.click()">
                <img :src="site.logo || '/favicon-256.png'" alt="" draggable="false" />
                <span class="av-over"><SIcon name="upload" :size="20" /></span>
              </button>
              <input ref="logoInput" type="file" accept="image/*" hidden @change="upload($event, 'logo')" />
              <div class="av-meta">
                <b>{{ t('studio.identity.logo') }}</b>
                <small>{{ t('studio.identity.logoSub') }}</small>
              </div>
              <div class="row-btns">
                <button v-if="site.logo" type="button" class="st-btn q sm" @click="site.logo = ''">{{ t('studio.identity.useBuiltinLogo') }}</button>
                <button type="button" class="st-btn g sm" @click="picking = 'logo'">{{ t('studio.identity.fromLibrary') }}</button>
                <button type="button" class="st-btn g sm" :disabled="!!uploading" @click="logoInput?.click()">{{ site.logo ? t('studio.identity.replace') : t('studio.identity.upload') }}</button>
              </div>
            </div>

            <div class="av-block">
              <button type="button" class="av" :class="{ busy: uploading === 'avatar' }" :title="t('studio.identity.avatarPick')" @click="avatarInput?.click()">
                <img :src="about.avatar || site.logo || '/favicon-256.png'" alt="" draggable="false" />
                <span class="av-over"><SIcon name="upload" :size="20" /></span>
              </button>
              <input ref="avatarInput" type="file" accept="image/*" hidden @change="upload($event, 'avatar')" />
              <div class="av-meta">
                <b>{{ t('studio.identity.avatar') }}</b>
                <small>{{ t('studio.identity.avatarSub') }}</small>
              </div>
              <div class="row-btns">
                <button v-if="about.avatar" type="button" class="st-btn q sm" @click="about.avatar = ''">{{ t('studio.identity.useLogo') }}</button>
                <button type="button" class="st-btn g sm" @click="picking = 'avatar'">{{ t('studio.identity.fromLibrary') }}</button>
                <button type="button" class="st-btn g sm" :disabled="!!uploading" @click="avatarInput?.click()">{{ about.avatar ? t('studio.identity.replace') : t('studio.identity.upload') }}</button>
              </div>
            </div>

            <div class="pt-block">
              <figure class="pt-thumb" :class="[`fade-${about.portrait.fade}`, { logo: !about.portrait.src, busy: uploading === 'portrait' }]" @click="portraitInput?.click()">
                <span class="pt-frame" :style="{ borderRadius: radiusAuto ? 'var(--r-lg)' : `${Math.round((about.portrait.radius ?? 24) * 0.5)}px` }">
                  <img :src="about.portrait.src || site.logo || '/logo-1024.webp'" alt="" draggable="false" :style="{ objectPosition: about.portrait.focus || '50% 40%' }" />
                </span>
                <span class="pt-over"><SIcon name="upload" :size="20" /></span>
              </figure>
              <input ref="portraitInput" type="file" accept="image/*" hidden @change="upload($event, 'portrait')" />
              <div class="pt-ctrl">
                <div class="pt-head">
                  <div>
                    <b>{{ t('studio.identity.portrait') }}</b>
                    <small>{{ t('studio.identity.portraitSub') }}</small>
                  </div>
                  <div class="row-btns">
                    <button v-if="about.portrait.src" type="button" class="st-btn q sm" @click="about.portrait.src = ''">{{ t('studio.identity.useLogo') }}</button>
                    <button type="button" class="st-btn g sm" @click="picking = 'portrait'">{{ t('studio.identity.fromLibrary') }}</button>
                    <button type="button" class="st-btn g sm" :disabled="!!uploading" @click="portraitInput?.click()">{{ about.portrait.src ? t('studio.identity.replace') : t('studio.identity.upload') }}</button>
                  </div>
                </div>
                <div class="ctl">
                  <span class="st-flabel">{{ t('studio.identity.fade') }}</span>
                  <div class="seg">
                    <button v-for="f in FADES" :key="f" type="button" :class="{ on: about.portrait.fade === f }" @click="about.portrait.fade = f">{{ t(`studio.identity.fade_${f}`) }}</button>
                  </div>
                </div>
                <div class="ctl">
                  <span class="st-flabel">
                    {{ t('studio.identity.radius') }}
                    <em>{{ radiusAuto ? t('studio.identity.radiusAuto') : `${about.portrait.radius}px` }}</em>
                  </span>
                  <div class="slide">
                    <label class="chk"><input v-model="radiusAuto" type="checkbox" />{{ t('studio.identity.radiusAuto') }}</label>
                    <input v-model.number="about.portrait.radius" type="range" min="0" max="80" :disabled="radiusAuto" />
                  </div>
                </div>
                <div class="ctl two">
                  <label>
                    <span class="st-flabel">{{ t('studio.identity.focusX') }}<em>{{ focus.x }}%</em></span>
                    <input type="range" min="0" max="100" :value="focus.x" @input="setFocus('x', Number(($event.target as HTMLInputElement).value))" />
                  </label>
                  <label>
                    <span class="st-flabel">{{ t('studio.identity.focusY') }}<em>{{ focus.y }}%</em></span>
                    <input type="range" min="0" max="100" :value="focus.y" @input="setFocus('y', Number(($event.target as HTMLInputElement).value))" />
                  </label>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- 名片头图（移动端关于页名片顶部横幅，可关闭） -->
        <section class="st-card st-rise" style="--i: 1">
          <div class="st-sec-t bn-t">
            <div><h2>{{ t('studio.identity.banner') }}</h2><span>{{ t('studio.identity.bannerSub') }}</span></div>
            <StSwitch v-model="about.banner.show" />
          </div>
          <!-- 收起 / 展开：grid 行高 0fr ⇄ 1fr 过渡，内容常驻，高度连续变化不跳 -->
          <div class="bn-fold" :class="{ open: about.banner.show }" :aria-hidden="!about.banner.show" :inert="!about.banner.show">
            <div class="bn-clip">
            <div class="bn">
              <figure class="bn-pv" :class="{ busy: uploading === 'banner' }">
                <img :src="bannerSrc" alt="" draggable="false" :style="{ objectPosition: about.banner.focus || '50% 50%' }" />
              </figure>
              <div class="bn-grid">
                <button
                  v-for="id in SCENES"
                  :key="id"
                  type="button"
                  :class="{ on: bannerSrc === coverSrc(id) }"
                  :title="SCENE_LABELS[id]"
                  @click="about.banner.src = coverSrc(id)"
                ><img :src="`/covers/${id}-s.webp`" alt="" draggable="false" /></button>
                <button type="button" class="up" :title="t('studio.identity.fromLibrary')" @click="picking = 'banner'">
                  <SIcon name="image" :size="18" />
                </button>
                <button type="button" class="up" :disabled="!!uploading" :title="t('studio.identity.upload')" @click="bannerInput?.click()">
                  <SIcon name="upload" :size="18" />
                </button>
              </div>
              <input ref="bannerInput" type="file" accept="image/*" hidden @change="upload($event, 'banner')" />
            </div>
            </div>
          </div>
        </section>

        <!-- 名片：问候 / 名字 / 签名 / 自述 / 建站日期 -->
        <section class="st-card st-rise" style="--i: 1">
          <div class="st-sec-t"><h2>{{ t('studio.identity.card') }}</h2><span>{{ t('studio.identity.cardSub') }}</span></div>
          <div class="grid g-12">
            <label class="c3"><span class="st-flabel">{{ t('studio.identity.hello') }}</span><span class="st-field"><input v-model="about.hello" :placeholder="t('studio.identity.helloPh')" /></span></label>
            <label class="c5"><span class="st-flabel">{{ t('studio.identity.name') }}</span><span class="st-field"><input v-model="about.name" /></span></label>
            <label class="c4">
              <span class="st-flabel">{{ t('studio.identity.alias') }}<em>{{ t('studio.identity.optional') }}</em></span>
              <span class="st-field"><input v-model="about.alias" :placeholder="t('studio.identity.aliasPh')" /></span>
            </label>
            <label class="c12">
              <span class="st-flabel">{{ t('studio.identity.tagline') }}<em>{{ t('studio.identity.taglineHint') }}</em></span>
              <span class="st-field"><input v-model="about.tagline" /></span>
            </label>
            <label class="c12"><span class="st-flabel">{{ t('studio.identity.bio') }}</span><span class="st-field ta"><textarea v-model="about.bio" rows="4" /></span></label>
            <label class="c8">
              <span class="st-flabel">{{ t('studio.identity.siteName') }}<em>{{ t('studio.identity.siteNameHint') }}</em></span>
              <span class="st-field"><input v-model="site.title" /></span>
            </label>
            <label class="c4"><span class="st-flabel">{{ t('studio.identity.founded') }}</span><span class="st-field"><input v-model="about.foundedAt" type="date" /></span></label>
          </div>
        </section>

        <!-- 状态 -->
        <section class="st-card st-rise" style="--i: 2">
          <div class="st-sec-t"><h2>{{ t('studio.identity.status') }}</h2><span>{{ t('studio.identity.statusSub') }}</span></div>
          <div class="grid g-12">
            <label class="c5"><span class="st-flabel">{{ t('studio.identity.doing') }}</span><span class="st-field"><input v-model="about.status.doing" :placeholder="t('studio.identity.doingPh')" /></span></label>
            <label class="c4"><span class="st-flabel">{{ t('studio.identity.city') }}</span><span class="st-field"><input v-model="about.status.city" :placeholder="t('studio.identity.cityPh')" /></span></label>
            <label class="c3">
              <span class="st-flabel">{{ t('studio.identity.tz') }}</span>
              <span class="st-field"><span class="suffix">UTC</span><input v-model.number="about.status.tz" type="number" min="-12" max="14" step="0.5" /></span>
            </label>
          </div>
        </section>

        <!-- 链接 -->
        <section class="st-card st-rise" style="--i: 3">
          <div class="st-sec-t"><h2>{{ t('studio.identity.links') }}</h2><span>{{ t('studio.identity.linksSub') }}</span></div>
          <TransitionGroup name="lk" tag="div" class="links">
            <div v-for="(l, i) in about.links" :key="i" class="lk" :class="{ pri: l.primary }">
              <label class="ic" :title="t('studio.identity.icon')">
                <KitIcon :name="l.icon" :size="20" />
                <select v-model="l.icon" :aria-label="t('studio.identity.icon')">
                  <option v-for="ic in SOCIAL_ICONS" :key="ic" :value="ic">{{ iconLabel(ic) }}</option>
                </select>
              </label>
              <span class="st-field nm"><input v-model="l.name" :placeholder="t('studio.identity.linkName')" /></span>
              <span class="st-field hd"><input v-model="l.handle" :placeholder="t('studio.identity.linkHandle')" /></span>
              <span class="st-field url mono-in"><SIcon name="link" :size="15" /><input v-model="l.url" placeholder="https://…" /></span>
              <div class="ops">
                <span class="chips">
                  <button type="button" class="pri-chip" :class="{ on: l.card }" :title="t('studio.identity.onCardTip')" @click="toggleCard(l)">{{ t('studio.identity.onCard') }}</button>
                  <button type="button" class="pri-chip" :class="{ on: l.primary }" @click="setPrimary(l)">{{ t('studio.identity.primary') }}</button>
                </span>
                <span class="mv">
                  <button type="button" class="st-ibtn sm" :disabled="i === 0" :title="t('studio.identity.up')" @click="moveLink(i, -1)"><SIcon name="arrowL" :size="16" class="rot" /></button>
                  <button type="button" class="st-ibtn sm" :disabled="i === about.links.length - 1" :title="t('studio.identity.down')" @click="moveLink(i, 1)"><SIcon name="arrowR" :size="16" class="rot" /></button>
                  <button type="button" class="st-ibtn sm" :title="t('studio.delete')" @click="removeLink(i)"><SIcon name="trash" :size="16" /></button>
                </span>
              </div>
            </div>
          </TransitionGroup>
          <button type="button" class="add-link" @click="addLink"><SIcon name="plus" :size="16" />{{ t('studio.identity.addLink') }}</button>
        </section>

        <!-- 格言 -->
        <section class="st-card st-rise" style="--i: 4">
          <div class="st-sec-t"><h2>{{ t('studio.identity.motto') }}</h2><span>{{ t('studio.identity.mottoSub') }}</span></div>
          <div class="grid g-12">
            <label class="c8"><span class="st-flabel">{{ t('studio.identity.mottoText') }}</span><span class="st-field"><input v-model="about.motto" /></span></label>
            <label class="c4"><span class="st-flabel">{{ t('studio.identity.mottoSign') }}</span><span class="st-field"><input v-model="about.mottoSign" placeholder="NAME · SINCE 2026" /></span></label>
          </div>
        </section>
      </div>

      <!-- 实时预览 -->
      <aside class="pv st-rise" style="--i: 1">
        <div class="pv-head">
          <b>{{ t('studio.identity.preview') }}</b>
          <a class="st-link" href="/about" target="_blank" rel="noopener">{{ t('studio.identity.openAbout') }}<SIcon name="external" :size="14" /></a>
        </div>
        <div ref="previewBox" class="pv-stage" :style="{ height: `${previewH}px` }">
          <div ref="previewInner" class="pv-inner" :style="{ width: `${PREVIEW_W}px`, transform: `scale(${scale})` }">
            <AboutModules :modules="previewModules" :about="about" />
          </div>
        </div>
        <p class="pv-foot">{{ t('studio.identity.previewFoot') }}</p>
      </aside>
    </div>
  
    <MediaPicker
      :open="!!picking"
      :title="picking ? t(`studio.identity.pick_${picking}`) : ''"
      :current="pickCurrent"
      @pick="onPicked"
      @close="picking = ''"
    />
  </section>
</template>

<style scoped lang="scss">
.view {
  max-width: 1440px;
  margin: 0 auto;
  padding: 32px 48px 72px;
}

.layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(320px, 400px);
  gap: 24px;
  align-items: start;
}

.form { display: flex; flex-direction: column; gap: 20px; min-width: 0; }

.st-flabel em { font-style: normal; font-weight: 400; color: var(--st-ink-4); white-space: nowrap; }

/* 12 栏表单网格 */
.grid.g-12 {
  display: grid;
  grid-template-columns: repeat(12, minmax(0, 1fr));
  gap: 16px 14px;

  > label { display: block; min-width: 0; }
  .c3 { grid-column: span 3; }
  .c4 { grid-column: span 4; }
  .c5 { grid-column: span 5; }
  .c8 { grid-column: span 8; }
  .c12 { grid-column: 1 / -1; }
}

.row-btns { display: flex; gap: 6px; flex-wrap: wrap; }

/* ---------- 形象 ---------- */
.look {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

/* 头像：横排 —— 圆形头像 ｜ 说明 ｜ 操作 */
.av-block {
  display: flex;
  align-items: center;
  gap: 18px;
  padding: 18px;
  border-radius: var(--r-md);
  background: var(--well);
}

.av {
  position: relative;
  flex: none;
  width: 72px;
  height: 72px;
  padding: 0;
  border: 0;
  border-radius: 50%;
  overflow: hidden;
  cursor: pointer;
  box-shadow: 0 0 0 1px var(--line-2), 0 10px 24px -14px var(--st-shade);

  img { width: 100%; height: 100%; object-fit: cover; display: block; }
}

.av-over,
.pt-over {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  color: #fff;
  background: rgb(0 0 0 / 0.38);
  opacity: 0;
  transition: opacity var(--dur-fast);
}

.av:hover .av-over,
.av.busy .av-over,
.pt-thumb:hover .pt-over,
.pt-thumb.busy .pt-over { opacity: 1; }

.av-meta,
.pt-head > div {
  display: flex;
  flex-direction: column;
  gap: 4px;

  b { font-size: 14.5px; font-weight: 600; color: var(--st-ink); }
  small { font-size: 12.5px; line-height: 1.5; color: var(--st-ink-3); }
}

.av-meta { flex: 1; min-width: 0; }
.av-block > .row-btns { flex: none; }

/* ---------- 名片头图 ---------- */
.bn-t {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;

  > div { display: flex; flex-direction: column; gap: 2px; }
}

.bn-fold {
  display: grid;
  grid-template-rows: 0fr;
  opacity: 0;
  transition:
    grid-template-rows var(--dur-slow) var(--ease-out),
    opacity var(--dur) var(--ease-out);

  &.open {
    grid-template-rows: 1fr;
    opacity: 1;
  }
}

.bn-clip { min-height: 0; overflow: hidden; }

/* 内边距代替外边距：收起时不留残余高度 */
.bn {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.1fr);
  gap: 18px;
  padding-top: 16px;
  transform: translateY(-8px);
  transition: transform var(--dur-slow) var(--ease-out);
}

.bn-fold.open .bn { transform: none; }

@media (prefers-reduced-motion: reduce) {
  .bn-fold, .bn { transition: none; }
}

.bn-pv {
  aspect-ratio: 16 / 6;
  border-radius: var(--r-md);
  overflow: hidden;
  background: var(--well);
  box-shadow: 0 0 0 1px var(--line-2);
  transition: opacity var(--dur-fast);

  &.busy { opacity: 0.5; }
  img { width: 100%; height: 100%; object-fit: cover; display: block; }
}

.bn-grid {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: 6px;
  align-content: start;

  button {
    aspect-ratio: 4 / 3;
    border-radius: var(--r-sm);
    overflow: hidden;
    box-shadow: 0 0 0 1px var(--line-2);
    transition: box-shadow var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

    &:hover { transform: translateY(-1px); }
    &.on { box-shadow: 0 0 0 2px var(--ink); }
    img { width: 100%; height: 100%; object-fit: cover; display: block; }
  }

  .up {
    display: grid;
    place-items: center;
    color: var(--st-ink-3);
    background: var(--well);

    &:hover { color: var(--st-ink); }
  }
}


.pt-block {
  display: grid;
  grid-template-columns: 168px minmax(0, 1fr);
  gap: 22px;
  padding: 18px;
  border-radius: var(--r-md);
  background: var(--well);
}

/* 形象图缩略：与前台同一套渐隐遮罩，按比例缩小 */
.pt-thumb {
  position: relative;
  align-self: start;
  margin: 0;
  aspect-ratio: 4 / 5;
  cursor: pointer;
  border-radius: var(--r-md);
  overflow: hidden;
  background: var(--paper);
  box-shadow: 0 0 0 1px var(--line) inset;
}

.pt-frame {
  position: absolute;
  inset: 0;
  overflow: hidden;

  img { width: 100%; height: 100%; object-fit: cover; display: block; }
}

/* 未上传形象图：logo 原图自带圆角底，缩略图去掉纸面底与描边，按方形显示 */
.pt-thumb.logo {
  aspect-ratio: 1;
  background: none;
  box-shadow: none;
  border-radius: 0;

  .pt-frame img { object-fit: contain; }
  .pt-over { border-radius: 22%; }
}

.fade-left .pt-frame {
  -webkit-mask-image: linear-gradient(to right, transparent, rgb(0 0 0 / 0.3) 20%, rgb(0 0 0 / 0.86) 40%, #000 52%);
  mask-image: linear-gradient(to right, transparent, rgb(0 0 0 / 0.3) 20%, rgb(0 0 0 / 0.86) 40%, #000 52%);
}

.fade-bottom .pt-frame {
  -webkit-mask-image: linear-gradient(to top, transparent, rgb(0 0 0 / 0.42) 20%, rgb(0 0 0 / 0.94) 38%, #000 44%);
  mask-image: linear-gradient(to top, transparent, rgb(0 0 0 / 0.42) 20%, rgb(0 0 0 / 0.94) 38%, #000 44%);
}

.pt-ctrl { display: flex; flex-direction: column; gap: 14px; min-width: 0; }

.pt-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  flex-wrap: wrap;
}

.ctl {
  .st-flabel { margin-bottom: 6px; }

  &.two { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
}

.seg {
  display: inline-flex;
  gap: 2px;
  padding: 3px;
  border-radius: var(--r-sm);
  background: var(--paper);
  box-shadow: 0 0 0 1px var(--line) inset;

  button {
    height: 30px;
    padding: 0 13px;
    white-space: nowrap;
    border-radius: calc(var(--r-sm) - 2px);
    font-size: 13px;
    color: var(--st-ink-2);
    transition: background var(--dur-fast), color var(--dur-fast), box-shadow var(--dur-fast);

    &:hover { color: var(--st-ink); }
    &.on { background: var(--lift); color: var(--lift-fg); box-shadow: var(--lift-shadow); }
  }
}

.slide {
  display: flex;
  align-items: center;
  gap: 14px;

  input[type='range'] { flex: 1; }
}

input[type='range'] { width: 100%; accent-color: var(--solid); }
input[type='range']:disabled { opacity: 0.4; }

.chk {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--st-ink-2);
  white-space: nowrap;
  cursor: pointer;

  input { accent-color: var(--solid); }
}

/* ---------- 链接 ---------- */
.links { display: flex; flex-direction: column; gap: 10px; }

/* 每条链接一块两行小卡：图标 ｜ 名称 ｜ 账号 ｜ 操作；第二行完整地址 */
.lk {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr) minmax(0, 1fr) auto;
  grid-template-areas: 'ic nm hd ops' 'ic url url ops';
  gap: 8px 10px;
  padding: 12px;
  border-radius: var(--r-md);
  background: var(--paper);
  box-shadow: 0 0 0 1px var(--line-2);
  transition: box-shadow var(--dur-fast);

  &.pri { box-shadow: 0 0 0 1px color-mix(in oklab, var(--ink) 45%, transparent), 0 0 0 3px color-mix(in oklab, var(--ink) 10%, transparent); }

  .nm { grid-area: nm; }
  .hd { grid-area: hd; }
  .url { grid-area: url; gap: 8px; }
  .st-field { min-width: 0; }

  /* 图标：方块按钮，透明 select 覆盖在上面承接选择 */
  .ic {
    grid-area: ic;
    position: relative;
    align-self: stretch;
    display: grid;
    place-items: center;
    border-radius: var(--r-sm);
    background: var(--well);
    color: var(--st-ink);
    box-shadow: 0 0 0 1px var(--line) inset;
    cursor: pointer;
    transition: background var(--dur-fast);

    &:hover { background: var(--hover); }
    select { position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer; }
  }

  .ops {
    grid-area: ops;
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    justify-content: space-between;
    gap: 8px;
  }

  .mv { display: flex; gap: 2px; }
  .chips { display: flex; gap: 6px; }
  .rot { rotate: 90deg; }
}

.pri-chip {
  height: 28px;
  padding: 0 11px;
  border-radius: var(--r-pill);
  font-size: 12.5px;
  color: var(--st-ink-3);
  box-shadow: 0 0 0 1px var(--line-2) inset;
  white-space: nowrap;
  transition: background var(--dur-fast), color var(--dur-fast), box-shadow var(--dur-fast);

  &:hover { color: var(--st-ink); }
  &.on { background: var(--solid); color: var(--on-solid); box-shadow: none; }
}

.lk-move { transition: transform var(--dur) var(--ease-out); }
.lk-enter-active, .lk-leave-active { transition: opacity var(--dur) var(--ease-out), transform var(--dur) var(--ease-out); }
.lk-enter-from, .lk-leave-to { opacity: 0; transform: translateY(-6px); }
.lk-leave-active { position: absolute; }

.add-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 40px;
  border-radius: var(--r-sm);
  font-size: 13.5px;
  color: var(--st-ink-3);
  box-shadow: 0 0 0 1.5px var(--line-2) inset;
  transition: color var(--dur-fast), box-shadow var(--dur-fast);

  &:hover { color: var(--ink); box-shadow: 0 0 0 1.5px color-mix(in oklab, var(--ink) 55%, transparent) inset; }
}

/* ---------- 实时预览 ---------- */
.pv {
  position: sticky;
  top: 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 18px;
  border-radius: var(--r-lg);
  background: var(--paper);
  box-shadow: 0 0 0 1px var(--line-2), 0 14px 30px -22px color-mix(in oklab, var(--st-shade) 40%, transparent);
}

.pv-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;

  b { font: 700 18px/1.3 var(--font-serif); color: var(--st-ink); }
}

/* 预览舞台：前台底色，内部按 1180px 真实宽度渲染再整体缩放 */
.pv-stage {
  position: relative;
  overflow: hidden;
  border-radius: var(--r-md);
  background: var(--bg);
  box-shadow: 0 0 0 1px var(--line) inset;
  transition: height var(--dur) var(--ease-out);
}

.pv-inner {
  position: absolute;
  top: 0;
  left: 0;
  padding: 8px 48px 0;
  transform-origin: 0 0;
  pointer-events: none;
}

.pv-foot { margin: 0; font-size: 12.5px; color: var(--st-ink-3); line-height: 1.6; }

@media (max-width: 1100px) {
  .view { padding: 28px 32px 64px; }
  .layout { grid-template-columns: minmax(0, 1fr); }
  .pv { position: static; order: -1; }
}

@media (max-width: 720px) {
  .view { padding: 20px 16px 48px; }
  .grid.g-12 > label { grid-column: 1 / -1; }
  .pt-block { grid-template-columns: 1fr; }
  .pt-thumb { width: 150px; }
  .av-block { flex-wrap: wrap; }
  .ctl.two { grid-template-columns: 1fr; }
  .lk {
    grid-template-columns: 44px minmax(0, 1fr);
    grid-template-areas: 'ic nm' 'ic hd' 'url url' 'ops ops';
  }
  .lk .ops { flex-direction: row; align-items: center; }
}

/* 站点 logo：圆角方块（与头像的圆形区分） */
.av.logo { border-radius: var(--r-lg); }
.av.logo img { border-radius: inherit; }
</style>
