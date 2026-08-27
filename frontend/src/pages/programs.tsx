import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { getProgram, listPrograms } from "../api/programs";
import { isValidEntityID } from "../api/entityIds";
import { useSessionStore } from "../store/session";
const difficulty: Record<string, string> = {
  beginner: "Начальный",
  intermediate: "Средний",
  advanced: "Продвинутый",
};
export function ProgramsPage() {
  const token = useSessionStore((s) => s.accessToken),
    q = useQuery({
      queryKey: ["programs"],
      queryFn: () => listPrograms(token!),
      enabled: !!token,
    });
  if (q.isLoading)
    return <p className="notice skeleton">Загружаем программы…</p>;
  if (q.isError)
    return <p className="notice error">Не удалось загрузить программы.</p>;
  return (
    <div className="stack">
      <h2>Программы тренировок</h2>
      {q.data?.programs.map((p) => (
        <Link className="workout-card" to={`/programs/${p.id}`} key={p.id}>
          <h3>{p.name}</h3>
          <p>{p.description}</p>
          <footer>
            <span>{difficulty[p.difficulty] ?? p.difficulty}</span>
            <b>{p.workout_count} тренировок</b>
          </footer>
        </Link>
      ))}
    </div>
  );
}
export function ProgramDetailPage() {
  const { id = "" } = useParams(),
    token = useSessionStore((s) => s.accessToken),
    valid = isValidEntityID(id),
    q = useQuery({
      queryKey: ["program", id],
      queryFn: () => getProgram(token!, id),
      enabled: !!token && valid,
    });
  if (!valid) return <p className="notice error">Некорректная программа.</p>;
  if (q.isLoading)
    return <p className="notice skeleton">Загружаем программу…</p>;
  if (q.isError || !q.data)
    return <p className="notice error">Программа не найдена.</p>;
  const p = q.data.program;
  const levels = p.levels?.length
    ? p.levels
    : [{
        id: "legacy",
        level_number: 1,
        title: "Тренировки",
        description: "Тренировки программы",
        difficulty: p.difficulty,
        unlock_rule_type: "none",
        unlock_rule_value: 0,
        workouts: p.workouts ?? [],
      }];
  return (
    <div className="stack">
      <Link className="text-link" to="/programs">
        ← Все программы
      </Link>
      <section className="hero-card">
        <p className="eyebrow">{difficulty[p.difficulty] ?? p.difficulty}</p>
        <h2>{p.name}</h2>
        <span>{p.description}</span>
        <b>{p.workout_count} тренировок</b>
      </section>
      <h3>Этапы программы</h3>
      {levels.map((level, levelIndex) => (
        <section className="stack card" key={level.id}>
          <header>
            <p className="eyebrow">ЭТАП {levelIndex + 1}</p>
            <h3>{level.title}</h3>
            <p>{level.description}</p>
          </header>
          {level.workouts.map((w, workoutIndex) => (
            <Link className="workout-card" to={`/workouts/${w.id}`} key={w.id}>
              <p className="eyebrow">
                {workoutIndex + 1}. {difficulty[w.difficulty] ?? w.difficulty}
              </p>
              <h3>{w.title}</h3>
              <p>{w.description}</p>
              <footer><span>{w.estimated_minutes} мин</span></footer>
            </Link>
          ))}
        </section>
      ))}
    </div>
  );
}
