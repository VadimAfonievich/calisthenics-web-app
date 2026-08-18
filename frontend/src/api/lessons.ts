import { api } from './client'
export type Lesson = { id:string; category_name:string; title:string; short_description:string; content:string; difficulty:string; duration_minutes:number; completed:boolean; progress_percent:number }
export const listLessons=(token:string)=>api<{lessons:Lesson[]}>('/lessons',{},token)
export const getLesson=(token:string,id:string)=>api<{lesson:Lesson}>(`/lessons/${id}`,{},token)
export const completeLesson=(token:string,id:string)=>api<{xp_earned:number;total_xp:number;already_completed:boolean}>(`/lessons/${id}/complete`,{method:'POST'},token)
