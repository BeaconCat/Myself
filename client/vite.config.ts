import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

/** 开发代理目标：默认本机 3100，可用 VITE_API 覆盖（并行开发时各自指向不同后端端口） */
const API = process.env.VITE_API || 'http://localhost:3100';

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: '../server/web/dist',
    emptyOutDir: true,
  },
  server: {
    proxy: {
      '/api': API,
      '/uploads': API,
      '/feed': API,
    },
  },
});
