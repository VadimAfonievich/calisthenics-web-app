import { api } from "./client";
import { requireEntityID } from "./entityIds";
export type CoachRole = "coach" | "admin" | "super_admin";
export type Dashboard = {
  lessons: number;
  lessons_published: number;
  exercises: number;
  workouts: number;
  workouts_published: number;
  programs: number;
  skills: number;
  media: number;
};
export type ContentKind =
  "lessons" | "exercises" | "workouts" | "programs" | "skills";
export type ContentItem = {
  id: string;
  name: string;
  slug?: string;
  status: "draft" | "published" | "archived";
  difficulty: string;
  owner_user_id?: string;
  updated_at: string;
};
export type LessonBlock = {
  type:
    | "heading"
    | "text"
    | "image"
    | "video"
    | "tip"
    | "warning"
    | "checklist"
    | "divider";
  text?: string;
  media_id?: string;
  items?: string[];
};
export type LessonInput = {
  category_id: string;
  title: string;
  slug?: string;
  short_description: string;
  content: string;
  difficulty: string;
  duration_minutes: number;
  cover_media_id?: string;
  blocks: LessonBlock[];
};
export type BuilderExercise = {
  exercise_id: string;
  sets: number;
  target_reps?: number;
  target_duration_seconds?: number;
  rest_seconds: number;
  notes?: string;
  sort_order: number;
};
export type BuilderLevel = {
  level_number: number;
  title: string;
  description: string;
  difficulty: string;
  unlock_rule_type: string;
  unlock_rule_value: number;
  criterion_type: string;
  criterion_value: number;
  program_level_id?: string;
  sort_order: number;
};
export type BuilderInput = {
  name?: string;
  code?: string;
  title?: string;
  description: string;
  difficulty: string;
  duration_weeks?: number;
  estimated_minutes?: number;
  day_number?: number;
  program_id?: string;
  program_level_id?: string;
  movement_type?: string;
  instructions?: string;
  common_mistakes?: string;
  coach_tips?: string;
  muscle_groups?: string[];
  equipment?: string[];
  category?: string;
  map_group?: "basic" | "floor" | "bar" | "parallel_bars";
  icon?: string;
  xp_reward?: number;
  final_criterion_type?: string;
  final_criterion_value?: number;
  cover_media_id?: string;
  exercises?: BuilderExercise[];
  levels?: BuilderLevel[];
  requirements?: string[];
  workouts?: ProgramWorkout[];
  sort_order?: number;
  warmup_enabled?: boolean;
  warmup_workout_id?: string;
};
export type ProgramWorkout = { workout_id: string; sort_order: number };
export type Option = {
  id: string;
  name: string;
  status?: "draft" | "published" | "archived";
  difficulty?: string;
  minutes?: number;
  sort_order?: number;
  owner_user_id?: string;
  parent_id?: string;
};
export type CoachOptions = {
  categories: Option[];
  exercises: Option[];
  programs: Option[];
  program_levels: Option[];
  workouts: Option[];
  warmups: Option[];
  skills: Option[];
  media: Option[];
};
export type MediaAsset = {
  id: string;
  type: "image" | "video";
  url: string;
  thumbnail_url?: string;
  original_filename: string;
  mime_type: string;
  size_bytes: number;
  references: number;
  status: string;
};
export const coachMe = (t: string) =>
    api<{ role: CoachRole }>("/coach/me", {}, t),
  dashboard = (t: string) => api<Dashboard>("/coach/dashboard", {}, t),
  analytics = (t: string) =>
    api<Record<string, unknown>>("/coach/analytics", {}, t),
  content = (t: string, kind: string, search = "", status = "") =>
    api<{ items: ContentItem[] }>(
      `/coach/${kind}?${new URLSearchParams({ search, status })}`,
      {},
      t,
    ),
  contentDetail = (t: string, kind: ContentKind, id: string) =>
    api<{ item: Record<string, unknown> }>(
      `/coach/${kind}/${requireEntityID(id, "Content")}`,
      {},
      t,
    ),
  coachOptions = (t: string) => api<CoachOptions>("/coach/options", {}, t),
  createLesson = (t: string, x: LessonInput) =>
    api<{ id: string }>(
      "/coach/lessons",
      { method: "POST", body: JSON.stringify(x) },
      t,
    ),
  updateLesson = (t: string, id: string, x: LessonInput) =>
    api<{ id: string }>(
      `/coach/lessons/${requireEntityID(id, "Lesson")}`,
      { method: "PUT", body: JSON.stringify(x) },
      t,
    ),
  saveBuilder = (
    t: string,
    kind: Exclude<ContentKind, "lessons">,
    x: BuilderInput,
    id?: string,
  ) =>
    api<{ id: string }>(
      id
        ? `/coach/${kind}/${requireEntityID(id, "Content")}`
        : `/coach/${kind}`,
      { method: id ? "PUT" : "POST", body: JSON.stringify(x) },
      t,
    ),
  action = (
    t: string,
    kind: string,
    id: string,
    name: "publish" | "unpublish" | "archive" | "restore" | "duplicate",
  ) =>
    api<{ id: string; status?: string }>(
      `/coach/${kind}/${requireEntityID(id, "Content")}/${name}`,
      { method: "POST" },
      t,
    ),
  media = (t: string) => api<{ media: MediaAsset[] }>("/coach/media", {}, t),
  addExternalMedia = (
    t: string,
    x: {
      type: "image" | "video";
      url: string;
      original_filename: string;
      mime_type: string;
      size_bytes: number;
    },
  ) =>
    api<{ media: MediaAsset }>(
      "/coach/media",
      { method: "POST", body: JSON.stringify(x) },
      t,
    ),
  deleteMedia = (t: string, id: string) =>
    api<void>(
      `/coach/media/${requireEntityID(id, "Media")}`,
      { method: "DELETE" },
      t,
    );
