import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, type MediaFolder } from '../../../api';
import { useDialogStore } from '../../../stores/dialog';
import { toast } from './toast';
import './i18n';

/** 当前浏览位置：全部素材 / 未归类 / 某个文件夹路径 */
export const ALL = '\u0000all';

/** 拖动素材到文件夹时 dataTransfer 里的类型（值为 JSON 数组：素材名） */
export const DRAG_TYPE = 'application/x-myself-media';

/**
 * 素材文件夹的数据与操作（素材页与素材库模态框共用）：列表、新建 / 重命名 / 删除（确认后）、移动素材。
 * 所有改动后自动重新拉取列表；onChanged 让调用方刷新素材。
 */
export function useFolders(onChanged?: () => void) {
  const { t } = useI18n();
  const dialog = useDialogStore();
  const folders = ref<MediaFolder[]>([]);

  const errText = (e: unknown): string => {
    const code = (e as Error).message;
    return t(code === 'invalid_folder' || code === 'folder_too_deep' ? `studio.folder.err_${code}` : 'studio.saveFailed');
  };

  async function load(): Promise<void> {
    try {
      folders.value = await adminApi.mediaFolders();
    } catch {
      folders.value = [];
    }
  }

  async function create(parent = ''): Promise<string | null> {
    const name = await dialog.prompt({
      title: parent ? t('studio.folder.newIn', { name: parent }) : t('studio.folder.new'),
      placeholder: t('studio.folder.namePh'),
      confirmText: t('studio.folder.create'),
    });
    if (!name?.trim()) return null;
    try {
      const { path } = await adminApi.createFolder(parent ? `${parent}/${name.trim()}` : name.trim());
      await load();
      return path;
    } catch (e) {
      toast(errText(e), { icon: 'x' });
      return null;
    }
  }

  async function rename(path: string): Promise<string | null> {
    const i = path.lastIndexOf('/');
    const parent = i >= 0 ? path.slice(0, i) : '';
    const name = await dialog.prompt({
      title: t('studio.folder.rename'),
      inputValue: path.slice(i + 1),
      confirmText: t('studio.editor.ok'),
    });
    if (!name?.trim() || name.trim() === path.slice(i + 1)) return null;
    try {
      const res = await adminApi.renameFolder(path, parent ? `${parent}/${name.trim()}` : name.trim());
      await load();
      onChanged?.();
      return res.path;
    } catch (e) {
      toast(errText(e), { icon: 'x' });
      return null;
    }
  }

  async function remove(path: string): Promise<boolean> {
    const ok = await dialog.confirm({
      title: t('studio.folder.deleteTitle', { name: path }),
      message: t('studio.folder.deleteBody'),
      confirmText: t('studio.delete'),
      danger: true,
    });
    if (!ok) return false;
    try {
      await adminApi.deleteFolder(path);
      await load();
      onChanged?.();
      return true;
    } catch (e) {
      toast(errText(e), { icon: 'x' });
      return false;
    }
  }

  async function move(names: string[], folder: string): Promise<boolean> {
    if (!names.length) return false;
    try {
      const { moved } = await adminApi.moveMedia(names, folder);
      toast(t('studio.folder.moved', { n: moved, name: folder || t('studio.folder.unfiled') }), { icon: 'check' });
      await load();
      onChanged?.();
      return true;
    } catch (e) {
      toast(errText(e), { icon: 'x' });
      return false;
    }
  }

  return { folders, load, create, rename, remove, move };
}
