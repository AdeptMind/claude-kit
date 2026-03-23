<script>
  import { currentProject } from '../stores/project.js'
  import MarkdownEditor from '../components/MarkdownEditor.svelte'

  let stories = $state([])
  let stats = $state({ done: 0, inProgress: 0, todo: 0, total: 0 })
  let loading = $state(true)
  let selectedStoryId = $state(null)

  let statusFilter = $state('all')
  let priorityFilter = $state('all')
  let componentFilter = $state('all')
  let search = $state('')

  let projectPath = $state('')
  currentProject.subscribe(p => {
    projectPath = p?.path || ''
    loadStories()
  })

  let components = $derived([...new Set(stories.map(s => s.component).filter(Boolean))])

  let filtered = $derived(() => {
    return stories.filter(s => {
      if (statusFilter !== 'all' && s.status !== statusFilter) return false
      if (priorityFilter !== 'all' && s.priority !== Number(priorityFilter)) return false
      if (componentFilter !== 'all' && s.component !== componentFilter) return false
      if (search && !s.title.toLowerCase().includes(search.toLowerCase()) && !s.id.toLowerCase().includes(search.toLowerCase())) return false
      return true
    })
  })

  let todoStories = $derived(filtered().filter(s => s.status === 'todo'))
  let inProgressStories = $derived(filtered().filter(s => s.status === 'in-progress'))
  let doneStories = $derived(filtered().filter(s => s.status === 'done'))

  // Selected story — reactive so it updates if stories change (status change)
  let selectedStory = $derived(selectedStoryId ? stories.find(s => s.id === selectedStoryId) : null)

  const priorityColors = { 1: 'bg-ck-rose/20 text-ck-rose border-ck-rose/30', 2: 'bg-ck-dim/20 text-ck-dim border-ck-dim/30' }
  const statusColors = { todo: 'bg-gray-500', 'in-progress': 'bg-ck-gold', done: 'bg-ck-green' }
  const statusLabels = { todo: 'Todo', 'in-progress': 'In Progress', done: 'Done' }

  // Auto-refresh stories every 5s to catch status changes
  let refreshInterval
  $effect(() => {
    loadStories()
    refreshInterval = setInterval(loadStories, 5000)
    return () => clearInterval(refreshInterval)
  })

  async function loadStories() {
    try {
      const { List, GetStats } = await import('../../wailsjs/go/main/StoryService.js')
      const [list, st] = await Promise.all([List(projectPath), GetStats(projectPath)])
      stories = list || []
      stats = st || stats
    } catch {
      stories = [
        { id: 'CK-001', title: 'Setup Wails project structure', status: 'done', priority: 1, round: 1, component: 'infra', type: 'setup', acceptanceCriteria: ['Wails init with Svelte', 'TailwindCSS configured', 'Build produces .app'], dependsOn: [] },
        { id: 'CK-002', title: 'Implement sidebar navigation', status: 'done', priority: 1, round: 1, component: 'ui', type: 'feature', acceptanceCriteria: ['5 nav icons', 'Active state highlighted', 'Responsive collapse'], dependsOn: ['CK-001'] },
        { id: 'CK-003', title: 'Dashboard page with phase progress', status: 'done', priority: 1, round: 1, component: 'ui', type: 'feature', acceptanceCriteria: ['Project name + role', 'Phase progress bar', 'Recent files list'], dependsOn: ['CK-002'] },
        { id: 'CK-004', title: 'File manager with preview', status: 'in-progress', priority: 1, round: 2, component: 'ui', type: 'feature', acceptanceCriteria: ['Tree view', 'File preview', 'Drag and drop'], dependsOn: ['CK-002'] },
        { id: 'CK-005', title: 'Workflow launcher integration', status: 'in-progress', priority: 1, round: 2, component: 'backend', type: 'integration', acceptanceCriteria: ['Phase diagram', 'Artifact monitor', 'Open in Cowork'], dependsOn: ['CK-003'] },
        { id: 'CK-006', title: 'Story board view', status: 'in-progress', priority: 1, round: 2, component: 'ui', type: 'feature', acceptanceCriteria: ['3-column board', 'Filters', 'Story detail modal'], dependsOn: ['CK-002'] },
        { id: 'CK-007', title: 'Role-based sidebar switching', status: 'todo', priority: 2, round: 3, component: 'ui', type: 'feature', acceptanceCriteria: ['Management sidebar', 'Dev sidebar', 'Immediate switch'], dependsOn: ['CK-004'] },
        { id: 'CK-008', title: 'Profile editor with Cowork integration', status: 'todo', priority: 2, round: 3, component: 'ui', type: 'feature', acceptanceCriteria: ['Agent selector', 'Form fields', 'Live preview', 'Save to Cowork'], dependsOn: ['CK-002'] },
        { id: 'CK-009', title: 'Settings persistence to disk', status: 'todo', priority: 2, round: 3, component: 'backend', type: 'config', acceptanceCriteria: ['Save to ~/.claude-kit/state.json', 'Load on startup'], dependsOn: ['CK-001'] },
        { id: 'CK-010', title: 'macOS app signing and notarization', status: 'todo', priority: 2, round: 4, component: 'infra', type: 'infra', acceptanceCriteria: ['Apple Developer cert', 'Notarize with xcrun', 'No Gatekeeper warning'], dependsOn: ['CK-009'] },
      ]
      stats = { done: 3, inProgress: 3, todo: 4, total: 10 }
    }
    loading = false
  }

  function openStory(story) {
    selectedStoryId = story.id
  }

  let editingStory = $state(false)
  let editContent = $state('')

  async function editStory() {
    try {
      const { GetStoryMarkdown, LockStory } = await import('../../wailsjs/go/main/StoryService.js')
      editContent = await GetStoryMarkdown(projectPath, selectedStory.id)
      try { await LockStory(projectPath, selectedStory.id) } catch {}
    } catch {
      editContent = storyToMarkdown(selectedStory)
    }
    editingStory = true
  }

  async function saveStoryEdit(content) {
    try {
      const { SaveStoryMarkdown, UnlockStory } = await import('../../wailsjs/go/main/StoryService.js')
      await SaveStoryMarkdown(projectPath, selectedStory.id, content)
      try { await UnlockStory(projectPath, selectedStory.id) } catch {}
    } catch {}
    editingStory = false
    loadStories()
  }

  async function cancelEdit() {
    try {
      const { UnlockStory } = await import('../../wailsjs/go/main/StoryService.js')
      await UnlockStory(projectPath, selectedStory.id)
    } catch {}
    editingStory = false
  }

  function storyToMarkdown(s) {
    let md = `# ${s.id}: ${s.title}\n\n`
    md += `**Priority:** P${s.priority}\n`
    if (s.round) md += `**Round:** ${s.round}\n`
    if (s.component) md += `**Component:** ${s.component}\n`
    md += `**Status:** ${s.status}\n\n`
    if (s.acceptanceCriteria?.length) {
      md += `## Acceptance Criteria\n`
      s.acceptanceCriteria.forEach(ac => md += `- ${ac}\n`)
      md += '\n'
    }
    if (s.dependsOn?.length) {
      md += `## Dependencies\n`
      s.dependsOn.forEach(d => md += `- ${d}\n`)
      md += '\n'
    }
    md += `## Notes\n_Add your notes here..._\n`
    return md
  }

  function closeStory() {
    selectedStoryId = null
    editingStory = false
  }

  $effect(() => {
    function handleEditStory() {
      if (selectedStoryId && !editingStory) editStory()
    }
    function handleCloseModal() {
      if (editingStory) {
        cancelEdit()
      } else if (selectedStoryId) {
        closeStory()
      }
    }
    window.addEventListener('ck:edit-story', handleEditStory)
    window.addEventListener('ck:close-modal', handleCloseModal)
    return () => {
      window.removeEventListener('ck:edit-story', handleEditStory)
      window.removeEventListener('ck:close-modal', handleCloseModal)
    }
  })
</script>

{#if loading}
  <div class="flex items-center justify-center h-full">
    <div class="w-6 h-6 border-2 border-ck-pink border-t-transparent rounded-full animate-spin"></div>
  </div>
{:else}
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-bold text-white">Stories</h1>
      <p class="text-sm text-ck-dim mt-1">
        {stats.total} stories &mdash; {stats.done} done, {stats.inProgress} in progress, {stats.todo} todo
      </p>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap items-center gap-3">
      <select bind:value={statusFilter}
        class="bg-ck-dark border border-gray-700 rounded-lg px-3 py-1.5 text-sm text-white focus:border-ck-pink focus:outline-none">
        <option value="all">All statuses</option>
        <option value="todo">Todo</option>
        <option value="in-progress">In Progress</option>
        <option value="done">Done</option>
      </select>

      <select bind:value={priorityFilter}
        class="bg-ck-dark border border-gray-700 rounded-lg px-3 py-1.5 text-sm text-white focus:border-ck-pink focus:outline-none">
        <option value="all">All priorities</option>
        <option value="1">P1</option>
        <option value="2">P2</option>
      </select>

      <select bind:value={componentFilter}
        class="bg-ck-dark border border-gray-700 rounded-lg px-3 py-1.5 text-sm text-white focus:border-ck-pink focus:outline-none">
        <option value="all">All components</option>
        {#each components as comp}
          <option value={comp}>{comp}</option>
        {/each}
      </select>

      <input
        type="text"
        placeholder="Search by title or ID..."
        bind:value={search}
        class="bg-ck-dark border border-gray-700 rounded-lg px-3 py-1.5 text-sm text-white
          placeholder-ck-dim focus:border-ck-pink focus:outline-none flex-1 min-w-[200px]"
      />
    </div>

    <!-- Board -->
    <div class="grid grid-cols-3 gap-4">
      {#snippet column(title, items, accent)}
        <div class="space-y-3">
          <div class="flex items-center gap-2">
            <span class="w-2 h-2 rounded-full {accent}"></span>
            <h2 class="text-sm font-semibold text-ck-dim uppercase tracking-wider">{title}</h2>
            <span class="text-xs text-ck-dim">({items.length})</span>
          </div>
          <div class="space-y-2">
            {#each items as story}
              <button
                onclick={() => openStory(story)}
                class="w-full text-left bg-ck-dark rounded-lg p-4 border transition-colors cursor-pointer
                  {selectedStoryId === story.id ? 'border-ck-rose ring-1 ring-ck-rose/30' : 'border-gray-800 hover:border-gray-600'}"
              >
                <div class="flex items-start justify-between gap-2 mb-2">
                  <span class="text-xs font-mono px-2 py-0.5 rounded bg-ck-gold/20 text-ck-gold border border-ck-gold/30">
                    {story.id}
                  </span>
                  <span class="text-xs px-2 py-0.5 rounded border {priorityColors[story.priority] || priorityColors[2]}">
                    P{story.priority}
                  </span>
                </div>
                <p class="text-sm text-white mb-3">
                  {#if story.locked}<span class="text-ck-gold text-xs mr-1" title="Being edited">🔒</span>{/if}{story.title}
                </p>
                <div class="flex items-center gap-2 flex-wrap">
                  {#if story.component}
                    <span class="text-xs px-2 py-0.5 rounded bg-white/5 text-ck-dim">{story.component}</span>
                  {/if}
                  {#if story.round}
                    <span class="text-xs text-ck-dim">R{story.round}</span>
                  {/if}
                </div>
              </button>
            {/each}
            {#if items.length === 0}
              <div class="text-center py-8 text-ck-dim text-xs">No stories</div>
            {/if}
          </div>
        </div>
      {/snippet}

      {@render column('Todo', todoStories, 'bg-gray-500')}
      {@render column('In Progress', inProgressStories, 'bg-ck-pink')}
      {@render column('Done', doneStories, 'bg-green-500')}
    </div>
  </div>

  <!-- Story detail modal -->
  {#if selectedStory}
    <!-- Backdrop -->
    <button
      class="fixed inset-0 bg-black/50 z-40 cursor-default"
      onclick={closeStory}
      aria-label="Close"
    ></button>

    <div class="fixed inset-x-0 top-[15%] mx-auto w-[600px] max-h-[70vh] bg-ck-dark border border-gray-700 rounded-xl shadow-2xl z-50 flex flex-col overflow-hidden">
      <!-- Header -->
      <div class="flex items-start justify-between px-6 py-4 border-b border-gray-800">
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-3 mb-2">
            <span class="text-sm font-mono px-2.5 py-1 rounded bg-ck-gold/20 text-ck-gold border border-ck-gold/30">
              {selectedStory.id}
            </span>
            <span class="text-xs px-2.5 py-1 rounded border {priorityColors[selectedStory.priority] || priorityColors[2]}">
              P{selectedStory.priority}
            </span>
            <span class="text-xs px-2.5 py-1 rounded flex items-center gap-1.5 {
              selectedStory.status === 'done' ? 'bg-green-500/20 text-green-400 border border-green-500/30' :
              selectedStory.status === 'in-progress' ? 'bg-ck-gold/20 text-ck-gold border border-ck-gold/30' :
              'bg-gray-500/20 text-gray-400 border border-gray-500/30'
            }">
              <span class="w-1.5 h-1.5 rounded-full {statusColors[selectedStory.status] || 'bg-gray-500'}"></span>
              {statusLabels[selectedStory.status] || selectedStory.status}
            </span>
          </div>
          <h2 class="text-lg font-semibold text-white">{selectedStory.title}</h2>
        </div>
        <div class="flex items-center gap-2 ml-4 shrink-0">
          <button onclick={editStory} class="px-3 py-1 text-xs rounded bg-ck-gold/20 text-ck-gold border border-ck-gold/30 hover:bg-ck-gold/30 transition-colors">
            Edit
          </button>
          <button
            onclick={closeStory}
            class="text-ck-dim hover:text-white transition-colors text-xl"
          >✕</button>
        </div>
      </div>

      <!-- Body -->
      <div class="flex-1 overflow-y-auto px-6 py-4 space-y-5">
        <!-- Metadata -->
        <div class="flex flex-wrap gap-3">
          {#if selectedStory.component}
            <div class="text-xs">
              <span class="text-ck-dim">Component:</span>
              <span class="ml-1 text-white px-2 py-0.5 rounded bg-white/5">{selectedStory.component}</span>
            </div>
          {/if}
          {#if selectedStory.type}
            <div class="text-xs">
              <span class="text-ck-dim">Type:</span>
              <span class="ml-1 text-white px-2 py-0.5 rounded bg-white/5">{selectedStory.type}</span>
            </div>
          {/if}
          {#if selectedStory.round}
            <div class="text-xs">
              <span class="text-ck-dim">Round:</span>
              <span class="ml-1 text-white px-2 py-0.5 rounded bg-white/5">{selectedStory.round}</span>
            </div>
          {/if}
        </div>

        <!-- Ralph Status -->
        <div>
          <h3 class="text-xs font-semibold text-ck-dim uppercase tracking-wider mb-2">Ralph Status</h3>
          <div class="flex items-center gap-2 text-sm">
            <span class="w-2 h-2 rounded-full {statusColors[selectedStory.status]}"></span>
            <span class="text-gray-300">{statusLabels[selectedStory.status]}</span>
            {#if selectedStory.status === 'done'}
              <span class="text-ck-green text-xs">— validated by Ralph</span>
            {:else if selectedStory.status === 'in-progress'}
              <span class="text-ck-gold text-xs">— dependencies met, eligible for implementation</span>
            {:else}
              <span class="text-ck-dim text-xs">— waiting for dependencies</span>
            {/if}
          </div>
        </div>

        <!-- Acceptance Criteria -->
        {#if selectedStory.acceptanceCriteria?.length > 0}
          <div>
            <h3 class="text-xs font-semibold text-ck-dim uppercase tracking-wider mb-2">Acceptance Criteria</h3>
            <ul class="space-y-1.5">
              {#each selectedStory.acceptanceCriteria as ac}
                <li class="flex items-start gap-2 text-sm text-gray-300">
                  <span class="shrink-0 mt-0.5 {selectedStory.status === 'done' ? 'text-ck-green' : 'text-ck-dim'}">
                    {selectedStory.status === 'done' ? '✓' : '○'}
                  </span>
                  {ac}
                </li>
              {/each}
            </ul>
          </div>
        {/if}

        <!-- Dependencies -->
        {#if selectedStory.dependsOn?.length > 0}
          <div>
            <h3 class="text-xs font-semibold text-ck-dim uppercase tracking-wider mb-2">Dependencies</h3>
            <div class="flex flex-wrap gap-2">
              {#each selectedStory.dependsOn as dep}
                {@const depStory = stories.find(s => s.id === dep)}
                <button
                  onclick={() => { selectedStoryId = dep }}
                  class="text-xs font-mono px-2 py-1 rounded border transition-colors cursor-pointer
                    {depStory?.status === 'done' ? 'bg-green-500/10 text-green-400 border-green-500/30 hover:bg-green-500/20' :
                     depStory?.status === 'in-progress' ? 'bg-ck-gold/10 text-ck-gold border-ck-gold/30 hover:bg-ck-gold/20' :
                     'bg-white/5 text-ck-dim border-gray-700 hover:bg-white/10'}"
                >
                  {dep}
                  {#if depStory}
                    <span class="ml-1 opacity-60">
                      {depStory.status === 'done' ? '✓' : depStory.status === 'in-progress' ? '◐' : '○'}
                    </span>
                  {/if}
                </button>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    </div>

    {#if editingStory}
      <MarkdownEditor
        content={editContent}
        filename="{selectedStory.id}.md"
        onSave={saveStoryEdit}
        onCancel={cancelEdit}
      />
    {/if}
  {/if}
{/if}
