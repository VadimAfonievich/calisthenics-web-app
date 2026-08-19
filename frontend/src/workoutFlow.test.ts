import {describe,expect,it} from 'vitest'
import type {CompletedSet,Workout} from './api/workouts'
import {actualTimedSetSeconds,addRestSeconds,elapsedSeconds,exerciseTarget,nextSet,progressPercent,skipRest,totalSets,workoutSummary} from './workoutFlow'

const workout:Workout={id:'50000000-0000-0000-0000-000000000001',title:'Test',description:'Plan',estimated_minutes:30,difficulty:'beginner',program_id:'40000000-0000-0000-0000-000000000001',program_name:'Program',exercises:[{id:'30000000-0000-0000-0000-000000000001',name:'Reps',sets:2,target_reps:5,rest_seconds:30},{id:'30000000-0000-0000-0000-000000000002',name:'Timer',sets:1,target_duration_seconds:20,rest_seconds:45}]}
const first:CompletedSet={exercise_id:workout.exercises[0].id,set_number:1,reps:4,completed:true}

describe('workout flow',()=>{
  it('renders reps and timed targets for preview',()=>{expect(exerciseTarget(workout.exercises[0])).toBe('2 × 5 повт.');expect(exerciseTarget(workout.exercises[1])).toBe('1 × 20 сек')})
  it('advances to the next set and exercise',()=>{expect(nextSet(workout,[])?.setNumber).toBe(1);expect(nextSet(workout,[first])?.setNumber).toBe(2);expect(nextSet(workout,[first,{...first,set_number:2}])?.exercise.name).toBe('Timer')})
  it('rest controls skip and add 15 seconds',()=>{expect(skipRest()).toBe(0);expect(addRestSeconds(30)).toBe(45)})
  it('tracks total set progress and completion',()=>{expect(totalSets(workout)).toBe(3);expect(progressPercent(workout,[first])).toBe(33);expect(nextSet(workout,[first,{...first,set_number:2},{exercise_id:workout.exercises[1].id,set_number:1,duration_seconds:18,completed:true}])).toBeNull()})
  it('restores elapsed time from backend started_at',()=>{expect(elapsedSeconds('2026-01-01T00:00:00Z',Date.parse('2026-01-01T00:01:05Z'))).toBe(65)})
  it('records actual early timed-set duration',()=>{expect(actualTimedSetSeconds(1000,13500)).toBe(13)})
  it('builds completion totals without calculating XP',()=>{expect(workoutSummary([first,{exercise_id:workout.exercises[1].id,set_number:1,duration_seconds:18,completed:true}])).toEqual({completed_sets:2,reps:4,timed_seconds:18})})
})
