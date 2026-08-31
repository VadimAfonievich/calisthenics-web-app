import { api } from "./client";
import { requireEntityID } from "./entityIds";
export type ProgramWorkout = {
  id: string;
  title: string;
  description: string;
  estimated_minutes: number;
  difficulty: string;
  category: string;
};
export type Program = {
  id: string;
  name: string;
  slug: string;
  description: string;
  difficulty: string;
  duration_weeks: number;
  category: string;
  cover_media_url?: string;
  workout_count: number;
  workouts?: ProgramWorkout[];
  levels?: ProgramLevel[];
  progress_status?: "active" | "completed";
  current_level?: number;
  current_stage?: string;
  next_workout?: ProgramWorkout;
};
export type ProgramLevel = {
  id: string;
  level_number: number;
  title: string;
  description: string;
  difficulty: string;
  unlock_rule_type: string;
  unlock_rule_value: number;
  workouts: ProgramWorkout[];
  status?: "completed" | "current" | "locked";
  mastery_type?:"duration"|"reps"|"manual";
  mastery_value?:number;
  mastery_name?:string;
  mastery_description?:string;
  mastered_at?:string;
};
export const listPrograms = (token: string) =>
  api<{ programs: Program[] }>("/programs", {}, token);
export const getProgram = (token: string, id: string) =>
  api<{ program: Program }>(
    `/programs/${requireEntityID(id, "Program")}`,
    {},
    token,
  );
export const startProgram = (token: string, id: string) =>
  api<{ progress: { program_id: string; status: string; current_level: number } }>(
    `/programs/${requireEntityID(id, "Program")}/start`,
    { method: "POST" },
    token,
  );
export const confirmStageMastery=(token:string,programID:string,levelID:string)=>api<{mastery:{program_id:string;program_level_id:string;status:string;current_level:number;mastered_at:string;program_completed:boolean}}>(`/programs/${requireEntityID(programID,"Program")}/levels/${requireEntityID(levelID,"Program level")}/mastery`,{method:"POST"},token)
