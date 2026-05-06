/**
 * mapper-tool-pan.js -- Default tool for panning the map canvas.
 *
 * Dragging on empty space pans the camera. When no drag is active the tool
 * renders a dashed "ghost cell" outline with a "+" on whichever empty grid
 * cell the cursor hovers, hinting that clicking will open the context menu
 * for room creation.
 *
 * Pan math differs between 2D (simple pixel-to-grid ratio) and 3D (inverse
 * isometric projection), so both paths are handled in onMouseMove.
 */
/* jshint esversion: 11, browser: true */
/* globals MapperTools, MapperState, MapperRender, MapperEvents,
   BASE_STEP_2D, ROOM_SIZE_2D, SYMBOL_FONT_SIZE_2D,
   GHOST_CELL_BORDER, GHOST_CELL_FILL, GHOST_CELL_SYMBOL,
   ZOOM_STEP, ZOOM_MIN, ZOOM_MAX */
'use strict';

(function() {

    var tool = {
        name: 'pan',

        // -----------------------------------------------------------------
        //  Lifecycle
        // -----------------------------------------------------------------

        onActivate: function() {},
        onDeactivate: function() {
            var cam = MapperState.camera;
            cam.dragActive = false;
        },

        // -----------------------------------------------------------------
        //  Mouse handlers
        // -----------------------------------------------------------------

        onMouseDown: function(e, cx, cy, roomId, gridCell) {
            if (roomId !== null) return false;
            if (e.shiftKey) return false; // shift+empty starts a selection rect

            var cam = MapperState.camera;
            cam.dragActive = true;
            cam.dragStartPxX = e.clientX;
            cam.dragStartPxY = e.clientY;
            cam.dragStartPanX = cam.panOffsetX;
            cam.dragStartPanY = cam.panOffsetY;
            return true; // claim
        },

        onMouseMove: function(e, cx, cy, roomId, gridCell) {
            var cam = MapperState.camera;

            if (cam.dragActive) {
                var step = BASE_STEP_2D * cam.spacingScale2d * cam.zoomScale;
                cam.panOffsetX = cam.dragStartPanX - (e.clientX - cam.dragStartPxX) / step;
                cam.panOffsetY = cam.dragStartPanY - (e.clientY - cam.dragStartPxY) / step;
                MapperRender.scheduleRender();
                return;
            }
        },

        onMouseUp: function(e, cx, cy) {
            var cam = MapperState.camera;
            if (!cam.dragActive) return;
            var dx = e.clientX - cam.dragStartPxX;
            var dy = e.clientY - cam.dragStartPxY;
            cam.dragActive = false;
            // Suppress the click event when the user clearly intended a drag
            if (Math.abs(dx) > 4 || Math.abs(dy) > 4) {
                MapperEvents.emit('pan:suppressClick');
            }
        },

        onKeyDown: function() {},

        renderOverlay2d: function(ctx, rs) {
            if (MapperState.roomDrag.active) return;
            if (MapperState.quickBuildMode.active) return;
            if (MapperState.exitDrawMode.active) return;

            var hoveredGridCell = rs.hoveredGridCell;
            if (!hoveredGridCell) return;
            if (rs.gridCellOccupied(hoveredGridCell.gx, hoveredGridCell.gy, rs.activeZ2d)) return;

            var gp = rs.gridToCanvas2d(hoveredGridCell.gx, hoveredGridCell.gy);
            var scaledSize = rs.scaledSize;
            var ghalf = scaledSize / 2;

            ctx.strokeStyle = GHOST_CELL_BORDER;
            ctx.lineWidth = Math.max(1, 1.5 * rs.zoomScale);
            ctx.setLineDash([Math.max(2, 4 * rs.zoomScale), Math.max(2, 4 * rs.zoomScale)]);
            ctx.strokeRect(gp.px - ghalf, gp.py - ghalf, scaledSize, scaledSize);
            ctx.setLineDash([]);

            ctx.fillStyle = GHOST_CELL_FILL;
            ctx.fillRect(gp.px - ghalf, gp.py - ghalf, scaledSize, scaledSize);

            ctx.fillStyle = GHOST_CELL_SYMBOL;
            ctx.font = 'bold ' + Math.max(10, rs.scaledFont * 0.8) + 'px monospace';
            ctx.textAlign = 'center';
            ctx.textBaseline = 'middle';
            ctx.fillText('+', gp.px, gp.py);
        }
    };

    MapperTools.register(tool);

})();
