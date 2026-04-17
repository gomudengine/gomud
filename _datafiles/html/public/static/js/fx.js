const FX = {

    // -----------------------------------------------------------------------
    // Confetti — coloured squares fall from above.
    // duration: seconds the animation runs.
    // -----------------------------------------------------------------------
    Confetti(duration = 1.5) {
        const canvas = document.createElement('canvas');
        const ctx = canvas.getContext('2d', { willReadFrequently: true });
        document.body.appendChild(canvas);
        canvas.style.cssText = 'position:fixed;top:0;left:0;pointer-events:none;z-index:99999;';

        function resize() { canvas.width = window.innerWidth; canvas.height = window.innerHeight; }
        resize();
        window.addEventListener('resize', resize);

        const colors = ['#ff0', '#f0f', '#0ff', '#0f0', '#f00', '#00f'];
        const pieces = Array.from({ length: 400 }, () => ({
            x:             Math.random() * canvas.width,
            y:             Math.random() * -canvas.height,
            size:          Math.random() * 8 + 4,
            color:         colors[Math.floor(Math.random() * colors.length)],
            velocityX:     (Math.random() - 0.5) * 4,
            velocityY:     Math.random() * 3 + 2,
            rotation:      Math.random() * 360,
            rotationSpeed: (Math.random() - 0.5) * 10,
        }));

        const start = performance.now();
        const durationMs = duration * 1000;

        (function animate(now) {
            ctx.clearRect(0, 0, canvas.width, canvas.height);
            pieces.forEach(p => {
                p.velocityY += 0.15;
                p.velocityX += (Math.random() - 0.5) * 0.05;
                p.x += p.velocityX;
                p.y += p.velocityY;
                p.rotation += p.rotationSpeed;
                ctx.save();
                ctx.translate(p.x, p.y);
                ctx.rotate(p.rotation * Math.PI / 180);
                ctx.fillStyle = p.color;
                ctx.fillRect(-p.size / 2, -p.size / 2, p.size, p.size);
                ctx.restore();
            });
            if (now - start < durationMs) {
                requestAnimationFrame(animate);
            } else {
                window.removeEventListener('resize', resize);
                document.body.removeChild(canvas);
            }
        })(performance.now());
    },

    // -----------------------------------------------------------------------
    // Flash — a brief full-screen colour overlay that fades out.
    // color:    CSS colour string, e.g. 'rgba(255,0,0,0.45)' for damage,
    //           'rgba(80,255,80,0.35)' for healing, 'rgba(255,220,0,0.4)' for
    //           level-up.
    // duration: fade-out time in seconds.
    // -----------------------------------------------------------------------
    Flash(color = 'rgba(255,0,0,0.45)', duration = 0.5) {
        const el = document.createElement('div');
        el.style.cssText = [
            'position:fixed', 'inset:0', 'pointer-events:none', 'z-index:99999',
            'background:' + color,
            'transition:opacity ' + duration + 's ease-out',
            'opacity:1',
        ].join(';');
        document.body.appendChild(el);
        requestAnimationFrame(() => requestAnimationFrame(() => { el.style.opacity = '0'; }));
        setTimeout(() => { if (el.parentNode) { el.parentNode.removeChild(el); } }, duration * 1000 + 50);
    },

    // -----------------------------------------------------------------------
    // Shake — briefly shakes the #main-container (or the whole body).
    // intensity: max pixel offset.
    // duration:  total shake time in seconds.
    // -----------------------------------------------------------------------
    Shake(intensity = 8, duration = 0.4) {
        const target = document.getElementById('main-container') || document.body;
        const start  = performance.now();
        const durationMs = duration * 1000;
        const original   = target.style.transform;

        (function animate(now) {
            const elapsed  = now - start;
            const progress = elapsed / durationMs;
            if (progress >= 1) {
                target.style.transform = original;
                return;
            }
            const decay = 1 - progress;
            const dx = (Math.random() - 0.5) * 2 * intensity * decay;
            const dy = (Math.random() - 0.5) * 2 * intensity * decay;
            target.style.transform = 'translate(' + dx + 'px,' + dy + 'px)';
            requestAnimationFrame(animate);
        })(performance.now());
    },

    // -----------------------------------------------------------------------
    // Sparks — golden particles burst upward from the bottom of the screen.
    // Useful for kill blows, treasure, or level-up moments.
    // count:    number of spark particles.
    // duration: seconds before the canvas is removed.
    // -----------------------------------------------------------------------
    Sparks(count = 120, duration = 1.2) {
        const canvas = document.createElement('canvas');
        const ctx    = canvas.getContext('2d', { willReadFrequently: true });
        document.body.appendChild(canvas);
        canvas.style.cssText = 'position:fixed;top:0;left:0;pointer-events:none;z-index:99999;';

        function resize() { canvas.width = window.innerWidth; canvas.height = window.innerHeight; }
        resize();
        window.addEventListener('resize', resize);

        const colors = ['#ffe066', '#ffd700', '#ffaa00', '#fff4a0', '#ff8800'];
        const sparks = Array.from({ length: count }, () => {
            const angle = (Math.random() * 120 - 60) * Math.PI / 180; // -60..+60 deg from straight up
            const speed = Math.random() * 10 + 4;
            return {
                x:     Math.random() * canvas.width,
                y:     canvas.height + Math.random() * 20,
                vx:    Math.sin(angle) * speed,
                vy:    -Math.cos(angle) * speed,
                size:  Math.random() * 3 + 1,
                color: colors[Math.floor(Math.random() * colors.length)],
                life:  Math.random() * 0.5 + 0.5,
            };
        });

        const start      = performance.now();
        const durationMs = duration * 1000;

        (function animate(now) {
            const elapsed = now - start;
            ctx.clearRect(0, 0, canvas.width, canvas.height);
            const t = elapsed / durationMs;
            sparks.forEach(s => {
                s.vy += 0.3;
                s.x  += s.vx;
                s.y  += s.vy;
                const alpha = Math.max(0, s.life - t) / s.life;
                ctx.globalAlpha = alpha;
                ctx.beginPath();
                ctx.arc(s.x, s.y, s.size, 0, Math.PI * 2);
                ctx.fillStyle = s.color;
                ctx.fill();
            });
            ctx.globalAlpha = 1;
            if (elapsed < durationMs) {
                requestAnimationFrame(animate);
            } else {
                window.removeEventListener('resize', resize);
                document.body.removeChild(canvas);
            }
        })(performance.now());
    },

    // -----------------------------------------------------------------------
    // Rain — streaks fall from the top of the screen.
    // color:    streak colour, e.g. '#66aaff' for rain, '#aaffaa' for acid.
    // duration: seconds the effect runs.
    // -----------------------------------------------------------------------
    Rain(color = '#66aaff', duration = 2.0) {
        const canvas = document.createElement('canvas');
        const ctx    = canvas.getContext('2d', { willReadFrequently: true });
        document.body.appendChild(canvas);
        canvas.style.cssText = 'position:fixed;top:0;left:0;pointer-events:none;z-index:99999;';

        function resize() { canvas.width = window.innerWidth; canvas.height = window.innerHeight; }
        resize();
        window.addEventListener('resize', resize);

        const drops = Array.from({ length: 200 }, () => ({
            x:      Math.random() * window.innerWidth,
            y:      Math.random() * -window.innerHeight,
            length: Math.random() * 20 + 10,
            speed:  Math.random() * 8 + 6,
            alpha:  Math.random() * 0.5 + 0.3,
        }));

        const start      = performance.now();
        const durationMs = duration * 1000;

        (function animate(now) {
            const elapsed = now - start;
            ctx.clearRect(0, 0, canvas.width, canvas.height);
            drops.forEach(d => {
                d.y += d.speed;
                if (d.y > canvas.height) {
                    d.y = -d.length;
                    d.x = Math.random() * canvas.width;
                }
                ctx.beginPath();
                ctx.moveTo(d.x, d.y);
                ctx.lineTo(d.x - 1, d.y + d.length);
                ctx.strokeStyle = color;
                ctx.globalAlpha = d.alpha;
                ctx.lineWidth   = 1;
                ctx.stroke();
            });
            ctx.globalAlpha = 1;
            if (elapsed < durationMs) {
                requestAnimationFrame(animate);
            } else {
                window.removeEventListener('resize', resize);
                document.body.removeChild(canvas);
            }
        })(performance.now());
    },

    // -----------------------------------------------------------------------
    // Ripple — concentric rings expand outward from the center of the screen.
    // Useful for magic casts, area-of-effect spells, or tremors.
    // color:    ring colour.
    // rings:    how many rings to emit.
    // duration: seconds for each ring to fully expand and fade.
    // -----------------------------------------------------------------------
    Ripple(color = '#3ad4b8', rings = 4, duration = 1.0) {
        const canvas = document.createElement('canvas');
        const ctx    = canvas.getContext('2d', { willReadFrequently: true });
        document.body.appendChild(canvas);
        canvas.style.cssText = 'position:fixed;top:0;left:0;pointer-events:none;z-index:99999;';

        function resize() { canvas.width = window.innerWidth; canvas.height = window.innerHeight; }
        resize();
        window.addEventListener('resize', resize);

        const cx = canvas.width  / 2;
        const cy = canvas.height / 2;
        const maxRadius = Math.sqrt(cx * cx + cy * cy);
        const durationMs = duration * 1000;
        const stagger    = durationMs / rings;

        const ripples = Array.from({ length: rings }, (_, i) => ({
            startTime: performance.now() + i * stagger,
        }));

        let allDone = false;

        (function animate(now) {
            ctx.clearRect(0, 0, canvas.width, canvas.height);
            allDone = true;
            ripples.forEach(r => {
                if (now < r.startTime) { allDone = false; return; }
                const t = (now - r.startTime) / durationMs;
                if (t >= 1) { return; }
                allDone = false;
                const radius = t * maxRadius;
                const alpha  = 1 - t;
                ctx.beginPath();
                ctx.arc(cx, cy, radius, 0, Math.PI * 2);
                ctx.strokeStyle = color;
                ctx.globalAlpha = alpha;
                ctx.lineWidth   = 2 + (1 - t) * 3;
                ctx.stroke();
            });
            ctx.globalAlpha = 1;
            if (!allDone) {
                requestAnimationFrame(animate);
            } else {
                window.removeEventListener('resize', resize);
                document.body.removeChild(canvas);
            }
        })(performance.now());
    },

};
