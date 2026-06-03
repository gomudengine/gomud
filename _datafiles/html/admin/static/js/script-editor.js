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
                white-space: pre-wrap; max-width: 420px; word-break: break-word;
                pointer-events: none;
                opacity: 0; transition: opacity 0.1s;
            }
            .script-editor-tooltip.visible { opacity: 1; }
        `;
        document.head.appendChild(s);
    }

    // Post-processes the raw highlighted HTML string to wrap predefined engine
    // function names with the .hljs-engine-fn class.
    //
    // highlight.js uses two different span classes depending on how the identifier
    // is cased and used:
    //   - "hljs-title function_"  : function definitions AND lowercase-starting call sites
    //   - "hljs-title class_"     : PascalCase/UpperCamelCase call sites (e.g. GetRoom, RandInt)
    //
    // defNames covers event callback definitions; callNames covers global engine
    // functions that scripts call. dynamicPrefixes is an array of strings for
    // dynamic functions like onCommand_ whose suffix is user-defined.
    // All receive the same visual treatment.
    function applyEngineFnHighlight(html, defNames, callNames, dynamicPrefixes) {
        return html.replace(
            /<span class="hljs-title (?:function_|class_)">([^<]*)<\/span>/g,
            function (match, name) {
                const trimmed = name.trim();
                if ((defNames && defNames.has(trimmed)) || (callNames && callNames.has(trimmed))) {
                    return '<span class="hljs-engine-fn" data-engine-fn="' + trimmed + '">' + name + '</span>';
                }
                if (dynamicPrefixes) {
                    for (let i = 0; i < dynamicPrefixes.length; i++) {
                        const p = dynamicPrefixes[i];
                        if (trimmed.length > p.prefix.length && trimmed.indexOf(p.prefix) === 0) {
                            // data-engine-fn holds the concrete name (e.g. onCommand_pull);
                            // data-engine-schema holds the schema key for meta lookup.
                            return '<span class="hljs-engine-fn" data-engine-fn="' + trimmed + '" data-engine-schema="' + p.schemaName + '">' + name + '</span>';
                        }
                    }
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
        document.body.appendChild(tooltip);

        function buildTooltipContent(el) {
            const fnName = el.dataset.engineFn;
            const schemaName = el.dataset.engineSchema || fnName;
            // Check global engine functions.
            if (fnName && engineCallMeta && engineCallMeta[fnName]) {
                const meta = engineCallMeta[fnName];
                const paramStr = (meta.params || []).map(function (p) {
                    return p.name + ': ' + p.type;
                }).join(', ');
                const retStr = meta.returnType ? ' \u2192 ' + meta.returnType : '';
                let content = meta.name + '(' + paramStr + ')' + retStr;
                if (meta.description) { content += '\n' + meta.description; }
                return content;
            }
            // Check event callbacks (static and dynamic). Dynamic entries use the
            // concrete function name (fnName) in the signature but look up metadata
            // via the schema name (schemaName, e.g. onCommand_{command}).
            if (schemaName && engineFnMeta && engineFnMeta[schemaName]) {
                const meta = engineFnMeta[schemaName];
                const paramStr = (meta.params || []).map(function (p) {
                    return p.name + ': ' + p.type;
                }).join(', ');
                let content = fnName + '(' + paramStr + ')';
                if (meta.description) { content += '\n' + meta.description; }
                return content;
            }
            return 'Engine Function';
        }

        pre.addEventListener('mouseover', function (e) {
            if (e.target.classList.contains('hljs-engine-fn')) {
                tooltip.textContent = buildTooltipContent(e.target);
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

        let engineFnNames = null;    // Set<string>: static event callback definition names
        let engineFnMeta = null;      // object: schemaName -> ScriptFuncDef (for static callbacks)
        let engineDynamic = null;     // array: [{prefix, schemaName}] for dynamic callbacks
        let engineCallNames = null;   // Set<string>: global engine function names
        let engineCallMeta = null;    // object: name -> EngineGlobalFuncDef metadata
        let dirty = false;

        function highlight(src) {
            let html = hljs.highlight(src, { language: 'javascript' }).value;
            if (engineFnNames || engineCallNames || engineDynamic) {
                html = applyEngineFnHighlight(html, engineFnNames, engineCallNames, engineDynamic);
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
                const schema = res.data && res.data.data;

                // Event callback names for this script type (function definitions).
                const typeDef = schema && schema.scriptTypes && schema.scriptTypes[scriptType];
                if (typeDef && typeDef.functions) {
                    const names = new Set();
                    const meta = {};
                    const dynPrefixes = [];
                    typeDef.functions.forEach(function (fn) {
                        if (fn.dynamic) {
                            // Dynamic entries like onCommand_{command}: extract the
                            // literal prefix before the placeholder and match any
                            // function name that starts with it.
                            const placeholder = fn.dynamic.placeholder; // e.g. "{command}"
                            const idx = fn.name.indexOf(placeholder);
                            if (idx > 0) {
                                const prefix = fn.name.slice(0, idx); // e.g. "onCommand_"
                                dynPrefixes.push({ prefix: prefix, schemaName: fn.name });
                                meta[fn.name] = fn; // keyed by schema name for tooltip lookup
                            }
                        } else {
                            names.add(fn.name);
                            meta[fn.name] = fn;
                        }
                    });
                    engineFnNames = names;
                    engineFnMeta = meta;
                    engineDynamic = dynPrefixes.length > 0 ? dynPrefixes : null;
                }

                // Global engine function names and their metadata (call sites).
                if (schema && schema.engineFunctions) {
                    const callNames = new Set();
                    const metaMap = {};
                    schema.engineFunctions.forEach(function (fn) {
                        callNames.add(fn.name);
                        metaMap[fn.name] = fn;
                    });
                    engineCallNames = callNames;
                    engineCallMeta = metaMap;
                }

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
