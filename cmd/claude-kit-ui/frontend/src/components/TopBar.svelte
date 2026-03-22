<script>
  import { currentProject, recentProjects, setProject, loadRecents } from '../stores/project.js'

  let project = $state(null)
  let recents = $state([])
  let dropdownOpen = $state(false)
  let manualInput = $state(false)
  let manualPath = $state('')

  currentProject.subscribe(v => project = v)
  recentProjects.subscribe(v => recents = v)

  $effect(() => { loadRecents() })

  function selectProject(p) {
    setProject(p)
    dropdownOpen = false
  }

  async function openFolderPicker() {
    try {
      const { OpenDirectoryDialog } = await import('../../wailsjs/runtime/runtime.js')
      const path = await OpenDirectoryDialog({ title: 'Select Project Folder' })
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

  function handleKeydown(e) {
    if (e.key === 'Escape') {
      dropdownOpen = false
      manualInput = false
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

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
          onclick={() => { dropdownOpen = false; manualInput = false }}
          aria-label="Close dropdown"
        ></button>

        <div class="absolute top-full mt-1 left-1/2 -translate-x-1/2 w-72 bg-ck-dark border border-gray-700 rounded-lg shadow-xl z-50 overflow-hidden">
          {#if recents.length > 0}
            <div class="max-h-60 overflow-y-auto">
              {#each recents as p}
                <button
                  onclick={() => selectProject(p)}
                  class="w-full text-left px-3 py-2 hover:bg-white/5 transition-colors
                    {project?.path === p.path ? 'bg-ck-rose/10 border-l-2 border-ck-rose' : 'border-l-2 border-transparent'}"
                >
                  <div class="text-sm text-white font-medium truncate">{p.name}</div>
                  <div class="text-xs text-ck-dim truncate">{p.path}</div>
                </button>
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

  <!-- Right: Version badge -->
  <span class="text-xs text-ck-gold font-mono select-none">v0.5.0</span>
</div>
