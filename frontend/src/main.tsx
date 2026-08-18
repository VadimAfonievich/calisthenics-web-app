import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './styles.css'

const apiBaseURL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api/v1'

function App() {
  return (
    <main className="mx-auto flex min-h-dvh max-w-md flex-col justify-center px-6 py-10">
      <p className="text-sm font-semibold tracking-wide text-emerald-400">CALISTHENICS COACH</p>
      <h1 className="mt-3 text-4xl font-bold tracking-tight text-white">Основа приложения готова</h1>
      <p className="mt-4 text-base leading-7 text-slate-300">
        Telegram-авторизация, маршруты и тренировочный функционал будут подключаться в следующих фазах.
      </p>
      <p className="mt-8 rounded-xl bg-slate-800 px-4 py-3 font-mono text-xs text-slate-300">
        API: {apiBaseURL}
      </p>
    </main>
  )
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
