<script>
  const { onselect } = $props()

  const managementRoles = [
    { id: 'po', label: 'Product Owner' },
    { id: 'business-analyst', label: 'Business Analyst' },
    { id: 'project-manager', label: 'Project Manager' },
    { id: 'scrum-master', label: 'Scrum Master' },
    { id: 'people-ops', label: 'People Ops' },
    { id: 'change-manager', label: 'Change Manager' },
    { id: 'architect', label: 'Architect' },
    { id: 'ux-designer', label: 'UX Designer' },
    { id: 'finops', label: 'FinOps' },
    { id: 'soc2-compliance', label: 'SOC2 Compliance' },
  ]

  const devRoles = [
    { id: 'dev', label: 'Developer' },
    { id: 'backend', label: 'Backend' },
    { id: 'frontend', label: 'Frontend' },
    { id: 'devops', label: 'DevOps' },
    { id: 'security', label: 'Security' },
    { id: 'sre', label: 'SRE' },
    { id: 'golang', label: 'Go' },
    { id: 'python', label: 'Python' },
    { id: 'typescript', label: 'TypeScript' },
  ]

  let selected = $state(null)
  let devExpanded = $state(false)

  function select(role) {
    selected = role.id
    onselect?.(role)
  }
</script>

<div class="space-y-3">
  <h3 class="text-xs font-semibold uppercase tracking-wider text-ck-dim">Management Roles</h3>
  <div class="flex flex-wrap gap-1.5">
    {#each managementRoles as role}
      <button
        onclick={() => select(role)}
        class="px-2.5 py-1 text-xs rounded-md transition-colors
          {selected === role.id ? 'bg-ck-rose text-white' : 'bg-ck-dark text-ck-dim hover:text-white hover:bg-ck-pink/20'}"
      >{role.label}</button>
    {/each}
  </div>

  <button
    onclick={() => devExpanded = !devExpanded}
    class="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider text-ck-dim hover:text-white transition-colors"
  >
    <span class="transition-transform {devExpanded ? 'rotate-90' : ''}">&rsaquo;</span>
    Developer Roles
  </button>

  {#if devExpanded}
    <div class="flex flex-wrap gap-1.5">
      {#each devRoles as role}
        <button
          onclick={() => select(role)}
          class="px-2.5 py-1 text-xs rounded-md transition-colors
            {selected === role.id ? 'bg-ck-rose text-white' : 'bg-ck-dark text-ck-dim hover:text-white hover:bg-ck-pink/20'}"
        >{role.label}</button>
      {/each}
    </div>
  {/if}
</div>
