import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import {
  completeSkillLevel,
  confirmSkillCriterion,
  getSkill,
  getSkillMap,
  masterSkill,
  type Skill,
} from "../api/skills";
import { isValidEntityID } from "../api/entityIds";
import { useSessionStore } from "../store/session";

const statusLabels = {
  locked: "Закрыт",
  available: "Доступен",
  in_progress: "В процессе",
  mastered: "Освоен",
};
const skillGroups = [
  { id: "basic", title: "Базовые навыки", description: "Фундамент силы, мобильности и контроля тела" },
  { id: "floor", title: "Пол", description: "Упоры, балансы и элементы без снарядов" },
  { id: "bar", title: "Турник", description: "Тяговые и силовые элементы на перекладине" },
  { id: "parallel_bars", title: "Брусья", description: "Упоры и силовые элементы на брусьях" },
] as const;
const criterion = (type: string, value: number) =>
  type === "duration_seconds"
    ? `${value} сек`
    : type === "repetitions"
      ? `${value} повторений`
      : type === "workout_completed"
        ? `${value} тренировка`
        : "Подтверждение";
const SkillNotice = () => (
  <div className="empty-state">
    <h2>Навыки доступны в Telegram</h2>
  </div>
);
export function SkillsPage() {
  const token = useSessionStore((s) => s.accessToken);
  const query = useQuery({
    queryKey: ["skill-map"],
    queryFn: () => getSkillMap(token!),
    enabled: !!token,
  });
  if (!token) return <SkillNotice />;
  if (query.isLoading)
    return <div className="notice skeleton">Строим карту навыков…</div>;
  if (query.isError || !query.data)
    return (
      <div className="notice error">Не удалось загрузить карту навыков.</div>
    );
  const byID = new Map(query.data.nodes.map((x) => [x.id, x]));
  const requirementsBySkill = new Map<string,string[]>();
  query.data.requirements.forEach((requirement) => {
    requirementsBySkill.set(requirement.skill_id,[
      ...(requirementsBySkill.get(requirement.skill_id) ?? []),
      requirement.required_skill_id,
    ]);
  });
  const depthOf = (skillID:string, trail=new Set<string>()):number => {
    if(trail.has(skillID)) return 0;
    const parents=requirementsBySkill.get(skillID) ?? [];
    if(!parents.length) return 0;
    const nextTrail=new Set(trail).add(skillID);
    return 1+Math.max(...parents.map(parent=>depthOf(parent,nextTrail)));
  };
  return (
    <div className="stack">
      <div>
        <p className="eyebrow">КАРТА РАЗВИТИЯ</p>
        <h2>Навыки</h2>
      </div>
      <div className="skill-map">
        {skillGroups.map(group => {
          const skills=query.data.nodes.filter(skill=>(skill.map_group ?? "basic")===group.id);
          if(!skills.length)return null;
          return <section className="skill-group" key={group.id}>
            <header><p className="eyebrow">{group.title}</p><small>{group.description}</small></header>
            <div className="skill-group-tree">
              {skills.map(skill => {
                const parents=requirementsBySkill.get(skill.id) ?? [];
                return <div key={skill.id} className={`skill-branch depth-${Math.min(depthOf(skill.id),3)}`}>
                  {parents.length>0&&<i className="skill-line" />}
                  <SkillNode skill={skill} />
                  {parents.length>0&&<small>После: {parents.map(parent=>byID.get(parent)?.name).filter(Boolean).join(" + ")}</small>}
                </div>;
              })}
            </div>
          </section>;
        })}
      </div>
    </div>
  );
}
function SkillNode({ skill }: { skill: Skill }) {
  const content = (
    <>
      <span>{skill.icon}</span>
      <div>
        <p className="eyebrow">{skill.difficulty}</p>
        <h3>{skill.name}</h3>
        <div className="progress-track">
          <i style={{ width: `${skill.progress_percent}%` }} />
        </div>
        <small>
          {statusLabels[skill.status]} · {skill.progress_percent}%
        </small>
      </div>
    </>
  );
  return (
    <Link className={`skill-node ${skill.status}`} to={`/skills/${skill.id}`}>
      {content}
    </Link>
  );
}
export function SkillDetailPage() {
  const { id = "" } = useParams(),
    token = useSessionStore((s) => s.accessToken),
    client = useQueryClient(),
    valid = isValidEntityID(id);
  const query = useQuery({
    queryKey: ["skill", id],
    queryFn: () => getSkill(token!, id),
    enabled: !!token && valid,
  });
  const complete = useMutation({
    mutationFn: ({ level, value }: { level: number; value: number }) =>
      completeSkillLevel(token!, id, level, value),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: ["skill", id] });
      client.invalidateQueries({ queryKey: ["skill-map"] });
    },
  });
  const master = useMutation({
    mutationFn: () =>
      masterSkill(token!, id, query.data!.skill.final_criterion_value),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: ["skill", id] });
      client.invalidateQueries({ queryKey: ["skill-map"] });
      client.invalidateQueries({ queryKey: ["progress"] });
    },
  });
  const confirmCriterion = useMutation({
    mutationFn: (criterionID: string) =>
      confirmSkillCriterion(token!, id, criterionID),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: ["skill", id] });
      client.invalidateQueries({ queryKey: ["skill-map"] });
      client.invalidateQueries({ queryKey: ["progress"] });
    },
  });
  if (!token) return <SkillNotice />;
  if (!valid)
    return (
      <div className="notice error">Некорректный идентификатор навыка.</div>
    );
  if (query.isLoading)
    return <div className="notice skeleton">Загружаем прогрессию…</div>;
  if (query.isError || !query.data)
    return <div className="notice error">Навык не найден.</div>;
  const { skill, levels, criteria = [] } = query.data,
    allComplete =
      levels.length > 0 && levels.every((x) => x.status === "completed");
  return (
    <div className="stack skill-detail">
      <Link className="text-link" to="/skills">
        ← Карта навыков
      </Link>
      {skill.cover_media_url && (
        <img className="lesson-cover" src={skill.cover_media_url} alt="" />
      )}
      <section className="hero-card">
        <span className="skill-icon">{skill.icon}</span>
        <p className="eyebrow">
          {skill.difficulty} · {statusLabels[skill.status]}
        </p>
        <h2>{skill.name}</h2>
        <span>{skill.description}</span>
        <div className="progress-track">
          <i style={{ width: `${skill.progress_percent}%` }} />
        </div>
        <b>{skill.progress_percent}%</b>
      </section>
      {criteria.length > 0 && (
        <section className="stack base-criteria">
          <h3>Ваши показатели</h3>
          <p>
            {criteria.filter((x) => x.confirmed).length} из {criteria.length}{" "}
            выполнено
          </p>
          <div className="progress-track">
            <i
              style={{
                width: `${Math.round((criteria.filter((x) => x.confirmed).length * 100) / criteria.length)}%`,
              }}
            />
          </div>
          {criteria.map((item) => (
            <article
              className={`criterion-row ${item.confirmed ? "completed" : ""}`}
              key={item.id}
            >
              <span>{item.confirmed ? "✓" : "○"}</span>
              <b>{item.title}</b>
              {!item.confirmed && (
                <button
                  disabled={confirmCriterion.isPending}
                  onClick={() => confirmCriterion.mutate(item.id)}
                >
                  Подтвердить
                </button>
              )}
            </article>
          ))}
          {criteria.every((x) => x.confirmed) ? (
            <p className="notice success">
              Базовая подготовка пройдена! Теперь вам доступны первые элементы
              калистеники.
            </p>
          ) : (
            <aside className="card">
              <h3>Пока не получается выполнить базу?</h3>
              <p>Развивайте силу и готовьте тело к будущим элементам.</p>
              <Link className="primary-button" to="/workouts?category=strength">
                Перейти к базовым тренировкам
              </Link>
            </aside>
          )}
        </section>
      )}
      {levels.map((level) => (
        <article className={`skill-level ${level.status}`} key={level.id}>
          <i>
            {level.status === "completed"
              ? "✓"
              : level.status === "locked"
                ? "🔒"
                : level.level_number === skill.current_level
                  ? "●"
                  : "○"}
          </i>
          <div>
            <p className="eyebrow">УРОВЕНЬ {level.level_number}</p>
            <h3>{level.name}</h3>
            <p>{level.description}</p>
            <small>
              Критерий: {criterion(level.criterion_type, level.criterion_value)}
            </small>
            {level.workouts.map((w) => (
              <Link
                className="level-workout"
                to={`/workouts/${w.id}`}
                key={w.id}
              >
                {w.title} · {w.estimated_minutes} мин →
              </Link>
            ))}
            {level.status !== "locked" && level.status !== "completed" && (
              <button
                disabled={complete.isPending}
                onClick={() =>
                  complete.mutate({
                    level: level.level_number,
                    value: level.criterion_value,
                  })
                }
              >
                {complete.isPending ? "Проверяем…" : "Подтвердить уровень"}
              </button>
            )}
          </div>
        </article>
      ))}
      {allComplete && skill.status !== "mastered" && (
        <button
          className="primary-button workout-cta"
          disabled={master.isPending}
          onClick={() => master.mutate()}
        >
          {master.isPending
            ? "Проверяем…"
            : `Подтвердить навык · +${skill.xp_reward} XP`}
        </button>
      )}
      {(complete.isError || master.isError) && (
        <p className="notice error">
          {(complete.error ?? master.error) instanceof Error
            ? (complete.error ?? master.error)?.message
            : "Не удалось обновить навык."}
        </p>
      )}
      {master.isSuccess && (
        <p className="success">Навык освоен! +{master.data.xp_earned} XP</p>
      )}
    </div>
  );
}
