const configuredAPIBaseURL = import.meta.env.VITE_API_URL ?? import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api/v1'
export const apiBaseURL = configuredAPIBaseURL.replace(/\/+$/, '')
export const isAPIBaseConfigured = Boolean(import.meta.env.VITE_API_URL ?? import.meta.env.VITE_API_BASE_URL)
export const buildAPIURL = (path: string, baseURL = apiBaseURL) => `${baseURL.replace(/\/+$/, '')}/${path.replace(/^\/+/, '')}`
export const authRequestURL = buildAPIURL('/auth/telegram')
export const healthProbeURL = new URL('/healthz', apiBaseURL).toString()
export class APIError extends Error { constructor(public readonly code: string, message: string, public readonly status: number) { super(message) } }
export async function api<T>(path: string, options: RequestInit = {}, accessToken?: string): Promise<T> {
  const response = await fetch(buildAPIURL(path), { ...options, headers: { 'Content-Type': 'application/json', ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}), ...options.headers } })
  if (!response.ok) { const body = await response.json().catch(() => null) as { error?: { code?: string; message?: string } } | null; throw new APIError(body?.error?.code ?? 'REQUEST_FAILED', body?.error?.message ?? 'Не удалось выполнить запрос', response.status) }
  return response.json() as Promise<T>
}
