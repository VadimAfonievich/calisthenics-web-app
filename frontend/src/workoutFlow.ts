import type { CompletedSet, Workout, WorkoutExercise } from './api/workouts'

export const totalSets=(workout:Workout)=>workout.exercises.reduce((sum,x)=>sum+x.sets,0)
export const completedCount=(sets:CompletedSet[])=>sets.filter(x=>x.completed).length
export const isSetComplete=(sets:CompletedSet[],exerciseID:string,setNumber:number)=>sets.some(x=>x.completed&&x.exercise_id===exerciseID&&x.set_number===setNumber)
export function nextSet(workout:Workout,sets:CompletedSet[]){for(let exerciseIndex=0;exerciseIndex<workout.exercises.length;exerciseIndex++){const exercise=workout.exercises[exerciseIndex];for(let setNumber=1;setNumber<=exercise.sets;setNumber++)if(!isSetComplete(sets,exercise.id,setNumber))return{exercise,exerciseIndex,setNumber}}return null}
export const progressPercent=(workout:Workout,sets:CompletedSet[])=>{const total=totalSets(workout);return total?Math.round(completedCount(sets)/total*100):0}
export const elapsedSeconds=(startedAt:string,now=Date.now())=>Math.max(0,Math.floor((now-new Date(startedAt).getTime())/1000))
export const formatClock=(seconds:number)=>{const safe=Math.max(0,Math.floor(seconds));const h=Math.floor(safe/3600),m=Math.floor(safe%3600/60),s=safe%60;return h?`${String(h).padStart(2,'0')}:${String(m).padStart(2,'0')}:${String(s).padStart(2,'0')}`:`${String(m).padStart(2,'0')}:${String(s).padStart(2,'0')}`}
export const exerciseTarget=(exercise:WorkoutExercise)=>exercise.target_reps!==undefined?`${exercise.sets} × ${exercise.target_reps} повт.`:`${exercise.sets} × ${exercise.target_duration_seconds ?? 0} сек`
export const workoutSummary=(sets:CompletedSet[])=>({completed_sets:completedCount(sets),reps:sets.reduce((sum,x)=>sum+(x.reps??0),0),timed_seconds:sets.reduce((sum,x)=>sum+(x.duration_seconds??0),0)})
export const addRestSeconds=(remaining:number)=>Math.max(0,remaining)+15
export const skipRest=()=>0
export const actualTimedSetSeconds=(startedAt:number,now=Date.now())=>Math.max(0,Math.round((now-startedAt)/1000))
export const PREPARATION_SECONDS=5
export const isPreparationTime=(remaining:number)=>remaining<=PREPARATION_SECONDS
