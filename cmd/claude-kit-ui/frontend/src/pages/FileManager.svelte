<script>
  import FileTree from '../components/FileTree.svelte'
  import FilePreview from '../components/FilePreview.svelte'
  import { currentProject } from '../stores/project.js'

  let selectedPath = $state('')
  let previewContent = $state('')
  let dragging = $state(false)
  let fileTree = $state([])
  let projectPath = $state('')

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
    <div class="px-3 py-2 border-b border-gray-800 text-xs font-semibold text-ck-gold uppercase tracking-wider">
      Explorer
    </div>
    <div class="flex-1 overflow-auto py-1">
      <FileTree nodes={fileTree} {selectedPath} onSelect={handleSelect} onOpen={handleOpen} />
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
