<script>
  let { content = '', filename = '', onSave = () => {}, onCancel = () => {} } = $props()

  let htmlContent = $state('')
  let editor = $state(null)
  let isBold = $state(false)
  let isItalic = $state(false)

  function markdownToHtml(md) {
    if (!md) return '<p><br></p>'
    let html = md
      // Headings
      .replace(/^### (.+)$/gm, '<h3>$1</h3>')
      .replace(/^## (.+)$/gm, '<h2>$1</h2>')
      .replace(/^# (.+)$/gm, '<h1>$1</h1>')
      // Inline
      .replace(/`([^`]+)`/g, '<code>$1</code>')
      .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
      .replace(/\*(.+?)\*/g, '<em>$1</em>')
      // Links
      .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2">$1</a>')
      // HR
      .replace(/^---$/gm, '<hr>')
      // Lists — collect consecutive lines
      .replace(/(?:^- .+\n?)+/gm, (match) => {
        const items = match.trim().split('\n').map(l => `<li>${l.replace(/^- /, '')}</li>`).join('')
        return `<ul>${items}</ul>`
      })
      .replace(/(?:^\d+\. .+\n?)+/gm, (match) => {
        const items = match.trim().split('\n').map(l => `<li>${l.replace(/^\d+\. /, '')}</li>`).join('')
        return `<ol>${items}</ol>`
      })
    // Wrap remaining bare lines in <p> tags
    html = html.split('\n\n').map(block => {
      const trimmed = block.trim()
      if (!trimmed) return ''
      if (/^<(h[1-3]|ul|ol|hr|p|div|blockquote)/.test(trimmed)) return trimmed
      return `<p>${trimmed.replace(/\n/g, '<br>')}</p>`
    }).join('\n')
    return html || '<p><br></p>'
  }

  function htmlToMarkdown(html) {
    let md = html
    // Block elements first
    md = md.replace(/<h1[^>]*>([\s\S]*?)<\/h1>/gi, (_, c) => `# ${stripTags(c)}\n\n`)
    md = md.replace(/<h2[^>]*>([\s\S]*?)<\/h2>/gi, (_, c) => `## ${stripTags(c)}\n\n`)
    md = md.replace(/<h3[^>]*>([\s\S]*?)<\/h3>/gi, (_, c) => `### ${stripTags(c)}\n\n`)
    md = md.replace(/<hr[^>]*\/?>/gi, '---\n\n')
    // Lists
    md = md.replace(/<ul[^>]*>([\s\S]*?)<\/ul>/gi, (_, inner) => {
      return inner.replace(/<li[^>]*>([\s\S]*?)<\/li>/gi, (_, li) => `- ${stripTags(li).trim()}\n`) + '\n'
    })
    md = md.replace(/<ol[^>]*>([\s\S]*?)<\/ol>/gi, (_, inner) => {
      let i = 0
      return inner.replace(/<li[^>]*>([\s\S]*?)<\/li>/gi, (_, li) => `${++i}. ${stripTags(li).trim()}\n`) + '\n'
    })
    // Links before stripping other tags
    md = md.replace(/<a[^>]+href="([^"]*)"[^>]*>([\s\S]*?)<\/a>/gi, (_, href, text) => `[${stripTags(text)}](${href})`)
    // Inline formatting
    md = md.replace(/<(strong|b)[^>]*>([\s\S]*?)<\/\1>/gi, (_, _t, c) => `**${c}**`)
    md = md.replace(/<(em|i)[^>]*>([\s\S]*?)<\/\1>/gi, (_, _t, c) => `*${c}*`)
    md = md.replace(/<code[^>]*>([\s\S]*?)<\/code>/gi, (_, c) => `\`${c}\``)
    // Paragraphs and line breaks
    md = md.replace(/<\/p>\s*<p[^>]*>/gi, '\n\n')
    md = md.replace(/<br\s*\/?>/gi, '\n')
    md = md.replace(/<div[^>]*>/gi, '\n')
    md = md.replace(/<\/div>/gi, '')
    // Strip remaining tags
    md = md.replace(/<\/?p[^>]*>/gi, '')
    md = md.replace(/<[^>]+>/g, '')
    // Decode entities
    md = md.replace(/&amp;/g, '&').replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&nbsp;/g, ' ').replace(/&quot;/g, '"')
    // Clean up whitespace: collapse 3+ newlines to 2
    md = md.replace(/\n{3,}/g, '\n\n')
    return md.trim() + '\n'
  }

  function stripTags(html) {
    return html.replace(/<[^>]+>/g, '')
  }

  $effect(() => { htmlContent = markdownToHtml(content) })

  function updateToolbarState() {
    isBold = document.queryCommandState('bold')
    isItalic = document.queryCommandState('italic')
  }

  function onSelectionChange() {
    updateToolbarState()
  }

  $effect(() => {
    document.addEventListener('selectionchange', onSelectionChange)
    return () => document.removeEventListener('selectionchange', onSelectionChange)
  })

  function execCmd(cmd, value = null) {
    document.execCommand(cmd, false, value)
    if (editor) htmlContent = editor.innerHTML
    updateToolbarState()
  }

  function bold() { execCmd('bold') }
  function italic() { execCmd('italic') }
  function h1() { execCmd('formatBlock', 'h1') }
  function h2() { execCmd('formatBlock', 'h2') }
  function h3() { execCmd('formatBlock', 'h3') }
  function paragraph() { execCmd('formatBlock', 'p') }
  function bulletList() { execCmd('insertUnorderedList') }
  function numberedList() { execCmd('insertOrderedList') }
  function hr() { execCmd('insertHTML', '<hr>') }

  function code() {
    const sel = window.getSelection()
    if (sel.rangeCount > 0) {
      const text = sel.toString()
      if (text) {
        execCmd('insertHTML', `<code>${text}</code>`)
      }
    }
  }

  function link() {
    const sel = window.getSelection()
    const text = sel.toString()
    const url = prompt('URL:', 'https://')
    if (url) {
      if (text) {
        execCmd('createLink', url)
      } else {
        execCmd('insertHTML', `<a href="${url}">${url}</a>`)
      }
    }
  }

  function handleInput() {
    if (editor) htmlContent = editor.innerHTML
  }

  function handleSave() {
    onSave(htmlToMarkdown(htmlContent))
  }

  let charCount = $derived(stripTags(htmlContent).replace(/&nbsp;/g, ' ').replace(/&amp;/g, '&').replace(/&lt;/g, '<').replace(/&gt;/g, '>').length)
  let lineCount = $derived(htmlToMarkdown(htmlContent).split('\n').filter(l => l !== '').length)
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
        <button onclick={bold} class="px-2 py-1 text-xs font-bold rounded transition-colors {isBold ? 'bg-ck-pink/30 text-white' : 'text-gray-400 hover:bg-ck-pink/20 hover:text-white'}" title="Bold">B</button>
        <button onclick={italic} class="px-2 py-1 text-xs italic rounded transition-colors {isItalic ? 'bg-ck-pink/30 text-white' : 'text-gray-400 hover:bg-ck-pink/20 hover:text-white'}" title="Italic">I</button>
        <span class="w-px h-4 bg-gray-700"></span>
        <button onclick={h1} class="px-2 py-1 text-xs text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Heading 1">H1</button>
        <button onclick={h2} class="px-2 py-1 text-xs text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Heading 2">H2</button>
        <button onclick={h3} class="px-2 py-1 text-xs text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Heading 3">H3</button>
        <button onclick={paragraph} class="px-2 py-1 text-xs text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Normal text">P</button>
        <span class="w-px h-4 bg-gray-700"></span>
        <button onclick={code} class="px-2 py-1 text-xs font-mono text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Inline code">&lt;/&gt;</button>
        <button onclick={bulletList} class="px-2 py-1 text-xs text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Bullet list">&#8226; List</button>
        <button onclick={numberedList} class="px-2 py-1 text-xs text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Numbered list">1. List</button>
        <button onclick={link} class="px-2 py-1 text-xs text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Link">Link</button>
        <button onclick={hr} class="px-2 py-1 text-xs text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Horizontal rule">---</button>
      </div>

      <div class="flex items-center gap-2">
        <button
          onclick={handleSave}
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

    <!-- WYSIWYG editor area -->
    <div
      bind:this={editor}
      contenteditable="true"
      oninput={handleInput}
      class="flex-1 overflow-y-auto p-6 prose prose-invert prose-headings:text-ck-pink prose-strong:text-white prose-code:text-ck-gold prose-code:bg-ck-bg prose-code:px-1 prose-code:py-0.5 prose-code:rounded prose-a:text-ck-rose prose-hr:border-gray-700 max-w-none focus:outline-none text-sm text-gray-300 leading-relaxed bg-ck-bg"
      style="min-height: 100%"
    >
      {@html htmlContent}
    </div>

    <!-- Bottom bar -->
    <div class="flex items-center justify-between px-4 py-2 border-t border-gray-700 bg-ck-bg rounded-b-xl">
      <div class="flex items-center gap-4 text-[11px] text-ck-dim">
        <span>{charCount} characters</span>
        <span>{lineCount} lines</span>
      </div>
      <span class="text-[11px] text-ck-dim">WYSIWYG editor</span>
    </div>
  </div>
</div>
