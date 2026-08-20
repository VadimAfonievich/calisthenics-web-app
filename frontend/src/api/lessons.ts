import { api } from './client'
import { requireEntityID } from './entityIds'
export type LessonBlock = { type:'heading'|'text'|'image'|'video'|'tip'|'warning'|'checklist'|'divider'; text?:string; url?:string; mime_type?:string; alt?:string; items?:string[] }
export type Lesson = { id:string; category_id:string; category_name:string; title:string; short_description:string; content:string; content_blocks:LessonBlock[]; cover_media_url?:string; difficulty:string; duration_minutes:number; completed:boolean; progress_percent:number }
export const listLessons=(token:string)=>api<{lessons:Lesson[]}>('/lessons',{},token)
export const getLesson=(token:string,id:string)=>api<{lesson:Lesson}>(`/lessons/${requireEntityID(id,'Lesson')}`,{},token)
export const completeLesson=(token:string,id:string)=>api<{xp_earned:number;total_xp:number;already_completed:boolean}>(`/lessons/${requireEntityID(id,'Lesson')}/complete`,{method:'POST'},token)
