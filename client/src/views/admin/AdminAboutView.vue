<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../api';
import { useConfigStore, type SiteConfig } from '../../stores/config';

/** 关于管理：关于页所有模块的可视化配置 */
const { t } = useI18n();
const configStore = useConfigStore();

type AboutCfg = SiteConfig['about'];
const about = reactive<AboutCfg>(JSON.parse(JSON.stringify(configStore.cfg.about)));
const skillsText = ref('');
const message = ref('');
const busy = ref(false);

async function load(): Promise<void> {
  const remote = await adminApi.settings() as unknown as SiteConfig;
  Object.assign(about, JSON.parse(JSON.stringify(remote.about)));
  skillsText.value = about.skills.join(', ');
}

async function save(): Promise<void> {
  if (busy.value) return;
  busy.value = true;
  message.value = '';
  try {
    about.skills = skillsText.value.split(/[,，]+/).map((s) => s.trim()).filter(Boolean);
    await adminApi.saveSettings({ about } as unknown as Record<string, unknown>);
    await configStore.load();
    message.value = t('admin.saved');
  } catch {
    message.value = t('admin.saveFailed');
  } finally {
    busy.value = false;
  }
}

/* 头像上传 */
const avatarInput = ref<HTMLInputElement | null>(null);
const avatarBusy = ref(false);

async function onAvatar(e: Event): Promise<void> {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = '';
  if (!file || avatarBusy.value) return;
  avatarBusy.value = true;
  try {
    const [uploaded] = await adminApi.uploadMedia([file]);
    if (uploaded) about.avatar = uploaded.url;
  } finally {
    avatarBusy.value = false;
  }
}

/* 列表模块增删 */
function addGroup(): void {
  about.skillGroups.push({ title: '新分组', items: [] });
}

function addMilestone(): void {
  about.milestones.push({ year: '', text: '' });
}

function addSocial(): void {
  about.socials.push({ name: '', url: '', icon: 'link' });
}

/** 分组条目用逗号文本编辑 */
const groupItemsText = reactive<Record<number, string>>({});

function syncGroupText(): void {
  about.skillGroups.forEach((g, i) => { groupItemsText[i] = g.items.join(', '); });
}

function applyGroupText(i: number): void {
  about.skillGroups[i].items = (groupItemsText[i] ?? '')
    .split(/[,，]+/).map((s) => s.trim()).filter(Boolean);
}

onMounted(async () => {
  await load();
  syncGroupText();
});
</script>

<template>
  <div class="about-admin">
    <header class="a-head">
      <div>
        <h1>{{ t('admin.menuAbout') }}</h1>
        <p>{{ t('admin.aboutAdminHint') }}</p>
      </div>
      <div class="actions">
        <span v-if="message" class="msg">{{ message }}</span>
        <button class="a-btn primary" :disabled="busy" @click="save">{{ t('admin.saveAll') }}</button>
      </div>
    </header>

    <!-- 身份 -->
    <section class="a-card block">
      <h2 class="sec-h">{{ t('admin.aboutIdentity') }}</h2>
      <div class="identity">
        <button class="avatar" :disabled="avatarBusy" @click="avatarInput?.click()">
          <img :src="about.avatar || '/favicon-256.png'" alt="" />
          <span class="avatar-tip">{{ avatarBusy ? t('admin.uploading') : t('admin.changeAvatar') }}</span>
        </button>
        <input ref="avatarInput" type="file" accept="image/*" hidden @change="onAvatar" />
        <div class="identity-fields">
          <div class="row2">
            <label><span>{{ t('admin.aboutName') }}</span><input v-model="about.name" class="a-input" type="text" /></label>
            <label><span>{{ t('admin.aboutTagline') }}</span><input v-model="about.tagline" class="a-input" type="text" /></label>
          </div>
          <label><span>{{ t('admin.aboutBio') }}</span><textarea v-model="about.bio" class="a-input" rows="3" /></label>
          <div class="row2">
            <label><span>{{ t('admin.aboutMotto') }}</span><input v-model="about.motto" class="a-input" type="text" /></label>
            <label><span>{{ t('admin.aboutFounded') }}</span><input v-model="about.foundedAt" class="a-input" type="date" /></label>
          </div>
          <label><span>{{ t('admin.aboutSkills') }}</span><input v-model="skillsText" class="a-input" type="text" /></label>
        </div>
      </div>
    </section>

    <!-- 技能分组 -->
    <section class="a-card block">
      <h2 class="sec-h">{{ t('admin.aboutGroups') }}</h2>
      <div v-for="(group, i) in about.skillGroups" :key="i" class="line">
        <input v-model="group.title" class="a-input title-in" type="text" :placeholder="t('admin.groupTitle')" />
        <input
          v-model="groupItemsText[i]"
          class="a-input flex-in"
          type="text"
          :placeholder="t('admin.tagsPlaceholder')"
          @change="applyGroupText(i)"
        />
        <button class="op danger" @click="about.skillGroups.splice(i, 1)">{{ t('admin.delete') }}</button>
      </div>
      <button class="a-btn ghost" @click="addGroup">{{ t('admin.addItem') }}</button>
    </section>

    <!-- 历程 -->
    <section class="a-card block">
      <h2 class="sec-h">{{ t('admin.aboutMilestones') }}</h2>
      <div v-for="(m, i) in about.milestones" :key="i" class="line">
        <input v-model="m.year" class="a-input title-in" type="text" :placeholder="t('admin.msYear')" />
        <input v-model="m.text" class="a-input flex-in" type="text" :placeholder="t('admin.msText')" />
        <button class="op danger" @click="about.milestones.splice(i, 1)">{{ t('admin.delete') }}</button>
      </div>
      <button class="a-btn ghost" @click="addMilestone">{{ t('admin.addItem') }}</button>
    </section>

    <!-- 社交链接 -->
    <section class="a-card block">
      <h2 class="sec-h">{{ t('admin.aboutSocials') }}</h2>
      <div v-for="(s, i) in about.socials" :key="i" class="line">
        <input v-model="s.name" class="a-input title-in" type="text" :placeholder="t('admin.socialName')" />
        <input v-model="s.url" class="a-input flex-in" type="text" placeholder="https://…" />
        <select v-model="s.icon" class="a-input icon-in">
          <option value="github">GitHub</option>
          <option value="mail">Mail</option>
          <option value="rss">RSS</option>
          <option value="link">Link</option>
        </select>
        <button class="op danger" @click="about.socials.splice(i, 1)">{{ t('admin.delete') }}</button>
      </div>
      <button class="a-btn ghost" @click="addSocial">{{ t('admin.addItem') }}</button>
    </section>
  </div>
</template>

<style scoped lang="scss">
.about-admin { max-width: 860px; }

.actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.msg { font-size: 13px; color: var(--primary); }

.block { margin-bottom: 18px; }

.sec-h {
  font-size: 16px;
  margin-bottom: 16px;
  padding-left: 12px;
  border-left: 4px solid var(--primary);
}

.identity {
  display: flex;
  gap: 22px;
  align-items: flex-start;
}

.avatar {
  position: relative;
  width: 108px;
  height: 108px;
  border-radius: 50%;
  overflow: hidden;
  border: 2px solid var(--border);
  padding: 0;
  background: var(--surface-2);
  cursor: pointer;
  flex-shrink: 0;
  transition: border-color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  img { width: 100%; height: 100%; object-fit: cover; display: block; }

  .avatar-tip {
    position: absolute;
    inset: auto 0 0;
    padding: 6px 0 8px;
    font-size: 11px;
    font-weight: 700;
    color: #fff;
    background: rgba(0, 0, 0, 0.55);
    opacity: 0;
    transition: opacity var(--dur-fast);
  }

  &:hover {
    border-color: var(--primary);
    transform: scale(1.03);

    .avatar-tip { opacity: 1; }
  }
}

.identity-fields {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.row2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

label {
  display: flex;
  flex-direction: column;
  gap: 6px;

  span { font-size: 12px; font-weight: 600; color: var(--text-2); }
}

.line {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.title-in { width: 130px; flex-shrink: 0; }
.flex-in { flex: 1; }
.icon-in { width: 110px; flex-shrink: 0; }

.op {
  border: none;
  background: none;
  font-size: 13px;
  font-weight: 600;
  flex-shrink: 0;

  &.danger { color: var(--accent-red); }
  &:hover { opacity: 0.75; }
}

@media (max-width: 720px) {
  .identity { flex-direction: column; }
  .row2 { grid-template-columns: 1fr; }
}
</style>
