/* Fullscreen and viewport negotiation are progressive enhancements. */
let gameSocket = null, requestedViewport = '', viewportTimer;
function desiredViewport(w, h) {
    const target = w > 1000 ? 46 : (w > h ? 27 : 30);
    const tile = Math.max(target, w / 32, h / 32);
    return { rows: Math.max(8, Math.min(32, Math.floor(h / tile))), columns: Math.max(8, Math.min(32, Math.floor(w / tile))) };
}
function syncGameViewport() {
    const page = document.querySelector('.play-page');
    if (!page) return;
    page.style.setProperty('--play-height', `${window.visualViewport?.height || window.innerHeight}px`);
    resizeCanvas();
    clearTimeout(viewportTimer);
    viewportTimer = setTimeout(() => {
        const screen = document.getElementById('screen');
        if (!gameSocket || !screen || !document.querySelector('.play-page')) return;
        const rect = screen.getBoundingClientRect();
        const viewport = desiredViewport(rect.width, rect.height);
        const key = `${viewport.rows},${viewport.columns}`;
        if (key === requestedViewport) return;
        requestedViewport = key;
        gameSocket.send(JSON.stringify({eventname:'viewport', ...viewport}));
    }, 120);
    updateFullscreenControl();
}
async function toggleGameFullscreen() {
    const target = document.documentElement;
    try {
        if (document.fullscreenElement) await document.exitFullscreen();
        else if (target.requestFullscreen) await target.requestFullscreen();
        else { showGameNotice('Fullscreen is unavailable here. Add Bloopworld to your Home Screen for more room.'); return; }
    } catch (_) { showGameNotice('Fullscreen could not open. You can keep playing in this view.'); }
    syncGameViewport();
}
function updateFullscreenControl() {
    const button = document.getElementById('fullscreen-toggle');
    if (!button) return;
    const active = !!document.fullscreenElement;
    button.setAttribute('aria-label',active?'Exit fullscreen':'Enter fullscreen');
    button.setAttribute('aria-pressed',String(active));
    button.title=active?'Exit fullscreen':'Fullscreen';
}
function showGameNotice(text) {
    const target=document.getElementById('bottom_text');
    if (target) { target.textContent=text; target.hidden=false; }
}
document.addEventListener('htmx:wsAfterMessage', e => {
    if (e.target.id !== 'controls') return;
    gameSocket=e.detail.socketWrapper;
    const notice=document.getElementById("bottom_text");
    if (notice) notice.hidden=!notice.textContent.trim();
    if (!requestedViewport) syncGameViewport();
});
document.addEventListener('htmx:wsOpen', e => {
    if (e.target.id !== 'controls') return;
    requestedViewport='';
    const status=document.getElementById('connection-state'); if(status) status.textContent='';
});
document.addEventListener('htmx:wsClose', e => {
    if (e.target.id !== 'controls') return;
    gameSocket=null; requestedViewport='';
    const status=document.getElementById('connection-state'); if(status) status.textContent='Connection lost. Reconnecting…';
});
document.addEventListener('htmx:afterSwap', syncGameViewport);
document.addEventListener('htmx:oobAfterSwap', e => {
    if (e.target.id === 'modal_background') document.querySelector('#modal_menu a')?.focus();
});
window.addEventListener('resize',syncGameViewport);
window.visualViewport?.addEventListener('resize',syncGameViewport);
document.addEventListener('fullscreenchange',syncGameViewport);
document.addEventListener('keydown', e => {
    if (!document.querySelector('.play-page')) return;
    if (e.target.matches('input:not([type=hidden]),textarea,select,[contenteditable=true]')) { e.stopPropagation(); return; }
    const modal = document.getElementById('modal_menu');
    if (modal && e.key === 'Tab') {
        const links = [...modal.querySelectorAll('a,button,[tabindex="0"]')];
        if (links.length) {
            const index = links.indexOf(document.activeElement);
            const next = (index + (e.shiftKey ? -1 : 1) + links.length) % links.length;
            e.preventDefault(); links[next].focus();
        }
    }
    if (['ArrowUp','ArrowDown','ArrowLeft','ArrowRight',' '].includes(e.key)) e.preventDefault();
}, true);

function openGameMenu() {
    gameSocket?.send(JSON.stringify({eventname:document.querySelector('.modal_bg')?'menuOff':'menuOn'}));
}
