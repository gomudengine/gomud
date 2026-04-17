/* global Client, VirtualWindows, injectStyles, Terminal, FitAddon */

/**
 * window-modal.js
 *
 * Generic scrollable modal overlay.
 *
 * Public API (global):
 *   GameModal.open({ title, body, format })
 *     title  — string shown in the header bar
 *     body   — content string
 *     format — "terminal" (default) renders ANSI/MUD output via xterm.js
 *              "html"     renders raw HTML inside a styled container
 *
 *   GameModal.close()
 *
 * GMCP:
 *   Responds to the "Help" namespace. Payload: { title, body, format }
 *
 * Invokable from any other JS:
 *   GameModal.open({ title: 'Help: cast', body: '...', format: 'terminal' });
 */

'use strict';

(function() {

    injectStyles(`
        /* ---- Backdrop ---- */
        #game-modal-backdrop {
            display: none;
            position: fixed;
            inset: 0;
            z-index: 10000;
            background: rgba(0, 0, 0, 0.72);
            align-items: center;
            justify-content: center;
        }

        #game-modal-backdrop.open {
            display: flex;
        }

        /* ---- Panel ---- */
        #game-modal-panel {
            position: relative;
            display: flex;
            flex-direction: column;
            width: min(780px, 92vw);
            max-height: 82vh;
            background: #111;
            border: 1px solid #1c6b60;
            border-radius: 6px;
            box-shadow: 0 8px 40px rgba(0, 0, 0, 0.9);
            overflow: hidden;
        }

        /* ---- Header ---- */
        #game-modal-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: 7px 12px 7px 14px;
            background: #0d2e28;
            border-bottom: 1px solid #1c6b60;
            flex-shrink: 0;
        }

        #game-modal-title {
            font-size: 0.88em;
            color: #3ad4b8;
            text-transform: uppercase;
            letter-spacing: 0.06em;
            font-weight: bold;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        #game-modal-close {
            background: none;
            border: none;
            color: #7ab8a0;
            font-size: 1.1em;
            cursor: pointer;
            padding: 0 2px;
            line-height: 1;
            flex-shrink: 0;
            transition: color 0.15s;
        }

        #game-modal-close:hover {
            color: #dffbd1;
        }

        /* ---- Body ---- */
        #game-modal-body {
            flex: 1;
            overflow: hidden;
            display: flex;
            flex-direction: column;
            min-height: 0;
        }

        /* Terminal mode: xterm.js fills the body */
        #game-modal-term-container {
            flex: 1;
            overflow: hidden;
            padding: 8px 6px 6px;
            box-sizing: border-box;
        }

        #game-modal-term-container .xterm-viewport::-webkit-scrollbar       { width: 6px; }
        #game-modal-term-container .xterm-viewport::-webkit-scrollbar-track  { background: #1a1a1a; }
        #game-modal-term-container .xterm-viewport::-webkit-scrollbar-thumb  { background: #1c6b60; border-radius: 3px; }
        #game-modal-term-container .xterm-viewport::-webkit-scrollbar-thumb:hover { background: #3ad4b8; }
        #game-modal-term-container .xterm-viewport { scrollbar-width: thin; scrollbar-color: #1c6b60 #1a1a1a; }

        /* HTML mode: scrollable prose container */
        #game-modal-html-container {
            flex: 1;
            overflow-y: auto;
            padding: 12px 16px;
            color: #dffbd1;
            font-size: 0.84em;
            line-height: 1.6;
        }

        #game-modal-html-container::-webkit-scrollbar       { width: 5px; }
        #game-modal-html-container::-webkit-scrollbar-track  { background: #111; }
        #game-modal-html-container::-webkit-scrollbar-thumb  { background: #1c6b60; border-radius: 3px; }
        #game-modal-html-container::-webkit-scrollbar-thumb:hover { background: #3ad4b8; }

        /* ---- Footer hint ---- */
        #game-modal-footer {
            flex-shrink: 0;
            padding: 4px 14px 5px;
            border-top: 1px solid #0f3333;
            font-size: 0.66em;
            color: #3a5a52;
            text-align: right;
        }
    `);

    // -----------------------------------------------------------------------
    // DOM — built once on DOMContentLoaded, hidden until opened
    // -----------------------------------------------------------------------
    let backdrop, panel, titleEl, closeBtn, termContainer, htmlContainer;
    let modalTerm     = null;
    let modalFitAddon = null;

    function _buildDOM() {
        backdrop = document.createElement('div');
        backdrop.id = 'game-modal-backdrop';

        panel = document.createElement('div');
        panel.id = 'game-modal-panel';

        const header = document.createElement('div');
        header.id = 'game-modal-header';

        titleEl = document.createElement('span');
        titleEl.id = 'game-modal-title';
        titleEl.textContent = '';

        closeBtn = document.createElement('button');
        closeBtn.id          = 'game-modal-close';
        closeBtn.textContent = '\u00d7';
        closeBtn.setAttribute('aria-label', 'Close');

        header.appendChild(titleEl);
        header.appendChild(closeBtn);

        const body = document.createElement('div');
        body.id = 'game-modal-body';

        termContainer = document.createElement('div');
        termContainer.id = 'game-modal-term-container';

        htmlContainer = document.createElement('div');
        htmlContainer.id = 'game-modal-html-container';

        body.appendChild(termContainer);
        body.appendChild(htmlContainer);

        const footer = document.createElement('div');
        footer.id          = 'game-modal-footer';
        footer.textContent = 'Press Esc or click outside to close';

        panel.appendChild(header);
        panel.appendChild(body);
        panel.appendChild(footer);
        backdrop.appendChild(panel);
        document.body.appendChild(backdrop);

        closeBtn.addEventListener('click', close);

        backdrop.addEventListener('click', function(e) {
            if (e.target === backdrop) { close(); }
        });
    }

    function ensureTerminal() {
        if (modalTerm) { return; }
        modalTerm = new Terminal({
            cols:            80,
            rows:            24,
            cursorBlink:     false,
            disableStdin:    true,
            scrollback:      2000,
            fontSize:        14,
            theme: {
                background:  '#111111',
                foreground:  '#dffbd1',
                cursor:      '#3ad4b8',
                black:       '#1e1e1e',
                red:         '#e06060',
                green:       '#3ad4b8',
                yellow:      '#d4a843',
                blue:        '#7ab8a0',
                magenta:     '#c06090',
                cyan:        '#3ad4b8',
                white:       '#dffbd1',
                brightBlack: '#555',
                brightWhite: '#ffffff',
            },
        });
        modalFitAddon = new FitAddon.FitAddon();
        modalTerm.loadAddon(modalFitAddon);
        modalTerm.open(termContainer);
    }

    // -----------------------------------------------------------------------
    // Open / close
    // -----------------------------------------------------------------------
    function open(opts) {
        opts = opts || {};
        const title   = opts.title  || '';
        const content = opts.body   || '';
        const format  = (opts.format || 'terminal').toLowerCase();

        titleEl.textContent = title;

        if (format === 'html') {
            termContainer.style.display = 'none';
            htmlContainer.style.display = '';
            htmlContainer.innerHTML     = content;
            htmlContainer.scrollTop     = 0;
        } else {
            htmlContainer.style.display = 'none';
            termContainer.style.display = '';
        }

        // Show the backdrop first so the terminal container has layout dimensions
        backdrop.classList.add('open');

        if (format !== 'html') {
            requestAnimationFrame(function() {
                ensureTerminal();
                modalTerm.reset();
                modalFitAddon.fit();

                const lines = content.split('\n');
                lines.forEach(function(line) {
                    modalTerm.writeln(line.replace(/\r$/, ''));
                });

                // Double rAF: let xterm finish its own post-write scroll before
                // we force back to the top.
                requestAnimationFrame(function() {
                    modalTerm.scrollToTop();
                });
            });
        }
    }

    function close() {
        backdrop.classList.remove('open');
    }

    // -----------------------------------------------------------------------
    // Public API — exposed before DOMContentLoaded so callers can queue calls
    // -----------------------------------------------------------------------
    window.GameModal = { open: open, close: close };

    // -----------------------------------------------------------------------
    // Init on DOMContentLoaded — document.body is guaranteed to exist here
    // -----------------------------------------------------------------------
    document.addEventListener('DOMContentLoaded', function() {
        _buildDOM();

        document.addEventListener('keydown', function(e) {
            if (e.key === 'Escape' && backdrop.classList.contains('open')) {
                close();
            }
        });

        VirtualWindows.register({
            window:       null,
            gmcpHandlers: ['Help'],
            onGMCP: function(namespace, payload) {
                if (!payload) { return; }
                open({
                    title:  payload.title  || 'Help',
                    body:   payload.body   || '',
                    format: payload.format || 'terminal',
                });
            },
        });
    });

})();
