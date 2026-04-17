/* global Client, VirtualWindow, VirtualWindows, injectStyles */
/**
 * window-debug-log.js
 *
 * Virtual window: Debug Log
 *
 * Receives GMCP:
 *   Comm - incoming channel message
 *
 * Prints raw structured output to a scrolling monospace log window.
 */

'use strict';

(function() {

    injectStyles(`
        #debug-log-output {
            width: 100%;
            height: 100%;
            display: flex;
            flex-direction: column;
            background: #1e1e1e;
            color: #cfcfcf;
            font-family: monospace;
            font-size: 0.75em;
            line-height: 1.35em;
            overflow: hidden;
        }

        #debug-log-stream {
            flex: 1;
            overflow-y: auto;
            padding: 6px;
            box-sizing: border-box;
            white-space: pre-wrap;
            word-break: break-word;
        }

        .log-line {
            margin: 0;
            padding: 0;
        }

        .log-text {
            color: #dddddd;
            white-space: pre-wrap;

        }
    `);

    // -----------------------------------------------------------------------
    // DOM
    // -----------------------------------------------------------------------
    function createDOM() {
        const root = document.createElement('div');
        root.id = 'debug-log-output';

        const stream = document.createElement('div');
        stream.id = 'debug-log-stream';

        root.appendChild(stream);
        document.body.appendChild(root);

        return root;
    }

    // -----------------------------------------------------------------------
    // Window
    // -----------------------------------------------------------------------
    const win = new VirtualWindow('Debug Log', {
        dock: 'right',
        defaultDocked: true,
        dockedHeight: 500,
        offOnLoad: true,
        factory() {
            const el = createDOM();
            return {
                title: 'Debug Log',
                mount: el,
                background: '#1e1e1e',
                border: 1,
                x: 'right',
                y: 450,
                width: 420,
                height: 320,
                header: 20,
                bottom: 0,
            };
        },
    });

    // -----------------------------------------------------------------------
    // Logging
    // -----------------------------------------------------------------------
    function logMessage(namespace) {
        const stream = document.getElementById('debug-log-stream');
        if (!stream) return;

        const line = document.createElement('div');
        line.className = 'log-line';

        const payload = JSON.stringify(Client.GetGMCP(namespace), null, 2);

        line.innerHTML =
            `<span class="log-text">[GMCP] ${namespace} ${payload}</span>`;

        stream.appendChild(line);

        // Auto-scroll
        stream.scrollTop = stream.scrollHeight;

        // Trim old logs if too large
        while (stream.childElementCount > 500) {
            stream.removeChild(stream.firstElementChild);
        }
    }

    // -----------------------------------------------------------------------
    // GMCP update
    // -----------------------------------------------------------------------
    function updateLog(namespace) {
        win.open();
        if (!win.isOpen()) return;
        logMessage(namespace);
    }

    // -----------------------------------------------------------------------
    // Registration
    // -----------------------------------------------------------------------
    VirtualWindows.register({
        window: win,
        gmcpHandlers: ['*'],
        onGMCP(namespace) {
            updateLog(namespace);
        },
    });

})();
