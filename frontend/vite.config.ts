import path from 'path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [
    react(),
    {
      name: 'agent-debug-vite-requests',
      configureServer(server) {
        // #region agent log
        fetch('http://127.0.0.1:7487/ingest/66eada7b-49c2-4766-979c-edf5216a8dfb',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'73e4d6'},body:JSON.stringify({sessionId:'73e4d6',runId:'pre-fix',hypothesisId:'H5',location:'frontend/vite.config.ts:12',message:'vite_server_configured',data:{node:process.version,pid:process.pid},timestamp:Date.now()})}).catch(()=>{})
        // #endregion
        server.middlewares.use((req, res, next) => {
          const url = req.url ?? ''
          if (
            url === '/' ||
            url.startsWith('/@vite/client') ||
            url.startsWith('/@react-refresh') ||
            url.startsWith('/src/main.tsx')
          ) {
            const started = Date.now()
            res.on('finish', () => {
              // #region agent log
              fetch('http://127.0.0.1:7487/ingest/66eada7b-49c2-4766-979c-edf5216a8dfb',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'73e4d6'},body:JSON.stringify({sessionId:'73e4d6',runId:'pre-fix',hypothesisId:'H1',location:'frontend/vite.config.ts:26',message:'vite_request_finished',data:{url,statusCode:res.statusCode,durationMs:Date.now()-started},timestamp:Date.now()})}).catch(()=>{})
              // #endregion
            })
          }
          next()
        })
        process.on('SIGTERM', () => {
          // #region agent log
          fetch('http://127.0.0.1:7487/ingest/66eada7b-49c2-4766-979c-edf5216a8dfb',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'73e4d6'},body:JSON.stringify({sessionId:'73e4d6',runId:'pre-fix',hypothesisId:'H6',location:'frontend/vite.config.ts:40',message:'vite_process_sigterm',data:{pid:process.pid},timestamp:Date.now()})}).catch(()=>{})
          // #endregion
        })
        process.on('SIGINT', () => {
          // #region agent log
          fetch('http://127.0.0.1:7487/ingest/66eada7b-49c2-4766-979c-edf5216a8dfb',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'73e4d6'},body:JSON.stringify({sessionId:'73e4d6',runId:'pre-fix',hypothesisId:'H6',location:'frontend/vite.config.ts:45',message:'vite_process_sigint',data:{pid:process.pid},timestamp:Date.now()})}).catch(()=>{})
          // #endregion
        })
        process.on('exit', (code) => {
          // #region agent log
          fetch('http://127.0.0.1:7487/ingest/66eada7b-49c2-4766-979c-edf5216a8dfb',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'73e4d6'},body:JSON.stringify({sessionId:'73e4d6',runId:'pre-fix',hypothesisId:'H6',location:'frontend/vite.config.ts:50',message:'vite_process_exit',data:{pid:process.pid,code},timestamp:Date.now()})}).catch(()=>{})
          // #endregion
        })
      },
    },
  ],
  resolve: {
    alias: { '@': path.resolve(__dirname, './src') },
  },
  clearScreen: false,
  server: {
    port: 5173,
    strictPort: true,
    host: true,
    open: true,
    proxy: {
      '/api': { target: 'http://127.0.0.1:8080', changeOrigin: true },
      '/health': { target: 'http://127.0.0.1:8080', changeOrigin: true },
    },
  },
  envPrefix: ['VITE_'],
})
