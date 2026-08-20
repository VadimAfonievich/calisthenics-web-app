import { api } from './client'
export type Exercise={id:string;name:string;description:string;instructions:string;common_mistakes:string;coach_tips?:string;cover_media_url?:string;difficulty:string;muscle_groups:string[];equipment:string[]}
export const listExercises=(t:string)=>api<{exercises:Exercise[]}>('/exercises',{},t)
export const getExercise=(t:string,id:string)=>api<{exercise:Exercise}>(`/exercises/${id}`,{},t)
