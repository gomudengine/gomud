# Admin HTML Directory Context

## Overview

The `_datafiles/html/admin/` directory contains all front-end assets for the GoMud admin interface. The path is configured via `FilePaths.AdminHtml` in `config.yaml` and defaults to `_datafiles/html/admin`. Files are served by `internal/web`:

- HTML pages are rendered server-side via Go's `text/template` engine by the handlers in `internal/web/admin.go` and `internal/web/admin_config.go`.
- Non-HTML files (`.js`, images, etc.) are served as static assets via `serveAdminStaticFile` at `GET /admin/{file}` (single-segment) and `GET /admin/static/{path...}` (multi-segment), authenticated but not mud-locked. JavaScript files live in `_datafiles/html/admin/static/js/`.

No external CDN dependencies. No Bootstrap, jQuery, or HTMX.

---

## Page Skeleton

Every admin page uses the same three-part structure:

```html
{{template "header" .}}

<style>
    /* Page-specific CSS goes here — never in external files */
</style>

<!-- Page HTML content -->
<h1>Page Title</h1>
<p class="subtitle">One-line description.</p>

<!-- ... page body ... -->

<script>
    /* Page-specific JavaScript goes here — never in external files */
</script>

{{template "footer" .}}
```

The `_header.html` template provides: DOCTYPE, `<head>`, CSS reset/base styles, nav bar, all shared JS/CSS imports, and the opening `<main>` tag. The `_footer.html` template closes `</main></body></html>`.

All page-specific CSS and JS are **inline** within the page file. Shared utilities live in `static/js/`.

## File Naming

| Type | Pattern | Example |
|---|---|---|
| Editor/viewer page | `pagename.html` | `races.html` |
| API documentation | `pagename-api.html` | `races-api.html` |
| Shared templates | `_name.html` (underscore prefix) | `_header.html` |
| Shared JavaScript | `static/js/name.js` | `static/js/api.js` |
| Shared CSS | `static/css/name.css` | `static/css/highlight.css` |

Every editor page that calls APIs **must** have a corresponding `pagename-api.html` documenting those endpoints. Link to it from the page subtitle.

## Template Data

Pages receive Go template data from their handler. Common fields:

- `.CONFIG` — the `configs.Config` struct (server settings, file paths, etc.)
- `.STATS` — the `web.Stats` struct (connection counts, ports)
- `.NAV` — navigation entries (consumed by `_header.html`, not your page)

Use `{{.CONFIG.FilePaths.WebDomain}}` in API doc curl examples for the server hostname.

---

## Layout Patterns

### Two-Column Layout (Sidebar + Editor)

This is the standard pattern for CRUD pages (items, buffs, races, quests, colorpatterns). A left sidebar lists records; the right panel shows the editor form.

```html
<div class="page-layout">
    <div class="sidebar">
        <div class="sidebar-header">
            <input type="search" id="searchInput" placeholder="Search..." oninput="filterList()" />
        </div>
        <div id="list" class="thing-list">
            <div class="no-things">Loading...</div>
        </div>
        <button class="btn-new" onclick="newThing()">+ New Thing</button>
    </div>

    <div class="editor-panel">
        <div id="editorPlaceholder" class="editor-placeholder">
            Select an item to edit, or click "New Thing".
        </div>
        <div id="editorForm" style="display:none">
            <div id="statusBar" class="status-bar"></div>
            <div class="form-grid">
                <!-- form fields go here -->
            </div>
        </div>
    </div>
</div>
```

Required CSS for this layout:

```css
.page-layout {
    display: grid;
    grid-template-columns: 260px 1fr;   /* sidebar width: 260-300px */
    gap: 1.25rem;
    align-items: start;
}
@media (max-width: 800px) {
    .page-layout { grid-template-columns: 1fr; }
}
```

### Dashboard Layout (Cards Grid)

Used by `index.html` for summary/stats pages:

```html
<div class="grid">
    <div class="card">
        <h2>Card Title</h2>
        <div class="value">42</div>
        <div class="detail">Additional info</div>
    </div>
    <!-- more cards -->
</div>
```

```css
.grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1rem; }
.card { background: #fff; border: 1px solid #ddd; border-radius: 6px; padding: 1.25rem; }
```

### Full-Width Layout

Used by `config.html`, `keywords.html`, and `users.html` for table/search UIs. No sidebar — content fills the `<main>` container directly.

---

## CSS Conventions

### Core Rules

1. **All page styles are inline** in a `<style>` block at the top of the page, after `{{template "header" .}}`.
2. **No external CSS frameworks.** No Bootstrap, Tailwind, or similar.
3. **No shared page CSS.** Each page re-declares its own styles. This avoids cross-page coupling.
4. **Copy proven CSS patterns** from existing pages rather than inventing new ones.

### Color Palette

| Usage | Value |
|---|---|
| Primary dark (nav, buttons, focus) | `#1a1a2e` |
| Primary dark hover | `#2d2d4e` |
| Success green (background) | `#e6f4ea` |
| Success green (text) | `#1e6e34` |
| Error red (background) | `#fde8e8` |
| Error red (text) | `#8a0000` |
| Page background | `#f5f5f5` |
| Card/panel background | `#fff` |
| Border | `#ddd` |
| Subtle border | `#eee`, `#f0f0f0` |
| Label text | `#555` |
| Muted text | `#888`, `#999` |
| Button green (new/add) | `#1e6e34` / hover `#145728` |

### Typography

| Element | Size | Weight | Notes |
|---|---|---|---|
| `h1` | `1.5rem` | default (700) | Page title |
| `.subtitle` | `0.9rem` | normal | Below the h1, color `#555` |
| Labels | `0.78rem` | 600 | Uppercase, letter-spacing `0.03em` |
| Body/inputs | `0.875rem` | normal | |
| Small badges | `0.68rem`–`0.75rem` | normal | Monospace for IDs |

### Panels and Cards

```css
/* Standard white panel (sidebar, editor, cards) */
background: #fff;
border: 1px solid #ddd;
border-radius: 6px;
```

### Sidebar Rows

Each page names its rows with a page-specific prefix:

```css
.race-row { /* or .item-row, .buff-row, .quest-row, .pattern-row */
    padding: 0.5rem 0.9rem;
    cursor: pointer;
    border-bottom: 1px solid #f0f0f0;
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
    font-size: 0.875rem;
}
.race-row:hover { background: #f5f7ff; }
.race-row.active { background: #1a1a2e; color: #fff; }
.race-id { /* ID badge */
    font-size: 0.75rem;
    color: #999;
    font-family: monospace;
    flex-shrink: 0;
}
.race-row.active .race-id { color: #8899cc; }
```

### Form Grid

Two-column grid inside the editor panel. Fields span one column by default; use `.span2` for full-width.

```css
.form-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.85rem 1.25rem;
}
.form-grid .span2 { grid-column: 1 / -1; }
```

### Form Fields

```html
<div class="field">
    <label for="fName">Name</label>
    <input type="text" id="fName" placeholder="..." />
</div>

<div class="field span2">
    <label for="fDesc">Description</label>
    <textarea id="fDesc" rows="3" placeholder="..."></textarea>
    <div class="field-hint">Help text shown below the input.</div>
</div>
```

```css
.field label {
    display: block; font-size: 0.78rem; font-weight: 600;
    color: #555; margin-bottom: 0.25rem;
    text-transform: uppercase; letter-spacing: 0.03em;
}
.field input, .field select, .field textarea {
    width: 100%; padding: 0.4rem 0.6rem;
    border: 1px solid #ccc; border-radius: 4px;
    font-size: 0.875rem; font-family: inherit; background: #fff;
}
.field input:focus, .field select:focus, .field textarea:focus {
    outline: 2px solid #1a1a2e; outline-offset: 1px; border-color: transparent;
}
.field-hint { font-size: 0.75rem; color: #888; margin-top: 0.2rem; }
```

### Buttons

```css
/* Primary action (Save) */
.btn-save {
    padding: 0.45rem 1.2rem; background: #1a1a2e; color: #fff;
    border: none; border-radius: 4px; font-size: 0.875rem;
    font-weight: 600; cursor: pointer;
}
.btn-save:hover { background: #2d2d4e; }

/* Destructive action (Delete) */
.btn-delete {
    padding: 0.45rem 1rem; background: #fff; color: #8a0000;
    border: 1px solid #e0a0a0; border-radius: 4px; font-size: 0.875rem;
    font-weight: 600; cursor: pointer;
}
.btn-delete:hover { background: #fde8e8; }

/* New/Create in sidebar */
.btn-new {
    margin: 0.6rem 0.9rem; padding: 0.4rem 0.9rem;
    border: none; border-radius: 4px; background: #1e6e34; color: #fff;
    font-size: 0.85rem; font-weight: 600; cursor: pointer;
    display: block; text-align: center;
}
.btn-new:hover { background: #145728; }

/* Small add/remove */
.btn-add-sm {
    font-size: 0.8rem; padding: 0.3rem 0.7rem;
    border: 1px dashed #aaa; background: none;
    border-radius: 4px; cursor: pointer; color: #555;
}
.btn-icon {
    background: none; border: none; cursor: pointer;
    font-size: 1rem; color: #888; padding: 0 0.2rem;
}
.btn-icon:hover { color: #c00; }
```

### Section Dividers

Use inside `.form-grid` to group related fields:

```html
<div class="section-title">Base Stats</div>
```

```css
.section-title {
    font-size: 0.78rem; font-weight: 700; text-transform: uppercase;
    letter-spacing: 0.06em; color: #666;
    padding: 0.75rem 0 0.4rem; border-bottom: 1px solid #eee;
    margin-bottom: 0.75rem; grid-column: 1 / -1;
}
```

### Status Bar

```html
<div id="statusBar" class="status-bar"></div>
```

```css
.status-bar {
    display: none; padding: 0.55rem 0.9rem; border-radius: 4px;
    font-size: 0.85rem; margin-bottom: 1rem;
}
.status-bar.success { background: #e6f4ea; color: #1e6e34; display: block; }
.status-bar.error   { background: #fde8e8; color: #8a0000; display: block; }
```

### Action Row

Buttons at the bottom of the editor form:

```html
<div class="btn-row">
    <button class="btn-save" onclick="saveThing()">Save</button>
    <button class="btn-delete" id="btnDelete" onclick="deleteThing()" style="display:none">Delete</button>
</div>
```

```css
.btn-row {
    display: flex; gap: 0.6rem; margin-top: 1.25rem; grid-column: 1 / -1;
}
```

---

## JavaScript Conventions

### General Rules

1. **All page JS is inline** in a `<script>` block at the bottom of the page, before `{{template "footer" .}}`.
2. **No external frameworks.** Vanilla JS only.
3. Functions called from `onclick`/`oninput` HTML attributes must be `window`-scoped (either use `window.fn = function()` or declare them as plain `function` statements at the top level of the `<script>` block).
4. **No `async`/`await` at the top level.** Wrap initialization in an `async function init()` and call it at the end of the script block.

### Event Handler Conventions

- `oninput="filterList()"` — real-time search filtering
- `onclick="saveThing()"` — button actions
- `onchange="handleChange()"` — dropdowns
- `onkeydown="if(event.key==='Enter') save()"` — enter-to-submit
- `onclick="event.stopPropagation()"` — prevent event bubbling when needed

### State Management

- Store data in top-level variables (`let allData = {}`, `let activeId = null`, `let isNew = false`).
- Each page is self-contained. No cross-page state or communication.
- UI state is tracked via DOM class toggles (`.active`, `.open`) and `style.display`.
- Never mutate the data store optimistically before the API confirms success.

### Standard Page Lifecycle

```js
let allData = {};
let activeId = null;
let isNew = false;

// 1. Load data on page init
async function loadData() {
    const res = await AdminAPI.get('/admin/api/v1/things');
    if (!res.ok) { console.error('Load failed:', res.error); return; }
    allData = (res.data && res.data.data) || {};
    renderList();
}

// 2. Render sidebar list
function renderList() {
    const list = document.getElementById('thingList');
    const ids = Object.keys(allData).sort((a, b) => parseInt(a) - parseInt(b));
    if (ids.length === 0) {
        list.innerHTML = '<div class="no-things">No items found.</div>';
        return;
    }
    list.innerHTML = ids.map(id => {
        const t = allData[id];
        return `<div class="thing-row" data-id="${id}" data-name="${escAttr(t.Name || '')}"
            onclick="selectThing('${id}')">
            <span class="thing-id">#${id}</span>
            <span>${escHtml(t.Name || '(unnamed)')}</span>
        </div>`;
    }).join('');
    if (activeId !== null) highlightRow(activeId);
}

// 3. Filter sidebar
function filterList() {
    const q = document.getElementById('searchInput').value.toLowerCase();
    document.querySelectorAll('.thing-row').forEach(row => {
        const name = (row.dataset.name || '').toLowerCase();
        row.style.display = name.includes(q) ? '' : 'none';
    });
}

// 4. Select / New
function selectThing(id) {
    activeId = id;
    isNew = false;
    highlightRow(id);
    populateForm(allData[id]);
}

function newThing() {
    activeId = null;
    isNew = true;
    document.querySelectorAll('.thing-row').forEach(r => r.classList.remove('active'));
    populateForm(null);
}

// 5. Populate form
function populateForm(thing) {
    document.getElementById('editorPlaceholder').style.display = 'none';
    document.getElementById('editorForm').style.display = '';
    document.getElementById('statusBar').className = 'status-bar';
    document.getElementById('btnDelete').style.display = isNew ? 'none' : '';
    // ... set form field values ...
}

// 6. Save
async function saveThing() {
    const payload = buildPayload();
    if (!payload.Name) { showStatus('error', 'Name is required.'); return; }

    let res;
    if (isNew) {
        res = await AdminAPI.post('/admin/api/v1/things', payload);
    } else {
        res = await AdminAPI.patch(`/admin/api/v1/things/${activeId}`, payload);
    }

    if (!res.ok) { showStatus('error', res.error || 'Save failed.'); return; }
    // Update local data, re-render list, show success
    showStatus('success', 'Saved.');
}

// 7. Delete
async function deleteThing() {
    if (!confirm('Delete this? This cannot be undone.')) return;
    const res = await AdminAPI.delete(`/admin/api/v1/things/${activeId}`);
    if (!res.ok) { showStatus('error', res.error || 'Delete failed.'); return; }
    delete allData[activeId];
    activeId = null;
    renderList();
    document.getElementById('editorForm').style.display = 'none';
    document.getElementById('editorPlaceholder').style.display = '';
}

// Utilities
function highlightRow(id) {
    document.querySelectorAll('.thing-row').forEach(r => r.classList.remove('active'));
    const row = document.querySelector(`.thing-row[data-id="${id}"]`);
    if (row) row.classList.add('active');
}

function showStatus(type, msg) {
    const bar = document.getElementById('statusBar');
    bar.className = 'status-bar ' + type;
    bar.textContent = msg;
    if (type === 'success') setTimeout(() => { bar.className = 'status-bar'; }, 3000);
}

function escHtml(s) { return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;'); }
function escAttr(s) { return String(s).replace(/"/g,'&quot;'); }

// Kick off
loadData();
```

---

## Shared Libraries

These are loaded by `_header.html` and available on every page:

| Global | File | Purpose |
|---|---|---|
| `AdminAPI` | `static/js/api.js` | HTTP client with caching |
| `AnsiColors` | `static/js/ansi-colors.js` | ANSI-256 to hex conversion and swatch rendering |
| `SelectDialog` | `static/js/select-dialog.js` | Reusable search/select modal |
| `ScriptEditor` | `static/js/script-editor.js` | Syntax-highlighted textarea for JS scripts |
| `hljs` | `static/js/highlight.js` | Code syntax highlighting engine |
| `AnsiTags` | `static/js/ansitags.js` | ANSI tag parsing and rendering |

### AdminAPI (`static/js/api.js`)

Thin, promise-based HTTP client for admin pages.

```js
// Single requests — return Promise<APIResult>
AdminAPI.get(path)
AdminAPI.post(path, body)
AdminAPI.put(path, body)
AdminAPI.patch(path, body)
AdminAPI.delete(path, body?)

// Parallel dispatch — resolves when all settle
AdminAPI.all([promise, ...])

// Fluent queue builder — dispatches all in parallel on .run()
AdminAPI.queue()
  .get(path)
  .patch(path, body)
  .run()  // returns Promise<APIResult[]>
```

#### `APIResult` shape

```js
{
  ok: boolean,     // true when HTTP status is 2xx
  status: number,  // HTTP status code
  data: any,       // parsed JSON body (null on parse failure)
  error: string,   // error message, or empty string on success
}
```

**Never use try/catch** — errors are returned as `ok: false`, not thrown.

#### Usage examples

```js
// GET — cached for 30 seconds
const res = await AdminAPI.get('/admin/api/v1/things');

// POST/PATCH/PUT/DELETE — invalidates cache for the resource root
const res = await AdminAPI.post('/admin/api/v1/things', payload);
const res = await AdminAPI.patch('/admin/api/v1/things/42', payload);
const res = await AdminAPI.delete('/admin/api/v1/things/42');

// Parallel requests
const [things, buffs] = await AdminAPI.all([
    AdminAPI.get('/admin/api/v1/things'),
    AdminAPI.get('/admin/api/v1/buffs'),
]);

// Fluent queue
const results = await AdminAPI.queue()
    .get('/admin/api/v1/things')
    .patch('/admin/api/v1/things/42', payload)
    .run();
```

#### Implementation details

- Uses `fetch` with `credentials: 'same-origin'` and `Content-Type: application/json`.
- Never rejects — network errors are caught and returned as `APIResult` with `ok: false`.
- Response body is parsed as JSON; falls back to raw text on parse failure.
- GET results are cached for 30 seconds. Mutations (POST/PATCH/PUT/DELETE) invalidate the cache for the resource root path.
- `AdminAPI.all` delegates to `Promise.all` (already non-rejecting because each request handles its own errors).
- `AdminAPI.queue().run()` splices the internal pending array so the same queue instance can be reused.

### AnsiColors (`static/js/ansi-colors.js`)

Converts ANSI 256-color codes to CSS hex values and renders color previews.

```js
AnsiColors.toHex(n)                    // ANSI code (0–255) → CSS hex string e.g. '#ff0000'
AnsiColors.swatchHtml(colors, max?)    // array of ANSI codes → HTML string of colored <span>s
AnsiColors.previewTextHtml(text, colors)  // colorize a text string using bounce pattern → HTML
```

- `toHex(n)` handles all 256 codes: 0–15 (standard), 16–231 (6×6×6 cube), 232–255 (grayscale).
- `swatchHtml(colors, max?)` returns `<span style="background:#hex"></span>` for each code, optionally limited to `max` entries.
- `previewTextHtml(text, colors)` applies the default bounce pattern (per-character, reversing at ends) and returns HTML with inline `color:` styles.

### SelectDialog (`static/js/select-dialog.js`)

Reusable search-and-select modal dialog.

```js
SelectDialog.open({
  title: 'Dialog title',          // string shown in modal header
  apiPath: '/admin/api/v1/...',   // GET endpoint to fetch data from
  transform: (data) => [...],     // maps API .data into [{label, value}, ...]
  onSelect: (value) => { ... },   // called with selected value (string)
                                  // or array of values if multi:true
  multi: false,                   // optional, enables multi-select with checkboxes
});

SelectDialog.close();             // close programmatically
```

`transform(data)` receives the parsed `APIResult.data` field and must return an array of `{label: string, value: string}` objects.

In **single-select** mode (default), clicking an item calls `onSelect(value)` and closes the dialog immediately. In **multi-select** mode, items show checkboxes and a "Select" button appears in the footer.

A search/filter input is always shown. The dialog closes on Cancel, the x button, or clicking the overlay backdrop.

#### Common usage: inject a color pattern name into a text input

```js
SelectDialog.open({
  title: 'Pick a Color Pattern',
  apiPath: '/admin/api/v1/colorpatterns',
  transform: data => Object.keys(data).sort().map(k => ({ label: k, value: k })),
  onSelect: value => {
    myInput.value += ':' + value;
  },
});
```

### ScriptEditor (`static/js/script-editor.js`)

Overlays syntax highlighting on a `<textarea>` for editing JavaScript. Supports tab indentation (4 spaces) and synchronized scroll.

```js
const syncHighlight = ScriptEditor.init('scriptTextarea');
// After programmatically changing textarea value:
syncHighlight();
```

---

## API Documentation Pages

Every page that uses API endpoints must have a matching `-api.html` file (e.g. `items.html` -> `items-api.html`).

### Page structure

1. **Title**: `<h1>` with the resource name + "API Reference" (e.g. "Items API Reference").
2. **Subtitle**: `<p class="subtitle">` — one sentence describing what the endpoints manage.
3. **Endpoint list**: a `<div class="api-list">` containing one `<details class="api-entry">` per endpoint.

### Per-endpoint entry

Each `<details class="api-entry">` contains:

1. **Summary line** (always visible):
   - **Method badge**: `<span class="method method-{method}">` — one of `method-get`, `method-post`, `method-patch`, `method-put`, `method-delete`.
   - **Path**: `<span class="api-path">` — the full route (e.g. `/admin/api/v1/items/{itemId}`).
   - **Description**: `<span class="api-desc">` — short phrase describing the action.

2. **Body** (inside `<div class="api-body">`):
   - **Description paragraph**: `<p>` explaining what the endpoint does, any special parameters, and behavior notes.
   - **curl example**: a `<div class="curl-block">` showing a complete curl command. Use `{{.CONFIG.FilePaths.WebDomain}}` for the host. Color spans: `.kw` for the command name, `.flag` for flags, `.str` for string values.
   - **Response examples** (required): a `<div class="response-examples">` containing one or more `<details class="resp-entry">` blocks:
     - Each has a **summary** with a status badge (`status-ok` or `status-err`) and a label (e.g. "Success", "Not Found").
     - Each has a **body** with a `<div class="curl-block">` showing an example JSON response.

### Template

```html
{{template "header" .}}

<style>
    /* Copy the standard API page styles from any existing -api.html */
</style>

<h1>Things API Reference</h1>
<p class="subtitle">REST endpoints for managing things.</p>

<div class="api-list">
    <details class="api-entry">
        <summary>
            <span class="method method-get">GET</span>
            <span class="api-path">/admin/api/v1/things</span>
            <span class="api-desc">Return all things</span>
        </summary>
        <div class="api-body">
            <p>Description of the endpoint.</p>
            <div class="curl-block">
                <span class="kw">curl</span> <span class="flag">-s</span>
                <span class="flag">-u</span> <span class="str">admin:password</span> \
     <span class="str">http://{{.CONFIG.FilePaths.WebDomain}}/admin/api/v1/things</span>
            </div>
            <div class="response-examples">
                <details class="resp-entry">
                    <summary>
                        <span class="method resp-label status-ok">200</span>
                        <span class="resp-label">Success</span>
                    </summary>
                    <div class="resp-body">
                        <div class="curl-block">{"success":true,"data":{...}}</div>
                    </div>
                </details>
            </div>
        </div>
    </details>
    <!-- more endpoints -->
</div>

{{template "footer" .}}
```

### Required response examples

Every endpoint must include at minimum:

| Method | Required responses |
|---|---|
| GET (collection) | 200 Success |
| GET (single) | 200 Success, 404 Not Found |
| POST | 200/201 Success, 400 Bad Request |
| PATCH | 200 Success, 400 Bad Request, 404 Not Found |
| PUT | 200 Success, 400 Bad Request, 404 Not Found |
| DELETE | 200 Success, 404 Not Found |

Additional error codes (409 Conflict, 500 Internal Server Error) should be documented when the endpoint can produce them.

### CSS classes

All `-api.html` pages share the same `<style>` block. The required classes are:

| Class | Purpose |
|---|---|
| `.api-list` | Flex column container for all entries |
| `.api-entry` | Collapsible `<details>` wrapper |
| `.method` | Base style for HTTP method badges |
| `.method-get` | Green badge |
| `.method-post` | Blue badge |
| `.method-patch` | Orange badge |
| `.method-put` | Orange badge (same as patch) |
| `.method-delete` | Red badge |
| `.api-path` | Monospace route path |
| `.api-desc` | Right-aligned description text |
| `.api-body` | Expanded content area |
| `.curl-block` | Dark code block for curl commands and JSON |
| `.response-examples` | Container for response entries |
| `.resp-entry` | Collapsible response example |
| `.status-ok`, `.status-err` | Green/red status badges |

### Curl block syntax highlighting

| Class | Usage | Color |
|---|---|---|
| `.kw` | `curl`, HTTP methods | Blue `#79c0ff` |
| `.flag` | `-s`, `-X`, `-H`, `-d` | Purple `#d2a8ff` |
| `.str` | URLs, JSON payloads, strings | Light blue `#a5d6ff` |

---

## Reusable UI Components

### Tabs

```html
<div class="tab-bar">
    <button class="tab-btn active" onclick="switchTab('general')">General</button>
    <button class="tab-btn" onclick="switchTab('advanced')">Advanced</button>
</div>
<div id="tab-general" class="tab-panel active">...</div>
<div id="tab-advanced" class="tab-panel">...</div>
```

```css
.tab-bar { display: flex; gap: 0; border-bottom: 2px solid #ddd; margin-bottom: 1rem; }
.tab-btn {
    padding: 0.5rem 1rem; border: none; background: none;
    cursor: pointer; font-size: 0.85rem; color: #666;
    border-bottom: 2px solid transparent; margin-bottom: -2px;
}
.tab-btn.active { color: #1a1a2e; border-bottom-color: #1a1a2e; font-weight: 600; }
.tab-panel { display: none; }
.tab-panel.active { display: block; }
```

### Collapsible Sections

```html
<div class="help-group">
    <div class="help-group-header" onclick="this.parentElement.classList.toggle('open')">
        Section Title
    </div>
    <div class="help-group-body">
        <!-- content -->
    </div>
</div>
```

### Chips (Tags with Remove)

```html
<div class="buff-ids-wrap">
    <span class="chip">
        Buff #5
        <span class="x" onclick="removeBuff(5)">&times;</span>
    </span>
</div>
```

### Stat Modifier Grid

Used by items and buffs for editing stat modifications. Groups stats by category with labeled inputs.

### Flags Grid

Grid of labeled checkboxes for boolean flags:

```html
<div class="flags-grid">
    <label class="flag-item">
        <input type="checkbox" value="hidden" />
        <span class="flag-name">hidden</span>
        <span class="flag-desc">Not visible in buff list</span>
    </label>
</div>
```

---

## Checklist for New Pages

1. Start with `{{template "header" .}}` and end with `{{template "footer" .}}`.
2. Add a `<style>` block with page-specific CSS. Copy the standard patterns from an existing page like `races.html` (simplest CRUD example) or `items.html` (full-featured example).
3. Add `<h1>` title and `<p class="subtitle">` with a link to the API docs page.
4. Use the two-column layout for CRUD pages, or full-width for specialized UIs.
5. Add a `<script>` block with the standard lifecycle: load, render, filter, select, populate, save, delete, showStatus, escHtml, escAttr.
6. Use `AdminAPI` for all HTTP requests. Never use raw `fetch`.
7. Create the matching `pagename-api.html` documenting all endpoints.
8. Register the Go handler and route in `internal/web/`. Add the page to the navigation entries so it appears in the nav bar.

---

## Reference: Existing Pages by Complexity

| Page | Layout | Complexity | Good template for... |
|---|---|---|---|
| `index.html` | Cards grid | Simple | Dashboard/summary pages |
| `users.html` | Full-width table | Simple | Search/list-only pages |
| `https.html` | Full-width | Simple | Info/status pages |
| `races.html` | Sidebar + editor | Medium | Basic CRUD with form fields |
| `colorpatterns.html` | Sidebar + editor | Medium | CRUD with preview/visualization |
| `quests.html` | Sidebar + editor | Complex | Nested/dynamic sub-editors (steps) |
| `buffs.html` | Sidebar + editor | Complex | Stat mods, flags, script editor |
| `items.html` | Sidebar + editor | Complex | All features (stat mods, scripts, SelectDialog, chips) |
| `keywords.html` | Full-width + tabs | Complex | Tabbed UIs with multiple data types |
| `config.html` | Full-width table | Complex | Inline editing, staged changes |

---

## Detailed File Reference

### `_header.html`

Defines the Go template named `header`. Included by every admin page via `{{template "header" .}}`.

Contents:
- Minimal HTML5 shell (`<!DOCTYPE html>`, `<head>`, `<body>`)
- Inline CSS reset and base styles (system font stack, nav bar, `<main>` container)
- Top navigation bar built from `.NAV` template data, with support for nested dropdowns via `.SubItems`
- `<script>` tags loading shared JS libraries from `/admin/static/js/` — `api.js`, `ansi-colors.js`, `select-dialog.js`, `script-editor.js`, `highlight.js`, and `ansitags.js`
- `<link>` to `/admin/static/css/highlight.css`

### `_footer.html`

Defines the Go template named `footer`. Included by every admin page via `{{template "footer" .}}`.

Contents: closing `</main>`, `</body>`, `</html>` tags.

### `index.html`

Admin dashboard page. Rendered by `adminIndex` at `GET /admin/`.

Template data: `.CONFIG` (configs.Config), `.STATS` (web.Stats).

Sections:
- **Header**: MUD name and version from `.CONFIG`
- **Stats grid** (3 cards): Telnet ports, WebSocket port, SSH port — each with connection count from `.STATS`
- **API reference card**: link to Config API docs

### `config.html`

Live configuration editor page. Rendered by `adminConfig` at `GET /admin/config`.

Template data: `.CONFIG` (used only for the subtitle), `.STATS`.

Behaviour:
1. On page load, calls `AdminAPI.get('/admin/api/v1/config')` and builds a sortable, section-grouped table.
2. **Inline editing**: clicking a non-redacted value shows an `<input>` with Save/Cancel. Enter stages the edit; Escape cancels.
3. **Staging**: edits accumulate in a `pending` map with yellow-highlighted rows. A **Pending Changes** panel lists staged changes.
4. **Apply**: `applyChanges()` calls `AdminAPI.patch('/admin/api/v1/config', pending)`. Rejected (locked) keys remain staged.
5. **Discard**: `discardAll()` reverts all staged changes.
6. **Filtering**: search input filters by key/value; section dropdown filters by prefix.
7. **Locked keys**: displayed with a `locked` badge, not clickable.
8. **Redacted secrets**: displayed in italic grey, no click handler.

JavaScript functions (all `window`-scoped):
- `filterTable()`, `startEdit(key)`, `cancelEdit(key)`, `editKeydown(event, key)`, `stageEdit(key)`, `removePending(key)`, `discardAll()`, `applyChanges()`

Utility helpers: `escHtml(s)`, `escAttr(s)`, `cssId(key)`.

## Template Rendering Flow

All admin pages follow the same pattern:

```
handler parses:
  _header.html  (defines "header")
  <page>.html   (defines page body, calls {{template "header" .}} and {{template "footer" .}})
  _footer.html  (defines "footer")

template.New("<page>.html").Funcs(funcMap).ParseFiles(...).Execute(w, templateData)
```

`Cache-Control: no-store` is set on every response to prevent stale admin pages.

## Relationship to Go Source

| HTML file | Go handler | Route |
|---|---|---|
| `index.html` | `internal/web/admin.go` `adminIndex` | `GET /admin/` |
| `config.html` | `internal/web/admin_config.go` `adminConfig` | `GET /admin/config` |
| `static/js/*.js` | `internal/web/web.go` `serveAdminStaticFile` | `GET /admin/static/{path...}` |
| `_header.html`, `_footer.html` | included by both page handlers | — |
