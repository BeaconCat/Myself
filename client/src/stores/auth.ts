import { defineStore } from 'pinia';

const TOKEN_KEY = 'myself.token';

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem(TOKEN_KEY) ?? '',
  }),
  getters: {
    loggedIn: (s) => !!s.token,
  },
  actions: {
    setToken(token: string) {
      this.token = token;
      localStorage.setItem(TOKEN_KEY, token);
    },
    logout() {
      this.token = '';
      localStorage.removeItem(TOKEN_KEY);
    },
  },
});

/** api 层直接读，避免与 pinia 循环依赖 */
export function readToken(): string {
  return localStorage.getItem(TOKEN_KEY) ?? '';
}
