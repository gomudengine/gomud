/* global Client, RoomGridSVG, VirtualWindow, VirtualWindows, injectStyles */

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

class RoomGridSVG {
  constructor(selector, options = {}) {
      // ── Configurable options & defaults ───────────────────────────────
      this.cellSize = options.cellSize || 100;
      this.cellMargin = options.cellMargin || 20;
      this.spacing = this.cellSize + this.cellMargin;
      this.zoomStep = options.zoomStep || 1.2;
      this.zoomLevel = options.initialZoom || 1;
      this.onRoomClick = options.onRoomClick || (() => {});
      this.zoomButtonSize = options.zoomButtonSize || 25;
      this.controlsMargin = options.controlsMargin || 10;
      this.roomEdgeColor = options.roomEdgeColor || "#1c6b60";
      this.visitingColor = options.visitingColor || "#c20000";
      // ── Internal state ────────────────────────────────────────────────
      // rooms: Map<RoomId, { room, group, defaultColor }>
      this.rooms = new Map();
      this.drawnEdges = new Set(); // to avoid dup lines
      this.currentCenterId = null; // for highlight

      // ── Build container & SVG ─────────────────────────────────────────
      this.container = document.querySelector(selector);
      this.container.style.position = 'relative';

      this.svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
      this.svg.setAttribute('preserveAspectRatio', 'xMidYMid meet');
      this.svg.style.width = '100%';
      this.svg.style.height = '100%';
      this.container.appendChild(this.svg);

      // Connections under rooms:
      this.connectionsGroup = document.createElementNS(this.svg.namespaceURI, 'g');
      this.svg.appendChild(this.connectionsGroup);
      // Rooms on top:
      this.roomsGroup = document.createElementNS(this.svg.namespaceURI, 'g');
      this.svg.appendChild(this.roomsGroup);

      // Default tiny viewBox until rooms exist:
      this.svg.setAttribute('viewBox', '0 0 1 1');

      // ── HTML overlay zoom controls ────────────────────────────────────
      this._createHTMLControls();
  }

  // ── Public API ───────────────────────────────────────────────────────

  /**
   * Add or update a room.
   * - Pre-adds any Exits given as {RoomId,x,y,…}
   * - If room already exists, updates its position, color, text, & redraws edges.
   */
  addRoom(room) {
      const id = room.RoomId;

      // 1) Pre-add exit-defined rooms
      if (Array.isArray(room.Exits)) {
          room.Exits.forEach(e => {
              if (e && typeof e === 'object' && e.RoomId != null) {

                  if (this.rooms.has(e.RoomId)) return;

                  this.addRoom({
                      RoomId: e.RoomId,
                      Text: e.Text != null ? e.Text : String(e.RoomId),
                      x: e.x,
                      y: e.y,
                      Exits: Array.isArray(e.Exits) ? e.Exits : []
                  });
              }
          });
      }

      // prepare defaults
      const defaultColor = room.Color || '#fff';
      const displayText = room.Text != null ?
          room.Text :
          String(room.RoomId);

      // 2) UPDATE existing
      if (this.rooms.has(id)) {
          const entry = this.rooms.get(id);
          // update stored data
          entry.room.x = room.x;
          entry.room.y = room.y;
          entry.room.Exits = Array.isArray(room.Exits) ? room.Exits : [];
          entry.room.Color = room.Color;
          entry.room.Text = room.Text;
          entry.defaultColor = defaultColor;

          // move & recolor rect
          const rect = this.svg.querySelector(`rect[data-room-rect="${id}"]`);
          rect.setAttribute('x', room.x * this.spacing);
          rect.setAttribute('y', room.y * this.spacing);
          if (this.currentCenterId === id) {
              rect.setAttribute('fill', this.visitingColor);
          } else {
              rect.setAttribute('fill', defaultColor);
          }

          // move & update label
          const txtEl = this.svg.querySelector(`g[data-room-id="${id}"] text`);
          txtEl.setAttribute('x', room.x * this.spacing + this.cellSize / 2);
          txtEl.setAttribute('y', room.y * this.spacing + this.cellSize / 2 + 5);
          txtEl.textContent = displayText;

          // redraw any new edges
          this._drawEdgesForRoom(id);

          // refresh bounds & view
          this._updateBounds();
          this._applyZoom();
          return;
      }

      // 3) NEW room → draw group
      const g = document.createElementNS(this.svg.namespaceURI, 'g');
      g.setAttribute('data-room-id', id);

      // square
      const rect = document.createElementNS(this.svg.namespaceURI, 'rect');
      rect.setAttribute('width', this.cellSize);
      rect.setAttribute('height', this.cellSize);
      rect.setAttribute('x', room.x * this.spacing);
      rect.setAttribute('y', room.y * this.spacing);
      rect.setAttribute('stroke', this.roomEdgeColor);
      rect.setAttribute('stroke-width', '4');
      rect.setAttribute('rx', this.cellSize / 10); // corner radius X
      rect.setAttribute('ry', this.cellSize / 10); // corner radius Y    
      rect.setAttribute('data-room-rect', id);
      rect.setAttribute('fill', defaultColor);
      rect.style.cursor = 'pointer';
      rect.addEventListener('click', () => this.onRoomClick(room));
      g.appendChild(rect);

      // label
      const label = document.createElementNS(this.svg.namespaceURI, 'text');
      label.setAttribute('x', room.x * this.spacing + this.cellSize / 2);
      label.setAttribute('y', room.y * this.spacing + this.cellSize / 2 + 5);
      label.setAttribute('text-anchor', 'middle');
      label.setAttribute('font-size', this.cellSize * 0.3);
      label.textContent = displayText;
      g.appendChild(label);

      this.roomsGroup.appendChild(g);
      this.rooms.set(id, {
          room,
          group: g,
          defaultColor
      });

      // draw edges for this new room
      this._drawEdgesForRoom(id);

      // refresh bounds & view
      this._updateBounds();
      this._applyZoom();
  }

  /**
   * Bulk‐set rooms (wipes existing).
   */
  setRooms(arr) {
      this.reset();
      arr.forEach(r => this.addRoom(r));
  }

  /**
   * Clear everything.
   */
  reset() {
      this.rooms.clear();
      this.drawnEdges.clear();
      this.currentCenterId = null;
      this.zoomLevel = 1;
      this.svg.setAttribute('viewBox', '0 0 1 1');
      this.roomsGroup.innerHTML = '';
      this.connectionsGroup.innerHTML = '';
  }

  /**
   * Center & highlight a room.  Previous one reverts to its default color.
   */
  centerOnRoom(id) {
      const entry = this.rooms.get(id);
      if (!entry) return;

      // un-highlight previous
      if (this.currentCenterId != null) {
          const prevRect = this.svg.querySelector(
              `rect[data-room-rect="${this.currentCenterId}"]`
          );
          if (prevRect) {
              const prevEntry = this.rooms.get(this.currentCenterId);
              prevRect.setAttribute('fill', prevEntry.defaultColor);
          }
      }

      // compute new view center
      this.center = {
          x: entry.room.x * this.spacing + this.cellSize / 2,
          y: entry.room.y * this.spacing + this.cellSize / 2
      };
      this._applyZoom();

      // highlight new
      const newRect = this.svg.querySelector(
          `rect[data-room-rect="${id}"]`
      );
      if (newRect) newRect.setAttribute('fill', this.visitingColor);

      this.currentCenterId = id;
  }

  zoomIn() {
      this.zoomLevel *= this.zoomStep;
      this._applyZoom();
  }
  zoomOut() {
      this.zoomLevel /= this.zoomStep;
      this._applyZoom();
  }

  drawConnection(a, b) {
      if (!this.rooms.has(a) || !this.rooms.has(b)) return;
      this._drawEdge(a, b);
      this._applyZoom();
  }

  // ── Private draw helpers ───────────────────────────────────────────────

  _createHTMLControls() {
      const div = document.createElement('div');
      div.style.cssText = `
    position:absolute;
    top:${this.controlsMargin}px;
    right:${this.controlsMargin}px;
    display:flex; gap:5px;
  `;
      const mk = (lbl, cb) => {
          const b = document.createElement('button');
          b.textContent = lbl;
          b.style.cssText = `
      width:${this.zoomButtonSize}px;
      height:${this.zoomButtonSize}px;
      font-size:${this.zoomButtonSize*0.6}px;
      line-height:1;
    `;
          b.addEventListener('click', cb);
          return b;
      };
      div.append(mk('−', () => this.zoomOut()), mk('+', () => this.zoomIn()));
      this.container.appendChild(div);
  }

  _drawEdgesForRoom(id) {
      const me = this.rooms.get(id)
          .room;
      const exits = Array.isArray(me.Exits) ? me.Exits : [];

      // draw its own exits
      exits.forEach(e => {
          const to = (typeof e === 'object') ? e.RoomId : e;
          if (this.rooms.has(to)) this._drawEdge(id, to);
      });

      // draw others’ exits back to it
      this.rooms.forEach(({
          room
      }, otherId) => {
          if (otherId === id) return;
          const oe = Array.isArray(room.Exits) ? room.Exits : [];
          if (oe.some(x => ((typeof x === 'object') ? x.RoomId : x) === id)) {
              this._drawEdge(otherId, id);
          }
      });
  }

  _drawEdge(a, b) {
      const key = a < b ? `${a}-${b}` : `${b}-${a}`;
      if (this.drawnEdges.has(key)) return;
      this.drawnEdges.add(key);

      const ra = this.rooms.get(a)
          .room;
      const rb = this.rooms.get(b)
          .room;
      const x1 = ra.x * this.spacing + this.cellSize / 2;
      const y1 = ra.y * this.spacing + this.cellSize / 2;
      const x2 = rb.x * this.spacing + this.cellSize / 2;
      const y2 = rb.y * this.spacing + this.cellSize / 2;

      const line = document.createElementNS(this.svg.namespaceURI, 'line');
      line.setAttribute('x1', x1);
      line.setAttribute('y1', y1);
      line.setAttribute('x2', x2);
      line.setAttribute('y2', y2);
      line.setAttribute('stroke', this.roomEdgeColor);
      line.setAttribute('stroke-width', '20');
      this.connectionsGroup.appendChild(line);
  }

  _updateBounds() {
      if (!this.rooms.size) {
          this.bounds = {
              minX: 0,
              maxX: 0,
              minY: 0,
              maxY: 0
          };
      } else {
          const xs = [...this.rooms.values()].map(e => e.room.x);
          const ys = [...this.rooms.values()].map(e => e.room.y);
          this.bounds = {
              minX: Math.min(...xs),
              maxX: Math.max(...xs),
              minY: Math.min(...ys),
              maxY: Math.max(...ys)
          };
      }
      this.worldWidth = (this.bounds.maxX - this.bounds.minX + 1) * this.spacing;
      this.worldHeight = (this.bounds.maxY - this.bounds.minY + 1) * this.spacing;

      if (!this.center && this.rooms.size) {
          this.center = {
              x: this.bounds.minX * this.spacing + this.worldWidth / 2,
              y: this.bounds.minY * this.spacing + this.worldHeight / 2
          };
      }
  }

  _applyZoom() {
      const hw = this.worldWidth / (2 * this.zoomLevel);
      const hh = this.worldHeight / (2 * this.zoomLevel);
      const x0 = (this.center ? this.center.x : this.worldWidth / 2) - hw;
      const y0 = (this.center ? this.center.y : this.worldHeight / 2) - hh;
      this.svg.setAttribute('viewBox', `${x0} ${y0} ${hw*2} ${hh*2}`);
  }
}



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

