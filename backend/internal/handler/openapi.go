package handler

const openAPI = `openapi: 3.0.3
info:
  title: Calisthenics Coach API
  version: 0.1.0
  description: API contract for the Calisthenics Coach Mini App.
paths:
  /healthz:
    get:
      summary: Check API and infrastructure readiness
      responses:
        '200':
          description: API, PostgreSQL, and Redis are ready
        '503':
          description: A required dependency is unavailable
  /api/v1/auth/telegram:
    post:
      summary: Validate Telegram Mini App initData and issue an access token
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [init_data]
              properties:
                init_data: {type: string}
      responses:
        '200': {description: Authenticated user and Bearer access token}
        '401': {description: Invalid or expired Telegram initData}
  /api/v1/me:
    get:
      summary: Get the current user profile
      security: [{bearerAuth: []}]
      responses:
        '200': {description: Current user}
        '401': {description: Missing or invalid access token}
  /api/v1/lessons:
    get:
      summary: List published lessons with the current user's progress
      security: [{bearerAuth: []}]
      responses:
        '200': {description: Published lessons}
  /api/v1/lessons/{id}:
    get:
      summary: Get a published lesson with current-user progress
      security: [{bearerAuth: []}]
      responses:
        '200': {description: Lesson}
        '404': {description: Lesson not found}
  /api/v1/lessons/{id}/complete:
    post:
      summary: Complete a lesson and award XP once
      security: [{bearerAuth: []}]
      responses:
        '200': {description: Completion result}
        '404': {description: Lesson not found}
  /api/v1/programs:
    get:
      summary: List published programs
      security: [{bearerAuth: []}]
      responses:
        '200': {description: Published programs}
        '401': {description: Missing or invalid access token}
  /api/v1/programs/{id}:
    get:
      summary: Get a published program
      security: [{bearerAuth: []}]
      parameters:
        - {name: id, in: path, required: true, schema: {type: string, format: uuid}}
      responses:
        '200': {description: Published program}
        '401': {description: Missing or invalid access token}
        '404': {description: Program not found}
  /api/v1/workouts:
    get:
      summary: List workouts from published programs
      security: [{bearerAuth: []}]
      responses:
        '200': {description: Workout catalog with program metadata and latest user status}
        '401': {description: Missing or invalid access token}
  /api/v1/workouts/{id}:
    get:
      summary: Get a published workout plan without starting a session
      security: [{bearerAuth: []}]
      parameters:
        - {name: id, in: path, required: true, schema: {type: string, format: uuid}}
      responses:
        '200': {description: Workout plan with ordered exercises}
        '400': {description: Invalid workout UUID}
        '404': {description: Workout not found}
  /api/v1/workouts/{id}/start:
    post:
      summary: Create or resume an active workout session
      security: [{bearerAuth: []}]
      parameters:
        - {name: id, in: path, required: true, schema: {type: string, format: uuid}}
      responses:
        '201': {description: Active workout session}
        '400': {description: Invalid workout UUID}
  /api/v1/workout-sessions/{id}:
    get:
      summary: Resume an owned workout session with its plan and recorded sets
      security: [{bearerAuth: []}]
      parameters:
        - {name: id, in: path, required: true, schema: {type: string, format: uuid}}
      responses:
        '200': {description: Session, workout plan, and completed sets}
        '400': {description: Invalid session UUID}
        '403': {description: Session belongs to another user}
  /api/v1/workout-sessions/{id}/sets:
    post:
      summary: Idempotently record a completed exercise set
      security: [{bearerAuth: []}]
      responses:
        '204': {description: Set recorded}
        '400': {description: Invalid set payload or UUID}
        '403': {description: Session belongs to another user or is closed}
  /api/v1/workout-sessions/{id}/complete:
    post:
      summary: Complete a workout and evaluate XP, streak, and achievements once
      security: [{bearerAuth: []}]
      parameters:
        - {name: duration_seconds, in: query, required: true, schema: {type: integer, minimum: 0, maximum: 43200}}
      responses:
        '200': {description: Completion summary}
        '400': {description: Invalid duration or UUID}
        '403': {description: Session belongs to another user}
  /api/v1/progress:
    get:
      summary: Get XP, level, streak, and lifetime totals
      security: [{bearerAuth: []}]
      responses:
        '200': {description: Current progress summary}
  /api/v1/stats:
    get:
      summary: Get lifetime and weekly workout statistics
      security: [{bearerAuth: []}]
      responses:
        '200': {description: Workout statistics}
  /api/v1/history:
    get:
      summary: Get completed workout history
      security: [{bearerAuth: []}]
      responses:
        '200': {description: Completed workout history}
  /api/v1/achievements:
    get:
      summary: Get the achievement catalog and unlock state
      security: [{bearerAuth: []}]
      responses:
        '200': {description: Achievement catalog}
  /api/v1/admin/lessons:
    post:
      summary: Create a lesson (administrator only)
      security: [{bearerAuth: []}]
      requestBody: {required: true, content: {application/json: {schema: {$ref: '#/components/schemas/AdminLessonInput'}}}}
      responses: {'201': {description: Created lesson ID}, '400': {description: Invalid payload}, '403': {description: Administrator role required}}
  /api/v1/admin/lessons/{id}:
    put:
      summary: Update a lesson
      security: [{bearerAuth: []}]
      requestBody: {required: true, content: {application/json: {schema: {$ref: '#/components/schemas/AdminLessonInput'}}}}
      responses: {'200': {description: Updated lesson ID}, '400': {description: Invalid payload}, '404': {description: Lesson not found}}
    delete:
      summary: Delete a lesson
      security: [{bearerAuth: []}]
      responses: {'204': {description: Deleted}, '404': {description: Lesson not found}}
  /api/v1/admin/lessons/{id}/publish:
    post:
      summary: Publish or unpublish a lesson
      security: [{bearerAuth: []}]
      requestBody: {required: true, content: {application/json: {schema: {$ref: '#/components/schemas/PublishInput'}}}}
      responses: {'200': {description: Updated lesson ID}, '400': {description: Invalid payload}, '404': {description: Lesson not found}}
  /api/v1/admin/exercises:
    post:
      summary: Create an exercise (administrator only)
      security: [{bearerAuth: []}]
      requestBody: {required: true, content: {application/json: {schema: {$ref: '#/components/schemas/AdminExerciseInput'}}}}
      responses: {'201': {description: Created exercise ID}, '400': {description: Invalid payload}, '403': {description: Administrator role required}}
  /api/v1/admin/exercises/{id}:
    put:
      summary: Update an exercise
      security: [{bearerAuth: []}]
      requestBody: {required: true, content: {application/json: {schema: {$ref: '#/components/schemas/AdminExerciseInput'}}}}
      responses: {'200': {description: Updated exercise ID}, '400': {description: Invalid payload}, '404': {description: Exercise not found}}
    delete:
      summary: Delete an exercise
      security: [{bearerAuth: []}]
      responses: {'204': {description: Deleted}, '404': {description: Exercise not found}}
  /api/v1/admin/programs:
    post:
      summary: Create a program (administrator only)
      security: [{bearerAuth: []}]
      requestBody: {required: true, content: {application/json: {schema: {$ref: '#/components/schemas/AdminProgramInput'}}}}
      responses: {'201': {description: Created program ID}, '400': {description: Invalid payload}, '403': {description: Administrator role required}}
  /api/v1/admin/programs/{id}:
    put:
      summary: Update a program
      security: [{bearerAuth: []}]
      requestBody: {required: true, content: {application/json: {schema: {$ref: '#/components/schemas/AdminProgramInput'}}}}
      responses: {'200': {description: Updated program ID}, '400': {description: Invalid payload}, '404': {description: Program not found}}
    delete:
      summary: Delete a program
      security: [{bearerAuth: []}]
      responses: {'204': {description: Deleted}, '404': {description: Program not found}}
  /api/v1/admin/programs/{id}/publish:
    post:
      summary: Publish or unpublish a program
      security: [{bearerAuth: []}]
      requestBody: {required: true, content: {application/json: {schema: {$ref: '#/components/schemas/PublishInput'}}}}
      responses: {'200': {description: Updated program ID}, '400': {description: Invalid payload}, '404': {description: Program not found}}
  /api/v1/admin/workouts:
    post:
      summary: Create a workout (administrator only)
      security: [{bearerAuth: []}]
      requestBody: {required: true, content: {application/json: {schema: {$ref: '#/components/schemas/AdminWorkoutInput'}}}}
      responses: {'201': {description: Created workout ID}, '400': {description: Invalid payload}, '403': {description: Administrator role required}}
  /api/v1/admin/workouts/{id}:
    put:
      summary: Update a workout
      security: [{bearerAuth: []}]
      requestBody: {required: true, content: {application/json: {schema: {$ref: '#/components/schemas/AdminWorkoutInput'}}}}
      responses: {'200': {description: Updated workout ID}, '400': {description: Invalid payload}, '404': {description: Workout not found}}
    delete:
      summary: Delete a workout
      security: [{bearerAuth: []}]
      responses: {'204': {description: Deleted}, '404': {description: Workout not found}}
  /api/v1/admin/workouts/{id}/exercises:
    post:
      summary: Add an exercise to a workout
      security: [{bearerAuth: []}]
      requestBody: {required: true, content: {application/json: {schema: {$ref: '#/components/schemas/AdminWorkoutExerciseInput'}}}}
      responses: {'201': {description: Created workout exercise ID}, '400': {description: Invalid payload}}
components:
  securitySchemes:
    bearerAuth: {type: http, scheme: bearer, bearerFormat: JWT}
  schemas:
    PublishInput:
      type: object
      required: [published]
      properties:
        published: {type: boolean}
    AdminLessonInput:
      type: object
      required: [category_id, title, slug, short_description, content, difficulty, duration_minutes]
      properties:
        category_id: {type: string, format: uuid}
        title: {type: string}
        slug: {type: string}
        short_description: {type: string}
        content: {type: string}
        difficulty: {type: string, enum: [beginner, intermediate, advanced]}
        duration_minutes: {type: integer, minimum: 1, maximum: 1440}
        sort_order: {type: integer, minimum: 0}
        published: {type: boolean}
    AdminExerciseInput:
      type: object
      required: [name, slug, description, instructions, common_mistakes, difficulty, muscle_groups]
      properties:
        name: {type: string}
        slug: {type: string}
        description: {type: string}
        instructions: {type: string}
        common_mistakes: {type: string}
        difficulty: {type: string, enum: [beginner, intermediate, advanced]}
        muscle_groups: {type: array, minItems: 1, items: {type: string}}
        equipment: {type: array, items: {type: string}}
        video_url: {type: string, nullable: true}
        image_url: {type: string, nullable: true}
    AdminProgramInput:
      type: object
      required: [name, slug, description, difficulty, duration_weeks]
      properties:
        name: {type: string}
        slug: {type: string}
        description: {type: string}
        difficulty: {type: string, enum: [beginner, intermediate, advanced]}
        duration_weeks: {type: integer, minimum: 1, maximum: 520}
        published: {type: boolean}
    AdminWorkoutInput:
      type: object
      required: [program_id, day_number, title, description, estimated_minutes]
      properties:
        program_id: {type: string, format: uuid}
        day_number: {type: integer, minimum: 1}
        title: {type: string}
        description: {type: string}
        estimated_minutes: {type: integer, minimum: 1, maximum: 1440}
        sort_order: {type: integer, minimum: 0}
    AdminWorkoutExerciseInput:
      type: object
      required: [exercise_id, sets, rest_seconds]
      properties:
        exercise_id: {type: string, format: uuid}
        sets: {type: integer, minimum: 1, maximum: 100}
        target_reps: {type: integer, minimum: 1, nullable: true}
        target_duration_seconds: {type: integer, minimum: 1, nullable: true}
        rest_seconds: {type: integer, minimum: 0, maximum: 3600}
        sort_order: {type: integer, minimum: 0}
        notes: {type: string, nullable: true}
`
