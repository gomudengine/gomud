// Standard Picker configurations for common admin entity types.
//
// Each entry is a plain object that can be spread into Picker.open():
//
//   Picker.open({
//     ...PickerConfigs.buffs,
//     onSelect: (buff) => appendBuffChip(listEl, buff.BuffId, buff.Name),
//   });
//
//   Picker.open({
//     ...PickerConfigs.buffs,
//     title:    'Select Worn Buff',
//     filter:   b => !b.Secret,
//     multi:    true,
//     selected: collectBuffIds('wornbuffids-list'),
//     onSelect: (buffs) => replaceBuffChips('wornbuffids-list', buffs),
//   });

const PickerConfigs = {

    buffs: {
        title:   'Select Buff',
        idKey:   'BuffId',
        columns: [
            { key: 'BuffId',      label: '#',    width: '4rem',  mono: true },
            { key: 'Name',        label: 'Name', flex: true },
            { key: 'TriggerRate', label: 'Rate', width: '6rem' },
        ],
        searchKeys: ['BuffId', 'Name'],
        sort: (a, b) => a.BuffId - b.BuffId,
        // buffs API returns { data: { specs: { ... } } } — unwrap via source fn
        source: async () => {
            const res = await AdminAPI.get('/admin/api/v1/buffs');
            if (!res.ok) throw new Error(res.error || 'Failed to load buffs');
            const specs = (res.data && res.data.data && res.data.data.specs) || {};
            return Object.values(specs).sort((a, b) => a.BuffId - b.BuffId);
        },
    },

    items: {
        title:   'Select Item',
        source:  '/admin/api/v1/items',
        idKey:   'ItemId',
        columns: [
            { key: 'ItemId', label: '#',    width: '4rem', mono: true },
            { key: 'Name',   label: 'Name', flex: true },
            { key: 'Type',   label: 'Type', width: '8rem' },
        ],
        searchKeys: ['ItemId', 'Name'],
        sort: (a, b) => a.ItemId - b.ItemId,
    },

    quests: {
        title:   'Select Quest',
        source:  '/admin/api/v1/quests',
        idKey:   'QuestId',
        columns: [
            { key: 'QuestId', label: '#',    width: '4rem', mono: true },
            { key: 'Name',    label: 'Name', flex: true },
        ],
        searchKeys: ['QuestId', 'Name'],
        sort: (a, b) => a.QuestId - b.QuestId,
    },

    mobs: {
        title:   'Select Mob',
        source:  '/admin/api/v1/mobs',
        idKey:   'MobId',
        columns: [
            { key: 'MobId', label: '#',    width: '4rem', mono: true },
            { key: 'Name',  label: 'Name', flex: true,
              render: (_, item) => (item.Character && item.Character.Name) || '(unnamed)' },
            { key: 'Zone',  label: 'Zone', width: '10rem' },
        ],
        searchKeys: ['MobId', 'Zone'],
        sort: (a, b) => a.MobId - b.MobId,
    },

};

// buffName(id) — synchronously resolves a buff name from the AdminAPI cache.
// Returns '#ID Name' if the cache is warm (i.e. the buffs API has already been
// fetched this session), or '#ID' if it hasn't been fetched yet.
PickerConfigs.buffName = function (id) {
    const cached = AdminAPI._buffNameCache || (AdminAPI._buffNameCache = {});
    if (cached[id] !== undefined) return cached[id];
    // Try to read from the already-cached AdminAPI GET response (synchronous).
    const entry = (window._adminBuffSpecs || {})[String(id)];
    if (entry) {
        cached[id] = '#' + id + ' ' + entry.Name;
        return cached[id];
    }
    return '#' + id;
};

// Pre-warm _adminBuffSpecs whenever buffs are fetched so buffName() works
// synchronously on subsequent calls.
AdminAPI.get('/admin/api/v1/buffs').then(res => {
    if (res.ok && res.data && res.data.data && res.data.data.specs) {
        window._adminBuffSpecs = res.data.data.specs;
    }
});

// QuestTokenPicker
//
// Usage:
//   QuestTokenPicker.pick((token, label) => {
//     // token: e.g. "1001-start"
//     // label: e.g. "#1001 The Lost Artifact — start"
//   });
//
// Optional: pass excludeQuestId to filter out the quest currently being edited
// (prevents a quest from referencing its own tokens).
//   QuestTokenPicker.pick(callback, { excludeQuestId: currentQuestId });

const QuestTokenPicker = (() => {
    function pick(onPicked, opts) {
        const exclude = opts && opts.excludeQuestId;

        Picker.open({
            ...PickerConfigs.quests,
            title:  'Select Quest',
            filter: exclude != null ? (q => q.QuestId !== exclude) : undefined,
            onSelect: (quest) => {
                const steps = (quest.Steps || []);

                // Build step items — always include start/end even if Steps is sparse
                const stepItems = steps.map(s => ({
                    Id:          s.Id,
                    Description: s.Description || '',
                }));

                Picker.open({
                    title:   'Select Step — #' + quest.QuestId + ' ' + quest.Name,
                    source:  async () => stepItems,
                    idKey:   'Id',
                    columns: [
                        { key: 'Id',          label: 'Step',        width: '8rem', mono: true },
                        { key: 'Description', label: 'Description', flex: true },
                    ],
                    searchKeys: ['Id', 'Description'],
                    onSelect: (step) => {
                        const token = quest.QuestId + '-' + step.Id;
                        const label = '#' + quest.QuestId + ' ' + quest.Name + ' \u2014 ' + step.Id;
                        onPicked(token, label);
                    },
                });
            },
        });
    }

    return { pick };
})();

// UserPicker — search-driven user picker backed by /admin/api/v1/users/search.
//
// Unlike Picker, this modal does not pre-load all users. Instead it fires a
// debounced search request as the user types and renders live results.
//
// Usage:
//   UserPicker.open({
//     onSelect: (user) => {
//       // user: { user_id, username, role, email }
//     },
//   });

const UserPicker = (() => {
    'use strict';

    let overlay = null;
    let _triggerEl = null;

    function injectStyles() {
        if (document.getElementById('user-picker-styles')) return;
        const style = document.createElement('style');
        style.id = 'user-picker-styles';
        style.textContent = `
            .up-overlay {
                position: fixed; inset: 0; background: rgba(0,0,0,0.45);
                display: flex; align-items: center; justify-content: center;
                z-index: 9999;
            }
            .up-modal {
                background: #fff; border-radius: 8px; box-shadow: 0 8px 32px rgba(0,0,0,0.25);
                width: 520px; max-width: 96vw; max-height: 80vh;
                display: flex; flex-direction: column; overflow: hidden;
            }
            .up-header {
                padding: 0.85rem 1rem 0.7rem; border-bottom: 1px solid #e5e5e5;
                display: flex; align-items: center; justify-content: space-between; flex-shrink: 0;
            }
            .up-title { font-size: 1rem; font-weight: 700; color: #1a1a2e; }
            .up-close {
                background: none; border: none; font-size: 1.25rem; cursor: pointer;
                color: #888; line-height: 1; padding: 0 0.2rem;
            }
            .up-close:hover { color: #333; }
            .up-search-wrap { padding: 0.6rem 1rem; border-bottom: 1px solid #eee; flex-shrink: 0; }
            .up-search {
                width: 100%; padding: 0.4rem 0.65rem; border: 1px solid #ccc;
                border-radius: 4px; font-size: 0.875rem;
            }
            .up-search:focus { outline: 2px solid #1a1a2e; outline-offset: 1px; border-color: transparent; }
            .up-hint { font-size: 0.75rem; color: #aaa; margin-top: 0.3rem; }
            .up-table-wrap { overflow-y: auto; flex: 1; }
            .up-table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
            .up-table thead { position: sticky; top: 0; background: #f5f5f5; z-index: 1; }
            .up-table th {
                padding: 0.4rem 0.75rem; text-align: left; font-size: 0.72rem;
                font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em;
                color: #666; border-bottom: 1px solid #e0e0e0; white-space: nowrap;
            }
            .up-table td { padding: 0.45rem 0.75rem; border-bottom: 1px solid #f0f0f0; vertical-align: middle; }
            .up-table tr:last-child td { border-bottom: none; }
            .up-table tbody tr { cursor: pointer; }
            .up-table tbody tr:hover { background: #f5f7ff; }
            .up-table tbody tr.up-highlighted { background: #eef0ff; }
            .up-role-badge {
                display: inline-block; font-size: 0.7rem; padding: 0.1rem 0.4rem;
                border-radius: 3px; background: #e8eaf6; color: #3949ab;
                font-weight: 600; text-transform: uppercase;
            }
            .up-role-badge.up-role-admin { background: #fde8e8; color: #8a0000; }
            .up-empty { padding: 2rem 1rem; text-align: center; color: #aaa; font-size: 0.875rem; }
            .up-loading { padding: 2rem 1rem; text-align: center; color: #888; font-size: 0.875rem; }
            .up-footer {
                padding: 0.65rem 1rem; border-top: 1px solid #eee;
                display: flex; justify-content: flex-end; flex-shrink: 0;
            }
            .up-btn-cancel {
                padding: 0.4rem 1rem; border-radius: 4px; font-size: 0.85rem;
                font-weight: 600; cursor: pointer; background: #f0f0f0;
                color: #444; border: 1px solid #ccc;
            }
            .up-btn-cancel:hover { background: #e5e5e5; }
        `;
        document.head.appendChild(style);
    }

    function escHtml(s) {
        return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
    }

    function close() {
        if (overlay) { overlay.remove(); overlay = null; }
        if (_triggerEl) { _triggerEl.focus(); _triggerEl = null; }
    }

    function open({ onSelect }) {
        injectStyles();
        close();

        _triggerEl = document.activeElement || null;

        overlay = document.createElement('div');
        overlay.className = 'up-overlay';
        overlay.setAttribute('role', 'dialog');
        overlay.setAttribute('aria-modal', 'true');
        overlay.setAttribute('aria-label', 'Select User');
        overlay.addEventListener('click', e => { if (e.target === overlay) close(); });

        const modal = document.createElement('div');
        modal.className = 'up-modal';

        const header = document.createElement('div');
        header.className = 'up-header';
        const titleEl = document.createElement('span');
        titleEl.className = 'up-title';
        titleEl.textContent = 'Select User';
        const closeBtn = document.createElement('button');
        closeBtn.className = 'up-close';
        closeBtn.textContent = '\u00d7';
        closeBtn.setAttribute('aria-label', 'Close');
        closeBtn.addEventListener('click', close);
        header.appendChild(titleEl);
        header.appendChild(closeBtn);

        const searchWrap = document.createElement('div');
        searchWrap.className = 'up-search-wrap';
        const searchInput = document.createElement('input');
        searchInput.type = 'search';
        searchInput.className = 'up-search';
        searchInput.placeholder = 'Type a username to search\u2026';
        searchInput.setAttribute('aria-label', 'Search users');
        const hint = document.createElement('div');
        hint.className = 'up-hint';
        hint.textContent = 'Type at least 2 characters. Returns exact and prefix matches.';
        searchWrap.appendChild(searchInput);
        searchWrap.appendChild(hint);

        const tableWrap = document.createElement('div');
        tableWrap.className = 'up-table-wrap';
        tableWrap.innerHTML = '<div class="up-empty">Start typing to search for a user.</div>';

        const footer = document.createElement('div');
        footer.className = 'up-footer';
        const cancelBtn = document.createElement('button');
        cancelBtn.className = 'up-btn-cancel';
        cancelBtn.textContent = 'Cancel';
        cancelBtn.addEventListener('click', close);
        footer.appendChild(cancelBtn);

        modal.append(header, searchWrap, tableWrap, footer);
        overlay.appendChild(modal);
        document.body.appendChild(overlay);
        searchInput.focus();

        let highlightIdx = -1;
        let visibleUsers = [];
        let debounceTimer = null;

        function renderResults(users) {
            visibleUsers = users;
            highlightIdx = -1;

            if (!users.length) {
                tableWrap.innerHTML = '<div class="up-empty">No users found.</div>';
                return;
            }

            const table = document.createElement('table');
            table.className = 'up-table';

            const thead = document.createElement('thead');
            thead.innerHTML = '<tr><th style="width:4rem">ID</th><th>Username</th><th style="width:6rem">Role</th></tr>';
            table.appendChild(thead);

            const tbody = document.createElement('tbody');
            users.forEach((u, i) => {
                const tr = document.createElement('tr');
                tr.dataset.idx = i;
                const roleClass = (u.role || '').toLowerCase() === 'admin' ? ' up-role-admin' : '';
                tr.innerHTML =
                    '<td style="font-family:monospace">' + escHtml(String(u.user_id)) + '</td>' +
                    '<td><strong>' + escHtml(u.username || '') + '</strong>' +
                    (u.email ? '<br><span style="font-size:0.78rem;color:#888">' + escHtml(u.email) + '</span>' : '') +
                    '</td>' +
                    '<td><span class="up-role-badge' + roleClass + '">' + escHtml(u.role || 'user') + '</span></td>';
                tr.addEventListener('click', () => { close(); onSelect(u); });
                tbody.appendChild(tr);
            });
            table.appendChild(tbody);

            tableWrap.innerHTML = '';
            tableWrap.appendChild(table);
        }

        function setHighlight(idx) {
            const rows = tableWrap.querySelectorAll('tbody tr');
            if (highlightIdx >= 0 && rows[highlightIdx]) rows[highlightIdx].classList.remove('up-highlighted');
            highlightIdx = Math.max(0, Math.min(idx, visibleUsers.length - 1));
            if (rows[highlightIdx]) {
                rows[highlightIdx].classList.add('up-highlighted');
                rows[highlightIdx].scrollIntoView({ block: 'nearest' });
            }
        }

        async function doSearch(q) {
            if (q.trim().length < 2) {
                tableWrap.innerHTML = '<div class="up-empty">Start typing to search for a user.</div>';
                visibleUsers = [];
                return;
            }
            tableWrap.innerHTML = '<div class="up-loading">Searching\u2026</div>';
            const res = await AdminAPI.get('/admin/api/v1/users/search?name=' + encodeURIComponent(q.trim()), true);
            if (!res.ok) {
                tableWrap.innerHTML = '<div class="up-empty">Search failed: ' + escHtml(res.error || 'unknown error') + '</div>';
                return;
            }
            renderResults((res.data && res.data.data) || []);
        }

        searchInput.addEventListener('input', () => {
            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(() => doSearch(searchInput.value), 200);
        });

        overlay.addEventListener('keydown', e => {
            if (e.key === 'Escape') { e.preventDefault(); close(); return; }
            if (e.key === 'ArrowDown') { e.preventDefault(); setHighlight(highlightIdx < 0 ? 0 : highlightIdx + 1); return; }
            if (e.key === 'ArrowUp')   { e.preventDefault(); setHighlight(highlightIdx <= 0 ? 0 : highlightIdx - 1); return; }
            if (e.key === 'Enter') {
                e.preventDefault();
                if (highlightIdx >= 0 && visibleUsers[highlightIdx]) {
                    close();
                    onSelect(visibleUsers[highlightIdx]);
                }
            }
        });
    }

    return { open, close };
})();
