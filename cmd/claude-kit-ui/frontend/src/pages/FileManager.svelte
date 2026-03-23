<script>
  import { onMount, onDestroy } from 'svelte'
  import FileTree from '../components/FileTree.svelte'
  import FilePreview from '../components/FilePreview.svelte'
  import { currentProject } from '../stores/project.js'

  let selectedPath = $state('')
  let previewContent = $state('')
  let dragging = $state(false)
  let fileTree = $state([])
  let projectPath = $state('')
  let fileSearch = $state('')
  let searchOpen = $state(false)

  // Filter tree recursively
  function filterTree(nodes, query) {
    if (!query) return nodes
    const q = query.toLowerCase()
    return nodes.reduce((acc, node) => {
      if (node.name.toLowerCase().includes(q)) {
        acc.push(node)
      } else if (node.children) {
        const filtered = filterTree(node.children, query)
        if (filtered.length > 0) {
          acc.push({ ...node, children: filtered, expanded: true })
        }
      }
      return acc
    }, [])
  }

  let filteredTree = $derived(filterTree(fileTree, fileSearch))

  onMount(() => {
    const handler = () => {
      searchOpen = true
      setTimeout(() => {
        const input = document.querySelector('[data-file-search]')
        if (input) input.focus()
      }, 50)
    }
    window.addEventListener('ck:file-search', handler)
    return () => window.removeEventListener('ck:file-search', handler)
  })

  currentProject.subscribe(p => {
    projectPath = p?.path || ''
    if (p?.path) loadTree(p.path)
  })

  async function loadTree(path) {
    try {
      const { Tree } = await import('../../wailsjs/go/main/FileService.js')
      const tree = await Tree(path)
      fileTree = tree || []
    } catch {
      fileTree = []
    }
  }

  async function handleSelect(path) {
    selectedPath = path
    if (!projectPath) {
      previewContent = `// No project selected`
      return
    }
    try {
      const { ReadPreview } = await import('../../wailsjs/go/main/FileService.js')
      const content = await ReadPreview(projectPath, path)
      previewContent = content || `// No preview available for ${path}`
    } catch {
      previewContent = `// No preview available for ${path}`
    }
  }

  function handleOpen(path) {
    // TODO: OpenExternal with projectPath
    console.log('Open external:', path)
  }

  function handleDragOver(e) {
    e.preventDefault()
    dragging = true
  }

  function handleDragLeave() {
    dragging = false
  }

  function handleDrop(e) {
    e.preventDefault()
    dragging = false
    // When Wails bindings available: copy files via runtime
    const files = e.dataTransfer?.files
    if (files?.length) {
      console.log('Dropped files:', [...files].map(f => f.name))
    }
  }
</script>

<div class="flex h-full gap-0 rounded-lg overflow-hidden border border-gray-800">
  <!-- Tree panel -->
  <div
    class="w-64 shrink-0 bg-ck-dark border-r border-gray-800 flex flex-col relative"
    ondragover={handleDragOver}
    ondragleave={handleDragLeave}
    ondrop={handleDrop}
    role="tree"
    tabindex="0"
  >
    <div class="border-b border-gray-800">
      <div class="flex items-center justify-between px-3 py-2">
        <span class="text-xs font-semibold text-ck-gold uppercase tracking-wider">Explorer</span>
        <button
          onclick={() => { searchOpen = !searchOpen; if (!searchOpen) fileSearch = '' }}
          class="text-xs text-ck-dim hover:text-white transition-colors"
          title="Search files (⌘F)"
        >🔍</button>
      </div>
      {#if searchOpen}
        <div class="px-2 pb-2">
          <input
            type="text"
            data-file-search
            placeholder="Filter files..."
            bind:value={fileSearch}
            class="w-full px-2 py-1 text-xs bg-ck-bg border border-gray-700 rounded text-white
              placeholder-ck-dim focus:border-ck-pink focus:outline-none"
          />
        </div>
      {/if}
    </div>
    <div class="flex-1 overflow-auto py-1">
      <FileTree nodes={filteredTree} {selectedPath} onSelect={handleSelect} onOpen={handleOpen} />
    </div>

    {#if dragging}
      <div class="absolute inset-0 bg-ck-rose/20 border-2 border-dashed border-ck-rose rounded-lg flex items-center justify-center z-10 pointer-events-none">
        <span class="text-sm text-ck-pink font-medium">Drop files here</span>
      </div>
    {/if}
  </div>

  <!-- Preview panel -->
  <div class="flex-1 bg-ck-bg min-w-0">
    <FilePreview path={selectedPath} content={previewContent} />
  </div>
</div>
