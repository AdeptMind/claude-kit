<script>
  import PhaseProgress from '../components/PhaseProgress.svelte'
  import QuickActions from '../components/QuickActions.svelte'
  import { navigateTo } from '../stores/navigation.js'

  let project = $state(null)
  let role = $state(null)
  let workflowStatus = $state(null)
  let loading = $state(true)

  const phases = [
    { label: 'Break', status: 'done' },
    { label: 'Model', status: 'done' },
    { label: 'Analyze', status: 'active' },
    { label: 'Act', status: 'pending' },
    { label: 'Deliver', status: 'pending' },
  ]

  const recentFiles = [
    { name: 'problem.yaml', type: 'yaml', modified: '2 min ago' },
    { name: 'architecture.yaml', type: 'yaml', modified: '5 min ago' },
    { name: 'backlog.yaml', type: 'yaml', modified: '12 min ago' },
    { name: 'principles.md', type: 'md', modified: '1h ago' },
    { name: 'App.svelte', type: 'svelte', modified: '3h ago' },
  ]

  const fileIcons = {
    yaml: '📋',
    md: '📝',
    svelte: '🧩',
    ts: '🔷',
    js: '🟡',
    go: '🔵',
  }

  const quickActions = [
    {
      icon: '👥',
      title: 'Open in Cowork',
      description: 'Launch collaborative session',
      handler: () => {},
    },
    {
      icon: '🚀',
      title: 'Run Workflow',
      description: 'Execute BMAD pipeline',
      handler: () => navigateTo('workflow'),
    },
    {
      icon: '📁',
      title: 'Manage Files',
      description: 'Browse project artifacts',
      handler: () => navigateTo('files'),
    },
  ]

  $effect(() => {
    loadProject()
  })

  async function loadProject() {
    try {
      const { GetState } = await import('../../wailsjs/go/main/ProjectService.js')
      const state = await GetState()
      project = state
    } catch {
      // Bindings not available yet — use placeholder
      project = { name: 'claude-kit-ui', path: '~/workspace/claude-kit' }
    }

    try {
      const { GetCurrent } = await import('../../wailsjs/go/main/RoleService.js')
      role = await GetCurrent()
    } catch {
      role = { name: 'Architect', color: 'gold' }
    }

    try {
      const { GetStatus } = await import('../../wailsjs/go/main/WorkflowService.js')
      workflowStatus = await GetStatus()
    } catch {
      workflowStatus = { currentPhase: 'analyze' }
    }

    loading = false
  }

  function handleNewProject() {
    // Will call Create() when bindings available
  }

  function handleOpenProject() {
    // Will call Open() when bindings available
  }
</script>

{#if loading}
  <div class="flex items-center justify-center h-full">
    <div class="w-6 h-6 border-2 border-ck-pink border-t-transparent rounded-full animate-spin"></div>
  </div>
{:else if !project}
  <!-- Welcome screen -->
  <div class="flex items-center justify-center h-full">
    <div class="text-center max-w-md">
      <h1 class="text-3xl font-bold text-white mb-2">Welcome to Claude Kit</h1>
      <p class="text-ck-dim mb-8">Get started by creating a new project or opening an existing one.</p>
      <div class="flex gap-4 justify-center">
        <button
          onclick={handleNewProject}
          class="px-6 py-2.5 rounded-lg bg-ck-rose text-white font-medium text-sm
            hover:bg-ck-pink transition-colors"
        >
          New Project
        </button>
        <button
          onclick={handleOpenProject}
          class="px-6 py-2.5 rounded-lg bg-ck-dark border border-gray-700 text-white font-medium text-sm
            hover:border-ck-dim transition-colors"
        >
          Open Project
        </button>
      </div>
    </div>
  </div>
{:else}
  <!-- Dashboard -->
  <div class="space-y-8">
    <!-- Header -->
    <div class="flex items-center gap-4">
      <h1 class="text-2xl font-bold text-white">{project.name}</h1>
      {#if role}
        <span class="px-3 py-1 rounded-full text-xs font-semibold bg-ck-gold/20 text-ck-gold border border-ck-gold/30">
          {role.name}
        </span>
      {/if}
    </div>

    <!-- Phase Progress -->
    <section>
      <h2 class="text-sm font-semibold text-ck-dim uppercase tracking-wider mb-4">BMAD Progress</h2>
      <div class="bg-ck-dark rounded-xl p-6 border border-gray-800">
        <PhaseProgress {phases} />
      </div>
    </section>

    <!-- Recent Files -->
    <section>
      <h2 class="text-sm font-semibold text-ck-dim uppercase tracking-wider mb-4">Recent Files</h2>
      <div class="bg-ck-dark rounded-xl border border-gray-800 divide-y divide-gray-800">
        {#each recentFiles as file}
          <div class="flex items-center justify-between px-5 py-3 hover:bg-white/5 transition-colors cursor-pointer">
            <div class="flex items-center gap-3">
              <span class="text-base">{fileIcons[file.type] || '📄'}</span>
              <span class="text-sm text-white">{file.name}</span>
            </div>
            <span class="text-xs text-ck-dim">{file.modified}</span>
          </div>
        {/each}
      </div>
    </section>

    <!-- Quick Actions -->
    <section>
      <h2 class="text-sm font-semibold text-ck-dim uppercase tracking-wider mb-4">Quick Actions</h2>
      <QuickActions actions={quickActions} />
    </section>
  </div>
{/if}
