<script>
  import { onMount, onDestroy } from 'svelte'
  import { Editor } from '@tiptap/core'
  import StarterKit from '@tiptap/starter-kit'
  import Placeholder from '@tiptap/extension-placeholder'

  let { content = '', filename = '', onSave = () => {}, onCancel = () => {} } = $props()

  let editor = $state(null)
  let element = $state(null)

  function markdownToHtml(md) {
    if (!md) return '<p></p>'
    let html = md
      .replace(/^### (.+)$/gm, '<h3>$1</h3>')
      .replace(/^## (.+)$/gm, '<h2>$1</h2>')
      .replace(/^# (.+)$/gm, '<h1>$1</h1>')
      .replace(/`([^`]+)`/g, '<code>$1</code>')
      .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
      .replace(/\*(.+?)\*/g, '<em>$1</em>')
      .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2">$1</a>')
      .replace(/^---$/gm, '<hr>')
      .replace(/(?:^- .+\n?)+/gm, (match) => {
        const items = match.trim().split('\n').map(l => `<li>${l.replace(/^- /, '')}</li>`).join('')
        return `<ul>${items}</ul>`
      })
      .replace(/(?:^\d+\. .+\n?)+/gm, (match) => {
        const items = match.trim().split('\n').map(l => `<li>${l.replace(/^\d+\. /, '')}</li>`).join('')
        return `<ol>${items}</ol>`
      })
    html = html.split('\n\n').map(block => {
      const trimmed = block.trim()
      if (!trimmed) return ''
      if (/^<(h[1-3]|ul|ol|hr|p|div|blockquote|pre)/.test(trimmed)) return trimmed
      return `<p>${trimmed.replace(/\n/g, '<br>')}</p>`
    }).join('\n')
    return html || '<p></p>'
  }

  function htmlToMarkdown(html) {
    let md = html
    md = md.replace(/<h1[^>]*>([\s\S]*?)<\/h1>/gi, (_, c) => `# ${stripTags(c)}\n\n`)
    md = md.replace(/<h2[^>]*>([\s\S]*?)<\/h2>/gi, (_, c) => `## ${stripTags(c)}\n\n`)
    md = md.replace(/<h3[^>]*>([\s\S]*?)<\/h3>/gi, (_, c) => `### ${stripTags(c)}\n\n`)
    md = md.replace(/<hr[^>]*\/?>/gi, '---\n\n')
    md = md.replace(/<pre[^>]*><code[^>]*>([\s\S]*?)<\/code><\/pre>/gi, (_, c) => '```\n' + c.replace(/&amp;/g, '&').replace(/&lt;/g, '<').replace(/&gt;/g, '>') + '\n```\n\n')
    md = md.replace(/<ul[^>]*>([\s\S]*?)<\/ul>/gi, (_, inner) => {
      return inner.replace(/<li[^>]*>([\s\S]*?)<\/li>/gi, (_, li) => `- ${stripTags(li).trim()}\n`) + '\n'
    })
    md = md.replace(/<ol[^>]*>([\s\S]*?)<\/ol>/gi, (_, inner) => {
      let i = 0
      return inner.replace(/<li[^>]*>([\s\S]*?)<\/li>/gi, (_, li) => `${++i}. ${stripTags(li).trim()}\n`) + '\n'
    })
    md = md.replace(/<a[^>]+href="([^"]*)"[^>]*>([\s\S]*?)<\/a>/gi, (_, href, text) => `[${stripTags(text)}](${href})`)
    md = md.replace(/<(strong|b)[^>]*>([\s\S]*?)<\/\1>/gi, (_, _t, c) => `**${c}**`)
    md = md.replace(/<(em|i)[^>]*>([\s\S]*?)<\/\1>/gi, (_, _t, c) => `*${c}*`)
    md = md.replace(/<code[^>]*>([\s\S]*?)<\/code>/gi, (_, c) => `\`${c}\``)
    md = md.replace(/<\/p>\s*<p[^>]*>/gi, '\n\n')
    md = md.replace(/<br\s*\/?>/gi, '\n')
    md = md.replace(/<div[^>]*>/gi, '\n')
    md = md.replace(/<\/div>/gi, '')
    md = md.replace(/<\/?p[^>]*>/gi, '')
    md = md.replace(/<[^>]+>/g, '')
    md = md.replace(/&amp;/g, '&').replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&nbsp;/g, ' ').replace(/&quot;/g, '"')
    md = md.replace(/\n{3,}/g, '\n\n')
    return md.trim() + '\n'
  }

  function stripTags(html) {
    return html.replace(/<[^>]+>/g, '')
  }

  onMount(() => {
    editor = new Editor({
      element: element,
      extensions: [
        StarterKit.configure({
          heading: { levels: [1, 2, 3] },
        }),
        Placeholder.configure({
          placeholder: 'Start writing...',
        }),
      ],
      content: markdownToHtml(content),
      editorProps: {
        attributes: {
          class: 'prose prose-invert max-w-none focus:outline-none min-h-full p-6',
        },
      },
      onTransaction: () => {
        editor = editor
      },
    })
  })

  onDestroy(() => {
    if (editor) editor.destroy()
  })

  function handleSave() {
    if (!editor) return
    onSave(htmlToMarkdown(editor.getHTML()))
  }

  function btnClass(active) {
    return active
      ? 'px-2 py-1 text-xs rounded transition-colors bg-ck-pink/30 text-white'
      : 'px-2 py-1 text-xs rounded transition-colors text-gray-400 hover:bg-ck-pink/20 hover:text-white'
  }

  let charCount = $derived(editor ? editor.storage.characterCount?.characters?.() ?? editor.getText().length : 0)
  let lineCount = $derived(editor ? editor.getText().split('\n').filter(l => l !== '').length : 0)
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
      {#if editor}
      <div class="flex items-center gap-1">
        <button onclick={() => editor.chain().focus().toggleBold().run()} class="{btnClass(editor.isActive('bold'))} font-bold" title="Bold">B</button>
        <button onclick={() => editor.chain().focus().toggleItalic().run()} class="{btnClass(editor.isActive('italic'))} italic" title="Italic">I</button>
        <span class="w-px h-4 bg-gray-700"></span>
        <button onclick={() => editor.chain().focus().toggleHeading({ level: 1 }).run()} class={btnClass(editor.isActive('heading', { level: 1 }))} title="Heading 1">H1</button>
        <button onclick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()} class={btnClass(editor.isActive('heading', { level: 2 }))} title="Heading 2">H2</button>
        <button onclick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()} class={btnClass(editor.isActive('heading', { level: 3 }))} title="Heading 3">H3</button>
        <button onclick={() => editor.chain().focus().setParagraph().run()} class={btnClass(!editor.isActive('heading'))} title="Normal text">P</button>
        <span class="w-px h-4 bg-gray-700"></span>
        <button onclick={() => editor.chain().focus().toggleCode().run()} class="{btnClass(editor.isActive('code'))} font-mono" title="Inline code">&lt;/&gt;</button>
        <button onclick={() => editor.chain().focus().toggleCodeBlock().run()} class={btnClass(editor.isActive('codeBlock'))} title="Code block">{'{ }'}</button>
        <button onclick={() => editor.chain().focus().toggleBulletList().run()} class={btnClass(editor.isActive('bulletList'))} title="Bullet list">&#8226; List</button>
        <button onclick={() => editor.chain().focus().toggleOrderedList().run()} class={btnClass(editor.isActive('orderedList'))} title="Numbered list">1. List</button>
        <button onclick={() => editor.chain().focus().setHorizontalRule().run()} class="px-2 py-1 text-xs text-gray-400 rounded hover:bg-ck-pink/20 hover:text-white transition-colors" title="Horizontal rule">---</button>
      </div>
      {/if}

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

    <!-- Tiptap editor area -->
    <div class="flex-1 overflow-y-auto bg-ck-bg" bind:this={element}></div>

    <!-- Bottom bar -->
    <div class="flex items-center justify-between px-4 py-2 border-t border-gray-700 bg-ck-bg rounded-b-xl">
      <div class="flex items-center gap-4 text-[11px] text-ck-dim">
        <span>{charCount} characters</span>
        <span>{lineCount} lines</span>
      </div>
      <span class="text-[11px] text-ck-dim">Tiptap editor</span>
    </div>
  </div>
</div>

<style>
  :global(.ProseMirror) {
    min-height: 100%;
    outline: none;
  }
  :global(.ProseMirror h1) {
    font-size: 1.875rem;
    font-weight: 700;
    color: #FF69B4;
    margin-bottom: 0.5rem;
  }
  :global(.ProseMirror h2) {
    font-size: 1.5rem;
    font-weight: 600;
    color: #E91E8B;
    margin-bottom: 0.5rem;
  }
  :global(.ProseMirror h3) {
    font-size: 1.25rem;
    font-weight: 600;
    color: #FFD700;
    margin-bottom: 0.5rem;
  }
  :global(.ProseMirror p) {
    color: #d1d5db;
    margin-bottom: 0.5rem;
  }
  :global(.ProseMirror strong) {
    color: #ffffff;
  }
  :global(.ProseMirror code) {
    color: #FFD700;
    background: rgba(255, 215, 0, 0.1);
    padding: 0.1rem 0.3rem;
    border-radius: 0.25rem;
    font-family: monospace;
  }
  :global(.ProseMirror ul),
  :global(.ProseMirror ol) {
    padding-left: 1.5rem;
    margin-bottom: 0.5rem;
  }
  :global(.ProseMirror li) {
    color: #d1d5db;
    margin-bottom: 0.25rem;
  }
  :global(.ProseMirror hr) {
    border-color: #374151;
    margin: 1rem 0;
  }
  :global(.ProseMirror .is-editor-empty:first-child::before) {
    content: attr(data-placeholder);
    color: #6C6C6C;
    pointer-events: none;
    float: left;
    height: 0;
  }
  :global(.ProseMirror pre) {
    background: #0d1117;
    border: 1px solid #374151;
    border-radius: 0.5rem;
    padding: 0.75rem 1rem;
    margin-bottom: 0.5rem;
  }
  :global(.ProseMirror pre code) {
    background: none;
    padding: 0;
    color: #d1d5db;
  }
</style>
