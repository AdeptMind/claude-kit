<script>
  let { phase = $bindable(null) } = $props()

  let artifacts = $state([])
  let selectedArtifact = $state(null)
  let highlightedId = $state(null)

  $effect(() => {
    if (phase) {
      artifacts = (phase.artifacts ?? []).map((a, i) => ({
        id: `${phase.id}-${i}`,
        name: a,
        preview: `Content of ${a} (phase: ${phase.label})`,
      }))
      selectedArtifact = null
    } else {
      artifacts = []
      selectedArtifact = null
    }
  })

  $effect(() => {
    const handler = (data) => {
      if (!phase || data?.phase !== phase.id) return
      const newArt = {
        id: `${phase.id}-${Date.now()}`,
        name: data.name,
        preview: data.preview ?? `New artifact: ${data.name}`,
      }
      highlightedId = newArt.id
      artifacts = [...artifacts, newArt]
      setTimeout(() => { highlightedId = null }, 2000)
    }
    const unsub = window.runtime?.EventsOn?.('artifact:new', handler)
    return () => { if (typeof unsub === 'function') unsub() }
  })

  function select(art) {
    selectedArtifact = art
  }
</script>

{#if phase}
  <div class="flex flex-col lg:flex-row gap-4 w-full">
    <div class="lg:w-1/3 flex flex-col gap-1">
      <h3 class="text-sm font-mono text-ck-dim mb-2">
        Artifacts — <span class="text-white">{phase.label}</span>
      </h3>
      {#each artifacts as art (art.id)}
        <button
          onclick={() => select(art)}
          class="text-left px-3 py-2 rounded text-sm font-mono transition-all duration-300
            {selectedArtifact?.id === art.id ? 'bg-ck-rose/30 text-white' : 'text-ck-dim hover:text-white hover:bg-white/5'}
            {highlightedId === art.id ? 'ring-1 ring-ck-gold bg-ck-gold/10' : ''}"
        >
          {art.name}
        </button>
      {/each}
      {#if artifacts.length === 0}
        <p class="text-ck-dim text-xs italic">No artifacts yet.</p>
      {/if}
    </div>

    <div class="lg:w-2/3 bg-ck-dark rounded-lg border border-gray-800 p-4 min-h-[120px]">
      {#if selectedArtifact}
        <h4 class="text-sm font-mono text-ck-pink mb-2">{selectedArtifact.name}</h4>
        <pre class="text-xs text-ck-dim whitespace-pre-wrap">{selectedArtifact.preview}</pre>
      {:else}
        <p class="text-ck-dim text-xs italic">Select an artifact to preview.</p>
      {/if}
    </div>
  </div>
{:else}
  <p class="text-ck-dim text-sm italic text-center py-4">Click a completed phase to view its artifacts.</p>
{/if}
