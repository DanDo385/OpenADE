import path from 'path'
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const backendPort = env.OPENADE_PORT?.trim() || '8080'
  const proxyTarget = `http://127.0.0.1:${backendPort}`

  return {
    plugins: [react()],
    resolve: {
      alias: { '@': path.resolve(__dirname, './src') },
    },
    clearScreen: false,
    server: {
      port: 5173,
      strictPort: true,
      host: true,
      proxy: {
        '/api': { target: proxyTarget, changeOrigin: true },
        '/health': { target: proxyTarget, changeOrigin: true },
      },
    },
    envPrefix: ['VITE_'],
  }
})
