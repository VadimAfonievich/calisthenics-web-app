import { api } from './client'
import { requireEntityID } from './entityIds'
export type Workout={id:string;title:string;description:string;estimated_minutes:number;exercises:Array<{id:string;name:string;sets:number;target_reps?:number;target_duration_seconds?:number;rest_seconds:number}>}
export const today=(t:string)=>api<{workout:Workout}>('/workouts/today',{},t)
export const start=(t:string,id:string)=>api<{session:{id:string;workout_id:string}}> (`/workouts/${requireEntityID(id,'Workout')}/start`,{method:'POST'},t)
export const saveSet=(t:string,s:string,x:object)=>api<void>(`/workout-sessions/${requireEntityID(s,'Workout session')}/sets`,{method:'POST',body:JSON.stringify(x)},t)
export const complete=(t:string,s:string,d:number)=>api<{session:{xp_earned:number;current_streak:number;unlocked_achievements:string[]}}>(`/workout-sessions/${requireEntityID(s,'Workout session')}/complete?duration_seconds=${d}`,{method:'POST'},t)
