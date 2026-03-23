<script>
    import { onMount, onDestroy } from 'svelte'
    import { Terminal } from '@xterm/xterm'
    import { FitAddon } from '@xterm/addon-fit'
    import { WebLinksAddon } from '@xterm/addon-web-links'
    import { currentProject } from '../stores/project.js'
    import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js'

    let terminalEl = $state(null)
    let term = null
    let fitAddon = null
    let running = $state(false)
    let projectPath = $state('')
    let projectName = $state('')
    let sessionCount = $state(0)
    let sessionInterval = null

    currentProject.subscribe(async (p) => {
        const newPath = p?.path || ''
        projectPath = newPath
        projectName = p?.name || ''

        if (!newPath || !term) return

        // Clear display for switch
        term.clear()

        try {
            const { SwitchTo, IsRunning } = await import('../../wailsjs/go/main/TerminalService.js')
            const [hasSession] = await SwitchTo(newPath)
            running = await IsRunning()
            if (hasSession) {
                term.writeln(`\x1b[90m— Resumed session: ${p.name} —\x1b[0m`)
            } else {
                // No existing session — auto-start
                setTimeout(() => startSession(), 300)
            }
        } catch {
            running = false
        }
    })

    onMount(() => {
        term = new Terminal({
            theme: {
                background: '#0d1117',
                foreground: '#d1d5db',
                cursor: '#FF69B4',
                cursorAccent: '#0d1117',
                selectionBackground: '#E91E8B40',
                black: '#1a1a2e',
                red: '#FF5555',
                green: '#00FF87',
                yellow: '#FFD700',
                blue: '#6C6C6C',
                magenta: '#FF69B4',
                cyan: '#E91E8B',
                white: '#FAFAFA',
            },
            fontFamily: 'Menlo, Monaco, "Courier New", monospace',
            fontSize: 13,
            lineHeight: 1.4,
            cursorBlink: true,
            cursorStyle: 'bar',
            scrollback: 10000,
        })

        fitAddon = new FitAddon()
        term.loadAddon(fitAddon)

        // Let Cmd+key shortcuts pass through to the global handler
        term.attachCustomKeyEventHandler((e) => {
            if ((e.metaKey || e.ctrlKey) && ['s','d','f','j','k','e','ArrowLeft','ArrowRight'].includes(e.key)) {
                return false // Don't let xterm handle it — let it bubble to window
            }
            return true
        })
        term.loadAddon(new WebLinksAddon())

        term.open(terminalEl)
        fitAddon.fit()

        // Send keystrokes to Go PTY
        term.onData(async (data) => {
            try {
                const { Write } = await import('../../wailsjs/go/main/TerminalService.js')
                await Write(data)
            } catch {}
        })

        // Resize PTY when terminal resizes
        term.onResize(async ({ cols, rows }) => {
            try {
                const { Resize } = await import('../../wailsjs/go/main/TerminalService.js')
                await Resize(cols, rows)
            } catch {}
        })

        // Listen for output from Go PTY
        EventsOn('terminal:output', (data) => {
            term.write(data)
        })

        EventsOn('terminal:exit', () => {
            running = false
            term.writeln('\r\n\x1b[90m--- Session ended ---\x1b[0m')
        })

        // Handle window resize
        const resizeObserver = new ResizeObserver(() => {
            if (fitAddon) fitAddon.fit()
        })
        resizeObserver.observe(terminalEl)

        // Auto-start if a project is selected
        if (projectPath && !running) {
            setTimeout(() => startSession(), 300)
        }

        // Refresh session count periodically
        sessionInterval = setInterval(async () => {
            try {
                const { ListSessions } = await import('../../wailsjs/go/main/TerminalService.js')
                const sessions = await ListSessions()
                sessionCount = sessions?.length || 0
            } catch {}
        }, 2000)

        return () => {
            resizeObserver.disconnect()
        }
    })

    onDestroy(() => {
        EventsOff('terminal:output')
        EventsOff('terminal:exit')
        if (sessionInterval) clearInterval(sessionInterval)
        if (term) term.dispose()
    })

    async function startSession() {
        if (!projectPath) {
            term.writeln('\x1b[31mNo project selected. Select a project first.\x1b[0m')
            return
        }
        try {
            const { Start } = await import('../../wailsjs/go/main/TerminalService.js')
            // Pass terminal dimensions so PTY matches xterm.js size
            const cols = term?.cols || 120
            const rows = term?.rows || 40
            await Start(projectPath, cols, rows)
            running = true
            term.clear()
            term.focus()
        } catch (e) {
            term.writeln(`\x1b[31mFailed to start: ${e}\x1b[0m`)
        }
    }

    async function stopSession() {
        try {
            const { Stop } = await import('../../wailsjs/go/main/TerminalService.js')
            await Stop()
        } catch {}
        running = false
    }
</script>

<div class="flex flex-col h-full">
    <!-- Header -->
    <div class="flex items-center justify-between px-4 py-2 border-b border-gray-800 shrink-0">
        <div class="flex items-center gap-3">
            <h1 class="text-sm font-semibold text-white">Claude Chat</h1>
            {#if projectName}
                <span class="text-xs text-ck-dim font-mono">{projectName}</span>
            {/if}
            {#if running}
                <span class="flex items-center gap-1.5 text-xs text-ck-green">
                    <span class="w-1.5 h-1.5 rounded-full bg-ck-green animate-pulse"></span>
                    Running
                </span>
            {:else}
                <span class="text-xs text-ck-dim">Not connected</span>
            {/if}
            {#if sessionCount > 1}
                <span class="text-xs text-ck-dim">{sessionCount} sessions</span>
            {/if}
        </div>
        <div class="flex items-center gap-2">
            {#if running}
                <button
                    onclick={stopSession}
                    class="px-3 py-1 text-xs rounded bg-red-500/20 text-red-400 border border-red-500/30 hover:bg-red-500/30 transition-colors"
                >
                    Stop
                </button>
            {:else}
                <button
                    onclick={startSession}
                    class="px-3 py-1 text-xs rounded bg-ck-green/20 text-ck-green border border-ck-green/30 hover:bg-ck-green/30 transition-colors"
                >
                    Start Claude
                </button>
            {/if}
        </div>
    </div>

    <!-- Terminal -->
    <div
        bind:this={terminalEl}
        class="flex-1 p-1"
        style="background: #0d1117;"
    ></div>
</div>

<style>
    :global(.xterm) {
        height: 100%;
    }
    :global(.xterm-viewport) {
        scrollbar-width: thin;
        scrollbar-color: #333 #0d1117;
    }
</style>
