import { api } from './client'
export type AppMode = 'student' | 'coach' | 'admin'
export type UserRole = 'user' | 'coach' | 'admin' | 'super_admin'
export type TenantContext={id:string;slug:string;name:string;role:'coach'|'student';description?:string}
export type CurrentUser = { id: string; first_name: string; display_name: string; username?: string; photo_url?: string; level: number; xp: number; current_streak: number; timezone: string; role: UserRole; available_modes: AppMode[]; current_tenant?:TenantContext; tenants?:TenantContext[] }
type AuthResponse = { access_token: string; expires_in: number; user: CurrentUser }
export const authenticateTelegram = (initData: string) => api<AuthResponse>('/auth/telegram', { method: 'POST', body: JSON.stringify({ init_data: initData }) })
export const getCurrentUser = (accessToken: string) => api<{ user: CurrentUser }>('/me', {}, accessToken)
