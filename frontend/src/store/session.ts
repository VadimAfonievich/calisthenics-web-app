import { create } from 'zustand'
import { authenticateTelegram, type CurrentUser } from '../api/auth'
import { APIError, apiBaseURL, authRequestURL, healthProbeURL, isAPIBaseConfigured } from '../api/client'
import { runHealthProbe, safeNetworkError } from '../api/networkDiagnostics'
import type { TelegramDiagnostics } from '../telegram/webapp'
type SessionStatus = 'loading' | 'authenticated' | 'demo' | 'error'
export type RuntimeDiagnostics = TelegramDiagnostics & { apiBaseConfigured: boolean; apiBaseURL: string; authRequestURL: string; authRequestMethod: 'POST'; authStarted: boolean; authCompleted: boolean; authHTTPStatus?: number; authError?: string; healthProbeURL: string; healthProbeStarted: boolean; healthProbeCompleted: boolean; healthProbeHTTPStatus?: number; healthProbeError?: string }
type SessionState = { status: SessionStatus; accessToken?: string; user?: CurrentUser; error?: string; diagnostics: RuntimeDiagnostics; bootstrap: (initData: string) => Promise<void>; probeHealth: () => Promise<void>; setTelegramDiagnostics: (diagnostics: TelegramDiagnostics) => void; failInitialization: (error: unknown) => void; signOut: () => void }
const demoUser: CurrentUser = { id: 'demo', first_name: 'Гость', display_name: 'Гость', level: 1, xp: 0, current_streak: 0, timezone: 'UTC' }
const initialDiagnostics: RuntimeDiagnostics = { sdkLoaded: false, webAppDetected: false, initDataPresent: false, initDataLength: 0, apiBaseConfigured: isAPIBaseConfigured, apiBaseURL, authRequestURL, authRequestMethod: 'POST', authStarted: false, authCompleted: false, healthProbeURL, healthProbeStarted: false, healthProbeCompleted: false }
export const useSessionStore = create<SessionState>((set) => ({
  status: 'loading',
  diagnostics: initialDiagnostics,
  setTelegramDiagnostics(diagnostics) { set((state) => ({ diagnostics: { ...state.diagnostics, ...diagnostics } })) },
  failInitialization(error) { set((state) => ({ status: 'error', error: 'Ошибка инициализации Telegram', diagnostics: { ...state.diagnostics, authError: safeNetworkError(error) } })) },
  async probeHealth() {
    set((state) => ({ diagnostics: { ...state.diagnostics, healthProbeStarted: true, healthProbeCompleted: false, healthProbeHTTPStatus: undefined, healthProbeError: undefined } }))
    const result = await runHealthProbe()
    set((state) => ({ diagnostics: { ...state.diagnostics, healthProbeCompleted: result.completed, healthProbeHTTPStatus: result.status, healthProbeError: result.error } }))
  },
  async bootstrap(initData) {
    if (!initData) { set((state) => ({ status: 'demo', user: demoUser, diagnostics: { ...state.diagnostics, initDataPresent: false, initDataLength: 0 } })); return }
    set((state) => ({ diagnostics: { ...state.diagnostics, authStarted: true, authCompleted: false, authHTTPStatus: undefined, authError: undefined } }))
    try {
      const session = await authenticateTelegram(initData)
      sessionStorage.setItem('access_token', session.access_token)
      set((state) => ({ status: 'authenticated', accessToken: session.access_token, user: session.user, diagnostics: { ...state.diagnostics, authCompleted: true, authHTTPStatus: 200 } }))
    } catch (error) {
      set((state) => ({ status: 'error', error: error instanceof Error ? error.message : 'Ошибка авторизации', diagnostics: { ...state.diagnostics, authCompleted: true, authHTTPStatus: error instanceof APIError ? error.status : undefined, authError: error instanceof APIError ? `HTTPError: ${error.status} (${error.message})`.slice(0, 200) : safeNetworkError(error) } }))
    }
  },
  signOut() { sessionStorage.removeItem('access_token'); set({ status: 'demo', accessToken: undefined, user: demoUser, error: undefined }) },
}))
