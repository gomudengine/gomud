/* jshint esversion: 11, browser: true */
/* globals MapperState, MapperTools, MapperEvents,
   ZOOM_STEP, ZOOM_MIN, ZOOM_MAX, CENTER_EASE_DURATION,
   ROOM_SIZE_2D, ROOM_GAP_2D, BASE_STEP_2D, CONNECTION_WIDTH_2D, ROOM_BORDER_WIDTH_2D, SYMBOL_FONT_SIZE_2D, MAP_BG_2D, ROOM_BORDER_COLOR_2D,
   CONNECTION_COLOR, ABNORMAL_CONNECTION_COLOR, SELECTED_ROOM_COLOR, SELECTED_ROOM_TEXT_COLOR, SYMBOL_TEXT_COLOR,
   exitDelta, isDirectionalExit, darkenColor, smoothstep, isExitConstraintSatisfied */

/**
 * MapperRender — canvas drawing engine for the admin map editor (2D only).
 *
 * Public API (returned IIFE object):
 *   Setup          — setCanvas, initResizeObserver, resizeCanvas
 *   Rendering      — render, getRenderState
 *   Coord helpers  — gridToCanvas2d, canvasToGrid2d, canvasToGrid
 *   Hit testing    — roomAtPoint, roomAtPoint2d, currentZ, gridCellOccupied
 *   Drawing prims  — drawRoom2d, drawLineBadge2d
 */
var MapperRender = (function() {
    'use strict';

    // --- Canvas References ---

    var canvas = null;
    var ctx = null;

    function setCanvas(c) {
        canvas = c;
        ctx = c.getContext('2d');
    }

    // --- Coordinate Transforms: 2D ---

    function gridToCanvas2d(gx, gy) {
        var cam = MapperState.camera;
        var midX = Math.floor(canvas.width / 2);
        var midY = Math.floor(canvas.height / 2);
        var step = BASE_STEP_2D * cam.spacingScale2d * cam.zoomScale;
        return {
            px: midX + (gx - cam.cameraX - cam.panOffsetX) * step,
            py: midY + (gy - cam.cameraY - cam.panOffsetY) * step
        };
    }

    function canvasToGrid2d(cx, cy) {
        var cam = MapperState.camera;
        var midX = Math.floor(canvas.width / 2);
        var midY = Math.floor(canvas.height / 2);
        var step = BASE_STEP_2D * cam.spacingScale2d * cam.zoomScale;
        return {
            gx: Math.round((cx - midX) / step + cam.cameraX + cam.panOffsetX),
            gy: Math.round((cy - midY) / step + cam.cameraY + cam.panOffsetY)
        };
    }

    function canvasToGrid(cx, cy) {
        return canvasToGrid2d(cx, cy);
    }

    // --- Hit Testing ---

    function roomAtPoint(cx, cy) {
        return roomAtPoint2d(cx, cy);
    }

    function roomAtPoint2d(cx, cy) {
        var cam = MapperState.camera;
        var half = (ROOM_SIZE_2D * cam.zoomScale) / 2;
        var found = null;
        MapperState.data.rooms.forEach(function(room, id) {
            if (found !== null) return;
            if (!room.HasCoordinates || room.MapZ !== cam.activeZ2d) return;
            var p = gridToCanvas2d(room.MapX, room.MapY);
            if (cx >= p.px - half && cx <= p.px + half && cy >= p.py - half && cy <= p.py + half) {
                found = id;
            }
        });
        return found;
    }

    // --- Current Z / Grid Occupancy ---

    function currentZ() {
        return MapperState.camera.activeZ2d;
    }

    function gridCellOccupied(gx, gy, gz) {
        return MapperState.data.roomsByCoord.has(gx + ',' + gy + ',' + gz);
    }

    // --- 2D Drawing Primitives ---

    /** Draws a small badge (secret "?" or key icon) at the midpoint of a connection line. */
    function drawLineBadge2d(mx, my, type) {
        var cam = MapperState.camera;
        var sz = Math.max(7, Math.round(CONNECTION_WIDTH_2D * cam.zoomScale * 2.5));
        var half = sz / 2;

        ctx.save();
        ctx.fillStyle = MAP_BG_2D;
        ctx.fillRect(mx - half, my - half, sz, sz);

        if (type === 'secret') {
            ctx.fillStyle = '#d4a843';
            ctx.font = 'bold ' + Math.round(sz * 0.85) + 'px monospace';
            ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
            ctx.fillText('?', mx, my);
        } else {
            var kc = '#9ab0d4', lw = Math.max(1, sz * 0.14);
            ctx.strokeStyle = kc; ctx.fillStyle = kc;
            ctx.lineWidth = lw; ctx.lineCap = 'round';
            var bowR = sz * 0.22, bowCx = mx - sz * 0.14;
            ctx.beginPath(); ctx.arc(bowCx, my, bowR, 0, Math.PI * 2); ctx.stroke();
            var shaftX1 = bowCx + bowR, shaftX2 = mx + half * 0.82;
            ctx.beginPath(); ctx.moveTo(shaftX1, my); ctx.lineTo(shaftX2, my); ctx.stroke();
            var toothH = sz * 0.18;
            var t1x = shaftX1 + (shaftX2 - shaftX1) * 0.45;
            var t2x = shaftX1 + (shaftX2 - shaftX1) * 0.72;
            ctx.beginPath();
            ctx.moveTo(t1x, my); ctx.lineTo(t1x, my + toothH);
            ctx.moveTo(t2x, my); ctx.lineTo(t2x, my + toothH);
            ctx.stroke();
        }
        ctx.restore();
    }

    /** Draws a single room tile in 2D: filled square, border, symbol, and Z-arrow indicators. */
    function drawRoom2d(p, room, id) {
        var cam = MapperState.camera;
        var scaledSize = ROOM_SIZE_2D * cam.zoomScale;
        var scaledBorder = ROOM_BORDER_WIDTH_2D * cam.zoomScale;
        var scaledFont = SYMBOL_FONT_SIZE_2D * cam.zoomScale;
        var half = scaledSize / 2;

        var isSelected = MapperState.selected.has(id);
        var fill = isSelected ? SELECTED_ROOM_COLOR : room._color;
        var rx = p.px - half, ry = p.py - half;

        ctx.fillStyle = fill;
        ctx.fillRect(rx, ry, scaledSize, scaledSize);

        if (!isSelected && room.HasScript) {
            // Glow: multiple progressively thicker, more transparent strokes
            var glowColor = '#d4a843';
            ctx.save();
            ctx.shadowColor = glowColor;
            ctx.shadowBlur = Math.max(4, scaledSize * 0.4) * cam.zoomScale;
            ctx.strokeStyle = glowColor;
            ctx.lineWidth = scaledBorder * 2;
            ctx.strokeRect(rx, ry, scaledSize, scaledSize);
            ctx.restore();
            ctx.strokeStyle = glowColor;
            ctx.lineWidth = scaledBorder * 1.5;
            ctx.strokeRect(rx, ry, scaledSize, scaledSize);
        } else {
            ctx.strokeStyle = ROOM_BORDER_COLOR_2D;
            ctx.lineWidth = scaledBorder;
            ctx.strokeRect(rx, ry, scaledSize, scaledSize);
        }

        ctx.fillStyle = isSelected ? SELECTED_ROOM_TEXT_COLOR : (room._symbolColor || SYMBOL_TEXT_COLOR);
        ctx.font = 'bold ' + scaledFont + 'px monospace';
        ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
        ctx.fillText(room._symbol || '•', p.px, p.py);

        var hasUp = false, hasDown = false;
        if (room.Exits) {
            for (var dir in room.Exits) {
                var delta = exitDelta(dir, room);
                if (delta && delta[2] > 0) hasUp = true;
                if (delta && delta[2] < 0) hasDown = true;
            }
        }
        if (hasUp || hasDown) {
            var arrowSize = Math.max(10, scaledSize * 0.56);
            var margin = Math.max(2, scaledSize * 0.1);
            ctx.font = 'bold ' + arrowSize + 'px monospace';
            ctx.fillStyle = '#ff00ff';
            if (hasDown) {
                ctx.textAlign = 'left'; ctx.textBaseline = 'alphabetic';
                ctx.fillText('▾', rx + margin, ry + scaledSize - margin);
            }
            if (hasUp) {
                ctx.textAlign = 'right'; ctx.textBaseline = 'top';
                ctx.fillText('▴', rx + scaledSize - margin, ry - margin);
            }
        }
    }

    // --- Tool Overlay Dispatch ---

    function getRenderState() {
        var cam = MapperState.camera;
        return {
            ctx: ctx, canvas: canvas, zoomScale: cam.zoomScale,
            activeTab: cam.activeTab, activeZ2d: cam.activeZ2d,
            selectedRoomIds: MapperState.selected,
            hoveredRoomId: MapperState.hoveredRoomId,
            hoveredGridCell: MapperState.hoveredGridCell,
            gridToCanvas2d: gridToCanvas2d,
            canvasToGrid2d: canvasToGrid2d,
            canvasToGrid: canvasToGrid,
            gridCellOccupied: gridCellOccupied,
            drawRoom2d: drawRoom2d,
            scaledSize: ROOM_SIZE_2D * cam.zoomScale,
            scaledBorder: ROOM_BORDER_WIDTH_2D * cam.zoomScale,
            scaledFont: SYMBOL_FONT_SIZE_2D * cam.zoomScale,
            half: (ROOM_SIZE_2D * cam.zoomScale) / 2
        };
    }

    function renderToolOverlays2d() {
        var rs = getRenderState();
        var tools = MapperTools.all();
        for (var name in tools) {
            if (tools[name] && typeof tools[name].renderOverlay2d === 'function') {
                tools[name].renderOverlay2d(ctx, rs);
            }
        }
    }

    // --- 2D Core Renderer ---

    function render2d() {
        var cam = MapperState.camera;
        var data = MapperState.data;

        ctx.clearRect(0, 0, canvas.width, canvas.height);
        ctx.fillStyle = MAP_BG_2D;
        ctx.fillRect(0, 0, canvas.width, canvas.height);

        var rooms = data.rooms;
        if (rooms.size === 0) { renderToolOverlays2d(); return; }

        ctx.lineCap = 'round';
        var drawnEdges = new Set();
        var abnormalEdges = [];

        // Pass 1: normal directional edges
        ctx.strokeStyle = CONNECTION_COLOR;
        ctx.lineWidth = CONNECTION_WIDTH_2D * cam.zoomScale;
        rooms.forEach(function(room, id) {
            if (!room.HasCoordinates || room.MapZ !== cam.activeZ2d) return;
            if (!room.Exits) return;
            for (var dir in room.Exits) {
                var ex = room.Exits[dir];
                var dest = rooms.get(ex.RoomId);
                if (!dest || !dest.HasCoordinates) continue;
                var key = Math.min(id, ex.RoomId) + '-' + Math.max(id, ex.RoomId) + ':' + dir;
                if (drawnEdges.has(key)) continue;
                drawnEdges.add(key);
                var delta = exitDelta(dir, room);
                var directional = isDirectionalExit(dir);
                if (!directional || !delta) {
                    abnormalEdges.push({ room: room, dest: dest, dir: dir, ex: ex });
                    continue;
                }
                if (delta[2] !== 0) continue;
                if (dest.MapZ !== cam.activeZ2d) continue;
                var pA = gridToCanvas2d(room.MapX, room.MapY);
                var pB = gridToCanvas2d(dest.MapX, dest.MapY);
                ctx.beginPath(); ctx.moveTo(pA.px, pA.py); ctx.lineTo(pB.px, pB.py); ctx.stroke();
                if (ex.Secret || ex.HasLock) {
                    drawLineBadge2d((pA.px + pB.px) / 2, (pA.py + pB.py) / 2, ex.Secret ? 'secret' : 'key');
                }
            }
        });

        // Pass 2: abnormal edges (yellow dotted arcs)
        if (abnormalEdges.length > 0) {
            ctx.strokeStyle = ABNORMAL_CONNECTION_COLOR;
            ctx.lineWidth = Math.max(1, CONNECTION_WIDTH_2D * cam.zoomScale * 0.7);
            ctx.setLineDash([Math.max(3, 8 * cam.zoomScale), Math.max(4, 10 * cam.zoomScale)]);
            abnormalEdges.forEach(function(ae) {
                var pA = gridToCanvas2d(ae.room.MapX, ae.room.MapY);
                var pB = gridToCanvas2d(ae.dest.MapX, ae.dest.MapY);
                var mx = (pA.px + pB.px) / 2, my = (pA.py + pB.py) / 2;
                var dx = pB.px - pA.px, dy = pB.py - pA.py;
                var dist = Math.sqrt(dx * dx + dy * dy);
                var bulge = Math.max(15, dist * 0.25);
                var cpx = mx + (-dy / dist) * bulge;
                var cpy = my + (dx / dist) * bulge;
                ctx.beginPath(); ctx.moveTo(pA.px, pA.py);
                ctx.quadraticCurveTo(cpx, cpy, pB.px, pB.py); ctx.stroke();
            });
            ctx.setLineDash([]);
        }

        // Rooms
        var dragGroup = MapperState.roomDrag;
        var dragGroupSet = dragGroup.active ? new Set(dragGroup.group.keys()) : new Set();

        rooms.forEach(function(room, id) {
            if (!room.HasCoordinates || room.MapZ !== cam.activeZ2d) return;
            if (dragGroupSet.has(id)) return;
            drawRoom2d(gridToCanvas2d(room.MapX, room.MapY), room, id);
        });

        renderToolOverlays2d();
    }

    // --- Render Dispatch ---

    function render() {
        render2d();
    }

    // --- Canvas Resize ---

    var viewport = null;

    function resizeCanvas() {
        if (!canvas) return;
        if (!viewport) viewport = canvas.parentElement;
        canvas.width = (viewport ? viewport.clientWidth : window.innerWidth) || 1;
        canvas.height = (viewport ? viewport.clientHeight : window.innerHeight) || 1;
    }

    var resizeObserver = null;
    function initResizeObserver() {
        if (typeof ResizeObserver === 'undefined') return;
        if (!canvas) return;
        viewport = canvas.parentElement;
        if (!viewport) return;
        resizeObserver = new ResizeObserver(function() { resizeCanvas(); render(); });
        resizeObserver.observe(viewport);
    }

    return {
        setCanvas: setCanvas,
        initResizeObserver: initResizeObserver,
        resizeCanvas: resizeCanvas,
        render: render,
        getRenderState: getRenderState,
        gridToCanvas2d: gridToCanvas2d,
        canvasToGrid2d: canvasToGrid2d,
        canvasToGrid: canvasToGrid,
        roomAtPoint: roomAtPoint,
        roomAtPoint2d: roomAtPoint2d,
        currentZ: currentZ,
        gridCellOccupied: gridCellOccupied,
        drawRoom2d: drawRoom2d,
        drawLineBadge2d: drawLineBadge2d
    };

})();
