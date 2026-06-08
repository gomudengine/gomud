// ScriptEditor - Monaco editor in an isolated iframe.
//
// The editor is embedded inline below the script section buttons. Clicking
// "Pop Out" moves the same iframe into a fullscreen modal. Closing the modal
// moves it back inline. One iframe instance is created per textareaId and
// reused for the lifetime of the page.
//
// The <textarea> is always the form's source of truth.
//
// Public API:
//   ScriptEditor.init(textareaId, scriptType) -> syncFn
//     syncFn()        - push textarea.value into the editor
//     syncFn(value)   - set textarea.value and push into editor
//
//   ScriptEditor.open(textareaId)       - pop out to fullscreen modal
//   ScriptEditor.getEditor(textareaId)  - returns iframe contentWindow or null

const ScriptEditor = (() => {
    'use strict';

    const MONACO_BASE = '/admin/static/js/monaco/vs';
    const FRAME_URL   = '/admin/static/js/monaco-editor-frame.html';
    const INLINE_HEIGHT = '400px';

    // textareaId -> { scriptType, iframe, iframeWin, inlineContainer, ready }
    const registry = {};

    // -------------------------------------------------------------------------
    // Public API
    // -------------------------------------------------------------------------

    function init(textareaId, scriptType) {
        const textarea = document.getElementById(textareaId);
        if (!textarea) return function () {};

        textarea.style.display = 'none';

        const rec = {
            scriptType: scriptType || null,
            iframe: null,
            iframeWin: null,
            inlineContainer: null,
            ready: false,
        };
        registry[textareaId] = rec;

        // Build the inline container that sits in the page below the button row.
        const container = document.createElement('div');
        container.style.cssText = [
            'width:100%',
            'height:' + INLINE_HEIGHT,
            'margin-top:0.5rem',
            'position:relative',
        ].join(';');
        rec.inlineContainer = container;

        // Insert after the textarea (which is hidden, sitting inside script-section).
        textarea.parentNode.insertBefore(container, textarea.nextSibling);

        // Build the iframe and place it in the inline container.
        const iframe = _buildIframe(textareaId, textarea, rec);
        rec.iframe = iframe;
        container.appendChild(iframe);

        return function syncFn(newValue) {
            if (newValue !== undefined) {
                textarea.value = newValue;
                if (rec.iframeWin) {
                    rec.iframeWin.postMessage({ type: 'monaco-set', value: newValue }, '*');
                }
            } else {
                if (rec.iframeWin) {
                    rec.iframeWin.postMessage({ type: 'monaco-set', value: textarea.value }, '*');
                }
            }
        };
    }

    function getEditor(textareaId) {
        const rec = registry[textareaId];
        return (rec && rec.iframeWin) || null;
    }

    function open(textareaId) {
        if (document.getElementById('monaco-modal-overlay')) return;
        const rec = registry[textareaId];
        if (!rec || !rec.iframe) return;

        const textarea = document.getElementById(textareaId);

        // ---- Overlay ----
        const overlay = document.createElement('div');
        overlay.id = 'monaco-modal-overlay';
        overlay.style.cssText = [
            'position:fixed', 'inset:0', 'z-index:9998',
            'background:rgba(0,0,0,0.75)',
            'display:flex', 'align-items:stretch', 'justify-content:stretch',
        ].join(';');

        // ---- Modal shell ----
        const modal = document.createElement('div');
        modal.style.cssText = [
            'display:flex', 'flex-direction:column',
            'flex:1', 'margin:1.5rem',
            'background:#1e1e1e',
            'box-shadow:0 8px 40px rgba(0,0,0,0.7)',
            'min-height:0', 'min-width:0',
        ].join(';');

        // ---- Header ----
        const header = document.createElement('div');
        header.style.cssText = [
            'display:flex', 'align-items:center', 'gap:0.75rem',
            'padding:0.45rem 0.75rem',
            'background:#2d2d2d', 'border-bottom:1px solid #444',
            'flex-shrink:0',
            'border-radius:6px 6px 0 0',
        ].join(';');

        const titleEl = document.createElement('span');
        titleEl.style.cssText = 'flex:1;font-size:0.8rem;color:#aaa;font-family:monospace;';
        titleEl.textContent = 'Script Editor';

        const addHandlerBtn = document.createElement('button');
        addHandlerBtn.type = 'button';
        addHandlerBtn.textContent = '+ Add Event Handler';
        addHandlerBtn.style.cssText = [
            'font-size:0.78rem', 'padding:0.25rem 0.7rem',
            'border:1px solid #666', 'border-radius:3px',
            'background:#3a3a3a', 'color:#ccc', 'cursor:pointer', 'flex-shrink:0',
        ].join(';');
        addHandlerBtn.addEventListener('mouseenter', function () {
            addHandlerBtn.style.background = '#555'; addHandlerBtn.style.color = '#fff';
        });
        addHandlerBtn.addEventListener('mouseleave', function () {
            addHandlerBtn.style.background = '#3a3a3a'; addHandlerBtn.style.color = '#ccc';
        });
        addHandlerBtn.addEventListener('click', function () {
            const rec = registry[textareaId];
            ScriptWizard.open({ scriptType: rec.scriptType, textareaId: textareaId });
        });

        const hintEl = document.createElement('span');
        hintEl.style.cssText = 'font-size:0.75rem;color:#666;';
        hintEl.textContent = 'Esc to close';

        const closeBtn = document.createElement('button');
        closeBtn.type = 'button';
        closeBtn.textContent = '\u2715 Close';
        closeBtn.style.cssText = [
            'font-size:0.78rem', 'padding:0.25rem 0.7rem',
            'border:1px solid #666', 'border-radius:3px',
            'background:#3a3a3a', 'color:#ccc', 'cursor:pointer', 'flex-shrink:0',
        ].join(';');
        closeBtn.addEventListener('mouseenter', function () {
            closeBtn.style.background = '#555'; closeBtn.style.color = '#fff';
        });
        closeBtn.addEventListener('mouseleave', function () {
            closeBtn.style.background = '#3a3a3a'; closeBtn.style.color = '#ccc';
        });

        header.appendChild(titleEl);
        header.appendChild(addHandlerBtn);
        header.appendChild(hintEl);
        header.appendChild(closeBtn);

        // ---- iframe mount inside modal ----
        // A plain div that the iframe will be moved into.
        // No overflow:hidden, no border-radius — avoids compositing layer
        // issues that misalign Monaco's pointer-event hit-testing.
        const mount = document.createElement('div');
        mount.style.cssText = 'flex:1;min-height:0;min-width:0;position:relative;';

        modal.appendChild(header);
        modal.appendChild(mount);
        overlay.appendChild(modal);
        document.body.appendChild(overlay);

        // Move the iframe from inline container into the modal mount.
        // Reset its size to fill the modal.
        const iframe = rec.iframe;
        iframe.style.cssText = 'width:100%;height:100%;border:none;display:block;border-radius:0 0 6px 6px;';
        mount.appendChild(iframe);

        // Tell Monaco to re-measure now that it's in a new (larger) container.
        if (rec.iframeWin) {
            rec.iframeWin.postMessage({ type: 'monaco-layout' }, '*');
        }

        // ---- Close: move iframe back inline ----
        function close() {
            // Restore inline sizing and move back.
            iframe.style.cssText = 'width:100%;height:100%;border:none;display:block;';
            rec.inlineContainer.appendChild(iframe);

            if (rec.iframeWin) {
                rec.iframeWin.postMessage({ type: 'monaco-layout' }, '*');
            }

            overlay.remove();
            document.removeEventListener('keydown', onKeyDown);
        }

        function onKeyDown(e) {
            if (e.key === 'Escape') close();
        }

        closeBtn.addEventListener('click', close);
        document.addEventListener('keydown', onKeyDown);
        overlay.addEventListener('mousedown', function (e) {
            if (e.target === overlay) close();
        });
    }

    // -------------------------------------------------------------------------
    // Internal: build the iframe
    // -------------------------------------------------------------------------

    function _buildIframe(textareaId, textarea, rec) {
        const iframe = document.createElement('iframe');
        iframe.src = FRAME_URL;
        iframe.style.cssText = 'width:100%;height:100%;border:none;display:block;';
        iframe.setAttribute('sandbox', 'allow-scripts allow-same-origin');

        // Message handler for this iframe.
        function onMessage(e) {
            if (e.source !== iframe.contentWindow) return;
            const msg = e.data;
            if (!msg || !msg.type) return;

            if (msg.type === 'monaco-ready') {
                rec.iframeWin = iframe.contentWindow;
                rec.ready = true;
            } else if (msg.type === 'monaco-change') {
                textarea.value = msg.value;
            } else if (msg.type === 'monaco-value') {
                textarea.value = msg.value;
            } else if (msg.type === 'monaco-open-wizard') {
                ScriptWizard.open({ scriptType: rec.scriptType, textareaId: textareaId });
            }
        }
        window.addEventListener('message', onMessage);

        // Once the iframe's HTML has loaded, send the init config.
        iframe.addEventListener('load', function () {
            iframe.contentWindow.postMessage({
                type: 'monaco-init',
                monacoBase: MONACO_BASE,
                scriptType: rec.scriptType,
                initialValue: textarea.value,
            }, '*');
        });

        return iframe;
    }

    // -------------------------------------------------------------------------

    return { init, getEditor, open };
})();
