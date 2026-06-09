// monaco-editor-frame.js
// Runs inside the isolated iframe. Receives configuration and content from
// the parent page via postMessage, hosts the Monaco editor, and posts changes
// back to the parent.

(function () {
    'use strict';

    var editor = null;
    var monacoBase = null;
    var scriptType = null;
    var pendingMessages = [];

    // -------------------------------------------------------------------------
    // Bootstrap: wait for the parent to send config before loading Monaco
    // -------------------------------------------------------------------------

    window.addEventListener('message', function (e) {
        if (!e.data || !e.data.type) return;

        if (e.data.type === 'monaco-init') {
            monacoBase = e.data.monacoBase;
            scriptType = e.data.scriptType || '';
            loadMonaco(e.data.initialValue || '');
            return;
        }

        // Queue messages that arrive before the editor is ready
        if (!editor) {
            pendingMessages.push(e.data);
            return;
        }

        handleMessage(e.data);
    });

    function handleMessage(msg) {
        if (msg.type === 'monaco-set') {
            if (editor.getValue() !== msg.value) {
                editor.setValue(msg.value);
            }

        } else if (msg.type === 'monaco-get') {
            parent.postMessage({ type: 'monaco-value', value: editor.getValue() }, '*');

        } else if (msg.type === 'monaco-layout') {
            editor.layout();

        } else if (msg.type === 'monaco-insert') {
            insertAtEnd(msg.stub);

        } else if (msg.type === 'monaco-jump-or-insert') {
            var model = editor.getModel();
            var text = model.getValue();
            var escaped = msg.fnName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
            var m = new RegExp('(^|\n)[ \t]*function\\s+' + escaped + '\\s*\\(').exec(text);
            if (m) {
                var lineNum = text.substring(0, m.index + m[0].indexOf('function')).split('\n').length;
                editor.revealLineInCenter(lineNum);
                editor.setPosition({ lineNumber: lineNum, column: 1 });
                editor.focus();
            } else {
                insertAtEnd(msg.stub);
            }
        }
    }

    function insertAtEnd(stub) {
        var model = editor.getModel();
        var lineCount = model.getLineCount();
        var lastLine = model.getLineContent(lineCount);
        var prefix = (lastLine.trim().length > 0 ? '\n' : '') + '\n';
        editor.executeEdits('script-wizard', [{
            range: new monaco.Range(lineCount, lastLine.length + 1, lineCount, lastLine.length + 1),
            text: prefix + stub,
            forceMoveMarkers: true
        }]);
        editor.revealLine(model.getLineCount());
        editor.focus();
    }

    // -------------------------------------------------------------------------
    // Monaco loading
    // -------------------------------------------------------------------------

    function loadMonaco(initialValue) {
        var loaderScript = document.createElement('script');
        loaderScript.src = monacoBase + '/loader.js';
        loaderScript.onload = function () {
            require.config({
                paths: { vs: monacoBase },
                'vs/css': { disabled: true }
            });
            window.MonacoEnvironment = {
                getWorkerUrl: function () {
                    return window.location.origin + '/admin/static/js/monaco/vs/base/worker/workerMain.js';
                }
            };
            require(['vs/editor/editor.main'], function () {
                createEditor(initialValue);
            });
        };
        document.head.appendChild(loaderScript);
    }

    function createEditor(initialValue) {
        editor = monaco.editor.create(document.getElementById('editor'), {
            value: initialValue,
            language: 'javascript',
            theme: 'vs-dark',
            automaticLayout: true,
            minimap: { enabled: false },
            fontSize: 13,
            lineNumbers: 'on',
            scrollBeyondLastLine: false,
            tabSize: 4,
            insertSpaces: true,
            wordWrap: 'off',
            renderWhitespace: 'none',
            folding: true,
            glyphMargin: false,
            overviewRulerLanes: 0,
            fixedOverflowWidgets: true
        });

        // Register a custom action so "Add Event Handler" appears in the
        // right-click context menu and F1 Command Palette.
        editor.addAction({
            id: 'gomud-add-event-handler',
            label: 'Add Event Handler',
            contextMenuGroupId: 'navigation',
            contextMenuOrder: 1,
            run: function () {
                parent.postMessage({ type: 'monaco-open-wizard' }, '*');
            }
        });

        // Notify parent of content changes
        editor.onDidChangeModelContent(function () {
            parent.postMessage({ type: 'monaco-change', value: editor.getValue() }, '*');
        });

        // Apply intellisense
        if (scriptType) {
            applyIntellisense(scriptType);
        }

        // Drain any messages that arrived before the editor was ready
        pendingMessages.forEach(function (msg) { handleMessage(msg); });
        pendingMessages = [];

        editor.focus();
        parent.postMessage({ type: 'monaco-ready' }, '*');
    }

    function applyIntellisense(type) {
        // Configure the JavaScript language service for a clean engine-only
        // environment. We want full semantic analysis (for hover, signature help,
        // and go-to-definition) but without browser/DOM globals polluting
        // completions. Setting lib:[] gives an empty standard lib; our .d.ts
        // then provides the only globals Monaco knows about.
        monaco.languages.typescript.javascriptDefaults.setCompilerOptions({
            allowNonTsExtensions: true,
            allowJs: true,
            checkJs: true,
            target: monaco.languages.typescript.ScriptTarget.ES2015,
            lib: []
        });
        // Keep all diagnostics on so hover info and signature help work.
        // Only suppress semantic errors that would fire on unknown globals
        // (the engine injects many globals at runtime that TypeScript can't see).
        monaco.languages.typescript.javascriptDefaults.setDiagnosticsOptions({
            noSemanticValidation: false,
            noSyntaxValidation: false,
            diagnosticCodesToIgnore: [2304, 2339, 2349, 2540]
        });
        fetch('/admin/api/v1/scripting/types.d.ts?type=' + encodeURIComponent(type), { credentials: 'include' })
            .then(function (r) { return r.text(); })
            .then(function (dts) {
                monaco.languages.typescript.javascriptDefaults.addExtraLib(dts, 'file:///gomud-engine.d.ts');
            })
            .catch(function () {});
    }

})();
