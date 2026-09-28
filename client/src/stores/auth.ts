import { defineStore } from 'pinia';

/**
 * 管理员登录态。令牌只存在服务端下发的 HttpOnly 会话 Cookie 里，页面脚本读不到；
 * 这里仅保存一个「已登录」提示标记（非凭据），用于导航显示与路由预判，真实校验以接口 401 为准。
 */
const SESSION_KEY = 'myself.session';
/** 旧版把 JWT 存在 localStorage：启动即清除 */
localStorage.removeItem('myself.token');

export const useAuthStore = defineStore('auth', {
  state: () => ({
    loggedIn: localStorage.getItem(SESSION_KEY) === '1',
  }),
  actions: {
    /** 登录 / 初始化 / 改密成功后：服务端已写入会话 Cookie，这里只记下提示标记 */
    markLoggedIn() {
      this.loggedIn = true;
      localStorage.setItem(SESSION_KEY, '1');
    },
    /** 启动校准：本地标记存在但会话已过期 / 被吊销时，撤掉标记（导航不再显示后台入口） */
    async sync() {
      if (!this.loggedIn) return;
      try {
        const res = await fetch('/api/v1/auth/session', { credentials: 'same-origin' });
        const data = (await res.json()) as { loggedIn: boolean };
        if (!data.loggedIn) {
          this.loggedIn = false;
          localStorage.removeItem(SESSION_KEY);
        }
      } catch { /* 后端不可达时保持原状 */ }
    },
    /** 退出：让服务端清除会话 Cookie，再清本地标记 */
    async logout() {
      this.loggedIn = false;
      localStorage.removeItem(SESSION_KEY);
      await fetch('/api/v1/auth/logout', { method: 'POST', headers: { 'X-Requested-With': 'myself' } }).catch(() => undefined);
    },
  },
});

/** 会话失效（401）：清本地标记。供 api 层直接调用，避免与 pinia 循环依赖 */
export function clearSessionHint(): void {
  localStorage.removeItem(SESSION_KEY);
}

export function hasSessionHint(): boolean {
  return localStorage.getItem(SESSION_KEY) === '1';
}
