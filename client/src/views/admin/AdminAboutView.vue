<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../api';
import { useConfigStore, type AboutModule, type SiteConfig } from '../../stores/config';
import { createModule, metaOf } from '../../about/registry';
import ModulePicker from '../../components/admin/ModulePicker.vue';
import { MODULE_EDITORS } from '../../components/admin/modules';

/** 关于管理：身份默认置顶，其余模块条目式增删/拖拽/行内编辑 */
const { t } = useI18n();
const configStore = useConfigStore();

type AboutCfg = SiteConfig['about'];
const about = reactive<AboutCfg>(JSON.parse(JSON.stringify(configStore.cfg.about)));
const skillsText = ref('');
const message = ref('');
const busy = ref(false);
const pickerOpen = ref(false);
const expanded = ref('');

async function load(): Promise<void> {
  const remote = await adminApi.settings() as unknown as SiteConfig;
  Object.assign(about, JSON.parse(JSON.stringify(remote.about)));
  if (!Array.isArray(about.modules)) about.modules = [];
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

/* 头像 */
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

/* 模块操作 */
function addModules(types: string[]): void {
  for (const type of types) about.modules.push(createModule(type));
  pickerOpen.value = false;
}

function removeModule(mod: AboutModule): void {
  about.modules = about.modules.filter((m) => m.id !== mod.id);
}

let dragIndex = -1;

function onDragStart(i: number): void {
  dragIndex = i;
}

function onDrop(i: number): void {
  if (dragIndex < 0 || dragIndex === i) return;
  const [moved] = about.modules.splice(dragIndex, 1);
  about.modules.splice(i, 0, moved);
  dragIndex = -1;
}

onMounted(load);
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
        <button class="a-btn ghost" @click="pickerOpen = true">{{ t('admin.addModule') }}</button>
        <button class="a-btn primary" :disabled="busy" @click="save">{{ t('admin.saveAll') }}</button>
      </div>
    </header>

    <!-- 身份（默认模块，置顶不可删） -->
    <section class="a-card block identity-card">
      <div class="row-head">
        <span class="m-icon locked">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="8" r="4" />
            <path d="M4 21c0-4 3.6-6.5 8-6.5s8 2.5 8 6.5" />
          </svg>
        </span>
        <strong>{{ t('admin.aboutIdentity') }}</strong>
        <span class="lock-tag">{{ t('admin.moduleLocked') }}</span>
      </div>
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

    <!-- 模块条目列表 -->
    <div class="mod-list">
      <div
        v-for="(mod, i) in about.modules"
        :key="mod.id"
        class="mod-row a-card"
        :class="{ open: expanded === mod.id }"
        draggable="true"
        @dragstart="onDragStart(i)"
        @dragover.prevent
        @drop="onDrop(i)"
      >
        <div class="row-head">
          <span class="grip" aria-hidden="true">⋮⋮</span>
          <span class="m-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
              <path :d="metaOf(mod.type)?.icon" />
            </svg>
          </span>
          <strong>{{ metaOf(mod.type)?.name }}</strong>
          <span class="summary">{{ metaOf(mod.type)?.summary(mod.data) }}</span>
          <button class="op" @click="expanded = expanded === mod.id ? '' : mod.id">
            {{ expanded === mod.id ? t('admin.collapse') : t('admin.edit') }}
          </button>
          <button class="op danger" @click="removeModule(mod)">{{ t('admin.delete') }}</button>
        </div>

        <!-- 行内编辑器（按类型映射到 components/admin/modules/Editor*.vue） -->
        <div v-if="expanded === mod.id" class="editor">
          <component :is="MODULE_EDITORS[mod.type]" :mod="mod" />
        </div>
      </div>
    </div>

    <p v-if="!about.modules.length" class="empty">{{ t('admin.noModules') }}</p>

    <ModulePicker v-if="pickerOpen" @close="pickerOpen = false" @add="addModules" />
  </div>
</template>

<style scoped lang="scss">
.about-admin { max-width: 860px; }

.actions { display: flex; align-items: center; gap: 12px; }
.msg { font-size: 13px; color: var(--primary); }
.block { margin-bottom: 16px; }
.hint { font-size: 13px; color: var(--text-2); }
.empty { color: var(--text-2); text-align: center; padding: 40px 0; }

/* 行头 */
.row-head {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;

  strong { font-size: 14.5px; flex-shrink: 0; }
}

.grip {
  color: var(--text-2);
  letter-spacing: -2px;
  user-select: none;
  cursor: grab;

  &:active { cursor: grabbing; }
}

.m-icon {
  width: 32px;
  height: 32px;
  border-radius: 9px;
  display: grid;
  place-items: center;
  flex-shrink: 0;
  background: rgba(var(--primary-rgb), 0.1);
  color: var(--primary);

  svg { width: 17px; height: 17px; }
}

.lock-tag {
  font-size: 10.5px;
  font-weight: 700;
  padding: 2px 10px;
  border-radius: 999px;
  background: var(--surface-2);
  color: var(--text-2);
}

.summary {
  flex: 1;
  min-width: 0;
  font-size: 12.5px;
  color: var(--text-2);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 模块列表 */
.mod-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.mod-row {
  padding: 14px 18px;
  transition: border-color var(--dur-fast), box-shadow var(--dur);

  &.open { border-color: rgba(var(--primary-rgb), 0.45); }
  &:hover { box-shadow: 0 8px 24px -12px rgba(var(--primary-rgb), 0.2); }
}

.editor {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px dashed var(--border);
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.op {
  border: none;
  background: none;
  font-size: 13px;
  font-weight: 600;
  color: var(--primary);
  flex-shrink: 0;

  &.danger { color: var(--accent-red); }
  &:hover { opacity: 0.75; }
}

/* 身份卡（模块编辑器内部样式见 components/admin/modules/module-editor.scss） */
.identity-card .row-head { margin-bottom: 16px; }

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

@media (max-width: 720px) {
  .identity { flex-direction: column; }
  .row2 { grid-template-columns: 1fr; }
}
</style>
