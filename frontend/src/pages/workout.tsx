import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Link,
  useNavigate,
  useParams,
  useSearchParams,
} from "react-router-dom";
import {
  complete,
  getSession,
  getWorkout,
  listWorkouts,
  normalizeWorkoutDuration,
  saveSet,
  start,
  type CompletedSet,
  type Workout,
  type WorkoutCatalogItem,
  type WorkoutExercise,
  type WorkoutSession,
} from "../api/workouts";
import { ExerciseDemoMedia } from "../components/ExerciseDemoMedia";
import { createPlanned } from "../api/calendar";
import { isValidEntityID, workoutRoute } from "../api/entityIds";
import {
  actualTimedSetSeconds,
  addRestSeconds,
  completedCount,
  elapsedSeconds,
  exerciseTarget,
  formatClock,
  isPreparationTime,
  nextSet,
  progressPercent,
  totalSets,
  workoutSummary,
} from "../workoutFlow";
import { useSessionStore } from "../store/session";
import { listPrograms } from "../api/programs";
import { listSkills } from "../api/skills";
import {VoiceCoach,saveVoiceCoachPreference,voiceCoachPreference} from "../voiceCoach";

const Notice = () => (
  <div className="empty-state">
    <h2>Тренировки доступны в Telegram</h2>
  </div>
);

export function WorkoutPlayerRoute() {
  const {id=""}=useParams();
  return <WorkoutPlayerPage key={id}/>;
}
export const canonicalCategory = (value: string) =>
  ({
    WARMUP: "warmup",
    MORNING_ROUTINE: "morning",
    BASE_STRENGTH: "strength",
    SKILL: "skill",
    MOBILITY: "warmup",
    OTHER: "strength",
  })[value] ?? value;
const categoryLabels: Record<string, string> = {
  warmup: "Разминка",
  morning: "Зарядки",
  strength: "Развитие силы",
  skill: "Тренировка навыков",
};

export function WorkoutCatalogPage() {
  const token = useSessionStore((s) => s.accessToken),
    [params, setParams] = useSearchParams(),
    selected = params.get("category") ?? "";
  const query = useQuery({
    queryKey: ["workouts"],
    queryFn: () => listWorkouts(token!),
    enabled: !!token,
  });
  const programs = useQuery({
    queryKey: ["programs"],
    queryFn: () => listPrograms(token!),
    enabled: !!token,
  });
  const skills = useQuery({
    queryKey: ["skills-preview"],
    queryFn: () => listSkills(token!),
    enabled: !!token,
  });
  if (!token) return <Notice />;
  if (query.isLoading)
    return <div className="notice skeleton">Загружаем тренировки…</div>;
  if (query.isError)
    return (
      <div className="notice error">
        Не удалось загрузить каталог тренировок.
      </div>
    );
  const visible = (query.data?.workouts ?? []).filter(
    (w) => !selected || canonicalCategory(w.category) === selected,
  );
  const grouped = visible.reduce<Record<string, WorkoutCatalogItem[]>>(
    (out, w) => {
      (out[canonicalCategory(w.category || "strength")] ??= []).push(w);
      return out;
    },
    {},
  );
  return (
    <div className="stack">
      <div>
        <p className="eyebrow">ТРЕНИРОВОЧНЫЙ ПУТЬ</p>
        <h2>Тренировки</h2>
      </div>
      <section className="training-categories">
        <button
          className={!selected ? "active" : ""}
          onClick={() => setParams({})}
        >
          Все
        </button>
        {Object.entries(categoryLabels).map(([key, label]) => (
          <button
            className={selected === key ? "active" : ""}
            onClick={() => setParams({ category: key })}
            key={key}
          >
            {label}
          </button>
        ))}
      </section>
      {!selected && <section className="stack prominent-programs">
        <div className="section-heading"><h3>Программы</h3><Link to="/programs">Все программы →</Link></div>
        {programs.data?.programs.slice(0,3).map(p=><Link className="workout-card" to={`/programs/${p.id}`} key={p.id}><h3>{p.name}</h3><p>{p.description}</p><footer><b>{p.workout_count} тренировок</b></footer></Link>)}
      </section>}
      {!Object.keys(grouped).length && (
        <div className="empty-state compact">
          <h3>Опубликованных тренировок пока нет</h3>
          <p>Сохранённые черновики появятся здесь после публикации тренером.</p>
        </div>
      )}
      {Object.entries(grouped).map(([category, items]) => (
        <section className="stack workout-category" key={category}>
          <h3>{categoryLabels[category] ?? categoryLabels.OTHER}</h3>
          {items.map((w) => {
            const route =
              w.status === "started" && isValidEntityID(w.active_session_id)
                ? `/workout-session/${w.active_session_id}`
                : workoutRoute(w.id);
            return route ? (
              <Link className="workout-card" to={route} key={w.id}>
                <div>
                  <p className="eyebrow">
                    {w.program_name ? `${w.program_name} · ` : ""}{w.difficulty}
                  </p>
                  <h3>{w.title}</h3>
                  <p>{w.description}</p>
                </div>
                <footer>
                  <span>{w.estimated_minutes} мин</span>
                  <span>{w.exercise_count} упр.</span>
                  {w.status && (
                    <b>{w.status === "started" ? "Продолжить" : w.status}</b>
                  )}
                </footer>
              </Link>
            ) : (
              <div className="notice error" key={w.title}>
                Тренировка «{w.title}» недоступна.
              </div>
            );
          })}
        </section>
      ))}
      {!selected && (
        <>
          <section className="stack">
            <div className="section-heading">
              <h3>Освоение элементов</h3>
              <Link to="/skills">Карта навыков →</Link>
            </div>
            {skills.data?.skills.slice(0, 3).map((skill) => (
              <Link
                className={`skill-node ${skill.status}`}
                to={`/skills/${skill.id}`}
                key={skill.id}
              >
                <span>{skill.icon}</span>
                <div>
                  <h3>{skill.name}</h3>
                  <small>
                    {skill.progress_percent}% ·{" "}
                    {skill.status === "locked" ? "Закрыт" : "Доступен"}
                  </small>
                </div>
              </Link>
            ))}
          </section>
        </>
      )}
    </div>
  );
}

export function WorkoutPreviewPage() {
  const { id = "" } = useParams(),
    token = useSessionStore((s) => s.accessToken),
    navigate = useNavigate(),
    [params] = useSearchParams(),
    valid = isValidEntityID(id);
  const query = useQuery({
    queryKey: ["workout", id],
    queryFn: () => getWorkout(token!, id),
    enabled: !!token && valid,
  });
  const begin = useMutation({
    mutationFn: async ({workoutID,next,usePlanned}:{workoutID:string;next?:string;usePlanned:boolean}) => {
      let planned = params.get("planned_workout_id") ?? undefined;
      const date = params.get("scheduled_date"),
        schedule = params.get("schedule_id") ?? undefined;
      if ((usePlanned || next) && !planned && date) {
        const result = await createPlanned(token!, {
          workout_id: next ?? workoutID,
          scheduled_date: date,
          scheduled_time: params.get("scheduled_time") ?? undefined,
          source_schedule_id: schedule,
        });
        planned = result.planned_workout.id;
      }
      return start(token!, workoutID, {
        ...(usePlanned && planned ? {planned_workout_id:planned} : {}),
        ...(next ? {follow_up_workout_id:next, ...(planned?{planned_workout_id:planned}:{})} : {}),
      });
    },
    onSuccess: (result) => navigate(`/workout-session/${result.session.id}`),
  });
  if (!token) return <Notice />;
  if (!valid)
    return (
      <div className="notice error">Некорректный идентификатор тренировки.</div>
    );
  if (query.isLoading)
    return <div className="notice skeleton">Загружаем план…</div>;
  if (query.isError || !query.data)
    return <div className="notice error">Тренировка не найдена.</div>;
  const w = query.data.workout;
  return (
    <div className="stack workout-preview">
      <Link className="text-link" to="/workouts">
        ← Все тренировки
      </Link>
      {w.cover_media_url && (
        <img className="lesson-cover" src={w.cover_media_url} alt="" />
      )}
      <section className="hero-card">
        <p className="eyebrow">
          {w.program_name ? `${w.program_name} · ` : ""}{w.difficulty}
        </p>
        <h2>{w.title}</h2>
        <span>{w.description}</span>
        <div className="workout-meta">
          <b>{w.estimated_minutes} мин</b>
          <b>{w.exercises.length} упражнений</b>
        </div>
      </section>
      <Link className="text-link" to={`/calendar?workout=${w.id}`}>
        Добавить в расписание →
      </Link>
      {w.warmup_enabled && w.default_warmup && (
        <section className="card warmup-choice stack">
          <div><p className="eyebrow">ПЕРЕД ТРЕНИРОВКОЙ</p><h3>{w.default_warmup.title}</h3><span>Разминка · ~{w.default_warmup.estimated_minutes} минут</span></div>
          <button className="primary-button" disabled={begin.isPending} onClick={()=>begin.mutate({workoutID:w.default_warmup!.id,next:w.id,usePlanned:false})}>Начать с разминки</button>
          <button disabled={begin.isPending} onClick={()=>begin.mutate({workoutID:w.id,usePlanned:true})}>Начать без разминки</button>
        </section>
      )}
      <section className="stack">
        <h3>План тренировки</h3>
        {w.exercises.map((x, i) => (
          <article className="plan-row" key={x.id}>
            <i>{i + 1}</i>
            <div>
              <h3>{x.name}</h3>
              <p>
                {exerciseTarget(x)} · Отдых {x.rest_seconds} сек
              </p>
            </div>
          </article>
        ))}
      </section>
      {!(w.warmup_enabled && w.default_warmup) && <button
        className="primary-button workout-cta"
        disabled={begin.isPending}
        onClick={() => begin.mutate({workoutID:w.id,usePlanned:true})}
      >
        {begin.isPending ? "Начинаем…" : "Начать тренировку"}
      </button>}
      {begin.isError && (
        <p className="notice error">
          {begin.error instanceof Error
            ? begin.error.message
            : "Не удалось начать тренировку."}
        </p>
      )}
    </div>
  );
}

type Phase = "intro" | "prepare" | "set" | "timer" | "rest" | "done" | "transition";
export function WorkoutPlayerPage() {
  const { id = "" } = useParams(),
    token = useSessionStore((s) => s.accessToken),
    qc = useQueryClient(),
    navigate = useNavigate(),
    valid = isValidEntityID(id);
  const query = useQuery({
    queryKey: ["workout-session", id],
    queryFn: () => getSession(token!, id),
    enabled: !!token && valid,
    retry: false,
  });
  const [dataSets, setDataSets] = useState<CompletedSet[]>([]),
    [phase, setPhase] = useState<Phase>("intro"),
    [now, setNow] = useState(Date.now()),
    [remaining, setRemaining] = useState(0),
    [rest, setRest] = useState(0),
    [actualReps, setActualReps] = useState(0),
    [voices,setVoices]=useState<SpeechSynthesisVoice[]>([]),
    [voiceEnabled,setVoiceEnabled]=useState(voiceCoachPreference);
  const timerStarted = useRef(0);
  const autoSaved = useRef("");
  const phaseDeadline=useRef(0);
  const voice=useRef(new VoiceCoach(voiceEnabled));
  useEffect(()=>()=>voice.current.cancel(),[]);
  useEffect(()=>{if(!voice.current.isSupported())return;const refresh=()=>setVoices(voice.current.voices());refresh();window.speechSynthesis.addEventListener?.("voiceschanged",refresh);return()=>window.speechSynthesis.removeEventListener?.("voiceschanged",refresh)},[]);
  useEffect(() => {
    const i = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(i);
  }, []);
  useEffect(() => {
    if (query.data) {
      setDataSets(query.data.completed_sets);
      if(query.data.session.status!=="completed"&&phaseDeadline.current===0){
        const first=nextSet(query.data.workout,query.data.completed_sets);
        if(first){voice.current.announceSessionStart(query.data.workout.category,first.exercise,query.data.session.id);setRemaining(4);phaseDeadline.current=Date.now()+4000;setPhase("intro");}
      }
    }
  }, [query.data]);
  const current = useMemo(
    () => (query.data ? nextSet(query.data.workout, dataSets) : null),
    [query.data, dataSets],
  );
  useEffect(() => {
    if (current?.exercise.target_reps !== undefined)
      setActualReps(current.exercise.target_reps);
  }, [current?.exercise.id, current?.setNumber]);
  useEffect(() => {
    if ((phase !== "timer" && phase !== "rest" && phase !== "prepare" && phase !== "intro") || !phaseDeadline.current) return;
    const calculated=Math.max(0,Math.ceil((phaseDeadline.current-now)/1000));
    const next=phase==="intro"?Math.min(4,calculated):phase==="prepare"?Math.min(5,calculated):calculated;
    setRemaining(current=>current===next?current:next);
  }, [phase,now]);
  useEffect(() => {
    if (phase === "timer" && isPreparationTime(remaining) && timerStarted.current) {
      window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred?.(
        "success",
      );
    }
    if (phase === "timer" && remaining>0 && remaining<=5 && current) voice.current.countdown(`ending:${current.exercise.id}:${current.setNumber}`,remaining);
    if (phase === "rest" && isPreparationTime(remaining)) {
      voice.current.cancel();
      window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred?.("success");
      setRemaining(5); setPhase("prepare");
    }
    if (phase === "intro" && remaining===0) {voice.current.cancel();setRemaining(5);phaseDeadline.current=Date.now()+5000;setPhase("prepare");}
    if (phase === "prepare" && remaining>0 && remaining<=5) voice.current.preparationCountdown(`prepare:${current?.exercise.id}:${current?.setNumber}`,remaining);
    if (phase === "prepare" && remaining===0 && current) {
      if(current.exercise.target_duration_seconds!==undefined){
        voice.current.announceStart(`${current.exercise.id}:${current.setNumber}`); const seconds=current.exercise.target_duration_seconds; timerStarted.current=Date.now(); phaseDeadline.current=Date.now()+seconds*1000; setRemaining(seconds); setPhase("timer");
      } else {
        voice.current.announceStart(`${current.exercise.id}:${current.setNumber}`);
        setPhase("set");
      }
    }
  }, [phase, remaining, current]);
  const record = useMutation({
    mutationFn: (payload: { reps?: number; duration_seconds?: number }) =>
      saveSet(token!, id, {
        exercise_id: current!.exercise.id,
        set_number: current!.setNumber,
        ...payload,
        completed: true,
      }),
    onSuccess: (_, payload) => {
      voice.current.cancel();
      const completedEvent=`${id}:${current!.exercise.id}:${current!.setNumber}`;
      voice.current.announceFinished(completedEvent);
      const item = {
        exercise_id: current!.exercise.id,
        set_number: current!.setNumber,
        ...payload,
        completed: true,
      };
      const updated = [
        ...dataSets.filter(
          (x) =>
            !(
              x.exercise_id === item.exercise_id &&
              x.set_number === item.set_number
            ),
        ),
        item,
      ];
      setDataSets(updated);
      timerStarted.current = 0;
      autoSaved.current = "";
      const final =
        completedCount(dataSets) + 1 >= totalSets(query.data!.workout);
      if (final) {voice.current.announceCompletion(query.data!.workout.category,!!query.data!.session.follow_up_workout_id,completedEvent);setPhase("done");}
      else if (current!.exercise.rest_seconds > 0) {
        const upcoming=nextSet(query.data!.workout,updated);
        const nextExercise=upcoming?.exercise.id!==current!.exercise.id?upcoming?.exercise:undefined;
        voice.current.announceTransition(current!.exercise.rest_seconds,nextExercise,completedEvent);
        setRest(current!.exercise.rest_seconds);
        setRemaining(current!.exercise.rest_seconds);
        phaseDeadline.current=Date.now()+current!.exercise.rest_seconds*1000;
        setPhase("rest");
      } else {
        const upcoming=nextSet(query.data!.workout,updated);
        if (upcoming && upcoming.exercise.id!==current!.exercise.id) {voice.current.announceTransition(0,upcoming.exercise,completedEvent);setRemaining(4);phaseDeadline.current=Date.now()+4000;setPhase("intro");}
        else {setRemaining(5); phaseDeadline.current=Date.now()+5000; setPhase("prepare");}
      }
    },
  });
  useEffect(()=>{
    if(phase!=="timer"||remaining!==0||!timerStarted.current||!current||record.isPending)return;
    const key=`${current.exercise.id}:${current.setNumber}`;
    if(autoSaved.current===key)return;
    autoSaved.current=key;
    record.mutate({duration_seconds:current.exercise.target_duration_seconds??actualTimedSetSeconds(timerStarted.current)});
  },[phase,remaining,current,record.isPending]);
  const finish = useMutation({
    mutationFn: () =>
      complete(token!, id, normalizeWorkoutDuration(elapsedSeconds(query.data!.session.started_at))),
    onSuccess: ({session:completed}) => {
      qc.invalidateQueries({ queryKey: ["progress"] });
      qc.invalidateQueries({ queryKey: ["workouts"] });
      if(completed.next_session?.id||completed.continued_session_id){
        phaseDeadline.current=0;setPhase("transition");
      }
    },
  });
  if (!token) return <Notice />;
  if (!valid)
    return (
      <div className="notice error">Некорректный идентификатор сессии.</div>
    );
  if (query.isLoading)
    return <div className="notice skeleton">Восстанавливаем тренировку…</div>;
  if (query.isError || !query.data)
    return (
      <div className="notice error">
        Активная тренировка не найдена или принадлежит другому пользователю.
      </div>
    );
  const { workout, session } = query.data,
    total = totalSets(workout),
    done = completedCount(dataSets),
    elapsed = normalizeWorkoutDuration(elapsedSeconds(session.started_at, now)),
    summary = workoutSummary(dataSets),
    continuesToWorkout = !!session.follow_up_workout_id;
  const continuation=finish.data?.session.next_session?.id??finish.data?.session.continued_session_id??session.continued_session_id;
  if ((phase==="transition"||session.status==="completed") && continuation)
    return <WarmupTransition sessionID={continuation} title={finish.data?.session.follow_up_workout_title??session.follow_up_workout_title??"Основная тренировка"} onContinue={(mainSessionID)=>navigate(`/workout-session/${mainSessionID}`)}/>;
  if (finish.isSuccess)
    return (
      <WorkoutComplete
        title={workout.title}
        duration={finish.data.session.duration_seconds}
        exercises={workout.exercises.length}
        total={total}
        summary={summary}
        session={finish.data.session}
      />
    );
  if (session.status === "completed")
    return (
      <div className="empty-state">
        <div className="empty-icon">✓</div>
        <h2>Тренировка уже завершена</h2>
        <Link className="primary-button" to="/workouts">
          К тренировкам
        </Link>
      </div>
    );
  const early = () => {
    if (
      done === total ||
      window.confirm(
        done === 0
          ? "Нет завершённых подходов. Завершить тренировку?"
          : `Выполнено ${done} из ${total} подходов. Завершить досрочно?`,
      )
    )
      finish.mutate();
  };
  return (
    <div className="workout-player stack">
      <header>
        <div>
          <p className="eyebrow">АКТИВНАЯ ТРЕНИРОВКА</p>
          <h2>{workout.title}</h2>
        </div>
        <b>{formatClock(elapsed)}</b>
      </header>
      <div className="voice-controls">
        <button onClick={()=>{const enabled=!voiceEnabled;setVoiceEnabled(enabled);saveVoiceCoachPreference(enabled);voice.current.setEnabled(enabled);}} aria-pressed={voiceEnabled}>🔊 Голос: {voiceEnabled?"Вкл":"Выкл"}</button>
        <button disabled={!voice.current.isSupported()||!voiceEnabled} onClick={()=>voice.current.test()}>Проверить голос</button>
        {voices.length>0&&<label>Голос<select aria-label="Голос" defaultValue={voice.current.selectedVoice()?.voiceURI} onChange={(e)=>{voice.current.setVoice(e.target.value);voice.current.test()}}>{voices.map(v=><option key={v.voiceURI} value={v.voiceURI}>{v.name}</option>)}</select></label>}
        {!voice.current.isSupported()&&<small>Голосовые подсказки недоступны на этом устройстве.</small>}
      </div>
      <div className="progress-track">
        <i style={{ width: `${progressPercent(workout, dataSets)}%` }} />
      </div>
      <div className="workout-stats">
        <span>
          Упражнение{" "}
          {(current?.exerciseIndex ?? workout.exercises.length - 1) + 1} /{" "}
          {workout.exercises.length}
        </span>
        <span>
          Подходы {done} / {total}
        </span>
        <span>{progressPercent(workout, dataSets)}%</span>
      </div>
      {phase === "rest" && current ? (
        <RestView
          seconds={remaining}
          original={rest}
          next={current.exercise}
          exerciseFinished={current.setNumber === 1}
          onSkip={() => {
            voice.current.cancel();
            phaseDeadline.current=0;
            setRemaining(0);
            setRemaining(5);phaseDeadline.current=Date.now()+5000;setPhase("prepare");
          }}
          onAdd={() => {const next=addRestSeconds(remaining);phaseDeadline.current=Date.now()+next*1000;setRemaining(next)}}
        />
      ) : phase === "done" || !current ? (
        <section className="player-focus">
          <p className="eyebrow">ПЛАН ВЫПОЛНЕН</p>
          <h2>Все подходы завершены</h2>
          <button
            className="primary-button"
            disabled={finish.isPending}
            onClick={() => finish.mutate()}
          >
            {finish.isPending
              ? continuesToWorkout ? "Переходим…" : "Завершаем…"
              : continuesToWorkout ? "Перейти к тренировке" : "Завершить тренировку"}
          </button>
        </section>
      ) : (
        <SetView
          current={current}
          phase={phase}
          remaining={remaining}
          reps={actualReps}
          pending={record.isPending}
          onMinus={() => setActualReps((v) => Math.max(0, v - 1))}
          onPlus={() => setActualReps((v) => v + 1)}
          onStart={() => {
            voice.current.cancel();
            timerStarted.current = Date.now();
            const seconds=current.exercise.target_duration_seconds ?? 0;
            phaseDeadline.current=Date.now()+seconds*1000;
            setRemaining(seconds);
            setPhase("timer");
          }}
          onComplete={() =>
            record.mutate(
              current.exercise.target_reps !== undefined
                ? { reps: actualReps }
                : {
                    duration_seconds: actualTimedSetSeconds(
                      timerStarted.current,
                    ),
                  },
            )
          }
        />
      )}{" "}
      {(record.isError || finish.isError) && (
        <p className="notice error">
          {(record.error ?? finish.error) instanceof Error
            ? (record.error ?? finish.error)?.message
            : "Не удалось сохранить результат."}
        </p>
      )}
      <WorkoutOutline
        workout={workout}
        sets={dataSets}
        currentID={current?.exercise.id}
      />
      {phase !== "done" && (
        <button
          className="danger-link"
          disabled={finish.isPending}
          onClick={early}
        >
          Завершить досрочно
        </button>
      )}
    </div>
  );
}

function SetView({
  current,
  phase,
  remaining,
  reps,
  pending,
  onMinus,
  onPlus,
  onStart,
  onComplete,
}: {
  current: {
    exercise: WorkoutExercise;
    exerciseIndex: number;
    setNumber: number;
  };
  phase: Phase;
  remaining: number;
  reps: number;
  pending: boolean;
  onMinus: () => void;
  onPlus: () => void;
  onStart: () => void;
  onComplete: () => void;
}) {
  const x = current.exercise;
  return (
    <section className="player-focus">
      <p className="eyebrow">УПРАЖНЕНИЕ {current.exerciseIndex + 1}</p>
      <h2>{x.name}</h2>
      <ExerciseDemoMedia compact media={x.demo_media_url?{url:x.demo_media_url,type:x.demo_media_type||"video",mime_type:x.demo_media_mime_type||"video/mp4",poster_url:x.demo_poster_url}:undefined}/>
      {x.notes&&<p className="lesson-callout"><b>Подсказка тренера:</b> {x.notes}</p>}
      <p>
        Подход {current.setNumber} из {x.sets}
      </p>
      {phase==="intro" ? <span>Слушайте следующую подсказку…</span> : phase==="prepare" ? <><span>Приготовьтесь…</span><strong className="countdown">{remaining}</strong></> : x.target_reps !== undefined ? (
        <>
          <span>Цель: {x.target_reps} повторений</span>
          <div className="rep-control">
            <button onClick={onMinus} disabled={pending}>
              −
            </button>
            <b>{reps}</b>
            <button onClick={onPlus} disabled={pending}>
              +
            </button>
          </div>
          <button
            className="primary-button"
            disabled={pending}
            onClick={onComplete}
          >
            {pending ? "Сохраняем…" : "Готово"}
          </button>
        </>
      ) : (
        <>
          <span>Цель: {x.target_duration_seconds} секунд</span>
          {phase === "timer" && (
            <strong className="countdown">{formatClock(remaining)}</strong>
          )}
          {phase === "timer" ? (
            <button
              className="primary-button"
              disabled={pending}
              onClick={onComplete}
            >
              {pending
                ? "Сохраняем…"
                : remaining === 0
                  ? "Завершить подход"
                  : "Завершить раньше"}
            </button>
          ) : <span>Приготовьтесь…</span>}
        </>
      )}
    </section>
  );
}
export function WarmupTransition({sessionID,title,onContinue}:{sessionID:string;title:string;onContinue:(sessionID:string)=>void|Promise<void>}) {
  const [starting,setStarting]=useState(false),[error,setError]=useState("");
  const transitionStarted=useRef(false);
  const continueToMainWorkout=useCallback(async()=>{
    if(transitionStarted.current)return;
    transitionStarted.current=true;setStarting(true);setError("");
    try{await onContinue(sessionID)}catch{
      transitionStarted.current=false;setStarting(false);setError("Не удалось открыть основную тренировку.");
    }
  },[onContinue,sessionID]);
  useEffect(()=>{void continueToMainWorkout()},[continueToMainWorkout]);
  return <section className="player-focus warmup-transition"><p className="eyebrow">РАЗМИНКА ЗАВЕРШЕНА</p><h2>{title}</h2><p>{starting?"Открываем основную тренировку…":"Готово к переходу"}</p>{error&&<p className="notice error">{error}</p>}<button className="primary-button" disabled={starting} onClick={()=>void continueToMainWorkout()}>{error?"Повторить":"Открываем…"}</button></section>;
}
function RestView({
  seconds,
  next,
  exerciseFinished,
  onSkip,
  onAdd,
}: {
  seconds: number;
  original: number;
  next: WorkoutExercise;
  exerciseFinished: boolean;
  onSkip: () => void;
  onAdd: () => void;
}) {
  return (
    <section className="player-focus rest">
      <p className="eyebrow">
        {exerciseFinished ? "УПРАЖНЕНИЕ ЗАВЕРШЕНО" : "ОТДЫХ"}
      </p>
      <strong className="countdown">{formatClock(seconds)}</strong>
      <p>Следующий: {next.name}</p>
      <div className="action-row">
        <button onClick={onSkip}>Пропустить отдых</button>
        <button onClick={onAdd}>+15 сек</button>
      </div>
    </section>
  );
}
function WorkoutOutline({
  workout,
  sets,
  currentID,
}: {
  workout: Workout;
  sets: CompletedSet[];
  currentID?: string;
}) {
  return (
    <details className="card workout-outline">
      <summary>План тренировки</summary>
      {workout.exercises.map((x) => {
        const done = sets.filter(
          (s) => s.exercise_id === x.id && s.completed,
        ).length;
        return (
          <p key={x.id}>
            <b>{done === x.sets ? "✓" : x.id === currentID ? "●" : "○"}</b>{" "}
            {x.name}{" "}
            <span>
              {done}/{x.sets}
            </span>
          </p>
        );
      })}
    </details>
  );
}
function WorkoutComplete({
  title,
  duration,
  exercises,
  total,
  summary,
  session,
}: {
  title: string;
  duration: number;
  exercises: number;
  total: number;
  summary: ReturnType<typeof workoutSummary>;
  session: WorkoutSession;
}) {
  const partial = summary.completed_sets < total;
  return (
    <div className="empty-state workout-summary">
      <div className="empty-icon">✓</div>
      <h2>
        {partial ? "Тренировка завершена досрочно" : "Тренировка завершена"}
      </h2>
      <p>{title}</p>
      <div className="summary-grid">
        <span>
          <b>{formatClock(duration)}</b>Время
        </span>
        <span>
          <b>{exercises}</b>Упражнений
        </span>
        <span>
          <b>
            {summary.completed_sets}/{total}
          </b>
          Подходов
        </span>
        <span>
          <b>{summary.reps}</b>Повторений
        </span>
        <span>
          <b>{summary.timed_seconds}</b>Секунд
        </span>
        <span>
          <b>+{session.xp_earned}</b>XP
        </span>
      </div>
      <p>🔥 Серия: {session.current_streak} дней</p>
      {session.unlocked_achievements.map((x) => (
        <span key={x}>🏆 {x}</span>
      ))}
      <Link className="primary-button" to="/workouts">
        К тренировкам
      </Link>
    </div>
  );
}
