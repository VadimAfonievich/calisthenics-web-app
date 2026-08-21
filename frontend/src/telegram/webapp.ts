export type TelegramWebApp = { initData: string; ready: () => void; expand: () => void; setHeaderColor?: (color: string) => void; setBackgroundColor?: (color: string) => void; themeParams?: Record<string, string | undefined>; HapticFeedback?: { notificationOccurred?: (type: 'success'|'warning'|'error') => void; impactOccurred?: (style:'light'|'medium'|'heavy')=>void }; BackButton?: {show:()=>void;hide?:()=>void;onClick:(callback:()=>void)=>void;offClick:(callback:()=>void)=>void} }
declare global { interface Window { Telegram?: { WebApp?: TelegramWebApp } } }
export type TelegramDiagnostics = { sdkLoaded: boolean; webAppDetected: boolean; initDataPresent: boolean; initDataLength: number; initializationError?: string }
const message = (error: unknown) => error instanceof Error ? error.message.slice(0, 160) : 'Unknown initialization error'
export function initializeTelegram() {
  const webApp = window.Telegram?.WebApp
  const root = document.documentElement
  const theme = webApp?.themeParams
  let initializationError: string | undefined
  const safely = (name: string, action: () => void) => { try { action() } catch (error) { initializationError ??= `${name}: ${message(error)}` } }
  safely('theme', () => {
    if (theme?.bg_color) root.style.setProperty('--tg-bg', theme.bg_color)
    if (theme?.text_color) root.style.setProperty('--tg-text', theme.text_color)
    if (theme?.hint_color) root.style.setProperty('--tg-muted', theme.hint_color)
    if (theme?.button_color) root.style.setProperty('--tg-accent', theme.button_color)
  })
  safely('ready', () => webApp?.ready())
  safely('expand', () => webApp?.expand())
  safely('header color', () => webApp?.setHeaderColor?.(theme?.bg_color ?? '#0b1220'))
  safely('background color', () => webApp?.setBackgroundColor?.(theme?.bg_color ?? '#0b1220'))
  let initData = ''
  safely('initData', () => { initData = webApp?.initData ?? '' })
  return { initData, diagnostics: { sdkLoaded: Boolean(window.Telegram), webAppDetected: Boolean(webApp), initDataPresent: initData.length > 0, initDataLength: initData.length, initializationError } satisfies TelegramDiagnostics }
}
