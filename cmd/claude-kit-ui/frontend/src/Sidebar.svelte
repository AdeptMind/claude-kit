<script>
  import { currentPage, navigateTo, currentRole, setRole, isManagementRole, refreshRole } from './stores/navigation.js'
  import { currentProject } from './stores/project.js'

  let projectPath = $state('')
  currentProject.subscribe(p => {
    projectPath = p?.path || ''
    if (p?.path) refreshRole(p.path)
  })

  const managementItems = [
    { id: 'dashboard', icon: '\u{1F4CA}', label: 'Dashboard' },
    { id: 'stories', icon: '\u{1F4CB}', label: 'Stories' },
    { id: 'files', icon: '\u{1F4C1}', label: 'Files' },
    { id: 'profile', icon: '\u{1F464}', label: 'Profile' },
    { id: 'workflow', icon: '\u{1F504}', label: 'Workflow' },
    { id: 'knowledge', icon: '\u{1F9E0}', label: 'Knowledge' },
    { id: 'chat', icon: '\u{1F4AC}', label: 'Chat' },
    { id: 'settings', icon: '\u2699\uFE0F', label: 'Settings' },
  ]

  const devItems = [
    { id: 'dashboard', icon: '\u{1F4CA}', label: 'Dashboard' },
    { id: 'code', icon: '\u{1F4BB}', label: 'Code' },
    { id: 'tests', icon: '\u{1F9EA}', label: 'Tests' },
    { id: 'architecture', icon: '\u{1F3D7}', label: 'Architecture' },
    { id: 'chat', icon: '\u{1F4AC}', label: 'Chat' },
    { id: 'settings', icon: '\u2699\uFE0F', label: 'Settings' },
  ]

  const allRoles = [
    { group: 'Management', roles: ['po', 'business-analyst', 'project-manager', 'scrum-master', 'people-ops', 'change-manager', 'architect', 'ux-designer', 'finops', 'soc2-compliance', 'access-review', 'data-governance'] },
    { group: 'Dev', roles: ['dev', 'backend', 'frontend', 'devops', 'security', 'sre', 'golang', 'python', 'typescript'] },
    { group: 'Special', roles: ['all'] },
  ]

  let expanded = $state(true)
  let active = $state('dashboard')
  let role = $state('po')
  let roleSelectorOpen = $state(false)
  let searchQuery = $state('')
  let favorites = $state(loadFavorites())

  currentPage.subscribe(v => active = v)
  currentRole.subscribe(v => role = v)

  let navItems = $derived(isManagementRole(role) ? managementItems : devItems)

  // Filtered roles based on search
  let filteredRoles = $derived(() => {
    const q = searchQuery.toLowerCase().trim()
    if (!q) return null // null = show normal groups
    const matches = []
    for (const group of allRoles) {
      for (const r of group.roles) {
        if (r.includes(q)) matches.push(r)
      }
    }
    return matches
  })

  // Favorite roles sorted at top
  let favoriteRoles = $derived(
    allRoles.flatMap(g => g.roles).filter(r => favorites.has(r))
  )

  function selectRole(newRole) {
    setRole(newRole, projectPath)
    roleSelectorOpen = false
    searchQuery = ''
  }

  function toggleFavorite(e, r) {
    e.stopPropagation()
    const next = new Set(favorites)
    if (next.has(r)) {
      next.delete(r)
    } else {
      next.add(r)
    }
    favorites = next
    saveFavorites(next)
  }

  function loadFavorites() {
    try {
      const stored = localStorage.getItem('ck-role-favorites')
      return stored ? new Set(JSON.parse(stored)) : new Set()
    } catch { return new Set() }
  }

  function saveFavorites(favs) {
    localStorage.setItem('ck-role-favorites', JSON.stringify([...favs]))
  }

  function openSelector() {
    roleSelectorOpen = !roleSelectorOpen
    searchQuery = ''
  }

  function handleResize() {
    expanded = window.innerWidth >= 768
  }

  $effect(() => {
    handleResize()
    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  })
</script>

<aside class="flex flex-col h-full bg-ck-dark border-r border-gray-800 transition-all duration-200 {expanded ? 'w-48' : 'w-16'}">
  <div class="flex items-center gap-2 px-4 py-4 border-b border-gray-800">
    <span class="text-ck-pink font-bold text-lg">CK</span>
    {#if expanded}
      <span class="text-ck-gold text-xs font-mono">ui</span>
    {/if}
  </div>

  <!-- Role indicator — click to open selector -->
  <div class="relative">
    <button
      onclick={openSelector}
      class="w-full flex items-center gap-2 px-4 py-2.5 border-b border-gray-800 hover:bg-ck-pink/10 transition-colors cursor-pointer"
    >
      {#if favorites.has(role)}
        <span class="text-ck-gold text-xs">★</span>
      {/if}
      <span class="text-sm font-semibold text-ck-gold truncate">{role}</span>
      {#if expanded}
        <span class="text-ck-dim text-xs ml-auto">{roleSelectorOpen ? '▲' : '▼'}</span>
      {/if}
    </button>

    <!-- Role selector dropdown -->
    {#if roleSelectorOpen}
      <div class="absolute left-0 top-full z-50 w-60 max-h-96 flex flex-col bg-ck-dark border border-gray-700 rounded-lg shadow-xl">

        <!-- Search input -->
        <div class="p-2 border-b border-gray-800">
          <!-- svelte-ignore a11y_autofocus -->
          <input
            type="text"
            placeholder="Search roles..."
            bind:value={searchQuery}
            autofocus
            class="w-full px-3 py-1.5 text-sm bg-ck-bg border border-gray-700 rounded text-white placeholder-ck-dim focus:border-ck-pink focus:outline-none"
          />
        </div>

        <div class="overflow-y-auto flex-1">

          <!-- Search results mode -->
          {#if filteredRoles() !== null}
            {#each filteredRoles() as r}
              <button
                onclick={() => selectRole(r)}
                class="w-full text-left px-4 py-2 text-sm flex items-center gap-2 transition-colors
                  {r === role ? 'bg-ck-rose text-white' : 'text-gray-300 hover:bg-ck-pink/20 hover:text-white'}"
              >
                <span role="button" tabindex="0" onclick={(e) => toggleFavorite(e, r)} onkeydown={(e) => e.key === 'Enter' && toggleFavorite(e, r)} class="text-base cursor-pointer hover:scale-125 transition-transform {favorites.has(r) ? 'text-ck-gold' : 'text-gray-600'}">
                  {favorites.has(r) ? '★' : '☆'}
                </span>
                <span>{r}</span>
              </button>
            {/each}
            {#if filteredRoles().length === 0}
              <div class="px-4 py-3 text-sm text-ck-dim">No roles match "{searchQuery}"</div>
            {/if}

          <!-- Normal grouped mode -->
          {:else}

            <!-- Favorites section (if any) -->
            {#if favoriteRoles.length > 0}
              <div class="px-3 py-1.5 text-[10px] uppercase tracking-wider text-ck-gold border-b border-gray-800">
                ★ Favorites
              </div>
              {#each favoriteRoles as r}
                <button
                  onclick={() => selectRole(r)}
                  class="w-full text-left px-4 py-2 text-sm flex items-center gap-2 transition-colors
                    {r === role ? 'bg-ck-rose text-white' : 'text-gray-300 hover:bg-ck-pink/20 hover:text-white'}"
                >
                  <span role="button" tabindex="0" onclick={(e) => toggleFavorite(e, r)} onkeydown={(e) => e.key === 'Enter' && toggleFavorite(e, r)} class="text-base text-ck-gold cursor-pointer hover:scale-125 transition-transform">★</span>
                  <span>{r}</span>
                </button>
              {/each}
            {/if}

            <!-- All groups -->
            {#each allRoles as group}
              <div class="px-3 py-1.5 text-[10px] uppercase tracking-wider text-ck-dim border-b border-gray-800">
                {group.group}
              </div>
              {#each group.roles as r}
                <button
                  onclick={() => selectRole(r)}
                  class="w-full text-left px-4 py-2 text-sm flex items-center gap-2 transition-colors
                    {r === role ? 'bg-ck-rose text-white' : 'text-gray-300 hover:bg-ck-pink/20 hover:text-white'}"
                >
                  <span role="button" tabindex="0" onclick={(e) => toggleFavorite(e, r)} onkeydown={(e) => e.key === 'Enter' && toggleFavorite(e, r)} class="text-base cursor-pointer hover:scale-125 transition-transform {favorites.has(r) ? 'text-ck-gold' : 'text-gray-600 hover:text-gray-400'}">
                    {favorites.has(r) ? '★' : '☆'}
                  </span>
                  <span>{r}</span>
                </button>
              {/each}
            {/each}
          {/if}
        </div>
      </div>
    {/if}
  </div>

  <nav class="flex-1 flex flex-col gap-1 p-2">
    {#each navItems as item}
      <button
        onclick={() => navigateTo(item.id)}
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-colors
          {active === item.id ? 'bg-ck-rose text-white' : 'text-ck-dim hover:bg-ck-pink/20 hover:text-white'}"
      >
        <span class="text-lg">{item.icon}</span>
        {#if expanded}
          <span>{item.label}</span>
        {/if}
      </button>
    {/each}
  </nav>

  {#if expanded}
    <div class="px-3 py-2 border-t border-gray-800">
      <span class="text-[10px] text-ck-dim leading-relaxed">
        <kbd class="font-mono">&#8984;K</kbd> search
        <span class="mx-0.5">&middot;</span>
        <kbd class="font-mono">&#8984;J</kbd> chat
        <span class="mx-0.5">&middot;</span>
        <kbd class="font-mono">&#8984;S</kbd> stories
      </span>
    </div>
  {/if}
</aside>
