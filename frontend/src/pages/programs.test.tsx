// @vitest-environment jsdom
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { useSessionStore } from "../store/session";
import { ProgramDetailPage, ProgramsPage } from "./programs";

const mocks = vi.hoisted(() => ({ listPrograms: vi.fn(), getProgram: vi.fn(),startProgram:vi.fn(),confirmStageMastery:vi.fn() }));
vi.mock("../api/programs", () => mocks);
const id = "40000000-0000-0000-0000-000000000001";
const program = { id, name: "Базовая сила", slug: "base", description: "Силовой фундамент", difficulty: "beginner", duration_weeks: 4, category: "BASE_STRENGTH", workout_count: 1 };
const renderPage = (element: ReactNode, path = "/programs") => render(<QueryClientProvider client={new QueryClient({defaultOptions:{queries:{retry:false}}})}><MemoryRouter initialEntries={[path]}><Routes><Route path="/programs" element={element}/><Route path="/programs/:id" element={element}/></Routes></MemoryRouter></QueryClientProvider>);

beforeEach(() => { useSessionStore.setState({accessToken:"token",status:"authenticated"}); vi.clearAllMocks(); });
afterEach(cleanup);

describe("student programs", () => {
  it("shows published programs and their workout count", async () => { mocks.listPrograms.mockResolvedValue({programs:[program]}); renderPage(<ProgramsPage/>); expect(await screen.findByText("Базовая сила")).toBeTruthy(); expect(screen.getByText("1 тренировок")).toBeTruthy(); });
  it("shows ordered program workouts on detail", async () => { mocks.getProgram.mockResolvedValue({program:{...program,workouts:[{id:"50000000-0000-0000-0000-000000000001",title:"Тренировка 1",description:"Начало",difficulty:"beginner",estimated_minutes:25,category:"strength"}]}}); renderPage(<ProgramDetailPage/>,`/programs/${id}`); expect(await screen.findByText("Тренировка 1")).toBeTruthy(); expect(screen.getByText("25 мин")).toBeTruthy(); });
  it("groups workouts by program stage", async () => { mocks.getProgram.mockResolvedValue({program:{...program,levels:[{id:"level-1",level_number:1,title:"База горизонта",description:"Подготовка прямых рук",difficulty:"beginner",unlock_rule_type:"none",unlock_rule_value:0,workouts:[{id:"50000000-0000-0000-0000-000000000001",title:"Сила прямых рук",description:"Начало",difficulty:"beginner",estimated_minutes:25,category:"skill"}]}]}}); renderPage(<ProgramDetailPage/>,`/programs/${id}`); expect((await screen.findAllByText("База горизонта")).length).toBeGreaterThan(0); expect(screen.getByText("Сила прямых рук")).toBeTruthy(); });
  it("shows mastery goal and confirms the current stage",async()=>{const user=userEvent.setup(),levelID="41000000-0000-0000-0000-000000000001";mocks.getProgram.mockResolvedValue({program:{...program,progress_status:"active",current_level:1,current_stage:"Накат",levels:[{id:levelID,level_number:1,title:"Накат",description:"Практика",difficulty:"beginner",unlock_rule_type:"none",unlock_rule_value:0,mastery_type:"duration",mastery_value:15,mastery_name:"Накат",mastery_description:"Удержи Накат не менее 15 секунд с правильной техникой.",status:"current",workouts:[{id:"50000000-0000-0000-0000-000000000001",title:"Практика Наката",description:"Повторяй",difficulty:"beginner",estimated_minutes:10,category:"skill"}]}]}});mocks.confirmStageMastery.mockResolvedValue({mastery:{program_id:id,program_level_id:levelID,status:"active",current_level:2,mastered_at:"now",program_completed:false}});renderPage(<ProgramDetailPage/>,`/programs/${id}`);expect(await screen.findByText("15 сек")).toBeTruthy();expect(screen.getByText("Ещё не подтверждено")).toBeTruthy();expect(screen.getByText("Повторить тренировку")).toBeTruthy();await user.click(screen.getByRole("button",{name:"Подтвердить навык"}));expect(screen.getByRole("alertdialog").textContent).toContain("После подтверждения откроется следующий этап");await user.click(screen.getByRole("button",{name:"Да, навык освоен"}));expect(mocks.confirmStageMastery).toHaveBeenCalledWith("token",id,levelID)})
});
