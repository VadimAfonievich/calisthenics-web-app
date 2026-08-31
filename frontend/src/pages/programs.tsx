import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {useState} from "react";
import { Link, useParams } from "react-router-dom";
import { confirmStageMastery,getProgram, listPrograms, startProgram } from "../api/programs";
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
    client = useQueryClient(),
    valid = isValidEntityID(id),
    q = useQuery({
      queryKey: ["program", id],
      queryFn: () => getProgram(token!, id),
      enabled: !!token && valid,
    });
  const [confirmLevel,setConfirmLevel]=useState<string>();
  const start = useMutation({mutationFn:()=>startProgram(token!,id),onSuccess:()=>{client.invalidateQueries({queryKey:["program",id]});client.invalidateQueries({queryKey:["home-programs"]})}});
  const mastery=useMutation({mutationFn:(levelID:string)=>confirmStageMastery(token!,id,levelID),onSuccess:()=>{setConfirmLevel(undefined);client.invalidateQueries({queryKey:["program",id]});client.invalidateQueries({queryKey:["home-programs"]})}})
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
        mastery_type:"manual" as const,mastery_name:"Цель этапа",mastery_description:"Подтвердите, что вы освоили навык этого этапа.",
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
      {!p.progress_status && <button className="primary-button" disabled={start.isPending} onClick={()=>start.mutate()}>{start.isPending?"Запускаем…":"Начать программу"}</button>}
      {p.progress_status==="active"&&<p className="notice success">Программа активна · текущий этап: {p.current_stage}</p>}
      {p.progress_status==="completed"&&<p className="notice success">Программа завершена</p>}
      <h3>Этапы программы</h3>
      {levels.map((level, levelIndex) => (
        <section className={`stack card program-level ${level.status??""}`} key={level.id}>
          <header>
            <p className="eyebrow">ЭТАП {levelIndex + 1}</p>
            <h3>{level.title}</h3>
            <p>{level.description}</p>
            {level.status&&<b>{level.status==="completed"?"✓ Завершён":level.status==="current"?"→ Текущий этап":"🔒 Закрыт"}</b>}
          </header>
          <div className="stage-goal"><p className="eyebrow">ЦЕЛЬ ЭТАПА</p><h4>{level.mastery_name||level.title}</h4><p>{level.mastery_description||"Подтвердите, что вы освоили навык этого этапа."}</p>{level.mastery_type!=="manual"&&level.mastery_value&&<b>{level.mastery_value} {level.mastery_type==="duration"?"сек":"повторений"}</b>}<span>{level.mastered_at?"✓ Навык освоен":"Ещё не подтверждено"}</span></div>
          <p className="eyebrow">ТРЕНИРОВКИ ЭТАПА · {level.workouts.length} ДОСТУПНО</p>
          {level.workouts.map((w, workoutIndex) => (
            <Link className="workout-card" to={level.status==="locked"?"#":`/workouts/${w.id}`} aria-disabled={level.status==="locked"} key={w.id}>
              <p className="eyebrow">
                {workoutIndex + 1}. {difficulty[w.difficulty] ?? w.difficulty}
              </p>
              <h3>{w.title}</h3>
              <p>{w.description}</p>
              <footer><span>{w.estimated_minutes} мин</span>{level.status!=="locked"&&<b>Повторить тренировку</b>}</footer>
            </Link>
          ))}
          {level.status==="current"&&<section className="stage-mastery-action"><p className="eyebrow">ГОТОВ ПЕРЕЙТИ ДАЛЬШЕ?</p><p>Если ты уже выполняешь цель этапа, подтверди освоение навыка.</p><button type="button" onClick={()=>setConfirmLevel(level.id)}>Подтвердить навык</button></section>}
        </section>
      ))}
      {confirmLevel&&(()=>{const level=levels.find(x=>x.id===confirmLevel)!;return <div className="modal-backdrop" role="presentation"><section className="card confirm-dialog" role="alertdialog" aria-modal="true" aria-labelledby="mastery-confirm-title"><h3 id="mastery-confirm-title">Ты действительно освоил этот этап?</h3><p><b>Цель:</b><br/>{level.mastery_description}</p><p>После подтверждения откроется следующий этап.</p>{mastery.isError&&<p className="field-error">Не удалось подтвердить навык. Обновите страницу и попробуйте снова.</p>}<div className="form-actions"><button className="secondary-button" onClick={()=>setConfirmLevel(undefined)}>Пока нет</button><button disabled={mastery.isPending} onClick={()=>mastery.mutate(level.id)}>Да, навык освоен</button></div></section></div>})()}
    </div>
  );
}
