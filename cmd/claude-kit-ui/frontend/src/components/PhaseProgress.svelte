<script>
  let { phases = [] } = $props()
</script>

<div class="flex items-center gap-1 w-full">
  {#each phases as phase, i}
    <div class="flex items-center {i < phases.length - 1 ? 'flex-1' : ''}">
      <div class="flex flex-col items-center gap-1.5">
        <div
          class="w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold border-2 transition-colors
            {phase.status === 'done' ? 'bg-ck-green/20 border-ck-green text-ck-green' :
             phase.status === 'active' ? 'bg-ck-gold/20 border-ck-gold text-ck-gold animate-pulse' :
             'bg-ck-dark border-gray-700 text-ck-dim'}"
        >
          {#if phase.status === 'done'}
            ✓
          {:else}
            {i + 1}
          {/if}
        </div>
        <span class="text-[10px] font-medium whitespace-nowrap
          {phase.status === 'done' ? 'text-ck-green' :
           phase.status === 'active' ? 'text-ck-gold' :
           'text-ck-dim'}"
        >
          {phase.label}
        </span>
      </div>
      {#if i < phases.length - 1}
        <div class="flex-1 h-0.5 mx-2 mt-[-1rem]
          {phase.status === 'done' ? 'bg-ck-green' : 'bg-gray-700'}">
        </div>
      {/if}
    </div>
  {/each}
</div>
