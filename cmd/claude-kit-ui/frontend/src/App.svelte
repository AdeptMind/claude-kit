<script>
  import { currentPage } from './stores/navigation.js'
  import Sidebar from './Sidebar.svelte'
  import Dashboard from './pages/Dashboard.svelte'
  import FileManager from './pages/FileManager.svelte'
  import ProfileEditor from './pages/ProfileEditor.svelte'
  import WorkflowLauncher from './pages/WorkflowLauncher.svelte'
  import Settings from './pages/Settings.svelte'

  const pages = {
    dashboard: Dashboard,
    files: FileManager,
    profile: ProfileEditor,
    workflow: WorkflowLauncher,
    settings: Settings,
  }

  let activePage = $state('dashboard')
  currentPage.subscribe(v => activePage = v)
</script>

<main class="flex h-screen bg-ck-bg text-white font-sans">
  <Sidebar />
  <div class="flex-1 overflow-auto p-6">
    {#each Object.entries(pages) as [id, Component]}
      {#if activePage === id}
        <Component />
      {/if}
    {/each}
  </div>
</main>
