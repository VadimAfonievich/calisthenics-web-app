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
function Analytics({ token }: { token: string }) {
  const q = useQuery({
    queryKey: ["coach-analytics"],
    queryFn: () => analytics(token),
  });
  return (
    <div className="stack">
      <Link className="text-link" to="/coach">
        ← Панель тренера
      </Link>
      <h2>Аналитика</h2>
      <pre className="card analytics-json">
        {JSON.stringify(q.data ?? {}, null, 2)}
      </pre>
      <p className="notice">
        Финансовая аналитика появится после подключения оплат.
      </p>
    </div>
  );
}
