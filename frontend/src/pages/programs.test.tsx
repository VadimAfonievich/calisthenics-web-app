// @vitest-environment jsdom
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { useSessionStore } from "../store/session";
import { ProgramDetailPage, ProgramsPage } from "./programs";

const mocks = vi.hoisted(() => ({ listPrograms: vi.fn(), getProgram: vi.fn() }));
vi.mock("../api/programs", () => mocks);
const id = "40000000-0000-0000-0000-000000000001";
const program = { id, name: "Базовая сила", slug: "base", description: "Силовой фундамент", difficulty: "beginner", duration_weeks: 4, category: "BASE_STRENGTH", workout_count: 1 };
const renderPage = (element: ReactNode, path = "/programs") => render(<QueryClientProvider client={new QueryClient({defaultOptions:{queries:{retry:false}}})}><MemoryRouter initialEntries={[path]}><Routes><Route path="/programs" element={element}/><Route path="/programs/:id" element={element}/></Routes></MemoryRouter></QueryClientProvider>);

beforeEach(() => { useSessionStore.setState({accessToken:"token",status:"authenticated"}); vi.clearAllMocks(); });
afterEach(cleanup);

describe("student programs", () => {
  it("shows published programs and their workout count", async () => { mocks.listPrograms.mockResolvedValue({programs:[program]}); renderPage(<ProgramsPage/>); expect(await screen.findByText("Базовая сила")).toBeTruthy(); expect(screen.getByText("1 тренировок")).toBeTruthy(); });
  it("shows ordered program workouts on detail", async () => { mocks.getProgram.mockResolvedValue({program:{...program,workouts:[{id:"50000000-0000-0000-0000-000000000001",title:"Тренировка 1",description:"Начало",difficulty:"beginner",estimated_minutes:25,category:"strength"}]}}); renderPage(<ProgramDetailPage/>,`/programs/${id}`); expect(await screen.findByText("Тренировка 1")).toBeTruthy(); expect(screen.getByText("25 мин")).toBeTruthy(); });
});
