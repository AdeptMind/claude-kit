<script>
  let { content = '', filename = '', onSave = () => {}, onCancel = () => {} } = $props()

  let text = $state('')
  let textarea = $state(null)

  $effect(() => { text = content })

  function renderMarkdown(raw) {
    if (!raw) return '<p class="text-ck-dim italic">Start typing to see preview...</p>'
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

  function insertAtCursor(before, after = '') {
    if (!textarea) return
    const start = textarea.selectionStart
    const end = textarea.selectionEnd
    const selected = text.slice(start, end)
    const replacement = before + (selected || 'text') + after
    text = text.slice(0, start) + replacement + text.slice(end)
    // Restore cursor after the inserted text
    const cursorPos = start + replacement.length
    requestAnimationFrame(() => {
      textarea.focus()
      textarea.setSelectionRange(cursorPos, cursorPos)
    })
  }

  function insertAtLineStart(prefix) {
    if (!textarea) return
    const start = textarea.selectionStart
    // Find the beginning of the current line
    const lineStart = text.lastIndexOf('\n', start - 1) + 1
    text = text.slice(0, lineStart) + prefix + text.slice(lineStart)
    const cursorPos = start + prefix.length
    requestAnimationFrame(() => {
      textarea.focus()
      textarea.setSelectionRange(cursorPos, cursorPos)
    })
  }

  function insertBlock(block) {
    if (!textarea) return
    const pos = textarea.selectionStart
    text = text.slice(0, pos) + block + text.slice(pos)
    const cursorPos = pos + block.length
    requestAnimationFrame(() => {
      textarea.focus()
      textarea.setSelectionRange(cursorPos, cursorPos)
    })
  }

  function bold() { insertAtCursor('**', '**') }
  function italic() { insertAtCursor('*', '*') }
  function code() { insertAtCursor('`', '`') }
  function h1() { insertAtLineStart('# ') }
  function h2() { insertAtLineStart('## ') }
  function h3() { insertAtLineStart('### ') }
  function bulletList() { insertAtLineStart('- ') }
  function numberedList() { insertAtLineStart('1. ') }
  function link() { insertAtCursor('[', '](url)') }
  function hr() { insertBlock('\n---\n') }

  let lineCount = $derived(text.split('\n').length)
  let charCount = $derived(text.length)
</script>

<!-- Full-screen modal overlay -->
<div class="fixed inset-0 z-50 bg-black/60 flex items-center justify-center">
  <div class="w-[92%] h-[90%] bg-ck-dark rounded-xl border border-gray-700 flex flex-col shadow-2xl">

    <!-- Top bar / toolbar -->
    <div class="flex items-center justify-between px-4 py-2.5 bg-ck-bg border-b border-gray-700 rounded-t-xl">
      <div class="flex items-center gap-3">
        <span class="text-sm font-medium text-white">{filename}</span>
        <span class="text-[10px] px-2 py-0.5 rounded-full bg-ck-pink/20 text-ck-pink border border-ck-pink/30 font-semibold">Editing</span>
      </div>

      <!-- Formatting buttons -->
      <div class="flex items-center gap-1">
        <button onclick={bold} class="px-2 py-1 text-xs font-bold text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Bold">B</button>
        <button onclick={italic} class="px-2 py-1 text-xs italic text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Italic">I</button>
        <span class="w-px h-4 bg-gray-700"></span>
        <button onclick={h1} class="px-2 py-1 text-xs text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Heading 1">H1</button>
        <button onclick={h2} class="px-2 py-1 text-xs text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Heading 2">H2</button>
        <button onclick={h3} class="px-2 py-1 text-xs text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Heading 3">H3</button>
        <span class="w-px h-4 bg-gray-700"></span>
        <button onclick={code} class="px-2 py-1 text-xs font-mono text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Inline code">&lt;/&gt;</button>
        <button onclick={bulletList} class="px-2 py-1 text-xs text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Bullet list">&#8226; List</button>
        <button onclick={numberedList} class="px-2 py-1 text-xs text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Numbered list">1. List</button>
        <button onclick={link} class="px-2 py-1 text-xs text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Link">Link</button>
        <button onclick={hr} class="px-2 py-1 text-xs text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Horizontal rule">---</button>
      </div>

      <div class="flex items-center gap-2">
        <button
          onclick={() => onSave(text)}
          class="px-4 py-1.5 text-xs font-medium rounded-lg bg-ck-green/20 text-ck-green border border-ck-green/30 hover:bg-ck-green/30 transition-colors"
        >
          Save
        </button>
        <button
          onclick={onCancel}
          class="px-4 py-1.5 text-xs font-medium rounded-lg bg-ck-dark text-ck-dim border border-gray-700 hover:text-white transition-colors"
        >
          Cancel
        </button>
      </div>
    </div>

    <!-- Editor area: split view -->
    <div class="flex-1 flex min-h-0">
      <!-- Left: raw markdown textarea -->
      <div class="w-1/2 flex flex-col border-r border-gray-700">
        <div class="px-3 py-1.5 text-[10px] uppercase tracking-wider text-ck-dim border-b border-gray-800 bg-ck-bg/50">Markdown</div>
        <textarea
          bind:this={textarea}
          bind:value={text}
          class="flex-1 w-full bg-ck-bg p-4 text-sm font-mono text-gray-300 resize-none focus:outline-none leading-relaxed"
          spellcheck="false"
          placeholder="Start writing markdown..."
        ></textarea>
      </div>

      <!-- Right: rendered preview -->
      <div class="w-1/2 flex flex-col">
        <div class="px-3 py-1.5 text-[10px] uppercase tracking-wider text-ck-dim border-b border-gray-800 bg-ck-bg/50">Preview</div>
        <div class="flex-1 overflow-y-auto p-4 text-sm text-gray-300 leading-relaxed">
          {@html renderMarkdown(text)}
        </div>
      </div>
    </div>

    <!-- Bottom bar -->
    <div class="flex items-center justify-between px-4 py-2 border-t border-gray-700 bg-ck-bg rounded-b-xl">
      <div class="flex items-center gap-4 text-[11px] text-ck-dim">
        <span>{charCount} characters</span>
        <span>{lineCount} lines</span>
      </div>
      <span class="text-[11px] text-ck-dim">Markdown supported</span>
    </div>
  </div>
</div>
