import { api } from './client'
export type Exercise={id:string;name:string;description:string;instructions:string;common_mistakes:string;coach_tips?:string;cover_media_url?:string;difficulty:string;muscle_groups:string[];equipment:string[];tags:string[]}
export type ExerciseFilters={difficulty?:string;muscle_group?:string;movement_type?:string;equipment?:string;tag?:string}
export const listExercises=(t:string,filters:ExerciseFilters={})=>api<{exercises:Exercise[]}>(`/exercises?${new URLSearchParams(Object.entries(filters).filter((entry):entry is [string,string]=>!!entry[1]))}`,{},t)
export const getExercise=(t:string,id:string)=>api<{exercise:Exercise}>(`/exercises/${id}`,{},t)
