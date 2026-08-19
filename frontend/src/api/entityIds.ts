import { APIError } from './client'

const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i

export const isValidEntityID = (id: unknown): id is string => typeof id === 'string' && uuidPattern.test(id)

export function requireEntityID(id: unknown, entity: string): string {
  if (!isValidEntityID(id)) throw new APIError('INVALID_INPUT', `${entity} id is invalid`, 400)
  return id
}

export const lessonRoute = (id: unknown) => isValidEntityID(id) ? `/lessons/${id}` : null
export const workoutRoute = (id: unknown) => isValidEntityID(id) ? `/workouts/${id}` : null
