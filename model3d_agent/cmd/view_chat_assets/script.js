function copyRaw() {
    const pre = document.getElementById('raw-json');
    const text = pre?.textContent || '';
    navigator.clipboard.writeText(text).then(() => {
        const btn = document.getElementById('copy-btn');
        if (btn) { btn.textContent = 'Copied!'; setTimeout(() => btn.textContent = 'Copy raw JSON', 1200); }
    });
}
function toggleRaw() {
    const el = document.getElementById('raw-wrap');
    if (!el) return;
    el.style.display = (el.style.display === 'none' ? 'block' : 'none');
}