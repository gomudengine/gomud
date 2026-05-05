/**
 * mapper-tool-room-drag.js -- Two-phase drag system for moving rooms.
 *
 * Phase 1 ("armed"): On mousedown over a room the tool records the start
 *   position but does NOT become the active tool yet. This lets a plain
 *   click fall through to selection / context menu handling.
 *
 * Phase 2 ("promoted"): Once the cursor moves more than 4 px from the
 *   armed position the tool promotes itself via MapperTools.activate(),
 *   builds the drag group (single room or current selection), computes
 *   exit constraints, and begins tracking pixel + grid deltas.
 *
 * On mouseup the group is committed (if the target cells are unoccupied)
 *   and the tool deactivates back to pan.
 *
 * Overlay rendering draws origin markers, snap indicators (color-coded by
 * droppability), predicted connection lines, and the rooms themselves at
 * the current pixel offset -- in both 2D and 3D projections.
 */
/* jshint esversion: 11, browser: true */
/* globals MapperTools, MapperState, MapperRender, MapperEvents,
   BASE_STEP_2D, ROOM_SIZE_2D, SYMBOL_FONT_SIZE_2D, ROOM_BORDER_WIDTH_2D,
   CONNECTION_WIDTH_2D,
   DRAG_ORIGIN_MARKER, DRAG_SNAP_BLOCKED, DRAG_SNAP_BROKEN, DRAG_SNAP_CLEAN,
   DRAG_CONSTRAINT_BROKEN, DRAG_CONSTRAINT_OK, DRAG_GHOST_BROKEN_FILL,
   SELECTED_ROOM_COLOR, SELECTED_ROOM_TEXT_COLOR, SYMBOL_TEXT_COLOR,
   buildDragConstraints, isExitConstraintSatisfied, darkenColor */
'use strict';

(function() {

    // =====================================================================
    //  Armed state -- held before threshold promotion
    // =====================================================================

    var armed = false;
    var armedRoomId = null;
    var armedPxX = 0;
    var armedPxY = 0;

    function clearArmed() {
        armed = false;
        armedRoomId = null;
        armedPxX = 0;
        armedPxY = 0;
    }

    // =====================================================================
    //  Tool definition
    // =====================================================================

    var tool = {
        name: 'room-drag',
        cursor: 'move',

        // -----------------------------------------------------------------
        //  Lifecycle
        // -----------------------------------------------------------------

        onActivate: function() {},

        onDeactivate: function() {
            var rd = MapperState.roomDrag;
            rd.active = false;
            rd.anchorId = null;
            rd.group = new Map();
            rd.brokenExits = [];
            rd.allConstraints = [];
            clearArmed();
        },

        // -----------------------------------------------------------------
        //  Intercept -- called by init BEFORE normal mousedown dispatch.
        //  Returns true when the event is claimed (arms the drag).
        // -----------------------------------------------------------------

        interceptMouseDown: function(e, cx, cy, roomId) {
            if (roomId === null) return false;
            if (e.shiftKey) return false; // shift-click toggles selection
            armed = true;
            armedRoomId = roomId;
            armedPxX = e.clientX;
            armedPxY = e.clientY;
            return true;
        },

        // -----------------------------------------------------------------
        //  Mouse handlers (active after promotion)
        // -----------------------------------------------------------------

        onMouseDown: function() { return false; },

        onMouseMove: function(e, cx, cy, roomId, gridCell) {
            var rd = MapperState.roomDrag;

            // --- Threshold promotion: armed but not yet active ---
            if (armed && !rd.active) {
                var movedPx = Math.abs(e.clientX - armedPxX) + Math.abs(e.clientY - armedPxY);
                if (movedPx > 4) {
                    var anchorRoom = MapperState.data.rooms.get(armedRoomId);
                    if (anchorRoom && anchorRoom.HasCoordinates) {
                        // Build group from selection (if anchor is selected) or just the anchor
                        var groupIds = MapperState.selected.has(armedRoomId) && MapperState.selected.size > 1
                            ? new Set(MapperState.selected) : new Set([armedRoomId]);
                        var groupMap = new Map();
                        var groupSet = new Set();
                        groupIds.forEach(function(rid) {
                            var r = MapperState.data.rooms.get(rid);
                            if (r && r.HasCoordinates) {
                                groupMap.set(rid, { startGx: r.MapX, startGy: r.MapY });
                                groupSet.add(rid);
                            }
                        });

                        // Pre-compute exit constraints for the entire group
                        var allC = [];
                        groupSet.forEach(function(rid) {
                            allC = allC.concat(buildDragConstraints(rid, groupSet));
                        });

                        var ap2 = MapperRender.gridToCanvas2d(anchorRoom.MapX, anchorRoom.MapY);
                        var anchorP = { px: ap2.px, py: ap2.py };

                        rd.active = true;
                        rd.anchorId = armedRoomId;
                        rd.group = groupMap;
                        rd.deltaGx = 0;
                        rd.deltaGy = 0;
                        rd.pixelDx = 0;
                        rd.pixelDy = 0;
                        rd.anchorCanvasPx = anchorP.px;
                        rd.anchorCanvasPy = anchorP.py;
                        rd.droppable = true;
                        rd.brokenExits = [];
                        rd.allConstraints = allC;

                        MapperTools.activate('room-drag');
                        MapperRender.scheduleRender();
                        return;
                    }
                }
                return; // still below threshold
            }

            // --- Active drag tracking ---
            if (!rd.active) return;

            rd.pixelDx = cx - rd.anchorCanvasPx;
            rd.pixelDy = cy - rd.anchorCanvasPy;

            var gc = MapperRender.canvasToGrid(cx, cy);
            var anchorStart = rd.group.get(rd.anchorId);
            var newDx = gc.gx - anchorStart.startGx;
            var newDy = gc.gy - anchorStart.startGy;

            // Only recheck constraints when the snapped grid delta changes
            if (newDx !== rd.deltaGx || newDy !== rd.deltaGy) {
                rd.deltaGx = newDx;
                rd.deltaGy = newDy;

                // Collision check: ensure no group member lands on an occupied cell
                var canDrop = true;
                rd.group.forEach(function(start, rid) {
                    var pGx = start.startGx + newDx;
                    var pGy = start.startGy + newDy;
                    var room = MapperState.data.rooms.get(rid);
                    var rz = room ? room.MapZ : MapperRender.currentZ();
                    var coordKey = pGx + ',' + pGy + ',' + rz;
                    var occupant = MapperState.data.roomsByCoord.get(coordKey);
                    if (occupant !== undefined && !rd.group.has(occupant)) canDrop = false;
                });
                rd.droppable = canDrop;

                // Identify broken exit constraints for the overlay warning lines
                if (canDrop && (newDx !== 0 || newDy !== 0)) {
                    var broken = [];
                    rd.allConstraints.forEach(function(c) {
                        if (!isExitConstraintSatisfied(c, c.ownerGx + newDx, c.ownerGy + newDy)) broken.push(c);
                    });
                    rd.brokenExits = broken;
                } else {
                    rd.brokenExits = [];
                }
            }
            MapperRender.scheduleRender();
        },

        onMouseUp: function(e, cx, cy) {
            var rd = MapperState.roomDrag;

            // Armed but never promoted -- clear and let click fall through
            if (armed && !rd.active) {
                clearArmed();
                return;
            }

            if (!rd.active) return;

            var wasDroppable = rd.droppable;
            var dGx = rd.deltaGx;
            var dGy = rd.deltaGy;
            var groupCopy = new Map(rd.group);

            rd.active = false;
            rd.anchorId = null;
            rd.group = new Map();
            rd.brokenExits = [];
            rd.allConstraints = [];
            clearArmed();

            MapperEvents.emit('pan:suppressClick');

            if (dGx !== 0 || dGy !== 0) {
                if (wasDroppable) {
                    MapperState.applyGroupMove(groupCopy, dGx, dGy);
                } else {
                    MapperState.showToast('Cannot move — room collision detected');
                }
            }

            MapperTools.activate('pan');
            MapperRender.render();
        },

        onKeyDown: function() {},

        renderOverlay2d: function(ctx, rs) {
            var rd = MapperState.roomDrag;
            if (!rd.active) return;

            var scaledSize = rs.scaledSize;
            var half = rs.half;
            var hasBroken = rd.brokenExits.length > 0;
            var hasMoved = rd.deltaGx !== 0 || rd.deltaGy !== 0;

            // Origin markers: faint dashed outline where each room started
            if (hasMoved) {
                rd.group.forEach(function(start) {
                    var origP = rs.gridToCanvas2d(start.startGx, start.startGy);
                    ctx.strokeStyle = DRAG_ORIGIN_MARKER;
                    ctx.lineWidth = Math.max(1, 1 * rs.zoomScale);
                    ctx.setLineDash([Math.max(2, 3 * rs.zoomScale), Math.max(2, 3 * rs.zoomScale)]);
                    ctx.strokeRect(origP.px - half, origP.py - half, scaledSize, scaledSize);
                    ctx.setLineDash([]);
                });
            }

            // Snap indicators: dashed outline at the target cell
            // Color: red = blocked, orange = broken exits, blue = clean drop
            if (hasMoved) {
                rd.group.forEach(function(start) {
                    var snapP = rs.gridToCanvas2d(start.startGx + rd.deltaGx, start.startGy + rd.deltaGy);
                    var snapColor = !rd.droppable ? DRAG_SNAP_BLOCKED :
                                    hasBroken ? DRAG_SNAP_BROKEN : DRAG_SNAP_CLEAN;
                    ctx.strokeStyle = snapColor;
                    ctx.lineWidth = Math.max(1, 1.5 * rs.zoomScale);
                    ctx.setLineDash([Math.max(2, 4 * rs.zoomScale), Math.max(2, 4 * rs.zoomScale)]);
                    ctx.strokeRect(snapP.px - half, snapP.py - half, scaledSize, scaledSize);
                    ctx.setLineDash([]);
                });
            }

            // Predicted connection lines from snap positions to neighbors
            if (rd.droppable && hasMoved) {
                ctx.lineWidth = CONNECTION_WIDTH_2D * rs.zoomScale * 0.7;
                ctx.lineCap = 'round';
                ctx.setLineDash([Math.max(2, 3 * rs.zoomScale), Math.max(2, 3 * rs.zoomScale)]);
                rd.allConstraints.forEach(function(c) {
                    var isBroken = !isExitConstraintSatisfied(c, c.ownerGx + rd.deltaGx, c.ownerGy + rd.deltaGy);
                    var fromP = rs.gridToCanvas2d(c.ownerGx + rd.deltaGx, c.ownerGy + rd.deltaGy);
                    var toP = rs.gridToCanvas2d(c.refX, c.refY);
                    ctx.strokeStyle = isBroken ? DRAG_CONSTRAINT_BROKEN : DRAG_CONSTRAINT_OK;
                    ctx.beginPath();
                    ctx.moveTo(fromP.px, fromP.py);
                    ctx.lineTo(toP.px, toP.py);
                    ctx.stroke();
                });
                ctx.setLineDash([]);
            }

            // Drag-ghost rooms: solid tiles following the cursor at pixel offset
            rd.group.forEach(function(start, rid) {
                var origP = rs.gridToCanvas2d(start.startGx, start.startGy);
                var dragP = { px: origP.px + rd.pixelDx, py: origP.py + rd.pixelDy };
                var dragRoom = MapperState.data.rooms.get(rid);
                if (!dragRoom) return;

                var blocked = !rd.droppable && hasMoved;
                if (blocked) { ctx.globalAlpha = 0.5; }
                rs.drawRoom2d(dragP, dragRoom, rid);
                if (blocked) { ctx.globalAlpha = 1.0; }

                if (rd.droppable && hasBroken) {
                    ctx.fillStyle = DRAG_GHOST_BROKEN_FILL;
                    ctx.fillRect(dragP.px - half, dragP.py - half, scaledSize, scaledSize);
                }
            });
        }
    };

    MapperTools.register(tool);

})();
