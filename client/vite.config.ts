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
    // 保留浏览器的 Host（不改写为后端地址）：后端 CSRF 校验比对 Origin 与 Host，
    // 生成的邀请 / 重置链接也指向开发服务器
    proxy: {
      '/api': { target: API, changeOrigin: false },
      '/uploads': { target: API, changeOrigin: false },
      '/feed': { target: API, changeOrigin: false },
    },
  },
});
