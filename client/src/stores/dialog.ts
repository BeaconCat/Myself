import { defineStore } from 'pinia';

/**
 * 全局模态服务：替代原生 alert/confirm/prompt（项目规范禁用原生弹窗）。
 * 用法：await useDialogStore().confirm({ title, message })
 */
export interface DialogOptions {
  title?: string;
  message?: string;
  /** prompt 模式：显示输入框 */
  input?: boolean;
  inputValue?: string;
  placeholder?: string;
  /** prompt 输入框的可访问名称（缺省用 placeholder，再缺省用标题） */
  label?: string;
  confirmText?: string;
  cancelText?: string;
  /** 危险操作：确认键红色 */
  danger?: boolean;
  /** alert 模式：只有确认键 */
  alertOnly?: boolean;
}

interface ActiveDialog extends DialogOptions {
  resolve: (value: string | boolean | null) => void;
}

export const useDialogStore = defineStore('dialog', {
  state: () => ({
    active: null as ActiveDialog | null,
  }),
  actions: {
    open(options: DialogOptions): Promise<string | boolean | null> {
      return new Promise((resolve) => {
        this.active = { ...options, resolve };
      });
    },
    /** 确认框：resolve true/false */
    confirm(options: DialogOptions): Promise<boolean> {
      return this.open(options) as Promise<boolean>;
    },
    /** 输入框：resolve 字符串或 null（取消） */
    prompt(options: DialogOptions): Promise<string | null> {
      return this.open({ ...options, input: true }) as Promise<string | null>;
    },
    /** 提示框 */
    alert(options: DialogOptions): Promise<boolean> {
      return this.open({ ...options, alertOnly: true }) as Promise<boolean>;
    },
    settle(value: string | boolean | null) {
      this.active?.resolve(value);
      this.active = null;
    },
  },
});
