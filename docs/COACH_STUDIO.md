# Coach Content Studio

Coach Studio is available at `/coach` and its API under `/api/v1/coach`. Access is decided by `admin_users.role`: `coach` manages rows whose `owner_user_id` is their user ID, while `admin` and `super_admin` can manage all content. A normal user receives `403`. Existing seed rows have a null owner and remain system content.

Content uses `draft`, `published`, and `archived`. Legacy `published` flags on lessons and programs are synchronized for compatibility. Student APIs expose published content only. Publishing records `published_at` and `published_by`; editing does not copy analytics or learner progress.

Lessons retain legacy `content` and add ordered JSONB blocks. Allowed blocks are `heading`, `text`, `image`, `video`, `tip`, `warning`, `checklist`, and `divider`. The backend validates block shape. The mobile editor supports adding, removing, and reordering blocks. Other content lists expose search, status filtering, lifecycle actions, and provide the foundation for specialized workout/program/skill builders.

Analytics is aggregate-only: user activity, lesson/workout completions, popular workouts, skill progress, and achievements. There is no coach-to-student tenancy relation yet, so coach analytics currently describes the global learner population; individual CRM and payments are intentionally excluded.
