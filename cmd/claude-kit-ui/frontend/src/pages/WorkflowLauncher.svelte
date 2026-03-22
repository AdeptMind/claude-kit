<script>
  import PhaseDiagram from '../components/PhaseDiagram.svelte'
  import ArtifactMonitor from '../components/ArtifactMonitor.svelte'

  let phases = $state([
    { id: 'break', status: 'done', label: 'Break', artifacts: ['problem.md'] },
    { id: 'model', status: 'done', label: 'Model', artifacts: ['architecture.md', 'backlog.md'] },
    { id: 'analyze', status: 'active', label: 'Analyze', artifacts: ['analysis-report.md'] },
    { id: 'act', status: 'pending', label: 'Act', artifacts: [] },
    { id: 'deliver', status: 'pending', label: 'Deliver', artifacts: [] },
  ])

  let selectedPhase = $state(null)
  let launching = $state(false)

  $effect(() => {
    try {
      import('../../wailsjs/go/main/WorkflowService.js').then(({ GetStatus }) => {
        GetStatus().then(status => {
          if (status?.phases) phases = status.phases
        }).catch(() => {})
      }).catch(() => {})
    } catch {}
  })

  function handleSelectPhase(phase) {
    selectedPhase = phase
  }

  async function openCowork() {
    launching = true
    try {
      const { LaunchCowork } = await import('../../wailsjs/go/main/WorkflowService.js')
      await LaunchCowork()
    } catch {
      // Wails binding not available
    } finally {
      launching = false
    }
  }
</script>

<div class="flex flex-col h-full gap-6">
  <div class="text-center">
    <h2 class="text-xl font-bold text-white mb-1">Workflow</h2>
    <p class="text-xs text-ck-dim font-mono">Break &rarr; Model &rarr; Analyze &rarr; Act &rarr; Deliver</p>
  </div>

  <PhaseDiagram bind:phases onSelectPhase={handleSelectPhase} />

  <div class="flex justify-center">
    <button
      onclick={openCowork}
      disabled={launching}
      class="px-6 py-2.5 rounded-lg bg-ck-rose text-white font-bold text-sm
        hover:bg-ck-pink transition-colors duration-200 disabled:opacity-50"
    >
      {launching ? 'Launching...' : 'Open in Cowork'}
    </button>
  </div>

  <div class="flex-1 overflow-auto">
    <ArtifactMonitor bind:phase={selectedPhase} />
  </div>
</div>
