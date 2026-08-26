import { useEffect, useState } from "react";

export type DemoMedia = {
  url: string;
  type: "video" | "image";
  mime_type: string;
  poster_url?: string;
};

export function ExerciseDemoMedia({
  media,
  compact = false,
  label = "Демонстрация упражнения",
}: {
  media?: DemoMedia;
  compact?: boolean;
  label?: string;
}) {
  const [failed, setFailed] = useState(false);
  const [reduced, setReduced] = useState(false);

  useEffect(() => setFailed(false), [media?.url]);
  useEffect(() => {
    const query = window.matchMedia?.("(prefers-reduced-motion: reduce)");
    if (!query) return;
    const sync = () => setReduced(query.matches);
    sync();
    query.addEventListener?.("change", sync);
    return () => query.removeEventListener?.("change", sync);
  }, []);

  if (!media || failed) {
    return compact ? null : (
      <div className="exercise-demo-placeholder">
        Демонстрация пока не добавлена
      </div>
    );
  }
  if (media.type === "image" || reduced) {
    return media.poster_url || media.type === "image" ? (
      <img
        className={`exercise-demo ${compact ? "compact" : ""}`}
        src={media.poster_url || media.url}
        alt={label}
        loading="lazy"
        onError={() => setFailed(true)}
      />
    ) : (
      <div className="exercise-demo-placeholder">
        Анимация отключена в настройках движения
      </div>
    );
  }
  return (
    <video
      className={`exercise-demo ${compact ? "compact" : ""}`}
      aria-label={label}
      src={media.url}
      poster={media.poster_url}
      autoPlay
      muted
      loop
      playsInline
      preload="metadata"
      onError={() => setFailed(true)}
    />
  );
}
