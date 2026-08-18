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
components:
  securitySchemes:
    bearerAuth: {type: http, scheme: bearer, bearerFormat: JWT}
`
