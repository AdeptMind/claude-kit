<script>
  import FileTree from '../components/FileTree.svelte'
  import FilePreview from '../components/FilePreview.svelte'

  // Placeholder data until Wails bindings are available
  // import { Tree, ReadPreview, OpenExternal } from '../../wailsjs/go/main/FileService.js'

  let selectedPath = $state('')
  let previewContent = $state('')
  let dragging = $state(false)

  const placeholderTree = [
    {
      name: '.claude', path: '.claude', expanded: true, children: [
        { name: 'CLAUDE.md', path: '.claude/CLAUDE.md' },
        {
          name: 'agents', path: '.claude/agents', expanded: false, children: [
            { name: 'ralph.md', path: '.claude/agents/ralph.md' },
            { name: 'architect.md', path: '.claude/agents/architect.md' },
          ]
        },
        {
          name: 'skills', path: '.claude/skills', expanded: false, children: [
            { name: 'brainstorm', path: '.claude/skills/brainstorm', expanded: false, children: [
              { name: 'SKILL.md', path: '.claude/skills/brainstorm/SKILL.md' },
            ]},
          ]
        },
        {
          name: 'output', path: '.claude/output', expanded: false, children: [
            { name: 'problem.yaml', path: '.claude/output/problem.yaml' },
            { name: 'architecture.yaml', path: '.claude/output/architecture.yaml' },
            { name: 'backlog.yaml', path: '.claude/output/backlog.yaml' },
          ]
        },
      ]
    },
    {
      name: 'cmd', path: 'cmd', expanded: false, children: [
        { name: 'main.go', path: 'cmd/main.go' },
      ]
    },
    { name: 'go.mod', path: 'go.mod' },
    { name: 'README.md', path: 'README.md' },
  ]

  const placeholderContents = {
    '.claude/CLAUDE.md': '# BMAD Project Template\n\nSee workflow commands below...',
    '.claude/agents/ralph.md': '---\nname: ralph\nrole: Full-stack implementer\n---\n\n# Ralph Agent\n\nImplements stories from backlog.',
    '.claude/output/problem.yaml': 'name: claude-kit-ui\nversion: 1.0\nfeatures:\n  - file-manager\n  - dashboard\n  - workflow-launcher',
    'go.mod': 'module github.com/adrien-barret/claude-kit\n\ngo 1.23',
    'README.md': '# Claude Kit\n\nBMAD workflow orchestration toolkit.',
  }

  function handleSelect(path) {
    selectedPath = path
    // When Wails bindings available: previewContent = await ReadPreview(path)
    previewContent = placeholderContents[path] || `// No preview available for ${path}`
  }

  function handleOpen(path) {
    // When Wails bindings available: OpenExternal(path)
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
      <FileTree nodes={placeholderTree} {selectedPath} onSelect={handleSelect} onOpen={handleOpen} />
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
