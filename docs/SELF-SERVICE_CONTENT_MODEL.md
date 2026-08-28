# Self-service content model

Coaches create and publish exercises, workouts, programs, and skills once. Students discover published content and start programs themselves; manual per-student assignment is not required.

`user_program_progress` is user-scoped and supports multiple simultaneous active programs. Program levels advance from completed workout sessions, and locked program workouts are protected by the backend.

Published-content authorization goes through `internal/access.Service`. Phase 19 permits authenticated users. A future subscription implementation must replace this policy at this single boundary rather than adding billing checks to handlers.

Skills remain `category → skill → progression levels`. A skill level linked to a program level can use actual workout completion.

Super administrators can promote users to coach or demote coaches in the Mini App. Admin and super-admin escalation is intentionally excluded; changes are recorded in `role_change_audit`.
