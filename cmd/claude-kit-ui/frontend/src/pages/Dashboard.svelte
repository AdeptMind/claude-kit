<script>
  import PhaseProgress from '../components/PhaseProgress.svelte'
  import QuickActions from '../components/QuickActions.svelte'
  import { navigateTo } from '../stores/navigation.js'

  let project = $state(null)
  let role = $state(null)
  let workflowStatus = $state(null)
  let loading = $state(true)
  let selectedFile = $state(null)
  let fileContent = $state('')
  let editMode = $state(false)
  let saving = $state(false)

  const phases = [
    { label: 'Break', status: 'done' },
    { label: 'Model', status: 'done' },
    { label: 'Analyze', status: 'active' },
    { label: 'Act', status: 'pending' },
    { label: 'Deliver', status: 'pending' },
  ]

  const recentFiles = [
    { name: 'problem.yaml', type: 'yaml', modified: '2 min ago', path: '.claude/output/problem.yaml' },
    { name: 'architecture.yaml', type: 'yaml', modified: '5 min ago', path: '.claude/output/architecture.yaml' },
    { name: 'backlog.yaml', type: 'yaml', modified: '12 min ago', path: '.claude/output/backlog.yaml' },
    { name: 'principles.md', type: 'md', modified: '1h ago', path: '.claude/output/principles.md' },
    { name: 'user-journey.yaml', type: 'yaml', modified: '3h ago', path: '.claude/output/user-journey.yaml' },
  ]

  const fileIcons = {
    yaml: '📋',
    md: '📝',
    svelte: '🧩',
    ts: '🔷',
    js: '🟡',
    go: '🔵',
  }

  // Files that are editable (not yet consumed by a BMAD phase)
  const editableFiles = new Set(['principles.md', 'user-journey.yaml', 'problem.yaml'])

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

  async function openFile(file) {
    selectedFile = file
    editMode = false
    try {
      const { ReadPreview } = await import('../../wailsjs/go/main/FileService.js')
      const result = await ReadPreview(file.path)
      fileContent = result
    } catch {
      // Placeholder content for demo
      const placeholders = {
        'problem.yaml': 'project_name: my-project\nversion: "1.0"\nphase: break\n\nproblem_statement:\n  summary: >\n    Your project description here\n  target_users:\n    - User type 1\n  pain_points:\n    - Pain point 1',
        'architecture.yaml': 'project_name: my-project\nversion: "1.0"\nphase: model\n\ncomponents:\n  - name: api-server\n    type: service\n    responsibility: Handle HTTP requests',
        'backlog.yaml': 'project_name: my-project\nversion: "1.0"\nphase: model\n\ntasks:\n  - id: T-001\n    title: Setup project\n    priority: P1',
        'principles.md': '# Project Principles\n\n## Code Quality\n- DRY, KISS, SOLID\n- Pattern-first coding\n\n## Testing\n- TDD enforced\n- 80% coverage target',
        'user-journey.yaml': 'journeys:\n  - name: "User login"\n    persona: "end user"\n    story_refs: [US-001]\n    steps:\n      - action: "Navigate to /login"\n        expected: "Login form displayed"',
      }
      fileContent = placeholders[file.name] || `# ${file.name}\n\nFile content will appear here when connected to a project.`
    }
  }

  async function saveFile() {
    saving = true
    try {
      // Will use FileService.Write when available
      await new Promise(resolve => setTimeout(resolve, 500))
    } catch {}
    saving = false
    editMode = false
  }

  function closeFile() {
    selectedFile = null
    fileContent = ''
    editMode = false
  }

  function handleNewProject() {}
  function handleOpenProject() {}
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
          class="px-6 py-2.5 rounded-lg bg-ck-rose text-white font-medium text-sm hover:bg-ck-pink transition-colors"
        >
          New Project
        </button>
        <button
          onclick={handleOpenProject}
          class="px-6 py-2.5 rounded-lg bg-ck-dark border border-gray-700 text-white font-medium text-sm hover:border-ck-dim transition-colors"
        >
          Open Project
        </button>
      </div>
    </div>
  </div>
{:else}
  <div class="flex h-full gap-0">
    <!-- Main dashboard content -->
    <div class="flex-1 space-y-8 overflow-y-auto {selectedFile ? 'pr-4' : ''}">
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
            <button
              onclick={() => openFile(file)}
              class="w-full flex items-center justify-between px-5 py-3 hover:bg-white/5 transition-colors cursor-pointer
                {selectedFile?.name === file.name ? 'bg-ck-rose/10 border-l-2 border-l-ck-rose' : ''}"
            >
              <div class="flex items-center gap-3">
                <span class="text-base">{fileIcons[file.type] || '📄'}</span>
                <span class="text-sm text-white">{file.name}</span>
                {#if editableFiles.has(file.name)}
                  <span class="text-[10px] px-1.5 py-0.5 rounded bg-ck-gold/10 text-ck-gold border border-ck-gold/20">editable</span>
                {/if}
              </div>
              <span class="text-xs text-ck-dim">{file.modified}</span>
            </button>
          {/each}
        </div>
      </section>

      <!-- Quick Actions -->
      <section>
        <h2 class="text-sm font-semibold text-ck-dim uppercase tracking-wider mb-4">Quick Actions</h2>
        <QuickActions actions={quickActions} />
      </section>
    </div>

    <!-- File viewer/editor panel (slides in from right) -->
    {#if selectedFile}
      <div class="w-[45%] min-w-[400px] flex flex-col bg-ck-dark border-l border-gray-800 rounded-l-xl ml-4">
        <!-- Panel header -->
        <div class="flex items-center justify-between px-4 py-3 border-b border-gray-800">
          <div class="flex items-center gap-2">
            <span>{fileIcons[selectedFile.type] || '📄'}</span>
            <span class="text-sm font-medium text-white">{selectedFile.name}</span>
          </div>
          <div class="flex items-center gap-2">
            {#if editableFiles.has(selectedFile.name)}
              {#if editMode}
                <button
                  onclick={saveFile}
                  disabled={saving}
                  class="px-3 py-1 text-xs rounded bg-ck-green/20 text-ck-green border border-ck-green/30 hover:bg-ck-green/30 transition-colors disabled:opacity-50"
                >
                  {saving ? 'Saving...' : 'Save'}
                </button>
                <button
                  onclick={() => editMode = false}
                  class="px-3 py-1 text-xs rounded bg-ck-dark text-ck-dim border border-gray-700 hover:text-white transition-colors"
                >
                  Cancel
                </button>
              {:else}
                <button
                  onclick={() => editMode = true}
                  class="px-3 py-1 text-xs rounded bg-ck-gold/20 text-ck-gold border border-ck-gold/30 hover:bg-ck-gold/30 transition-colors"
                >
                  Edit
                </button>
              {/if}
            {/if}
            <button
              onclick={closeFile}
              class="px-2 py-1 text-ck-dim hover:text-white transition-colors text-lg"
            >
              ✕
            </button>
          </div>
        </div>

        <!-- File content -->
        <div class="flex-1 overflow-y-auto p-4">
          {#if editMode}
            <textarea
              bind:value={fileContent}
              class="w-full h-full bg-ck-bg border border-gray-700 rounded-lg p-4 text-sm font-mono text-gray-300 resize-none focus:border-ck-pink focus:outline-none"
              spellcheck="false"
            ></textarea>
          {:else}
            <pre class="text-sm font-mono text-gray-300 whitespace-pre-wrap leading-relaxed">{fileContent}</pre>
          {/if}
        </div>
      </div>
    {/if}
  </div>
{/if}
