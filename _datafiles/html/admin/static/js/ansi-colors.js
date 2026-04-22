// AnsiColors — global utility for converting ANSI 256-color codes to CSS hex
// and rendering color-swatch HTML. Loaded by _header.html and available on all
// admin pages.

const AnsiColors = (() => {
    'use strict';

    const _base16 = [
        '#000000','#800000','#008000','#808000','#000080','#800080','#008080','#c0c0c0',
        '#808080','#ff0000','#00ff00','#ffff00','#0000ff','#ff00ff','#00ffff','#ffffff',
    ];
    const _cubeSteps = [0, 95, 135, 175, 215, 255];

    function toHex(n) {
        n = parseInt(n, 10);
        if (n < 16) return _base16[n] || '#888';
        if (n > 231) {
            const v = 8 + 10 * (n - 232);
            const h = v.toString(16).padStart(2, '0');
            return '#' + h + h + h;
        }
        n -= 16;
        const r = _cubeSteps[Math.floor(n / 36)];
        const g = _cubeSteps[Math.floor((n % 36) / 6)];
        const b = _cubeSteps[n % 6];
        const hx = v => v.toString(16).padStart(2, '0');
        return '#' + hx(r) + hx(g) + hx(b);
    }

    function swatchHtml(colors, maxColors) {
        const list = maxColors ? colors.slice(0, maxColors) : colors;
        return list.map(c => '<span style="background:' + toHex(c) + '"></span>').join('');
    }

    function previewTextHtml(text, colors) {
        if (!colors || colors.length === 0) return text;
        let html = '';
        let dir = 1, pos = 0;
        for (const ch of text) {
            const hex = toHex(colors[pos]);
            html += '<span style="color:' + hex + '">' + ch.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;') + '</span>';
            if (ch !== ' ') {
                if (pos === colors.length - 1) dir = -1;
                else if (pos === 0) dir = 1;
                pos += dir;
            }
        }
        return html;
    }

    return { toHex, swatchHtml, previewTextHtml };
})();
