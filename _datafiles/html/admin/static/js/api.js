/**
 * GoMud Admin API client library.
 *
 * Usage
 * -----
 * Single request:
 *   const res = await AdminAPI.get('/admin/api/v1/config');
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
     * @param {string} path
     * @returns {Promise<APIResult>}
     */
    function get(path) {
        return request('GET', path);
    }

    /**
     * POST request.
     * @param {string} path
     * @param {Object} body
     * @returns {Promise<APIResult>}
     */
    function post(path, body) {
        return request('POST', path, body);
    }

    /**
     * PUT request.
     * @param {string} path
     * @param {Object} body
     * @returns {Promise<APIResult>}
     */
    function put(path, body) {
        return request('PUT', path, body);
    }

    /**
     * PATCH request.
     * @param {string} path
     * @param {Object} body
     * @returns {Promise<APIResult>}
     */
    function patch(path, body) {
        return request('PATCH', path, body);
    }

    /**
     * DELETE request.
     * @param {string} path
     * @param {Object|null} body
     * @returns {Promise<APIResult>}
     */
    function del(path, body = null) {
        return request('DELETE', path, body);
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
             * @param {string} path
             * @returns {typeof q}
             */
            get(path) {
                pending.push(get(path));
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
