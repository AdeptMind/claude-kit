<script>
  let { actions = [] } = $props()
</script>

<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
  {#each actions as action}
    <button
      onclick={action.handler}
      class="group flex flex-col items-start gap-3 p-5 rounded-xl bg-ck-dark border border-gray-800
        hover:border-ck-rose/50 hover:bg-ck-dark/80 transition-all duration-200"
    >
      <span class="text-2xl">{action.icon}</span>
      <div>
        <h3 class="text-sm font-semibold text-white group-hover:text-ck-pink transition-colors">
          {action.title}
        </h3>
        <p class="text-xs text-ck-dim mt-1">{action.description}</p>
      </div>
    </button>
  {/each}
</div>
