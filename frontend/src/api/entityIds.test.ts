// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from 'vitest'
import { lessonRoute, workoutRoute } from './entityIds'
import { getLesson } from './lessons'
import { start } from './workouts'

const lessonID = '20000000-0000-0000-0000-000000000001'
const workoutID = '50000000-0000-0000-0000-000000000001'

afterEach(() => { vi.unstubAllGlobals() })

describe('entity routes', () => {
  it('builds a lesson card route with the UUID', () => {
    expect(lessonRoute(lessonID)).toBe(`/lessons/${lessonID}`)
  })

  it('builds a workout navigation route with the UUID', () => {
    expect(workoutRoute(workoutID)).toBe(`/workout/${workoutID}`)
  })

  it('rejects an undefined lesson id before fetch', async () => {
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    expect(() => getLesson('token', undefined as unknown as string)).toThrow('Lesson id is invalid')
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('uses the exact workout UUID in the start request', async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, json: async () => ({ session: { id: lessonID, workout_id: workoutID } }) })
    vi.stubGlobal('fetch', fetchMock)
    await start('token', workoutID)
    expect(fetchMock).toHaveBeenCalledWith(expect.stringMatching(new RegExp(`/workouts/${workoutID}/start$`)), expect.objectContaining({ method: 'POST' }))
  })

  it('rejects an undefined workout id before fetch', async () => {
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    expect(() => start('token', undefined as unknown as string)).toThrow('Workout id is invalid')
    expect(fetchMock).not.toHaveBeenCalled()
  })
})
