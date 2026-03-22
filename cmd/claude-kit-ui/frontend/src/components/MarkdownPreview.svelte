<script>
  const { content } = $props()

  function renderMarkdown(text) {
    if (!text) return '<p class="text-ck-dim italic">Start typing to see preview...</p>'
    return text
      .replace(/^### (.+)$/gm, '<h3 class="text-base font-semibold text-ck-gold mt-3 mb-1">$1</h3>')
      .replace(/^## (.+)$/gm, '<h2 class="text-lg font-bold text-ck-pink mt-4 mb-1">$1</h2>')
      .replace(/^# (.+)$/gm, '<h1 class="text-xl font-bold text-ck-pink mt-4 mb-2">$1</h1>')
      .replace(/\*\*(.+?)\*\*/g, '<strong class="text-white">$1</strong>')
      .replace(/^- (.+)$/gm, '<li class="ml-4 list-disc text-gray-300">$1</li>')
      .replace(/\n{2,}/g, '<br/><br/>')
      .replace(/\n/g, '<br/>')
  }
</script>

<div class="h-full flex flex-col">
  <h3 class="text-xs font-semibold uppercase tracking-wider text-ck-dim mb-3">Preview: about-me.md</h3>
  <div class="flex-1 overflow-auto bg-ck-dark rounded-lg p-4 text-sm text-gray-300 leading-relaxed">
    {@html renderMarkdown(content)}
  </div>
</div>
