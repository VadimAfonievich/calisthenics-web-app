import { healthProbeURL } from './client'

export type HealthProbeResult = { completed: true; status?: number; error?: string }

export const safeNetworkError = (error: unknown) => {
  if (error instanceof DOMException && error.name === 'AbortError') return 'AbortError: request aborted'
  if (error instanceof Error && error.name === 'TimeoutError') return 'TimeoutError: request timed out'
  if (error instanceof TypeError) return `TypeError: network failure (${error.message})`.slice(0, 200)
  if (error instanceof Error) return `${error.name}: ${error.message}`.slice(0, 200)
  return 'Unknown network error'
}

export async function runHealthProbe(url = healthProbeURL): Promise<HealthProbeResult> {
  try {
    const response = await fetch(url, { method: 'GET' })
    return { completed: true, status: response.status, ...(!response.ok ? { error: `HTTPError: ${response.status}` } : {}) }
  } catch (error) {
    return { completed: true, error: safeNetworkError(error) }
  }
}
