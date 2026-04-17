
const Triggers = (() => {

    const STORAGE_KEY = 'triggers';

    // -----------------------------------------------------------------------
    // Built-in (default) triggers.
    // These are merged into storage on first load, starting disabled.
    // Edit the body string to change what the trigger does.
    // -----------------------------------------------------------------------
    const _defaults = [
        {
            pattern: 'On the Ground: {number} gold',
            body: [
                'Client.SendInput("get gold");',
            ].join('\n'),
        },
    ];

    // -----------------------------------------------------------------------
    // Storage
    // Each record: { pattern, body, enabled, isDefault }
    // -----------------------------------------------------------------------
    function _load() {
        try {
            const raw = localStorage.getItem(STORAGE_KEY);
            if (!raw) { return null; }
            const parsed = JSON.parse(raw);
            return Array.isArray(parsed) ? parsed : null;
        } catch (e) {
            return null;
        }
    }

    function _save(list) {
        try {
            localStorage.setItem(STORAGE_KEY, JSON.stringify(list));
        } catch (e) {
            console.warn('Triggers: could not save to localStorage', e);
        }
    }

    // Merge defaults into a stored list: add any default whose pattern is not
    // already present. Existing entries (including user edits) are untouched.
    function _mergeDefaults(list) {
        _defaults.forEach(def => {
            const exists = list.some(t => t.pattern === def.pattern && t.isDefault);
            if (!exists) {
                list.push({ pattern: def.pattern, body: def.body, enabled: false, isDefault: true });
            }
        });
        return list;
    }

    // Initialise: load from storage, merge defaults, persist.
    let _triggers = _mergeDefaults(_load() || []);
    _save(_triggers);

    // -----------------------------------------------------------------------
    // Utilities
    // -----------------------------------------------------------------------
    function ParseNumber(value, locales = navigator.languages) {
        const example = Intl.NumberFormat(locales).format('1.1');
        const cleanPattern = new RegExp(`[^-+0-9${ example.charAt(1) }]`, 'g');
        const cleaned = value.replace(cleanPattern, '');
        const normalized = cleaned.replace(example.charAt(1), '.');
        return parseFloat(normalized);
    }

    function matchPattern(pattern, str) {
        const escapeRegex = s => s.replace(/[-/\\^$+?.()|[\]{}]/g, '\\$&');
        let groupTypes = [];
        let regexPattern = escapeRegex(pattern);

        regexPattern = regexPattern.replace(/\\\{text\\\}/g, () => {
            groupTypes.push('text');
            return '(.+?)';
        });
        regexPattern = regexPattern.replace(/\\\{number\\\}/g, () => {
            groupTypes.push('number');
            return '([-+]?\\d[\\d,]*(?:\\.\\d+)?)';
        });

        const match = str.match(new RegExp(regexPattern));
        if (!match) { return null; }

        return match.slice(1).map((value, i) => {
            if (groupTypes[i] === 'number') { return ParseNumber(value); }
            return value;
        });
    }

    function stripAnsi(str) {
        return str.replace(/\x1B\[[0-?]*[ -/]*[@-~]/g, '');
    }

    // Validate a function body string. Returns null on success, error message on failure.
    function validateBody(body) {
        try {
            // eslint-disable-next-line no-new-func
            new Function('matches', body);
            return null;
        } catch (e) {
            return e.message;
        }
    }

    // -----------------------------------------------------------------------
    // Public API
    // -----------------------------------------------------------------------

    // Run all enabled triggers against a line of output.
    function Try(str) {
        str = stripAnsi(str);
        _triggers.forEach(trigger => {
            if (!trigger.enabled) { return; }
            const matches = matchPattern(trigger.pattern, str);
            if (!matches) { return; }
            try {
                // eslint-disable-next-line no-new-func
                const fn = new Function('matches', trigger.body);
                fn(matches);
            } catch (e) {
                console.warn('Trigger error [' + trigger.pattern + ']:', e);
            }
        });
    }

    // Return a shallow copy of the trigger list.
    function getTriggers() {
        return _triggers.map(t => Object.assign({}, t));
    }

    // Update the pattern and body of an existing trigger by index.
    // Returns null on success, error string on validation failure.
    function saveTrigger(idx, pattern, body) {
        const err = validateBody(body);
        if (err) { return err; }
        if (idx < 0 || idx >= _triggers.length) { return 'Invalid trigger index.'; }
        _triggers[idx].pattern = pattern;
        _triggers[idx].body    = body;
        _save(_triggers);
        return null;
    }

    // Add a new user trigger. Returns null on success, error string on failure.
    function addTrigger(pattern, body) {
        const err = validateBody(body);
        if (err) { return err; }
        if (!pattern.trim()) { return 'Pattern cannot be empty.'; }
        _triggers.push({ pattern, body, enabled: true, isDefault: false });
        _save(_triggers);
        return null;
    }

    // Remove a trigger by index.
    function removeTrigger(idx) {
        if (idx < 0 || idx >= _triggers.length) { return; }
        _triggers.splice(idx, 1);
        _save(_triggers);
    }

    // Enable or disable a trigger by index.
    function setEnabled(idx, enabled) {
        if (idx < 0 || idx >= _triggers.length) { return; }
        _triggers[idx].enabled = !!enabled;
        _save(_triggers);
    }

    return {
        Try,
        getTriggers,
        saveTrigger,
        addTrigger,
        removeTrigger,
        setEnabled,
        validateBody,
        matchPattern,
        stripAnsi,
        ParseNumber,
    };

})();
