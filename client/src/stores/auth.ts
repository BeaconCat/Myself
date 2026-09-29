import { defineStore } from 'pinia';
import type { SessionUser, UserRole } from '../api';

/**
 * 登录态。令牌只存在服务端下发的 HttpOnly 会话 Cookie 里，页面脚本读不到；
 * 这里保存「已登录」提示标记（非凭据，用于路由预判）与当前用户资料（启动时经 /auth/session 校准）。
 * 角色：admin 站长、author 协作作者、reader 读者。
 */
const SESSION_KEY = 'myself.session';
const ROLE_KEY = 'myself.role';
/** 旧版把 JWT 存在 localStorage：启动即清除 */
localStorage.removeItem('myself.token');

export const useAuthStore = defineStore('auth', {
  state: () => ({
    loggedIn: localStorage.getItem(SESSION_KEY) === '1',
    role: (localStorage.getItem(ROLE_KEY) as UserRole | null) ?? null,
    user: null as SessionUser | null,
    synced: false,
  }),
  getters: {
    /** 可进后台：站长与协作作者 */
    staff: (s) => s.loggedIn && (s.role === 'admin' || s.role === 'author'),
    isAdmin: (s) => s.loggedIn && s.role === 'admin',
  },
  actions: {
    /** 登录 / 注册 / 初始化 / 改密成功后：服务端已写入会话 Cookie，这里记下提示与用户资料 */
    markLoggedIn(user?: SessionUser) {
      this.loggedIn = true;
      localStorage.setItem(SESSION_KEY, '1');
      if (user) this.setUser(user);
    },
    setUser(user: SessionUser) {
      this.user = user;
      this.role = user.role;
      localStorage.setItem(ROLE_KEY, user.role);
    },
    clear() {
      this.loggedIn = false;
      this.role = null;
      this.user = null;
      localStorage.removeItem(SESSION_KEY);
      localStorage.removeItem(ROLE_KEY);
    },
    /** 启动校准：以服务端会话为准（过期 / 被吊销 / 角色被关闭时撤掉本地标记） */
    async sync() {
      try {
        const res = await fetch('/api/v1/auth/session', { credentials: 'same-origin' });
        const data = (await res.json()) as { loggedIn: boolean; user?: SessionUser };
        if (data.loggedIn && data.user) this.markLoggedIn(data.user);
        else this.clear();
      } catch { /* 后端不可达时保持原状 */ }
      this.synced = true;
    },
    /** 退出：让服务端清除会话 Cookie，再清本地 */
    async logout() {
      this.clear();
      await fetch('/api/v1/auth/logout', { method: 'POST', headers: { 'X-Requested-With': 'myself' } }).catch(() => undefined);
    },
  },
});

/** 会话失效（401）：清本地标记。供 api 层直接调用，避免与 pinia 循环依赖 */
export function clearSessionHint(): void {
  localStorage.removeItem(SESSION_KEY);
  localStorage.removeItem(ROLE_KEY);
}

export function hasSessionHint(): boolean {
  return localStorage.getItem(SESSION_KEY) === '1';
}

/** 路由预判用的角色提示（真实权限以服务端为准） */
export function roleHint(): UserRole | null {
  return localStorage.getItem(ROLE_KEY) as UserRole | null;
}

/** 协作作者可进入的后台页：自己的文章、写文章、素材 */
export const AUTHOR_ROUTES = new Set(['admin-posts', 'admin-write-post', 'admin-media']);

/** 该角色能否进入某个后台页 */
export function canEnterAdmin(role: UserRole | null, name: string): boolean {
  if (role === 'admin') return true;
  return role === 'author' && AUTHOR_ROUTES.has(name);
}

/** 登录后的落脚点：站长进概览、作者进文章、读者回前台 */
export function staffHome(role: UserRole): { name: string } | { path: string } {
  if (role === 'admin') return { name: 'admin-today' };
  if (role === 'author') return { name: 'admin-posts' };
  return { path: '/' };
}
