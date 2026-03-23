<script>
  import { currentProject, recentProjects, setProject, removeProject, loadRecents, projectFinderOpen } from '../stores/project.js'
  import { navigateTo } from '../stores/navigation.js'

  let project = $state(null)
  let recents = $state([])
  let dropdownOpen = $state(false)
  let manualInput = $state(false)
  let manualPath = $state('')
  let projectSearch = $state('')

  currentProject.subscribe(v => project = v)
  recentProjects.subscribe(v => recents = v)

  $effect(() => { loadRecents() })

  projectFinderOpen.subscribe(v => {
    if (v) {
      dropdownOpen = true
      setTimeout(() => {
        const input = document.querySelector('[data-project-search]')
        if (input) input.focus()
      }, 50)
      projectFinderOpen.set(false)
    }
  })

  function selectProject(p) {
    setProject(p)
    dropdownOpen = false
    projectSearch = ''
  }

  async function openFolderPicker() {
    try {
      const { OpenFolderDialog } = await import('../../wailsjs/go/main/App.js')
      const path = await OpenFolderDialog()
      if (path) {
        const name = path.split('/').filter(Boolean).pop() || path
        setProject({ name, path })
        dropdownOpen = false
      }
    } catch {
      manualInput = true
    }
  }

  function submitManualPath() {
    if (!manualPath.trim()) return
    const path = manualPath.trim()
    const name = path.split('/').filter(Boolean).pop() || path
    setProject({ name, path })
    manualInput = false
    manualPath = ''
    dropdownOpen = false
  }

  function handleCloseModal() {
    if (dropdownOpen) {
      dropdownOpen = false
      manualInput = false
      projectSearch = ''
    }
  }

  $effect(() => {
    window.addEventListener('ck:close-modal', handleCloseModal)
    return () => window.removeEventListener('ck:close-modal', handleCloseModal)
  })
</script>

<div class="h-10 bg-ck-dark border-b border-gray-800 flex items-center px-4 shrink-0 relative z-50">
  <!-- Left: App icon -->
  <span class="text-sm font-bold text-ck-pink tracking-tight select-none">CK</span>

  <!-- Center: Project selector -->
  <div class="flex-1 flex justify-center">
    <div class="relative">
      <button
        onclick={() => dropdownOpen = !dropdownOpen}
        class="flex items-center gap-1.5 px-3 py-1 rounded-md text-sm font-semibold text-white
          hover:bg-white/5 transition-colors"
      >
        {project?.name || 'No project selected'}
        <svg class="w-3 h-3 text-ck-dim transition-transform {dropdownOpen ? 'rotate-180' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {#if dropdownOpen}
        <!-- Backdrop -->
        <button
          class="fixed inset-0 z-40 cursor-default"
          onclick={() => { dropdownOpen = false; manualInput = false; projectSearch = '' }}
          aria-label="Close dropdown"
        ></button>

        <div class="absolute top-full mt-1 left-1/2 -translate-x-1/2 w-72 bg-ck-dark border border-gray-700 rounded-lg shadow-xl z-50 overflow-hidden">
          <!-- Search input -->
          <div class="p-2 border-b border-gray-700">
            <input
              type="text"
              data-project-search
              placeholder="Search projects..."
              bind:value={projectSearch}
              class="w-full px-3 py-1.5 text-sm bg-ck-bg border border-gray-700 rounded text-white
                placeholder-ck-dim focus:border-ck-pink focus:outline-none"
            />
          </div>
          {#if recents.length > 0}
            <div class="max-h-60 overflow-y-auto">
              {#each recents.filter(p => !projectSearch || p.name.toLowerCase().includes(projectSearch.toLowerCase()) || p.path.toLowerCase().includes(projectSearch.toLowerCase())) as p}
                <div
                  class="group flex items-center hover:bg-white/5 transition-colors
                    {project?.path === p.path ? 'bg-ck-rose/10 border-l-2 border-ck-rose' : 'border-l-2 border-transparent'}"
                >
                  <button
                    onclick={() => selectProject(p)}
                    class="flex-1 text-left px-3 py-2 min-w-0"
                    title={p.path}
                  >
                    <div class="text-sm text-white font-medium truncate">{p.name}</div>
                    <div class="text-xs text-ck-dim truncate">{p.path}</div>
                  </button>
                  <button
                    onclick={(e) => { e.stopPropagation(); removeProject(p.path) }}
                    class="shrink-0 w-6 h-6 mr-2 flex items-center justify-center rounded-full
                      text-red-400/0 group-hover:text-red-400/50 hover:!text-red-400 hover:bg-red-400/10
                      transition-all text-xs font-bold"
                    title="Remove from list"
                  >
                    −
                  </button>
                </div>
              {/each}
            </div>
          {:else}
            <div class="px-3 py-4 text-center text-xs text-ck-dim">No recent projects</div>
          {/if}

          <div class="border-t border-gray-700 p-2">
            {#if manualInput}
              <form onsubmit={(e) => { e.preventDefault(); submitManualPath() }} class="flex gap-1.5">
                <input
                  type="text"
                  bind:value={manualPath}
                  placeholder="/path/to/project"
                  class="flex-1 bg-ck-bg border border-gray-700 rounded px-2 py-1 text-xs text-white
                    placeholder-ck-dim focus:border-ck-pink focus:outline-none"
                />
                <button type="submit" class="px-2 py-1 bg-ck-rose rounded text-white text-xs font-medium hover:bg-ck-pink transition-colors">
                  OK
                </button>
              </form>
            {:else}
              <button
                onclick={openFolderPicker}
                class="w-full flex items-center justify-center gap-1.5 px-3 py-1.5 bg-ck-rose rounded-md text-white text-xs font-medium
                  hover:bg-ck-pink transition-colors"
              >
                <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path d="M12 4v16m8-8H4" />
                </svg>
                Open project
              </button>
            {/if}
          </div>
        </div>
      {/if}
    </div>
  </div>

  <!-- Right: Chat button + Version -->
  <div class="flex items-center gap-3">
    <button
      onclick={() => navigateTo('chat')}
      class="flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium transition-colors
        text-ck-dim hover:text-white hover:bg-white/5"
      title="Open Claude Chat"
    >
      💬 Chat
    </button>
    <span class="text-xs text-ck-gold font-mono select-none">v0.5.0</span>
  </div>
</div>
