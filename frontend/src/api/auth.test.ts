// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from 'vitest'

afterEach(() => { vi.unstubAllGlobals() })

describe('authenticateTelegram', () => {
  it('posts original initData to the configured auth route', async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, json: async () => ({ access_token: 'jwt', expires_in: 86400, user: {} }) })
    vi.stubGlobal('fetch', fetchMock)
    const { authenticateTelegram } = await import('./auth')
    await authenticateTelegram('original-init-data')
    expect(fetchMock).toHaveBeenCalledWith(expect.stringMatching(/\/auth\/telegram$/), expect.objectContaining({ method: 'POST', body: JSON.stringify({ init_data: 'original-init-data' }) }))
  })
})
