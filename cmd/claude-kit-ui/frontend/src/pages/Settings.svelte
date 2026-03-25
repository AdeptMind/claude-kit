<script>
  import { currentProject } from '../stores/project.js'

  let coworkFolder = $state('')
  let workspaceRoot = $state('')
  let saving = $state(false)
  let saved = $state(false)
  let pushing = $state(false)
  let pushed = $state(false)
  let launching = $state(false)
  let error = $state('')

  // MCP Sync config
  let mcpName = $state('')
  let mcpCommand = $state('')
  let mcpArgs = $state('')
  let mcpEnv = $state('')
  let mcpSaving = $state(false)
  let mcpSaved = $state(false)
  let mcpTools = $state([])
  let mcpTesting = $state(false)

  let projectPath = $state('')
  let mcpEndpoint = $state('')
  currentProject.subscribe(p => {
    projectPath = p?.path || ''
    loadMCPConfig()
    loadMCPEndpoint()
  })

  let registering = $state(false)
  let registered = $state(false)

  async function loadMCPEndpoint() {
    try {
      const svc = await import('../../wailsjs/go/main/MCPServerService.js')
      mcpEndpoint = await svc.GetEndpoint()
    } catch {}
  }

  async function registerInClaudeDesktop() {
    registering = true
    error = ''
    try {
      const svc = await import('../../wailsjs/go/main/MCPServerService.js')
      await svc.RegisterInClaudeDesktop()
      registered = true
      setTimeout(() => registered = false, 3000)
    } catch (e) {
      error = 'Failed to register: ' + (e?.message || e)
    } finally {
      registering = false
    }
  }

  async function loadMCPConfig() {
    if (!projectPath) return
    try {
      const svc = await import('../../wailsjs/go/main/MCPService.js')
      const cfg = await svc.GetConfig(projectPath)
      if (cfg) {
        mcpName = cfg.name || ''
        mcpCommand = cfg.command || ''
        mcpArgs = (cfg.args || []).join(' ')
        mcpEnv = Object.entries(cfg.env || {}).map(([k, v]) => `${k}=${v}`).join('\n')
      }
    } catch {}
  }

  async function saveMCPConfig() {
    mcpSaving = true
    mcpSaved = false
    error = ''
    try {
      const svc = await import('../../wailsjs/go/main/MCPService.js')
      const envMap = {}
      mcpEnv.split('\n').filter(Boolean).forEach(line => {
        const idx = line.indexOf('=')
        if (idx > 0) envMap[line.slice(0, idx)] = line.slice(idx + 1)
      })
      await svc.SaveConfig(projectPath, {
        name: mcpName,
        command: mcpCommand,
        args: mcpArgs.split(/\s+/).filter(Boolean),
        env: envMap,
      })
      mcpSaved = true
      setTimeout(() => mcpSaved = false, 2000)
    } catch (e) {
      error = 'Failed to save MCP config'
    } finally {
      mcpSaving = false
    }
  }

  async function testMCPConnection() {
    mcpTesting = true
    mcpTools = []
    error = ''
    try {
      const svc = await import('../../wailsjs/go/main/MCPService.js')
      mcpTools = await svc.ListTools(projectPath) || []
    } catch (e) {
      error = e.message || 'MCP connection failed'
    } finally {
      mcpTesting = false
    }
  }

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

    <!-- MCP Skills Server -->
    {#if mcpEndpoint}
      <section class="space-y-3">
        <h3 class="text-sm font-semibold text-ck-gold uppercase tracking-wide">MCP Skills Server</h3>
        <p class="text-xs text-ck-dim">Expose skills to Claude Desktop via MCP.</p>
        <pre class="px-3 py-2 bg-ck-dark border border-gray-700 rounded-md text-xs text-green-400 font-mono overflow-x-auto">{`"claude-kit": { "url": "${mcpEndpoint}" }`}</pre>
        <button onclick={registerInClaudeDesktop} disabled={registering}
          class="w-full py-2 bg-ck-dark border border-ck-rose text-ck-rose text-sm font-semibold rounded-md hover:bg-ck-rose hover:text-white transition-colors disabled:opacity-50">
          {#if registered}Registered!{:else if registering}Registering...{:else}Register in Claude Desktop{/if}
        </button>
      </section>
    {/if}

    <!-- MCP Story Sync -->
    <section class="space-y-3">
      <h3 class="text-sm font-semibold text-ck-gold uppercase tracking-wide">MCP Story Sync</h3>
      <p class="text-xs text-ck-dim">Connect an MCP server to sync stories with Jira, Linear, etc.</p>

      <label class="block">
        <span class="text-xs text-ck-dim">Server Name</span>
        <input type="text" bind:value={mcpName} placeholder="Jira"
          class="w-full mt-1 px-3 py-2 bg-ck-dark border border-gray-700 rounded-md text-sm text-white focus:outline-none focus:ring-2 focus:ring-ck-pink" />
      </label>

      <label class="block">
        <span class="text-xs text-ck-dim">Command</span>
        <input type="text" bind:value={mcpCommand} placeholder="npx"
          class="w-full mt-1 px-3 py-2 bg-ck-dark border border-gray-700 rounded-md text-sm text-white focus:outline-none focus:ring-2 focus:ring-ck-pink" />
      </label>

      <label class="block">
        <span class="text-xs text-ck-dim">Arguments</span>
        <input type="text" bind:value={mcpArgs} placeholder="-y @anthropic/mcp-server-jira"
          class="w-full mt-1 px-3 py-2 bg-ck-dark border border-gray-700 rounded-md text-sm text-white focus:outline-none focus:ring-2 focus:ring-ck-pink" />
      </label>

      <label class="block">
        <span class="text-xs text-ck-dim">Environment Variables (one per line: KEY=VALUE)</span>
        <textarea bind:value={mcpEnv} rows="3" placeholder="JIRA_URL=https://your-org.atlassian.net&#10;JIRA_TOKEN=your-token"
          class="w-full mt-1 px-3 py-2 bg-ck-dark border border-gray-700 rounded-md text-sm text-white focus:outline-none focus:ring-2 focus:ring-ck-pink font-mono"></textarea>
      </label>

      <div class="flex gap-3">
        <button onclick={saveMCPConfig} disabled={mcpSaving}
          class="flex-1 py-2 bg-ck-dark border border-ck-gold text-ck-gold text-sm font-semibold rounded-md hover:bg-ck-gold hover:text-black transition-colors disabled:opacity-50">
          {#if mcpSaved}Saved!{:else if mcpSaving}Saving...{:else}Save MCP Config{/if}
        </button>
        <button onclick={testMCPConnection} disabled={mcpTesting}
          class="flex-1 py-2 bg-ck-dark border border-gray-600 text-ck-dim text-sm font-semibold rounded-md hover:border-white hover:text-white transition-colors disabled:opacity-50">
          {mcpTesting ? 'Testing...' : 'Test Connection'}
        </button>
      </div>

      {#if mcpTools.length > 0}
        <div class="rounded-lg bg-green-500/10 border border-green-500/30 p-3 text-xs text-green-400">
          Connected! Available tools: {mcpTools.join(', ')}
        </div>
      {/if}
    </section>

    <!-- Save -->
    <button onclick={saveSettings} disabled={saving}
      class="w-full py-2 bg-ck-rose text-white text-sm font-semibold rounded-md hover:bg-ck-pink transition-colors disabled:opacity-50">
      {saving ? 'Saving...' : 'Save Settings'}
    </button>
  </div>
</div>
