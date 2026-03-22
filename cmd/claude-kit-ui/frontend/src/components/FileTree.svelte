<script>
  import FileTree from './FileTree.svelte'

  let { nodes = [], selectedPath = '', onSelect = () => {}, onOpen = () => {} } = $props()
</script>

{#each nodes as node}
  {@const isDir = node.children?.length > 0}
  <div class="select-none">
    <button
      onclick={() => {
        if (isDir) {
          node.expanded = !node.expanded
        } else {
          onSelect(node.path)
        }
      }}
      ondblclick={() => { if (!isDir) onOpen(node.path) }}
      class="flex items-center gap-1.5 w-full px-2 py-1 text-sm rounded hover:bg-ck-pink/20 transition-colors
        {selectedPath === node.path ? 'bg-ck-rose/30 text-white' : 'text-ck-dim hover:text-white'}"
    >
      <span class="text-xs w-4 text-center shrink-0">
        {#if isDir}
          {node.expanded ? '▾' : '▸'}
        {/if}
      </span>
      <span class="shrink-0">{isDir ? (node.expanded ? '📂' : '📁') : fileIcon(node.name)}</span>
      <span class="truncate">{node.name}</span>
    </button>

    {#if isDir && node.expanded}
      <div class="ml-3 pl-2 border-l border-gray-700/50">
        <FileTree nodes={node.children} {selectedPath} {onSelect} {onOpen} />
      </div>
    {/if}
  </div>
{/each}

<script module>
  function fileIcon(name) {
    const ext = name.split('.').pop()?.toLowerCase()
    const icons = {
      md: '📄', yaml: '📊', yml: '📊', csv: '📋',
      xlsx: '📑', xls: '📑', pdf: '📎',
      js: '💻', ts: '💻', go: '💻', py: '💻', svelte: '💻',
      json: '💻', toml: '💻', sh: '💻',
    }
    return icons[ext] || '📄'
  }
</script>
