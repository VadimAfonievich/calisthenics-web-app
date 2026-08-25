import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { getAchievements, getHistory, getProgress, getStats } from '../api/progress'
import { useSessionStore } from '../store/session'

const minutes = (seconds: number) => Math.round(seconds / 60)

export function ProgressPage() {
  const token = useSessionStore(state => state.accessToken)
  const progress = useQuery({ queryKey: ['progress'], queryFn: () => getProgress(token!), enabled: !!token })
  const stats = useQuery({ queryKey: ['stats'], queryFn: () => getStats(token!), enabled: !!token })
  const history = useQuery({ queryKey: ['history'], queryFn: () => getHistory(token!), enabled: !!token })
  const achievements = useQuery({ queryKey: ['achievements'], queryFn: () => getAchievements(token!), enabled: !!token })

  if (!token) return <div className="empty-state"><h2>Прогресс доступен в Telegram</h2></div>
  if (progress.isLoading || stats.isLoading || history.isLoading || achievements.isLoading)
    return <div className="unit notice skeleton">Собираем статистику…</div>
  if (!progress.data || !stats.data || !history.data || !achievements.data)
    return <div className="notice error">Не удалось загрузить прогресс.</div>

  const max = Math.max(1, ...stats.data.weeks.map(week => week.workouts))
  const unlocked = achievements.data.filter(item => item.unlocked).length
  const preview = achievements.data.slice(0, 4)

  return <div className="stack">
    <section className="hero-card">
      <p className="eyebrow">УРОВЕНЬ {progress.data.level.number}</p>
      <h2>{progress.data.level.name}</h2>
      <span>{progress.data.xp} XP · {progress.data.next_level ? `${progress.data.xp_to_next_level} XP до ${progress.data.next_level.name}` : 'максимальный уровень'}</span>
      <div className="progress-track"><i style={{ width: `${progress.data.level_progress * 100}%` }} /></div>
    </section>

    <section className="metric-grid">
      <article><b>🔥 {progress.data.current_streak}</b><span>текущая серия</span></article>
      <article><b>{progress.data.longest_streak}</b><span>лучшая серия</span></article>
      <article><b>{progress.data.total_workouts}</b><span>тренировок</span></article>
    </section>

    <section className="card">
      <p className="eyebrow">12 НЕДЕЛЬ</p><h3>Ритм тренировок</h3>
      <div className="week-chart">{stats.data.weeks.length ? stats.data.weeks.map(week =>
        <div key={week.week_start} title={`${week.workouts} тренировок`} style={{ height: `${Math.max(12, week.workouts / max * 100)}%` }} />
      ) : <p>Завершите первую тренировку — график появится здесь.</p>}</div>
    </section>

    <section className="metric-grid">
      <article><b>{progress.data.completed_exercises}</b><span>упражнений</span></article>
      <article><b>{minutes(progress.data.training_seconds)}</b><span>минут</span></article>
      <article><b>{history.data.length}</b><span>в истории</span></article>
    </section>

    <section className="card achievements-preview">
      <header><div><p className="eyebrow">КОЛЛЕКЦИЯ</p><h3>Достижения</h3></div><b>{unlocked} из {achievements.data.length}</b></header>
      <div className="achievements-preview-grid">
        {preview.map(item => <article className={`achievement-mini ${item.unlocked ? 'unlocked' : 'locked'}`} key={item.code}>
          <i>{item.icon}</i><span><b>{item.title}</b><small>{item.unlocked ? `Открыто · +${item.xp_reward} XP` : 'Пока не открыто'}</small></span>
        </article>)}
      </div>
      <Link className="text-link" to="/achievements">Все достижения →</Link>
    </section>
  </div>
}
