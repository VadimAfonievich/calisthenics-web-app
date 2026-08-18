import { create } from 'zustand'
import { authenticateTelegram, type CurrentUser } from '../api/auth'
type SessionStatus = 'loading' | 'authenticated' | 'demo' | 'error'
type SessionState = { status: SessionStatus; accessToken?: string; user?: CurrentUser; error?: string; bootstrap: (initData: string) => Promise<void>; signOut: () => void }
const demoUser: CurrentUser = { id: 'demo', first_name: 'Гость', display_name: 'Гость', level: 1, xp: 0, current_streak: 0, timezone: 'UTC' }
export const useSessionStore = create<SessionState>((set) => ({
  status: 'loading',
  async bootstrap(initData) { if (!initData) { set({ status: 'demo', user: demoUser }); return }; try { const session = await authenticateTelegram(initData); sessionStorage.setItem('access_token', session.access_token); set({ status: 'authenticated', accessToken: session.access_token, user: session.user }) } catch (error) { set({ status: 'error', error: error instanceof Error ? error.message : 'Ошибка авторизации' }) } },
  signOut() { sessionStorage.removeItem('access_token'); set({ status: 'demo', accessToken: undefined, user: demoUser, error: undefined }) },
}))
