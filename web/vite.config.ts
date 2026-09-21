import { writeFileSync } from 'node:fs'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [
    vue(),
    // dist/.gitkeep 入库占位，让 go:embed 在未构建时也能编译；构建会清空 dist，这里补回
    { name: 'keep-dist-placeholder', closeBundle: () => writeFileSync('dist/.gitkeep', '') },
  ],
  server: {
    port: 5173,
    proxy: {
      '/admin': 'http://127.0.0.1:8080',
      '/v1': 'http://127.0.0.1:8080',
      '/assets': 'http://127.0.0.1:8080',
    },
  },
  build: {
    outDir: 'dist',
    // vendor 分包：把 tdesign / echarts / vue 全家桶各自拆出去，
    // 首屏只需要 index + tdesign 两块，且版本不变时可长期命中缓存。
    rollupOptions: {
      output: {
        manualChunks(id: string) {
          if (!id.includes('node_modules')) return undefined
          if (id.includes('echarts') || id.includes('zrender')) return 'vendor-echarts'
          if (id.includes('tdesign')) return 'vendor-tdesign'
          if (id.includes('/vue/') || id.includes('@vue/') || id.includes('vue-router') || id.includes('vue-i18n') || id.includes('pinia')) return 'vendor-vue'
          return 'vendor-misc'
        },
      },
    },
  },
})
