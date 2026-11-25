<script lang="ts">
  import { onMount } from 'svelte'
  import { Label } from '$lib/components/ui/label'
  import { Separator } from '$lib/components/ui/separator'
  import { Button } from '$lib/components/ui/button'
  import ExternalLinkIcon from '@lucide/svelte/icons/external-link'
  import { Browser } from '@wailsio/runtime'
  import { commands } from '$lib/wails-bindings'
  import { confirmQuitIfDirty, flushAllStores } from '$lib/lifecycle'
  import { toast } from '$lib/stores/toast'
  import { t } from '$lib/i18n/t.svelte'

  let appName = $state('…')
  let appVersion = $state('…')
  let wailsVersion = $state('…')
  let updaterConfigured = $state(false)
  let platformInfo = $state('…')
  let updateStatus = $state<'idle' | 'checking' | 'downloading' | 'ready'>(
    'idle',
  )

  onMount(async () => {
    const [info, configured] = await Promise.all([
      commands.getSystemInfo(),
      commands.updateConfigured(),
    ])
    appName = info.appName
    appVersion = info.appVersion
    wailsVersion = info.wailsVersion
    platformInfo = `${info.osName} ${info.osVersion} (${info.osArch})`
    updaterConfigured = configured
  })

  function openLink(urlKey: string): void {
    void Browser.OpenURL(t(urlKey))
  }

  async function checkForUpdates(): Promise<void> {
    updateStatus = 'checking'
    try {
      const update = await commands.checkForUpdate()
      if (!update) {
        toast.success(t('preferences.about.upToDate'))
        updateStatus = 'idle'
        return
      }
      updateStatus = 'downloading'
      await commands.installUpdate()
      updateStatus = 'ready'
      toast.success(t('preferences.about.updateReady'))
      if (await confirmQuitIfDirty()) {
        await flushAllStores()
        await commands.restartForUpdate()
      } else {
        updateStatus = 'idle'
      }
    } catch {
      toast.error(t('preferences.about.updateError'))
      updateStatus = 'idle'
    }
  }
</script>

<div class="flex items-start justify-between gap-4">
  <div class="space-y-1">
    <p class="text-lg font-semibold">{appName}</p>
    <p class="text-muted-foreground text-sm">
      {t('preferences.about.versionLabel', { version: appVersion })}
    </p>
  </div>
  <Button
    variant="default"
    size="sm"
    disabled={!updaterConfigured || updateStatus !== 'idle'}
    onclick={checkForUpdates}
  >
    {#if updateStatus === 'checking'}
      {t('preferences.about.updateChecking')}
    {:else if updateStatus === 'downloading'}
      {t('preferences.about.updateDownloading')}
    {:else if updateStatus === 'ready'}
      {t('preferences.about.updateReady')}
    {:else}
      {t('preferences.about.checkForUpdates')}
    {/if}
  </Button>
</div>

<Separator />

<div class="space-y-3">
  <Label>{t('preferences.about.systemSection')}</Label>
  <div class="divide-y rounded-md border">
    <div class="flex items-center justify-between gap-4 px-3 py-2 text-sm">
      <span class="text-muted-foreground">
        {t('preferences.about.wailsVersionLabel')}
      </span>
      <span class="font-mono text-xs">{wailsVersion}</span>
    </div>
    <div class="flex items-center justify-between gap-4 px-3 py-2 text-sm">
      <span class="text-muted-foreground">
        {t('preferences.about.platformLabel')}
      </span>
      <span class="font-mono text-xs">{platformInfo}</span>
    </div>
  </div>
</div>

<Separator />

<div class="space-y-3">
  <Label>{t('preferences.about.linksSection')}</Label>
  <div class="flex flex-wrap gap-2">
    <Button
      variant="outline"
      size="sm"
      onclick={() => openLink('preferences.about.githubUrl')}
    >
      <ExternalLinkIcon class="size-3.5" />
      {t('preferences.about.githubLink')}
    </Button>
    <Button
      variant="outline"
      size="sm"
      onclick={() => openLink('preferences.about.websiteUrl')}
    >
      <ExternalLinkIcon class="size-3.5" />
      {t('preferences.about.websiteLink')}
    </Button>
    <Button
      variant="outline"
      size="sm"
      onclick={() => openLink('preferences.about.changelogUrl')}
    >
      <ExternalLinkIcon class="size-3.5" />
      {t('preferences.about.changelogLink')}
    </Button>
  </div>
</div>

<Separator />

<div class="space-y-2">
  <Label>{t('preferences.about.licenseSection')}</Label>
  <p class="text-muted-foreground text-sm">
    {t('preferences.about.licenseText')}
  </p>
</div>
