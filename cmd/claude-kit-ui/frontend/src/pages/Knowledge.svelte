<script>
  import MarkdownPreview from '../components/MarkdownPreview.svelte'

  let nodes = $state([])
  let searchResults = $state(null)
  let loading = $state(true)
  let query = $state('')
  let semantic = $state(false)
  let hasEmbedder = $state(false)
  let bannerDismissed = $state(loadBannerState())
  let selectedNode = $state(null)
  let selectedEdges = $state([])
  let breadcrumb = $state([])
  let stats = $state(null)
  let focusedIndex = $state(-1)
  let debounceTimer = null

  // Type colors
  const typeColors = {
    decision: { bg: 'bg-purple-500/20', text: 'text-purple-400', dot: '🟣' },
    pattern: { bg: 'bg-blue-500/20', text: 'text-blue-400', dot: '🔵' },
    concept: { bg: 'bg-ck-green/20', text: 'text-ck-green', dot: '🟢' },
    bug: { bg: 'bg-red-500/20', text: 'text-red-400', dot: '🔴' },
    note: { bg: 'bg-gray-500/20', text: 'text-gray-400', dot: '⚪' },
    reference: { bg: 'bg-ck-gold/20', text: 'text-ck-gold', dot: '🟡' },
  }

  $effect(() => {
    loadNodes()
    checkEmbedder()
  })

  async function loadNodes() {
    try {
      const svc = await import('../../wailsjs/go/main/KnowledgeService.js')
      const list = await svc.ListNodes(50, 0)
      nodes = list || []
      stats = await svc.GetStats()
    } catch (e) {
      nodes = []
    }
    loading = false
  }

  async function checkEmbedder() {
    try {
      const svc = await import('../../wailsjs/go/main/KnowledgeService.js')
      hasEmbedder = await svc.HasEmbedder()
    } catch { hasEmbedder = false }
  }

  function handleSearch() {
    clearTimeout(debounceTimer)
    debounceTimer = setTimeout(async () => {
      if (!query.trim()) {
        searchResults = null
        return
      }
      try {
        const svc = await import('../../wailsjs/go/main/KnowledgeService.js')
        const results = await svc.SearchNodes(query, semantic, 20)
        searchResults = results || []
      } catch { searchResults = [] }
    }, 300)
  }

  function displayNodes() {
    if (searchResults !== null) return searchResults.map(r => ({
      id: r.node.id,
      title: r.node.title,
      type: r.node.type,
      dimensions: r.node.dimensions || [],
      createdAt: r.node.createdAt,
      score: r.score,
      matchType: r.matchType,
    }))
    return nodes
  }

  async function selectNode(node) {
    try {
      const svc = await import('../../wailsjs/go/main/KnowledgeService.js')
      selectedNode = await svc.GetNode(node.id)
      const edges = await svc.GetEdges(node.id)
      selectedEdges = edges || []
      // Add to breadcrumb
      const existing = breadcrumb.findIndex(b => b.id === node.id)
      if (existing >= 0) {
        breadcrumb = breadcrumb.slice(0, existing + 1)
      } else {
        breadcrumb = [...breadcrumb.slice(-4), { id: node.id, title: node.title }]
      }
      focusedIndex = -1
    } catch {}
  }

  async function navigateToConnected(nodeId) {
    try {
      const svc = await import('../../wailsjs/go/main/KnowledgeService.js')
      const node = await svc.GetNode(nodeId)
      if (node) await selectNode(node)
    } catch {}
  }

  function closeDetail() {
    selectedNode = null
    selectedEdges = []
    breadcrumb = []
  }

  function handleKeydown(e) {
    const items = displayNodes()
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      focusedIndex = Math.min(focusedIndex + 1, items.length - 1)
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      focusedIndex = Math.max(focusedIndex - 1, 0)
    } else if (e.key === 'Enter' && focusedIndex >= 0) {
      e.preventDefault()
      selectNode(items[focusedIndex])
    } else if (e.key === 'Escape') {
      if (selectedNode) {
        closeDetail()
      } else if (query) {
        query = ''
        searchResults = null
      }
    }
  }

  function relativeDate(dateStr) {
    if (!dateStr) return ''
    const d = new Date(dateStr)
    const now = new Date()
    const diff = now - d
    const mins = Math.floor(diff / 60000)
    if (mins < 60) return `${mins}m ago`
    const hours = Math.floor(mins / 60)
    if (hours < 24) return `${hours}h ago`
    const days = Math.floor(hours / 24)
    if (days < 7) return `${days}d ago`
    const weeks = Math.floor(days / 7)
    return `${weeks}w ago`
  }

  function loadBannerState() {
    try { return localStorage.getItem('ck-kg-banner-dismissed') === 'true' } catch { return false }
  }
  function dismissBanner() {
    bannerDismissed = true
    localStorage.setItem('ck-kg-banner-dismissed', 'true')
  }

  function getTypeStyle(type) {
    return typeColors[type] || typeColors.note
  }
</script>

<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<div class="h-full flex flex-col" onkeydown={handleKeydown} tabindex="0">
  <!-- Search bar -->
  <div class="flex items-center gap-3 mb-4">
    <div class="flex-1 relative">
      <input
        type="text"
        placeholder="Search knowledge..."
        bind:value={query}
        oninput={handleSearch}
        class="w-full px-4 py-2.5 bg-ck-dark border border-gray-700 rounded-lg text-white placeholder-ck-dim focus:border-ck-pink focus:outline-none"
      />
    </div>
    <div class="flex items-center gap-2 bg-ck-dark border border-gray-700 rounded-lg px-1 py-1">
      <button
        onclick={() => { semantic = false; handleSearch() }}
        class="px-3 py-1.5 rounded text-xs font-medium transition-colors
          {!semantic ? 'bg-ck-rose text-white' : 'text-ck-dim hover:text-white'}"
      >Keyword</button>
      <button
        onclick={() => { if (hasEmbedder) { semantic = true; handleSearch() } }}
        disabled={!hasEmbedder}
        title={hasEmbedder ? 'AI-powered semantic search' : 'Install Ollama for semantic search'}
        class="px-3 py-1.5 rounded text-xs font-medium transition-colors
          {semantic ? 'bg-ck-gold text-black' : hasEmbedder ? 'text-ck-dim hover:text-white' : 'text-gray-600 cursor-not-allowed'}"
      >Semantic</button>
    </div>
  </div>

  <!-- Embedder banner -->
  {#if !hasEmbedder && !bannerDismissed}
    <div class="flex items-center gap-2 px-4 py-2 mb-4 bg-ck-gold/10 border border-ck-gold/30 rounded-lg text-sm text-ck-gold">
      <span>Semantic search unavailable — install Ollama for AI-powered search</span>
      <button onclick={dismissBanner} class="ml-auto text-ck-dim hover:text-white">✕</button>
    </div>
  {/if}

  <!-- Content area -->
  <div class="flex-1 flex overflow-hidden gap-4">
    <!-- Node list -->
    <div class="overflow-y-auto {selectedNode ? 'w-2/5' : 'w-full'} transition-all duration-200"
         style={selectedNode ? '' : ''}
         role="listbox" aria-label="Knowledge nodes">
      {#if loading}
        {#each [1,2,3] as _}
          <div class="p-4 mb-2 bg-ck-dark rounded-lg animate-pulse">
            <div class="h-4 bg-gray-700 rounded w-3/4 mb-2"></div>
            <div class="h-3 bg-gray-800 rounded w-1/2"></div>
          </div>
        {/each}
      {:else if displayNodes().length === 0}
        <div class="text-center py-12 text-ck-dim">
          {#if query}
            <p>No results for "{query}"</p>
            {#if !semantic && hasEmbedder}
              <p class="text-sm mt-1">Try semantic search for AI-powered results</p>
            {/if}
          {:else}
            <p>No knowledge yet.</p>
            <p class="text-sm mt-2">Agents will capture knowledge as you work,<br/>or add nodes with <code class="text-ck-pink">ck knowledge add</code></p>
          {/if}
        </div>
      {:else}
        {#each displayNodes() as node, i}
          {@const style = getTypeStyle(node.type)}
          <button
            onclick={() => selectNode(node)}
            role="option"
            aria-selected={selectedNode?.id === node.id}
            class="w-full text-left p-3 mb-1.5 rounded-lg transition-colors
              {focusedIndex === i ? 'bg-ck-pink/20 border border-ck-pink/40' :
               selectedNode?.id === node.id ? 'bg-ck-rose/20 border border-ck-rose/30' :
               'bg-ck-dark border border-transparent hover:bg-ck-pink/10 hover:border-gray-700'}"
          >
            <div class="flex items-center gap-2 mb-1">
              <span class="text-xs">{style.dot}</span>
              <span class="text-sm font-medium text-white truncate">{node.title}</span>
            </div>
            <div class="flex items-center gap-2 text-xs text-ck-dim">
              <span class="{style.bg} {style.text} px-1.5 py-0.5 rounded text-[10px]">{node.type}</span>
              {#if node.dimensions?.length > 0}
                {#each node.dimensions.slice(0, 3) as dim}
                  <span class="text-ck-dim">#{dim}</span>
                {/each}
              {/if}
              <span class="ml-auto">{relativeDate(node.createdAt)}</span>
            </div>
          </button>
        {/each}
      {/if}

      <!-- Stats footer -->
      {#if stats && !loading}
        <div class="text-[10px] text-ck-dim text-center py-2 mt-2">
          {stats.nodes} nodes · {stats.nodesWithEmbedding} embedded
        </div>
      {/if}
    </div>

    <!-- Detail panel -->
    {#if selectedNode}
      <div
        class="w-3/5 bg-ck-dark rounded-lg border border-gray-800 overflow-y-auto p-5"
        role="complementary"
        aria-label="Node detail"
        style="animation: slideIn 0.2s ease-out"
      >
        <!-- Breadcrumb -->
        {#if breadcrumb.length > 1}
          <div class="flex items-center gap-1 mb-3 text-xs text-ck-dim overflow-x-auto">
            {#each breadcrumb as crumb, i}
              {#if i > 0}<span>›</span>{/if}
              <button
                onclick={() => navigateToConnected(crumb.id)}
                class="hover:text-ck-pink truncate max-w-[120px]
                  {crumb.id === selectedNode.id ? 'text-white' : ''}"
              >{crumb.title}</button>
            {/each}
          </div>
        {/if}

        <!-- Header -->
        <div class="flex items-start justify-between mb-4">
          <div>
            <h2 class="text-lg font-bold text-white">{selectedNode.title}</h2>
            <div class="flex items-center gap-2 mt-1 text-xs text-ck-dim">
              {@const style = getTypeStyle(selectedNode.type)}
              <span class="{style.bg} {style.text} px-2 py-0.5 rounded">{selectedNode.type}</span>
              <span>{selectedNode.id}</span>
              <span>{relativeDate(selectedNode.createdAt)}</span>
            </div>
          </div>
          <button onclick={closeDetail} class="text-ck-dim hover:text-white text-lg">✕</button>
        </div>

        <!-- Content -->
        {#if selectedNode.content}
          <div class="prose prose-invert prose-sm max-w-none mb-6">
            <MarkdownPreview content={selectedNode.content} />
          </div>
        {/if}

        <!-- Connections -->
        {#if selectedEdges.length > 0}
          <div>
            <h3 class="text-sm font-semibold text-ck-dim uppercase tracking-wider mb-2">Connections</h3>
            {#each selectedEdges as edge}
              {@const isOutgoing = edge.fromNodeId === selectedNode.id}
              {@const otherNodeId = isOutgoing ? edge.toNodeId : edge.fromNodeId}
              <button
                onclick={() => navigateToConnected(otherNodeId)}
                class="w-full text-left flex items-center gap-2 p-2 rounded hover:bg-ck-pink/10 transition-colors"
                aria-label="Connected to {otherNodeId} via {edge.type}"
              >
                <span class="text-ck-dim text-xs">{isOutgoing ? '→' : '←'}</span>
                <span class="bg-gray-700/50 text-ck-dim px-1.5 py-0.5 rounded text-[10px]">{edge.type}</span>
                <span class="text-sm text-white truncate">{otherNodeId}</span>
                {#if edge.explanation}
                  <span class="text-xs text-ck-dim truncate ml-auto">— {edge.explanation}</span>
                {/if}
              </button>
            {/each}
          </div>
        {:else}
          <p class="text-xs text-ck-dim">No connections yet</p>
        {/if}
      </div>
    {/if}
  </div>
</div>

<style>
  @keyframes slideIn {
    from { opacity: 0; transform: translateX(20px); }
    to { opacity: 1; transform: translateX(0); }
  }
  @media (prefers-reduced-motion: reduce) {
    @keyframes slideIn {
      from { opacity: 1; transform: none; }
      to { opacity: 1; transform: none; }
    }
  }
</style>
