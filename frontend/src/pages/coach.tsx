import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import {
  action,
  addExternalMedia,
  analytics,
  coachMe,
  content,
  createLesson,
  dashboard,
  media,
  type AnalyticsMetric,
  type LessonBlock,
  type LessonInput,
} from "../api/coach";
import { useSessionStore } from "../store/session";
import "../coach.css";
const sections = [
  ["lessons", "Уроки"],
  ["exercises", "Упражнения"],
  ["workouts", "Тренировки"],
  ["programs", "Программы"],
  ["skills", "Прогрессии"],
  ["media", "Медиафайлы"],
] as const;
function Gate({ children }: { children: (token: string) => React.ReactNode }) {
  const token = useSessionStore((s) => s.accessToken),
    role = useQuery({
      queryKey: ["coach-role"],
      queryFn: () => coachMe(token!),
      enabled: !!token,
      retry: false,
    });
  if (!token) return <div className="notice error">Требуется вход.</div>;
  if (role.isLoading)
    return <div className="notice skeleton">Проверяем роль…</div>;
  if (role.isError)
    return (
      <div className="empty-state">
        <h2>Раздел доступен только тренеру</h2>
      </div>
    );
  return children(token);
}
export function CoachDashboardPage() {
  return <Gate>{(token) => <Dashboard token={token} />}</Gate>;
}
function Dashboard({ token }: { token: string }) {
  const q = useQuery({
    queryKey: ["coach-dashboard"],
    queryFn: () => dashboard(token),
  });
  return (
    <div className="stack">
      <section className="hero-card">
        <p className="eyebrow">COACH STUDIO</p>
        <h2>Панель тренера</h2>
        <span>Создавайте и публикуйте учебный контент.</span>
      </section>
      <div className="coach-grid">
        {sections.map(([key, label]) => (
          <Link
            className="card"
            to={key === "media" ? "/coach/media" : `/coach/content/${key}`}
            key={key}
          >
            <h3>{label}</h3>
            <b>{q.data?.[key] ?? "—"}</b>
            {key === "lessons" && (
              <small>{q.data?.lessons_published ?? 0} опубликовано</small>
            )}
            {key === "workouts" && (
              <small>{q.data?.workouts_published ?? 0} опубликовано</small>
            )}
          </Link>
        ))}
      </div>
      <Link className="card" to="/coach/analytics">
        <h3>Аналитика учеников</h3>
        <p>Агрегированные тренировки, программы и навыки →</p>
      </Link>
    </div>
  );
}
export function CoachContentPage() {
  const { kind = "lessons" } = useParams();
  return <Gate>{(token) => <ContentList token={token} kind={kind} />}</Gate>;
}
function ContentList({ token, kind }: { token: string; kind: string }) {
  const [search, setSearch] = useState(""),
    [status, setStatus] = useState(""),
    [editor, setEditor] = useState(false),
    qc = useQueryClient(),
    q = useQuery({
      queryKey: ["coach-content", kind, search, status],
      queryFn: () => content(token, kind, search, status),
    }),
    act = useMutation({
      mutationFn: ({
        id,
        name,
      }: {
        id: string;
        name: "publish" | "unpublish" | "archive" | "duplicate";
      }) => action(token, kind, id, name),
      onSuccess: () =>
        qc.invalidateQueries({ queryKey: ["coach-content", kind] }),
    });
  return (
    <div className="stack">
      <Link className="text-link" to="/coach">
        ← Панель тренера
      </Link>
      <h2>{sections.find((x) => x[0] === kind)?.[1] ?? kind}</h2>
      <div className="coach-filters">
        <input
          placeholder="Поиск"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <select value={status} onChange={(e) => setStatus(e.target.value)}>
          <option value="">Все статусы</option>
          <option value="draft">Черновики</option>
          <option value="published">Опубликовано</option>
          <option value="archived">Архив</option>
        </select>
      </div>
      {kind === "lessons" && (
        <button className="primary-button" onClick={() => setEditor((v) => !v)}>
          + Создать урок
        </button>
      )}
      {editor && (
        <LessonEditor
          token={token}
          done={() => {
            setEditor(false);
            qc.invalidateQueries({ queryKey: ["coach-content", "lessons"] });
          }}
        />
      )}
      {q.data?.items.map((x) => (
        <article className="card coach-row" key={x.id}>
          <div>
            <h3>{x.name}</h3>
            <p>
              {x.difficulty} · {x.status}
            </p>
          </div>
          <div className="coach-actions">
            {x.status !== "published" && (
              <button onClick={() => act.mutate({ id: x.id, name: "publish" })}>
                Опубликовать
              </button>
            )}
            {x.status === "published" && (
              <button
                onClick={() => act.mutate({ id: x.id, name: "unpublish" })}
              >
                Снять
              </button>
            )}
            <button onClick={() => act.mutate({ id: x.id, name: "archive" })}>
              Архив
            </button>
            {kind === "lessons" && (
              <button
                onClick={() => act.mutate({ id: x.id, name: "duplicate" })}
              >
                Дубликат
              </button>
            )}
          </div>
        </article>
      ))}
    </div>
  );
}
const blank: LessonInput = {
  category_id: "",
  title: "",
  slug: "",
  short_description: "",
  content: "",
  difficulty: "beginner",
  duration_minutes: 5,
  blocks: [],
};
function LessonEditor({ token, done }: { token: string; done: () => void }) {
  const [value, setValue] = useState(blank),
    save = useMutation({
      mutationFn: () => createLesson(token, value),
      onSuccess: done,
    });
  const add = (type: LessonBlock["type"]) =>
      setValue((v) => ({
        ...v,
        blocks: [
          ...v.blocks,
          { type, ...(type === "checklist" ? { items: [""] } : { text: "" }) },
        ],
      })),
    move = (i: number, d: number) =>
      setValue((v) => {
        const b = [...v.blocks],
          [x] = b.splice(i, 1);
        b.splice(i + d, 0, x);
        return { ...v, blocks: b };
      });
  return (
    <form
      className="card coach-editor"
      onSubmit={(e) => {
        e.preventDefault();
        save.mutate();
      }}
    >
      <input
        placeholder="Название"
        value={value.title}
        onChange={(e) => setValue({ ...value, title: e.target.value })}
      />
      <input
        placeholder="Slug"
        value={value.slug}
        onChange={(e) => setValue({ ...value, slug: e.target.value })}
      />
      <input
        placeholder="Category UUID"
        value={value.category_id}
        onChange={(e) => setValue({ ...value, category_id: e.target.value })}
      />
      <textarea
        placeholder="Краткое описание"
        value={value.short_description}
        onChange={(e) =>
          setValue({ ...value, short_description: e.target.value })
        }
      />
      <div className="block-picker">
        {(
          ["heading", "text", "tip", "warning", "checklist", "divider"] as const
        ).map((x) => (
          <button type="button" onClick={() => add(x)} key={x}>
            + {x}
          </button>
        ))}
      </div>
      {value.blocks.map((b, i) => (
        <div className="content-block" key={i}>
          <b>{b.type}</b>
          {b.type !== "divider" && (
            <textarea
              value={b.type === "checklist" ? b.items?.join("\n") : b.text}
              onChange={(e) =>
                setValue((v) => ({
                  ...v,
                  blocks: v.blocks.map((x, j) =>
                    j === i
                      ? b.type === "checklist"
                        ? { ...x, items: e.target.value.split("\n") }
                        : { ...x, text: e.target.value }
                      : x,
                  ),
                }))
              }
            />
          )}
          <button type="button" disabled={i === 0} onClick={() => move(i, -1)}>
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
              setValue((v) => ({
                ...v,
                blocks: v.blocks.filter((_, j) => j !== i),
              }))
            }
          >
            Удалить
          </button>
        </div>
      ))}
      <button className="primary-button">Сохранить черновик</button>
      {save.isError && (
        <p className="notice error">Не удалось сохранить урок.</p>
      )}
    </form>
  );
}
export function CoachMediaPage() {
  return <Gate>{(token) => <Media token={token} />}</Gate>;
}
const imageMimeFromURL = (url: string) => {
  const path = url.split(/[?#]/)[0].toLowerCase();
  return path.endsWith(".png")
    ? "image/png"
    : path.endsWith(".jpg") || path.endsWith(".jpeg")
      ? "image/jpeg"
      : "image/webp";
};
const nameFromURL = (url: string) => {
  try {
    return decodeURIComponent(
      new URL(url).pathname.split("/").filter(Boolean).pop() || "изображение",
    );
  } catch {
    return "изображение";
  }
};
function Media({ token }: { token: string }) {
  const qc = useQueryClient(),
    q = useQuery({ queryKey: ["coach-media"], queryFn: () => media(token) }),
    [url, setURL] = useState(""),
    [error, setError] = useState(""),
    [pendingFile, setPendingFile] = useState<File>(),
    add = useMutation({
      mutationFn: async () => {
        if (pendingFile) {
          if (pendingFile.size > 10 * 1024 * 1024)
            throw new Error("Файл больше 10 МБ");
          if (
            !["image/jpeg", "image/png", "image/webp"].includes(
              pendingFile.type,
            )
          )
            throw new Error("Поддерживаются JPG, PNG и WebP");
          const data = await new Promise<string>((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = () => resolve(String(reader.result));
            reader.onerror = () => reject(reader.error);
            reader.readAsDataURL(pendingFile);
          });
          return addExternalMedia(token, {
            type: "image",
            url: data,
            original_filename: pendingFile.name,
            mime_type: pendingFile.type,
            size_bytes: pendingFile.size,
          });
        }
        if (!url.startsWith("https://"))
          throw new Error("Ссылка должна начинаться с https://");
        return addExternalMedia(token, {
          type: "image",
          url,
          original_filename: nameFromURL(url),
          mime_type: imageMimeFromURL(url),
          size_bytes: 0,
        });
      },
      onSuccess: () => {
        setURL("");
        setPendingFile(undefined);
        setError("");
        qc.invalidateQueries({ queryKey: ["coach-media"] });
        qc.invalidateQueries({ queryKey: ["coach-options"] });
      },
      onError: (e) =>
        setError(
          e instanceof Error ? e.message : "Не удалось добавить изображение",
        ),
    });
  return (
    <div className="stack">
      <Link className="text-link" to="/coach">
        ← Панель тренера
      </Link>
      <h2>Медиатека</h2>
      <section className="card coach-editor">
        <h3>Добавить изображение</h3>
        <label>
          С устройства
          <input
            type="file"
            accept="image/jpeg,image/png,image/webp"
            onChange={(e) => {
              setPendingFile(e.target.files?.[0]);
              setURL("");
              setError("");
            }}
          />
        </label>
        <p>или по прямой HTTPS-ссылке</p>
        <input
          placeholder="https://example.com/photo.jpg"
          value={url}
          onChange={(e) => {
            setURL(e.target.value);
            setPendingFile(undefined);
            setError("");
          }}
        />
        <button
          type="button"
          className="primary-button"
          disabled={add.isPending || (!pendingFile && !url)}
          onClick={() => add.mutate()}
        >
          {add.isPending ? "Добавляем…" : "Добавить изображение"}
        </button>
        {error && <p className="notice error">{error}</p>}
        <small>Файлы: JPG, PNG или WebP, до 10 МБ.</small>
      </section>
      <div className="media-grid">
        {q.data?.media.map((x) => (
          <article className="card" key={x.id}>
            {x.type === "image" ? (
              <img src={x.url} alt={x.original_filename} />
            ) : (
              <video controls preload="metadata" src={x.url} />
            )}
            <small>{x.original_filename}</small>
            <b>{x.references} ссылок</b>
          </article>
        ))}
      </div>
    </div>
  );
}
export function CoachAnalyticsPage() {
  return <Gate>{(token) => <Analytics token={token} />}</Gate>;
}

const plural = (value: number, one: string, few: string, many: string) => {
  const tens = value % 100, units = value % 10;
  if (tens >= 11 && tens <= 14) return many;
  if (units === 1) return one;
  return units >= 2 && units <= 4 ? few : many;
};

const metricDuration = (seconds: number) => {
  const minutes = Math.round(seconds / 60);
  return minutes < 60
    ? `${minutes} мин`
    : `${Math.floor(minutes / 60)} ч ${minutes % 60} мин`;
};

function AnalyticsRanking({
  items,
  kind,
}: {
  items: AnalyticsMetric[];
  kind: "workouts" | "skills" | "achievements";
}) {
  if (!items.length)
    return <div className="analytics-empty">Пока недостаточно данных</div>;
  const max = Math.max(1, ...items.map((item) => item.value));
  return (
    <div className="analytics-ranking">
      {items.map((item, index) => (
        <article className="analytics-rank-row" key={item.name}>
          <span className="analytics-rank">{index + 1}</span>
          <div>
            <header><b>{item.name}</b><strong>{item.value}</strong></header>
            <div className="analytics-bar"><i style={{ width: `${Math.max(4, item.value / max * 100)}%` }} /></div>
            <small>
              {kind === "workouts"
                ? `Среднее время: ${metricDuration(item.secondary)}`
                : kind === "skills"
                  ? `Средний уровень: ${item.secondary.toFixed(1)}`
                  : plural(item.value, "получение", "получения", "получений")}
            </small>
          </div>
        </article>
      ))}
    </div>
  );
}

export function Analytics({ token }: { token: string }) {
  const q = useQuery({
    queryKey: ["coach-analytics"],
    queryFn: () => analytics(token),
  });
  if (q.isLoading)
    return <div className="notice skeleton">Собираем аналитику…</div>;
  if (q.isError || !q.data)
    return <div className="notice error">Не удалось загрузить аналитику.</div>;
  const data = q.data;
  const activeShare = data.total_users
    ? Math.round(data.active_users_30d / data.total_users * 100)
    : 0;
  return (
    <div className="stack coach-analytics">
      <Link className="text-link" to="/coach">
        ← Панель тренера
      </Link>
      <div className="coach-title">
        <div><p className="eyebrow">ОБЩАЯ СТАТИСТИКА</p><h2>Аналитика</h2></div>
        <span className="analytics-period">30 дней</span>
      </div>
      <section className="analytics-kpis" aria-label="Ключевые показатели">
        <article><span>Ученики</span><b>{data.total_users}</b><small>{activeShare}% активны за месяц</small></article>
        <article><span>Тренировки</span><b>{data.total_workouts_completed}</b><small>завершено всего</small></article>
        <article><span>Уроки</span><b>{data.total_lessons_completed}</b><small>изучено учениками</small></article>
      </section>
      <section className="card analytics-activity">
        <header><p className="eyebrow">АКТИВНОСТЬ</p><h3>Последние периоды</h3></header>
        <div className="analytics-period-grid">
          <article><span>7 дней</span><b>{data.active_users_7d}</b><small>{plural(data.active_users_7d, "активный ученик", "активных ученика", "активных учеников")}</small><strong>{data.workouts_7d} тренировок</strong></article>
          <article><span>30 дней</span><b>{data.active_users_30d}</b><small>{plural(data.active_users_30d, "активный ученик", "активных ученика", "активных учеников")}</small><strong>{data.workouts_30d} тренировок</strong></article>
        </div>
      </section>
      <section className="card analytics-section"><header><div><p className="eyebrow">ПОПУЛЯРНОСТЬ</p><h3>Тренировки</h3></div><span>{data.popular_workouts.length}</span></header><AnalyticsRanking items={data.popular_workouts} kind="workouts" /></section>
      <section className="card analytics-section"><header><div><p className="eyebrow">ПРОГРЕСС</p><h3>Навыки учеников</h3></div><span>{data.skill_progress.length}</span></header><AnalyticsRanking items={data.skill_progress} kind="skills" /></section>
      <section className="card analytics-section"><header><div><p className="eyebrow">МОТИВАЦИЯ</p><h3>Достижения</h3></div><span>{data.top_achievements.length}</span></header><AnalyticsRanking items={data.top_achievements} kind="achievements" /></section>
      <p className="analytics-footnote">Показатели агрегированы по всем ученикам платформы.</p>
    </div>
  );
}
