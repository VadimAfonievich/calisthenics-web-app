// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from 'vitest'
import { initializeTelegram, type TelegramWebApp } from './webapp'

afterEach(() => { delete window.Telegram })

describe('initializeTelegram', () => {
  it('reports a missing SDK and WebApp without throwing', () => {
    const result = initializeTelegram()
    expect(result).toMatchObject({ initData: '', diagnostics: { sdkLoaded: false, webAppDetected: false, initDataPresent: false, initDataLength: 0 } })
  })

  it('reports a loaded SDK with no WebApp', () => {
    window.Telegram = {}
    expect(initializeTelegram().diagnostics).toMatchObject({ sdkLoaded: true, webAppDetected: false, initDataPresent: false })
  })

  it('reads original initData and invokes ready and expand', () => {
    const ready = vi.fn(), expand = vi.fn()
    window.Telegram = { WebApp: { initData: 'query_id=original&hash=signed', ready, expand } }
    const result = initializeTelegram()
    expect(result.initData).toBe('query_id=original&hash=signed')
    expect(result.diagnostics).toMatchObject({ webAppDetected: true, initDataPresent: true, initDataLength: 29 })
    expect(ready).toHaveBeenCalledOnce()
    expect(expand).toHaveBeenCalledOnce()
  })

  it('continues with initData when an SDK method throws', () => {
    const webApp: TelegramWebApp = { initData: 'signed', ready: () => { throw new Error('unsupported') }, expand: vi.fn() }
    window.Telegram = { WebApp: webApp }
    const result = initializeTelegram()
    expect(result.initData).toBe('signed')
    expect(result.diagnostics.initializationError).toBe('ready: unsupported')
    expect(webApp.expand).toHaveBeenCalledOnce()
  })
})
