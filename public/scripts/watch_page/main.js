(async () => {
    const id = location.pathname.split('/').pop();
    // Request HLS stream from server
    const r = await fetch(`/stream/${id}`);
    const { m3u8 } = await r.json();

    // Add cache-buster to avoid reusing an old playlist
    const src = m3u8 + (m3u8.includes('?') ? '&' : '?') + 't=' + Date.now();

    const video = document.getElementById('video');
    if (video.canPlayType('application/vnd.apple.mpegurl')) {
    // Native HLS (Safari / iOS)
    video.src = src;
    video.play();
    } else if (Hls.isSupported()) {
    // hls.js for browsers without native HLS
    const hls = new Hls({
        manifestLoadingRetryDelay: 1000,
        manifestLoadingMaxRetry: 10,
        fragLoadingMaxRetry: 10
    });
    hls.loadSource(src);
    hls.attachMedia(video);
    hls.on(Hls.Events.MANIFEST_PARSED, () => video.play());
    } else {
    alert('Your browser does not support HLS.');
    }
})();