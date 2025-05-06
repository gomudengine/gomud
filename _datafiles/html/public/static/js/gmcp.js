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
