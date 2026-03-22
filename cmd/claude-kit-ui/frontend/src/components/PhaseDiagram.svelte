<script>
  const PHASES = [
    { id: 'break', label: 'Break', artifacts: ['problem.md'] },
    { id: 'model', label: 'Model', artifacts: ['architecture.md', 'backlog.md'] },
    { id: 'analyze', label: 'Analyze', artifacts: ['analysis-report.md'] },
    { id: 'act', label: 'Act', artifacts: ['src/'] },
    { id: 'deliver', label: 'Deliver', artifacts: ['CHANGELOG.md'] },
  ]

  let { phases = $bindable([]), onSelectPhase = () => {} } = $props()

  function statusFor(id) {
    return phases.find(p => p.id === id)?.status ?? 'pending'
  }

  function handleClick(phase) {
    const s = statusFor(phase.id)
    if (s !== 'pending') onSelectPhase(phase)
  }

  function statusClasses(status) {
    if (status === 'done') return 'bg-ck-green text-black border-ck-green'
    if (status === 'active') return 'bg-ck-gold text-black border-ck-gold animate-pulse'
    return 'bg-transparent text-ck-dim border-ck-dim'
  }

  function connectorClass(idx) {
    const prev = statusFor(PHASES[idx - 1]?.id)
    return prev === 'done' ? 'bg-ck-green' : 'bg-ck-dim/40'
  }
</script>

<div class="flex items-center justify-center gap-0 w-full py-6">
  {#each PHASES as phase, i}
    {#if i > 0}
      <div class="h-0.5 w-10 sm:w-16 {connectorClass(i)} transition-colors duration-300"></div>
    {/if}
    <button
      onclick={() => handleClick(phase)}
      class="flex flex-col items-center gap-2 group"
      disabled={statusFor(phase.id) === 'pending'}
    >
      <div
        class="w-12 h-12 rounded-full border-2 flex items-center justify-center text-sm font-bold transition-all duration-300
          {statusClasses(statusFor(phase.id))}
          {statusFor(phase.id) !== 'pending' ? 'cursor-pointer hover:scale-110' : 'cursor-default'}"
      >
        {#if statusFor(phase.id) === 'done'}
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
          </svg>
        {:else}
          {i + 1}
        {/if}
      </div>
      <span class="text-xs font-mono {statusFor(phase.id) === 'pending' ? 'text-ck-dim' : 'text-white'}">
        {phase.label}
      </span>
    </button>
  {/each}
</div>
