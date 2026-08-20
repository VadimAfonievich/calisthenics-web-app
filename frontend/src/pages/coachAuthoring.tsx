import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate, useParams } from "react-router-dom";
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
} from "../api/coach";
import { useSessionStore } from "../store/session";
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
        {items.map((x) => (
          <ContentCard
            key={`${x.kind}-${x.id}`}
            item={x}
            kind={x.kind}
            manageAll={manageAll}
          />
        ))}
        {!q.isLoading && !items.length && (
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
        <div className="notice error">Не удалось загрузить материалы.</div>
      )}
      {q.data?.items.map((x) => (
        <ContentCard item={x} kind={k.key} key={x.id} manageAll={manageAll} />
      ))}
      {!q.isLoading && !q.data?.items.length && (
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
    if (!readOnly) nav(`/coach/${kind}/${item.id}/edit`);
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
        <b>{readOnly ? "Только чтение" : "Редактировать ›"}</b>
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
export function CoachEditor() {
  const { kind = "lessons", id = "new" } = useParams();
  const k = kind as ContentKind;
  if (!kinds.some((x) => x.key === k))
    return <div className="notice error">Неизвестный тип материала.</div>;
  return k === "lessons" ? (
    <LessonEditor id={id} />
  ) : (
    <BuilderEditor kind={k} id={id} />
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
function LessonEditor({ id }: { id: string }) {
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
      mutationFn: () =>
        edit ? updateLesson(token(), id, value) : createLesson(token(), value),
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
  const submit = () => {
    if (!value.title.trim()) return setValidation("Введите название урока.");
    if (!value.category_id) return setValidation("Выберите категорию урока.");
    if (!value.short_description.trim())
      return setValidation("Добавьте краткое описание урока.");
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
    save.mutate();
  };
  return (
    <form
      className="stack coach-editor-page"
      onSubmit={(e) => {
        e.preventDefault();
        submit();
      }}
    >
      <Back dirty={dirty} />
      <h2>{edit ? "Редактировать урок" : "Новый урок"}</h2>
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
            value={value.duration_minutes}
            onChange={(e) =>
              change({ ...value, duration_minutes: +e.target.value })
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
      <button className="primary-button sticky-save" disabled={save.isPending}>
        {save.isPending ? "Сохраняем…" : "Сохранить урок"}
      </button>
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
      }
    : {}),
  ...(kind === "workouts"
    ? {
        title: "",
        estimated_minutes: 20,
        day_number: 1,
        program_id: "",
        exercises: [],
      }
    : {}),
  ...(kind === "programs"
    ? { name: "", duration_weeks: 4, category: "OTHER", levels: [] }
    : {}),
  ...(kind === "skills"
    ? {
        name: "",
        category: "SKILL",
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
) {
  if (!(value.name ?? value.title)?.trim()) return "Введите название.";
  if (!value.description.trim()) return "Добавьте описание.";
  if (kind === "exercises") {
    if (!value.instructions?.trim()) return "Добавьте инструкцию выполнения.";
    if (!value.muscle_groups?.some((x) => x.trim()))
      return "Укажите хотя бы одну группу мышц.";
  }
  if (kind === "workouts") {
    if (!value.program_id) return "Выберите программу.";
    if (!value.day_number || value.day_number < 1)
      return "Укажите номер дня программы.";
    if (!value.estimated_minutes || value.estimated_minutes < 1)
      return "Укажите продолжительность тренировки.";
    if (!value.exercises?.length) return "Добавьте хотя бы одно упражнение.";
    const ids = value.exercises.map((x) => x.exercise_id);
    if (new Set(ids).size !== ids.length)
      return "Одно упражнение нельзя добавить в тренировку дважды.";
    if (
      value.exercises.some(
        (x) =>
          x.sets < 1 ||
          x.rest_seconds < 0 ||
          ((x.target_reps ?? 0) < 1 && (x.target_duration_seconds ?? 0) < 1),
      )
    )
      return "Проверьте подходы, повторения или время и отдых у каждого упражнения.";
  }
  if (
    kind === "programs" &&
    (!value.duration_weeks || value.duration_weeks < 1)
  )
    return "Укажите продолжительность программы.";
  if (kind === "skills") {
    if (!value.icon?.trim()) return "Добавьте значок прогрессии.";
    if (!value.final_criterion_value || value.final_criterion_value < 1)
      return "Укажите итоговый критерий прогрессии.";
    if (!value.levels?.length) return "Добавьте хотя бы один этап прогрессии.";
  }
  return "";
}
function BuilderEditor({
  kind,
  id,
}: {
  kind: Exclude<ContentKind, "lessons">;
  id: string;
}) {
  const edit = id !== "new",
    nav = useNavigate(),
    [value, setValue] = useState(() => blankBuilder(kind)),
    [dirty, setDirty] = useState(false),
    [success, setSuccess] = useState(false),
    [validation, setValidation] = useState("");
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
      const clean = { ...item };
      delete clean.levels;
      delete clean.exercises;
      setValue({
        ...blankBuilder(kind),
        ...clean,
        ...(levels ? { levels } : {}),
        ...(exercises ? { exercises } : {}),
      } as BuilderInput);
    }
  }, [detail.data, kind]);
  const change = (x: BuilderInput) => {
      setValue(x);
      setDirty(true);
      setValidation("");
    },
    save = useMutation({
      mutationFn: async (publish: boolean) => {
        const result = await saveBuilder(
          token(),
          kind,
          value,
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
    const error = validateBuilder(kind, value);
    if (error) return setValidation(error);
    save.mutate(publish);
  };
  const cover = String(
    (value as BuilderInput & { cover_media_id?: string }).cover_media_id ?? "",
  );
  return (
    <form
      className="stack coach-editor-page"
      onSubmit={(e) => {
        e.preventDefault();
        submit(false);
      }}
    >
      <Back dirty={dirty} />
      <h2>{edit ? `Редактировать: ${title.one}` : `Новая: ${title.one}`}</h2>
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
            Как измерять выполнение
            <select
              value={value.movement_type}
              onChange={(e) =>
                change({ ...value, movement_type: e.target.value })
              }
            >
              <option value="reps">Количество повторений</option>
              <option value="duration">Время исполнения</option>
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
        </>
      )}
      {kind === "workouts" && (
        <WorkoutFields value={value} change={change} opts={opts.data} />
      )}{" "}
      {kind === "programs" && (
        <>
          <label>
            Продолжительность, недель
            <input
              type="number"
              min="1"
              value={value.duration_weeks}
              onChange={(e) =>
                change({ ...value, duration_weeks: +e.target.value })
              }
            />
          </label>
          <Stages value={value} change={change} kind="programs" />
        </>
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
                value={value.xp_reward ?? 0}
                onChange={(e) =>
                  change({ ...value, xp_reward: +e.target.value })
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
              value={value.final_criterion_value ?? 1}
              onChange={(e) =>
                change({ ...value, final_criterion_value: +e.target.value })
              }
            />
          </label>
          <Stages value={value} change={change} kind="skills" />
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
                    {x.name}
                  </option>
                ))}
            </select>
            <small>Можно выбрать несколько навыков.</small>
          </label>
        </>
      )}
      {validation && <p className="notice error">{validation}</p>}
      {success && <p className="notice success">Материал сохранён.</p>}
      {save.isError && <ErrorBox error={save.error} />}
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
    </form>
  );
}
function WorkoutFields({
  value,
  change,
  opts,
}: {
  value: BuilderInput;
  change: (x: BuilderInput) => void;
  opts?: Awaited<ReturnType<typeof coachOptions>>;
}) {
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
        Программа
        <select
          value={value.program_id}
          onChange={(e) => change({ ...value, program_id: e.target.value })}
        >
          <option value="">Выберите программу</option>
          {opts?.programs.map((x) => (
            <option value={x.id} key={x.id}>
              {x.name}
            </option>
          ))}
        </select>
      </label>
      <label>
        День программы
        <input
          type="number"
          min="1"
          value={value.day_number ?? 1}
          onChange={(e) => change({ ...value, day_number: +e.target.value })}
        />
        <small>
          Номер дня должен быть уникальным внутри выбранной программы.
        </small>
      </label>
      <label>
        Продолжительность, минут
        <input
          type="number"
          min="1"
          value={value.estimated_minutes}
          onChange={(e) =>
            change({ ...value, estimated_minutes: +e.target.value })
          }
        />
      </label>
      <h3>Упражнения</h3>
      {list.map((x, i) => {
        const timed = x.target_duration_seconds !== undefined;
        return (
          <div className="card workout-builder-row" key={i}>
            <select
              value={x.exercise_id}
              onChange={(e) => update(i, { exercise_id: e.target.value })}
            >
              {opts?.exercises.map((o) => (
                <option value={o.id} key={o.id}>
                  {o.name}
                </option>
              ))}
            </select>
            <label>
              Подходов
              <input
                type="number"
                min="1"
                value={x.sets}
                onChange={(e) => update(i, { sets: +e.target.value })}
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
                value={timed ? x.target_duration_seconds : x.target_reps}
                onChange={(e) =>
                  update(
                    i,
                    timed
                      ? { target_duration_seconds: +e.target.value }
                      : { target_reps: +e.target.value },
                  )
                }
              />
            </label>
            <label>
              Отдых, секунд
              <input
                type="number"
                min="0"
                value={x.rest_seconds}
                onChange={(e) => update(i, { rest_seconds: +e.target.value })}
              />
            </label>
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
function Stages({
  value,
  change,
  kind,
}: {
  value: BuilderInput;
  change: (x: BuilderInput) => void;
  kind: "programs" | "skills";
}) {
  const levels = value.levels ?? [];
  return (
    <>
      <h3>{kind === "skills" ? "Этапы прогрессии" : "Этапы программы"}</h3>
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
                  value={x.unlock_rule_value}
                  onChange={(e) =>
                    change({
                      ...value,
                      levels: levels.map((v, j) =>
                        j === i
                          ? { ...v, unlock_rule_value: +e.target.value }
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
                  value={x.criterion_value}
                  onChange={(e) =>
                    change({
                      ...value,
                      levels: levels.map((v, j) =>
                        j === i
                          ? { ...v, criterion_value: +e.target.value }
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
