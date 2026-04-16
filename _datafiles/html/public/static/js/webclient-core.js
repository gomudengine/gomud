/**
 * webclient-core.js
 *
 * Core infrastructure for the GoMud web client. Provides:
 *   - Client namespace (shared state accessible by window modules)
 *   - VirtualWindow class (lifecycle management for WinBox panels)
 *   - VirtualWindows registry (GMCP handler dispatch)
 *   - WebSocket connection management
 *   - Terminal (xterm.js) setup
 *   - MSP audio (music + sound)
 *   - Volume slider UI
 *
 * Window modules call VirtualWindows.register(...) to add themselves.
 * The HTML file calls Client.init() on page load.
 */

'use strict';

// ---------------------------------------------------------------------------
// injectStyles
//
// Appends a <style> block to <head>. Called by window modules at load time
// so each module owns and ships its own CSS alongside its JS.
// ---------------------------------------------------------------------------
function injectStyles(css) {
    const style = document.createElement('style');
    style.textContent = css;
    document.head.appendChild(style);
}

// ---------------------------------------------------------------------------
// VirtualWindow
//
// Wraps a WinBox instance with a well-defined three-state lifecycle:
//   undefined  -> not yet created (first GMCP data triggers creation)
//   WinBox obj -> open
//   false      -> user closed it; will not reopen automatically
//
// Usage:
//   const win = new VirtualWindow('My Title', () => {
//       const el = document.createElement('div');
//       // ... build DOM ...
//       document.body.appendChild(el);
//       return { mount: el, width: 300, height: 200, x: 'right', y: 0 };
//   });
//   win.open();          // creates on first call, no-ops if closed by user
//   win.isOpen()         // true if the WinBox exists and is not closed
//   win.get()            // returns the WinBox instance, or null
// ---------------------------------------------------------------------------
class VirtualWindow {
    constructor(id, factory) {
        this._id      = id;
        this._factory = factory;
        this._win     = undefined;  // undefined | WinBox | false
    }

    open() {
        // Already open
        if (this._win && this._win !== false) {
            return;
        }
        // User explicitly closed it — do not reopen
        if (this._win === false) {
            return;
        }
        // First open: build the WinBox via the factory
        const opts = this._factory();
        if (!opts) {
            return;
        }
        // Inject the onclose handler so we record the user's intent
        const userOnClose = opts.onclose;
        opts.onclose = (force) => {
            this._win = false;
            if (typeof userOnClose === 'function') {
                return userOnClose(force);
            }
            return false;
        };
        this._win = new WinBox(opts);
    }

    isOpen() {
        return !!this._win && this._win !== false;
    }

    get() {
        return (this._win && this._win !== false) ? this._win : null;
    }
}

// ---------------------------------------------------------------------------
// VirtualWindows registry
//
// Window modules call VirtualWindows.register(descriptor) where descriptor is:
//   {
//       gmcpHandlers: ['Char.Vitals', 'Char'],   // GMCP namespaces this handles
//       onGMCP(namespace, data) { ... }           // called when any listed namespace updates
//   }
//
// Multiple modules may register for the same namespace — all handlers are called.
// handleGMCP(namespace, body) walks from the most-specific to least-specific
// namespace segment and calls every handler registered at the first level that
// has any handlers.
// ---------------------------------------------------------------------------
const VirtualWindows = (() => {
    // Map<gmcpNamespace, Array<handler function>>
    const _handlers = {};

    function register(descriptor) {
        if (!descriptor || !Array.isArray(descriptor.gmcpHandlers)) {
            console.error('VirtualWindows.register: descriptor must have gmcpHandlers array');
            return;
        }
        if (typeof descriptor.onGMCP !== 'function') {
            console.error('VirtualWindows.register: descriptor must have onGMCP function');
            return;
        }
        descriptor.gmcpHandlers.forEach(ns => {
            if (!_handlers[ns]) {
                _handlers[ns] = [];
            }
            _handlers[ns].push(descriptor.onGMCP.bind(descriptor));
        });
    }

    function handleGMCP(namespace, body) {
        // Walk from most-specific to least-specific namespace segment.
        // Call all handlers registered at the first matching level.
        const parts = namespace.split('.');
        for (let i = parts.length; i >= 1; i--) {
            const path = parts.slice(0, i).join('.');
            if (_handlers[path] && _handlers[path].length > 0) {
                _handlers[path].forEach(fn => fn(namespace, body));
                return;
            }
        }
        console.log('GMCP (unhandled):', namespace, body);
    }

    return { register, handleGMCP };
})();

// ---------------------------------------------------------------------------
// Client namespace
//
// Shared state and services that window modules may read or call.
// Nothing here is truly private — window modules are trusted collaborators.
// ---------------------------------------------------------------------------
const Client = (() => {

    // -----------------------------------------------------------------------
    // Audio
    // -----------------------------------------------------------------------
    let baseMp3Url = '';
    const MusicPlayer = new MP3Player(false);
    const SoundPlayer = new MP3Player(true);

    // -----------------------------------------------------------------------
    // Terminal
    // -----------------------------------------------------------------------
    const term = new window.Terminal({
        cols: 80,
        rows: 60,
        cursorBlink: true,
        fontSize: 20,
    });
    const fitAddon = new window.FitAddon.FitAddon();
    term.loadAddon(fitAddon);

    function resizeTerminal() {
        fitAddon.fit();
    }

    // -----------------------------------------------------------------------
    // Networking stats
    // -----------------------------------------------------------------------
    let payloadsReceived = 0;
    let totalBytesReceived = 0;
    let payloadsSent      = 0;
    let totalBytesSent    = 0;

    // -----------------------------------------------------------------------
    // Command history
    // -----------------------------------------------------------------------
    let commandHistory          = [];
    let historyPosition         = 0;
    const commandHistoryMaxLength = 30;

    // -----------------------------------------------------------------------
    // GMCP state store
    //
    // GMCPStructs holds the most-recently-received value for every namespace.
    // Window modules read from it inside their onGMCP callbacks.
    // -----------------------------------------------------------------------
    const GMCPStructs = {};

    function _applyGMCPPayload(namespace, body) {
        const parts        = namespace.split('.');
        const lastProperty = parts.pop();
        let cursor         = GMCPStructs;
        for (const seg of parts) {
            if (!cursor[seg]) {
                cursor[seg] = {};
            }
            cursor = cursor[seg];
        }
        cursor[lastProperty] = body;
    }

    // -----------------------------------------------------------------------
    // WebSocket
    // -----------------------------------------------------------------------
    let socket               = null;
    let pendingReconnectToken = null;
    let debugOutput           = false;  // set Client.debug = true from the console to enable

    function debugLog(msg) {
        if (debugOutput) {
            console.log(msg);
        }
    }

    function sendData(dataToSend) {
        if (!socket || socket.readyState !== WebSocket.OPEN) {
            return false;
        }
        payloadsSent++;
        totalBytesSent += dataToSend.length;
        socket.send(dataToSend);
        return true;
    }

    function _parseMSPProps(parts, startIndex) {
        const props = {};
        for (let i = startIndex; i < parts.length; i++) {
            const eq = parts[i].indexOf('=');
            if (eq !== -1) {
                props[parts[i].slice(0, eq)] = parts[i].slice(eq + 1);
            }
        }
        return props;
    }

    function _handleMusicCommand(raw) {
        const inner  = raw.slice(8, raw.length - 1);
        const parts  = inner.split(' ');
        const fileName = parts[0];
        const obj    = _parseMSPProps(parts, 1);

        if (fileName === 'Off') {
            if (obj.U) {
                baseMp3Url = obj.U;
                if (baseMp3Url[baseMp3Url.length - 1] !== '/') {
                    baseMp3Url += '/';
                }
            } else {
                MusicPlayer.stop();
            }
            return;
        }

        let loopMusic  = true;
        let soundLevel = 1.0;
        if (obj.L && obj.L !== '-1') { loopMusic  = false; }
        if (obj.V)                    { soundLevel = Number(obj.V) / 100; }

        if (!MusicPlayer.isPlaying(baseMp3Url + fileName)) {
            MusicPlayer.play(baseMp3Url + fileName, loopMusic, soundLevel * (sliderValues['music'] / 100));
        }
    }

    function _handleSoundCommand(raw) {
        const inner    = raw.slice(8, raw.length - 1);
        const parts    = inner.split(' ');
        const fileName = parts[0];
        const obj      = _parseMSPProps(parts, 1);

        if (fileName === 'Off') {
            if (obj.U) {
                baseMp3Url = obj.U;
                if (baseMp3Url[baseMp3Url.length - 1] !== '/') {
                    baseMp3Url += '/';
                }
            } else {
                SoundPlayer.stop();
            }
            return;
        }

        let soundLevel = 1.0;
        let loopSound  = true;
        if (obj.L && obj.L !== '-1') { loopSound  = false; }
        if (obj.V)                    { soundLevel = Number(obj.V) / 100; }

        const typeKey = ((obj.T || 'other').toLowerCase()) + ' sounds';
        SoundPlayer.play(baseMp3Url + fileName, false, soundLevel * (sliderValues[typeKey] / 100));
    }

    function _handleWebclientCommand(data) {
        if (data.startsWith('TEXTMASK:')) {
            debugLog(data);
            textInput.type = data.substring(9) === 'true' ? 'password' : 'text';
            return true;
        }
        if (data.startsWith('RELOGTKN:')) {
            pendingReconnectToken = data.substring(9);
            return true;
        }
        return false;
    }

    function _onMessage(event) {
        payloadsReceived++;
        totalBytesReceived += event.data.length;

        // Webclient protocol commands (TEXTMASK:, RELOGTKN:)
        if (_handleWebclientCommand(event.data)) {
            return;
        }

        // MSP / GMCP commands (all start with "!!")
        if (event.data.length > 2 && event.data.slice(0, 2) === '!!') {

            if (event.data.slice(0, 7) === '!!GMCP(') {
                const gmcpPayload = event.data.trim().slice(7, event.data.length - 1).trim();
                const lastChar    = gmcpPayload[gmcpPayload.length - 1];
                const jsonIndex   = (lastChar === '}') ? gmcpPayload.indexOf('{') : gmcpPayload.indexOf('[');
                if (jsonIndex === -1) {
                    return;
                }
                const gmcpNamespace = gmcpPayload.slice(0, jsonIndex).trim();
                const gmcpBody      = JSON.parse(gmcpPayload.slice(jsonIndex).trim());
                _applyGMCPPayload(gmcpNamespace, gmcpBody);
                VirtualWindows.handleGMCP(gmcpNamespace, gmcpBody);
                return;
            }

            if (event.data.slice(0, 8) === '!!MUSIC(') {
                _handleMusicCommand(event.data);
                return;
            }

            if (event.data.slice(0, 8) === '!!SOUND(') {
                _handleSoundCommand(event.data);
                return;
            }
        }

        term.write(event.data);
    }

    function attachSocketHandlers(openMessage, clearOnOpen) {
        socket.onopen = function() {
            if (clearOnOpen) { term.clear(); }
            term.writeln(openMessage);
            connectButton.style.display = 'none';
            connectButton.disabled = true;
            textInput.focus();
        };

        socket.onmessage = _onMessage;

        socket.onerror = function(error) {
            term.writeln('Error: ' + (error.message || 'unknown'));
        };

        socket.onclose = function(event) {
            if (event.wasClean) {
                term.writeln('Connection closed cleanly, code=' + event.code + ', reason=' + event.reason);
            } else {
                term.writeln('Connection died');
            }
            connectButton.style.display = 'block';
            connectButton.disabled = false;

            if (textInput.type === 'password') {
                textInput.value = '';
                textInput.type  = 'text';
            }

            if (pendingReconnectToken) {
                const token = pendingReconnectToken;
                pendingReconnectToken = null;
                setTimeout(() => reconnectWithToken(token), 500);
            }
        };
    }

    function reconnectWithToken(token) {
        debugLog('Reconnecting with copyover token');
        const wsUrl = (location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/ws';
        socket = new WebSocket(wsUrl);
        attachSocketHandlers('Reconnected after server reboot.', false);
        const origOnOpen = socket.onopen;
        socket.onopen = function() {
            origOnOpen();
            socket.send(token);
        };
    }

    // -----------------------------------------------------------------------
    // Volume sliders
    // -----------------------------------------------------------------------
    const defaultSliders = {
        'music':               75,
        'combat sounds':       75,
        'movement sounds':     75,
        'environment sounds':  75,
        'other sounds':        75,
    };

    let sliderValues        = { ...defaultSliders };
    let unmutedSliderValues = null;

    function getSpeakerIcon(value) {
        value = Number(value);
        if (value === 0)       { return '🔇'; }
        if (value < 33)        { return '🔈'; }
        if (value < 66)        { return '🔉'; }
        return '🔊';
    }

    function buildSliders() {
        const container = document.getElementById('sliders-container');
        container.innerHTML = '';

        Object.keys(sliderValues).forEach(key => {
            const wrapper = document.createElement('div');
            wrapper.className = 'slider-container';

            const label = document.createElement('label');
            label.textContent = key.toLowerCase().split(' ').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');

            const slider = document.createElement('input');
            slider.type  = 'range';
            slider.min   = 0;
            slider.max   = 100;
            slider.value = sliderValues[key];

            const iconSpan = document.createElement('span');
            iconSpan.className   = 'slider-icon';
            iconSpan.textContent = getSpeakerIcon(sliderValues[key]);

            slider.addEventListener('input', e => {
                const val = Number(e.target.value);
                sliderValues[key] = val;
                iconSpan.textContent = getSpeakerIcon(val);
                localStorage.setItem('sliderValues', JSON.stringify(sliderValues));
                MusicPlayer.setGlobalVolume(sliderValues['music'] / 100);

                const muteCheckbox = document.getElementById('mute-checkbox');
                if (muteCheckbox.checked && val > 0) {
                    muteCheckbox.checked = false;
                    localStorage.setItem('muteAllSound', JSON.stringify(false));
                    document.getElementById('mute-icon').textContent = '🔊';
                }
            });

            wrapper.appendChild(label);
            wrapper.appendChild(slider);
            wrapper.appendChild(iconSpan);
            container.appendChild(wrapper);
        });
    }

    function toggleMuteAll() {
        const muteCheckbox = document.getElementById('mute-checkbox');
        const muteIcon     = document.getElementById('mute-icon');
        const isChecked    = muteCheckbox.checked;

        if (isChecked) {
            unmutedSliderValues = { ...sliderValues };
            localStorage.setItem('unmutedSliderValues', JSON.stringify(unmutedSliderValues));
            Object.keys(sliderValues).forEach(k => { sliderValues[k] = 0; });
            localStorage.setItem('sliderValues', JSON.stringify(sliderValues));
            buildSliders();
            muteIcon.textContent = '🔇';
            MusicPlayer.setGlobalVolume(0);
            localStorage.setItem('muteAllSound', JSON.stringify(true));
        } else {
            const savedUnmuted = localStorage.getItem('unmutedSliderValues');
            if (savedUnmuted) {
                let loaded = JSON.parse(savedUnmuted) || {};
                loaded = { ...defaultSliders, ...loaded };
                unmutedSliderValues = { ...loaded };
                sliderValues = { ...unmutedSliderValues };
                localStorage.setItem('sliderValues', JSON.stringify(sliderValues));
            }
            buildSliders();
            muteIcon.textContent = '🔊';
            MusicPlayer.setGlobalVolume(sliderValues['music'] / 100);
            localStorage.setItem('muteAllSound', JSON.stringify(false));
        }
    }

    function toggleMenu() {
        const menu = document.getElementById('floating-menu');
        menu.style.display = (menu.style.display === 'none' || menu.style.display === '') ? 'block' : 'none';
    }

    // -----------------------------------------------------------------------
    // Net stats
    // -----------------------------------------------------------------------
    function printNetStats() {
        term.writeln('');
        term.writeln(' Request Ct: ' + String(payloadsSent));
        term.writeln(' Bytes Sent: ' + String(Math.round(totalBytesSent    / 1024 * 100) / 100) + 'kb');
        term.writeln('Response Ct: ' + String(payloadsReceived));
        term.writeln(' Bytes Rcvd: ' + String(Math.round(totalBytesReceived / 1024 * 100) / 100) + 'kb');
        term.writeln('');
    }

    // -----------------------------------------------------------------------
    // Keyboard shortcuts
    //
    // Window modules may call Client.registerShortcut(code, command) to add
    // their own bindings, e.g. Client.registerShortcut('KeyM', 'map').
    // -----------------------------------------------------------------------
    const codeShortcuts = {
        Numpad1: 'southwest', Numpad2: 'south',  Numpad3: 'southeast',
        Numpad4: 'west',      Numpad5: 'default', Numpad6: 'east',
        Numpad7: 'northwest', Numpad8: 'north',   Numpad9: 'northeast',
        F1: '=1', F2: '=2', F3: '=3',  F4: '=4',  F5: '=5',
        F6: '=6', F7: '=7', F8: '=8',  F9: '=9',  F10: '=10',
        ArrowUp: 'north', ArrowDown: 'south', ArrowLeft: 'west', ArrowRight: 'east',
    };

    function registerShortcut(code, command) {
        codeShortcuts[code] = command;
    }

    // -----------------------------------------------------------------------
    // Terminal commands
    //
    // Window modules may call Client.registerCommand(name, description, fn)
    // to add their own !commands processed before sending to the server.
    // fn receives the full input string and returns true if it handled it.
    // -----------------------------------------------------------------------
    const specialCommands = {
        '!net': { description: 'Print out network traffic stats', fn: () => { printNetStats(); return true; } },
    };

    function registerCommand(name, description, fn) {
        specialCommands[name] = { description, fn };
    }

    // -----------------------------------------------------------------------
    // DOM references (resolved at init time)
    // -----------------------------------------------------------------------
    let connectButton, textOutput, textInput;

    // -----------------------------------------------------------------------
    // init()
    // -----------------------------------------------------------------------
    function init() {
        connectButton = document.getElementById('connect-button');
        textOutput    = document.getElementById('terminal');
        textInput     = document.getElementById('command-input');

        // Mount terminal
        term.open(textOutput);
        window.addEventListener('resize', resizeTerminal);
        resizeTerminal();

        // Keep focus on terminal on click (not drag)
        let isDragging = false;
        textOutput.addEventListener('mousedown', () => { isDragging = false; });
        textOutput.addEventListener('mousemove', () => { isDragging = true; });
        textOutput.addEventListener('mouseup', () => {
            const selected = window.getSelection().toString();
            if (!isDragging && !selected) { textInput.focus(); }
            isDragging = false;
        });

        // Connect button
        connectButton.addEventListener('click', () => {
            if (socket && socket.readyState === WebSocket.OPEN) {
                socket.close();
                return;
            }
            const wsUrl = (location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/ws';
            debugLog('Connecting to: ' + wsUrl);
            socket = new WebSocket(wsUrl);
            attachSocketHandlers('Connected to the server!', true);
        });

        // Input keydown
        textInput.addEventListener('keydown', function(event) {
            // F-key macros
            if (event.key.substring(0, 1) === 'F' && event.key.length === 2) {
                sendData('=' + event.key.substring(1));
                if (event.preventDefault) { event.preventDefault(); }
                return false;
            }

            // Command history
            if (event.key === 'ArrowUp' || event.key === 'ArrowDown') {
                historyPosition += (event.key === 'ArrowUp') ? 1 : -1;
                if (historyPosition < 1) { historyPosition = 1; }
                if (historyPosition > commandHistory.length) { historyPosition = commandHistory.length; }
                event.target.value = commandHistory[commandHistory.length - historyPosition];
                return;
            }

            // Numpad / arrow shortcuts when input is empty
            if (textInput.value.length === 0 && codeShortcuts[event.code]) {
                sendData(codeShortcuts[event.code]);
                if (event.preventDefault) { event.preventDefault(); }
                return false;
            }

            // Enter
            if (event.key === 'Enter') {
                if (event.target.value !== '' && textInput.type !== 'password') {
                    commandHistory.push(event.target.value);
                    historyPosition = 0;
                    if (commandHistory.length > commandHistoryMaxLength) {
                        commandHistory = commandHistory.slice(commandHistory.length - commandHistoryMaxLength);
                    }
                }

                const cmd = specialCommands[event.target.value];
                if (cmd) {
                    if (cmd.fn(event.target.value)) {
                        event.target.value = '';
                        return;
                    }
                }

                if (sendData(event.target.value)) {
                    event.target.value = '';
                } else {
                    term.writeln('Not connected to the server. Did you click the Connect button?');
                }
            }
        });

        // Volume sliders: load from localStorage
        const savedValues = localStorage.getItem('sliderValues');
        if (savedValues) {
            try {
                sliderValues = { ...defaultSliders, ...JSON.parse(savedValues) };
            } catch (e) {
                console.warn('Could not parse saved sliderValues, using defaults.');
            }
        } else {
            localStorage.setItem('sliderValues', JSON.stringify(sliderValues));
        }

        const savedMute = localStorage.getItem('muteAllSound');
        if (savedMute) {
            try {
                document.getElementById('mute-checkbox').checked = JSON.parse(savedMute);
            } catch (e) {
                console.warn('Could not parse muteAllSound, ignoring.');
            }
        }

        buildSliders();

        const muteCheckbox = document.getElementById('mute-checkbox');
        const muteIcon     = document.getElementById('mute-icon');

        if (muteCheckbox.checked) {
            const savedUnmuted = localStorage.getItem('unmutedSliderValues');
            if (savedUnmuted) {
                try {
                    unmutedSliderValues = { ...defaultSliders, ...JSON.parse(savedUnmuted) };
                } catch (e) {
                    console.warn('Could not parse unmutedSliderValues.');
                }
            }
            Object.keys(sliderValues).forEach(k => { sliderValues[k] = 0; });
            localStorage.setItem('sliderValues', JSON.stringify(sliderValues));
            buildSliders();
            muteIcon.textContent = '🔇';
            MusicPlayer.setGlobalVolume(0);
        } else {
            MusicPlayer.setGlobalVolume(sliderValues['music'] / 100);
            muteIcon.textContent = '🔊';
        }

        // Log available commands to console
        console.log('%cterminal commands:', 'font-weight:bold;');
        let longest = 0;
        for (const k in specialCommands) { if (k.length > longest) { longest = k.length; } }
        for (const k in specialCommands) { console.log('  ' + k.padEnd(longest) + ' - ' + specialCommands[k].description); }
        console.log('%cconsole commands:', 'font-weight:bold;');
        console.log('  Client.debug = true   - enable debug logging');
        console.log('  Client.registerCommand(name, description, fn)  - add a terminal command');
        console.log('  Client.registerShortcut(code, command)          - add a keyboard shortcut');
    }

    // -----------------------------------------------------------------------
    // Public surface
    // -----------------------------------------------------------------------
    return {
        // Services
        get term()         { return term; },
        get MusicPlayer()  { return MusicPlayer; },
        get SoundPlayer()  { return SoundPlayer; },

        // Shared state (read by window modules)
        get GMCPStructs()  { return GMCPStructs; },
        // sliderValues is a `let` that gets reassigned on mute/unmute, so the
        // getter captures the variable binding, not a snapshot of the object.
        get sliderValues() { return sliderValues; },

        // Debug toggle: set Client.debug = true from the browser console
        get debug()        { return debugOutput; },
        set debug(v)       { debugOutput = !!v; },

        // Extension points for window modules
        registerCommand,
        registerShortcut,

        // Functions called from HTML event handlers
        init,
        toggleMenu,
        toggleMuteAll,

        // Utility
        sendData,
        debugLog,
    };
})();
