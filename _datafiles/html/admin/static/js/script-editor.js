// ScriptEditor - wraps a <textarea> with a highlight.js overlay for live
// syntax highlighting. Requires highlight.js to be loaded globally (window.hljs).
//
// Usage:
//   const sync = ScriptEditor.init('f-script', 'item');
//   // Call sync() after programmatically changing the textarea's value:
//   document.getElementById('f-script').value = newCode;
//   sync();
//
// The optional second argument is a script type string ('room', 'mob', 'item',
// 'spell', 'buff', 'pet'). When provided, the editor fetches the predefined
// engine function names for that type and highlights them distinctly from
// user-defined functions.

const ScriptEditor = (() => {
    'use strict';

    let stylesInjected = false;

    function injectStyles() {
        if (stylesInjected) return;
        stylesInjected = true;
        const s = document.createElement('style');
        s.textContent = `
            .script-editor { position: relative; }
            .script-editor pre {
                position: absolute; top: 0; left: 0; right: 0; bottom: 0;
                margin: 0; pointer-events: none; overflow: hidden;
                border: 1px solid transparent; border-radius: 4px;
                padding: 0.6rem 0.75rem;
                font-family: monospace; font-size: 0.82rem; line-height: 1.55;
            }
            .script-editor pre .hljs-engine-fn {
                pointer-events: auto;
                cursor: help;
            }
            .script-editor pre code.hljs {
                padding: 0; background: transparent;
                font-family: inherit; font-size: inherit; line-height: inherit;
                display: block; white-space: pre;
                overflow: visible;
            }
            .script-editor textarea {
                -webkit-text-fill-color: transparent;
                color: transparent !important;
                caret-color: var(--color-code-text);
                white-space: pre; overflow-x: auto;
            }
            .script-editor textarea::selection {
                -webkit-text-fill-color: var(--color-code-text);
                background: rgba(128, 180, 255, 0.35);
            }
            .script-editor textarea::placeholder {
                -webkit-text-fill-color: var(--color-text-subtle);
                color: var(--color-text-subtle);
            }
            /* Dark syntax colors (GitHub-dark inspired) */
            .script-editor .hljs { background: transparent; color: var(--color-code-text); }
            .script-editor .hljs-comment,
            .script-editor .hljs-quote { color: var(--color-code-comment); font-style: italic; }
            .script-editor .hljs-keyword,
            .script-editor .hljs-selector-tag { color: var(--color-code-keyword); font-weight: 700; }
            .script-editor .hljs-string,
            .script-editor .hljs-template-tag,
            .script-editor .hljs-deletion { color: var(--color-api-str); }
            .script-editor .hljs-number,
            .script-editor .hljs-literal { color: var(--color-api-kw); }
            .script-editor .hljs-built_in,
            .script-editor .hljs-title,
            .script-editor .hljs-section { color: var(--color-api-flag); }
            .script-editor .hljs-variable,
            .script-editor .hljs-template-variable,
            .script-editor .hljs-symbol { color: var(--color-code-symbol); }
            .script-editor .hljs-name,
            .script-editor .hljs-attribute,
            .script-editor .hljs-attr { color: var(--color-api-kw); }
            .script-editor .hljs-regexp,
            .script-editor .hljs-link { color: var(--color-code-link); }
            .script-editor .hljs-meta { color: var(--color-api-kw); }
            .script-editor .hljs-punctuation { color: var(--color-code-text); }
            .script-editor .hljs-addition { color: var(--color-code-addition); }
            .script-editor .hljs-operator { color: var(--color-code-keyword); }
            .script-editor .hljs-params { color: var(--color-code-text); }
            .script-editor .hljs-property { color: var(--color-api-kw); }
            .script-editor .hljs-type,
            .script-editor .hljs-selector-class,
            .script-editor .hljs-selector-id { color: var(--color-code-symbol); }
            /* Engine-defined callback functions (predefined by the scripting system) */
            .script-editor .hljs-engine-fn {
                color: var(--color-code-engine-fn); font-weight: 700;
                text-decoration: underline dashed var(--color-code-engine-fn);
                text-underline-offset: 3px;
            }
            /* Tooltip for engine functions */
            .script-editor-tooltip {
                position: fixed; z-index: 9999;
                background: #1a1a2e;
                color: #c9d1d9;
                border: 1px solid var(--color-code-engine-fn);
                border-radius: 4px;
                padding: 3px 8px;
                font-family: sans-serif; font-size: 0.75rem; font-weight: normal;
                white-space: nowrap; pointer-events: none;
                opacity: 0; transition: opacity 0.1s;
            }
            .script-editor-tooltip.visible { opacity: 1; }
        `;
        document.head.appendChild(s);
    }

    // Post-processes the raw highlighted HTML string to wrap predefined engine
    // function names with the .hljs-engine-fn class. highlight.js emits function
    // definition names as <span class="hljs-title function_"> (the scope
    // "title.function" maps to two classes: hljs-title and function_). Only
    // names in that span are promoted so call-site references, strings, and
    // comments that happen to contain the same text are left alone.
    function applyEngineFnHighlight(html, nameSet) {
        return html.replace(
            /<span class="hljs-title function_">([^<]*)<\/span>/g,
            function (match, name) {
                if (nameSet.has(name.trim())) {
                    return '<span class="hljs-engine-fn">' + name + '</span>';
                }
                return match;
            }
        );
    }

    function init(textareaId, scriptType) {
        const textarea = document.getElementById(textareaId);
        if (!textarea || !window.hljs) return function () {};

        injectStyles();
        textarea.setAttribute('wrap', 'off');

        const wrapper = document.createElement('div');
        wrapper.className = 'script-editor';
        textarea.parentNode.insertBefore(wrapper, textarea);
        wrapper.appendChild(textarea);

        const pre = document.createElement('pre');
        pre.setAttribute('aria-hidden', 'true');
        const code = document.createElement('code');
        code.className = 'language-javascript hljs';
        pre.appendChild(code);
        wrapper.insertBefore(pre, textarea);

        // Tooltip element shared across all engine-fn spans in this editor.
        const tooltip = document.createElement('div');
        tooltip.className = 'script-editor-tooltip';
        tooltip.textContent = 'Engine Function';
        document.body.appendChild(tooltip);

        pre.addEventListener('mouseover', function (e) {
            if (e.target.classList.contains('hljs-engine-fn')) {
                tooltip.classList.add('visible');
            }
        });
        pre.addEventListener('mousemove', function (e) {
            if (e.target.classList.contains('hljs-engine-fn')) {
                tooltip.style.left = e.clientX + 'px';
                tooltip.style.top  = (e.clientY - tooltip.offsetHeight - 8) + 'px';
            }
        });
        pre.addEventListener('mouseout', function (e) {
            if (e.target.classList.contains('hljs-engine-fn')) {
                tooltip.classList.remove('visible');
            }
        });
        // Clicks on the pre layer (engine-fn spans) must not swallow focus.
        // Forward the mousedown to the textarea so the caret is placed correctly.
        pre.addEventListener('mousedown', function (e) {
            e.preventDefault();
            textarea.focus();
        });

        let engineFnNames = null; // Set<string> once loaded, null until then
        let dirty = false;

        function highlight(src) {
            let html = hljs.highlight(src, { language: 'javascript' }).value;
            if (engineFnNames) {
                html = applyEngineFnHighlight(html, engineFnNames);
            }
            return html;
        }

        function sync() {
            dirty = false;
            if (!textarea.value) {
                code.innerHTML = '';
            } else {
                code.innerHTML = highlight(textarea.value) + '\n';
            }
            pre.scrollTop = textarea.scrollTop;
            pre.scrollLeft = textarea.scrollLeft;
        }

        textarea.addEventListener('input', function () { dirty = true; sync(); });
        textarea.addEventListener('scroll', function () {
            pre.scrollTop = textarea.scrollTop;
            pre.scrollLeft = textarea.scrollLeft;
        });

        textarea.addEventListener('keydown', function (e) {
            if (e.key === 'Tab') {
                e.preventDefault();
                var start = this.selectionStart;
                var end = this.selectionEnd;
                this.value = this.value.substring(0, start) + '    ' + this.value.substring(end);
                this.selectionStart = this.selectionEnd = start + 4;
                sync();
            }
        });

        // If a script type was provided, fetch the schema and extract the
        // predefined function names for that type. Re-sync once loaded so
        // any pre-filled script content is immediately highlighted correctly.
        if (scriptType) {
            AdminAPI.get('/admin/api/v1/scripting/functions').then(function (res) {
                if (!res.ok) return;
                const typeDef = res.data && res.data.data && res.data.data.scriptTypes && res.data.data.scriptTypes[scriptType];
                if (!typeDef || !typeDef.functions) return;
                const names = new Set();
                typeDef.functions.forEach(function (fn) {
                    // For dynamic functions like onCommand_{command} the base
                    // name contains a placeholder; skip those since the actual
                    // runtime name is user-defined.
                    if (!fn.dynamic) {
                        names.add(fn.name);
                    }
                });
                engineFnNames = names;
                sync();
            }).catch(function () { /* schema fetch failure is non-fatal */ });
        }

        // Re-apply highlighting every 3 seconds if the content has changed
        // since the last sync. This ensures engine-function colours are
        // applied after the schema fetch resolves even if the user typed
        // before it finished loading.
        setInterval(function () {
            if (dirty) sync();
        }, 3000);

        sync();
        return sync;
    }

    return { init };
})();
