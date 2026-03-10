import React from 'react'
import ReactDOM from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import App from './App'
import './index.css'

// #region agent log
fetch('http://127.0.0.1:7487/ingest/66eada7b-49c2-4766-979c-edf5216a8dfb',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'73e4d6'},body:JSON.stringify({sessionId:'73e4d6',runId:'pre-fix',hypothesisId:'H2',location:'frontend/src/main.tsx:8',message:'main_module_loaded',data:{href:window.location.href,userAgent:navigator.userAgent.slice(0,120)},timestamp:Date.now()})}).catch(()=>{});
// #endregion

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      staleTime: 15_000,
      retry: 0,
    },
  },
})

const rootEl = document.getElementById('root')
// #region agent log
fetch('http://127.0.0.1:7487/ingest/66eada7b-49c2-4766-979c-edf5216a8dfb',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'73e4d6'},body:JSON.stringify({sessionId:'73e4d6',runId:'pre-fix',hypothesisId:'H3',location:'frontend/src/main.tsx:23',message:'root_lookup',data:{rootFound:Boolean(rootEl)},timestamp:Date.now()})}).catch(()=>{});
// #endregion

if (!rootEl) {
  throw new Error('Root element #root not found')
}

ReactDOM.createRoot(rootEl).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>
  </React.StrictMode>,
)

// #region agent log
fetch('http://127.0.0.1:7487/ingest/66eada7b-49c2-4766-979c-edf5216a8dfb',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'73e4d6'},body:JSON.stringify({sessionId:'73e4d6',runId:'pre-fix',hypothesisId:'H3',location:'frontend/src/main.tsx:38',message:'react_render_called',data:{ok:true},timestamp:Date.now()})}).catch(()=>{});
// #endregion
