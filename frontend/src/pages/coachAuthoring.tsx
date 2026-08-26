import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Link,
  useNavigate,
  useParams,
  useSearchParams,
} from "react-router-dom";
import {
  action,
  coachMe,
  coachOptions,
  content,
  contentDetail,
  createLesson,
  saveBuilder,
  updateLesson,
  type BuilderExercise,
  type BuilderInput,
  type ContentItem,
  type ContentKind,
  type LessonBlock,
  type LessonInput,
  type ProgramWorkout,
} from "../api/coach";
import { useSessionStore } from "../store/session";
import { WorkoutBuilderFields } from "./WorkoutBuilderFields";
import "../authoringActions.css";
export const kinds: {
  key: ContentKind;
  one: string;
  many: string;
  description: string;
}[] = [
  {
    key: "lessons",
    one: "Урок",
    many: "Уроки",
    description: "Теория, техника и обучающие материалы",
  },
  {
    key: "exercises",
    one: "Упражнение",
    many: "Упражнения",
    description: "Отдельное упражнение и техника выполнения",
  },
  {
    key: "workouts",
    one: "Тренировка",
    many: "Тренировки",
    description: "Последовательность упражнений, подходов и отдыха",
  },
  {
    key: "programs",
    one: "Программа",
    many: "Программы",
    description: "План из нескольких тренировок и уровней",
  },
  {
    key: "skills",
    one: "Прогрессия",
    many: "Прогрессии",
    description: "Путь обучения навыку или элементу",
  },
];
const statusLabel = {
    draft: "Черновик",
    published: "Опубликован",
    archived: "В архиве",
  },
  difficultyLabel: Record<string, string> = {
    beginner: "Начальный",
    intermediate: "Средний",
    advanced: "Продвинутый",
  };
function token() {
  return useSessionStore.getState().accessToken!;
}
function ErrorBox({ error }: { error?: Error | null }) {
  return (
    <p className="notice error">
      {error?.message || "Не удалось сохранить. Проверьте заполненные поля."}
    </p>
  );
}
function RetryBox({ text, retry }: { text: string; retry: () => void }) {
  return (
    <div className="notice error">
      <p>{text}</p>
      <button type="button" className="secondary-button" onClick={retry}>
        Повторить
      </button>
    </div>
  );
}
const optionName = (option: { name: string; status?: string }) =>
  `${option.name}${option.status ? ` · ${statusLabel[option.status as keyof typeof statusLabel] ?? option.status}` : ""}`;
export function CoachContentHome() {
  const [create, setCreate] = useState(false),
    [filter, setFilter] = useState<"all" | ContentKind>("all");
  const q = useQuery({
    queryKey: ["coach-content-home"],
    queryFn: async () => ({
      items: (
        await Promise.all(
          kinds.map(async (k) =>
            (await content(token(), k.key)).items.map((x) => ({
              ...x,
              kind: k.key,
            })),
          ),
        )
      ).flat(),
    }),
  });
  const role = useQuery({
    queryKey: ["coach-role"],
    queryFn: () => coachMe(token()),
  });
  const manageAll =
    role.data?.role === "admin" || role.data?.role === "super_admin";
  const items =
    q.data?.items.filter((x) => filter === "all" || x.kind === filter) ?? [];
  return (
    <div className="stack">
      <div className="coach-title">
        <h2>Контент</h2>
        <button className="primary-button" onClick={() => setCreate(true)}>
          + Создать
        </button>
      </div>
      <section className="stack">
        <p className="eyebrow">СОЗДАТЬ</p>
        <div className="authoring-grid">
          {kinds.map((k) => (
            <Link
              className="card create-card"
              to={`/coach/${k.key}/new`}
              key={k.key}
            >
              <h3>{k.one}</h3>
              <p>{k.description}</p>
              <b>Создать →</b>
            </Link>
          ))}
        </div>
      </section>
      <section className="stack">
        <h3>Мой контент</h3>
        <div className="filter-chips">
          <button
            className={filter === "all" ? "active" : ""}
            onClick={() => setFilter("all")}
          >
            Все
          </button>
          {kinds.map((k) => (
            <button
              className={filter === k.key ? "active" : ""}
              onClick={() => setFilter(k.key)}
              key={k.key}
            >
              {k.many}
            </button>
          ))}
        </div>
        {q.isLoading && (
          <div className="notice skeleton">Загружаем материалы…</div>
        )}
        {q.isError && (
          <RetryBox
            text="Не удалось загрузить материалы Coach Studio."
            retry={() => q.refetch()}
          />
        )}
        {items.map((x) => (
          <ContentCard
            key={`${x.kind}-${x.id}`}
            item={x}
            kind={x.kind}
            manageAll={manageAll}
          />
        ))}
        {!q.isLoading && !q.isError && !items.length && (
          <div className="empty-state compact">
            <h3>Здесь пока нет материалов</h3>
            <p>Создайте первый материал для своих учеников.</p>
            <button className="primary-button" onClick={() => setCreate(true)}>
              + Создать
            </button>
          </div>
        )}
      </section>
      {create && <CreateSheet close={() => setCreate(false)} />}
    </div>
  );
}
function CreateSheet({ close }: { close: () => void }) {
  return (
    <div className="sheet-backdrop" onClick={close}>
      <section className="bottom-sheet" onClick={(e) => e.stopPropagation()}>
        <div className="sheet-handle" />
        <h2>Что создать?</h2>
        {kinds.map((k) => (
          <Link to={`/coach/${k.key}/new`} onClick={close} key={k.key}>
            {k.one}
            <span>›</span>
          </Link>
        ))}
        <button onClick={close}>Отмена</button>
      </section>
    </div>
  );
}
export function CoachContentList() {
  const { kind = "lessons" } = useParams();
  const [search, setSearch] = useState(""),
    [status, setStatus] = useState("");
  const role = useQuery({
    queryKey: ["coach-role"],
    queryFn: () => coachMe(token()),
  });
  const manageAll =
    role.data?.role === "admin" || role.data?.role === "super_admin";
  const k = kinds.find((x) => x.key === kind) ?? kinds[0],
    q = useQuery({
      queryKey: ["coach-content", kind, search, status],
      queryFn: () => content(token(), kind, search, status),
    });
  return (
    <div className="stack">
      <Link className="text-link" to="/coach/content">
        ← Контент
      </Link>
      <div className="coach-title">
        <h2>{k.many}</h2>
        <Link className="primary-button" to={`/coach/${kind}/new`}>
          + Создать
        </Link>
      </div>
      <div className="coach-filters">
        <input
          placeholder={`Поиск: ${k.many.toLowerCase()}`}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <select value={status} onChange={(e) => setStatus(e.target.value)}>
          <option value="">Все статусы</option>
          <option value="draft">Черновики</option>
          <option value="published">Опубликованные</option>
          <option value="archived">Архив</option>
        </select>
      </div>
      {q.isLoading && (
        <div className="notice skeleton">Загружаем материалы…</div>
      )}
      {q.isError && (
        <RetryBox
          text={`Не удалось загрузить: ${k.many.toLowerCase()}.`}
          retry={() => q.refetch()}
        />
      )}
      {q.data?.items.map((x) => (
        <ContentCard item={x} kind={k.key} key={x.id} manageAll={manageAll} />
      ))}
      {!q.isLoading && !q.isError && !q.data?.items.length && (
        <div className="empty-state compact">
          <h3>У вас пока нет своих материалов</h3>
          <p>Создайте первый материал и настройте его под своих учеников.</p>
          <Link className="primary-button" to={`/coach/${kind}/new`}>
            + Создать
          </Link>
        </div>
      )}
    </div>
  );
}
function ContentCard({
  item,
  kind,
  manageAll = false,
}: {
  item: ContentItem;
  kind: ContentKind;
  manageAll?: boolean;
}) {
  const nav = useNavigate(),
    qc = useQueryClient(),
    system = !item.owner_user_id,
    readOnly = system && !manageAll,
    act = useMutation({
      mutationFn: (
        name: "publish" | "unpublish" | "archive" | "restore" | "duplicate",
      ) => action(token(), kind, item.id, name),
      onSuccess: (x) => {
        qc.invalidateQueries({ queryKey: ["coach-content"] });
        if (x.id !== item.id) nav(`/coach/${kind}/${x.id}/edit`);
      },
    });
  const edit = () => {
    nav(`/coach/${kind}/${item.id}/edit${readOnly ? "?view=system" : ""}`);
  };
  return (
    <article className={`card content-card ${readOnly ? "read-only" : ""}`}>
      <button className="content-card-main" onClick={edit}>
        <span>
          <small>
            {kinds.find((k) => k.key === kind)?.one}
            {system ? " · Системный материал" : ""}
          </small>
          <h3>{item.name}</h3>
          <p>
            {difficultyLabel[item.difficulty] ?? item.difficulty} ·{" "}
            {statusLabel[item.status]}
          </p>
        </span>
        <b>{readOnly ? "Просмотреть ›" : "Редактировать ›"}</b>
      </button>
      <div className="coach-actions">
        {readOnly ? (
          <button onClick={() => act.mutate("duplicate")}>
            Создать свою копию
          </button>
        ) : (
          <>
            {item.status === "draft" && (
              <button onClick={() => act.mutate("publish")}>
                Опубликовать
              </button>
            )}
            {item.status === "published" && (
              <button onClick={() => act.mutate("unpublish")}>
                Снять с публикации
              </button>
            )}
            {item.status !== "archived" && (
              <button
                onClick={() =>
                  confirm("Переместить материал в архив?") &&
                  act.mutate("archive")
                }
              >
                В архив
              </button>
            )}
            {item.status === "archived" && (
              <button onClick={() => act.mutate("restore")}>
                Восстановить
              </button>
            )}
            <button onClick={() => act.mutate("duplicate")}>
              Создать копию
            </button>
          </>
        )}
      </div>
    </article>
  );
}
export const shouldLeave = (dirty: boolean, ask: () => boolean) =>
  !dirty || ask();
function useDirty(dirty: boolean) {
  useEffect(() => {
    const before = (e: BeforeUnloadEvent) => {
      if (dirty) {
        e.preventDefault();
        e.returnValue = "";
      }
    };
    window.addEventListener("beforeunload", before);
    const tg = window.Telegram?.WebApp;
    const back = () => {
      if (
        shouldLeave(dirty, () =>
          confirm("Есть несохранённые изменения. Выйти без сохранения?"),
        )
      )
        history.back();
    };
    if (dirty) {
      tg?.BackButton?.show();
      tg?.BackButton?.onClick(back);
    }
    return () => {
      window.removeEventListener("beforeunload", before);
      tg?.BackButton?.offClick(back);
    };
  }, [dirty]);
}
const diff = (
  <>
    <option value="beginner">Начальный</option>
    <option value="intermediate">Средний</option>
    <option value="advanced">Продвинутый</option>
  </>
);
export const workoutTarget = (mode: "reps" | "time") =>
  mode === "time"
    ? { target_reps: undefined, target_duration_seconds: 30 }
    : { target_reps: 10, target_duration_seconds: undefined };
export const numericValue = (raw: string) =>
  raw === "" ? undefined : Number(raw);
export const normalizeProgramWorkouts = (items: ProgramWorkout[]) =>
  items.map((item, sort_order) => ({ ...item, sort_order }));
export const moveProgramWorkout = (
  items: ProgramWorkout[],
  index: number,
  direction: -1 | 1,
) => {
  const target = index + direction;
  if (target < 0 || target >= items.length)
    return normalizeProgramWorkouts(items);
  const next = [...items];
  [next[index], next[target]] = [next[target], next[index]];
  return normalizeProgramWorkouts(next);
};
export const builderPayload = (
  kind: Exclude<ContentKind, "lessons">,
  value: BuilderInput,
): BuilderInput => {
  const common = {
    description: value.description,
    difficulty: value.difficulty,
    cover_media_id: value.cover_media_id,
  };
  if (kind === "workouts")
    return {
      ...common,
      title: value.title,
      category: value.category,
      estimated_minutes: value.estimated_minutes,
      warmup_enabled: value.category === "warmup" ? false : value.warmup_enabled,
      warmup_workout_id:
        value.category === "warmup" || value.warmup_enabled === false
          ? undefined
          : value.warmup_workout_id,
      exercises: value.exercises,
    };
  if (kind === "exercises")
    return {
      ...common,
      name: value.name,
      movement_type: value.movement_type,
      instructions: value.instructions,
      common_mistakes: value.common_mistakes,
      coach_tips: value.coach_tips,
      muscle_groups: value.muscle_groups,
      equipment: value.equipment,
      tags: value.tags,
      demo_media_id: value.demo_media_id,
    };
  if (kind === "programs")
    return {
      ...common,
      name: value.name,
      duration_weeks: value.duration_weeks,
      category: value.category,
      workouts: value.workouts,
    };
  return {
    ...common,
    name: value.name,
    code: value.code,
    category: value.category,
    icon: value.icon,
    xp_reward: value.xp_reward,
    final_criterion_type: value.final_criterion_type,
    final_criterion_value: value.final_criterion_value,
    map_group: value.map_group,
    levels: value.levels,
    requirements: value.requirements,
    sort_order: value.sort_order,
  };
};
export function CoachEditor() {
  const { kind = "lessons", id = "new" } = useParams();
  const [search] = useSearchParams();
  const readOnly = search.get("view") === "system";
  const k = kind as ContentKind;
  if (!kinds.some((x) => x.key === k))
    return <div className="notice error">Неизвестный тип материала.</div>;
  return k === "lessons" ? (
    <LessonEditor id={id} readOnly={readOnly} />
  ) : (
    <BuilderEditor kind={k} id={id} readOnly={readOnly} />
  );
}
const blankLesson: LessonInput = {
  category_id: "",
  title: "",
  short_description: "",
  content: "",
  difficulty: "beginner",
  duration_minutes: 10,
  blocks: [],
};
export const blockNames: Record<LessonBlock["type"], string> = {
  heading: "Заголовок",
  text: "Текст",
  tip: "Совет тренера",
  warning: "Важное замечание",
  image: "Изображение",
  video: "Видео",
  checklist: "Список",
  divider: "Разделитель",
};
function LessonEditor({
  id,
  readOnly = false,
}: {
  id: string;
  readOnly?: boolean;
}) {
  const edit = id !== "new",
    nav = useNavigate(),
    [value, setValue] = useState(blankLesson),
    [dirty, setDirty] = useState(false),
    [picker, setPicker] = useState(false),
    [success, setSuccess] = useState(false),
    [validation, setValidation] = useState("");
  useDirty(dirty);
  const opts = useQuery({
      queryKey: ["coach-options"],
      queryFn: () => coachOptions(token()),
    }),
    detail = useQuery({
      queryKey: ["coach-detail", "lessons", id],
      queryFn: () => contentDetail(token(), "lessons", id),
      enabled: edit,
    }),
    mediaOptions =
      (opts.data as unknown as { media?: { id: string; name: string }[] })
        ?.media ?? [];
  useEffect(() => {
    if (detail.data) {
      const x = detail.data.item;
      setValue({
        category_id: String(x.category_id ?? ""),
        title: String(x.title ?? ""),
        short_description: String(x.short_description ?? ""),
        content: String(x.content ?? ""),
        difficulty: String(x.difficulty ?? "beginner"),
        duration_minutes: Number(x.duration_minutes ?? 10),
        cover_media_id: x.cover_media_id ? String(x.cover_media_id) : undefined,
        blocks: (x.content_blocks as LessonBlock[]) ?? [],
      });
    }
  }, [detail.data]);
  const change = (v: LessonInput) => {
      setValue(v);
      setDirty(true);
      setValidation("");
    },
    save = useMutation({
      mutationFn: async (publish: boolean) => {
        const result = edit
          ? await updateLesson(token(), id, value)
          : await createLesson(token(), value);
        if (publish) await action(token(), "lessons", result.id, "publish");
        return result;
      },
      onSuccess: (x) => {
        setDirty(false);
        setSuccess(true);
        if (!edit) nav(`/coach/lessons/${x.id}/edit`, { replace: true });
      },
    });
  const add = (type: LessonBlock["type"]) => {
      change({
        ...value,
        blocks: [
          ...value.blocks,
          {
            type,
            ...(type === "image" || type === "video"
              ? { media_id: "" }
              : type === "checklist"
                ? { items: [""] }
                : { text: "" }),
          },
        ],
      });
      setPicker(false);
    },
    move = (i: number, d: number) => {
      const b = [...value.blocks],
        [x] = b.splice(i, 1);
      b.splice(i + d, 0, x);
      change({ ...value, blocks: b });
    };
  const submit = (publish = false) => {
    if (!value.title.trim()) return setValidation("Введите название урока.");
    if (!value.category_id) return setValidation("Выберите категорию урока.");
    if (!value.short_description.trim())
      return setValidation("Добавьте краткое описание урока.");
    if (!Number.isFinite(value.duration_minutes) || value.duration_minutes < 1)
      return setValidation("Укажите длительность урока.");
    if (
      value.blocks.some(
        (x) => (x.type === "image" || x.type === "video") && !x.media_id,
      )
    )
      return setValidation("Выберите файл для фото или видео блока.");
    if (
      value.blocks.some(
        (x) => x.type === "checklist" && !x.items?.some((item) => item.trim()),
      )
    )
      return setValidation("Добавьте хотя бы один пункт списка.");
    if (
      value.blocks.some(
        (x) =>
          !["image", "video", "checklist", "divider"].includes(x.type) &&
          !x.text?.trim(),
      )
    )
      return setValidation("Заполните добавленные текстовые блоки.");
    save.mutate(publish);
  };
  return (
    <form
      className={`stack coach-editor-page ${readOnly ? "system-view" : ""}`}
      onSubmit={(e) => {
        e.preventDefault();
        submit(false);
      }}
    >
      <Back dirty={dirty} />
      <h2>{edit ? "Редактировать урок" : "Новый урок"}</h2>
      {readOnly && (
        <p className="notice">
          Системный материал доступен только для просмотра. Создайте свою копию
          из списка, чтобы редактировать его.
        </p>
      )}
      {opts.isLoading && (
        <p className="notice skeleton">Загружаем справочники…</p>
      )}
      {opts.isError && (
        <RetryBox
          text="Не удалось загрузить категории и медиатеку."
          retry={() => opts.refetch()}
        />
      )}
      {detail.isError && (
        <RetryBox
          text="Не удалось загрузить урок."
          retry={() => detail.refetch()}
        />
      )}
      <label>
        Название *
        <input
          value={value.title}
          onChange={(e) => change({ ...value, title: e.target.value })}
        />
      </label>
      <label>
        Категория
        <select
          value={value.category_id}
          onChange={(e) => change({ ...value, category_id: e.target.value })}
        >
          <option value="">Выберите категорию</option>
          {opts.data?.categories.map((x) => (
            <option value={x.id} key={x.id}>
              {x.name}
            </option>
          ))}
        </select>
      </label>
      <label>
        Обложка — фото или видео
        <select
          value={value.cover_media_id ?? ""}
          onChange={(e) =>
            change({ ...value, cover_media_id: e.target.value || undefined })
          }
        >
          <option value="">Без обложки</option>
          {mediaOptions.map((x) => (
            <option value={x.id} key={x.id}>
              {x.name}
            </option>
          ))}
        </select>
      </label>
      <label>
        Краткое описание
        <textarea
          value={value.short_description}
          onChange={(e) =>
            change({ ...value, short_description: e.target.value })
          }
        />
      </label>
      <div className="form-grid">
        <label>
          Сложность
          <select
            value={value.difficulty}
            onChange={(e) => change({ ...value, difficulty: e.target.value })}
          >
            {diff}
          </select>
        </label>
        <label>
          Длительность, минут
          <input
            type="number"
            min="1"
            value={value.duration_minutes ?? ""}
            onChange={(e) =>
              change({
                ...value,
                duration_minutes: numericValue(e.target.value) as number,
              })
            }
          />
        </label>
      </div>
      <section className="stack">
        <h3>Материал урока</h3>
        {value.blocks.map((b, i) => (
          <div className="card lesson-block-editor" key={i}>
            <b>{blockNames[b.type]}</b>
            {b.type === "image" || b.type === "video" ? (
              <select
                value={b.media_id ?? ""}
                onChange={(e) =>
                  change({
                    ...value,
                    blocks: value.blocks.map((x, j) =>
                      j === i ? { ...x, media_id: e.target.value } : x,
                    ),
                  })
                }
              >
                <option value="">Выберите файл</option>
                {mediaOptions.map((x) => (
                  <option value={x.id} key={x.id}>
                    {x.name}
                  </option>
                ))}
              </select>
            ) : b.type === "divider" ? null : (
              <textarea
                placeholder={
                  b.type === "checklist"
                    ? "Каждый пункт с новой строки"
                    : `Введите: ${blockNames[b.type].toLowerCase()}`
                }
                value={
                  b.type === "checklist"
                    ? (b.items?.join("\n") ?? "")
                    : (b.text ?? "")
                }
                onChange={(e) =>
                  change({
                    ...value,
                    blocks: value.blocks.map((x, j) =>
                      j === i
                        ? b.type === "checklist"
                          ? { ...x, items: e.target.value.split("\n") }
                          : { ...x, text: e.target.value }
                        : x,
                    ),
                  })
                }
              />
            )}
            <div>
              <button type="button" disabled={!i} onClick={() => move(i, -1)}>
                ↑
              </button>
              <button
                type="button"
                disabled={i === value.blocks.length - 1}
                onClick={() => move(i, 1)}
              >
                ↓
              </button>
              <button
                type="button"
                onClick={() =>
                  change({
                    ...value,
                    blocks: value.blocks.filter((_, j) => j !== i),
                  })
                }
              >
                Удалить
              </button>
            </div>
          </div>
        ))}
        <button
          type="button"
          className="secondary-button"
          onClick={() => setPicker(true)}
        >
          + Добавить блок
        </button>
        <Link className="text-link" to="/coach/media">
          Добавить фото или видео в медиатеку →
        </Link>
      </section>
      {validation && <p className="notice error">{validation}</p>}
      {success && <p className="notice success">Урок сохранён.</p>}
      {save.isError && <ErrorBox error={save.error} />}
      {!readOnly && (
        <div className="publish-actions">
          <button className="secondary-button" disabled={save.isPending}>
            {save.isPending ? "Сохраняем…" : "Сохранить черновик"}
          </button>
          <button
            type="button"
            className="primary-button"
            disabled={save.isPending}
            onClick={() => submit(true)}
          >
            {save.isPending ? "Публикуем…" : "Сохранить и опубликовать"}
          </button>
        </div>
      )}
      {picker && <BlockSheet close={() => setPicker(false)} add={add} />}
    </form>
  );
}
function BlockSheet({
  close,
  add,
}: {
  close: () => void;
  add: (t: LessonBlock["type"]) => void;
}) {
  const choices: LessonBlock["type"][] = [
    "heading",
    "text",
    "tip",
    "warning",
    "image",
    "video",
    "checklist",
    "divider",
  ];
  return (
    <div className="sheet-backdrop" onClick={close}>
      <section className="bottom-sheet" onClick={(e) => e.stopPropagation()}>
        <h2>Добавить в урок</h2>
        {choices.map((x) => (
          <button onClick={() => add(x)} key={x}>
            {blockNames[x]}
            <span>+</span>
          </button>
        ))}
      </section>
    </div>
  );
}
const blankBuilder = (kind: ContentKind): BuilderInput => ({
  description: "",
  difficulty: "beginner",
  ...(kind === "exercises"
    ? {
        name: "",
        instructions: "",
        common_mistakes: "",
        movement_type: "reps",
        muscle_groups: ["Все тело"],
        equipment: [],
        tags: [],
      }
    : {}),
  ...(kind === "workouts"
    ? {
        title: "",
        category: "strength",
        estimated_minutes: 20,
        warmup_enabled: true,
        exercises: [],
      }
    : {}),
  ...(kind === "programs"
    ? { name: "", duration_weeks: 4, category: "OTHER", workouts: [] }
    : {}),
  ...(kind === "skills"
    ? {
        name: "",
        category: "SKILL",
        map_group: "basic",
        icon: "◇",
        xp_reward: 100,
        final_criterion_type: "repetitions",
        final_criterion_value: 1,
        levels: [
          {
            level_number: 1,
            title: "Первый этап",
            description: "Начальный этап",
            difficulty: "beginner",
            unlock_rule_type: "none",
            unlock_rule_value: 0,
            criterion_type: "repetitions",
            criterion_value: 1,
            sort_order: 0,
          },
        ],
        requirements: [],
      }
    : {}),
});
export function validateBuilder(
  kind: Exclude<ContentKind, "lessons">,
  value: BuilderInput,
  publish = false,
) {
  if ((kind !== "workouts" || publish) && !(value.name ?? value.title)?.trim()) return "Введите название.";
  if ((kind !== "workouts" || publish) && !value.description.trim()) return "Добавьте описание.";
  if (kind === "exercises") {
    if (!value.instructions?.trim()) return "Добавьте инструкцию выполнения.";
    if (!value.muscle_groups?.some((x) => x.trim()))
      return "Укажите хотя бы одну группу мышц.";
  }
  if (kind === "workouts") {
    if (!value.estimated_minutes || value.estimated_minutes < 1)
      return "Укажите продолжительность тренировки.";
    const exercises = value.exercises ?? [];
    if (publish && !exercises.length)
      return "Добавьте хотя бы одно упражнение.";
    const ids = exercises.map((x) => x.exercise_id);
    if (new Set(ids).size !== ids.length)
      return "Одно упражнение нельзя добавить в тренировку дважды.";
    if (
      exercises.some(
        (x) =>
          !Number.isFinite(x.sets) ||
          x.sets < 1 ||
          !Number.isFinite(x.rest_seconds) ||
          x.rest_seconds < 0 ||
          (x.target_reps == null && x.target_duration_seconds == null) ||
          (x.target_reps != null &&
            (!Number.isFinite(x.target_reps) || x.target_reps < 1)) ||
          (x.target_duration_seconds != null &&
            (!Number.isFinite(x.target_duration_seconds) ||
              x.target_duration_seconds < 1)),
      )
    )
      return "Проверьте подходы, повторения или время и отдых у каждого упражнения.";
  }
  if (
    kind === "programs" &&
    (!value.duration_weeks || value.duration_weeks < 1)
  )
    return "Укажите продолжительность программы.";
  if (kind === "programs" && publish && !value.workouts?.length)
    return "Добавьте хотя бы одну тренировку.";
  if (kind === "skills") {
    if (!value.icon?.trim()) return "Добавьте значок прогрессии.";
    if (!value.final_criterion_value || value.final_criterion_value < 1)
      return "Укажите итоговый критерий прогрессии.";
    if (!value.levels?.length) return "Добавьте хотя бы один этап прогрессии.";
    if (
      value.levels.some(
        (level) =>
          !Number.isFinite(level.criterion_value) || level.criterion_value < 1,
      )
    )
      return "Укажите значение критерия для каждого этапа.";
  }
  return "";
}
export function workoutValidationErrors(value: BuilderInput, publish: boolean, options: {id:string;name:string}[] = []) {
  const errors:string[]=[];
  if (publish && !value.title?.trim()) errors.push("Укажите название.");
  if (publish && !value.description.trim()) errors.push("Добавьте описание.");
  if (!value.estimated_minutes || value.estimated_minutes < 1) errors.push("Укажите ориентировочную длительность.");
  const items=value.exercises??[], names=new Map(options.map(x=>[x.id,x.name]));
  if (publish && !items.length) errors.push("Добавьте хотя бы одно упражнение.");
  const used=new Set<string>();
  items.forEach((item,index)=>{
    const name=names.get(item.exercise_id)??`Упражнение ${index+1}`;
    if(used.has(item.exercise_id))errors.push(`${name}: упражнение уже добавлено.`);used.add(item.exercise_id);
    if(!Number.isFinite(item.sets)||item.sets<1)errors.push(`${name}: количество подходов должно быть больше 0.`);
    if((item.target_reps==null)===(item.target_duration_seconds==null))errors.push(`${name}: выберите повторения или время выполнения.`);
    if(item.target_reps!=null&&(!Number.isFinite(item.target_reps)||item.target_reps<1))errors.push(`${name}: количество повторений должно быть больше 0.`);
    if(item.target_duration_seconds!=null&&(!Number.isFinite(item.target_duration_seconds)||item.target_duration_seconds<1))errors.push(`${name}: время выполнения должно быть больше 0.`);
    if(!Number.isFinite(item.rest_seconds)||item.rest_seconds<0)errors.push(`${name}: отдых не может быть отрицательным.`);
  });
  return errors;
}
function BuilderEditor({
  kind,
  id,
  readOnly = false,
}: {
  kind: Exclude<ContentKind, "lessons">;
  id: string;
  readOnly?: boolean;
}) {
  const edit = id !== "new",
    nav = useNavigate(),
    [value, setValue] = useState(() => blankBuilder(kind)),
    [dirty, setDirty] = useState(false),
    [success, setSuccess] = useState(false),
    [validation, setValidation] = useState(""),
    [workoutErrors, setWorkoutErrors] = useState<string[]>([]);
  useDirty(dirty);
  const opts = useQuery({
      queryKey: ["coach-options"],
      queryFn: () => coachOptions(token()),
    }),
    detail = useQuery({
      queryKey: ["coach-detail", kind, id],
      queryFn: () => contentDetail(token(), kind, id),
      enabled: edit,
    }),
    mediaOptions =
      (opts.data as unknown as { media?: { id: string; name: string }[] })
        ?.media ?? [];
  useEffect(() => {
    if (detail.data) {
      const item = detail.data.item;
      const levels = Array.isArray(item.levels)
        ? item.levels.map((level) => {
            const x = level as Record<string, unknown>;
            return {
              level_number: Number(x.level_number),
              title: String(x.title ?? x.name ?? ""),
              description: String(x.description ?? ""),
              difficulty: String(x.difficulty ?? item.difficulty ?? "beginner"),
              unlock_rule_type: String(x.unlock_rule_type ?? "none"),
              unlock_rule_value: Number(x.unlock_rule_value ?? 0),
              criterion_type: String(x.criterion_type ?? "repetitions"),
              criterion_value: Number(x.criterion_value ?? 1),
              ...(x.program_level_id
                ? { program_level_id: String(x.program_level_id) }
                : {}),
              sort_order: Number(x.sort_order ?? 0),
            };
          })
        : undefined;
      const exercises = Array.isArray(item.exercises)
        ? item.exercises.map((exercise) => {
            const x = exercise as Record<string, unknown>;
            return {
              exercise_id: String(x.exercise_id ?? ""),
              sets: Number(x.sets ?? 1),
              ...(x.target_reps == null
                ? {}
                : { target_reps: Number(x.target_reps) }),
              ...(x.target_duration_seconds == null
                ? {}
                : {
                    target_duration_seconds: Number(x.target_duration_seconds),
                  }),
              rest_seconds: Number(x.rest_seconds ?? 60),
              ...(x.notes == null ? {} : { notes: String(x.notes) }),
              sort_order: Number(x.sort_order ?? 0),
            };
          })
        : undefined;
      const workouts = Array.isArray(item.workouts)
        ? item.workouts.map((workout) => {
            const x = workout as Record<string, unknown>;
            return {
              workout_id: String(x.workout_id ?? ""),
              sort_order: Number(x.sort_order ?? 0),
            };
          })
        : undefined;
      const clean = { ...item };
      delete clean.levels;
      delete clean.exercises;
      delete clean.workouts;
      setValue({
        ...blankBuilder(kind),
        ...clean,
        ...(levels ? { levels } : {}),
        ...(exercises ? { exercises } : {}),
        ...(workouts ? { workouts } : {}),
      } as BuilderInput);
    }
  }, [detail.data, kind]);
  const change = (x: BuilderInput) => {
      setValue(x);
      setDirty(true);
      setValidation("");
      setWorkoutErrors([]);
    },
    save = useMutation({
      mutationFn: async (publish: boolean) => {
        const result = await saveBuilder(
          token(),
          kind,
          builderPayload(kind, value),
          edit ? id : undefined,
        );
        if (publish) await action(token(), kind, result.id, "publish");
        return result;
      },
      onSuccess: (x) => {
        setDirty(false);
        setSuccess(true);
        if (!edit) nav(`/coach/${kind}/${x.id}/edit`, { replace: true });
      },
    }),
    title = kinds.find((x) => x.key === kind)!;
  const submit = (publish = false) => {
    if(kind==="workouts"){
      const errors=workoutValidationErrors(value,publish,opts.data?.exercises);
      if(errors.length){setWorkoutErrors(errors);setTimeout(()=>document.querySelector(".validation-summary")?.scrollIntoView({behavior:"smooth",block:"center"}),0);return;}
    }
    const error = validateBuilder(kind, value, publish);
    if (error) return setValidation(error);
    save.mutate(publish);
  };
  const cover = String(
    (value as BuilderInput & { cover_media_id?: string }).cover_media_id ?? "",
  );
  const demo = String(value.demo_media_id ?? "");
  return (
    <form
      className={`stack coach-editor-page ${readOnly ? "system-view" : ""}`}
      onSubmit={(e) => {
        e.preventDefault();
        submit(false);
      }}
    >
      <Back dirty={dirty} />
      <h2>{edit ? `Редактировать: ${title.one}` : `Новая: ${title.one}`}</h2>
      {readOnly && (
        <p className="notice">
          Системный материал доступен только для просмотра. Создайте свою копию
          из списка, чтобы редактировать его.
        </p>
      )}
      {opts.isLoading && (
        <p className="notice skeleton">Загружаем справочники…</p>
      )}
      {opts.isError && (
        <RetryBox
          text="Не удалось загрузить справочники для формы."
          retry={() => opts.refetch()}
        />
      )}
      {detail.isError && (
        <RetryBox
          text="Не удалось загрузить материал."
          retry={() => detail.refetch()}
        />
      )}
      <label>
        Название *
        <input
          value={value.name ?? value.title ?? ""}
          onChange={(e) =>
            change(
              kind === "workouts"
                ? { ...value, title: e.target.value }
                : { ...value, name: e.target.value },
            )
          }
        />
      </label>
      <label>
        Описание
        <textarea
          value={value.description}
          onChange={(e) => change({ ...value, description: e.target.value })}
        />
      </label>
      <label>
        Сложность
        <select
          value={value.difficulty}
          onChange={(e) => change({ ...value, difficulty: e.target.value })}
        >
          {diff}
        </select>
      </label>
      {(kind === "programs" || kind === "skills") && (
        <label>
          Категория
          <select
            value={value.category ?? "OTHER"}
            onChange={(e) => change({ ...value, category: e.target.value })}
          >
            <option value="MORNING_ROUTINE">Утренняя рутина</option>
            <option value="WARMUP">Разминка</option>
            <option value="BASE_STRENGTH">Базовая сила</option>
            <option value="SKILL">Навык</option>
            <option value="MOBILITY">Мобильность</option>
            <option value="OTHER">Другое</option>
          </select>
        </label>
      )}
      <label>
        Фото или видео
        <select
          value={cover}
          onChange={(e) =>
            change(
              Object.assign({}, value, {
                cover_media_id: e.target.value || undefined,
              }),
            )
          }
        >
          <option value="">Без медиа</option>
          {mediaOptions.map((x) => (
            <option value={x.id} key={x.id}>
              {x.name}
            </option>
          ))}
        </select>
      </label>
      <Link className="text-link" to="/coach/media">
        Добавить фото или видео в медиатеку →
      </Link>
      {kind === "exercises" && (
        <>
          <label>
            Демонстрация упражнения
            <select
              disabled={readOnly}
              value={demo}
              onChange={(e) =>
                change({
                  ...value,
                  demo_media_id: e.target.value || undefined,
                })
              }
            >
              <option value="">Демонстрация пока не добавлена</option>
              {mediaOptions.map((x) => (
                <option value={x.id} key={x.id}>
                  {x.name}
                </option>
              ))}
            </select>
            <small>
              MP4/WebM/GIF или статичное изображение, до 5 МБ; видео 1–6
              секунд. Загрузка файлов требует настроенного object storage, сейчас
              можно выбрать зарегистрированный asset.
            </small>
          </label>
          <label>
            Как измерять выполнение
            <select
              value={value.movement_type}
              onChange={(e) =>
                change({ ...value, movement_type: e.target.value })
              }
            >
              <option value="reps">Количество повторений</option>
              <option value="duration">Время исполнения</option>
              <option value="distance">Дистанция</option>
              <option value="custom">Другой критерий</option>
            </select>
          </label>
          <label>
            Инструкция
            <textarea
              value={value.instructions}
              onChange={(e) =>
                change({ ...value, instructions: e.target.value })
              }
            />
          </label>
          <label>
            Советы тренера
            <textarea
              value={value.coach_tips}
              onChange={(e) => change({ ...value, coach_tips: e.target.value })}
            />
          </label>
          <label>
            Частые ошибки
            <textarea
              value={value.common_mistakes ?? ""}
              onChange={(e) =>
                change({ ...value, common_mistakes: e.target.value })
              }
            />
          </label>
          <label>
            Группы мышц
            <input
              value={(value.muscle_groups ?? []).join(", ")}
              onChange={(e) =>
                change({
                  ...value,
                  muscle_groups: e.target.value
                    .split(",")
                    .map((x) => x.trim())
                    .filter(Boolean),
                })
              }
              placeholder="грудь, трицепс"
            />
          </label>
          <label>
            Оборудование
            <input
              value={(value.equipment ?? []).join(", ")}
              onChange={(e) =>
                change({
                  ...value,
                  equipment: e.target.value
                    .split(",")
                    .map((x) => x.trim())
                    .filter(Boolean),
                })
              }
              placeholder="турник, резинка"
            />
          </label>
          <label>
            Теги навыка / семейства
            <input
              value={(value.tags ?? []).join(", ")}
              onChange={(e) => change({...value,tags:e.target.value.split(",").map((x)=>x.trim().toLowerCase()).filter(Boolean)})}
              placeholder="push, handstand, warm-up"
            />
          </label>
        </>
      )}
      {kind === "workouts" && (
        <>
          <label>
            Категория
            <select
              value={value.category ?? "strength"}
              onChange={(e) =>
                change({
                  ...value,
                  category: e.target.value,
                  warmup_enabled:
                    e.target.value === "warmup"
                      ? false
                      : (value.warmup_enabled ?? true),
                  warmup_workout_id:
                    e.target.value === "warmup" ? undefined : value.warmup_workout_id,
                })
              }
            >
              <option value="warmup">Разминка</option>
              <option value="morning">Зарядка</option>
              <option value="strength">Развитие силы</option>
              <option value="skill">Тренировка навыков</option>
            </select>
          </label>
          <WorkoutFields value={value} change={change} opts={opts.data} />
        </>
      )}{" "}
      {kind === "programs" && (
        <ProgramWorkouts value={value} change={change} opts={opts.data} />
      )}
      {kind === "skills" && (
        <>
          <div className="form-grid">
            <label>
              Значок
              <input
                value={value.icon ?? ""}
                onChange={(e) => change({ ...value, icon: e.target.value })}
              />
            </label>
            <label>
              Награда XP
              <input
                type="number"
                min="0"
                value={value.xp_reward ?? ""}
                onChange={(e) =>
                  change({ ...value, xp_reward: numericValue(e.target.value) })
                }
              />
            </label>
          </div>
          <label>
            Тип итогового критерия
            <select
              value={value.final_criterion_type ?? "repetitions"}
              onChange={(e) =>
                change({ ...value, final_criterion_type: e.target.value })
              }
            >
              <option value="repetitions">Повторения</option>
              <option value="duration_seconds">Время, секунд</option>
              <option value="manual_confirmation">Ручное подтверждение</option>
              <option value="workout_count">Количество тренировок</option>
            </select>
          </label>
          <label>
            Значение итогового критерия
            <input
              type="number"
              min="1"
              value={value.final_criterion_value ?? ""}
              onChange={(e) =>
                change({
                  ...value,
                  final_criterion_value: numericValue(e.target.value),
                })
              }
            />
          </label>
          <label>
            Раздел карты
            <select
              value={value.map_group ?? "basic"}
              onChange={(e) => change({
                ...value,
                map_group: e.target.value as BuilderInput["map_group"],
              })}
            >
              <option value="basic">Базовые навыки</option>
              <option value="floor">Пол</option>
              <option value="bar">Турник</option>
              <option value="parallel_bars">Брусья</option>
            </select>
            <small>Определяет раздел навыка в общей карте ученика.</small>
          </label>
          <label>
            Расположение в карте навыков
            <select
              value={
                opts.data?.skills.find(
                  (skill) => skill.sort_order === (value.sort_order ?? 0) - 1,
                )?.id ?? ""
              }
              onChange={(e) => {
                const previous = opts.data?.skills.find(
                  (skill) => skill.id === e.target.value,
                );
                change({
                  ...value,
                  sort_order: previous ? (previous.sort_order ?? 0) + 1 : 0,
                });
              }}
            >
              <option value="">В конце карты</option>
              {opts.data?.skills
                .filter((skill) => skill.id !== id)
                .map((skill) => (
                  <option value={skill.id} key={skill.id}>
                    После: {skill.name}
                  </option>
                ))}
            </select>
            <small>
              Расположение не изменяет обязательные предыдущие навыки.
            </small>
          </label>
          <Stages
            value={value}
            change={change}
            kind="skills"
            opts={opts.data}
          />
          <label>
            Предыдущий навык
            <select
              multiple
              value={value.requirements}
              onChange={(e) =>
                change({
                  ...value,
                  requirements: Array.from(
                    e.target.selectedOptions,
                    (x) => x.value,
                  ),
                })
              }
            >
              {opts.data?.skills
                .filter((x) => x.id !== id)
                .map((x) => (
                  <option value={x.id} key={x.id}>
                    {optionName(x)}
                  </option>
                ))}
            </select>
            <small>Можно выбрать несколько навыков.</small>
          </label>
        </>
      )}
      {validation && <p className="notice error">{validation}</p>}
      {!!workoutErrors.length&&<div className="notice error validation-summary" role="alert"><b>Нужно исправить {workoutErrors.length} {workoutErrors.length===1?"поле":"поля"}</b><ul>{workoutErrors.map((error,index)=><li key={index}>{error}</li>)}</ul></div>}
      {success && <p className="notice success">Материал сохранён.</p>}
      {save.isError && <ErrorBox error={save.error} />}
      {!readOnly && (
        <div className="publish-actions">
          <button
            type="submit"
            className="secondary-button"
            disabled={save.isPending}
          >
            {save.isPending ? "Сохраняем…" : "Сохранить черновик"}
          </button>
          <button
            type="button"
            className="primary-button"
            disabled={save.isPending}
            onClick={() => submit(true)}
          >
            {save.isPending ? "Публикуем…" : "Сохранить и опубликовать"}
          </button>
        </div>
      )}
    </form>
  );
}
export const WorkoutFields = WorkoutBuilderFields;
function LegacyWorkoutFields({
  value,
  change,
  opts,
}: {
  value: BuilderInput;
  change: (x: BuilderInput) => void;
  opts?: Awaited<ReturnType<typeof coachOptions>>;
}) {
  const [exerciseSearch, setExerciseSearch] = useState("");
  const [exerciseDifficulty, setExerciseDifficulty] = useState("");
  const matchingExercises = (opts?.exercises ?? []).filter(
    (exercise) =>
      exercise.name.toLowerCase().includes(exerciseSearch.trim().toLowerCase()) &&
      (!exerciseDifficulty || exercise.difficulty === exerciseDifficulty),
  );
  const add = () => {
      const used = new Set((value.exercises ?? []).map((x) => x.exercise_id));
      const first = opts?.exercises.find((x) => !used.has(x.id));
      if (first)
        change({
          ...value,
          exercises: [
            ...(value.exercises ?? []),
            {
              exercise_id: first.id,
              sets: 3,
              target_reps: 10,
              rest_seconds: 60,
              sort_order: value.exercises?.length ?? 0,
            },
          ],
        });
    },
    list = value.exercises ?? [];
  const update = (i: number, patch: Partial<BuilderExercise>) =>
    change({
      ...value,
      exercises: list.map((v, j) => (j === i ? { ...v, ...patch } : v)),
    });
  return (
    <>
      <label>
        Продолжительность, минут
        <input
          type="number"
          min="1"
          value={value.estimated_minutes ?? ""}
          onChange={(e) =>
            change({
              ...value,
              estimated_minutes: numericValue(e.target.value),
            })
          }
        />
      </label>
      {value.category !== "warmup" && (
        <label className="toggle-field">
          <span>
            <b>Разминка перед тренировкой</b>
            <small>Перед началом будет предложена стандартная разминка.</small>
          </span>
          <input
            aria-label="Разминка перед тренировкой"
            type="checkbox"
            checked={value.warmup_enabled ?? true}
            onChange={(e) => change({...value, warmup_enabled:e.target.checked, warmup_workout_id:e.target.checked?value.warmup_workout_id:undefined})}
          />
        </label>
      )}
      {value.category !== "warmup" && (value.warmup_enabled ?? true) && (
        <label>
          Выберите разминку
          <select value={value.warmup_workout_id ?? ""} onChange={(e)=>change({...value,warmup_workout_id:e.target.value||undefined})}>
            <option value="">Стандартная разминка</option>
            {opts?.warmups?.map((warmup)=><option value={warmup.id} key={warmup.id}>{optionName(warmup)}</option>)}
          </select>
          <small>Если ничего не выбрано, будет использована стандартная разминка.</small>
        </label>
      )}
      <h3>Упражнения</h3>
      {!!opts?.exercises.length && <>
        <label>
          Поиск по библиотеке
          <input
            type="search"
            value={exerciseSearch}
            onChange={(event) => setExerciseSearch(event.target.value)}
            placeholder="Название упражнения"
          />
          <small>Найдено: {matchingExercises.length}. Системные упражнения доступны всем тренерам.</small>
        </label>
        <label>
          Сложность
          <select value={exerciseDifficulty} onChange={(event) => setExerciseDifficulty(event.target.value)}>
            <option value="">Любая</option>
            <option value="beginner">Начальный</option>
            <option value="intermediate">Средний</option>
            <option value="advanced">Продвинутый</option>
          </select>
        </label>
      </>}
      {list.map((x, i) => {
        const timed = x.target_duration_seconds !== undefined;
        return (
          <div className="card workout-builder-row" key={i}>
            <select
              value={x.exercise_id}
              onChange={(e) => update(i, { exercise_id: e.target.value })}
            >
              {(opts?.exercises ?? []).filter((o) => o.id === x.exercise_id || matchingExercises.some((candidate) => candidate.id === o.id)).map((o) => (
                <option value={o.id} key={o.id}>
                  {optionName(o)} · {o.owner_user_id ? "моё" : "системное"}
                </option>
              ))}
            </select>
            <label>
              Подходов
              <input
                type="number"
                min="1"
                value={x.sets ?? ""}
                onChange={(e) =>
                  update(i, { sets: numericValue(e.target.value) as number })
                }
              />
            </label>
            <label>
              Тип выполнения
              <select
                value={timed ? "time" : "reps"}
                onChange={(e) =>
                  update(i, workoutTarget(e.target.value as "reps" | "time"))
                }
              >
                <option value="reps">Повторения</option>
                <option value="time">Время</option>
              </select>
            </label>
            <label>
              {timed ? "Время, секунд" : "Повторений"}
              <input
                type="number"
                min="1"
                value={
                  (timed ? x.target_duration_seconds : x.target_reps) ?? ""
                }
                onChange={(e) =>
                  update(
                    i,
                    timed
                      ? {
                          target_duration_seconds: numericValue(e.target.value),
                        }
                      : { target_reps: numericValue(e.target.value) },
                  )
                }
              />
            </label>
            <label>
              Отдых, секунд
              <input
                type="number"
                min="0"
                value={x.rest_seconds ?? ""}
                onChange={(e) =>
                  update(i, {
                    rest_seconds: numericValue(e.target.value) as number,
                  })
                }
              />
            </label>
            <label className="wide-field">
              Комментарий тренера
              <input
                value={x.notes ?? ""}
                onChange={(e) =>
                  update(i, { notes: e.target.value || undefined })
                }
                placeholder="Подсказка ученику"
              />
            </label>
            <div className="row-order-actions">
              <button
                type="button"
                disabled={i === 0}
                onClick={() => {
                  const next = [...list];
                  [next[i - 1], next[i]] = [next[i], next[i - 1]];
                  change({
                    ...value,
                    exercises: next.map((item, index) => ({
                      ...item,
                      sort_order: index,
                    })),
                  });
                }}
              >
                ↑ Выше
              </button>
              <button
                type="button"
                disabled={i === list.length - 1}
                onClick={() => {
                  const next = [...list];
                  [next[i], next[i + 1]] = [next[i + 1], next[i]];
                  change({
                    ...value,
                    exercises: next.map((item, index) => ({
                      ...item,
                      sort_order: index,
                    })),
                  });
                }}
              >
                ↓ Ниже
              </button>
            </div>
            <button
              type="button"
              onClick={() =>
                change({
                  ...value,
                  exercises: list
                    .filter((_, j) => j !== i)
                    .map((v, j) => ({ ...v, sort_order: j })),
                })
              }
            >
              Удалить
            </button>
          </div>
        );
      })}
      <button
        type="button"
        className="secondary-button"
        onClick={add}
        disabled={
          !opts?.exercises.some(
            (option) => !list.some((item) => item.exercise_id === option.id),
          )
        }
      >
        + Добавить упражнение
      </button>
    </>
  );
}
function ProgramWorkouts({
  value,
  change,
  opts,
}: {
  value: BuilderInput;
  change: (x: BuilderInput) => void;
  opts?: Awaited<ReturnType<typeof coachOptions>>;
}) {
  const [selecting, setSelecting] = useState(false),
    [search, setSearch] = useState("");
  const selected = value.workouts ?? [];
  const byID = new Map(opts?.workouts.map((x) => [x.id, x]) ?? []);
  const available = (opts?.workouts ?? []).filter(
    (x) =>
      x.owner_user_id &&
      !selected.some((item) => item.workout_id === x.id) &&
      x.name.toLowerCase().includes(search.toLowerCase()),
  );
  const replace = (items: ProgramWorkout[]) =>
    change({ ...value, workouts: normalizeProgramWorkouts(items) });
  return (
    <section className="stack program-workouts">
      <h3>Тренировки</h3>
      {selected.map((item, index) => {
        const workout = byID.get(item.workout_id);
        return (
          <article className="card program-workout-card" key={item.workout_id}>
            <div>
              <b>
                {index + 1}. {workout?.name ?? "Тренировка недоступна"}
              </b>
              <small>
                {difficultyLabel[workout?.difficulty ?? ""] ??
                  workout?.difficulty}{" "}
                · {workout?.minutes ?? "—"} минут ·{" "}
                {workout?.status ? statusLabel[workout.status] : "Недоступна"}
              </small>
            </div>
            <div className="row-order-actions">
              <button
                type="button"
                disabled={!index}
                onClick={() => {
                  replace(moveProgramWorkout(selected, index, -1));
                }}
              >
                ↑
              </button>
              <button
                type="button"
                disabled={index === selected.length - 1}
                onClick={() => {
                  replace(moveProgramWorkout(selected, index, 1));
                }}
              >
                ↓
              </button>
              <button
                type="button"
                onClick={() => replace(selected.filter((_, i) => i !== index))}
              >
                Удалить
              </button>
            </div>
          </article>
        );
      })}
      {!selected.length && (
        <p className="notice">
          Добавьте тренировки, которые войдут в программу.
        </p>
      )}
      <button
        type="button"
        className="secondary-button"
        onClick={() => setSelecting(!selecting)}
      >
        + Добавить тренировку
      </button>
      {selecting && (
        <div className="card workout-selector">
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Поиск тренировки"
          />
          {available.map((workout) => (
            <button
              type="button"
              key={workout.id}
              onClick={() => {
                replace([
                  ...selected,
                  { workout_id: workout.id, sort_order: selected.length },
                ]);
                setSelecting(false);
                setSearch("");
              }}
            >
              <b>{workout.name}</b>
              <small>
                {difficultyLabel[workout.difficulty ?? ""] ??
                  workout.difficulty}{" "}
                · {workout.minutes} минут ·{" "}
                {workout.status ? statusLabel[workout.status] : ""}
              </small>
            </button>
          ))}
          {!available.length && (
            <p>Подходящих тренировок нет. Сначала создайте свою тренировку.</p>
          )}
        </div>
      )}
    </section>
  );
}
function Stages({
  value,
  change,
  kind,
  opts,
}: {
  value: BuilderInput;
  change: (x: BuilderInput) => void;
  kind: "programs" | "skills";
  opts?: Awaited<ReturnType<typeof coachOptions>>;
}) {
  const levels = value.levels ?? [];
  return (
    <>
      <h3>Этапы прогрессии</h3>
      {levels.map((x, i) => (
        <div className="card stage-editor" key={i}>
          <label>
            Название этапа
            <input
              value={x.title}
              onChange={(e) =>
                change({
                  ...value,
                  levels: levels.map((v, j) =>
                    j === i ? { ...v, title: e.target.value } : v,
                  ),
                })
              }
            />
          </label>
          <label>
            Описание
            <textarea
              value={x.description}
              onChange={(e) =>
                change({
                  ...value,
                  levels: levels.map((v, j) =>
                    j === i ? { ...v, description: e.target.value } : v,
                  ),
                })
              }
            />
          </label>
          {kind === "programs" ? (
            <>
              <label>
                Условие открытия
                <select
                  value={x.unlock_rule_type}
                  onChange={(e) =>
                    change({
                      ...value,
                      levels: levels.map((v, j) =>
                        j === i
                          ? { ...v, unlock_rule_type: e.target.value }
                          : v,
                      ),
                    })
                  }
                >
                  <option value="none">Доступен сразу</option>
                  <option value="previous_level">Предыдущий этап</option>
                  <option value="workouts_completed">
                    Количество тренировок
                  </option>
                  <option value="criterion">По критерию</option>
                </select>
              </label>
              <label>
                Значение условия
                <input
                  type="number"
                  min="0"
                  value={x.unlock_rule_value ?? ""}
                  onChange={(e) =>
                    change({
                      ...value,
                      levels: levels.map((v, j) =>
                        j === i
                          ? {
                              ...v,
                              unlock_rule_value: numericValue(
                                e.target.value,
                              ) as number,
                            }
                          : v,
                      ),
                    })
                  }
                />
              </label>
            </>
          ) : (
            <>
              <label>
                Связанный этап программы
                <select
                  value={x.program_level_id ?? ""}
                  onChange={(e) =>
                    change({
                      ...value,
                      levels: levels.map((v, j) =>
                        j === i
                          ? {
                              ...v,
                              program_level_id: e.target.value || undefined,
                            }
                          : v,
                      ),
                    })
                  }
                >
                  <option value="">Без связи</option>
                  {opts?.program_levels.map((option) => (
                    <option value={option.id} key={option.id}>
                      {optionName(option)}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Критерий этапа
                <select
                  value={x.criterion_type}
                  onChange={(e) =>
                    change({
                      ...value,
                      levels: levels.map((v, j) =>
                        j === i ? { ...v, criterion_type: e.target.value } : v,
                      ),
                    })
                  }
                >
                  <option value="repetitions">Повторения</option>
                  <option value="duration_seconds">Время, секунд</option>
                  <option value="workout_completed">
                    Завершение тренировки
                  </option>
                  <option value="manual_confirmation">
                    Ручное подтверждение
                  </option>
                </select>
              </label>
              <label>
                Значение критерия
                <input
                  type="number"
                  min="1"
                  value={x.criterion_value ?? ""}
                  onChange={(e) =>
                    change({
                      ...value,
                      levels: levels.map((v, j) =>
                        j === i
                          ? {
                              ...v,
                              criterion_value: numericValue(
                                e.target.value,
                              ) as number,
                            }
                          : v,
                      ),
                    })
                  }
                />
              </label>
            </>
          )}
          <button
            type="button"
            onClick={() =>
              change({ ...value, levels: levels.filter((_, j) => j !== i) })
            }
          >
            Удалить
          </button>
        </div>
      ))}
      <button
        type="button"
        className="secondary-button"
        onClick={() =>
          change({
            ...value,
            levels: [
              ...levels,
              {
                level_number: levels.length + 1,
                title: `Этап ${levels.length + 1}`,
                description: "Описание этапа",
                difficulty: value.difficulty,
                unlock_rule_type: "none",
                unlock_rule_value: 0,
                criterion_type:
                  kind === "skills" ? "repetitions" : "workout_completed",
                criterion_value: 1,
                sort_order: levels.length,
              },
            ],
          })
        }
      >
        + Добавить этап
      </button>
    </>
  );
}
function Back({ dirty }: { dirty: boolean }) {
  const nav = useNavigate();
  return (
    <button
      type="button"
      className="text-link back-button"
      onClick={() => {
        if (
          shouldLeave(dirty, () =>
            confirm("Есть несохранённые изменения. Выйти без сохранения?"),
          )
        )
          nav("/coach/content");
      }}
    >
      ← Контент
    </button>
  );
}
