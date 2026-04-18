/* global Client, ResizeObserver, VirtualWindow, VirtualWindows, injectStyles, uiMenu */

/**
 * window-map3d.js
 *
 * Virtual window: isometric 3-D map.
 *
 * Rooms are drawn as isometric tiles (diamond top face + two side faces).
 * Unlike the flat map, exits that cross z-planes are followed, so vertically
 * connected rooms appear stacked above or below their neighbours.
 *
 * Responds to the same GMCP namespaces as window-map.js:
 *   Room      - incremental update as the player moves room-to-room
 *   World.Map - bulk snapshot of all visited rooms
 *
 * Reads:
 *   Client.GMCPStructs.Room.Info
 *   Client.GMCPStructs.World.Map
 */

'use strict';

(function () {

    // -------------------------------------------------------------------------
    // Layout constants
    // -------------------------------------------------------------------------

    /** Half-width of a tile diamond in pixels at zoom 1. */
    var TILE_HW = 20;

    /** Half-height of the tile top face in pixels at zoom 1 (≈ HW/2 for 2:1 iso). */
    var TILE_HH = 10;

    /** Height of the vertical side faces in pixels at zoom 1. */
    var TILE_DEPTH = 7;

    /**
     * Grid spacing multiplier in the X/Y plane.
     * A value of 1.0 means tiles are flush edge-to-edge; increase to add gaps.
     * The screen step between adjacent tile centres = TILE_HW * GRID_STEP_XY
     * (for the X component) and TILE_HH * GRID_STEP_XY (for the Y component).
     */
    var GRID_STEP_XY = 1.6;

    /**
     * Screen pixels between tile centres per z-level at zoom 1.
     * Intentionally large so different z-planes are unmistakably separated.
     */
    var Z_STEP = 120;

    /** Multiplicative step applied on each zoom in/out button press. */
    var ZOOM_STEP = 1.25;

    /** Minimum and maximum allowed zoom scale. */
    var ZOOM_MIN = 0.25;
    var ZOOM_MAX = 4.0;

    /**
     * Duration in seconds for the camera to ease to a new room.
     * Set to 0 to disable easing and snap instantly.
     */
    var CENTER_EASE_DURATION = 0.2;

    /** Stroke width of connection lines between rooms. */
    var CONNECTION_WIDTH = 2;

    /** Background color of the map canvas. */
    var MAP_BG = '#1e1e2e';

    /** Border color of tile edges. */
    var TILE_BORDER_COLOR = '#000000';

    /** Line width for tile borders. */
    var TILE_BORDER_WIDTH = 0.8;

    /** Color of connection lines between rooms. */
    var CONNECTION_COLOR = '#7a4a1a';

    /** Fill color for the player's current room top face. */
    var CURRENT_ROOM_COLOR = '#c20000';

    /** Side-face darkening factor (0–1). */
    var SIDE_DARKEN = 0.55;

    /** Text color for symbol labels. */
    var SYMBOL_TEXT_COLOR = '#e0e0e0';

    /** Text color for symbol inside the current room. */
    var CURRENT_ROOM_TEXT_COLOR = '#ffffff';

    /** Font size for the symbol drawn on each tile (in pixels at zoom 1). */
    var SYMBOL_FONT_SIZE = 10;

    /**
     * Symbol-to-color lookup.
     */
    var SYMBOL_COLORS = {
        // Water
        '~':  '#2a53f7',   // shore / water edge
        '≈':  '#0033cd',   // open water

        // Terrain
        '♣':  '#1a6b1a',   // forest — dark green
        '♨':  '#4a6b20',   // swamp — murky green
        '❄':  '#b8d8f0',   // snow — pale blue-white
        '⌬':  '#5a4a38',   // cave — dark brown-grey
        '⩕':  '#7a6a50',   // mountains — warm grey-brown
        '▼':  '#8a7a5a',   // cliffs — tan
        '⌂':  '#8a6a3a',   // house / building
        '*':  '#d4aa55',   // desert — sandy gold
        "'":  '#6a8a30',   // farmland — olive green
        '=':  '#a07840',   // road — light brown

        // Special room types
        '$':  '#2a7a2a',   // shop — green
        '%':  '#2a5a8a',   // trainer — blue
        '♜':  '#4a4a4a',   // wall — dark grey
        '+':  '#5fb7ff',   // healer — light blue
        '•':  '#3a3a4a',   // generic / default
    };

    var DEFAULT_ROOM_COLOR = '#3a3a4a';

    /**
     * Maps the GMCP `environment` field (biome name) to a map symbol when the
     * room has no explicit `mapsymbol` set.
     */
    var ENVIRONMENT_SYMBOLS = {
        'Forest':    '\u2663',
        'Swamp':     '\u2668',
        'Snow':      '\u2744',
        'Cave':      '\u232c',
        'Dungeon':   '\u232c',
        'Mountains': '\u2a55',
        'Cliffs':    '\u25bc',
        'House':     '\u2302',
        'Desert':    '*',
        'Farmland':  "'",
        'Road':      '=',
        'Shore':     '~',
        'Water':     '\u2248',
    };

    /**
     * Return the best symbol for a room given its GMCP info object.
     * Prefers the explicit mapsymbol; falls back to the environment name lookup;
     * then falls back to the default dot.
     */
    function symbolForRoom(info) {
        if (info.mapsymbol) { return info.mapsymbol; }
        if (info.environment && ENVIRONMENT_SYMBOLS[info.environment]) {
            return ENVIRONMENT_SYMBOLS[info.environment];
        }
        return '\u2022';
    }

    // -------------------------------------------------------------------------
    // Styles
    // -------------------------------------------------------------------------
    injectStyles([
        '#map3d-container {',
        '    width: 100%;',
        '    height: 100%;',
        '    background: ' + MAP_BG + ';',
        '    overflow: hidden;',
        '    position: relative;',
        '}',
        '#map3d-canvas {',
        '    display: block;',
        '    position: absolute;',
        '    top: 0; left: 0;',
        '    cursor: grab;',
        '}',
        '#map3d-tooltip {',
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
        '#map3d-tooltip .tt-name {',
        '    font-size: 0.85em;',
        '    font-weight: bold;',
        '    color: #dffbd1;',
        '    margin-bottom: 4px;',
        '    line-height: 1.3;',
        '}',
        '#map3d-tooltip .tt-divider {',
        '    border: none;',
        '    border-top: 1px solid #1c6b60;',
        '    margin: 5px 0;',
        '}',
        '#map3d-tooltip .tt-row {',
        '    display: flex;',
        '    justify-content: space-between;',
        '    align-items: baseline;',
        '    gap: 8px;',
        '    font-size: 0.75em;',
        '    line-height: 1.6;',
        '}',
        '#map3d-tooltip .tt-label {',
        '    color: #7ab8a0;',
        '    text-transform: uppercase;',
        '    letter-spacing: 0.04em;',
        '    font-size: 0.88em;',
        '    flex-shrink: 0;',
        '}',
        '#map3d-tooltip .tt-value {',
        '    color: #dffbd1;',
        '    text-align: right;',
        '}',
        '#map3d-tooltip .tt-badges {',
        '    display: flex;',
        '    flex-wrap: wrap;',
        '    gap: 3px;',
        '    margin-top: 4px;',
        '}',
        '#map3d-tooltip .tt-badge {',
        '    font-size: 0.62em;',
        '    padding: 1px 4px;',
        '    border-radius: 3px;',
        '    background: #1a2e28;',
        '    color: #7ab8a0;',
        '    border: 1px solid #1c6b60;',
        '}',
        '#map3d-tooltip .tt-badge.pvp     { background: #3d0f0f; color: #e06060; border-color: #6b1c1c; }',
        '#map3d-tooltip .tt-badge.bank    { background: #0f2e10; color: #56d44a; border-color: #1c6b1c; }',
        '#map3d-tooltip .tt-badge.trainer { background: #2e2000; color: #fdd;    border-color: #6b5010; }',
        '#map3d-tooltip .tt-badge.storage { background: #1a1200; color: #c8a800; border-color: #6b5010; }',
        '#map3d-zoom-controls {',
        '    position: absolute;',
        '    top: 6px;',
        '    right: 6px;',
        '    display: flex;',
        '    align-items: center;',
        '    gap: 2px;',
        '    z-index: 10;',
        '}',
        '#map3d-zoom-controls .ctrl-sep {',
        '    width: 1px;',
        '    height: 16px;',
        '    background: #555;',
        '    margin: 0 3px;',
        '}',
        '#map3d-zoom-controls button {',
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
        '#map3d-zoom-controls button:hover {',
        '    background: rgba(0,0,0,0.8);',
        '    color: #fff;',
        '}',
    ].join('\n'));

    // -------------------------------------------------------------------------
    // Module state
    // -------------------------------------------------------------------------

    var canvas    = null;
    var ctx       = null;
    var container = null;

    /**
     * rooms3d: Map<RoomId, { x, y, z, symbol }>
     * All rooms currently in the render set.
     */
    var rooms3d = new Map();

    /**
     * edges3d: Set of canonical "minId-maxId" strings.
     */
    var edges3d = new Map();

    /** RoomId of the player's current room. */
    var currentRoomId = null;

    /** Camera position in grid coordinates (interpolated toward current room). */
    var camX = 0;
    var camY = 0;
    var camZ = 0;

    /** Easing state. */
    var easeStartX    = 0;
    var easeStartY    = 0;
    var easeStartZ    = 0;
    var easeTargetX   = 0;
    var easeTargetY   = 0;
    var easeTargetZ   = 0;
    var easeStartTime = null;
    var easeRafId     = null;

    /** Pan offset in grid coordinates (cleared on room change). */
    var panOffsetX = 0;
    var panOffsetY = 0;

    /** Drag state. */
    var dragActive    = false;
    var dragStartPxX  = 0;
    var dragStartPxY  = 0;
    var dragStartPanX = 0;
    var dragStartPanY = 0;

    /** Current zoom scale. */
    var zoomScale = 1.0;

    /** Per-room GMCP info for tooltips. */
    var roomInfoStore3d = new Map();

    /** Tooltip DOM element. */
    var tooltip = null;
    var tooltipHideTimer = null;

    /**
     * roomCache3d: keyed by roomId.
     * { RoomId, zoneName, x, y, z, symbol, exits, stubs }
     * exits: Array<{ num, locked, secret }>  — includes cross-z exits
     * stubs: Array<{ dx, dy, dz, locked, secret }>
     */
    var roomCache3d = {};

    /**
     * Spacing scale factor, adjusted at runtime by the spacing buttons.
     * Multiplies both GRID_STEP_XY and Z_STEP.
     */
    var spacingScale = (function () {
        var saved = parseFloat(localStorage.getItem('map3d.spacingScale'));
        return (isFinite(saved) && saved >= 0.4 && saved <= 4.0) ? saved : 1.0;
    }());

    /** Step multiplier applied on each spacing in/out button press. */
    var SPACING_STEP = 1.25;

    /** Minimum and maximum spacing scale. */
    var SPACING_MIN = 0.4;
    var SPACING_MAX = 4.0;

    /** Z of the room currently under the mouse, or null when not hovering an off-plane room. */
    var hoveredZ = null;

    var currentRoomKey    = '';
    var worldMapRequested = false;

    // -------------------------------------------------------------------------
    // Helpers
    // -------------------------------------------------------------------------

    /**
     * Color lookup by biome/environment name.
     * Used when a room has no mapsymbol, or its symbol has no entry in SYMBOL_COLORS.
     */
    var ENVIRONMENT_COLORS = {
        'Forest':    '#1a6b1a',
        'Swamp':     '#4a6b20',
        'Snow':      '#b8d8f0',
        'Cave':      '#5a4a38',
        'Dungeon':   '#5a4a38',
        'Mountains': '#7a6a50',
        'Cliffs':    '#8a7a5a',
        'House':     '#8a6a3a',
        'Desert':    '#d4aa55',
        'Farmland':  '#6a8a30',
        'Road':      '#a07840',
        'Shore':     '#2a53f7',
        'Water':     '#0033cd',
        'City':      '#5a5a6a',
        'Fort':      '#5a5a6a',
        'Land':      '#3a3a4a',
    };

    function colorForSymbol(sym, env) {
        if (sym && SYMBOL_COLORS[sym]) { return SYMBOL_COLORS[sym]; }
        if (env && ENVIRONMENT_COLORS[env]) { return ENVIRONMENT_COLORS[env]; }
        return DEFAULT_ROOM_COLOR;
    }

    /**
     * Darken a hex color by multiplying each channel by `factor`.
     */
    function darkenColor(hex, factor) {
        var r = parseInt(hex.slice(1, 3), 16);
        var g = parseInt(hex.slice(3, 5), 16);
        var b = parseInt(hex.slice(5, 7), 16);
        r = Math.round(r * factor);
        g = Math.round(g * factor);
        b = Math.round(b * factor);
        return '#' +
            ('0' + r.toString(16)).slice(-2) +
            ('0' + g.toString(16)).slice(-2) +
            ('0' + b.toString(16)).slice(-2);
    }

    // -------------------------------------------------------------------------
    // Isometric projection
    // -------------------------------------------------------------------------

    /**
     * Convert grid (gx, gy, gz) to canvas pixel coordinates.
     * The camera position (camX, camY, camZ) maps to the canvas centre.
     *
     * 2:1 dimetric iso with explicit grid spacing:
     *   stepXY = TILE_HW * GRID_STEP_XY  (screen pixels per grid unit in X/Y)
     *   screenX = (relX - relY) * stepXY
     *   screenY = (relX + relY) * stepXY/2 - relZ * Z_STEP
     *
     * Z convention: positive z is higher (appears higher on screen, lower sy).
     * Negative z is below ground (appears lower on screen, higher sy).
     */
    function isoProject(gx, gy, gz) {
        var step = TILE_HW * GRID_STEP_XY * spacingScale * zoomScale;
        var zs   = Z_STEP  * spacingScale * zoomScale;
        var midX = Math.floor(canvas.width  / 2);
        var midY = Math.floor(canvas.height / 2);

        var relX = gx - camX - panOffsetX;
        var relY = gy - camY - panOffsetY;
        var relZ = gz - camZ;

        return {
            sx: midX + (relX - relY) * step,
            sy: midY + (relX + relY) * (step / 2) - relZ * zs,
        };
    }

    // -------------------------------------------------------------------------
    // Canvas resize
    // -------------------------------------------------------------------------

    function resizeCanvas() {
        if (!canvas || !container) { return; }
        canvas.width  = container.clientWidth  || 1;
        canvas.height = container.clientHeight || 1;
    }

    // -------------------------------------------------------------------------
    // Camera easing
    // -------------------------------------------------------------------------

    function smoothstep(t) {
        return t * t * (3 - 2 * t);
    }

    function setCameraTarget(tx, ty, tz) {
        panOffsetX = 0;
        panOffsetY = 0;

        if (CENTER_EASE_DURATION <= 0) {
            camX = tx; camY = ty; camZ = tz;
            render();
            return;
        }

        if (easeRafId !== null) { cancelAnimationFrame(easeRafId); easeRafId = null; }

        easeStartX = camX; easeStartY = camY; easeStartZ = camZ;
        easeTargetX = tx;  easeTargetY = ty;  easeTargetZ = tz;
        easeStartTime = null;

        function step(ts) {
            if (easeStartTime === null) { easeStartTime = ts; }
            var elapsed = (ts - easeStartTime) / 1000;
            var t = Math.min(elapsed / CENTER_EASE_DURATION, 1);
            var s = smoothstep(t);
            camX = easeStartX + (easeTargetX - easeStartX) * s;
            camY = easeStartY + (easeTargetY - easeStartY) * s;
            camZ = easeStartZ + (easeTargetZ - easeStartZ) * s;
            render();
            if (t < 1) {
                easeRafId = requestAnimationFrame(step);
            } else {
                easeRafId = null;
            }
        }

        easeRafId = requestAnimationFrame(step);
    }

    // -------------------------------------------------------------------------
    // Rendering
    // -------------------------------------------------------------------------

    /**
     * Compute the screen-space attachment point on a tile's top-face diamond
     * for a connection leaving in direction (dx, dy).
     *
     * Diamond vertices relative to tile centre (sx, sy):
     *   top    (0, -hh) = (-gx, -gy) corner
     *   right  (+hw, 0) = (+gx, -gy) corner
     *   bottom (0, +hh) = (+gx, +gy) corner
     *   left   (-hw, 0) = (-gx, +gy) corner
     *
     * Orthogonal exits use the midpoint of the edge facing that direction.
     * Diagonal exits use the corner vertex in that direction.
     * Vertical-only exits (dz != 0, dx == dy == 0) use the tile centre.
     */
    function tileAttachPoint(sx, sy, dx, dy) {
        var hw = TILE_HW * zoomScale;
        var hh = TILE_HH * zoomScale;

        // Diagonal: corner vertex
        if (dx !== 0 && dy !== 0) {
            // (+1,+1) -> bottom (0,+hh)  (+1,-1) -> right (+hw,0)
            // (-1,+1) -> left (-hw,0)    (-1,-1) -> top (0,-hh)
            if (dx > 0 && dy > 0) { return { sx: sx,       sy: sy + hh }; }
            if (dx > 0 && dy < 0) { return { sx: sx + hw,  sy: sy      }; }
            if (dx < 0 && dy > 0) { return { sx: sx - hw,  sy: sy      }; }
            /* dx < 0 && dy < 0 */  return { sx: sx,       sy: sy - hh };
        }

        // Orthogonal: midpoint of the edge facing that direction
        // +gx edge: between right (+hw,0) and bottom (0,+hh)  -> midpoint (+hw/2, +hh/2)
        // -gx edge: between top  (0,-hh) and left  (-hw,0)   -> midpoint (-hw/2, -hh/2)
        // +gy edge: between bottom (0,+hh) and left (-hw,0)  -> midpoint (-hw/2, +hh/2)
        // -gy edge: between top  (0,-hh) and right (+hw,0)   -> midpoint (+hw/2, -hh/2)
        if (dx > 0) { return { sx: sx + hw / 2, sy: sy + hh / 2 }; }
        if (dx < 0) { return { sx: sx - hw / 2, sy: sy - hh / 2 }; }
        if (dy > 0) { return { sx: sx - hw / 2, sy: sy + hh / 2 }; }
        if (dy < 0) { return { sx: sx + hw / 2, sy: sy - hh / 2 }; }

        // Vertical only (dz != 0, dx == dy == 0): use tile centre
        return { sx: sx, sy: sy };
    }

    /**
     * Draw a connection line between two rooms.
     *
     * Any connection that crosses z-planes always runs centre-to-centre so the
     * line is never obscured by a tile drawn on top of it.  Same-z connections
     * use the edge/corner attachment point based on exit direction.
     */
    function drawConnection(ax, ay, az, bx, by, bz, dx, dy, dz) {
        var pA = isoProject(ax, ay, az);
        var pB = isoProject(bx, by, bz);

        var startPt, endPt;
        if (dz !== 0) {
            // Any cross-z connection: centre to centre so tiles never obscure the line
            startPt = pA;
            endPt   = pB;
        } else {
            // Same-z: use the edge/corner attachment points
            startPt = tileAttachPoint(pA.sx, pA.sy,  dx,  dy);
            endPt   = tileAttachPoint(pB.sx, pB.sy, -dx, -dy);
        }

        ctx.beginPath();
        ctx.moveTo(startPt.sx, startPt.sy);
        ctx.lineTo(endPt.sx,   endPt.sy);
        ctx.stroke();
    }

    function render() {
        if (!ctx || !canvas) { return; }

        ctx.clearRect(0, 0, canvas.width, canvas.height);
        ctx.fillStyle = MAP_BG;
        ctx.fillRect(0, 0, canvas.width, canvas.height);

        if (rooms3d.size === 0) { return; }

        // Build a sorted render list: painter's algorithm.
        // Sort key: lower (gx + gy - gz) renders first (farther back in iso view).
        var list = [];
        rooms3d.forEach(function (room, id) {
            list.push({ id: id, x: room.x, y: room.y, z: room.z, symbol: room.symbol });
        });
        // Painter's algorithm: higher z (higher floors, lower sy) renders last (on top).
        // Within the same z-level, rooms further back in iso space (higher gx+gy) render first.
        list.sort(function (a, b) {
            var ka = a.x + a.y - a.z * 2;
            var kb = b.x + b.y - b.z * 2;
            return ka - kb;
        });

        // Determine the active z for fade: hover overrides player position
        var playerZ = (currentRoomId !== null && rooms3d.has(currentRoomId))
            ? rooms3d.get(currentRoomId).z
            : (camZ | 0);
        var activeZ = (hoveredZ !== null) ? hoveredZ : playerZ;

        // Draw connections first (behind tiles)
        ctx.strokeStyle = CONNECTION_COLOR;
        ctx.lineWidth   = CONNECTION_WIDTH * zoomScale;
        ctx.lineCap     = 'round';

        edges3d.forEach(function (edge, key) {
            var parts = key.split('-');
            var idA   = parseInt(parts[0], 10);
            var idB   = parseInt(parts[1], 10);
            var rA    = rooms3d.get(idA);
            var rB    = rooms3d.get(idB);
            if (!rA || !rB) { return; }
            var zDiff = Math.max(
                Math.abs(rA.z - activeZ),
                Math.abs(rB.z - activeZ)
            );
            ctx.globalAlpha = zDiff === 0 ? 1.0 : 0.25;
            drawConnection(rA.x, rA.y, rA.z, rB.x, rB.y, rB.z, edge.dx, edge.dy, edge.dz);
        });
        ctx.globalAlpha = 1.0;

        // Draw tiles in painter order, fading rooms on other z-planes
        list.forEach(function (item) {
            var isCurrent = (item.id === currentRoomId);
            var topColor  = isCurrent ? CURRENT_ROOM_COLOR : colorForSymbol(item.symbol, item.env);
            var zDiff     = Math.abs(item.z - activeZ);
            ctx.globalAlpha = zDiff === 0 ? 1.0 : 0.25;
            drawTileById(item.id, item.x, item.y, item.z, topColor, isCurrent, item.symbol);
        });
        ctx.globalAlpha = 1.0;
    }

    /**
     * Draw a single isometric tile at grid position (gx, gy, gz).
     * Composed of three faces: top diamond, left side face, right side face.
     */
    function drawTileById(id, gx, gy, gz, topColor, isCurrent, symbol) {
        var hw   = TILE_HW    * zoomScale;
        var hh   = TILE_HH    * zoomScale;
        var dep  = TILE_DEPTH * zoomScale;
        var bw   = TILE_BORDER_WIDTH * zoomScale;

        var p    = isoProject(gx, gy, gz);
        var sx   = p.sx;
        var sy   = p.sy;

        var leftColor  = darkenColor(topColor, SIDE_DARKEN * 0.8);
        var rightColor = darkenColor(topColor, SIDE_DARKEN);

        // Top face
        ctx.beginPath();
        ctx.moveTo(sx,      sy - hh);
        ctx.lineTo(sx + hw, sy     );
        ctx.lineTo(sx,      sy + hh);
        ctx.lineTo(sx - hw, sy     );
        ctx.closePath();
        ctx.fillStyle = topColor;
        ctx.fill();
        ctx.strokeStyle = TILE_BORDER_COLOR;
        ctx.lineWidth   = bw;
        ctx.stroke();

        // Left side face
        ctx.beginPath();
        ctx.moveTo(sx - hw, sy          );
        ctx.lineTo(sx,      sy + hh     );
        ctx.lineTo(sx,      sy + hh + dep);
        ctx.lineTo(sx - hw, sy      + dep);
        ctx.closePath();
        ctx.fillStyle = leftColor;
        ctx.fill();
        ctx.strokeStyle = TILE_BORDER_COLOR;
        ctx.lineWidth   = bw;
        ctx.stroke();

        // Right side face
        ctx.beginPath();
        ctx.moveTo(sx,      sy + hh     );
        ctx.lineTo(sx + hw, sy          );
        ctx.lineTo(sx + hw, sy      + dep);
        ctx.lineTo(sx,      sy + hh + dep);
        ctx.closePath();
        ctx.fillStyle = rightColor;
        ctx.fill();
        ctx.strokeStyle = TILE_BORDER_COLOR;
        ctx.lineWidth   = bw;
        ctx.stroke();

        // Symbol on top face
        var sym = symbol || '•';
        ctx.fillStyle    = isCurrent ? CURRENT_ROOM_TEXT_COLOR : SYMBOL_TEXT_COLOR;
        ctx.font         = 'bold ' + (SYMBOL_FONT_SIZE * zoomScale) + 'px monospace';
        ctx.textAlign    = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText(sym, sx, sy);
    }

    // -------------------------------------------------------------------------
    // Room and edge management
    // -------------------------------------------------------------------------

    function addRoom3d(id, gx, gy, gz, symbol, env) {
        rooms3d.set(id, { x: gx, y: gy, z: gz, symbol: symbol || '•', env: env || '' });
    }

    /**
     * Record an edge between two rooms, storing the direction delta from A to B.
     * The canonical key is always minId-maxId; the stored delta is always from
     * the lower-id room toward the higher-id room.
     */
    function addEdge3d(idA, idB, rA, rB) {
        var key, dx, dy, dz;
        if (idA < idB) {
            key = idA + '-' + idB;
            dx = rB.x - rA.x; dy = rB.y - rA.y; dz = rB.z - rA.z;
        } else {
            key = idB + '-' + idA;
            dx = rA.x - rB.x; dy = rA.y - rB.y; dz = rA.z - rB.z;
        }
        if (!edges3d.has(key)) {
            edges3d.set(key, { dx: dx, dy: dy, dz: dz });
        }
    }

    function resetMap3d() {
        rooms3d.clear();
        edges3d.clear();
        currentRoomId = null;
        camX = 0; camY = 0; camZ = 0;
        panOffsetX = 0; panOffsetY = 0;
        dragActive = false;
        if (easeRafId !== null) { cancelAnimationFrame(easeRafId); easeRafId = null; }
    }

    function replayZone3d(startId) {
        resetMap3d();

        if (!roomCache3d[startId]) { return; }

        var visited = {};
        var queue   = [startId];

        while (queue.length > 0) {
            var id = queue.shift();
            if (visited[id]) { continue; }
            visited[id] = true;

            var r = roomCache3d[id];
            if (!r) { continue; }

            addRoom3d(r.RoomId, r.x, r.y, r.z, r.symbol, r.env);

            if (Array.isArray(r.exits)) {
                r.exits.forEach(function (exit) {
                    if (!visited[exit.num] && roomCache3d[exit.num]) {
                        queue.push(exit.num);
                    }
                });
            }
        }

        // Second pass: add edges for rooms now in the render set
        rooms3d.forEach(function (room, id) {
            var r = roomCache3d[id];
            if (!r || !Array.isArray(r.exits)) { return; }
            r.exits.forEach(function (exit) {
                if (rooms3d.has(exit.num)) {
                    addEdge3d(id, exit.num, r, roomCache3d[exit.num]);
                }
            });
        });
    }

    // -------------------------------------------------------------------------
    // World.Map ingestion
    // -------------------------------------------------------------------------

    function ingestWorldMap3d(entries) {
        if (!Array.isArray(entries) || entries.length === 0) { return; }

        // First pass: populate roomInfoStore3d
        entries.forEach(function (info) {
            if (info.num) {
                roomInfoStore3d.set(info.num, info);
            }
        });

        // Second pass: build cache
        entries.forEach(function (info) {
            var id = info.num;
            if (!id) { return; }

            var coords = info.coords ? info.coords.split(',').map(function (s) { return s.trim(); }) : null;
            if (!coords || coords.length < 4) { return; }

            var zoneName = coords[0];
            var gx       = parseInt(coords[1], 10);
            var gy       = parseInt(coords[2], 10);
            var gz       = parseInt(coords[3], 10);

            var isZoneRoot = Array.isArray(info.details) && info.details.indexOf('root') !== -1;
            if (gx === 0 && gy === 0 && !isZoneRoot) { return; }

            upsertRoomCache3d(id, zoneName, gx, gy, gz, symbolForRoom(info), info.environment || '', info.exitsv2);
        });

        // Rebuild from current room
        if (currentRoomId && roomCache3d[currentRoomId]) {
            var savedId = currentRoomId;
            replayZone3d(savedId);
            currentRoomId = savedId;
        }

        render();
    }

    /**
     * Insert or update a room in roomCache3d.
     * Unlike the flat map, we include cross-z exits as full edges when the
     * destination is visited, not just same-z ones.
     */
    function upsertRoomCache3d(id, zoneName, gx, gy, gz, sym, env, exitsv2) {
        var exitIds   = [];
        var exitStubs = [];

        if (exitsv2) {
            for (var dir in exitsv2) {
                var exitInfo = exitsv2[dir];

                // Skip exits with no spatial delta (portals, etc.)
                if (exitInfo.dx === 0 && exitInfo.dy === 0 && exitInfo.dz === 0) { continue; }

                var isSecret    = Array.isArray(exitInfo.details) && exitInfo.details.indexOf('secret') !== -1;
                var isLocked    = Array.isArray(exitInfo.details) && exitInfo.details.indexOf('locked') !== -1;
                var destVisited = roomInfoStore3d.has(exitInfo.num);

                if (isSecret && !destVisited) { continue; }

                if (destVisited) {
                    exitIds.push({ num: exitInfo.num, locked: isLocked, secret: isSecret });
                } else {
                    exitStubs.push({ dx: exitInfo.dx, dy: exitInfo.dy, dz: exitInfo.dz, locked: isLocked, secret: isSecret });
                }
            }
        }

        roomCache3d[id] = {
            RoomId:   id,
            zoneName: zoneName,
            x:        gx,
            y:        gy,
            z:        gz,
            symbol:   sym,
            env:      env,
            exits:    exitIds,
            stubs:    exitStubs,
        };
    }

    // -------------------------------------------------------------------------
    // Hit-testing: find room under canvas-relative pixel position
    // -------------------------------------------------------------------------

    function roomAtPoint(cx, cy) {
        var step  = TILE_HW * GRID_STEP_XY * spacingScale * zoomScale;
        var hw    = step;
        var hh    = step / 2;
        var found = null;

        // Iterate in reverse painter order so the topmost tile wins
        var list = [];
        rooms3d.forEach(function (room, id) {
            list.push({ id: id, x: room.x, y: room.y, z: room.z });
        });
        list.sort(function (a, b) {
            return (b.x + b.y - b.z * 2) - (a.x + a.y - a.z * 2);
        });

        for (var i = 0; i < list.length; i++) {
            var item = list[i];
            var p    = isoProject(item.x, item.y, item.z);
            var dx   = cx - p.sx;
            var dy   = cy - p.sy;
            if (Math.abs(dx) / hw + Math.abs(dy) / hh <= 1) {
                found = item.id;
                break;
            }
        }
        return found;
    }

    // -------------------------------------------------------------------------
    // Tooltip
    // -------------------------------------------------------------------------

    function ensureTooltip() {
        if (tooltip) { return; }
        tooltip = document.createElement('div');
        tooltip.id = 'map3d-tooltip';
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
        if (info.coords) {
            var c = info.coords.split(',').map(function (s) { return s.trim(); });
            if (c.length >= 4) {
                rows.push({ label: 'Z', value: c[3] });
            }
        }

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
                return !isSecret || roomInfoStore3d.has(e.num);
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
    // Zoom and spacing controls
    // -------------------------------------------------------------------------

    function zoomIn() {
        zoomScale = Math.min(ZOOM_MAX, zoomScale * ZOOM_STEP);
        render();
    }

    function zoomOut() {
        zoomScale = Math.max(ZOOM_MIN, zoomScale / ZOOM_STEP);
        render();
    }

    function spacingIn() {
        spacingScale = Math.min(SPACING_MAX, spacingScale * SPACING_STEP);
        localStorage.setItem('map3d.spacingScale', spacingScale);
        render();
    }

    function spacingOut() {
        spacingScale = Math.max(SPACING_MIN, spacingScale / SPACING_STEP);
        localStorage.setItem('map3d.spacingScale', spacingScale);
        render();
    }

    // -------------------------------------------------------------------------
    // DOM factory
    // -------------------------------------------------------------------------

    function createDOM() {
        resetMap3d();
        currentRoomKey = '';
        roomCache3d    = {};

        container = document.createElement('div');
        container.id = 'map3d-container';

        canvas = document.createElement('canvas');
        canvas.id = 'map3d-canvas';
        container.appendChild(canvas);
        ctx = canvas.getContext('2d');

        canvas.addEventListener('mouseleave', function () {
            hideTooltip();
            if (hoveredZ !== null) {
                hoveredZ = null;
                render();
            }
            if (dragActive) {
                dragActive = false;
                canvas.style.cursor = '';
            }
        });

        // Drag to pan (horizontal only — pans in iso X/Y space)
        canvas.addEventListener('mousedown', function (e) {
            if (e.button !== 0) { return; }
            dragActive    = true;
            dragStartPxX  = e.clientX;
            dragStartPxY  = e.clientY;
            dragStartPanX = panOffsetX;
            dragStartPanY = panOffsetY;
            canvas.style.cursor = 'grabbing';
            e.preventDefault();
        });

        canvas.addEventListener('mousemove', function (e) {
            var rect = canvas.getBoundingClientRect();
            if (dragActive) {
                // Invert the iso projection to convert screen drag delta to grid delta.
                // screenX = (dgx - dgy) * step  =>  dgx - dgy = dsx / step
                // screenY = (dgx + dgy) * step/2  =>  dgx + dgy = dsy * 2 / step
                var step  = TILE_HW * GRID_STEP_XY * spacingScale * zoomScale;
                var dsx   = e.clientX - dragStartPxX;
                var dsy   = e.clientY - dragStartPxY;
                var dgx   = (dsx / step + dsy * 2 / step) / 2;
                var dgy   = (dsy * 2 / step - dsx / step) / 2;
                panOffsetX = dragStartPanX - dgx;
                panOffsetY = dragStartPanY - dgy;
                render();
                return;
            }
            var id   = roomAtPoint(e.clientX - rect.left, e.clientY - rect.top);
            var info = id !== null ? roomInfoStore3d.get(id) : null;
            if (info) {
                clearTimeout(tooltipHideTimer);
                showTooltip(e.clientX, e.clientY, info);
                // If hovered room is on a different z-plane, focus that plane
                var hRoom = rooms3d.get(id);
                var newHoveredZ = (hRoom && hRoom.z !== null) ? hRoom.z : null;
                if (newHoveredZ !== hoveredZ) {
                    hoveredZ = newHoveredZ;
                    render();
                }
            } else {
                hideTooltip();
                if (hoveredZ !== null) {
                    hoveredZ = null;
                    render();
                }
            }
        });

        canvas.addEventListener('mouseup', function (e) {
            if (!dragActive) { return; }
            var dx = e.clientX - dragStartPxX;
            var dy = e.clientY - dragStartPxY;
            dragActive = false;
            canvas.style.cursor = '';
            if (Math.abs(dx) > 4 || Math.abs(dy) > 4) {
                canvas.dataset.suppressClick = '1';
            }
        });

        canvas.addEventListener('click', function (e) {
            if (canvas.dataset.suppressClick) {
                delete canvas.dataset.suppressClick;
                return;
            }
            var charInfo = Client.GMCPStructs.Char && Client.GMCPStructs.Char.Info;
            if (!charInfo || charInfo.role !== 'admin') { return; }
            var rect = canvas.getBoundingClientRect();
            var id   = roomAtPoint(e.clientX - rect.left, e.clientY - rect.top);
            if (id === null) { return; }
            e.stopPropagation();
            uiMenu(e, [
                { label: 'teleport ' + id,  cmd: 'teleport ' + id  },
                { label: 'room info ' + id, cmd: 'room info ' + id },
            ]);
        });

        // Scroll wheel zoom
        canvas.addEventListener('wheel', function (e) {
            e.preventDefault();
            if (e.deltaY < 0) { zoomIn(); } else { zoomOut(); }
        }, { passive: false });

        var controls = document.createElement('div');
        controls.id = 'map3d-zoom-controls';

        var btnZoomOut = document.createElement('button');
        btnZoomOut.textContent = '\u2212';
        btnZoomOut.title = 'Zoom out';
        btnZoomOut.addEventListener('click', zoomOut);

        var btnZoomIn = document.createElement('button');
        btnZoomIn.textContent = '+';
        btnZoomIn.title = 'Zoom in';
        btnZoomIn.addEventListener('click', zoomIn);

        var sep = document.createElement('span');
        sep.className = 'ctrl-sep';

        var btnSpacingOut = document.createElement('button');
        btnSpacingOut.textContent = '\u2212';
        btnSpacingOut.title = 'Decrease spacing';
        btnSpacingOut.addEventListener('click', spacingOut);

        var btnSpacingIn = document.createElement('button');
        btnSpacingIn.textContent = '+';
        btnSpacingIn.title = 'Increase spacing';
        btnSpacingIn.addEventListener('click', spacingIn);

        controls.appendChild(btnZoomOut);
        controls.appendChild(btnZoomIn);
        controls.appendChild(sep);
        controls.appendChild(btnSpacingOut);
        controls.appendChild(btnSpacingIn);
        container.appendChild(controls);

        document.body.appendChild(container);
        return container;
    }

    // -------------------------------------------------------------------------
    // VirtualWindow instance
    // -------------------------------------------------------------------------

    var win = new VirtualWindow('Map3D', {
        dock:          'right',
        defaultDocked: true,
        dockedHeight:  363,
        factory: function () {
            var el = createDOM();
            return {
                title:      'Map (3D)',
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
        ingestWorldMap3d(worldData.Map);
    }

    function updateMap() {
        var obj = Client.GMCPStructs.Room;
        if (!obj || !obj.Info) { return; }

        win.open();
        if (!win.isOpen()) { return; }

        if (!worldMapRequested) {
            worldMapRequested = true;
            Client.GMCPRequest('World.Map');
        }

        var info = obj.Info;

        var winBox = win.get();
        if (winBox) {
            winBox.setTitle('map 3d (' + info.area + ')');
        }

        resizeCanvas();

        var coords   = info.coords.split(',').map(function (s) { return s.trim(); });
        var zoneName = coords[0];
        var gx       = parseInt(coords[1], 10);
        var gy       = parseInt(coords[2], 10);
        var gz       = parseInt(coords[3], 10);

        var sym = symbolForRoom(info);
        var env = info.environment || '';

        roomInfoStore3d.set(info.num, info);
        upsertRoomCache3d(info.num, zoneName, gx, gy, gz, sym, env, info.exitsv2);

        var roomKey = info.num + '';

        // Rebuild the 3D render set whenever the player enters a new room.
        // For revisited rooms (same id) we still do a cheap incremental update
        // to pick up any newly-discovered neighbours from World.Map.
        if (currentRoomKey !== roomKey) {
            currentRoomKey = roomKey;
            currentRoomId  = info.num;
            replayZone3d(info.num);
        } else {
            // Same room — incrementally add current room and its neighbours
            addRoom3d(info.num, gx, gy, gz, sym, env);
            var rc = roomCache3d[info.num];
            if (rc && Array.isArray(rc.exits)) {
                rc.exits.forEach(function (exit) {
                    var destRc = roomCache3d[exit.num];
                    if (destRc) {
                        if (!rooms3d.has(exit.num)) {
                            addRoom3d(exit.num, destRc.x, destRc.y, destRc.z, destRc.symbol, destRc.env);
                        }
                        addEdge3d(info.num, exit.num, roomCache3d[info.num], destRc);
                    }
                });
            }
        }

        currentRoomId = info.num;
        setCameraTarget(gx, gy, gz);
    }

    // -------------------------------------------------------------------------
    // ResizeObserver
    // -------------------------------------------------------------------------

    function setupResizeObserver() {
        if (typeof ResizeObserver === 'undefined') { return; }
        var ro = new ResizeObserver(function () {
            resizeCanvas();
            render();
        });
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
            } else if (namespace === 'Room.Info' || namespace === 'Room') {
                updateMap();
            }
        },
    });

}());
