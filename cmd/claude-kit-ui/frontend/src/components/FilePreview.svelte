<script>
  let { path = '', content = '' } = $props()

  const codeExts = ['js', 'ts', 'go', 'py', 'svelte', 'json', 'yaml', 'yml', 'toml', 'sh', 'css', 'html']

  function isCode(filePath) {
    const ext = filePath.split('.').pop()?.toLowerCase()
    return codeExts.includes(ext)
  }
</script>

{#if !path}
  <div class="flex items-center justify-center h-full text-ck-dim">
    <p class="text-sm">Select a file to preview</p>
  </div>
{:else}
  <div class="h-full flex flex-col">
    <div class="flex items-center gap-2 px-4 py-2 border-b border-gray-800 shrink-0">
      <span class="text-xs text-ck-dim font-mono truncate">{path}</span>
    </div>
    <div class="flex-1 overflow-auto p-4">
      <pre class="{isCode(path) ? 'font-mono text-xs' : 'text-sm'} text-gray-300 whitespace-pre-wrap break-words leading-relaxed">{content}</pre>
    </div>
  </div>
{/if}
