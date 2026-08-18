# CALISTHENICS TELEGRAM MINI APP — MASTER DEVELOPMENT SPECIFICATION

## 0. Роль Codex

Ты — основной software engineer проекта. Твоя задача — разработать полностью рабочий прототип Telegram Mini App для обучения калистенике.

Работай как инженер в существующем Git-репозитории:
- сначала изучи репозиторий и окружение;
- не ломай существующий рабочий код без необходимости;
- все решения должны быть воспроизводимыми;
- после каждого крупного этапа запускай проверки;
- не переходи к следующему этапу, если текущий этап не работает;
- не оставляй TODO вместо критической функциональности;
- если для продолжения требуется выбор, выбирай наиболее простой production-oriented вариант и фиксируй решение в документации;
- не добавляй микросервисы без реальной необходимости.

Главный принцип: сначала рабочий MVP, затем расширение.

---

# 1. Цель проекта

Создать Telegram Mini App для обучения калистенике.

Пользователь:
1. открывает Telegram-бота;
2. запускает Mini App;
3. автоматически авторизуется через Telegram;
4. получает профиль и текущую программу;
5. изучает уроки;
6. выполняет тренировки;
7. отмечает подходы;
8. получает XP;
9. отслеживает прогресс;
10. получает достижения.

Приложение должно быть пригодно для дальнейшего расширения в полноценный продукт.

---

# 2. MVP

В MVP обязательно реализовать:

- Telegram authentication;
- Telegram Mini App;
- Telegram Bot;
- пользовательский профиль;
- главную страницу;
- каталог уроков;
- страницу урока;
- каталог упражнений;
- страницу упражнения;
- программы тренировок;
- тренировку;
- подходы и повторения;
- завершение тренировки;
- историю тренировок;
- XP;
- уровни;
- streak;
- статистику;
- достижения;
- базовую админ-панель;
- PostgreSQL;
- Redis;
- Docker Compose;
- Swagger/OpenAPI;
- миграции;
- SQLC;
- тесты критической backend-логики.

Не реализовывать в MVP:
- социальную сеть;
- друзей;
- чаты;
- рейтинги;
- AI-тренера;
- Apple Health;
- Google Fit;
- wearable-интеграции;
- платежи и подписки;
- Kafka;
- Kubernetes;
- микросервисную архитектуру.

---

# 3. Архитектура

Использовать modular monolith.

Схема:

Telegram
    |
    +--> Telegram Bot
    |
    +--> Telegram Mini App
              |
              | REST API
              v
         Go Backend
           |     |
           v     v
      PostgreSQL Redis

Не создавать отдельные микросервисы.

Backend:

- Go;
- Gin или Chi — выбрать один фреймворк и использовать последовательно;
- PostgreSQL;
- SQLC;
- Redis;
- JWT;
- golang-migrate;
- Swagger/OpenAPI;
- Docker.

Frontend:

- React;
- TypeScript;
- Vite;
- React Router;
- TanStack Query;
- Zustand только там, где действительно нужен client state;
- Tailwind CSS или другой легковесный UI-подход;
- Telegram WebApp SDK/API.

---

# 4. Репозиторий

Если репозиторий пустой — создать структуру:

backend/
frontend/
docker/
docs/

Backend:

backend/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── auth/
│   ├── users/
│   ├── lessons/
│   ├── exercises/
│   ├── workouts/
│   ├── progress/
│   ├── achievements/
│   ├── stats/
│   └── admin/
├── handler/
├── service/
├── repository/
├── middleware/
├── telegram/
├── config/
├── db/
│   ├── migrations/
│   ├── queries/
│   └── sqlc/
└── docs/

Frontend:

frontend/
├── src/
│   ├── app/
│   ├── pages/
│   ├── components/
│   ├── api/
│   ├── hooks/
│   ├── store/
│   ├── types/
│   └── utils/
└── public/

---

# 5. Основные домены

## Users

Пользователь Telegram.

Хранить минимум:

- id;
- telegram_id;
- username;
- first_name;
- last_name;
- photo_url;
- created_at;
- updated_at.

Профиль:

- display_name;
- level;
- xp;
- current_streak;
- longest_streak;
- preferred difficulty;
- created_at;
- updated_at.

---

# 6. Lessons

Уроки — образовательный контент.

Категории:

- Основы;
- Техника;
- Разминка;
- Восстановление;
- Базовые упражнения;
- Продвинутые элементы.

Lesson:

- id;
- category_id;
- title;
- slug;
- short_description;
- content;
- video_url;
- image_url;
- difficulty;
- duration_minutes;
- sort_order;
- published;
- created_at;
- updated_at.

User lesson progress:

- user_id;
- lesson_id;
- completed;
- completed_at;
- progress_percent.

---

# 7. Exercises

Exercise:

- id;
- name;
- slug;
- description;
- instructions;
- common_mistakes;
- difficulty;
- muscle_groups;
- equipment;
- video_url;
- image_url;
- created_at;
- updated_at.

Примеры:

- Push Up;
- Pull Up;
- Squat;
- Plank;
- Dips;
- Australian Pull Up;
- Handstand;
- Muscle Up;
- Front Lever.

Не зашивать упражнения в код. Они должны храниться в БД.

---

# 8. Training Programs

Program:

- id;
- name;
- slug;
- description;
- difficulty;
- duration_weeks;
- published;
- created_at;
- updated_at.

Примеры:

- Старт с нуля;
- База калистеники;
- 10 подтягиваний;
- Muscle Up;
- Handstand;
- Front Lever.

Workout:

- id;
- program_id;
- day_number;
- title;
- description;
- estimated_minutes;
- sort_order.

Workout exercise:

- id;
- workout_id;
- exercise_id;
- sets;
- target_reps;
- target_duration_seconds;
- rest_seconds;
- sort_order;
- notes.

---

# 9. Workout execution

Пользователь начинает тренировку.

Создаётся workout_session:

- id;
- user_id;
- workout_id;
- started_at;
- completed_at;
- status;
- duration_seconds;
- xp_earned.

Для каждого подхода:

exercise_set:

- id;
- session_id;
- exercise_id;
- set_number;
- reps;
- duration_seconds;
- weight;
- completed;
- created_at.

Пользователь должен иметь возможность:
- начать тренировку;
- выполнить подход;
- пропустить подход;
- изменить фактическое количество повторений;
- завершить упражнение;
- завершить тренировку.

---

# 10. Главный экран

После авторизации:

Приветствие пользователя.

Показать:

- текущий уровень;
- XP;
- прогресс до следующего уровня;
- текущую серию;
- сегодняшнюю тренировку;
- количество тренировок;
- ближайшее достижение.

Главная CTA:

"Начать тренировку".

Нижняя навигация:

- Главная;
- Уроки;
- Прогресс;
- Профиль.

---

# 11. Экран уроков

Категории + список уроков.

Карточка:

- название;
- описание;
- difficulty;
- длительность;
- статус прохождения.

Страница урока:

- заголовок;
- изображение;
- видео;
- текст;
- инструкции;
- кнопка "Отметить как пройдено".

---

# 12. Экран упражнений

Каталог упражнений.

Фильтры:

- difficulty;
- muscle group;
- equipment.

Страница:

- название;
- видео/изображение;
- описание;
- техника;
- типичные ошибки;
- мышцы;
- оборудование.

---

# 13. Экран тренировки

Показывать:

- название;
- длительность;
- список упражнений;
- текущий прогресс.

Для упражнения:

Push Ups

3 подхода × 10 повторений

Set 1
10 [✓]

Set 2
10 [✓]

Set 3
8 [✓]

Кнопка:
"Следующее упражнение"

Для упражнений по времени использовать timer.

После завершения:

"Тренировка завершена!"

Показать:
- длительность;
- упражнения;
- количество подходов;
- XP;
- текущий streak.

---

# 14. Progress

Экран прогресса:

- уровень;
- XP;
- XP до следующего уровня;
- current streak;
- longest streak;
- всего тренировок;
- завершённых упражнений;
- общее время тренировок;
- история тренировок;
- график тренировок по неделям.

В будущем архитектура должна позволять добавить:
- максимум подтягиваний;
- максимум отжиманий;
- время планки;
- вес;
- фотографии прогресса.

---

# 15. XP

Базовые правила:

Lesson completed: +20 XP
Workout completed: +100 XP
7-day streak: +200 XP
30-day streak: +500 XP

Сделать XP calculation отдельным сервисом, а не размазывать формулы по handler'ам.

---

# 16. Levels

Базовая система:

Level 1 — Beginner
Level 2 — Novice
Level 3 — Intermediate
Level 4 — Advanced
Level 5 — Athlete

Пороги XP вынести в конфигурацию или таблицу БД.

Не делать уровни hardcoded внутри UI.

---

# 17. Streak

Streak считается по дням, в которые пользователь завершил хотя бы одну тренировку.

Нужно корректно обработать:
- повторную тренировку в тот же день;
- продолжение серии;
- пропуск дня;
- первый тренировочный день;
- timezone пользователя.

Timezone пользователя сохранить в профиле.

---

# 18. Achievements

Система достижений должна быть расширяемой.

Achievement:

- id;
- code;
- title;
- description;
- icon;
- xp_reward;
- condition_type;
- condition_value.

Примеры:

FIRST_WORKOUT
TEN_WORKOUTS
FIFTY_WORKOUTS
HUNDRED_WORKOUTS
SEVEN_DAY_STREAK
THIRTY_DAY_STREAK
FIRST_PULL_UP
FIRST_MUSCLE_UP
FIRST_HANDSTAND

User achievement:

- user_id;
- achievement_id;
- unlocked_at.

Проверку достижений вынести в отдельный service.

---

# 19. Authentication

Использовать Telegram Mini App initData.

Backend обязан валидировать подпись Telegram.

Flow:

Telegram
  ↓
Mini App
  ↓
initData
  ↓
POST /api/v1/auth/telegram
  ↓
validate Telegram signature
  ↓
find/create user
  ↓
issue JWT
  ↓
frontend stores auth state
  ↓
authenticated API calls

Не принимать telegram_id от frontend как доверенный идентификатор.

Telegram initData должен проверяться на backend.

JWT:
- access token;
- разумный expiration;
- secret из environment;
- middleware для защищённых endpoints.

---

# 20. Telegram Bot

Минимум:

/start

Ответ:

Привет!

Добро пожаловать в Calisthenics Coach.

Тренируйся, изучай технику и отслеживай прогресс.

Кнопка:
"Открыть приложение"

Бот должен открывать Mini App через Telegram WebApp.

Позже предусмотреть:
- напоминания;
- уведомление о тренировке;
- поздравление с достижением.

---

# 21. API

Использовать versioning:

/api/v1/

Auth:

POST /api/v1/auth/telegram

User:

GET /api/v1/me
GET /api/v1/me/profile

Lessons:

GET /api/v1/lessons
GET /api/v1/lessons/:id
POST /api/v1/lessons/:id/complete

Exercises:

GET /api/v1/exercises
GET /api/v1/exercises/:id

Programs:

GET /api/v1/programs
GET /api/v1/programs/:id

Workouts:

GET /api/v1/workouts/today
GET /api/v1/workouts/:id
POST /api/v1/workouts/:id/start
POST /api/v1/workouts/:id/complete

Workout sets:

POST /api/v1/workout-sessions/:id/sets

Progress:

GET /api/v1/progress
GET /api/v1/stats
GET /api/v1/history

Achievements:

GET /api/v1/achievements

---

# 22. Admin API

Admin authentication.

Lessons:

POST /api/v1/admin/lessons
PUT /api/v1/admin/lessons/:id
DELETE /api/v1/admin/lessons/:id
POST /api/v1/admin/lessons/:id/publish

Exercises:

POST /api/v1/admin/exercises
PUT /api/v1/admin/exercises/:id
DELETE /api/v1/admin/exercises/:id

Programs:

POST /api/v1/admin/programs
PUT /api/v1/admin/programs/:id
DELETE /api/v1/admin/programs/:id

Workouts:

POST /api/v1/admin/workouts
PUT /api/v1/admin/workouts/:id
DELETE /api/v1/admin/workouts/:id

Workout exercises:

POST /api/v1/admin/workouts/:id/exercises
PUT /api/v1/admin/workouts/:id/exercises/:exercise_id
DELETE /api/v1/admin/workouts/:id/exercises/:exercise_id

---

# 23. Database

Основные таблицы:

users
profiles
lesson_categories
lessons
user_lesson_progress
exercises
programs
workouts
workout_exercises
workout_sessions
exercise_sets
user_progress
user_exercise_stats
achievements
user_achievements
admin_users

Использовать:
- PostgreSQL;
- UUID или BIGINT — выбрать единый подход;
- timestamps;
- foreign keys;
- indexes;
- unique constraints;
- check constraints там, где полезно.

SQL schema должна быть нормализованной и пригодной для аналитики.

---

# 24. Redis

Использовать Redis только там, где он действительно нужен.

Первоначально:
- cache для часто читаемого контента;
- rate limiting;
- при необходимости хранение короткоживущих данных.

Не использовать Redis как primary database.

---

# 25. Security

Обязательно:

- Telegram initData validation;
- JWT authentication;
- authorization;
- admin role;
- request validation;
- SQL injection protection через SQLC/parameterized queries;
- rate limiting;
- CORS;
- secure headers;
- secrets только через environment;
- не логировать JWT;
- не логировать Telegram initData целиком;
- обработка ошибок без утечки внутренних деталей.

---

# 26. Docker

Development:

docker-compose.yml

Services:

- postgres;
- redis;
- backend;
- frontend.

Добавить:
- healthchecks;
- volumes;
- environment variables;
- dependency handling.

Backend должен ждать готовности PostgreSQL.

---

# 27. Environment

Создать:

.env.example

Минимум:

DATABASE_URL=
REDIS_URL=
JWT_SECRET=
TELEGRAM_BOT_TOKEN=
TELEGRAM_BOT_USERNAME=
TELEGRAM_WEBAPP_URL=
CORS_ORIGINS=
APP_ENV=
PORT=
LOG_LEVEL=

Никогда не коммитить реальные secrets.

---

# 28. Seed data

Создать seed/migration с демо-контентом.

Минимум:

10 lessons
10 exercises
2 programs
5 workouts
10 achievements

Один program:
"Старт с нуля"

Второй:
"База калистеники"

Данные должны позволять полностью пройти MVP flow без ручного заполнения БД.

---

# 29. Frontend UX

Telegram Mini App должен выглядеть как мобильное приложение.

Учитывать:
- Telegram theme;
- safe area;
- mobile viewport;
- dark/light theme;
- touch targets;
- loading states;
- empty states;
- error states;
- skeleton loaders.

Не делать desktop-first UI.

---

# 30. Frontend pages

Создать:

/                  Home
/lessons            Lessons
/lessons/:id        Lesson
/exercises          Exercises
/exercises/:id      Exercise
/workout/today      Today's workout
/workout/:id        Workout
/progress            Progress
/achievements        Achievements
/profile             Profile

Admin можно вынести:

/admin
/admin/lessons
/admin/exercises
/admin/programs
/admin/workouts

---

# 31. State management

Server state:
TanStack Query.

Local UI state:
React state.

Global client state:
Zustand только для действительно глобальных состояний, например auth/session/preferences.

Не хранить server data одновременно в нескольких state stores.

---

# 32. Error handling

Backend должен возвращать единый формат:

{
  "error": {
    "code": "WORKOUT_NOT_FOUND",
    "message": "Workout not found"
  }
}

Frontend должен иметь общий error handler.

Ошибки должны быть понятны пользователю.

---

# 33. Logging

Structured logging.

Каждый request должен иметь:
- request id;
- method;
- path;
- status;
- duration.

Не логировать:
- JWT;
- secrets;
- Telegram initData;
- sensitive personal data.

---

# 34. Testing

Backend:

- unit tests для services;
- tests Telegram auth validation;
- tests XP calculation;
- tests streak;
- tests achievements;
- repository integration tests;
- API tests для критических endpoints.

Frontend:
- component tests для критических компонентов;
- минимум smoke test основного user flow.

Главный E2E сценарий:

Telegram auth
→ Home
→ Start workout
→ Complete sets
→ Complete workout
→ XP
→ streak
→ progress.

---

# 35. Documentation

Создать:

README.md
docs/ARCHITECTURE.md
docs/API.md
docs/DATABASE.md
docs/DEVELOPMENT.md
docs/TELEGRAM.md

README должен содержать:
- описание;
- stack;
- prerequisites;
- запуск;
- environment;
- migration;
- seed;
- тесты;
- Swagger;
- Telegram setup.

---

# 36. Development phases

## PHASE 0 — Discovery & architecture

Перед изменением кода:

1. изучить существующий repository;
2. определить текущий stack;
3. проверить git status;
4. проверить существующие README;
5. определить entry points;
6. определить, что уже есть;
7. не удалять рабочий функционал без необходимости.

Создать:

docs/ARCHITECTURE.md

Зафиксировать решения.

---

## PHASE 1 — Foundation

Сделать:

- project structure;
- Docker Compose;
- PostgreSQL;
- Redis;
- Go backend;
- React frontend;
- configuration;
- logging;
- health endpoint;
- CORS;
- Swagger foundation.

Проверки:

docker compose up
backend health
frontend starts
database connection
redis connection

---

## PHASE 2 — Database

Создать migrations.

Реализовать schema.

Подключить SQLC.

Создать seed data.

Проверить:
- migrations up;
- migrations down;
- seed;
- queries.

---

## PHASE 3 — Telegram Authentication

Реализовать:

- Telegram initData validation;
- user creation;
- JWT;
- auth middleware;
- GET /me.

Написать tests.

---

## PHASE 4 — Frontend Foundation

Реализовать:

- Telegram WebApp initialization;
- auth flow;
- router;
- API client;
- Query provider;
- layout;
- bottom navigation;
- theme;
- loading/error states.

---

## PHASE 5 — Lessons

Реализовать backend + frontend:

- categories;
- list;
- details;
- progress;
- completion.

---

## PHASE 6 — Exercises

Реализовать:

- catalog;
- filters;
- details;
- media;
- exercise information.

---

## PHASE 7 — Training Engine

Реализовать:

- programs;
- workouts;
- workout sessions;
- sets;
- repetitions;
- duration;
- rest timer;
- completion;
- history.

Это критический этап.

---

## PHASE 8 — Progress & Gamification

Реализовать:

- XP;
- levels;
- streak;
- statistics;
- history;
- achievements.

---

## PHASE 9 — Admin

Реализовать минимальную CMS:

- lessons CRUD;
- exercises CRUD;
- programs CRUD;
- workouts CRUD;
- publish/unpublish.

---

## PHASE 10 — Hardening

Проверить:

- authentication;
- authorization;
- validation;
- rate limiting;
- CORS;
- error handling;
- database constraints;
- indexes;
- transactions;
- race conditions;
- duplicate workout completion;
- duplicate XP;
- streak calculation.

---

## PHASE 11 — Testing

Полностью прогнать:

- unit tests;
- integration tests;
- API tests;
- frontend tests;
- E2E/smoke flow.

Исправить найденные ошибки.

---

## PHASE 12 — Production

Подготовить:

- production Docker;
- HTTPS assumptions;
- Telegram Mini App URL;
- bot configuration;
- database migrations;
- backups documentation;
- healthcheck;
- logging;
- monitoring hooks;
- CI/CD.

---

# 37. Критические бизнес-правила

1. Нельзя получить XP дважды за одну и ту же завершённую тренировку.

2. Повторное нажатие "Complete workout" должно быть idempotent.

3. Lesson completion должен быть idempotent.

4. Streak должен изменяться только один раз за календарный день.

5. Все вычисления прогресса должны выполняться на backend.

6. Frontend не может сам назначать себе XP.

7. Frontend не может сам назначать себе уровень.

8. Frontend не может сообщить backend другой telegram_id вместо пользователя из проверенного initData.

9. Workout session принадлежит конкретному user.

10. Пользователь не может изменить чужую workout session.

11. Admin endpoints требуют admin authorization.

---

# 38. Transactions

Использовать PostgreSQL transaction там, где одна пользовательская операция меняет несколько сущностей.

Например завершение тренировки:

BEGIN

update workout_session
insert/update progress
calculate XP
update user XP
update streak
unlock achievements

COMMIT

При ошибке:
ROLLBACK.

---

# 39. UX после завершения тренировки

После завершения:

Workout completed!

25 min
4 exercises
12 sets

+100 XP

Level progress:
███████░░░

🔥 8 day streak

Achievements:
🏆 First Week

Кнопка:
"Продолжить"

---

# 40. Definition of Done

Проект считается MVP-ready, если:

1. docker compose up запускает инфраструктуру;
2. backend подключается к PostgreSQL;
3. backend подключается к Redis;
4. frontend запускается;
5. Telegram Mini App может отправить initData;
6. backend валидирует Telegram initData;
7. пользователь создаётся автоматически;
8. JWT работает;
9. Home загружается;
10. lessons работают;
11. exercises работают;
12. programs работают;
13. workout можно начать;
14. sets можно завершить;
15. workout можно завершить;
16. XP начисляется;
17. XP не начисляется повторно;
18. streak работает;
19. progress отображается;
20. achievements работают;
21. admin может управлять контентом;
22. seed наполняет приложение демо-данными;
23. тесты проходят;
24. README содержит инструкции;
25. нет критических TODO;
26. нет hardcoded secrets;
27. Swagger описывает API.

---

# 41. Правила работы Codex

1. Не переписывай весь проект без необходимости.

2. Перед изменениями изучай существующий код.

3. Делай небольшие логические изменения.

4. После каждого этапа запускай тесты.

5. После изменения backend запускай backend tests.

6. После изменения frontend запускай frontend checks.

7. После изменения DB запускай migrations и integration tests.

8. Если команда не работает — исправь причину, а не обходи проблему.

9. Не скрывай ошибки.

10. Не оставляй приложение в состоянии "почти работает".

11. Не добавляй зависимость без необходимости.

12. Не используй deprecated API, если есть актуальная альтернатива.

13. Не хранить бизнес-логику в HTTP handlers.

14. Не хранить бизнес-логику во frontend.

15. Backend является источником истины для:
   - XP;
   - level;
   - streak;
   - progress;
   - achievements;
   - workout completion.

16. Все публичные API должны иметь validation.

17. Все изменения БД должны идти через migrations.

18. Не редактировать SQLC generated files вручную, если это можно исправить через schema/query/config.

19. Не коммитить .env.

20. Обновлять документацию при изменении архитектуры.

---

# 42. Git workflow

Делать небольшие логические commits.

Примеры:

feat: initialize project
feat: add database schema
feat: add telegram authentication
feat: add lessons
feat: add exercises
feat: add workout engine
feat: add progress tracking
feat: add achievements
feat: add admin
test: add workout integration tests
docs: update architecture

Не делать один гигантский commit со всем проектом, если repository уже ведётся через git.

---

# 43. Приоритеты

Если приходится выбирать между:

красивым UI и корректной бизнес-логикой

выбирать бизнес-логику.

Если приходится выбирать между:

новой функцией и стабильностью существующих функций

выбирать стабильность.

Если приходится выбирать между:

сложной архитектурой и простой рабочей архитектурой

выбирать простую рабочую архитектуру.

---

# 44. Первый запуск Codex

Начни с PHASE 0.

НЕ начинай сразу писать весь MVP.

Сначала:

1. изучи repository;
2. покажи текущую структуру;
3. определи существующий stack;
4. проверь git status;
5. найди существующие приложения/services;
6. определи, что можно переиспользовать;
7. создай/обнови docs/ARCHITECTURE.md;
8. сформулируй конкретный implementation plan на основе фактического repository.

После этого переходи к PHASE 1.

---

# 45. Главная команда

Работай по этому документу как по Master Specification.

Не пытайся реализовать все фазы за один шаг.

Реализуй проект последовательно:

PHASE 0
→ PHASE 1
→ PHASE 2
→ PHASE 3
→ PHASE 4
→ PHASE 5
→ PHASE 6
→ PHASE 7
→ PHASE 8
→ PHASE 9
→ PHASE 10
→ PHASE 11
→ PHASE 12

После каждого этапа:
- запусти проверки;
- исправь ошибки;
- обнови документацию;
- сообщи, что именно сделано;
- сообщи, какие проверки прошли;
- только после этого переходи дальше.

Начинай с PHASE 0.
