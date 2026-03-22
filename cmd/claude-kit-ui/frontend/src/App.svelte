<script>
  import { currentPage } from './stores/navigation.js'
  import TopBar from './components/TopBar.svelte'
  import Sidebar from './Sidebar.svelte'
  import Dashboard from './pages/Dashboard.svelte'
  import FileManager from './pages/FileManager.svelte'
  import ProfileEditor from './pages/ProfileEditor.svelte'
  import WorkflowLauncher from './pages/WorkflowLauncher.svelte'
  import Stories from './pages/Stories.svelte'
  import Settings from './pages/Settings.svelte'
  import Chat from './pages/Chat.svelte'

  const pages = {
    dashboard: Dashboard,
    stories: Stories,
    files: FileManager,
    profile: ProfileEditor,
    workflow: WorkflowLauncher,
    chat: Chat,
    settings: Settings,
  }

  let activePage = $state('dashboard')
  currentPage.subscribe(v => activePage = v)
</script>

<div class="flex flex-col h-screen bg-ck-bg text-white font-sans">
  <TopBar />
  <div class="flex flex-1 overflow-hidden">
    <Sidebar />
    <main class="flex-1 overflow-auto {activePage === 'chat' ? '' : 'p-6'}">
      {#each Object.entries(pages) as [id, Component]}
        {#if activePage === id}
          <Component />
        {/if}
      {/each}
    </main>
  </div>
</div>
