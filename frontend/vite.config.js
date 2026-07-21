import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

const tunnelAllowedHosts = ['frp-fun.com']

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    host: '127.0.0.1',
    port: 5189,
    strictPort: true,
    allowedHosts: tunnelAllowedHosts,
    proxy: {
      // 外网只需要穿透前端端口；API 和媒体请求由 Vite 转发到本机 Hertz 服务。
      '/api': {
        target: 'http://127.0.0.1:20000',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
  preview: {
    host: '127.0.0.1',
    port: 5189,
    strictPort: true,
    allowedHosts: tunnelAllowedHosts,
  },
})
