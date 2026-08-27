// @vitest-environment jsdom
import { cleanup, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";
import type { BuilderInput, CoachOptions } from "../api/coach";
import { Stages } from "./coachAuthoring";

const options: CoachOptions = {
  categories: [],
  exercises: [],
  programs: [],
  program_levels: [],
  warmups: [],
  skills: [],
  media: [],
  workouts: [
    {
      id: "10000000-0000-0000-0000-000000000001",
      name: "Горизонт — база",
      status: "published",
      difficulty: "beginner",
      minutes: 25,
      owner_user_id: "90000000-0000-0000-0000-000000000001",
    },
  ],
};

const initial: BuilderInput = {
  name: "Путь к горизонту",
  description: "Программа",
  difficulty: "beginner",
  duration_weeks: 8,
  category: "SKILL",
  levels: [
    {
      level_number: 1,
      title: "База",
      description: "Подготовка",
      difficulty: "beginner",
      unlock_rule_type: "none",
      unlock_rule_value: 0,
      criterion_type: "workout_completed",
      criterion_value: 1,
      sort_order: 0,
      workouts: [],
    },
  ],
};

function Harness() {
  const [value, setValue] = useState(initial);
  return <Stages value={value} change={setValue} kind="programs" opts={options} />;
}

afterEach(cleanup);

describe("Coach program stages", () => {
  it("assigns a workout to a concrete program stage", async () => {
    render(<MemoryRouter><Harness /></MemoryRouter>);
    expect(screen.getByText("Этапы программы")).toBeTruthy();
    await userEvent.click(screen.getByText("+ Добавить тренировку"));
    const picker = screen.getByPlaceholderText("Поиск тренировки").parentElement!;
    await userEvent.click(within(picker).getByText("Горизонт — база"));
    expect(screen.getByText("1. Горизонт — база")).toBeTruthy();
  });

  it("creates the next stage locked by the previous stage", async () => {
    render(<MemoryRouter><Harness /></MemoryRouter>);
    await userEvent.click(screen.getByText("+ Добавить этап"));
    const rules = screen.getAllByLabelText("Условие открытия") as HTMLSelectElement[];
    expect(rules).toHaveLength(2);
    expect(rules[1].value).toBe("previous_level");
  });
});
