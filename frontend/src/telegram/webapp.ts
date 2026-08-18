export type TelegramWebApp = { initData: string; ready: () => void; expand: () => void; setHeaderColor?: (color: string) => void; setBackgroundColor?: (color: string) => void; themeParams?: Record<string, string | undefined> }
declare global { interface Window { Telegram?: { WebApp?: TelegramWebApp } } }
export function initializeTelegram() {
  const webApp = window.Telegram?.WebApp
  const root = document.documentElement
  const theme = webApp?.themeParams
  if (theme?.bg_color) root.style.setProperty('--tg-bg', theme.bg_color)
  if (theme?.text_color) root.style.setProperty('--tg-text', theme.text_color)
  if (theme?.hint_color) root.style.setProperty('--tg-muted', theme.hint_color)
  if (theme?.button_color) root.style.setProperty('--tg-accent', theme.button_color)
  webApp?.ready(); webApp?.expand(); webApp?.setHeaderColor?.(theme?.bg_color ?? '#0b1220'); webApp?.setBackgroundColor?.(theme?.bg_color ?? '#0b1220')
  return { initData: webApp?.initData ?? '' }
}
