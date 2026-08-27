import { describe, expect, it, vi } from "vitest";
import {
  blockNames,
  builderPayload,
  kinds,
  moveProgramWorkout,
  numericValue,
  shouldLeave,
  workoutTarget,
  validateBuilder,
} from "./coachAuthoring";
import { normalizeCoachOptions } from "../api/coach";
describe("Coach authoring UX", () => {
  it("maps internal lesson blocks to human labels", () => {
    expect(blockNames.heading).toBe("Заголовок");
    expect(blockNames.tip).toBe("Совет тренера");
    expect(blockNames.warning).toBe("Важное замечание");
  });
  it("offers all existing content editors", () =>
    expect(kinds.map((x) => x.key)).toEqual([
      "lessons",
      "exercises",
      "workouts",
      "programs",
      "skills",
    ]));
  it("guards unsaved changes", () => {
    const ask = vi.fn(() => false);
    expect(shouldLeave(true, ask)).toBe(false);
    expect(ask).toHaveBeenCalled();
    expect(shouldLeave(false, ask)).toBe(true);
  });
});
describe("coach editor options", () => {
  it("keeps the progression form renderable when empty API collections are null", () => {
    const options = normalizeCoachOptions({
      categories: null,
      exercises: null,
      programs: null,
      program_levels: null,
      workouts: null,
      warmups: null,
      skills: null,
      media: null,
    });
    expect(options.program_levels).toEqual([]);
    expect(options.skills).toEqual([]);
    expect(options.media).toEqual([]);
  });
});
describe("numeric inputs", () => {
  it("keeps an erased value empty and does not prefix the next digit with zero", () => {
    expect(numericValue("")).toBeUndefined();
    expect(numericValue("5")).toBe(5);
    expect(String(numericValue("5"))).toBe("5");
  });
});
describe("program workout ordering", () => {
  const workouts = [
    { workout_id: "a", sort_order: 0 },
    { workout_id: "b", sort_order: 1 },
    { workout_id: "c", sort_order: 2 },
  ];
  it("reorders workouts and persists contiguous sort order", () => {
    expect(moveProgramWorkout(workouts, 2, -1)).toEqual([
      { workout_id: "a", sort_order: 0 },
      { workout_id: "c", sort_order: 1 },
      { workout_id: "b", sort_order: 2 },
    ]);
  });
});
describe("workout targets", () => {
  it("uses exactly one of repetitions or duration", () => {
    expect(workoutTarget("reps")).toEqual({
      target_reps: 10,
      target_duration_seconds: undefined,
    });
    expect(workoutTarget("time")).toEqual({
      target_reps: undefined,
      target_duration_seconds: 30,
    });
  });
});
describe("workout validation", () => {
  const base = {
    title: "Сила",
    description: "План",
    difficulty: "beginner",
    program_id: "p",
    day_number: 2,
    estimated_minutes: 30,
    exercises: [
      {
        exercise_id: "e1",
        sets: 3,
        target_reps: 10,
        rest_seconds: 60,
        sort_order: 0,
      },
    ],
  };
  it("accepts a complete workout", () =>
    expect(validateBuilder("workouts", base)).toBe(""));
  it("explains duplicate exercises", () =>
    expect(
      validateBuilder("workouts", {
        ...base,
        exercises: [...base.exercises, { ...base.exercises[0], sort_order: 1 }],
      }),
    ).toContain("дважды"));
  it("accepts a standalone workout without program fields", () =>
    expect(validateBuilder("workouts", {...base,program_id:undefined,program_level_id:undefined,day_number:undefined})).toBe(""));
  it("allows an empty workout draft but blocks publishing it", () => {
    const draft = { ...base, exercises: [] };
    expect(validateBuilder("workouts", draft)).toBe("");
    expect(validateBuilder("workouts", draft, true)).toContain("упражнение");
  });
});

describe("workout update DTO", () => {
  it("uses the create contract and strips read-only detail fields", () => {
    const payload = builderPayload("workouts", {
      id: "server-id",
      status: "published",
      owner_user_id: "owner-id",
      title: "Сила 2",
      description: "План",
      difficulty: "beginner",
      category: "strength",
      estimated_minutes: 35,
      day_number: 2,
      program_id: "program-id",
      exercises: [{
        exercise_id: "exercise-id",
        sets: 4,
        target_reps: 12,
        rest_seconds: 75,
        sort_order: 0,
        notes: "Строгая техника",
      }],
    } as never);
    expect(payload).toMatchObject({
      title: "Сила 2",
      category: "strength",
      estimated_minutes: 35,
    });
    expect(payload.exercises?.[0]).toMatchObject({target_reps: 12, rest_seconds: 75, notes: "Строгая техника"});
    expect(payload).not.toHaveProperty("id");
    expect(payload).not.toHaveProperty("status");
    expect(payload).not.toHaveProperty("owner_user_id");
    expect(payload).not.toHaveProperty("program_id");
    expect(payload).not.toHaveProperty("program_level_id");
    expect(payload).not.toHaveProperty("day_number");
  });
});
