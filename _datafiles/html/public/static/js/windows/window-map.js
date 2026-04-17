/* global Client, VirtualWindow, VirtualWindows, injectStyles */

/**
 * window-map.js
 *
 * Virtual window: map (symbol-based room grid canvas).
 *
 * Rooms are drawn as fixed-size squares with their MapSymbol centered inside.
 * Background color per room is determined by a symbol-to-color lookup table.
 * Connections between rooms are drawn as brown lines behind the room squares.
 * The map stays centered on the player's current room at all times.
 * - / + overlay buttons zoom the map in and out.
 *
 * Responds to GMCP namespaces:
 *   Room      - incremental update as the player moves room-to-room
 *   World.Map - bulk snapshot of all visited rooms, requested once on connect
 *
 * Reads:
 *   Client.GMCPStructs.Room.Info
 *   Client.GMCPStructs.World.Map
 */

'use strict';

(function () {

    // -------------------------------------------------------------------------
    // Layout constants — edit these to adjust the visual appearance
    // -------------------------------------------------------------------------

    /** Width and height of each room square in pixels (at zoom level 1). */
    var ROOM_SIZE = 28;

    /**
     * Gap between room squares in pixels (at zoom level 1).
     * Total grid step = ROOM_SIZE + ROOM_GAP.
     */
    var ROOM_GAP = 14;

    /** Multiplicative step applied on each zoom in/out button press. */
    var ZOOM_STEP = 1.25;

    /** Minimum and maximum allowed zoom scale. */
    var ZOOM_MIN = 0.25;
    var ZOOM_MAX = 4.0;

    /** Stroke width of the brown connection lines between rooms. */
    var CONNECTION_WIDTH = 4;

    /** Stroke width of the black border drawn around each room square. */
    var ROOM_BORDER_WIDTH = 1.5;

    /** Font size for the symbol drawn inside each room square (in pixels). */
    var SYMBOL_FONT_SIZE = 14;

    /** Background color of the map canvas. */
    var MAP_BACKGROUND = '#2b2b2b';

    /** Border color drawn around each room square. */
    var ROOM_BORDER_COLOR = '#000000';

    /** Color of the connection lines between rooms. */
    var CONNECTION_COLOR = '#7a4a1a';

    /** Fill color used for the player's current room. */
    var CURRENT_ROOM_COLOR = '#c20000';

    /** Text color used for the symbol inside the player's current room. */
    var CURRENT_ROOM_TEXT_COLOR = '#ffffff';

    /**
     * Symbol-to-color lookup table.
     *
     * Maps a room's MapSymbol string to a background fill color for the room
     * square.  Add, remove, or change entries here to adjust the color scheme.
     * Symbols not listed fall back to DEFAULT_ROOM_COLOR.
     */
    var SYMBOL_COLORS = {
        // Biome defaults (matched to biome symbol values)
        '~':  '#2a53f7',   // shore / water edge
        '≈':  '#0033cd',   // open water
        '♣':  '#1a5c1a',   // forest
        '♨':  '#3d5c1a',   // swamp
        '❄':  '#a0c8e0',   // snow
        '⌬':  '#4a3a2a',   // cave
        '⩕':  '#6b5a3a',   // mountains
        '▼':  '#7a6a4a',   // cliffs
        '⌂':  '#7a5a2a',   // house
        '*':  '#c8a050',   // desert
        "'":  '#5a7a2a',   // farmland

        // Common room-specific symbols
        '$':  '#2a6a2a',   // shop
        '%':  '#2a5a7a',   // trainer
        '♜':  '#3a3a3a',   // wall
        '•':  '#3a3a4a',   // generic / default biome dot
    };

    /** Fallback color for symbols not found in SYMBOL_COLORS. */
    var DEFAULT_ROOM_COLOR = '#3a3a4a';

    /** Text color for symbol labels inside room squares. */
    var SYMBOL_TEXT_COLOR = '#e0e0e0';

    // -------------------------------------------------------------------------
    // Derived base values (do not edit)
    // -------------------------------------------------------------------------
    var BASE_STEP = ROOM_SIZE + ROOM_GAP;

    // -------------------------------------------------------------------------
    // Styles
    // -------------------------------------------------------------------------
    injectStyles([
        '#map-canvas-container {',
        '    width: 100%;',
        '    height: 100%;',
        '    background: ' + MAP_BACKGROUND + ';',
        '    overflow: hidden;',
        '    position: relative;',
        '}',
        '#map-canvas {',
        '    display: block;',
        '    position: absolute;',
        '    top: 0; left: 0;',
        '}',
        '#map-tooltip {',
        '    position: fixed;',
        '    z-index: 99999;',
        '    pointer-events: none;',
        '    background: #0d2e28;',
        '    border: 1px solid #1c6b60;',
        '    border-radius: 6px;',
        '    box-shadow: 0 4px 16px rgba(0,0,0,0.7);',
        '    padding: 8px 10px;',
        '    min-width: 140px;',
        '    max-width: 240px;',
        '    display: none;',
        '    font-family: monospace;',
        '}',
        '#map-tooltip .tt-name {',
        '    font-size: 0.85em;',
        '    font-weight: bold;',
        '    color: #dffbd1;',
        '    margin-bottom: 4px;',
        '    line-height: 1.3;',
        '}',
        '#map-tooltip .tt-divider {',
        '    border: none;',
        '    border-top: 1px solid #1c6b60;',
        '    margin: 5px 0;',
        '}',
        '#map-tooltip .tt-row {',
        '    display: flex;',
        '    justify-content: space-between;',
        '    align-items: baseline;',
        '    gap: 8px;',
        '    font-size: 0.75em;',
        '    line-height: 1.6;',
        '}',
        '#map-tooltip .tt-label {',
        '    color: #7ab8a0;',
        '    text-transform: uppercase;',
        '    letter-spacing: 0.04em;',
        '    font-size: 0.88em;',
        '    flex-shrink: 0;',
        '}',
        '#map-tooltip .tt-value {',
        '    color: #dffbd1;',
        '    text-align: right;',
        '}',
        '#map-tooltip .tt-badges {',
        '    display: flex;',
        '    flex-wrap: wrap;',
        '    gap: 3px;',
        '    margin-top: 4px;',
        '}',
        '#map-tooltip .tt-badge {',
        '    font-size: 0.62em;',
        '    padding: 1px 4px;',
        '    border-radius: 3px;',
        '    background: #1a2e28;',
        '    color: #7ab8a0;',
        '    border: 1px solid #1c6b60;',
        '}',
        '#map-tooltip .tt-badge.pvp     { background: #3d0f0f; color: #e06060; border-color: #6b1c1c; }',
        '#map-tooltip .tt-badge.bank    { background: #0f2e10; color: #56d44a; border-color: #1c6b1c; }',
        '#map-tooltip .tt-badge.trainer { background: #2e2000; color: #fdd;    border-color: #6b5010; }',
        '#map-tooltip .tt-badge.storage { background: #1a1200; color: #c8a800; border-color: #6b5010; }',
        '#map-zoom-controls {',
        '    position: absolute;',
        '    top: 6px;',
        '    right: 6px;',
        '    display: flex;',
        '    gap: 3px;',
        '    z-index: 10;',
        '}',
        '#map-zoom-controls button {',
        '    width: 22px;',
        '    height: 22px;',
        '    padding: 0;',
        '    font-size: 14px;',
        '    line-height: 1;',
        '    background: rgba(0,0,0,0.55);',
        '    color: #ccc;',
        '    border: 1px solid #555;',
        '    border-radius: 3px;',
        '    cursor: pointer;',
        '}',
        '#map-zoom-controls button:hover {',
        '    background: rgba(0,0,0,0.8);',
        '    color: #fff;',
        '}',
    ].join('\n'));

    // -------------------------------------------------------------------------
    // Module state
    // -------------------------------------------------------------------------

    /** Canvas element. */
    var canvas = null;
    /** 2D rendering context. */
    var ctx = null;
    /** Container div. */
    var container = null;

    /**
     * rooms: Map<RoomId, { x, y, symbol, isCurrent }>
     * Stores the grid position and symbol for every known room.
     */
    var rooms = new Map();

    /**
     * edges: Set of canonical "minId-maxId" strings to avoid duplicate lines.
     */
    var edges = new Set();

    /**
     * zoneExitStubs: Array of { roomId, dx, dy } for exits that leave the
     * current zone.  Drawn as short stub lines from the room outward.
     */
    var zoneExitStubs = [];

    /** RoomId of the player's current room. */
    var currentRoomId = null;

    /** Current zoom scale factor. */
    var zoomScale = 1.0;

    /** Map<RoomId, full GMCP info object> for tooltip data. */
    var roomInfoStore = new Map();

    /** Tooltip DOM element, created lazily. */
    var tooltip = null;
    /** setTimeout handle for hiding the tooltip. */
    var tooltipHideTimer = null;

    /** Per-zone room cache, keyed by "zoneName/z:N". */
    var allRooms = {
        currentZoneKey: '',
        roomZones: {},
    };

    /** True once World.Map has been requested this session, to avoid re-requesting. */
    var worldMapRequested = false;

    // -------------------------------------------------------------------------
    // Color lookup helper
    // -------------------------------------------------------------------------
    function colorForSymbol(sym) {
        if (!sym) { return DEFAULT_ROOM_COLOR; }
        return SYMBOL_COLORS[sym] || DEFAULT_ROOM_COLOR;
    }

    // -------------------------------------------------------------------------
    // Canvas helpers
    // -------------------------------------------------------------------------

    function resizeCanvas() {
        if (!canvas || !container) { return; }
        canvas.width  = container.clientWidth  || 1;
        canvas.height = container.clientHeight || 1;
    }

    /**
     * Convert a room's grid (x, y) to canvas pixel coordinates, centered on
     * the current room, respecting the current zoom scale.
     */
    function gridToCanvas(gx, gy) {
        var cx = currentRoomId !== null && rooms.has(currentRoomId) ?
            rooms.get(currentRoomId).x : 0;
        var cy = currentRoomId !== null && rooms.has(currentRoomId) ?
            rooms.get(currentRoomId).y : 0;

        var midX = Math.floor(canvas.width  / 2);
        var midY = Math.floor(canvas.height / 2);
        var step = BASE_STEP * zoomScale;

        return {
            px: midX + (gx - cx) * step,
            py: midY + (gy - cy) * step,
        };
    }

    // -------------------------------------------------------------------------
    // Rendering
    // -------------------------------------------------------------------------

    function render() {
        if (!ctx || !canvas) { return; }

        ctx.clearRect(0, 0, canvas.width, canvas.height);

        // Map background
        ctx.fillStyle = MAP_BACKGROUND;
        ctx.fillRect(0, 0, canvas.width, canvas.height);

        // Draw connection lines first (behind rooms)
        ctx.strokeStyle = CONNECTION_COLOR;
        ctx.lineWidth   = CONNECTION_WIDTH * zoomScale;
        ctx.lineCap     = 'round';

        edges.forEach(function (key) {
            var parts = key.split('-');
            var idA   = parseInt(parts[0], 10);
            var idB   = parseInt(parts[1], 10);
            var rA    = rooms.get(idA);
            var rB    = rooms.get(idB);
            if (!rA || !rB) { return; }

            var pA = gridToCanvas(rA.x, rA.y);
            var pB = gridToCanvas(rB.x, rB.y);

            ctx.beginPath();
            ctx.moveTo(pA.px, pA.py);
            ctx.lineTo(pB.px, pB.py);
            ctx.stroke();
        });

        // Draw zone-exit stubs: short lines from room center outward
        var stubLen = (BASE_STEP * zoomScale) * 0.55;
        zoneExitStubs.forEach(function (stub) {
            var r = rooms.get(stub.roomId);
            if (!r) { return; }
            var p   = gridToCanvas(r.x, r.y);
            var len = Math.sqrt(stub.dx * stub.dx + stub.dy * stub.dy);
            if (len === 0) { return; }
            var nx  = stub.dx / len;
            var ny  = stub.dy / len;
            ctx.beginPath();
            ctx.moveTo(p.px, p.py);
            ctx.lineTo(p.px + nx * stubLen, p.py + ny * stubLen);
            ctx.stroke();
        });

        // Draw room squares
        var scaledSize        = ROOM_SIZE        * zoomScale;
        var scaledBorderWidth = ROOM_BORDER_WIDTH * zoomScale;
        var scaledFontSize    = SYMBOL_FONT_SIZE  * zoomScale;
        var half              = scaledSize / 2;

        rooms.forEach(function (room, id) {
            var p         = gridToCanvas(room.x, room.y);
            var isCurrent = (id === currentRoomId);
            var fillColor = isCurrent ? CURRENT_ROOM_COLOR : colorForSymbol(room.symbol);
            var rx        = p.px - half;
            var ry        = p.py - half;

            // Room fill
            ctx.fillStyle = fillColor;
            ctx.fillRect(rx, ry, scaledSize, scaledSize);

            // Room border
            ctx.strokeStyle = ROOM_BORDER_COLOR;
            ctx.lineWidth   = scaledBorderWidth;
            ctx.strokeRect(rx, ry, scaledSize, scaledSize);

            // Symbol label
            var sym = room.symbol || '•';
            ctx.fillStyle    = isCurrent ? CURRENT_ROOM_TEXT_COLOR : SYMBOL_TEXT_COLOR;
            ctx.font         = 'bold ' + scaledFontSize + 'px monospace';
            ctx.textAlign    = 'center';
            ctx.textBaseline = 'middle';
            ctx.fillText(sym, p.px, p.py);
        });
    }

    // -------------------------------------------------------------------------
    // Room and edge management
    // -------------------------------------------------------------------------

    function addOrUpdateRoom(id, gx, gy, symbol) {
        rooms.set(id, { x: gx, y: gy, symbol: symbol || '•' });
    }

    function addEdge(idA, idB) {
        var key = idA < idB ? (idA + '-' + idB) : (idB + '-' + idA);
        edges.add(key);
    }

    function resetMap() {
        rooms.clear();
        edges.clear();
        zoneExitStubs = [];
        currentRoomId = null;
    }

    function replayZone(zoneKey) {
        resetMap();
        var zoneRooms = allRooms.roomZones[zoneKey];
        if (!Array.isArray(zoneRooms)) { return; }

        // First pass: add all rooms
        zoneRooms.forEach(function (r) {
            addOrUpdateRoom(r.RoomId, r.x, r.y, r.symbol);
        });

        // Second pass: add edges and zone-exit stubs
        zoneRooms.forEach(function (r) {
            if (Array.isArray(r.exits)) {
                r.exits.forEach(function (exitId) {
                    if (rooms.has(exitId)) {
                        addEdge(r.RoomId, exitId);
                    }
                });
            }
            if (Array.isArray(r.stubs)) {
                r.stubs.forEach(function (stub) {
                    zoneExitStubs.push({ roomId: r.RoomId, dx: stub.dx, dy: stub.dy });
                });
            }
        });
    }

    // -------------------------------------------------------------------------
    // World.Map ingestion
    // Loads the full visited-room snapshot from the server, populating all
    // zone caches and the roomInfoStore.  Called once per session.
    // -------------------------------------------------------------------------
    function ingestWorldMap(entries) {
        if (!Array.isArray(entries) || entries.length === 0) { return; }

        entries.forEach(function (info) {
            var id = info.num;
            if (!id) { return; }

            // Store for tooltip use
            roomInfoStore.set(id, info);

            // Parse coordinates from the coord string
            var coords   = info.coords ? info.coords.split(',').map(function (s) { return s.trim(); }) : null;
            if (!coords || coords.length < 4) { return; }

            var zoneName = coords[0];
            var gx       = parseInt(coords[1], 10);
            var gy       = parseInt(coords[2], 10);
            var gz       = parseInt(coords[3], 10);

            // Skip rooms at 0,0 that are not the zone root — they have no
            // valid mapped position and would render on top of the origin.
            var isZoneRoot = Array.isArray(info.details) && info.details.indexOf('root') !== -1;
            if (gx === 0 && gy === 0 && !isZoneRoot) { return; }

            var zoneKey  = zoneName + '/z:' + gz;

            if (!Array.isArray(allRooms.roomZones[zoneKey])) {
                allRooms.roomZones[zoneKey] = [];
            }

            // Build exit lines from exitsv2.
            // Secret exits are only drawn if the destination has been visited.
            // Cross-zone exits become direction stubs; same-zone exits are edges.
            var exitIds   = [];
            var exitStubs = [];
            if (info.exitsv2) {
                for (var dir in info.exitsv2) {
                    var exitInfo = info.exitsv2[dir];
                    if (exitInfo.dz !== 0) { continue; }

                    var isSecret    = Array.isArray(exitInfo.details) && exitInfo.details.indexOf('secret') !== -1;
                    var destVisited = roomInfoStore.has(exitInfo.num);

                    // Secret exits are suppressed until the destination is visited.
                    if (isSecret && !destVisited) { continue; }

                    // Determine whether the destination is in the same zone.
                    var destInfo    = roomInfoStore.get(exitInfo.num);
                    var isCrossZone = false;
                    if (destInfo && destInfo.coords) {
                        var destCoords  = destInfo.coords.split(',').map(function (s) { return s.trim(); });
                        if (destCoords.length >= 4) {
                            var destZoneKey = destCoords[0] + '/z:' + destCoords[3];
                            isCrossZone = (destZoneKey !== zoneKey);
                        }
                    }

                    if (isCrossZone || !destInfo) {
                        exitStubs.push({ dx: exitInfo.dx, dy: exitInfo.dy });
                    } else {
                        exitIds.push(exitInfo.num);
                    }
                }
            }

            var sym = info.mapsymbol || '•';

            // Upsert into zone cache
            var zoneArr  = allRooms.roomZones[zoneKey];
            var existing = -1;
            for (var i = 0; i < zoneArr.length; i++) {
                if (zoneArr[i].RoomId === id) { existing = i; break; }
            }
            var record = { RoomId: id, x: gx, y: gy, symbol: sym, exits: exitIds, stubs: exitStubs };
            if (existing === -1) {
                zoneArr.push(record);
            } else {
                zoneArr[existing] = record;
            }
        });

        // Rebuild the visible map for the current zone
        if (allRooms.currentZoneKey) {
            var savedCurrentId = currentRoomId;
            replayZone(allRooms.currentZoneKey);
            currentRoomId = savedCurrentId;
        }

        render();
    }

    // -------------------------------------------------------------------------
    // Hit-testing: find the room id under a canvas-relative pixel position.
    // Returns the room id or null.
    // -------------------------------------------------------------------------
    function roomAtCanvasPoint(cx, cy) {
        var half  = (ROOM_SIZE * zoomScale) / 2;
        var found = null;
        rooms.forEach(function (room, id) {
            var p = gridToCanvas(room.x, room.y);
            if (cx >= p.px - half && cx <= p.px + half &&
                cy >= p.py - half && cy <= p.py + half) {
                found = id;
            }
        });
        return found;
    }

    // -------------------------------------------------------------------------
    // Tooltip
    // -------------------------------------------------------------------------
    function ensureTooltip() {
        if (tooltip) { return; }
        tooltip = document.createElement('div');
        tooltip.id = 'map-tooltip';
        document.body.appendChild(tooltip);
    }

    function showTooltip(mouseX, mouseY, info) {
        ensureTooltip();
        clearTimeout(tooltipHideTimer);

        var html = '<div class="tt-name">' + (info.name || 'Unknown') + '</div>';

        var rows = [];
        if (info.environment) { rows.push({ label: 'Env',    value: info.environment }); }
        if (info.maplegend)   { rows.push({ label: 'Type',   value: info.maplegend   }); }
        if (info.mapsymbol)   { rows.push({ label: 'Symbol', value: info.mapsymbol   }); }
        if (info.area)        { rows.push({ label: 'Area',   value: info.area        }); }

        if (rows.length > 0) {
            html += '<hr class="tt-divider">';
            rows.forEach(function (r) {
                html += '<div class="tt-row">' +
                    '<span class="tt-label">' + r.label + '</span>' +
                    '<span class="tt-value">' + r.value + '</span>' +
                    '</div>';
            });
        }

        var details    = info.details || [];
        var badgeOrder = ['pvp', 'bank', 'trainer', 'storage', 'character', 'ephemeral'];
        var badges     = badgeOrder.filter(function (d) { return details.indexOf(d) !== -1; });
        if (badges.length > 0) {
            html += '<hr class="tt-divider"><div class="tt-badges">';
            badges.forEach(function (b) {
                html += '<span class="tt-badge ' + b + '">' + b + '</span>';
            });
            html += '</div>';
        }

        if (info.exitsv2) {
            var exitNames = Object.keys(info.exitsv2).filter(function (dir) {
                var e = info.exitsv2[dir];
                var isSecret = Array.isArray(e.details) && e.details.indexOf('secret') !== -1;
                return !isSecret || roomInfoStore.has(e.num);
            }).sort();
            if (exitNames.length > 0) {
                html += '<hr class="tt-divider">' +
                    '<div class="tt-row">' +
                    '<span class="tt-label">Exits</span>' +
                    '<span class="tt-value">' + exitNames.join(', ') + '</span>' +
                    '</div>';
            }
        }

        tooltip.innerHTML     = html;
        tooltip.style.display = 'block';
        positionTooltip(mouseX, mouseY);
    }

    function positionTooltip(mouseX, mouseY) {
        if (!tooltip) { return; }
        var ttW  = tooltip.offsetWidth;
        var ttH  = tooltip.offsetHeight;
        var vw   = window.innerWidth;
        var vh   = window.innerHeight;
        var left = mouseX + 14;
        if (left + ttW > vw - 8) { left = mouseX - ttW - 14; }
        left = Math.max(8, left);
        var top = mouseY - Math.floor(ttH / 2);
        if (top + ttH > vh - 8) { top = vh - ttH - 8; }
        top = Math.max(8, top);
        tooltip.style.left = left + 'px';
        tooltip.style.top  = top  + 'px';
    }

    function hideTooltip() {
        tooltipHideTimer = setTimeout(function () {
            if (tooltip) { tooltip.style.display = 'none'; }
        }, 80);
    }

    // -------------------------------------------------------------------------
    // Zoom controls
    // -------------------------------------------------------------------------
    function zoomIn() {
        zoomScale = Math.min(ZOOM_MAX, zoomScale * ZOOM_STEP);
        render();
    }

    function zoomOut() {
        zoomScale = Math.max(ZOOM_MIN, zoomScale / ZOOM_STEP);
        render();
    }

    // -------------------------------------------------------------------------
    // DOM factory
    // -------------------------------------------------------------------------
    function createDOM() {
        resetMap();
        allRooms.currentZoneKey = '';
        allRooms.roomZones      = {};

        container = document.createElement('div');
        container.id = 'map-canvas-container';

        canvas = document.createElement('canvas');
        canvas.id = 'map-canvas';
        container.appendChild(canvas);
        ctx = canvas.getContext('2d');

        canvas.addEventListener('mousemove', function (e) {
            var rect = canvas.getBoundingClientRect();
            var id   = roomAtCanvasPoint(e.clientX - rect.left, e.clientY - rect.top);
            var info = id !== null ? roomInfoStore.get(id) : null;
            if (info) {
                clearTimeout(tooltipHideTimer);
                showTooltip(e.clientX, e.clientY, info);
            } else {
                hideTooltip();
            }
        });

        canvas.addEventListener('mouseleave', hideTooltip);

        var controls = document.createElement('div');
        controls.id = 'map-zoom-controls';

        var btnOut = document.createElement('button');
        btnOut.textContent = '\u2212';
        btnOut.addEventListener('click', zoomOut);

        var btnIn = document.createElement('button');
        btnIn.textContent = '+';
        btnIn.addEventListener('click', zoomIn);

        controls.appendChild(btnOut);
        controls.appendChild(btnIn);
        container.appendChild(controls);

        document.body.appendChild(container);
        return container;
    }

    // -------------------------------------------------------------------------
    // VirtualWindow instance
    // -------------------------------------------------------------------------
    var win = new VirtualWindow('Map', {
        dock:          'right',
        defaultDocked: true,
        dockedHeight:  363,
        factory: function () {
            var el = createDOM();
            return {
                title:      'Map',
                mount:      el,
                background: '#1e1e1e',
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

    // -------------------------------------------------------------------------
    // Update logic
    // -------------------------------------------------------------------------

    function updateWorldMap() {
        var worldData = Client.GMCPStructs.World;
        if (!worldData || !Array.isArray(worldData.Map)) { return; }

        win.open();
        if (!win.isOpen()) { return; }

        resizeCanvas();
        ingestWorldMap(worldData.Map);
    }

    function updateMap() {
        var obj = Client.GMCPStructs.Room;
        if (!obj || !obj.Info) { return; }

        win.open();
        if (!win.isOpen()) { return; }

        // Request the full visited-room snapshot once per session
        if (!worldMapRequested) {
            worldMapRequested = true;
            Client.GMCPRequest('World.Map');
        }

        var info = obj.Info;

        var winBox = win.get();
        if (winBox) {
            winBox.setTitle('map (' + info.area + ')');
        }

        // Resize canvas to match container each update (handles dock resize)
        resizeCanvas();

        // Parse coordinate string: "zoneName, x, y, z"
        var coords   = info.coords.split(',').map(function (s) { return s.trim(); });
        var zoneName = coords[0];
        var zoneKey  = zoneName + '/z:' + coords[3];

        if (allRooms.currentZoneKey !== zoneKey) {
            allRooms.currentZoneKey = zoneKey;
            if (!Array.isArray(allRooms.roomZones[zoneKey])) {
                allRooms.roomZones[zoneKey] = [];
            }
            replayZone(zoneKey);
        }

        var gx = parseInt(coords[1], 10);
        var gy = parseInt(coords[2], 10);

        // Determine symbol: use room's own mapsymbol if set, else fall back to
        // the biome symbol embedded in the maplegend/environment fields.
        var sym = info.mapsymbol || '•';

        // Build exit lines from exitsv2.
        // Secret exits are only drawn if the destination has been visited.
        // Cross-zone exits are drawn as direction stubs; same-zone exits are
        // drawn as full edges to the destination room.
        var exitIds   = [];
        var exitStubs = [];
        for (var dir in info.exitsv2) {
            var exitInfo = info.exitsv2[dir];
            if (exitInfo.dz !== 0) { continue; }

            var isSecret     = Array.isArray(exitInfo.details) && exitInfo.details.indexOf('secret') !== -1;
            var destVisited  = roomInfoStore.has(exitInfo.num);

            // Secret exits are suppressed until the destination is visited.
            if (isSecret && !destVisited) { continue; }

            // Determine whether the destination is in the same zone.
            var destInfo     = roomInfoStore.get(exitInfo.num);
            var isCrossZone  = false;
            if (destInfo && destInfo.coords) {
                var destCoords  = destInfo.coords.split(',').map(function (s) { return s.trim(); });
                if (destCoords.length >= 4) {
                    var destZoneKey = destCoords[0] + '/z:' + destCoords[3];
                    isCrossZone = (destZoneKey !== zoneKey);
                }
            }

            if (isCrossZone) {
                // Draw a stub line in the exit direction.
                exitStubs.push({ dx: exitInfo.dx, dy: exitInfo.dy });
                zoneExitStubs.push({ roomId: info.num, dx: exitInfo.dx, dy: exitInfo.dy });
            } else if (destInfo) {
                // Same zone, destination is known — draw a full edge.
                exitIds.push(exitInfo.num);
                if (!rooms.has(exitInfo.num)) {
                    var destCoords2 = destInfo.coords.split(',').map(function (s) { return s.trim(); });
                    addOrUpdateRoom(
                        exitInfo.num,
                        parseInt(destCoords2[1], 10),
                        parseInt(destCoords2[2], 10),
                        destInfo.mapsymbol || '•'
                    );
                }
                addEdge(info.num, exitInfo.num);
            } else {
                // Destination not yet visited — draw a stub in the exit direction
                // so the player can see there is an exit without rendering an
                // unknown room square at an unverified position.
                exitStubs.push({ dx: exitInfo.dx, dy: exitInfo.dy });
                zoneExitStubs.push({ roomId: info.num, dx: exitInfo.dx, dy: exitInfo.dy });
            }
        }

        // Store GMCP info for tooltip use
        roomInfoStore.set(info.num, info);

        // Add / update the current room
        addOrUpdateRoom(info.num, gx, gy, sym);
        currentRoomId = info.num;

        // Persist to zone cache (deduplicated by RoomId)
        var zoneArr  = allRooms.roomZones[zoneKey];
        var existing = -1;
        for (var i = 0; i < zoneArr.length; i++) {
            if (zoneArr[i].RoomId === info.num) { existing = i; break; }
        }
        var record = { RoomId: info.num, x: gx, y: gy, symbol: sym, exits: exitIds, stubs: exitStubs };
        if (existing === -1) {
            zoneArr.push(record);
        } else {
            zoneArr[existing] = record;
        }

        render();
    }

    // -------------------------------------------------------------------------
    // Handle container resize (ResizeObserver when available)
    // -------------------------------------------------------------------------
    function setupResizeObserver() {
        if (typeof ResizeObserver === 'undefined') { return; }
        var ro = new ResizeObserver(function () {
            resizeCanvas();
            render();
        });
        // Observe lazily — container may not exist yet at registration time.
        var orig = win.open.bind(win);
        win.open = function () {
            orig();
            if (container) { ro.observe(container); }
        };
    }
    setupResizeObserver();

    // -------------------------------------------------------------------------
    // Registration
    // -------------------------------------------------------------------------
    VirtualWindows.register({
        window:       win,
        gmcpHandlers: ['Room', 'World'],
        onGMCP: function (namespace) {
            if (namespace === 'World.Map') {
                updateWorldMap();
            } else {
                updateMap();
            }
        },
    });

}());
