import { Navigate, Route, Routes } from "react-router-dom";
import type { ReactNode } from "react";
import { AppLayout } from "./shell";
import { HomePage } from "../pages/home";
import { LessonPage, LessonsPage } from "../pages/lessons";
import { ExercisePage, ExercisesPage } from "../pages/exercises";
import {
  WorkoutCatalogPage,
  WorkoutPlayerPage,
  WorkoutPreviewPage,
} from "../pages/workout";
import { ProgressPage } from "../pages/progress";
import { AchievementsPage } from "../pages/achievements";
import { AdminPage } from "../pages/admin";
import { SkillDetailPage, SkillsPage } from "../pages/skills";
import { CalendarPage, SchedulesPage } from "../pages/calendar";
import {
  CoachAnalyticsPage,
  CoachDashboardPage,
  CoachMediaPage,
} from "../pages/coach";
import { ProfilePage } from "../pages/profile";
import { CoachGuard } from "./coachGuard";
import {
  CoachContentHome,
  CoachContentList,
  CoachEditor,
} from "../pages/coachAuthoring";
import { ProgramsPage, ProgramDetailPage } from "../pages/programs";
const guarded = (page: ReactNode) => <CoachGuard>{page}</CoachGuard>;
export function AppRouter() {
  return (
    <Routes>
      <Route element={<AppLayout />}>
        <Route index element={<HomePage />} />
        <Route path="programs" element={<ProgramsPage />} />
        <Route path="programs/:id" element={<ProgramDetailPage />} />
        <Route path="lessons" element={<LessonsPage />} />
        <Route path="lessons/:id" element={<LessonPage />} />
        <Route path="exercises" element={<ExercisesPage />} />
        <Route path="exercises/:id" element={<ExercisePage />} />
        <Route path="workouts" element={<WorkoutCatalogPage />} />
        <Route path="workouts/:id" element={<WorkoutPreviewPage />} />
        <Route path="workout-session/:id" element={<WorkoutPlayerPage />} />
        <Route
          path="workout/today"
          element={<Navigate to="/workouts" replace />}
        />
        <Route path="calendar" element={<CalendarPage />} />
        <Route path="calendar/schedules" element={<SchedulesPage />} />
        <Route path="skills" element={<SkillsPage />} />
        <Route path="skills/:id" element={<SkillDetailPage />} />
        <Route path="progress" element={<ProgressPage />} />
        <Route path="achievements" element={<AchievementsPage />} />
        <Route path="coach" element={guarded(<CoachDashboardPage />)} />
        <Route path="coach/content" element={guarded(<CoachContentHome />)} />
        <Route
          path="coach/content/:kind"
          element={guarded(<CoachContentList />)}
        />
        <Route path="coach/:kind/new" element={guarded(<CoachEditor />)} />
        <Route path="coach/:kind/:id/edit" element={guarded(<CoachEditor />)} />
        <Route path="coach/media" element={guarded(<CoachMediaPage />)} />
        <Route
          path="coach/analytics"
          element={guarded(<CoachAnalyticsPage />)}
        />
        <Route path="admin" element={<AdminPage />} />
        <Route path="profile" element={<ProfilePage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  );
}
