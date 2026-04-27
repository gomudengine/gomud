/**
 * GoMud Admin API client library.
 *
 * Usage
 * -----
 * Single request:
 *   const res = await AdminAPI.get('/admin/api/v1/config');
 *   const res = await AdminAPI.get('/admin/api/v1/stats/memory', true); // bypass cache
 *   const res = await AdminAPI.patch('/admin/api/v1/config', { 'Server.MudName': 'My MUD' });
 *
 * Parallel requests (all settle before the callback fires):
 *   const [cfg, stats] = await AdminAPI.all([
 *     AdminAPI.get('/admin/api/v1/config'),
 *     AdminAPI.get('/admin/api/v1/stats'),
 *   ]);
 *
 * Queue builder (fluent):
 *   const results = await AdminAPI.queue()
 *     .get('/admin/api/v1/config')
 *     .patch('/admin/api/v1/config', { 'Server.MudName': 'My MUD' })
 *     .run();
 */

const AdminAPI = (() => {
    'use strict';

    const CACHE_TTL_MS = 300000; // how long GET results are cached (ms)
    const CACHE_PREFIX = 'adminapi_cache:';

    function _cacheKey(path) {
        return CACHE_PREFIX + path;
    }

    function _cacheGet(path) {
        try {
            const raw = sessionStorage.getItem(_cacheKey(path));
            if (!raw) return undefined;
            return JSON.parse(raw);
        } catch (_) {
            return undefined;
        }
    }

    function _cacheSet(path, entry) {
        try {
            sessionStorage.setItem(_cacheKey(path), JSON.stringify(entry));
        } catch (_) {
            // sessionStorage full or unavailable; skip caching
        }
    }

    function _cacheDelete(path) {
        try {
            sessionStorage.removeItem(_cacheKey(path));
        } catch (_) {}
    }

    function _cacheKeys() {
        const keys = [];
        try {
            for (let i = 0; i < sessionStorage.length; i++) {
                const k = sessionStorage.key(i);
                if (k && k.startsWith(CACHE_PREFIX)) {
                    keys.push(k.slice(CACHE_PREFIX.length));
                }
            }
        } catch (_) {}
        return keys;
    }

    // /admin/api/v1/items/attack-messages -> /admin/api/v1/items
    function _rootPath(path) {
        return path.split('/').slice(0, 5).join('/');
    }

    function _invalidate(path) {
        const prefix = _rootPath(path);
        for (const key of _cacheKeys()) {
            if (key.startsWith(prefix)) {
                console.log('Cache invalidated:', key);
                _cacheDelete(key);
            }
        }
    }

    /**
     * @typedef {Object} APIResult
     * @property {boolean} ok          - true when HTTP status is 2xx
     * @property {number}  status      - HTTP status code
     * @property {*}       data        - parsed JSON body (null on parse failure)
     * @property {string}  error       - error message, or empty string on success
     */

    /**
     * Core fetch wrapper. Returns a resolved APIResult regardless of outcome so
     * callers never need to catch network errors themselves.
     *
     * @param {string} method
     * @param {string} path
     * @param {Object|null} body
     * @returns {Promise<APIResult>}
     */
    async function request(method, path, body = null) {
        const init = {
            method,
            headers: { 'Content-Type': 'application/json' },
            credentials: 'same-origin',
        };

        if (body !== null) {
            init.body = JSON.stringify(body);
        }

        let status = 0;
        let data = null;
        let error = '';

        try {
            const response = await fetch(path, init);
            status = response.status;

            const text = await response.text();
            if (text.length > 0) {
                try {
                    data = JSON.parse(text);
                } catch (_) {
                    data = text;
                }
            }

            if (!response.ok) {
                error = (data && data.error) ? data.error : `HTTP ${status}`;
            }
        } catch (networkError) {
            error = networkError.message || 'Network error';
        }

        return { ok: status >= 200 && status < 300, status, data, error };
    }

    /**
     * GET request.
     * @param {string}  path
     * @param {boolean} [bypassCache=false] - when true, skips the cache read and
     *                                        writes the fresh result back into it.
     * @returns {Promise<APIResult>}
     */
    async function get(path, bypassCache = false) {
        if (!bypassCache) {
            const cached = _cacheGet(path);
            if (cached && Date.now() < cached.expires) {
                console.log('Loaded from cache:', path);
                return cached.result;
            }
        }

        const result = await request('GET', path);
        if (result.ok) {
            console.log('Loaded new:', path);
            _cacheSet(path, { result, expires: Date.now() + CACHE_TTL_MS });
        }
        return result;
    }

    /**
     * POST request.
     * @param {string} path
     * @param {Object} body
     * @returns {Promise<APIResult>}
     */
    async function post(path, body) {
        const result = await request('POST', path, body);
        _invalidate(path);
        return result;
    }

    /**
     * PUT request.
     * @param {string} path
     * @param {Object} body
     * @returns {Promise<APIResult>}
     */
    function put(path, body) {
        const result = request('PUT', path, body);
        _invalidate(path);
        return result;
    }

    /**
     * PATCH request.
     * @param {string} path
     * @param {Object} body
     * @returns {Promise<APIResult>}
     */
    async function patch(path, body) {
        const result = await request('PATCH', path, body);
        _invalidate(path);
        return result;
    }

    /**
     * DELETE request.
     * @param {string} path
     * @param {Object|null} body
     * @returns {Promise<APIResult>}
     */
    async function del(path, body = null) {
        const result = await request('DELETE', path, body);
        _invalidate(path);
        return result;
    }

    /**
     * Wait for all provided request promises to settle, then return their results
     * in the same order. Never rejects — failed requests are represented as
     * APIResult objects with ok=false.
     *
     * @param {Array<Promise<APIResult>>} promises
     * @returns {Promise<Array<APIResult>>}
     */
    function all(promises) {
        return Promise.all(promises);
    }

    /**
     * Fluent request queue. Collects requests and dispatches them all in parallel
     * when run() is called.
     *
     * @returns {{get, post, put, patch, delete, run}}
     */
    function queue() {
        const pending = [];

        const q = {
            /**
             * @param {string}  path
             * @param {boolean} [bypassCache=false]
             * @returns {typeof q}
             */
            get(path, bypassCache = false) {
                pending.push(get(path, bypassCache));
                return q;
            },
            /**
             * @param {string} path
             * @param {Object} body
             * @returns {typeof q}
             */
            post(path, body) {
                pending.push(post(path, body));
                return q;
            },
            /**
             * @param {string} path
             * @param {Object} body
             * @returns {typeof q}
             */
            put(path, body) {
                pending.push(put(path, body));
                return q;
            },
            /**
             * @param {string} path
             * @param {Object} body
             * @returns {typeof q}
             */
            patch(path, body) {
                pending.push(patch(path, body));
                return q;
            },
            /**
             * @param {string} path
             * @param {Object|null} body
             * @returns {typeof q}
             */
            delete(path, body = null) {
                pending.push(del(path, body));
                return q;
            },
            /**
             * Dispatch all queued requests in parallel and resolve with an array
             * of APIResult objects in the order they were queued.
             *
             * @returns {Promise<Array<APIResult>>}
             */
            run() {
                return all(pending.splice(0));
            },
        };

        return q;
    }

    return { get, post, put, patch, delete: del, all, queue, request };
})();
