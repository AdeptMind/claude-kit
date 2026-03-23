<script>
  import { currentProject } from '../stores/project.js'

  let agents = $state([])
  let loading = $state(true)
  let selectedAgent = $state(null)
  let saving = $state(false)
  let saved = $state(false)
  let errorMsg = $state('')
  let devExpanded = $state(false)
  let projectPath = $state('')

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

  let managementAgents = $derived(agents.filter(a => a.category === 'management'))
  let devAgents = $derived(agents.filter(a => a.category === 'dev'))

  let previewMd = $derived(generateMarkdown())

  currentProject.subscribe(p => {
    const newPath = p?.path || ''
    if (newPath !== projectPath) {
      projectPath = newPath
      agents = []
      selectedAgent = null
      resetForm()
      if (newPath) loadAgents()
      else loading = false
    }
  })

  function resetForm() {
    agentName = ''
    roleTitle = ''
    description = ''
    responsibilities = ''
    focusAreas = ''
    commStyle = 'Direct'
    language = 'Technical'
    customRules = ''
  }

  async function loadAgents() {
    if (!projectPath) { loading = false; return }
    loading = true
    errorMsg = ''
    try {
      const svc = await import('../../wailsjs/go/main/ProfileService.js')
      const list = await svc.ListAgents(projectPath)
      agents = list || []
    } catch {
      agents = []
    }
    loading = false
  }

  async function selectAgent(name) {
    selectedAgent = name
    agentName = name
    try {
      const svc = await import('../../wailsjs/go/main/ProfileService.js')
      const form = await svc.LoadAgent(projectPath, name)
      roleTitle = form.roleTitle || toTitle(name)
      description = form.description || ''
      responsibilities = form.responsibilities || ''
      focusAreas = form.focusAreas || ''
      commStyle = form.communicationStyle || 'Direct'
      language = form.languagePreference || 'Technical'
      customRules = form.customRules || ''
    } catch {
      roleTitle = toTitle(name)
    }
  }

  function toTitle(s) {
    return s.split('-').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ')
  }

  function generateMarkdown() {
    if (!roleTitle) return ''
    let md = `# About Me — ${roleTitle}\n\n`
    if (description) md += `## Role\n${description}\n\n`
    if (responsibilities) md += `## Responsibilities\n${responsibilities}\n\n`
    if (focusAreas) md += `## Focus Areas\n${focusAreas}\n\n`
    md += `## Communication Style\n${commStyle}\n\n`
    md += `## Language\n${language}\n`
    return md
  }

  async function save() {
    if (!projectPath) { errorMsg = 'No project selected'; return }
    saving = true
    saved = false
    errorMsg = ''
    try {
      const svc = await import('../../wailsjs/go/main/ProfileService.js')
      await svc.SaveProfile({
        agentName, roleTitle, description, responsibilities,
        focusAreas, communicationStyle: commStyle,
        languagePreference: language, customRules,
      }, projectPath)
      saved = true
      setTimeout(() => saved = false, 2000)
    } catch (e) {
      errorMsg = 'Failed to save'
    }
    saving = false
  }
</script>

<div class="h-full flex flex-col gap-4">
  <div class="flex items-center justify-between">
    <h2 class="text-xl font-bold text-ck-pink">Profile Editor</h2>
    <div class="flex items-center gap-2">
      {#if saved}<span class="text-xs text-ck-green">Saved!</span>{/if}
      {#if errorMsg}<span class="text-xs text-red-400">{errorMsg}</span>{/if}
    </div>
  </div>

  {#if !projectPath}
    <div class="flex items-center justify-center h-64">
      <p class="text-sm text-ck-dim">Select a project first.</p>
    </div>
  {:else if loading}
    <div class="flex items-center justify-center h-64">
      <div class="w-5 h-5 border-2 border-ck-pink border-t-transparent rounded-full animate-spin"></div>
    </div>
  {:else if agents.length === 0}
    <div class="flex flex-col items-center justify-center h-64 gap-3">
      <p class="text-sm text-ck-dim">No agents installed in this project.</p>
      <p class="text-xs text-ck-dim">Run <code class="text-ck-gold bg-ck-dark px-1.5 py-0.5 rounded">ck init</code> to install agents.</p>
    </div>
  {:else}
    <div class="flex-1 flex gap-4 min-h-0">
      <!-- Left: Agent selector + Form -->
      <div class="w-1/2 flex flex-col gap-4 overflow-y-auto">
        <!-- Agent selector -->
        <div class="space-y-2">
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
            <span class="transition-transform {devExpanded ? 'rotate-90' : ''}">›</span>
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
        </div>

        <!-- Form -->
        {#if selectedAgent}
          <div class="space-y-3">
            <div>
              <label class="text-xs text-ck-dim">Role Title</label>
              <input bind:value={roleTitle} class="w-full mt-1 px-3 py-1.5 text-sm bg-ck-dark border border-gray-700 rounded text-white focus:border-ck-pink focus:outline-none" />
            </div>
            <div>
              <label class="text-xs text-ck-dim">Description</label>
              <textarea bind:value={description} rows="2" class="w-full mt-1 px-3 py-1.5 text-sm bg-ck-dark border border-gray-700 rounded text-white focus:border-ck-pink focus:outline-none resize-none"></textarea>
            </div>
            <div>
              <label class="text-xs text-ck-dim">Responsibilities</label>
              <textarea bind:value={responsibilities} rows="3" class="w-full mt-1 px-3 py-1.5 text-sm bg-ck-dark border border-gray-700 rounded text-white focus:border-ck-pink focus:outline-none resize-none"></textarea>
            </div>
            <div>
              <label class="text-xs text-ck-dim">Focus Areas</label>
              <textarea bind:value={focusAreas} rows="2" class="w-full mt-1 px-3 py-1.5 text-sm bg-ck-dark border border-gray-700 rounded text-white focus:border-ck-pink focus:outline-none resize-none"></textarea>
            </div>
            <div class="flex gap-3">
              <div class="flex-1">
                <label class="text-xs text-ck-dim">Communication Style</label>
                <select bind:value={commStyle} class="w-full mt-1 px-3 py-1.5 text-sm bg-ck-dark border border-gray-700 rounded text-white focus:border-ck-pink focus:outline-none">
                  {#each commStyles as s}<option value={s}>{s}</option>{/each}
                </select>
              </div>
              <div class="flex-1">
                <label class="text-xs text-ck-dim">Language</label>
                <select bind:value={language} class="w-full mt-1 px-3 py-1.5 text-sm bg-ck-dark border border-gray-700 rounded text-white focus:border-ck-pink focus:outline-none">
                  {#each languages as l}<option value={l}>{l}</option>{/each}
                </select>
              </div>
            </div>
            <div>
              <label class="text-xs text-ck-dim">Custom Rules</label>
              <textarea bind:value={customRules} rows="4" class="w-full mt-1 px-3 py-1.5 text-sm bg-ck-dark border border-gray-700 rounded text-white focus:border-ck-pink focus:outline-none resize-none"></textarea>
            </div>
            <button
              onclick={save}
              disabled={saving}
              class="px-4 py-2 text-sm rounded-lg bg-ck-rose text-white font-medium hover:bg-ck-pink transition-colors disabled:opacity-50"
            >{saving ? 'Saving...' : 'Save Profile'}</button>
          </div>
        {:else}
          <p class="text-sm text-ck-dim">Select an agent above to configure its Cowork profile.</p>
        {/if}
      </div>

      <!-- Right: Preview -->
      <div class="w-1/2 bg-ck-dark rounded-xl border border-gray-800 p-4 overflow-y-auto">
        <h3 class="text-xs font-semibold text-ck-dim uppercase tracking-wider mb-3">Preview — about-me.md</h3>
        {#if previewMd}
          <pre class="text-sm text-gray-300 whitespace-pre-wrap leading-relaxed">{previewMd}</pre>
        {:else}
          <p class="text-sm text-ck-dim">Select an agent to see the profile preview.</p>
        {/if}
      </div>
    </div>
  {/if}
</div>
