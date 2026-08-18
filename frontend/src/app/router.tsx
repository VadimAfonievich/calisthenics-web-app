import { Navigate, Route, Routes } from 'react-router-dom'
import { AppLayout } from './shell'
import { PlaceholderPage } from '../pages/placeholder'
import { HomePage } from '../pages/home'
import { LessonPage, LessonsPage } from '../pages/lessons'
import { ExercisePage, ExercisesPage } from '../pages/exercises'
import { TodayWorkoutPage } from '../pages/workout'
export function AppRouter() { return <Routes><Route element={<AppLayout />}><Route index element={<HomePage />} /><Route path="lessons" element={<LessonsPage />} /><Route path="lessons/:id" element={<LessonPage />} /><Route path="exercises" element={<ExercisesPage />} /><Route path="exercises/:id" element={<ExercisePage />} /><Route path="workout/today" element={<TodayWorkoutPage />} /><Route path="workout/:id" element={<PlaceholderPage title="Тренировка" text="Экран выполнения тренировки будет добавлен позже." />} /><Route path="progress" element={<PlaceholderPage title="Прогресс" text="Статистика и серии тренировок появятся после подключения прогресса." />} /><Route path="achievements" element={<PlaceholderPage title="Достижения" text="Здесь будут ваши награды." />} /><Route path="profile" element={<PlaceholderPage title="Профиль" text="Настройки профиля будут доступны после завершения авторизации." />} /><Route path="*" element={<Navigate to="/" replace />} /></Route></Routes> }
