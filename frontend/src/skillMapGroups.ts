export const skillMapGroups = [
  {
    id: "basic",
    title: "Базовые навыки",
    description: "Фундамент силы, мобильности и контроля тела",
  },
  {
    id: "floor",
    title: "Пол",
    description: "Упоры, балансы и элементы без снарядов",
  },
  {
    id: "bar",
    title: "Турник",
    description: "Тяговые и силовые элементы на перекладине",
  },
  {
    id: "parallel_bars",
    title: "Брусья",
    description: "Упоры и силовые элементы на брусьях",
  },
] as const;

export type SkillMapGroup = (typeof skillMapGroups)[number]["id"];

export const skillMapGroupTitle = (id?: string) =>
  skillMapGroups.find((group) => group.id === id)?.title ??
  skillMapGroups[0].title;
