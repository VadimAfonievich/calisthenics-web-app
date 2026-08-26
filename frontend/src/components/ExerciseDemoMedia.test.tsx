// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ExerciseDemoMedia } from "./ExerciseDemoMedia";

beforeEach(() =>
  vi.stubGlobal(
    "matchMedia",
    vi.fn().mockReturnValue({
      matches: false,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }),
  ),
);
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe("ExerciseDemoMedia", () => {
  it("renders a muted looping inline video", () => {
    render(
      <ExerciseDemoMedia
        media={{
          url: "https://example.com/demo.mp4",
          type: "video",
          mime_type: "video/mp4",
        }}
      />,
    );
    const video = screen.getByLabelText(
      "Демонстрация упражнения",
    ) as HTMLVideoElement;
    expect(video.autoplay).toBe(true);
    expect(video.muted).toBe(true);
    expect(video.loop).toBe(true);
    expect(video.playsInline).toBe(true);
    expect(video.controls).toBe(false);
  });

  it.each(["image/gif", "image/webp"])("renders an %s image", (mimeType) => {
    render(
      <ExerciseDemoMedia
        media={{
          url: `https://example.com/demo.${mimeType === "image/gif" ? "gif" : "webp"}`,
          type: "image",
          mime_type: mimeType,
        }}
      />,
    );
    expect(screen.getByRole("img")).toBeTruthy();
  });

  it("renders missing and video-error fallbacks", () => {
    const { rerender } = render(<ExerciseDemoMedia />);
    expect(screen.getByText("Демонстрация пока не добавлена")).toBeTruthy();
    rerender(
      <ExerciseDemoMedia
        media={{
          url: "https://example.com/broken.mp4",
          type: "video",
          mime_type: "video/mp4",
        }}
      />,
    );
    fireEvent.error(screen.getByLabelText("Демонстрация упражнения"));
    expect(screen.getByText("Демонстрация пока не добавлена")).toBeTruthy();
  });

  it("uses the poster for reduced motion", () => {
    vi.stubGlobal(
      "matchMedia",
      vi.fn().mockReturnValue({
        matches: true,
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
      }),
    );
    render(
      <ExerciseDemoMedia
        media={{
          url: "https://example.com/demo.mp4",
          poster_url: "https://example.com/poster.webp",
          type: "video",
          mime_type: "video/mp4",
        }}
      />,
    );
    expect(screen.getByRole("img").getAttribute("src")).toContain("poster");
  });

  it("does not reserve space for missing compact media", () => {
    const { container } = render(<ExerciseDemoMedia compact />);
    expect(container.innerHTML).toBe("");
  });
});
