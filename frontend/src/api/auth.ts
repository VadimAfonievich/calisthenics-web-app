import { api } from './client'
export type CurrentUser = { id: string; first_name: string; display_name: string; level: number; xp: number; current_streak: number; timezone: string }
type AuthResponse = { access_token: string; expires_in: number; user: CurrentUser }
export const authenticateTelegram = (initData: string) => api<AuthResponse>('/auth/telegram', { method: 'POST', body: JSON.stringify({ init_data: initData }) })
export const getCurrentUser = (accessToken: string) => api<{ user: CurrentUser }>('/me', {}, accessToken)
