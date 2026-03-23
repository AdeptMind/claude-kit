<script>
  import { currentProject } from '../stores/project.js'

  let { show = $bindable(false), onComplete = () => {} } = $props()

  let agents = $state([])
  let loading = $state(true)
  let installing = $state(false)
  let error = $state('')
  let policyProfile = $state('none')
  let selected = $state(new Set())
  let bmadChecked = $state(false)

  let managementAgents = $derived(agents.filter(a => a.category === 'management'))
  let devAgents = $derived(agents.filter(a => a.category === 'dev'))
  let specialAgents = $derived(agents.filter(a => a.category === 'special'))

  let devExpanded = $state(false)
  let specialExpanded = $state(false)

  const bmadAgents = ['product-owner', 'architect', 'tech-lead']

  let summary = $state({ agents: 0, skills: 0, commands: 0, rules: 0 })

  $effect(() => {
    if (show) loadAvailable()
  })

  $effect(() => {
    updateSummary()
  })

  async function loadAvailable() {
    loading = true
    error = ''
    try {
      const svc = await import('../../wailsjs/go/main/InitService.js')
      const projectPath = $currentProject?.path || ''
      const list = await svc.ListAvailable(projectPath)
      agents = list || []
    } catch (e) {
      error = 'Failed to load available agents'
      agents = []
    }
    loading = false
  }

  function toggle(name) {
    const next = new Set(selected)
    if (next.has(name)) next.delete(name)
    else next.add(name)
    selected = next
  }

  function toggleBmad() {
    bmadChecked = !bmadChecked
    const next = new Set(selected)
    if (bmadChecked) {
      bmadAgents.forEach(a => next.add(a))
    } else {
      bmadAgents.forEach(a => next.delete(a))
    }
    selected = next
  }

  async function updateSummary() {
    const names = [...selected]
    if (names.length === 0) {
      summary = { agents: 0, skills: 0, commands: 0, rules: 0 }
      return
    }
    try {
      const svc = await import('../../wailsjs/go/main/InitService.js')
      summary = await svc.PreviewInstall(names)
    } catch {
      // fallback
    }
  }

  async function install() {
    installing = true
    error = ''
    try {
      const svc = await import('../../wailsjs/go/main/InitService.js')
      const projectPath = $currentProject?.path || ''
      await svc.Install(projectPath, [...selected], policyProfile)
      show = false
      onComplete()
    } catch (e) {
      error = 'Installation failed: ' + (e?.message || e)
    }
    installing = false
  }

  function close() {
    if (!installing) show = false
  }
</script>

{#if show}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm" onclick={close} onkeydown={(e) => e.key === 'Escape' && close()}>
    <!-- svelte-ignore a11y_no_static_element_interactions a11y_click_events_have_key_events -->
    <div class="bg-[#0d1117] border border-gray-800 rounded-xl shadow-2xl w-[680px] max-h-[85vh] flex flex-col" onclick={(e) => e.stopPropagation()}>

      <!-- Header -->
      <div class="px-6 py-4 border-b border-gray-800 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-ck-pink">Setup Project</h2>
          <p class="text-xs text-ck-dim mt-0.5">Select agents to install</p>
        </div>
        <button onclick={close} class="text-ck-dim hover:text-white transition-colors text-lg">&times;</button>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto px-6 py-4 space-y-5">
        {#if loading}
          <div class="flex items-center justify-center h-32">
            <div class="w-5 h-5 border-2 border-ck-pink border-t-transparent rounded-full animate-spin"></div>
          </div>
        {:else}
          {#if error}
            <p class="text-xs text-red-400">{error}</p>
          {/if}

          <!-- BMAD bundle -->
          <label class="flex items-center gap-3 p-3 rounded-lg bg-ck-dark border border-gray-700 hover:border-ck-rose/50 transition-colors cursor-pointer">
            <input type="checkbox" checked={bmadChecked} onchange={toggleBmad}
              class="w-4 h-4 rounded border-gray-600 text-ck-rose focus:ring-ck-rose bg-ck-dark" />
            <div>
              <span class="text-sm font-semibold text-ck-gold">BMAD Methodology</span>
              <p class="text-xs text-ck-dim">Pre-selects: product-owner, architect, tech-lead</p>
            </div>
          </label>

          <!-- Management -->
          <div>
            <h3 class="text-xs font-semibold uppercase tracking-wider text-ck-dim mb-2">Management</h3>
            <div class="grid grid-cols-2 gap-1.5">
              {#each managementAgents as agent}
                <label class="flex items-center gap-2 px-3 py-2 rounded-md hover:bg-ck-dark/50 transition-colors cursor-pointer
                  {agent.installed ? 'opacity-50' : ''}">
                  <input type="checkbox" checked={selected.has(agent.name)} onchange={() => toggle(agent.name)}
                    disabled={agent.installed}
                    class="w-3.5 h-3.5 rounded border-gray-600 text-ck-rose focus:ring-ck-rose bg-ck-dark" />
                  <div class="min-w-0">
                    <span class="text-sm text-white">{agent.name}</span>
                    {#if agent.description}
                      <span class="text-xs text-ck-dim ml-1.5">{agent.description}</span>
                    {/if}
                    {#if agent.installed}
                      <span class="text-[10px] text-ck-green ml-1">installed</span>
                    {/if}
                  </div>
                </label>
              {/each}
            </div>
          </div>

          <!-- Developer -->
          <div>
            <button onclick={() => devExpanded = !devExpanded}
              class="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider text-ck-dim hover:text-white transition-colors mb-2">
              <span class="transition-transform {devExpanded ? 'rotate-90' : ''}">&#x203A;</span>
              Developer ({devAgents.length})
            </button>
            {#if devExpanded}
              <div class="grid grid-cols-2 gap-1.5">
                {#each devAgents as agent}
                  <label class="flex items-center gap-2 px-3 py-2 rounded-md hover:bg-ck-dark/50 transition-colors cursor-pointer
                    {agent.installed ? 'opacity-50' : ''}">
                    <input type="checkbox" checked={selected.has(agent.name)} onchange={() => toggle(agent.name)}
                      disabled={agent.installed}
                      class="w-3.5 h-3.5 rounded border-gray-600 text-ck-rose focus:ring-ck-rose bg-ck-dark" />
                    <div class="min-w-0">
                      <span class="text-sm text-white">{agent.name}</span>
                      {#if agent.description}
                        <span class="text-xs text-ck-dim ml-1.5">{agent.description}</span>
                      {/if}
                      {#if agent.installed}
                        <span class="text-[10px] text-ck-green ml-1">installed</span>
                      {/if}
                    </div>
                  </label>
                {/each}
              </div>
            {/if}
          </div>

          <!-- Special -->
          {#if specialAgents.length > 0}
            <div>
              <button onclick={() => specialExpanded = !specialExpanded}
                class="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider text-ck-dim hover:text-white transition-colors mb-2">
                <span class="transition-transform {specialExpanded ? 'rotate-90' : ''}">&#x203A;</span>
                Special ({specialAgents.length})
              </button>
              {#if specialExpanded}
                <div class="grid grid-cols-2 gap-1.5">
                  {#each specialAgents as agent}
                    <label class="flex items-center gap-2 px-3 py-2 rounded-md hover:bg-ck-dark/50 transition-colors cursor-pointer
                      {agent.installed ? 'opacity-50' : ''}">
                      <input type="checkbox" checked={selected.has(agent.name)} onchange={() => toggle(agent.name)}
                        disabled={agent.installed}
                        class="w-3.5 h-3.5 rounded border-gray-600 text-ck-rose focus:ring-ck-rose bg-ck-dark" />
                      <div class="min-w-0">
                        <span class="text-sm text-white">{agent.name}</span>
                        {#if agent.description}
                          <span class="text-xs text-ck-dim ml-1.5">{agent.description}</span>
                        {/if}
                        {#if agent.installed}
                          <span class="text-[10px] text-ck-green ml-1">installed</span>
                        {/if}
                      </div>
                    </label>
                  {/each}
                </div>
              {/if}
            </div>
          {/if}

          <!-- Policy Profile -->
          <div>
            <h3 class="text-xs font-semibold uppercase tracking-wider text-ck-dim mb-2">Security Profile</h3>
            <div class="flex gap-3">
              {#each [['none', 'None'], ['permissive', 'Permissive'], ['moderate', 'Moderate'], ['strict', 'Strict']] as [value, label]}
                <label class="flex items-center gap-1.5 cursor-pointer">
                  <input type="radio" name="policy" {value} bind:group={policyProfile}
                    class="w-3.5 h-3.5 border-gray-600 text-ck-rose focus:ring-ck-rose bg-ck-dark" />
                  <span class="text-sm text-white">{label}</span>
                </label>
              {/each}
            </div>
            <p class="text-[10px] text-ck-dim mt-1">
              {#if policyProfile === 'none'}No security policy applied
              {:else if policyProfile === 'permissive'}Minimal restrictions
              {:else if policyProfile === 'moderate'}Standard protection
              {:else}Full lockdown{/if}
            </p>
          </div>
        {/if}
      </div>

      <!-- Footer -->
      <div class="px-6 py-4 border-t border-gray-800 flex items-center justify-between">
        <p class="text-xs text-ck-dim">
          {#if summary.agents > 0}
            Will install: {summary.agents} agents, ~{summary.skills} skills, ~{summary.commands} commands, ~{summary.rules} rules
          {:else}
            Select agents to see summary
          {/if}
        </p>
        <div class="flex gap-2">
          <button onclick={close} disabled={installing}
            class="px-4 py-2 text-sm rounded-lg border border-gray-700 text-ck-dim hover:text-white hover:border-gray-500 transition-colors disabled:opacity-50">
            Cancel
          </button>
          <button onclick={install} disabled={installing || selected.size === 0}
            class="px-4 py-2 text-sm rounded-lg bg-ck-rose text-white font-medium hover:bg-ck-pink transition-colors disabled:opacity-50">
            {installing ? 'Installing...' : 'Install'}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
