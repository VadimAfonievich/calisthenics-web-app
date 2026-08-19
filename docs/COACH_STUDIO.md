# Coach Content Studio

Coach Studio is available at `/coach` and its API under `/api/v1/coach`. Access is decided by `admin_users.role`: `coach` manages rows whose `owner_user_id` is their user ID, while `admin` and `super_admin` can manage all content. A normal user receives `403`. Existing seed rows have a null owner and remain system content.

## Authorization role and application mode

The PostgreSQL `admin_users.role` is the security authority. `GET /api/v1/me` returns that role plus server-derived `available_modes`. The frontend application mode (`student` or `coach`) only selects navigation and presentation and is stored in `localStorage` as `calisthenics_app_mode`. It never grants API permissions. Every `/api/v1/coach/*` request still checks the database-backed role and returns `403` for a normal user, even if local storage is manually changed.

Authorized coaches and administrators can switch modes from `/profile`. A saved coach mode is restored only when the current `/me` response still permits it; otherwise it falls back to student mode.

Content uses `draft`, `published`, and `archived`. Legacy `published` flags on lessons and programs are synchronized for compatibility. Student APIs expose published content only. Publishing records `published_at` and `published_by`; editing does not copy analytics or learner progress.

Lessons retain legacy `content` and add ordered JSONB blocks. Allowed blocks are `heading`, `text`, `image`, `video`, `tip`, `warning`, `checklist`, and `divider`. The backend validates block shape. The mobile editor supports adding, removing, and reordering blocks. Other content lists expose search, status filtering, lifecycle actions, and provide the foundation for specialized workout/program/skill builders.

Analytics is aggregate-only: user activity, lesson/workout completions, popular workouts, skill progress, and achievements. There is no coach-to-student tenancy relation yet, so coach analytics currently describes the global learner population; individual CRM and payments are intentionally excluded.
