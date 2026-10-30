import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

// Vite 配置：Vue3 + TS；开发期 /api 代理到后端 127.0.0.1:8090
export default defineConfig({
  // Electron 打包后用 loadFile（file:// 协议）加载 index.html，
  // 必须用相对路径 base，否则 /assets/xxx 会被解析成 file:///D:/assets/xxx 导致白屏
  base: './',
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 5173,
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8090',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    chunkSizeWarningLimit: 1200,
    rollupOptions: {
      output: {
        manualChunks: {
          vue: ['vue', 'vue-router', 'pinia'],
          element: ['element-plus', '@element-plus/icons-vue'],
          request: ['axios'],
          echarts: ['echarts']
        }
      }
    }
  }
})
