<script>
  let coworkFolder = $state('')
  let workspaceRoot = $state('')
  let saving = $state(false)
  let saved = $state(false)
  let pushing = $state(false)
  let pushed = $state(false)
  let launching = $state(false)
  let error = $state('')

  async function loadSettings() {
    try {
      // Placeholder: will call GetState binding when available
      const state = { coworkFolder: '', workspaceRoot: '' }
      coworkFolder = state.coworkFolder || ''
      workspaceRoot = state.workspaceRoot || ''
    } catch (e) {
      error = 'Failed to load settings'
    }
  }

  async function saveSettings() {
    saving = true
    saved = false
    error = ''
    try {
      // Placeholder: will call SetState binding when available
      await new Promise(r => setTimeout(r, 300))
      saved = true
      setTimeout(() => saved = false, 2000)
    } catch (e) {
      error = 'Failed to save settings'
    } finally {
      saving = false
    }
  }

  async function pushToCowork() {
    if (!coworkFolder) {
      error = 'Set a Cowork folder path first'
      return
    }
    pushing = true
    pushed = false
    error = ''
    try {
      // Placeholder: will call SaveProfile with cowork folder when binding available
      await new Promise(r => setTimeout(r, 500))
      pushed = true
      setTimeout(() => pushed = false, 2000)
    } catch (e) {
      error = 'Failed to push profile'
    } finally {
      pushing = false
    }
  }

  async function launchCowork() {
    launching = true
    error = ''
    try {
      // Placeholder: will call LaunchCowork binding when available
      await new Promise(r => setTimeout(r, 300))
    } catch (e) {
      error = 'Failed to launch Claude Desktop'
    } finally {
      launching = false
    }
  }

  $effect(() => { loadSettings() })
</script>

<div class="h-full flex flex-col gap-6">
  <div class="flex items-center justify-between">
    <h2 class="text-xl font-bold text-ck-pink">Settings</h2>
    {#if saved}<span class="text-xs text-ck-green">Settings saved!</span>{/if}
    {#if error}<span class="text-xs text-red-400">{error}</span>{/if}
  </div>

  <div class="space-y-6 max-w-xl">
    <!-- Cowork Integration -->
    <section class="space-y-3">
      <h3 class="text-sm font-semibold text-ck-gold uppercase tracking-wide">Cowork Integration</h3>

      <label class="block">
        <span class="text-xs text-ck-dim">Cowork Folder Path</span>
        <input type="text" bind:value={coworkFolder} placeholder="~/.claude/cowork-profiles"
          class="w-full mt-1 px-3 py-2 bg-ck-dark border border-gray-700 rounded-md text-sm text-white focus:outline-none focus:ring-2 focus:ring-ck-pink" />
      </label>

      <label class="block">
        <span class="text-xs text-ck-dim">Workspace Root</span>
        <input type="text" bind:value={workspaceRoot} placeholder="~/workspace"
          class="w-full mt-1 px-3 py-2 bg-ck-dark border border-gray-700 rounded-md text-sm text-white focus:outline-none focus:ring-2 focus:ring-ck-pink" />
      </label>

      <div class="flex gap-3">
        <button onclick={pushToCowork} disabled={pushing}
          class="flex-1 py-2 bg-ck-dark border border-ck-rose text-ck-rose text-sm font-semibold rounded-md hover:bg-ck-rose hover:text-white transition-colors disabled:opacity-50">
          {#if pushed}Pushed!{:else if pushing}Pushing...{:else}Push Profile to Cowork{/if}
        </button>
        <button onclick={launchCowork} disabled={launching}
          class="flex-1 py-2 bg-ck-dark border border-ck-gold text-ck-gold text-sm font-semibold rounded-md hover:bg-ck-gold hover:text-black transition-colors disabled:opacity-50">
          {launching ? 'Opening...' : 'Open Claude Desktop'}
        </button>
      </div>
    </section>

    <!-- Save -->
    <button onclick={saveSettings} disabled={saving}
      class="w-full py-2 bg-ck-rose text-white text-sm font-semibold rounded-md hover:bg-ck-pink transition-colors disabled:opacity-50">
      {saving ? 'Saving...' : 'Save Settings'}
    </button>
  </div>
</div>
