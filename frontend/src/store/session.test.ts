// @vitest-environment jsdom
import { beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => ({ authenticateTelegram: vi.fn() }))
vi.mock('../api/auth', () => ({ authenticateTelegram: mocks.authenticateTelegram }))

import { APIError } from '../api/client'
import { APP_MODE_KEY, resolveAppMode, useSessionStore } from './session'

const user = { id: 'user-id', first_name: 'Test', display_name: 'Test', level: 1, xp: 0, current_streak: 0, timezone: 'UTC', role: 'user' as const, available_modes: ['student' as const] }

beforeEach(() => {
  mocks.authenticateTelegram.mockReset()
  sessionStorage.clear()
  localStorage.clear()
  useSessionStore.setState({ status: 'loading', accessToken: undefined, user: undefined, error: undefined, diagnostics: { sdkLoaded: false, webAppDetected: false, initDataPresent: false, initDataLength: 0, apiBaseConfigured: true, apiBaseURL: 'https://example.com/api/v1', authRequestURL: 'https://example.com/api/v1/auth/telegram', authRequestMethod: 'POST', authStarted: false, authCompleted: false, healthProbeURL: 'https://example.com/healthz', healthProbeStarted: false, healthProbeCompleted: false } })
})

describe('application mode',()=>{
  it('allows coach roles to restore coach mode',()=>expect(resolveAppMode({...user,role:'coach',available_modes:['student','coach']},'coach')).toBe('coach'))
  it('rejects stale coach mode after role removal',()=>expect(resolveAppMode(user,'coach')).toBe('student'))
  it('switches and persists an allowed mode',()=>{useSessionStore.setState({user:{...user,role:'coach',available_modes:['student','coach']},appMode:'student'});expect(useSessionStore.getState().setAppMode('coach')).toBe(true);expect(localStorage.getItem(APP_MODE_KEY)).toBe('coach')})
  it('does not trust local storage for an unavailable mode',()=>{useSessionStore.setState({user,appMode:'student'});expect(useSessionStore.getState().setAppMode('coach')).toBe(false);expect(useSessionStore.getState().appMode).toBe('student')})
})

describe('session bootstrap', () => {
  it('ends loading in demo mode when initData is empty', async () => {
    await useSessionStore.getState().bootstrap('')
    expect(useSessionStore.getState().status).toBe('demo')
    expect(mocks.authenticateTelegram).not.toHaveBeenCalled()
  })

  it('authenticates with non-empty initData and ends loading', async () => {
    mocks.authenticateTelegram.mockResolvedValue({ access_token: 'test-jwt', expires_in: 86400, user })
    await useSessionStore.getState().bootstrap('original-init-data')
    expect(mocks.authenticateTelegram).toHaveBeenCalledWith('original-init-data')
    expect(useSessionStore.getState()).toMatchObject({ status: 'authenticated', diagnostics: { authStarted: true, authCompleted: true, authHTTPStatus: 200 } })
  })

  it('ends loading in error state after an auth failure', async () => {
    mocks.authenticateTelegram.mockRejectedValue(new APIError('INVALID_TELEGRAM_INIT_DATA', 'Invalid data', 401))
    await useSessionStore.getState().bootstrap('invalid-init-data')
    expect(useSessionStore.getState()).toMatchObject({ status: 'error', diagnostics: { authStarted: true, authCompleted: true, authHTTPStatus: 401 } })
  })

  it('classifies an auth fetch network failure', async () => {
    mocks.authenticateTelegram.mockRejectedValue(new TypeError('Failed to fetch'))
    await useSessionStore.getState().bootstrap('original-init-data')
    expect(useSessionStore.getState()).toMatchObject({ status: 'error', diagnostics: { authStarted: true, authCompleted: true, authHTTPStatus: undefined, authError: 'TypeError: network failure (Failed to fetch)' } })
  })

  it('ends loading in error state after an initialization exception', () => {
    useSessionStore.getState().failInitialization(new Error('SDK failure'))
    expect(useSessionStore.getState()).toMatchObject({ status: 'error', diagnostics: { authStarted: false, authCompleted: false } })
  })
})
