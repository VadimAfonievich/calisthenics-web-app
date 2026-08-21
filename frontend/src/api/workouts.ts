import { api } from './client'
import { requireEntityID } from './entityIds'

export type WorkoutExercise={id:string;name:string;sets:number;target_reps?:number;target_duration_seconds?:number;rest_seconds:number}
export type WarmupSummary={id:string;title:string;estimated_minutes:number}
export type Workout={id:string;title:string;description:string;estimated_minutes:number;difficulty:string;program_id?:string;program_name?:string;category?:string;warmup_enabled?:boolean;warmup_workout_id?:string;default_warmup?:WarmupSummary;cover_media_url?:string;exercises:WorkoutExercise[]}
export type WorkoutCatalogItem={id:string;title:string;description:string;estimated_minutes:number;difficulty:string;exercise_count:number;program_id?:string;program_name?:string;category:string;warmup_enabled:boolean;status?:string;active_session_id?:string}
export type CompletedSet={exercise_id:string;set_number:number;reps?:number;duration_seconds?:number;completed:boolean}
export type WorkoutSession={id:string;workout_id:string;status:string;duration_seconds:number;xp_earned:number;current_streak:number;unlocked_achievements:string[];started_at:string;planned_workout_id?:string;session_purpose?:"main"|"warmup";follow_up_workout_id?:string;follow_up_workout_title?:string;follow_up_planned_workout_id?:string;continued_session_id?:string;next_session?:WorkoutSession}
export type ActiveWorkout={session:WorkoutSession;workout:Workout;completed_sets:CompletedSet[]}

export const listWorkouts=(t:string)=>api<{workouts:WorkoutCatalogItem[]}>('/workouts',{},t)
export const getWorkout=(t:string,id:string)=>api<{workout:Workout}>(`/workouts/${requireEntityID(id,'Workout')}`,{},t)
export const today=(t:string)=>api<{workout:Workout}>('/workouts/today',{},t)
export type StartWorkoutInput={planned_workout_id?:string;follow_up_workout_id?:string}
export const start=(t:string,id:string,input:StartWorkoutInput={})=>api<{session:WorkoutSession}>(`/workouts/${requireEntityID(id,'Workout')}/start`,{method:'POST',body:JSON.stringify({...(input.planned_workout_id?{planned_workout_id:requireEntityID(input.planned_workout_id,'Planned workout')}:{ }),...(input.follow_up_workout_id?{follow_up_workout_id:requireEntityID(input.follow_up_workout_id,'Follow-up workout')}:{ })})},t)
export const getSession=(t:string,id:string)=>api<ActiveWorkout>(`/workout-sessions/${requireEntityID(id,'Workout session')}`,{},t)
export const saveSet=(t:string,s:string,x:{exercise_id:string;set_number:number;reps?:number;duration_seconds?:number;completed:true})=>api<void>(`/workout-sessions/${requireEntityID(s,'Workout session')}/sets`,{method:'POST',body:JSON.stringify(x)},t)
export const complete=(t:string,s:string,d:number)=>api<{session:WorkoutSession}>(`/workout-sessions/${requireEntityID(s,'Workout session')}/complete?duration_seconds=${Math.max(0,Math.floor(d))}`,{method:'POST'},t)
