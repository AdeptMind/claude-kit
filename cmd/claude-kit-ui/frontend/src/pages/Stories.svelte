<script>
  import { currentProject } from '../stores/project.js'

  let stories = $state([])
  let stats = $state({ done: 0, inProgress: 0, todo: 0, total: 0 })
  let loading = $state(true)

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

  const priorityColors = { 1: 'bg-ck-rose/20 text-ck-rose border-ck-rose/30', 2: 'bg-ck-dim/20 text-ck-dim border-ck-dim/30' }
  const statusLabels = { todo: 'Todo', 'in-progress': 'In Progress', done: 'Done' }

  $effect(() => { loadStories() })

  async function loadStories() {
    try {
      const { List, GetStats } = await import('../../wailsjs/go/main/StoryService.js')
      const [list, st] = await Promise.all([List(projectPath), GetStats(projectPath)])
      stories = list || []
      stats = st || stats
    } catch {
      stories = [
        { id: 'CK-001', title: 'Setup Wails project structure', status: 'done', priority: 1, round: 1, component: 'infra', type: 'setup', acceptanceCriteria: [], dependsOn: [] },
        { id: 'CK-002', title: 'Implement sidebar navigation', status: 'done', priority: 1, round: 1, component: 'ui', type: 'feature', acceptanceCriteria: [], dependsOn: ['CK-001'] },
        { id: 'CK-003', title: 'Dashboard page with phase progress', status: 'done', priority: 1, round: 1, component: 'ui', type: 'feature', acceptanceCriteria: [], dependsOn: ['CK-002'] },
        { id: 'CK-004', title: 'File manager with YAML viewer', status: 'in-progress', priority: 1, round: 2, component: 'ui', type: 'feature', acceptanceCriteria: [], dependsOn: ['CK-002'] },
        { id: 'CK-005', title: 'Workflow launcher integration', status: 'in-progress', priority: 1, round: 2, component: 'backend', type: 'integration', acceptanceCriteria: [], dependsOn: ['CK-003'] },
        { id: 'CK-006', title: 'Story board view', status: 'in-progress', priority: 1, round: 2, component: 'ui', type: 'feature', acceptanceCriteria: [], dependsOn: ['CK-002'] },
        { id: 'CK-007', title: 'Role-based sidebar switching', status: 'todo', priority: 2, round: 3, component: 'ui', type: 'feature', acceptanceCriteria: [], dependsOn: ['CK-004'] },
        { id: 'CK-008', title: 'Profile editor with account management', status: 'todo', priority: 2, round: 3, component: 'ui', type: 'feature', acceptanceCriteria: [], dependsOn: ['CK-002'] },
        { id: 'CK-009', title: 'Settings persistence to disk', status: 'todo', priority: 2, round: 3, component: 'backend', type: 'config', acceptanceCriteria: [], dependsOn: ['CK-001'] },
        { id: 'CK-010', title: 'macOS notarization pipeline', status: 'todo', priority: 2, round: 4, component: 'infra', type: 'infra', acceptanceCriteria: [], dependsOn: ['CK-009'] },
      ]
      stats = { done: 3, inProgress: 3, todo: 4, total: 10 }
    }
    loading = false
  }
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
              <div class="bg-ck-dark rounded-lg p-4 border border-gray-800 hover:border-gray-700 transition-colors">
                <div class="flex items-start justify-between gap-2 mb-2">
                  <span class="text-xs font-mono px-2 py-0.5 rounded bg-ck-gold/20 text-ck-gold border border-ck-gold/30">
                    {story.id}
                  </span>
                  <span class="text-xs px-2 py-0.5 rounded border {priorityColors[story.priority] || priorityColors[2]}">
                    P{story.priority}
                  </span>
                </div>
                <p class="text-sm text-white mb-3">{story.title}</p>
                <div class="flex items-center gap-2 flex-wrap">
                  {#if story.component}
                    <span class="text-xs px-2 py-0.5 rounded bg-white/5 text-ck-dim">{story.component}</span>
                  {/if}
                  {#if story.round}
                    <span class="text-xs text-ck-dim">R{story.round}</span>
                  {/if}
                </div>
              </div>
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
{/if}
