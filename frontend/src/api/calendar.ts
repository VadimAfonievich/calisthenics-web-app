import {api} from './client'
import {requireEntityID} from './entityIds'

export type CalendarStatus='scheduled'|'completed'|'skipped'|'cancelled'|'missed'
export type CalendarOccurrence={date:string;time?:string;workout_id:string;workout_title:string;schedule_id?:string;planned_workout_id?:string;status:CalendarStatus;completed_session_id?:string;category:string;difficulty:string;estimated_minutes:number;duration_seconds?:number;xp_earned?:number;completed_at?:string}
export type Schedule={id:string;workout_id:string;workout_title:string;weekdays:number[];preferred_time?:string;timezone:string;start_date:string;end_date?:string;active:boolean}
export type ScheduleInput={workout_id:string;weekdays:number[];preferred_time?:string;start_date:string;end_date?:string;timezone?:string;active?:boolean}
export type PlannedWorkout={id:string;workout_id:string;workout_title:string;scheduled_date:string;scheduled_time?:string;timezone:string;source_schedule_id?:string;status:CalendarStatus}
export type PlannedInput={workout_id:string;scheduled_date:string;scheduled_time?:string;source_schedule_id?:string}
export const getCalendar=(t:string,from:string,to:string)=>api<{occurrences:CalendarOccurrence[]}>(`/calendar?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`,{},t)
export const getToday=(t:string)=>api<{occurrences:CalendarOccurrence[]}>('/calendar/today',{},t)
export const listSchedules=(t:string)=>api<{schedules:Schedule[]}>('/training-schedules',{},t)
export const createSchedule=(t:string,x:ScheduleInput)=>api<{schedule:Schedule}>('/training-schedules',{method:'POST',body:JSON.stringify(x)},t)
export const updateSchedule=async(t:string,id:string,x:ScheduleInput)=>{await api<{schedule:Schedule}>(`/training-schedules/${requireEntityID(id,'Schedule')}`,{method:'PUT',body:JSON.stringify(x)},t)}
export const disableSchedule=(t:string,id:string)=>api<void>(`/training-schedules/${requireEntityID(id,'Schedule')}`,{method:'DELETE'},t)
export const createPlanned=(t:string,x:PlannedInput)=>api<{planned_workout:PlannedWorkout}>('/planned-workouts',{method:'POST',body:JSON.stringify(x)},t)
export const updatePlanned=(t:string,id:string,x:PlannedInput)=>api<{planned_workout:PlannedWorkout}>(`/planned-workouts/${requireEntityID(id,'Planned workout')}`,{method:'PUT',body:JSON.stringify(x)},t)
export const cancelPlanned=(t:string,id:string)=>api<void>(`/planned-workouts/${requireEntityID(id,'Planned workout')}`,{method:'DELETE'},t)
export const skipPlanned=(t:string,id:string)=>api<void>(`/planned-workouts/${requireEntityID(id,'Planned workout')}/skip`,{method:'POST'},t)
export const monthRange=(value:Date)=>{const first=new Date(value.getFullYear(),value.getMonth(),1),last=new Date(value.getFullYear(),value.getMonth()+1,0);return{from:localDate(first),to:localDate(last)}}
export const localDate=(d:Date)=>`${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`
export const monthCells=(value:Date)=>{const {from,to}=monthRange(value),first=new Date(`${from}T12:00:00`),last=new Date(`${to}T12:00:00`),lead=(first.getDay()+6)%7;return[...Array(lead).fill(null),...Array(last.getDate()).keys()].map((x,i)=>i<lead?null:new Date(value.getFullYear(),value.getMonth(),Number(x)+1))}
export const occurrenceRoute=(x:CalendarOccurrence)=>`/workouts/${requireEntityID(x.workout_id,'Workout')}?${new URLSearchParams({scheduled_date:x.date,...(x.time?{scheduled_time:x.time}:{}),...(x.schedule_id?{schedule_id:x.schedule_id}:{}),...(x.planned_workout_id?{planned_workout_id:x.planned_workout_id}:{})})}`
