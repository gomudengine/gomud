
class GraphRenderer {
  constructor({ container, onRoomClick }) {
    this.rooms       = new Map();
    this.onRoomClick = onRoomClick;

    // build SVG & viewport
    this.svg = document.createElementNS("http://www.w3.org/2000/svg","svg");
    this.svg.setAttribute("width","100%");
    this.svg.setAttribute("height","100%");
    this.svg.style.cursor = "grab";
    container.appendChild(this.svg);

    this.viewport = document.createElementNS(this.svg.namespaceURI,"g");
    this.svg.appendChild(this.viewport);

    // pan/zoom state
    this.tx = 0; this.ty = 0; this.k = 1;
    this._initPanZoom();
    this._initClick();
  }

  // ——— pan & zoom (unchanged) ———
  _initPanZoom() {
    this.svg.addEventListener("pointerdown", e => {
      this.svg.setPointerCapture(e.pointerId);
      this.isPanning      = true;
      this.panStart       = { x: e.clientX, y: e.clientY };
      this.transStart     = { tx: this.tx, ty: this.ty };
      this.svg.style.cursor = "grabbing";
    });
    this.svg.addEventListener("pointermove", e => {
      if (!this.isPanning) return;
      const dx = (e.clientX - this.panStart.x) / this.k;
      const dy = (e.clientY - this.panStart.y) / this.k;
      this.tx = this.transStart.tx + dx;
      this.ty = this.transStart.ty + dy;
      this._updateTransform();
    });
    this.svg.addEventListener("pointerup", e => {
      this.svg.releasePointerCapture(e.pointerId);
      this.isPanning        = false;
      this.svg.style.cursor = "grab";
    });
    this.svg.addEventListener("wheel", e => {
      e.preventDefault();
      const zoomSpeed = 0.002;
      const scale = Math.exp(-e.deltaY * zoomSpeed);
      const { width, height } = this.svg.getBoundingClientRect();
      const pt = this.svg.createSVGPoint();
      pt.x = width/2; pt.y = height/2;
      const before = pt.matrixTransform(this.viewport.getCTM().inverse());
      this.k *= scale;
      this._updateTransform();
      const after = pt.matrixTransform(this.viewport.getCTM().inverse());
      this.tx += (after.x - before.x);
      this.ty += (after.y - before.y);
      this._updateTransform();
    }, { passive: false });
  }

  _initClick() {
    this.viewport.addEventListener("click", e => {
      const hit = e.target.closest("[data-room-id]");
      if (hit) this.onRoomClick(+hit.dataset.roomId);
    });
  }

  _updateTransform() {
    this.viewport.setAttribute(
      "transform",
      `translate(${this.tx},${this.ty}) scale(${this.k})`
    );
  }

  _recomputeBounds() {
    if (!this.rooms.size) return;
    const xs = [], ys = [];
    for (let r of this.rooms.values()) {
      xs.push(r.x); ys.push(r.y);
    }
    const minX = Math.min(...xs), maxX = Math.max(...xs);
    const minY = Math.min(...ys), maxY = Math.max(...ys);
    const w = maxX - minX + 2, h = maxY - minY + 2;
    this.svg.setAttribute("viewBox",
      `${minX-1} ${minY-1} ${w} ${h}`
    );
  }

  _draw() {
    while (this.viewport.firstChild)
      this.viewport.removeChild(this.viewport.firstChild);

    // edges
    for (let r of this.rooms.values()) {
      for (let dir in r.Exits) {
        const toId = Number(r.Exits[dir].RoomId);
        const tgt  = this.rooms.get(toId);
        if (!tgt) continue;
        const line = document.createElementNS(this.svg.namespaceURI,"line");
        line.setAttribute("x1", r.x);
        line.setAttribute("y1", r.y);
        line.setAttribute("x2", tgt.x);
        line.setAttribute("y2", tgt.y);
        line.setAttribute("stroke","#444");
        line.setAttribute("stroke-width","0.02");
        this.viewport.appendChild(line);
      }
    }

    // rooms
    for (let r of this.rooms.values()) {
      const g = document.createElementNS(this.svg.namespaceURI,"g");
      g.setAttribute("data-room-id", r.RoomId);

      const rect = document.createElementNS(this.svg.namespaceURI,"rect");
      rect.setAttribute("x", r.x - 0.4);
      rect.setAttribute("y", r.y - 0.4);
      rect.setAttribute("width","0.8");
      rect.setAttribute("height","0.8");
      rect.setAttribute("fill","white");
      rect.setAttribute("stroke","#333");
      rect.style.cursor = "pointer";
      g.appendChild(rect);

      const text = document.createElementNS(this.svg.namespaceURI,"text");
      text.setAttribute("x", r.x);
      text.setAttribute("y", r.y + 0.1);
      text.setAttribute("text-anchor","middle");
      text.setAttribute("font-size","0.4");
      text.textContent = r.RoomId;
      text.style.pointerEvents = "none";
      g.appendChild(text);

      this.viewport.appendChild(g);
    }
  }

  // ——— collision‐free full‐graph layout ———
  _layoutAll() {
    if (!this.rooms.size) return;
  
    // clear coords & init occupied set
    const occupied = new Set();
    for (let r of this.rooms.values()) {
      r.x = undefined; r.y = undefined;
    }
  
    // seed root at (0,0)
    const root = this.rooms.values().next().value;
    root.x = 0; root.y = 0;
    occupied.add("0,0");
  
    const visited = new Set([ root.RoomId ]);
    const queue   = [ root.RoomId ];
  
    // direction vectors scaled
    const S = this.cellSize;
    const DIR = {
      north: [ 0, -S ],
      south: [ 0,  S ],
      east:  [  S, 0 ],
      west:  [ -S, 0 ]
    };
  
    // BFS assign
    while (queue.length) {
      const id   = queue.shift();
      const room = this.rooms.get(id);
  
      for (let dir in room.Exits) {
        const toId = Number(room.Exits[dir].RoomId);
        if (visited.has(toId)) continue;
        const nbr = this.rooms.get(toId);
        if (!nbr) continue;
  
        // start one cell over in the desired direction
        let [dx, dy] = DIR[dir] || [S, 0];
        let cx = room.x + dx,
            cy = room.y + dy;
  
        // walk further if occupied
        while (occupied.has(`${cx},${cy}`)) {
          cx += dx;
          cy += dy;
        }
  
        nbr.x = cx; nbr.y = cy;
        occupied.add(`${cx},${cy}`);
  
        visited.add(toId);
        queue.push(toId);
      }
    }
  
    // fallback for any unconnected rooms
    const placed = Array.from(visited).map(i => this.rooms.get(i));
    let maxX = Math.max(...placed.map(r => r.x));
    const minY = Math.min(...placed.map(r => r.y));
  
    for (let [id, r] of this.rooms) {
      if (visited.has(id)) continue;
      let cx = maxX + S, cy = minY;
      while (occupied.has(`${cx},${cy}`)) {
        cx += S;
      }
      r.x = cx; r.y = cy;
      occupied.add(`${cx},${cy}`);
      maxX = cx;
    }
  }
  

  // ——— Public API ———

  setRooms(arr) {
    this.rooms.clear();
    for (let r of arr) {
      r.RoomId = Number(r.RoomId);
      this.rooms.set(r.RoomId, r);
    }
    this._layoutAll();
    this._recomputeBounds();
    this._draw();
  }

  addRoom(room) {
    room.RoomId = Number(room.RoomId);
    this.rooms.set(room.RoomId, room);

    this._layoutAll();
    this._recomputeBounds();
    this._draw();
  }

  centerOn(roomId) {
    const r = this.rooms.get(Number(roomId));
    if (!r) return;
    const [ vx, vy, vw, vh ] = this.svg
      .getAttribute("viewBox")
      .split(/[\s,]+/).map(Number);
    const cx = vx + vw/2, cy = vy + vh/2;
    this.tx = cx/this.k - r.x;
    this.ty = cy/this.k - r.y;
    this._updateTransform();
  }
}
