import * as desktop from '../../bindings/github.com/pavlith/template/desktop'
import type * as models from '../../bindings/github.com/pavlith/template/models'

export type {
  CrashReportSummary,
  DiagnosticsReport,
  ImportedTheme,
  ThemeProfile,
} from '../../bindings/github.com/pavlith/template/models'

export type AppPreferences = Omit<
  models.AppPreferences,
  'importedThemes' | 'commandShortcuts'
> & {
  importedThemes: NonNullable<models.AppPreferences['importedThemes']>
  commandShortcuts: NonNullable<models.AppPreferences['commandShortcuts']>
}
export type PersistedAppState = Omit<
  models.PersistedAppState,
  'recentItems'
> & {
  recentItems: string[]
}
export type ShortcutPurpose = Exclude<`${models.ShortcutPurpose}`, ''>

type Result<T> = { status: 'ok'; data: T } | { status: 'error'; error: unknown }

function withResult<A extends unknown[], T>(call: (...args: A) => Promise<T>) {
  return async (...args: A): Promise<Result<T>> => {
    try {
      return { status: 'ok', data: await call(...args) }
    } catch (error) {
      return { status: 'error', error }
    }
  }
}

export function unwrapResult<T>(result: Result<T>): T {
  if (result.status === 'ok') return result.data
  throw result.error
}

export const commands = {
  loadPreferences: withResult(desktop.LoadPreferences),
  savePreferences: withResult(desktop.SavePreferences),
  loadAppState: withResult(desktop.LoadAppState),
  saveAppState: withResult(desktop.SaveAppState),
  openPreferencesFile: withResult(desktop.OpenPreferencesFile),
  hasRecentCrash: withResult(desktop.HasRecentCrash),
  getCrashReport: withResult(desktop.GetCrashReport),
  listCrashReports: withResult(
    async () => (await desktop.ListCrashReports()) ?? [],
  ),
  clearCrashReports: withResult(desktop.ClearCrashReports),
  collectDiagnostics: withResult(desktop.CollectDiagnostics),
  listSystemFonts: withResult(
    async () => (await desktop.ListSystemFonts()) ?? [],
  ),
  greet: withResult(desktop.Greet),
  registerGlobalShortcut: withResult(
    (purpose: ShortcutPurpose, accelerator: string) =>
      desktop.RegisterGlobalShortcut(
        purpose as models.ShortcutPurpose,
        accelerator,
      ),
  ),
  unregisterGlobalShortcut: withResult((purpose: ShortcutPurpose) =>
    desktop.UnregisterGlobalShortcut(purpose as models.ShortcutPurpose),
  ),
  toggleQuickPane: withResult(desktop.ToggleQuickPane),
  dismissQuickPane: withResult(desktop.DismissQuickPane),
  cleanupOldRecoveryFiles: withResult(desktop.CleanupOldRecoveryFiles),
  saveEmergencyData: withResult(desktop.SaveEmergencyData),
  logFrontendError: withResult(desktop.LogFrontendError),
  logFrontend: desktop.LogFrontend,
  quitApp: desktop.QuitApp,
  mainReady: desktop.MainReady,
  enableAutostart: desktop.EnableAutostart,
  disableAutostart: desktop.DisableAutostart,
  isAutostartEnabled: desktop.IsAutostartEnabled,
  notify: desktop.Notify,
  relaunch: desktop.Relaunch,
  getSystemInfo: desktop.GetSystemInfo,
  runShellDemo: desktop.RunShellDemo,
  pickThemeFile: desktop.PickThemeFile,
  updateConfigured: desktop.UpdateConfigured,
  checkForUpdate: desktop.CheckForUpdate,
  installUpdate: desktop.InstallUpdate,
  restartForUpdate: desktop.RestartForUpdate,
  setAppMenu: desktop.SetAppMenu,
  showContextMenu: desktop.ShowContextMenu,
}
