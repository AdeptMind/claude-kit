<script>
  import PhaseProgress from '../components/PhaseProgress.svelte'
  import QuickActions from '../components/QuickActions.svelte'
  import MarkdownEditor from '../components/MarkdownEditor.svelte'
  import ReadPreview from '../components/ReadPreview.svelte'
  import { navigateTo } from '../stores/navigation.js'

  let project = $state(null)
  let role = $state(null)
  let workflowStatus = $state(null)
  let loading = $state(true)

  // Modal state
  let readFile = $state(null)
  let readContent = $state('')
  let editFile = $state(null)
  let editContent = $state('')
  let deleteTarget = $state(null)

  const phases = [
    { label: 'Break', status: 'done' },
    { label: 'Model', status: 'done' },
    { label: 'Analyze', status: 'active' },
    { label: 'Act', status: 'pending' },
    { label: 'Deliver', status: 'pending' },
  ]

  const recentFiles = [
    { name: 'problem.md', type: 'md', modified: '2 min ago', path: '.claude/output/problem.md' },
    { name: 'architecture.md', type: 'md', modified: '5 min ago', path: '.claude/output/architecture.md' },
    { name: 'backlog.md', type: 'md', modified: '12 min ago', path: '.claude/output/backlog.md' },
    { name: 'principles.md', type: 'md', modified: '1h ago', path: '.claude/output/principles.md' },
    { name: 'user-journey.md', type: 'md', modified: '3h ago', path: '.claude/output/user-journey.md' },
  ]

  const fileIcons = {
    yaml: '📋',
    md: '📝',
    svelte: '🧩',
    ts: '🔷',
    js: '🟡',
    go: '🔵',
  }

  const editableFiles = new Set(['principles.md', 'user-journey.md', 'problem.md'])

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

  const placeholders = {
    'problem.md': '# Problem Definition\n\n## Project\n- **Name:** my-project\n- **Version:** 1.0\n- **Phase:** break\n\n## Problem Statement\nYour project description here\n\n## Target Users\n- User type 1\n\n## Pain Points\n- Pain point 1',
    'architecture.md': '# Architecture\n\n## Project\n- **Name:** my-project\n- **Version:** 1.0\n- **Phase:** model\n\n## Components\n\n### api-server\n- **Type:** service\n- **Responsibility:** Handle HTTP requests',
    'backlog.md': '# Backlog\n\n## Project\n- **Name:** my-project\n- **Version:** 1.0\n- **Phase:** model\n\n## Tasks\n\n### T-001 — Setup project\n- **Priority:** P1',
    'principles.md': '# Project Principles\n\n## Code Quality\n- DRY, KISS, SOLID\n- Pattern-first coding\n\n## Testing\n- TDD enforced\n- 80% coverage target',
    'user-journey.md': '# User Journeys\n\n## User login\n- **Persona:** end user\n- **Story refs:** US-001\n\n### Steps\n1. **Action:** Navigate to /login — **Expected:** Login form displayed',
  }

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

  async function loadFileContent(file) {
    let loaded = false
    try {
      const { ReadPreview: ReadPreviewFn } = await import('../../wailsjs/go/main/FileService.js')
      const result = await ReadPreviewFn(file.path)
      if (result && result.trim() && !result.includes('not yet implemented')) {
        return result
      }
    } catch {}
    return placeholders[file.name] || `# ${file.name}\n\nFile content will appear here when connected to a project.`
  }

  async function openRead(file) {
    readContent = await loadFileContent(file)
    readFile = file
  }

  async function openEdit(file) {
    editContent = await loadFileContent(file)
    editFile = file
  }

  async function handleSave(newContent) {
    try {
      // Will use FileService.Write when available
      await new Promise(resolve => setTimeout(resolve, 300))
    } catch {}
    editFile = null
    editContent = ''
  }

  function handleCancelEdit() {
    editFile = null
    editContent = ''
  }

  function closeRead() {
    readFile = null
    readContent = ''
  }

  function confirmDelete(file) {
    deleteTarget = file
  }

  async function executeDelete() {
    if (!deleteTarget) return
    try {
      // Will use FileService.Delete when available
      await new Promise(resolve => setTimeout(resolve, 300))
    } catch {}
    deleteTarget = null
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
          <div class="flex items-center justify-between px-5 py-3 hover:bg-white/5 transition-colors">
            <div class="flex items-center gap-3">
              <span class="text-base">{fileIcons[file.type] || '📄'}</span>
              <span class="text-sm text-white">{file.name}</span>
            </div>
            <div class="flex items-center gap-3">
              <span class="text-xs text-ck-dim mr-2">{file.modified}</span>
              <button
                onclick={() => openRead(file)}
                class="px-2.5 py-1 text-xs rounded border border-gray-700 text-ck-dim hover:text-white hover:border-gray-500 transition-colors"
              >
                Read
              </button>
              {#if editableFiles.has(file.name)}
                <button
                  onclick={() => openEdit(file)}
                  class="px-2.5 py-1 text-xs rounded border border-ck-gold/30 text-ck-gold hover:text-ck-pink hover:border-ck-pink/30 transition-colors"
                >
                  Edit
                </button>
              {/if}
              <button
                onclick={() => confirmDelete(file)}
                class="px-2 py-1 text-xs rounded border border-red-400/20 text-red-400/50 hover:text-red-400 hover:border-red-400/40 transition-colors"
              >
                ✕
              </button>
            </div>
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

  <!-- Read Preview Modal -->
  {#if readFile}
    <ReadPreview content={readContent} filename={readFile.name} onClose={closeRead} />
  {/if}

  <!-- Markdown Editor Modal -->
  {#if editFile}
    <MarkdownEditor content={editContent} filename={editFile.name} onSave={handleSave} onCancel={handleCancelEdit} />
  {/if}

  <!-- Delete Confirmation Modal -->
  {#if deleteTarget}
    <div class="fixed inset-0 z-50 bg-black/60 flex items-center justify-center">
      <div class="bg-ck-dark rounded-xl border border-gray-700 p-6 max-w-sm shadow-2xl">
        <h3 class="text-white font-semibold mb-2">Delete file?</h3>
        <p class="text-sm text-ck-dim mb-5">
          Are you sure you want to delete <span class="text-white font-medium">{deleteTarget.name}</span>? This action cannot be undone.
        </p>
        <div class="flex justify-end gap-3">
          <button
            onclick={() => deleteTarget = null}
            class="px-4 py-1.5 text-xs rounded-lg bg-ck-dark text-ck-dim border border-gray-700 hover:text-white transition-colors"
          >
            Cancel
          </button>
          <button
            onclick={executeDelete}
            class="px-4 py-1.5 text-xs rounded-lg bg-red-500/20 text-red-400 border border-red-500/30 hover:bg-red-500/30 transition-colors"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  {/if}
{/if}
