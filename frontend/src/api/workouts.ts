import { api } from './client'
export type Workout={id:string;title:string;description:string;estimated_minutes:number;exercises:Array<{id:string;name:string;sets:number;target_reps?:number;target_duration_seconds?:number;rest_seconds:number}>}
export const today=(t:string)=>api<{workout:Workout}>('/workouts/today',{},t)
export const start=(t:string,id:string)=>api<{session:{id:string}}> (`/workouts/${id}/start`,{method:'POST'},t)
export const saveSet=(t:string,s:string,x:object)=>api<void>(`/workout-sessions/${s}/sets`,{method:'POST',body:JSON.stringify(x)},t)
export const complete=(t:string,s:string,d:number)=>api<{session:{xp_earned:number}}>(`/workout-sessions/${s}/complete?duration_seconds=${d}`,{method:'POST'},t)
