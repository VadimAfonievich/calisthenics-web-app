// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from 'vitest'
import { runHealthProbe, safeNetworkError } from './networkDiagnostics'

afterEach(() => { vi.unstubAllGlobals() })

describe('health probe', () => {
  it('reports a successful HTTP response', async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200 })
    vi.stubGlobal('fetch', fetchMock)
    await expect(runHealthProbe('https://example.com/healthz')).resolves.toEqual({ completed: true, status: 200 })
    expect(fetchMock).toHaveBeenCalledWith('https://example.com/healthz', { method: 'GET' })
  })

  it('reports an HTTP failure', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 503 }))
    await expect(runHealthProbe('https://example.com/healthz')).resolves.toEqual({ completed: true, status: 503, error: 'HTTPError: 503' })
  })

  it('reports a fetch network failure safely', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))
    await expect(runHealthProbe('https://example.com/healthz')).resolves.toEqual({ completed: true, error: 'TypeError: network failure (Failed to fetch)' })
  })

  it('distinguishes an aborted request', () => {
    expect(safeNetworkError(new DOMException('Aborted', 'AbortError'))).toBe('AbortError: request aborted')
  })
})
