<script>
  let { content = '', filename = '', onClose = () => {} } = $props()

  function renderMarkdown(raw) {
    if (!raw) return '<p class="text-ck-dim italic">No content.</p>'
    return raw
      .replace(/^### (.+)$/gm, '<h3 class="text-base font-semibold text-ck-gold mt-3 mb-1">$1</h3>')
      .replace(/^## (.+)$/gm, '<h2 class="text-lg font-bold text-ck-pink mt-4 mb-1">$1</h2>')
      .replace(/^# (.+)$/gm, '<h1 class="text-xl font-bold text-ck-pink mt-4 mb-2">$1</h1>')
      .replace(/`([^`]+)`/g, '<code class="px-1 py-0.5 rounded bg-ck-bg text-ck-gold text-xs">$1</code>')
      .replace(/\*\*(.+?)\*\*/g, '<strong class="text-white font-bold">$1</strong>')
      .replace(/\*(.+?)\*/g, '<em class="italic text-gray-200">$1</em>')
      .replace(/^(\d+)\. (.+)$/gm, '<li class="ml-6 list-decimal text-gray-300">$2</li>')
      .replace(/^- (.+)$/gm, '<li class="ml-4 list-disc text-gray-300">$1</li>')
      .replace(/^---$/gm, '<hr class="border-gray-700 my-3"/>')
      .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a class="text-ck-pink underline" href="$2">$1</a>')
      .replace(/\n{2,}/g, '<br/><br/>')
      .replace(/\n/g, '<br/>')
  }

  function isCode(name) {
    return /\.(go|ts|js|py|sh|yaml|yml|json|toml|svelte|css|html)$/.test(name)
  }
</script>

<!-- Full-screen modal overlay -->
<div class="fixed inset-0 z-50 bg-black/60 flex items-center justify-center">
  <div class="w-[80%] h-[85%] bg-ck-dark rounded-xl border border-gray-700 flex flex-col shadow-2xl">

    <!-- Header -->
    <div class="flex items-center justify-between px-5 py-3 border-b border-gray-700 rounded-t-xl">
      <div class="flex items-center gap-3">
        <span class="text-sm font-medium text-white">{filename}</span>
        <span class="text-[10px] px-2 py-0.5 rounded-full bg-ck-dim/20 text-ck-dim border border-gray-600 font-semibold">Read-only</span>
      </div>
      <button
        onclick={onClose}
        class="px-3 py-1.5 text-xs font-medium rounded-lg bg-ck-dark text-ck-dim border border-gray-700 hover:text-white transition-colors"
      >
        Close
      </button>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto p-6">
      {#if isCode(filename)}
        <pre class="text-sm font-mono text-gray-300 whitespace-pre-wrap leading-relaxed">{content}</pre>
      {:else}
        <div class="text-sm text-gray-300 leading-relaxed">
          {@html renderMarkdown(content)}
        </div>
      {/if}
    </div>
  </div>
</div>
