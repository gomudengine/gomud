// SelectDialog — reusable search-and-select modal for admin pages.
//
// Usage (single select):
//   SelectDialog.open({
//     title: 'Select a Color Pattern',
//     apiPath: '/admin/api/v1/colorpatterns',
//     transform: (data) => Object.keys(data).sort().map(k => ({ label: k, value: k })),
//     onSelect: (value) => { myInput.value = ':' + value; },
//   });
//
// Usage (multi-select):
//   SelectDialog.open({
//     title: 'Select Buffs',
//     apiPath: '/admin/api/v1/buffs',
//     transform: (data) => Object.entries(data.specs).map(([id, s]) => ({ label: s.Name, value: id })),
//     onSelect: (values) => { console.log(values); },  // values is an array
//     multi: true,
//   });
//
// transform(data) receives the parsed JSON .data field and must return [{label, value, html?}, ...].
//   - label: displayed text (also used for search filtering)
//   - value: passed to onSelect
//   - html (optional): raw HTML string rendered after the label (e.g. color swatches)
// onSelect(value|values) is called with the selected value (string) or array of values (multi).
// The dialog is closed automatically after onSelect is called.

const SelectDialog = (() => {
    let overlay = null;

    function injectStyles() {
        if (document.getElementById('select-dialog-styles')) return;
        const style = document.createElement('style');
        style.id = 'select-dialog-styles';
        style.textContent = `
            .sd-overlay {
                position: fixed; inset: 0; background: rgba(0,0,0,0.45);
                display: flex; align-items: center; justify-content: center;
                z-index: 9999;
            }
            .sd-modal {
                background: #fff; border-radius: 8px; box-shadow: 0 8px 32px rgba(0,0,0,0.25);
                width: 480px; max-width: 95vw; max-height: 80vh;
                display: flex; flex-direction: column; overflow: hidden;
            }
            .sd-header {
                padding: 0.85rem 1rem 0.7rem; border-bottom: 1px solid #e5e5e5;
                display: flex; align-items: center; justify-content: space-between;
            }
            .sd-title { font-size: 1rem; font-weight: 700; color: #1a1a2e; }
            .sd-close {
                background: none; border: none; font-size: 1.25rem; cursor: pointer;
                color: #888; line-height: 1; padding: 0 0.2rem;
            }
            .sd-close:hover { color: #333; }
            .sd-search-wrap { padding: 0.6rem 1rem; border-bottom: 1px solid #eee; }
            .sd-search {
                width: 100%; padding: 0.4rem 0.65rem; border: 1px solid #ccc;
                border-radius: 4px; font-size: 0.875rem;
            }
            .sd-search:focus { outline: 2px solid #1a1a2e; outline-offset: 1px; border-color: transparent; }
            .sd-list { overflow-y: auto; flex: 1; padding: 0.35rem 0; }
            .sd-item {
                padding: 0.5rem 1rem; cursor: pointer; font-size: 0.875rem;
                display: flex; align-items: center; gap: 0.6rem; user-select: none;
            }
            .sd-item:hover { background: #f5f7ff; }
            .sd-item.sd-selected { background: #1a1a2e; color: #fff; }
            .sd-item input[type="checkbox"] { accent-color: #1a1a2e; flex-shrink: 0; }
            .sd-item-detail { margin-left: auto; flex-shrink: 0; display: flex; gap: 1px; align-items: center; font-size: 0.8rem; }
            .sd-item-detail span { display: inline-block; width: 8px; height: 14px; border-radius: 1px; }
            .sd-empty { padding: 2rem 1rem; text-align: center; color: #aaa; font-size: 0.85rem; }
            .sd-loading { padding: 2rem 1rem; text-align: center; color: #888; font-size: 0.85rem; }
            .sd-footer {
                padding: 0.65rem 1rem; border-top: 1px solid #eee;
                display: flex; justify-content: flex-end; gap: 0.5rem;
            }
            .sd-btn {
                padding: 0.4rem 1rem; border-radius: 4px; font-size: 0.85rem;
                font-weight: 600; cursor: pointer; border: 1px solid transparent;
            }
            .sd-btn-cancel { background: #f0f0f0; color: #444; border-color: #ccc; }
            .sd-btn-cancel:hover { background: #e5e5e5; }
            .sd-btn-confirm { background: #1a1a2e; color: #fff; }
            .sd-btn-confirm:hover { background: #2d2d4e; }
        `;
        document.head.appendChild(style);
    }

    function close() {
        if (overlay) {
            overlay.remove();
            overlay = null;
        }
    }

    function open({ title, apiPath, transform, onSelect, multi = false }) {
        injectStyles();
        close();

        overlay = document.createElement('div');
        overlay.className = 'sd-overlay';
        overlay.addEventListener('click', e => { if (e.target === overlay) close(); });

        const modal = document.createElement('div');
        modal.className = 'sd-modal';

        const header = document.createElement('div');
        header.className = 'sd-header';
        header.innerHTML = `<span class="sd-title">${title}</span>`;
        const closeBtn = document.createElement('button');
        closeBtn.className = 'sd-close';
        closeBtn.textContent = '×';
        closeBtn.addEventListener('click', close);
        header.appendChild(closeBtn);

        const searchWrap = document.createElement('div');
        searchWrap.className = 'sd-search-wrap';
        const searchInput = document.createElement('input');
        searchInput.type = 'search';
        searchInput.className = 'sd-search';
        searchInput.placeholder = 'Search…';
        searchWrap.appendChild(searchInput);

        const list = document.createElement('div');
        list.className = 'sd-list';
        list.innerHTML = '<div class="sd-loading">Loading…</div>';

        const footer = document.createElement('div');
        footer.className = 'sd-footer';
        const cancelBtn = document.createElement('button');
        cancelBtn.className = 'sd-btn sd-btn-cancel';
        cancelBtn.textContent = 'Cancel';
        cancelBtn.addEventListener('click', close);
        footer.appendChild(cancelBtn);

        let confirmBtn = null;
        if (multi) {
            confirmBtn = document.createElement('button');
            confirmBtn.className = 'sd-btn sd-btn-confirm';
            confirmBtn.textContent = 'Select';
            footer.appendChild(confirmBtn);
        }

        modal.append(header, searchWrap, list, footer);
        overlay.appendChild(modal);
        document.body.appendChild(overlay);
        searchInput.focus();

        let allItems = [];
        const selectedValues = new Set();

        function renderList(filter) {
            const q = filter.toLowerCase();
            const visible = allItems.filter(it => it.label.toLowerCase().includes(q));
            if (visible.length === 0) {
                list.innerHTML = '<div class="sd-empty">No results</div>';
                return;
            }
            list.innerHTML = '';
            for (const item of visible) {
                const row = document.createElement('div');
                row.className = 'sd-item';
                if (!multi && selectedValues.has(item.value)) row.classList.add('sd-selected');

                if (multi) {
                    const cb = document.createElement('input');
                    cb.type = 'checkbox';
                    cb.checked = selectedValues.has(item.value);
                    cb.addEventListener('change', () => {
                        if (cb.checked) selectedValues.add(item.value);
                        else selectedValues.delete(item.value);
                    });
                    row.appendChild(cb);
                }

                const lbl = document.createElement('span');
                lbl.textContent = item.label;
                row.appendChild(lbl);

                if (item.html) {
                    const detail = document.createElement('span');
                    detail.className = 'sd-item-detail';
                    detail.innerHTML = item.html;
                    row.appendChild(detail);
                }

                if (!multi) {
                    row.addEventListener('click', () => {
                        close();
                        onSelect(item.value);
                    });
                } else {
                    row.addEventListener('click', e => {
                        if (e.target.type === 'checkbox') return;
                        const cb = row.querySelector('input[type="checkbox"]');
                        cb.checked = !cb.checked;
                        if (cb.checked) selectedValues.add(item.value);
                        else selectedValues.delete(item.value);
                    });
                }

                list.appendChild(row);
            }
        }

        searchInput.addEventListener('input', () => renderList(searchInput.value));

        if (confirmBtn) {
            confirmBtn.addEventListener('click', () => {
                close();
                onSelect([...selectedValues]);
            });
        }

        AdminAPI.get(apiPath).then(res => {
            if (!res.ok) {
                list.innerHTML = `<div class="sd-empty">Error: ${res.error || 'Failed to load'}</div>`;
                return;
            }
            allItems = transform(res.data && res.data.data);
            renderList('');
        });
    }

    return { open, close };
})();
