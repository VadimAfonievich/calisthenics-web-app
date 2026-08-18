import { StrictMode, useEffect } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import { AppRouter } from './app/router'
import { useSessionStore } from './store/session'
import { initializeTelegram } from './telegram/webapp'
import './styles.css'

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: 1, refetchOnWindowFocus: false } } })

function Bootstrap() {
  const bootstrap = useSessionStore((state) => state.bootstrap)
  useEffect(() => { const telegram = initializeTelegram(); void bootstrap(telegram.initData) }, [bootstrap])
  return <AppRouter />
}

createRoot(document.getElementById('root')!).render(
  <StrictMode><QueryClientProvider client={queryClient}><BrowserRouter><Bootstrap /></BrowserRouter></QueryClientProvider></StrictMode>,
)
