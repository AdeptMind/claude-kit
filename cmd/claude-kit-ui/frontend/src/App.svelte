<script>
  import { currentPage, navigateTo } from './stores/navigation.js'
  import { switchProjectPrev, switchProjectNext, projectFinderOpen } from './stores/project.js'
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

  function isEditing() {
    const el = document.activeElement
    if (!el) return false
    const tag = el.tagName.toLowerCase()
    return tag === 'input' || tag === 'textarea' || el.isContentEditable
  }

  function handleGlobalKeydown(e) {
    // Accept both Cmd+key and Cmd+Shift+key (the latter works inside Chat terminal)
    const cmd = e.metaKey || e.ctrlKey

    if (e.key === 'Escape') {
      window.dispatchEvent(new CustomEvent('ck:close-modal'))
      return
    }

    if (!cmd) return

    if (e.key === 'ArrowLeft') {
      e.preventDefault()
      switchProjectPrev()
    } else if (e.key === 'ArrowRight') {
      e.preventDefault()
      switchProjectNext()
    } else if (e.key === 'k' || e.key === 'K') {
      e.preventDefault()
      projectFinderOpen.set(true)
    } else if (e.key === 'j' || e.key === 'J') {
      e.preventDefault()
      navigateTo('chat')
    } else if ((e.key === 's' || e.key === 'S') && !isEditing()) {
      e.preventDefault()
      navigateTo('stories')
    } else if ((e.key === 'd' || e.key === 'D') && !isEditing()) {
      e.preventDefault()
      navigateTo('dashboard')
    } else if ((e.key === 'f' || e.key === 'F') && !isEditing()) {
      e.preventDefault()
      navigateTo('files')
      window.dispatchEvent(new CustomEvent('ck:file-search'))
    } else if (e.key === 'e' || e.key === 'E') {
      e.preventDefault()
      window.dispatchEvent(new CustomEvent('ck:edit-story'))
    }
  }
</script>

<svelte:window on:keydown={handleGlobalKeydown} />

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
