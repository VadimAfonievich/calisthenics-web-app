import { api } from './client'
export type Level={number:number;name:string;min_xp:number}
export type Progress={xp:number;level:Level;next_level?:Level;xp_to_next_level:number;level_progress:number;current_streak:number;longest_streak:number;total_workouts:number;completed_exercises:number;training_seconds:number}
export type Stats={total_workouts:number;completed_exercises:number;training_seconds:number;weeks:Array<{week_start:string;workouts:number;training_seconds:number}>}
export type HistoryItem={id:string;title:string;completed_at:string;duration_seconds:number;xp_earned:number}
export type Achievement={code:string;title:string;description:string;icon:string;xp_reward:number;unlocked:boolean;unlocked_at?:string}
export const getProgress=(token:string)=>api<Progress>('/progress',{},token)
export const getStats=(token:string)=>api<Stats>('/stats',{},token)
export const getHistory=(token:string)=>api<HistoryItem[]>('/history',{},token)
export const getAchievements=(token:string)=>api<Achievement[]>('/achievements',{},token)
