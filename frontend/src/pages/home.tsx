import {useQuery} from '@tanstack/react-query'
import {Link} from 'react-router-dom'
import {getToday,occurrenceRoute} from '../api/calendar'
import {listPrograms} from '../api/programs'
import {listSkills} from '../api/skills'
import {useSessionStore} from '../store/session'

export function HomePage(){
  const user=useSessionStore(s=>s.user),token=useSessionStore(s=>s.accessToken)
  const skills=useQuery({queryKey:['home-skills'],queryFn:()=>listSkills(token!),enabled:!!token})
  const today=useQuery({queryKey:['calendar-today'],queryFn:()=>getToday(token!),enabled:!!token})
  const programs=useQuery({queryKey:['home-programs'],queryFn:()=>listPrograms(token!),enabled:!!token})
  const active=skills.data?.skills.filter(x=>x.status!=='mastered').slice(0,3)??[],planned=today.data?.occurrences??[]
  return <div className="stack">
    <section className="hero-card"><p>Сегодня — хороший день, чтобы стать сильнее.</p><h2>Начните с базы</h2><span>Короткие уроки и понятные тренировки без лишнего давления.</span><Link className="primary-button" to="/lessons">Открыть уроки</Link></section>
    <section className="metric-grid"><article><b>{user?.xp??0}</b><span>XP</span></article><article><b>{user?.current_streak??0}</b><span>дней подряд</span></article><article><b>{user?.level??1}</b><span>уровень</span></article></section>
    <section className="card today-card"><p className="eyebrow">СЕГОДНЯ</p><h3>Тренировки</h3>{today.isLoading?<p>Загружаем расписание…</p>:planned.length?planned.slice(0,3).map(x=><div className="today-row" key={`${x.workout_id}-${x.schedule_id??x.planned_workout_id}`}><span><b>{x.time??'—'}</b>{x.workout_title}<small>{x.estimated_minutes} мин · {x.status==='completed'?'выполнено':x.status}</small></span>{x.status==='scheduled'&&<Link to={occurrenceRoute(x)}>Начать</Link>}</div>):<p>Сегодня тренировок нет</p>}{planned.length>0&&planned.every(x=>x.status==='completed')&&<p className="success">Все тренировки на сегодня выполнены</p>}<Link className="text-link" to="/calendar">Открыть календарь →</Link></section>
    <section className="card"><p className="eyebrow">ТРЕНИРОВКИ</p><h3>Выберите тренировку</h3><p>Откройте каталог, изучите план и начните тренировку, когда будете готовы.</p><Link className="text-link" to="/workouts">Перейти к тренировкам →</Link></section>
    <section className="card stack"><div className="section-heading"><div><p className="eyebrow">ПРОГРАММЫ</p><h3>Программы тренировок</h3></div><Link to="/programs">Все программы →</Link></div>{programs.data?.programs.slice(0,3).map(p=><Link className="path-skill" to={`/programs/${p.id}`} key={p.id}><span>{p.name}</span><b>{p.workout_count} трен.</b></Link>)}</section>
    <section className="card"><p className="eyebrow">ТВОЙ ПУТЬ</p><h3>Активные навыки</h3>{active.length?active.map(skill=><Link className="path-skill" to={`/skills/${skill.id}`} key={skill.id}><span>{skill.icon} {skill.name}</span><b>{skill.progress_percent}%</b></Link>):<p>Выберите первый навык на карте развития.</p>}<Link className="text-link" to="/skills">Открыть карту навыков →</Link></section>
  </div>
}
