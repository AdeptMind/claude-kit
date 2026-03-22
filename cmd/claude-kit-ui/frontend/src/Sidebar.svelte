<script>
  import { currentPage, navigateTo, currentRole, isManagementRole } from './stores/navigation.js'

  const managementItems = [
    { id: 'dashboard', icon: '\u{1F4CA}', label: 'Dashboard' },
    { id: 'stories', icon: '\u{1F4CB}', label: 'Stories' },
    { id: 'files', icon: '\u{1F4C1}', label: 'Files' },
    { id: 'profile', icon: '\u{1F464}', label: 'Profile' },
    { id: 'workflow', icon: '\u{1F504}', label: 'Workflow' },
    { id: 'settings', icon: '\u2699\uFE0F', label: 'Settings' },
  ]

  const devItems = [
    { id: 'dashboard', icon: '\u{1F4CA}', label: 'Dashboard' },
    { id: 'code', icon: '\u{1F4BB}', label: 'Code' },
    { id: 'tests', icon: '\u{1F9EA}', label: 'Tests' },
    { id: 'architecture', icon: '\u{1F3D7}', label: 'Architecture' },
    { id: 'settings', icon: '\u2699\uFE0F', label: 'Settings' },
  ]

  let expanded = $state(true)
  let active = $state('dashboard')
  let role = $state('po')

  currentPage.subscribe(v => active = v)
  currentRole.subscribe(v => role = v)

  let navItems = $derived(isManagementRole(role) ? managementItems : devItems)

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

  <div class="flex items-center gap-2 px-4 py-2 border-b border-gray-800">
    <span class="text-xs font-mono text-ck-gold truncate">{role}</span>
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
</aside>
