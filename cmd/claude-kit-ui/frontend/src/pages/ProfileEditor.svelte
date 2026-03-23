<script>
  import MarkdownPreview from '../components/MarkdownPreview.svelte'
  import { ListAgents, LoadAgent, SaveProfile, GetSavedProfile } from '../../wailsjs/go/main/ProfileService'
  import { currentProject } from '../stores/project.js'

  let agents = $state([])
  let loading = $state(true)
  let selectedAgent = $state(null)
  let saving = $state(false)
  let saved = $state(false)
  let error = $state('')
  let devExpanded = $state(false)
  let specialExpanded = $state(false)
  let projectPath = $state('')

  currentProject.subscribe(p => {
    projectPath = p?.path || ''
    if (p?.path) loadAgents()
  })

  let agentName = $state('')
  let roleTitle = $state('')
  let description = $state('')
  let responsibilities = $state('')
  let focusAreas = $state('')
  let commStyle = $state('Direct')
  let language = $state('Technical')
  let customRules = $state('')

  const commStyles = ['Direct', 'Collaborative', 'Analytical', 'Supportive']
  const languages = ['Technical', 'Product', 'Mixed']

  const managementAgents = $derived(agents.filter(a => a.category === 'management'))
  const devAgents = $derived(agents.filter(a => a.category === 'dev'))
  const specialAgentsList = $derived(agents.filter(a => a.category === 'special'))

  async function loadAgents() {
    if (!projectPath) return
    loading = true
    selectedAgent = null
    try {
      agents = await ListAgents(projectPath)
      loading = false
      // try loading existing profile
      if (projectPath) {
        try {
          const existing = await GetSavedProfile(projectPath)
          if (existing.agentName) {
            fillForm(existing)
            selectedAgent = existing.agentName
          }
        } catch (_) { /* no saved profile */ }
      }
    } catch (e) {
      error = e.message || 'Failed to load agents'
      loading = false
    }
  }

  async function selectAgent(name) {
    selectedAgent = name
    try {
      const form = await LoadAgent(projectPath, name)
      fillForm(form)
    } catch (e) {
      error = e.message || 'Failed to load agent'
    }
  }

  function fillForm(form) {
    agentName = form.agentName || ''
    roleTitle = form.roleTitle || ''
    description = form.description || ''
    responsibilities = form.responsibilities || ''
    focusAreas = form.focusAreas || ''
    commStyle = form.communicationStyle || 'Direct'
    language = form.languagePreference || 'Technical'
    customRules = form.customRules || ''
  }

  function generateMarkdown() {
    const lines = [`# About Me — ${roleTitle || 'Untitled Role'}\n`]
    if (description) lines.push(`## Role\n${description}\n`)
    if (responsibilities) lines.push(`## Responsibilities\n${responsibilities}\n`)
    if (focusAreas) lines.push(`## Focus Areas\n${focusAreas}\n`)
    lines.push(`## Communication Style\n${commStyle}\n`)
    return lines.join('\n')
  }

  async function save() {
    if (!projectPath) {
      error = 'No project selected'
      return
    }
    saving = true
    saved = false
    error = ''
    try {
      await SaveProfile({
        agentName,
        roleTitle,
        description,
        responsibilities,
        focusAreas,
        communicationStyle: commStyle,
        languagePreference: language,
        customRules,
      }, projectPath)
      saved = true
      setTimeout(() => saved = false, 2000)
    } catch (e) {
      error = e.message || 'Failed to save profile'
    } finally {
      saving = false
    }
  }
</script>

<div class="h-full flex flex-col gap-4">
  <div class="flex items-center justify-between">
    <h2 class="text-xl font-bold text-ck-pink">Profile Editor</h2>
    <div class="flex items-center gap-2">
      {#if saved}<span class="text-xs text-ck-green">Saved!</span>{/if}
      {#if error}<span class="text-xs text-red-400">{error}</span>{/if}
    </div>
  </div>

  {#if loading}
    <p class="text-sm text-ck-dim">Loading agents...</p>
  {:else}
    <!-- Agent Selector -->
    <div class="space-y-3">
      <h3 class="text-xs font-semibold uppercase tracking-wider text-ck-dim">Management</h3>
      <div class="flex flex-wrap gap-1.5">
        {#each managementAgents as agent}
          <button
            onclick={() => selectAgent(agent.name)}
            title={agent.description}
            class="px-2.5 py-1 text-xs rounded-md transition-colors
              {selectedAgent === agent.name ? 'bg-ck-rose text-white' : 'bg-ck-dark text-ck-dim hover:text-white hover:bg-ck-pink/20'}"
          >{agent.name}</button>
        {/each}
      </div>

      <button
        onclick={() => devExpanded = !devExpanded}
        class="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider text-ck-dim hover:text-white transition-colors"
      >
        <span class="transition-transform {devExpanded ? 'rotate-90' : ''}">&rsaquo;</span>
        Developer ({devAgents.length})
      </button>
      {#if devExpanded}
        <div class="flex flex-wrap gap-1.5">
          {#each devAgents as agent}
            <button
              onclick={() => selectAgent(agent.name)}
              title={agent.description}
              class="px-2.5 py-1 text-xs rounded-md transition-colors
                {selectedAgent === agent.name ? 'bg-ck-rose text-white' : 'bg-ck-dark text-ck-dim hover:text-white hover:bg-ck-pink/20'}"
            >{agent.name}</button>
          {/each}
        </div>
      {/if}

      {#if specialAgentsList.length > 0}
        <button
          onclick={() => specialExpanded = !specialExpanded}
          class="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider text-ck-dim hover:text-white transition-colors"
        >
          <span class="transition-transform {specialExpanded ? 'rotate-90' : ''}">&rsaquo;</span>
          Special ({specialAgentsList.length})
        </button>
        {#if specialExpanded}
          <div class="flex flex-wrap gap-1.5">
            {#each specialAgentsList as agent}
              <button
                onclick={() => selectAgent(agent.name)}
                title={agent.description}
                class="px-2.5 py-1 text-xs rounded-md transition-colors
                  {selectedAgent === agent.name ? 'bg-ck-rose text-white' : 'bg-ck-dark text-ck-dim hover:text-white hover:bg-ck-pink/20'}"
              >{agent.name}</button>
            {/each}
          </div>
        {/if}
      {/if}
    </div>

    <div class="flex-1 grid grid-cols-2 gap-4 min-h-0">
      <!-- Form -->
      <div class="overflow-auto space-y-3 pr-2">
        <label class="block">
          <span class="text-xs text-ck-dim">Role Title</span>
          <input type="text" bind:value={roleTitle}
            class="w-full mt-1 px-3 py-2 bg-ck-dark border border-gray-700 rounded-md text-sm text-white focus:outline-none focus:ring-2 focus:ring-ck-pink" />
        </label>

        <label class="block">
          <span class="text-xs text-ck-dim">Description</span>
          <textarea bind:value={description} rows="2"
            class="w-full mt-1 px-3 py-2 bg-ck-dark border border-gray-700 rounded-md text-sm text-white resize-none focus:outline-none focus:ring-2 focus:ring-ck-pink"></textarea>
        </label>

        <label class="block">
          <span class="text-xs text-ck-dim">Responsibilities</span>
          <textarea bind:value={responsibilities} rows="3"
            class="w-full mt-1 px-3 py-2 bg-ck-dark border border-gray-700 rounded-md text-sm text-white resize-none focus:outline-none focus:ring-2 focus:ring-ck-pink"></textarea>
        </label>

        <label class="block">
          <span class="text-xs text-ck-dim">Focus Areas</span>
          <textarea bind:value={focusAreas} rows="2"
            class="w-full mt-1 px-3 py-2 bg-ck-dark border border-gray-700 rounded-md text-sm text-white resize-none focus:outline-none focus:ring-2 focus:ring-ck-pink"></textarea>
        </label>

        <div class="grid grid-cols-2 gap-3">
          <label class="block">
            <span class="text-xs text-ck-dim">Communication Style</span>
            <select bind:value={commStyle}
              class="w-full mt-1 px-3 py-2 bg-ck-dark border border-gray-700 rounded-md text-sm text-white focus:outline-none focus:ring-2 focus:ring-ck-pink">
              {#each commStyles as s}<option value={s}>{s}</option>{/each}
            </select>
          </label>
          <label class="block">
            <span class="text-xs text-ck-dim">Language</span>
            <select bind:value={language}
              class="w-full mt-1 px-3 py-2 bg-ck-dark border border-gray-700 rounded-md text-sm text-white focus:outline-none focus:ring-2 focus:ring-ck-pink">
              {#each languages as l}<option value={l}>{l}</option>{/each}
            </select>
          </label>
        </div>

        <label class="block">
          <span class="text-xs text-ck-dim">Custom Rules</span>
          <textarea bind:value={customRules} rows="4"
            class="w-full mt-1 px-3 py-2 bg-ck-dark border border-gray-700 rounded-md text-sm text-white resize-none focus:outline-none focus:ring-2 focus:ring-ck-pink"></textarea>
        </label>

        <button onclick={save} disabled={saving || !$currentProject}
          class="w-full py-2 bg-ck-rose text-white text-sm font-semibold rounded-md hover:bg-ck-pink transition-colors disabled:opacity-50">
          {saving ? 'Saving...' : 'Save Profile'}
        </button>
      </div>

      <!-- Preview -->
      <MarkdownPreview content={generateMarkdown()} />
    </div>
  {/if}
</div>
