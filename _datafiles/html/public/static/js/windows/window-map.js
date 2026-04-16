/**
 * window-map.js
 *
 * Virtual window: Map (room grid SVG).
 *
 * Responds to GMCP namespace:
 *   Room  - room info and exit data
 *
 * Reads: Client.GMCPStructs.Room.Info
 */

'use strict';

(function() {

    injectStyles(`
        #map-room-tooltip {
            position: fixed;
            z-index: 99999;
            pointer-events: none;
            background: #0d2e28;
            border: 1px solid #1c6b60;
            border-radius: 6px;
            box-shadow: 0 4px 16px rgba(0,0,0,0.7);
            padding: 8px 10px;
            min-width: 140px;
            max-width: 240px;
            display: none;
        }

        .map-tt-name {
            font-size: 0.85em;
            font-weight: bold;
            color: #dffbd1;
            margin-bottom: 4px;
            line-height: 1.3;
        }

        .map-tt-divider {
            border: none;
            border-top: 1px solid #1c6b60;
            margin: 5px 0;
        }

        .map-tt-row {
            display: flex;
            justify-content: space-between;
            align-items: baseline;
            gap: 8px;
            font-size: 0.75em;
            line-height: 1.6;
        }

        .map-tt-row-label {
            color: #7ab8a0;
            text-transform: uppercase;
            letter-spacing: 0.04em;
            font-size: 0.88em;
            flex-shrink: 0;
        }

        .map-tt-row-value {
            color: #dffbd1;
            text-align: right;
        }

        .map-tt-badges {
            display: flex;
            flex-wrap: wrap;
            gap: 3px;
            margin-top: 4px;
        }

        .map-tt-badge {
            font-size: 0.62em;
            padding: 1px 4px;
            border-radius: 3px;
            background: #1a2e28;
            color: #7ab8a0;
            border: 1px solid #1c6b60;
        }

        .map-tt-badge.pvp    { background: #3d0f0f; color: #e06060; border-color: #6b1c1c; }
        .map-tt-badge.bank   { background: #0f2e10; color: #56d44a; border-color: #1c6b1c; }
        .map-tt-badge.trainer { background: #2e2000; color: #fdd; border-color: #6b5010; }
        .map-tt-badge.storage { background: #1a1200; color: #c8a800; border-color: #6b5010; }
    `);

    // -----------------------------------------------------------------------
    // Tooltip
    // -----------------------------------------------------------------------
    let mapTooltip   = null;
    let mapHideTimer = null;

    function ensureMapTooltip() {
        if (mapTooltip) { return; }
        mapTooltip = document.createElement('div');
        mapTooltip.id = 'map-room-tooltip';
        document.body.appendChild(mapTooltip);
    }

    function showMapTooltip(svgEl, roomData) {
        ensureMapTooltip();
        clearTimeout(mapHideTimer);

        let html = '<div class="map-tt-name">' + (roomData.name || 'Unknown') + '</div>';

        const rows = [];
        if (roomData.environment) { rows.push({ label: 'Env',  value: roomData.environment }); }
        if (roomData.maplegend)   { rows.push({ label: 'Type', value: roomData.maplegend   }); }
        if (roomData.mapsymbol)   { rows.push({ label: 'Symbol', value: roomData.mapsymbol }); }

        if (rows.length > 0) {
            html += '<hr class="map-tt-divider">';
            rows.forEach(r => {
                html += '<div class="map-tt-row">' +
                    '<span class="map-tt-row-label">' + r.label + '</span>' +
                    '<span class="map-tt-row-value">' + r.value + '</span>' +
                '</div>';
            });
        }

        const details = roomData.details || [];
        const badgeOrder = ['pvp', 'bank', 'trainer', 'storage', 'character', 'ephemeral'];
        const badges = badgeOrder.filter(d => details.includes(d));
        if (badges.length > 0) {
            html += '<hr class="map-tt-divider"><div class="map-tt-badges">';
            badges.forEach(b => {
                html += '<span class="map-tt-badge ' + b + '">' + b + '</span>';
            });
            html += '</div>';
        }

        mapTooltip.innerHTML = html;
        mapTooltip.style.display = 'block';
        _positionMapTooltip(svgEl);
    }

    function _positionMapTooltip(svgEl) {
        if (!mapTooltip) { return; }
        // Use the bounding rect of the SVG element that triggered the event
        const rect = svgEl.getBoundingClientRect();
        const ttW  = mapTooltip.offsetWidth;
        const ttH  = mapTooltip.offsetHeight;
        const vw   = window.innerWidth;
        const vh   = window.innerHeight;

        let left = rect.right + 8;
        if (left + ttW > vw - 8) { left = rect.left - ttW - 8; }
        left = Math.max(8, left);

        let top = rect.top;
        if (top + ttH > vh - 8) { top = vh - ttH - 8; }
        top = Math.max(8, top);

        mapTooltip.style.left = left + 'px';
        mapTooltip.style.top  = top  + 'px';
    }

    function hideMapTooltip() {
        if (!mapTooltip) { return; }
        mapHideTimer = setTimeout(() => { mapTooltip.style.display = 'none'; }, 80);
    }

    function attachMapTooltip(gEl, roomId) {
        gEl.addEventListener('mouseenter', () => {
            const data = roomInfoStore.get(roomId);
            if (data) { showMapTooltip(gEl, data); }
        });
        gEl.addEventListener('mouseleave', hideMapTooltip);
        gEl.addEventListener('mousemove',  () => {
            if (mapTooltip && mapTooltip.style.display === 'block') {
                _positionMapTooltip(gEl);
            }
        });
    }

    // -----------------------------------------------------------------------
    // Module state
    // -----------------------------------------------------------------------
    let gr = null;  // RoomGridSVG instance, created when window first opens

    const allRooms = {
        currentZoneKey: '',
        roomZones: {},
    };

    // Map<RoomId, roomInfo> — stores the latest GMCP info for each known room
    const roomInfoStore = new Map();

    // -----------------------------------------------------------------------
    // DOM factory
    // -----------------------------------------------------------------------
    function createDOM() {
        gr = null;
        allRooms.currentZoneKey = '';
        allRooms.roomZones = {};
        const el = document.createElement('div');
        el.id = 'map-render';
        el.style.width  = '100%';
        el.style.height = '100%';
        document.body.appendChild(el);
        return el;
    }

    // -----------------------------------------------------------------------
    // VirtualWindow instance
    // -----------------------------------------------------------------------
    const win = new VirtualWindow('Map', {
        dock:          'right',
        defaultDocked: true,
        dockedHeight:  363,
        factory() {
            const el = createDOM();
            return {
                title:      'Map',
                mount:      el,
                background: '#1c6b60',
                border:     1,
                x:          'right',
                y:          66,
                width:      363,
                height:     20 + 363,
                header:     20,
                bottom:     60,
            };
        },
    });

    // -----------------------------------------------------------------------
    // Update logic
    // -----------------------------------------------------------------------
    function ensureGrid() {
        if (!gr) {
            gr = new RoomGridSVG('#map-render', {
                cellSize:    80,
                cellMargin:  80,
                initialZoom: 0.5,
            });
        }
    }

    function updateMap() {
        const obj = Client.GMCPStructs.Room;
        if (!obj || !obj.Info) {
            return;
        }

        win.open();
        if (!win.isOpen()) {
            return;
        }

        ensureGrid();

        const info   = obj.Info;
        const winBox = win.get();
        if (winBox) {
            winBox.setTitle('Map (' + info.area + ')');
        }

        // Parse coordinate string: "zoneName, x, y, z"
        const coords = info.coords.split(',').map(s => s.trim());
        const zoneName = coords[0];
        const zoneKey  = zoneName + '/z:' + coords[3];

        if (allRooms.currentZoneKey !== zoneKey) {
            gr.reset();
            allRooms.currentZoneKey = zoneKey;
            if (!allRooms.roomZones[zoneKey] || !Array.isArray(allRooms.roomZones[zoneKey])) {
                allRooms.roomZones[zoneKey] = [];
            }
            gr.setRooms(allRooms.roomZones[zoneKey]);
        }

        const r = {
            RoomId: info.num,
            Text:   '',
            x:      parseInt(coords[1]),
            y:      parseInt(coords[2]),
            z:      parseInt(coords[3]),
            Exits:  [],
        };

        // Annotate room based on NPC adjectives
        if (info.Contents && info.Contents.Npcs) {
            for (const i in info.Contents.Npcs) {
                if (info.Contents.Npcs[i].adjectives.indexOf('shop') !== -1) {
                    r.Color = '#AFE1AF';
                    r.Text  = '💰';
                    break;
                }
            }
        }

        // Annotate room based on room details flags
        if (info.details) {
            if (info.details.indexOf('bank') !== -1) {
                r.Color = '#56d44a';
                r.Text  = '🏛️';
            } else if (info.details.indexOf('trainer') !== -1) {
                r.Color = '#fdd';
                r.Text  = '🧠';
            } else if (info.details.indexOf('storage') !== -1) {
                r.Color = '#a88100';
                r.Text  = '🗄️';
            } else if (info.details.indexOf('character') !== -1) {
                r.Color = '#757575';
                r.Text  = '👤';
            }
        }

        // Build exits (2D only — skip exits with a z-delta)
        for (const p in info.exitsv2) {
            const exitInfo = info.exitsv2[p];
            if (exitInfo.dz === 0) {
                r.Exits.push({
                    RoomId: exitInfo.num,
                    x:      exitInfo.dx + parseInt(coords[1]),
                    y:      exitInfo.dy + parseInt(coords[2]),
                    Text:   '',
                });
            }
        }

        // Store the latest info for this room for tooltip use
        roomInfoStore.set(info.num, info);

        allRooms.roomZones[zoneKey].push(r);
        gr.addRoom(r);

        // Attach tooltip listeners. We attach to the current room and to any
        // exit rooms that were just pre-created by addRoom (they have g elements
        // but no info yet — tooltips will show once their info arrives).
        const attachedIds = new Set();
        const tryAttach = (roomId) => {
            if (attachedIds.has(roomId)) { return; }
            attachedIds.add(roomId);
            const gEl = gr.svg.querySelector(`g[data-room-id="${roomId}"]`);
            if (gEl && !gEl.dataset.tooltipAttached) {
                gEl.dataset.tooltipAttached = '1';
                attachMapTooltip(gEl, roomId);
            }
        };

        tryAttach(info.num);
        r.Exits.forEach(e => tryAttach(e.RoomId));

        gr.centerOnRoom(r.RoomId);
    }

    // -----------------------------------------------------------------------
    // Registration
    // -----------------------------------------------------------------------
    VirtualWindows.register({
        window:       win,
        gmcpHandlers: ['Room'],
        onGMCP(namespace, body) {
            updateMap();
        },
    });

})();
