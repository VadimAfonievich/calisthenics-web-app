# Telegram Mini App setup

Create a bot through BotFather and configure its Menu Button or Web App URL to the HTTPS value of `TELEGRAM_WEBAPP_URL`. Set that same URL as the production Mini App URL and use its origin in `CORS_ORIGINS`.

Set `TELEGRAM_BOT_TOKEN` only in the protected production environment file. Never put it in frontend build variables, commits, logs, or screenshots. Open the Mini App from Telegram and verify that the backend accepts the signed `initData`; direct browser mode intentionally remains a demo mode.
