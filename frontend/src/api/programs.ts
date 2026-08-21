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
};
export const listPrograms = (token: string) =>
  api<{ programs: Program[] }>("/programs", {}, token);
export const getProgram = (token: string, id: string) =>
  api<{ program: Program }>(
    `/programs/${requireEntityID(id, "Program")}`,
    {},
    token,
  );
