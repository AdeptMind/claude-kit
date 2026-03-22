<script>
  import AgentSelector from '../components/AgentSelector.svelte'
  import MarkdownPreview from '../components/MarkdownPreview.svelte'

  let roleTitle = $state('')
  let responsibilities = $state('')
  let focusAreas = $state('')
  let commStyle = $state('Direct')
  let language = $state('Technical')
  let customRules = $state('')
  let saving = $state(false)
  let saved = $state(false)

  const commStyles = ['Direct', 'Collaborative', 'Analytical', 'Supportive']
  const languages = ['Technical', 'Product', 'Mixed']

  function onAgentSelect(role) {
    roleTitle = role.label
    responsibilities = `Core ${role.label.toLowerCase()} responsibilities`
    focusAreas = ''
    commStyle = 'Direct'
    language = 'Technical'
    customRules = ''
  }

  function generateMarkdown() {
    const lines = [`# ${roleTitle || 'Untitled Role'}\n`]
    if (responsibilities) lines.push(`## Responsibilities\n${responsibilities}\n`)
    if (focusAreas) lines.push(`## Focus Areas\n${focusAreas}\n`)
    lines.push(`## Communication\n- **Style**: ${commStyle}\n- **Language**: ${language}\n`)
    if (customRules) lines.push(`## Custom Rules\n${customRules}`)
    return lines.join('\n')
  }

  async function save() {
    saving = true
    saved = false
    try {
      // Placeholder: will call SaveProfile binding when available
      await new Promise(r => setTimeout(r, 500))
      saved = true
      setTimeout(() => saved = false, 2000)
    } finally {
      saving = false
    }
  }
</script>

<div class="h-full flex flex-col gap-4">
  <div class="flex items-center justify-between">
    <h2 class="text-xl font-bold text-ck-pink">Profile Editor</h2>
    {#if saved}<span class="text-xs text-ck-green">Saved!</span>{/if}
  </div>

  <AgentSelector onselect={onAgentSelect} />

  <div class="flex-1 grid grid-cols-2 gap-4 min-h-0">
    <!-- Form -->
    <div class="overflow-auto space-y-3 pr-2">
      <label class="block">
        <span class="text-xs text-ck-dim">Role Title</span>
        <input type="text" bind:value={roleTitle}
          class="w-full mt-1 px-3 py-2 bg-ck-dark border border-gray-700 rounded-md text-sm text-white focus:outline-none focus:ring-2 focus:ring-ck-pink" />
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
        <textarea bind:value={customRules} rows="3"
          class="w-full mt-1 px-3 py-2 bg-ck-dark border border-gray-700 rounded-md text-sm text-white resize-none focus:outline-none focus:ring-2 focus:ring-ck-pink"></textarea>
      </label>

      <button onclick={save} disabled={saving}
        class="w-full py-2 bg-ck-rose text-white text-sm font-semibold rounded-md hover:bg-ck-pink transition-colors disabled:opacity-50">
        {saving ? 'Saving...' : 'Save Profile'}
      </button>
    </div>

    <!-- Preview -->
    <MarkdownPreview content={generateMarkdown()} />
  </div>
</div>
