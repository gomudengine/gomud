# GoMud Web Client — Developer Context

This document describes the architecture of the web client and explains how to
create new virtual windows that respond to GMCP payloads.

---

## File Map

```
webclient-pure.html                      Page shell, layout, init bridge
static/js/webclient-core.js             Core infrastructure (loaded first)
static/js/windows/window-vitals.js      Vitals window module
static/js/windows/window-map.js         Map window module
static/js/windows/window-comm.js        Communications window module
static/css/windows.css                  Shared dock/panel styles
```

`webclient-pure.html` is a Go template. All static asset paths must be prefixed
with `{{ .CONFIG.FilePaths.WebCDNLocation }}`. To add a new window module,
add one `<script>` tag after the existing window scripts.

---

## Architecture Overview

```
webclient-core.js defines three globals:

  injectStyles(css)        Append a <style> block to <head>
  VirtualWindows           Registry: register(), handleGMCP(), openAll()
  Client                   Shared state and services
  VirtualWindow            Class: lifecycle for a single panel
  DockSlots                Singleton: { left, right } DockSlot instances
```

### Load and init sequence

1. `webclient-core.js` executes — globals are defined, `DockSlots` are created
   on `DOMContentLoaded`.
2. Window module scripts execute — each calls `VirtualWindows.register(...)`,
   which records the window and its GMCP handlers.
3. The page `onload` fires, calling `Client.init()`.
4. `Client.init()` sets up the terminal, WebSocket, and volume sliders, then
   calls `VirtualWindows.openAll()` — every registered window opens immediately.
5. When the WebSocket receives a `!!GMCP(...)` message, `VirtualWindows.handleGMCP()`
   dispatches it to the matching handler(s). Handlers whose window has been
   closed by the user are silently skipped.

---

## VirtualWindow Lifecycle

A `VirtualWindow` has four states:

| `_win` value | Meaning |
|---|---|
| `undefined` | Never opened (before first `open()` call) |
| `'docked'` | Content is in a dock slot panel; no WinBox exists |
| WinBox instance | Floating, visible |
| `false` | Closed by the user; will not reopen for this session |

`open()` is idempotent — safe to call on every GMCP update. It is a no-op
if the window is already open or has been closed.

### Constructor options

```js
new VirtualWindow(id, {
    factory(),        // required — called once on first open; returns WinBox opts
    dock,             // optional — 'left' | 'right'; enables docking
    defaultDocked,    // optional — boolean; start docked instead of floating
    dockedHeight,     // optional — number (px); preferred height when docked
})
```

The `factory()` function must:
- Create the content DOM element
- Append it to `document.body` (required before WinBox can mount it)
- Return a WinBox options object with at minimum `{ title, mount: el }`

### Methods

```js
win.open()      // Open (first call) or no-op (already open or closed)
win.isOpen()    // true when floating or docked
win.get()       // Returns WinBox instance when floating, null otherwise
win.dock()      // Move from floating to docked
win.undock()    // Move from docked to floating
```

---

## VirtualWindows Registry

```js
VirtualWindows.register({
    window:       win,            // VirtualWindow instance (enables openAll + GMCP skip)
    gmcpHandlers: ['Foo.Bar'],    // GMCP namespaces this module handles
    onGMCP(namespace, body) {     // called on matching GMCP payload
        ...
    },
});
```

**Dispatch rules:**
- Namespaces are matched from most-specific to least-specific. A payload of
  `Char.Vitals` matches `Char.Vitals` before `Char`.
- Multiple modules may register for the same namespace — all are called.
- If a handler's associated `window` is closed (`_win === false`), it is
  skipped entirely. No handler is called for a closed window.

**`Client.GMCPStructs`** is updated before `onGMCP` is called, so handlers
can always read the latest value from it directly.

---

## Client API (available to window modules)

```js
Client.GMCPStructs          // Object tree of all received GMCP data
Client.sliderValues         // Current volume levels by category key
Client.MusicPlayer          // MP3Player instance (background music)
Client.SoundPlayer          // MP3Player instance (sound effects)
Client.term                 // xterm.js Terminal instance
Client.sendData(str)        // Send a string over the WebSocket; returns bool
Client.debugLog(msg)        // Log only when Client.debug === true
Client.debug                // get/set — enable verbose logging from console
Client.registerCommand(name, description, fn)   // Add a !terminal command
Client.registerShortcut(code, command)          // Add a keyboard shortcut
```

---

## Creating a New Window Module

### 1. Create the file

`static/js/windows/window-example.js`

Every module is wrapped in an IIFE to keep all state private.

```js
'use strict';

(function() {

    // -- Styles (injected once at script load time) --------------------------
    injectStyles(`
        #example-content {
            color: #fff;
            padding: 8px;
            font-family: monospace;
        }
    `);

    // -- DOM factory --------------------------------------------------------
    // Called once, on first open. Creates the content element and appends it
    // to document.body so WinBox can mount it.
    function createDOM() {
        const el = document.createElement('div');
        el.id = 'example-content';
        el.textContent = 'Waiting for data...';
        document.body.appendChild(el);
        return el;
    }

    // -- VirtualWindow instance ---------------------------------------------
    const win = new VirtualWindow('Example', {
        dock:          'right',   // enable docking to the right column
        defaultDocked: true,      // start docked on page load
        dockedHeight:  200,       // preferred height (px) when docked
        factory() {
            const el = createDOM();
            return {
                title:      'Example',
                mount:      el,
                background: '#1c6b60',
                border:     1,
                x:          'right',
                y:          0,
                width:      363,
                height:     220,
                header:     20,
                bottom:     60,
            };
        },
    });

    // -- Update logic -------------------------------------------------------
    function update() {
        // GMCPStructs is already updated before onGMCP fires.
        const data = Client.GMCPStructs.Some && Client.GMCPStructs.Some.Namespace;
        if (!data) { return; }

        // win.open() is a no-op if already open or closed by the user.
        win.open();
        if (!win.isOpen()) { return; }

        document.getElementById('example-content').textContent = data.someField;
    }

    // -- Registration -------------------------------------------------------
    VirtualWindows.register({
        window:       win,
        gmcpHandlers: ['Some.Namespace', 'Some'],
        onGMCP(namespace, body) {
            update();
        },
    });

})();
```

### 2. Register the script in `webclient-pure.html`

Add one line after the existing window scripts:

```html
<script src="{{ .CONFIG.FilePaths.WebCDNLocation }}/static/js/windows/window-example.js"></script>
```

That is all that is required. The window will open on page load, respond to
its GMCP namespaces, and stop receiving payloads if the user closes it.

---

## GMCP Namespace Conventions

Namespaces follow a dot-separated hierarchy. Register the most specific
namespace you need. If you need to handle both a parent and a child (e.g.
`Char` and `Char.Vitals`), register both in `gmcpHandlers` — the dispatch
walks from most-specific to least-specific and stops at the first level that
has any registered handlers, calling all of them.

```js
// Receives Char.Vitals payloads directly, and also bare Char payloads
// (which update GMCPStructs.Char and may include Vitals data).
gmcpHandlers: ['Char.Vitals', 'Char'],
```

Inside `onGMCP`, always read from `Client.GMCPStructs` rather than the `body`
argument. The store is the source of truth and is updated before dispatch.

---

## Dock Slot Behaviour

- Two slots exist: `#dock-left` and `#dock-right`, both flex children of
  `#main-container`.
- A slot is zero-width when empty and expands to its stored width when it
  contains panels. The slot-width drag handle appears between the slot and
  the terminal when panels are present.
- Panels within a slot can be resized vertically by dragging the handle
  between them.
- Each docked panel has a titlebar with a pop-out arrow (undocks to floating)
  and an X (closes the window for the session).
- When a floating window has `dock` configured, its WinBox header contains a
  dock button (downward arrow) that moves it back into the slot.

---

## Extension Points

### Adding a terminal command

```js
Client.registerCommand('!example', 'Print example info', (input) => {
    Client.term.writeln('Example command ran.');
    return true;  // return true to consume the input (clears the input field)
});
```

Commands are matched against the exact input string before it is sent to the
server. Return `true` to prevent the string from being sent.

### Adding a keyboard shortcut

```js
// Sends 'look' when the user presses Tab with an empty input field.
Client.registerShortcut('Tab', 'look');
```

The `code` value is a `KeyboardEvent.code` string (e.g. `'KeyM'`, `'Tab'`,
`'F11'`). Shortcuts only fire when the command input field is empty.
