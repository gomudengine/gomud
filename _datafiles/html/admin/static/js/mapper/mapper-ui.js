/**
 * mapper-ui.js -- UI controls and panels for the map editor.
 *
 * Owns the tab bar, zoom/spacing controls, camera easing, Z-level buttons,
 * room info panel, tooltip, stats line, zone dropdown, local room creation,
 * coordinate conversion, and save/discard workflow. All DOM interaction
 * funnels through a `dom` refs object set at init time so the module stays
 * decoupled from specific element IDs.
 */
/* jshint esversion: 11, browser: true */
/* globals MapperState, MapperRender, MapperEvents, MapperCtxMenu, AdminAPI,
   symbolForRoom, colorForSymbol, escapeHtml, smoothstep,
   getZonesAtPoint, closestZone,
   ZOOM_STEP, ZOOM_MIN, ZOOM_MAX, CENTER_EASE_DURATION, ROOM_SIZE_2D */
'use strict';

var MapperUI = (function() {

    // =====================================================================
    //  DOM references (set via init)
    // =====================================================================

    var dom = {
        viewport: null,
        spacingCtrlEl: null,
        zButtonsEl: null,
        saveCtrlEl: null,
        dirtyCountEl: null,
        changelogEl: null,
        clEntriesEl: null,
        statsEl: null,
        zoneSelect: null,
        loadingEl: null,
        infoEmptyEl: null,
        infoContentEl: null,
        tooltip: null
    };

    function init(domRefs) {
        if (domRefs.viewport) dom.viewport = domRefs.viewport;
        if (domRefs.spacingCtrlEl) dom.spacingCtrlEl = domRefs.spacingCtrlEl;
        if (domRefs.zButtonsEl) dom.zButtonsEl = domRefs.zButtonsEl;
        if (domRefs.saveCtrlEl) dom.saveCtrlEl = domRefs.saveCtrlEl;
        if (domRefs.dirtyCountEl) dom.dirtyCountEl = domRefs.dirtyCountEl;
        if (domRefs.changelogEl) dom.changelogEl = domRefs.changelogEl;
        if (domRefs.clEntriesEl) dom.clEntriesEl = domRefs.clEntriesEl;
        if (domRefs.statsEl) dom.statsEl = domRefs.statsEl;
        if (domRefs.zoneSelect) dom.zoneSelect = domRefs.zoneSelect;
        if (domRefs.loadingEl) dom.loadingEl = domRefs.loadingEl;
        if (domRefs.infoEmptyEl) dom.infoEmptyEl = domRefs.infoEmptyEl;
        if (domRefs.infoContentEl) dom.infoContentEl = domRefs.infoContentEl;
        if (domRefs.tooltip) dom.tooltip = domRefs.tooltip;
    }

    // =====================================================================
    //  Tab switching
    // =====================================================================

    function switchTab(tab) {
        MapperState.camera.activeTab = tab;
        document.querySelectorAll('.mapper-tabs .tab-btn').forEach(function(b) {
            b.classList.toggle('active', b.dataset.tab === tab);
        });
        MapperRender.resizeCanvas();
        updateZButtons();
        MapperRender.render();
    }

    // =====================================================================
    //  Zoom / spacing controls
    // =====================================================================

    function zoomIn() {
        MapperState.camera.zoomScale = Math.min(ZOOM_MAX, MapperState.camera.zoomScale * ZOOM_STEP);
        MapperRender.render();
    }

    function zoomOut() {
        MapperState.camera.zoomScale = Math.max(ZOOM_MIN, MapperState.camera.zoomScale / ZOOM_STEP);
        MapperRender.render();
    }

    // =====================================================================
    //  Camera centering and easing
    // =====================================================================

    function centerCamera() {
        MapperState.camera.panOffsetX = 0;
        MapperState.camera.panOffsetY = 0;
        if (MapperState.data.currentZone) {
            MapperState.centerOnZone(MapperState.data.currentZone);
        } else if (MapperState.data.zLevels.length > 0) {
            MapperState.camera.activeZ2d = MapperState.data.zLevels[0];
            updateZButtons();
            MapperRender.render();
        }
    }

    /** Smoothly animate the camera to (tx, ty, tz) using requestAnimationFrame. */
    function setCameraTarget(tx, ty, tz) {
        var cam = MapperState.camera;
        cam.panOffsetX = 0;
        cam.panOffsetY = 0;
        if (typeof tz === 'undefined') tz = cam.cameraZ;

        // Skip animation when easing is disabled
        if (CENTER_EASE_DURATION <= 0) {
            cam.cameraX = tx; cam.cameraY = ty; cam.cameraZ = tz;
            MapperRender.render();
            return;
        }

        // Cancel any in-progress ease before starting a new one
        if (cam.easeRafId !== null) { cancelAnimationFrame(cam.easeRafId); cam.easeRafId = null; }

        cam.easeStartX = cam.cameraX; cam.easeStartY = cam.cameraY; cam.easeStartZ = cam.cameraZ;
        cam.easeTargetX = tx; cam.easeTargetY = ty; cam.easeTargetZ = tz;
        cam.easeStartTime = null;

        function step(ts) {
            if (cam.easeStartTime === null) cam.easeStartTime = ts;
            var t = Math.min((ts - cam.easeStartTime) / 1000 / CENTER_EASE_DURATION, 1);
            var s = smoothstep(t);
            cam.cameraX = cam.easeStartX + (cam.easeTargetX - cam.easeStartX) * s;
            cam.cameraY = cam.easeStartY + (cam.easeTargetY - cam.easeStartY) * s;
            cam.cameraZ = cam.easeStartZ + (cam.easeTargetZ - cam.easeStartZ) * s;
            MapperRender.render();
            cam.easeRafId = t < 1 ? requestAnimationFrame(step) : null;
        }
        cam.easeRafId = requestAnimationFrame(step);
    }

    // =====================================================================
    //  Z-level buttons
    // =====================================================================

    function updateZButtons() {
        if (!dom.zButtonsEl) return;
        dom.zButtonsEl.innerHTML = '';
        if (MapperState.data.zLevels.length <= 1) return;

        var cam = MapperState.camera;
        var current = cam.activeZ2d;

        // Reverse so the highest Z is at the top, matching spatial intuition
        MapperState.data.zLevels.slice().reverse().forEach(function(z) {
            var btn = document.createElement('button');
            btn.textContent = z;
            btn.title = 'Z level ' + z;
            btn.style.cssText = 'width:24px;height:20px;padding:0;font-size:0.7rem;font-family:monospace;' +
                'border:1px solid var(--color-border);border-radius:3px;cursor:pointer;';
            if (z === current) {
                btn.style.background = 'var(--color-primary)';
                btn.style.color = '#fff';
                btn.style.borderColor = 'var(--color-primary)';
            } else {
                btn.style.background = 'var(--color-surface-white)';
                btn.style.color = 'var(--color-text-dim)';
            }
            btn.addEventListener('click', function() {
                cam.activeZ2d = z;
                updateZButtons();
                MapperRender.render();
            });
            dom.zButtonsEl.appendChild(btn);
        });
    }

    // =====================================================================
    //  Info panel (single room detail or multi-select summary)
    // =====================================================================

    function updateInfoPanel() {
        if (!dom.infoEmptyEl || !dom.infoContentEl) return;

        if (MapperState.selected.size === 0) {
            dom.infoEmptyEl.style.display = '';
            dom.infoContentEl.style.display = 'none';
            return;
        }

        dom.infoEmptyEl.style.display = 'none';
        dom.infoContentEl.style.display = '';

        var html = '';

        // --- Multi-select summary ---
        if (MapperState.selected.size > 1) {
            var ids = Array.from(MapperState.selected).sort(function(a, b) { return a - b; });
            html = '<div class="info-row"><span class="info-label">Selected</span><span class="info-value">' + ids.length + ' rooms</span></div>';
            html += '<hr class="info-divider">';
            ids.forEach(function(rid) {
                var r = MapperState.data.rooms.get(rid);
                var title = r ? escapeHtml(r.Title) : '?';
                html += '<div class="info-exit-row"><span class="info-exit-dir">#' + rid + '</span><span class="info-exit-id">' + title + '</span></div>';
            });
            dom.infoContentEl.innerHTML = html;
            return;
        }

        // --- Single room detail ---
        var selectedRoomId = Array.from(MapperState.selected)[0];
        var room = MapperState.data.rooms.get(selectedRoomId);
        if (!room) {
            dom.infoEmptyEl.style.display = '';
            dom.infoContentEl.style.display = 'none';
            return;
        }

        html = '';
        html += '<div class="info-row"><span class="info-label">ID</span><span class="info-value">' + selectedRoomId + '</span></div>';
        html += '<div class="info-row"><span class="info-label">Title</span><span class="info-value">' + escapeHtml(room.Title) + '</span></div>';
        html += '<div class="info-row"><span class="info-label">Biome</span><span class="info-value">' + escapeHtml(room.Biome || '-') + '</span></div>';
        html += '<div class="info-row"><span class="info-label">Symbol</span><span class="info-value">' + escapeHtml(room._symbol || '-') + '</span></div>';
        if (room.MapLegend) {
            html += '<div class="info-row"><span class="info-label">Legend</span><span class="info-value">' + escapeHtml(room.MapLegend) + '</span></div>';
        }
        if (room.Tags && room.Tags.length > 0) {
            var tagBadges = room.Tags.map(function(t) {
                var mod = MapperState.tagDescriptions[t];
                var tipText = mod ? 'module: ' + mod : t;
                return '<span class="info-badge" style="cursor:default" title="' + escapeHtml(tipText) + '">' + escapeHtml(t) + '</span>';
            }).join(' ');
            html += '<div class="info-row"><span class="info-label">Tags</span><span class="info-value info-badges">' + tagBadges + '</span></div>';
        }
        if (room.Nouns && Object.keys(room.Nouns).length > 0) {
            var nounBadges = Object.keys(room.Nouns).sort().map(function(n) {
                var desc = room.Nouns[n];
                var tip = desc ? 'title="' + escapeHtml(desc) + '"' : '';
                return '<span class="info-badge" style="cursor:default" ' + tip + '>' + escapeHtml(n) + '</span>';
            }).join(' ');
            html += '<div class="info-row"><span class="info-label">Nouns</span><span class="info-value info-badges">' + nounBadges + '</span></div>';
        }
        html += '<div class="info-row"><span class="info-label">Coords</span><span class="info-value">' + room.MapX + ', ' + room.MapY + ', ' + room.MapZ + '</span></div>';
        html += '<div class="info-row"><span class="info-label">Has Coords</span><span class="info-value">' + (room.HasCoordinates ? 'yes' : 'no') + '</span></div>';

        if (room.Exits && Object.keys(room.Exits).length > 0) {
            html += '<hr class="info-divider">';
            html += '<div class="info-exits">';
            var dirs = Object.keys(room.Exits).sort();
            dirs.forEach(function(dir) {
                var ex = room.Exits[dir];
                var badges = '';
                if (ex.Secret) badges += '<span class="info-badge secret">secret</span> ';
                if (ex.HasLock) badges += '<span class="info-badge locked">locked</span> ';
                html += '<div class="info-exit-row">';
                html += '<span class="info-exit-dir">' + escapeHtml(dir) + '</span>';
                html += '<span class="info-exit-id">' + badges + '#' + ex.RoomId + '</span>';
                html += '</div>';
            });
            html += '</div>';
        }

        html += '<hr class="info-divider">';
        html += '<a class="info-link" href="/admin/rooms#' + selectedRoomId + '">Edit Room &rarr;</a>';

        dom.infoContentEl.innerHTML = html;
    }

    // =====================================================================
    //  Tooltip
    // =====================================================================

    function showTooltip(mx, my, room, id) {
        if (!dom.tooltip) return;
        var html = '<div class="tt-title">' + escapeHtml(room.Title || 'Unknown') + '</div>';
        html += '<div class="tt-id">Room #' + id + '</div>';
        html += '<hr class="tt-divider">';
        if (room.Biome) html += '<div class="tt-row"><span class="tt-label">Biome</span><span class="tt-value">' + escapeHtml(room.Biome) + '</span></div>';
        if (room._symbol) html += '<div class="tt-row"><span class="tt-label">Symbol</span><span class="tt-value">' + escapeHtml(room._symbol) + '</span></div>';
        html += '<div class="tt-row"><span class="tt-label">Coords</span><span class="tt-value">' + room.MapX + ', ' + room.MapY + ', ' + room.MapZ + '</span></div>';
        if (room.Exits) {
            var dirs = Object.keys(room.Exits).sort();
            if (dirs.length > 0) {
                html += '<hr class="tt-divider">';
                html += '<div class="tt-row"><span class="tt-label">Exits</span><span class="tt-value">' + dirs.join(', ') + '</span></div>';
            }
        }
        dom.tooltip.innerHTML = html;
        dom.tooltip.style.display = 'block';
        positionTooltip(mx, my);
    }

    /** Keep the tooltip on-screen by flipping sides when it would overflow. */
    function positionTooltip(mx, my) {
        if (!dom.tooltip) return;
        var ttW = dom.tooltip.offsetWidth, ttH = dom.tooltip.offsetHeight;
        var vw = window.innerWidth, vh = window.innerHeight;
        var left = mx + 14;
        if (left + ttW > vw - 8) left = mx - ttW - 14;
        left = Math.max(8, left);
        var top = my - Math.floor(ttH / 2);
        if (top + ttH > vh - 8) top = vh - ttH - 8;
        top = Math.max(8, top);
        dom.tooltip.style.left = left + 'px';
        dom.tooltip.style.top  = top + 'px';
    }

    /** Brief delay before hiding prevents flicker when moving between rooms. */
    function hideTooltip() {
        if (!dom.tooltip) return;
        var cam = MapperState.camera;
        cam.tooltipHideTimer = setTimeout(function() {
            dom.tooltip.style.display = 'none';
        }, 80);
    }

    // =====================================================================
    //  Loading indicator
    // =====================================================================

    function showLoading(show) {
        if (!dom.loadingEl) return;
        dom.loadingEl.classList.toggle('hidden', !show);
    }

    // =====================================================================
    //  Stats display
    // =====================================================================

    function updateStats() {
        if (!dom.statsEl) return;
        var data = MapperState.data;
        var total = data.rooms.size;
        var coordinated = 0;
        data.rooms.forEach(function(r) { if (r.HasCoordinates) coordinated++; });
        var unmapped = total - coordinated;
        var txt = total + ' rooms';
        if (unmapped > 0) txt += ' (' + unmapped + ' unmapped)';
        txt += ' | ' + data.zLevels.length + ' z-level' + (data.zLevels.length !== 1 ? 's' : '');
        txt += ' | ' + data.visibleZones.size + '/' + data.allZones.length + ' zones';
        dom.statsEl.textContent = txt;
    }

    // =====================================================================
    //  Zone dropdown
    // =====================================================================

    function populateZoneDropdown() {
        if (!dom.zoneSelect) return;
        var current = dom.zoneSelect.value;
        while (dom.zoneSelect.options.length > 1) dom.zoneSelect.remove(1);
        MapperState.data.allZones.forEach(function(z) {
            var opt = document.createElement('option');
            opt.value = z.Name;
            opt.textContent = z.Name + ' (' + z.RoomCount + ' rooms)';
            dom.zoneSelect.appendChild(opt);
        });
        if (current) dom.zoneSelect.value = current;
    }

    // =====================================================================
    //  Zone picker modal
    // =====================================================================

    var zonePickerCallback = null;

    function showZonePicker(zones, callback) {
        var backdrop = document.getElementById('zone-picker-backdrop');
        var list = document.getElementById('zone-picker-list');
        var cancelBtn = document.getElementById('zone-picker-cancel');
        if (!backdrop || !list) { callback(zones[0]); return; }

        list.innerHTML = '';
        zones.forEach(function(zone) {
            var btn = document.createElement('button');
            btn.textContent = zone;
            btn.addEventListener('click', function() {
                hideZonePicker();
                callback(zone);
            });
            list.appendChild(btn);
        });

        zonePickerCallback = null;
        cancelBtn.onclick = hideZonePicker;
        backdrop.classList.add('visible');
    }

    function hideZonePicker() {
        var backdrop = document.getElementById('zone-picker-backdrop');
        if (backdrop) backdrop.classList.remove('visible');
        zonePickerCallback = null;
    }

    /**
     * Resolves which zone a new room at (gx, gy, gz) should belong to.
     * hintZone is used as a fallback (e.g. the source room's zone in quick-build).
     * Calls callback(zoneName) asynchronously (immediately or after user picks).
     */
    function resolveZoneForRoom(gx, gy, gz, hintZone, callback) {
        var candidates = getZonesAtPoint(gx, gy, gz);
        if (candidates.length === 0) {
            // Outside all bounding boxes — use hint or closest room
            callback(hintZone || closestZone(gx, gy, gz, MapperState.data.currentZone || ''));
        } else if (candidates.length === 1) {
            callback(candidates[0]);
        } else {
            showZonePicker(candidates, callback);
        }
    }

    // =====================================================================
    //  Zone creation
    // =====================================================================

    function createZoneAt(gx, gy, gz) {
        var backdrop  = document.getElementById('zone-name-backdrop');
        var input     = document.getElementById('zone-name-input');
        var errorEl   = document.getElementById('zone-name-error');
        var confirmBtn = document.getElementById('zone-name-confirm');
        var cancelBtn  = document.getElementById('zone-name-cancel');
        if (!backdrop || !input) return;

        input.value = '';
        errorEl.textContent = '';
        backdrop.classList.add('visible');
        setTimeout(function() { input.focus(); }, 50);

        function close() {
            backdrop.classList.remove('visible');
            confirmBtn.onclick = null;
            cancelBtn.onclick = null;
            input.onkeydown = null;
        }

        function submit() {
            var name = input.value.trim();
            if (!name) { errorEl.textContent = 'Zone name is required.'; return; }
            if (name.length < 2) { errorEl.textContent = 'Zone name must be at least 2 characters.'; return; }
            // Case-insensitive duplicate check against existing and pending zones
            var nameLower = name.toLowerCase();
            var exists = MapperState.data.allZones.some(function(z) { return z.Name.toLowerCase() === nameLower; }) ||
                         Array.from(MapperState.dirty.pendingZones).some(function(z) { return z.toLowerCase() === nameLower; });
            if (exists) { errorEl.textContent = 'A zone with that name already exists.'; return; }
            close();

            // Register the new zone locally so bounding boxes and zone resolution work
            MapperState.dirty.pendingZones.add(name);
            MapperState.data.allZones.push({ Name: name, RoomCount: 0, RoomId: 0, DefaultBiome: '' });
            populateZoneDropdown();

            // Create a local room assigned to the new zone at the chosen position
            var tempId = MapperState.createRoomLocally(gx, gy, gz, name);
            MapperState.selectRoom(tempId);

            // Switch the zone dropdown to the new zone
            if (dom.zoneSelect) {
                dom.zoneSelect.value = name;
                dom.zoneSelect.dispatchEvent(new Event('change'));
            }
            MapperRender.render();
        }

        confirmBtn.onclick = submit;
        cancelBtn.onclick = close;
        input.onkeydown = function(e) {
            if (e.key === 'Enter') submit();
            if (e.key === 'Escape') close();
        };
    }

    // =====================================================================
    //  Local room creation (deferred until save)
    // =====================================================================

    function createRoomAt(gx, gy, gz) {
        if (!MapperState.data.currentZone && MapperState.data.rooms.size === 0) {
            alert('Select a zone from the dropdown first to set which zone new rooms are created in.');
            return;
        }
        resolveZoneForRoom(gx, gy, gz, MapperState.data.currentZone, function(zone) {
            var tempId = MapperState.createRoomLocally(gx, gy, gz, zone);
            MapperState.selectRoom(tempId);
            MapperRender.render();
        });
    }

    // =====================================================================
    //  Server coordinate conversion
    // =====================================================================

    /**
     * Translate world-space coordinates back to zone-relative server
     * coordinates by subtracting the zone offset applied at load time.
     */
    function toServerCoords(room) {
        if (!room) return { x: 0, y: 0, z: 0 };
        var data = MapperState.data;
        var zone = room.Zone || data.currentZone;
        var off = zone ? data.zoneOffsets.get(zone) : null;
        if (!off) return { x: room.MapX, y: room.MapY, z: room.MapZ };
        return { x: room.MapX - off.dx, y: room.MapY - off.dy, z: room.MapZ };
    }

    // =====================================================================
    //  Save / Discard
    // =====================================================================

    async function saveAllChanges() {
        var data = MapperState.data;
        var dirty = MapperState.dirty;
        showLoading(true);

        // 0. Create any pending new zones on the server first.
        //    Capture the auto-created root room ID per zone so step 2 can
        //    patch it to the correct position instead of creating a second room.
        var zoneRootIds = {};  // zoneName -> server roomId
        for (var zoneName of dirty.pendingZones) {
            var zRes = await AdminAPI.post('/admin/api/v1/zones', { Name: zoneName });
            if (zRes.ok && zRes.data && zRes.data.data) {
                zoneRootIds[zoneName] = zRes.data.data.RoomId;
            }
        }

        // 1. Delete rooms on server
        for (var i = 0; i < dirty.deletedRooms.length; i++) {
            await AdminAPI.delete('/admin/api/v1/rooms/' + dirty.deletedRooms[i]);
        }

        // 2. Create new rooms on server and build a temp-to-real ID map.
        //    For rooms belonging to a pending zone, reuse the zone's auto-created
        //    root room rather than creating a second room.
        var tempToReal = new Map();
        for (var entry of dirty.createdRooms) {
            var tempId = entry[0], info = entry[1];
            var zone = info.zone || data.currentZone || '';
            var off = data.zoneOffsets.get(zone) || { dx: 0, dy: 0 };
            var serverX = info.gx - off.dx, serverY = info.gy - off.dy;
            var tempRoom = data.rooms.get(tempId);
            var realId = null;
            if (zoneRootIds[zone] !== undefined) {
                // Reuse the root room the zone API already created
                realId = zoneRootIds[zone];
                delete zoneRootIds[zone];  // only consume it once
            } else {
                var createRes = await AdminAPI.post('/admin/api/v1/rooms', { Zone: zone });
                if (createRes.ok && createRes.data && createRes.data.data) {
                    realId = createRes.data.data.RoomId;
                }
            }
            if (realId !== null) {
                tempToReal.set(tempId, realId);
                var patchData = {
                    MapX: serverX, MapY: serverY, MapZ: info.gz, HasCoordinates: true
                };
                if (tempRoom) {
                    if (tempRoom.Title && tempRoom.Title !== 'New Room') patchData.Title = tempRoom.Title;
                    if (tempRoom.Biome) patchData.Biome = tempRoom.Biome;
                    if (tempRoom.MapSymbol) patchData.MapSymbol = tempRoom.MapSymbol;
                    if (tempRoom.MapLegend) patchData.MapLegend = tempRoom.MapLegend;
                }
                await AdminAPI.patch('/admin/api/v1/rooms/' + realId, patchData);
            }
        }

        // 3. Move existing rooms on server
        for (var entry2 of dirty.movedRooms) {
            var roomId = entry2[0];
            if (roomId < 0) continue;
            var room = data.rooms.get(roomId);
            if (room) {
                var sc = toServerCoords(room);
                await AdminAPI.patch('/admin/api/v1/rooms/' + roomId, {
                    MapX: sc.x, MapY: sc.y, MapZ: sc.z, HasCoordinates: true
                });
            }
        }

        // 4. Apply exit changes (grouped by room: removals then additions)
        //    We merge against the server's current exit map so we don't
        //    clobber exits that were not touched locally.
        var exitChangesByRoom = new Map();
        dirty.exitRemovals.forEach(function(er) {
            var rId = er.roomId < 0 ? (tempToReal.get(er.roomId) || er.roomId) : er.roomId;
            if (rId < 0) return;
            if (!exitChangesByRoom.has(rId)) exitChangesByRoom.set(rId, { remove: new Set(), add: [] });
            exitChangesByRoom.get(rId).remove.add(er.dir);
        });
        dirty.exitAdditions.forEach(function(ea) {
            var rId = ea.roomId < 0 ? (tempToReal.get(ea.roomId) || ea.roomId) : ea.roomId;
            var tId = ea.targetRoomId < 0 ? (tempToReal.get(ea.targetRoomId) || ea.targetRoomId) : ea.targetRoomId;
            if (rId < 0 || tId < 0) return;
            if (!exitChangesByRoom.has(rId)) exitChangesByRoom.set(rId, { remove: new Set(), add: [] });
            exitChangesByRoom.get(rId).add.push({ dir: ea.dir, targetRoomId: tId });
        });
        for (var entry3 of exitChangesByRoom) {
            var rId = entry3[0], changes = entry3[1];
            var res = await AdminAPI.get('/admin/api/v1/rooms/' + rId);
            if (res.ok && res.data && res.data.data) {
                var fullRoom = res.data.data;
                var mergedExits = {};
                for (var d in (fullRoom.Exits || {})) {
                    if (!changes.remove.has(d)) mergedExits[d] = fullRoom.Exits[d];
                }
                changes.add.forEach(function(a) { mergedExits[a.dir] = { RoomId: a.targetRoomId }; });
                await AdminAPI.patch('/admin/api/v1/rooms/' + rId, { Exits: mergedExits });
            }
        }

        MapperState.clearDirty();
        await MapperState.loadAllRooms();
        showLoading(false);
        MapperState.selected.clear();
        updateInfoPanel();
        MapperRender.render();
    }

    async function discardChanges() {
        MapperState.clearDirty();
        showLoading(true);
        await MapperState.loadAllRooms();
        showLoading(false);
        MapperState.selected.clear();
        updateInfoPanel();
        MapperRender.render();
    }

    // =====================================================================
    //  Public API
    // =====================================================================

    return {
        init: init,
        switchTab: switchTab, zoomIn: zoomIn, zoomOut: zoomOut,
        centerCamera: centerCamera, setCameraTarget: setCameraTarget,
        updateZButtons: updateZButtons, updateInfoPanel: updateInfoPanel,
        showTooltip: showTooltip, positionTooltip: positionTooltip, hideTooltip: hideTooltip,
        showLoading: showLoading, updateStats: updateStats,
        populateZoneDropdown: populateZoneDropdown,
        createRoomAt: createRoomAt,
        createZoneAt: createZoneAt,
        resolveZoneForRoom: resolveZoneForRoom,
        hideZonePicker: hideZonePicker,
        toServerCoords: toServerCoords,
        saveAllChanges: saveAllChanges, discardChanges: discardChanges
    };

})();
