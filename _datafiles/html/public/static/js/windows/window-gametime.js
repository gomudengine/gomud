/**
 * window-gametime.js
 *
 * Virtual window: Time and Date.
 * Displays the current in-game time and date, with an animated sky showing
 * a sun or moon moving across the horizon based on the time of day.
 *
 * Responds to GMCP namespaces:
 *   Gametime  - full gametime update
 *
 * Reads: Client.GMCPStructs.Gametime
 */

'use strict';

(function() {

    injectStyles(`
        #gametime-panel {
            height: 100%;
            display: flex;
            flex-direction: column;
            overflow: hidden;
            background: #000;
            user-select: none;
        }

        #gametime-sky {
            position: relative;
            flex: 1;
            overflow: hidden;
            min-height: 0;
        }

        #gametime-sky canvas {
            display: block;
            width: 100%;
            height: 100%;
        }

        #gametime-labels {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 2px 8px;
            background: rgba(0,0,0,0.55);
            flex-shrink: 0;
        }

        #gametime-time {
            font-family: monospace;
            font-size: 0.85em;
            color: #e8d8a0;
            letter-spacing: 0.04em;
        }

        #gametime-date {
            font-family: monospace;
            font-size: 0.72em;
            color: #8899aa;
        }

        #gametime-tooltip {
            position: fixed;
            z-index: 99999;
            pointer-events: none;
            background: #0d2e28;
            border: 1px solid #1c6b60;
            border-radius: 6px;
            box-shadow: 0 4px 16px rgba(0,0,0,0.7);
            padding: 8px 10px;
            min-width: 140px;
            display: none;
            font-family: monospace;
            font-size: 0.78em;
            color: #dffbd1;
            white-space: nowrap;
        }
    `);

    // -----------------------------------------------------------------------
    // DOM factory
    // -----------------------------------------------------------------------
    function createDOM() {
        const panel = document.createElement('div');
        panel.id = 'gametime-panel';

        const sky = document.createElement('div');
        sky.id = 'gametime-sky';

        const canvas = document.createElement('canvas');
        canvas.id = 'gametime-canvas';
        sky.appendChild(canvas);

        const labels = document.createElement('div');
        labels.id = 'gametime-labels';
        labels.innerHTML =
            '<span id="gametime-time">--:-- --</span>' +
            '<span id="gametime-date">Day 1, Year 1</span>';

        panel.appendChild(sky);
        panel.appendChild(labels);

        document.body.appendChild(panel);
        return panel;
    }

    // -----------------------------------------------------------------------
    // Tooltip
    // -----------------------------------------------------------------------
    let tooltip      = null;
    let tooltipAnchor = null;

    function ensureTooltip() {
        if (tooltip) { return; }
        tooltip = document.createElement('div');
        tooltip.id = 'gametime-tooltip';
        document.body.appendChild(tooltip);
    }

    function showTooltip(anchorEl) {
        ensureTooltip();
        const data = Client.GMCPStructs.Gametime;
        if (!data) { return; }

        const hour24     = data.hour24     || 0;
        const minute     = data.minute     || 0;
        const dayStart   = data.day_start  || 6;
        const nightStart = data.night_start || 22;
        const night      = data.night;

        // Calculate hours remaining until the next transition.
        // All arithmetic is in fractional hours.
        const timeNow = hour24 + minute / 60;
        let hoursLeft;
        if (night) {
            // Time until dayStart (may wrap midnight)
            hoursLeft = dayStart - timeNow;
            if (hoursLeft <= 0) { hoursLeft += 24; }
        } else {
            // Time until nightStart
            hoursLeft = nightStart - timeNow;
            if (hoursLeft <= 0) { hoursLeft += 24; }
        }

        const h   = Math.floor(hoursLeft);
        const m   = Math.round((hoursLeft - h) * 60);
        const hStr = h > 0 ? h + 'h ' : '';
        const mStr = m + 'm';
        const label = night
            ? 'Sunrise in ' + hStr + mStr
            : 'Sunset in '  + hStr + mStr;

        tooltip.textContent  = label;
        tooltip.style.display = 'block';

        const rect = anchorEl.getBoundingClientRect();
        const ttW  = tooltip.offsetWidth;
        const vw   = window.innerWidth;
        let left = rect.right + 8;
        if (left + ttW > vw - 8) { left = rect.left - ttW - 8; }
        left = Math.max(8, left);
        tooltip.style.left = left + 'px';
        tooltip.style.top  = (rect.top + 4) + 'px';
    }

    function hideTooltip() {
        if (tooltip) { tooltip.style.display = 'none'; }
        tooltipAnchor = null;
    }

    function attachTooltip(el) {
        el.addEventListener('mouseenter', function() {
            tooltipAnchor = el;
            showTooltip(el);
        });
        el.addEventListener('mouseleave', hideTooltip);
    }

    // -----------------------------------------------------------------------
    // VirtualWindow instance
    // -----------------------------------------------------------------------
    const win = new VirtualWindow('Gametime', {
        dock:          'left',
        defaultDocked: true,
        dockedHeight:  100,
        factory() {
            const el = createDOM();
            attachTooltip(el);
            return {
                title:      'Time & Date',
                mount:      el,
                background: '#1c3a5e',
                border:     1,
                x:          0,
                y:          0,
                width:      300,
                height:     100,
                header:     20,
                bottom:     60,
            };
        },
    });

    // -----------------------------------------------------------------------
    // Sky rendering
    // -----------------------------------------------------------------------

    // Returns the 0..1 arc position of the celestial body.
    // Sun travels dayStart → nightStart; moon travels nightStart → dayStart.
    function celestialPosition(hour24, minute, dayStart, nightStart) {
        const timeNow = hour24 + minute / 60;

        const dayDuration = nightStart > dayStart
            ? nightStart - dayStart
            : (24 - dayStart) + nightStart;
        const nightDuration = 24 - dayDuration;

        const isDay = !(timeNow >= nightStart || timeNow < dayStart);

        if (isDay) {
            let elapsed = timeNow - dayStart;
            if (elapsed < 0) { elapsed += 24; }
            return elapsed / dayDuration;
        } else {
            let elapsed = timeNow - nightStart;
            if (elapsed < 0) { elapsed += 24; }
            return elapsed / nightDuration;
        }
    }

    function drawSky(data) {
        const canvas = document.getElementById('gametime-canvas');
        if (!canvas) { return; }

        const parent = canvas.parentElement;
        const w = parent.clientWidth  || 300;
        const h = parent.clientHeight || 60;

        if (canvas.width !== w || canvas.height !== h) {
            canvas.width  = w;
            canvas.height = h;
        }

        const ctx        = canvas.getContext('2d');
        const night      = data.night;
        const hour24     = data.hour24;
        const minute     = data.minute;
        const dayStart   = data.day_start   || 6;
        const nightStart = data.night_start || 22;

        // --- background ---
        const grad = ctx.createLinearGradient(0, 0, 0, h);
        if (night) {
            grad.addColorStop(0, '#020510');
            grad.addColorStop(1, '#0a0e20');
        } else {
            grad.addColorStop(0, '#1a6fbf');
            grad.addColorStop(1, '#87ceeb');
        }
        ctx.fillStyle = grad;
        ctx.fillRect(0, 0, w, h);

        // --- star field (night only) ---
        if (night) {
            ctx.fillStyle = 'rgba(255,255,255,0.7)';
            const starData = [
                [0.05,0.15],[0.12,0.55],[0.20,0.10],[0.28,0.70],[0.35,0.25],
                [0.42,0.80],[0.50,0.05],[0.57,0.60],[0.65,0.30],[0.73,0.85],
                [0.80,0.18],[0.88,0.65],[0.95,0.40],[0.08,0.90],[0.60,0.50],
                [0.33,0.45],[0.75,0.10],[0.18,0.38],[0.90,0.75],[0.48,0.92],
            ];
            starData.forEach(function(s) {
                const sx = s[0] * w;
                const sy = s[1] * h * 0.85;
                ctx.beginPath();
                ctx.arc(sx, sy, 0.8, 0, Math.PI * 2);
                ctx.fill();
            });
        }

        // --- celestial body ---
        const bodyRadius = Math.max(8, Math.min(14, h * 0.28));
        const pos    = celestialPosition(hour24, minute, dayStart, nightStart);
        const margin = bodyRadius * 1.5;
        const bodyX  = margin + pos * (w - margin * 2);
        const yBottom = h * 0.92;
        const yTop    = h * 0.18;
        const bodyY   = yBottom - (yBottom - yTop) * 4 * pos * (1 - pos);

        if (night) {
            // Full moon
            ctx.save();
            ctx.beginPath();
            ctx.arc(bodyX, bodyY, bodyRadius, 0, Math.PI * 2);
            ctx.fillStyle  = '#d8dde8';
            ctx.shadowColor = '#c0ccdd';
            ctx.shadowBlur  = 10;
            ctx.fill();
            ctx.restore();
        } else {
            // Sun — soft glow then disc
            const glow = ctx.createRadialGradient(bodyX, bodyY, bodyRadius * 0.5, bodyX, bodyY, bodyRadius * 2.2);
            glow.addColorStop(0, 'rgba(255,240,100,0.35)');
            glow.addColorStop(1, 'rgba(255,240,100,0)');
            ctx.beginPath();
            ctx.arc(bodyX, bodyY, bodyRadius * 2.2, 0, Math.PI * 2);
            ctx.fillStyle = glow;
            ctx.fill();

            ctx.save();
            ctx.beginPath();
            ctx.arc(bodyX, bodyY, bodyRadius, 0, Math.PI * 2);
            ctx.fillStyle   = '#fff5a0';
            ctx.shadowColor = '#fff5a0';
            ctx.shadowBlur  = 12;
            ctx.fill();
            ctx.restore();
        }
    }

    // -----------------------------------------------------------------------
    // Update logic
    // -----------------------------------------------------------------------
    function update() {
        const data = Client.GMCPStructs.Gametime;
        if (!data) { return; }

        win.open();
        if (!win.isOpen()) { return; }

        // Labels
        const timeEl = document.getElementById('gametime-time');
        const dateEl = document.getElementById('gametime-date');
        if (timeEl) {
            const min = String(data.minute).padStart(2, '0');
            timeEl.textContent = data.hour + ':' + min + ' ' + data.ampm;
        }
        if (dateEl) {
            const monthName = data.month_name || ('Month ' + data.month);
            const zodiac    = data.zodiac     || '';
            dateEl.textContent = 'Day ' + data.day + ' of ' + monthName +
                ', year ' + data.year + ' (the ' + zodiac + ').';
        }

        drawSky(data);

        if (tooltipAnchor && tooltip && tooltip.style.display === 'block') {
            showTooltip(tooltipAnchor);
        }
    }

    // -----------------------------------------------------------------------
    // Registration
    // -----------------------------------------------------------------------
    VirtualWindows.register({
        window:       win,
        gmcpHandlers: ['Gametime'],
        onGMCP() { update(); },
    });

})();
