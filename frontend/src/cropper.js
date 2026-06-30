(() => {
    const cropInputs = document.querySelectorAll('input[type="file"][data-crop-kind]');
    if (!cropInputs.length) return;

    const modal = document.createElement('div');
    modal.className = 'cropper-backdrop';
    modal.innerHTML = `
        <div class="cropper-window">
            <div class="cropper-head"><h3>Обрезка</h3><button type="button" class="btn btn-secondary" data-crop-close>Закрыть</button></div>
            <div class="cropper-body">
                <div class="cropper-stage" data-crop-stage><canvas class="cropper-canvas" data-crop-canvas></canvas><div class="cropper-frame" data-crop-frame></div></div>
                <aside class="cropper-side"><div class="cropper-preview" data-crop-preview-wrap><p class="text-muted">Превью</p><canvas data-crop-preview></canvas></div><label>Масштаб</label><input data-crop-zoom type="range" min="0.5" max="3" step="0.01" value="1"></aside>
            </div>
            <div class="cropper-actions"><button type="button" class="btn btn-secondary" data-crop-cancel>Отмена</button><button type="button" class="btn btn-primary" data-crop-apply>Применить</button></div>
        </div>`;
    document.body.appendChild(modal);

    // ← ИСПРАВЛЕНИЕ: явное приведение типов для всех элементов
    const stage = /** @type {HTMLElement} */ (modal.querySelector('[data-crop-stage]'));
    const canvas = /** @type {HTMLCanvasElement} */ (modal.querySelector('[data-crop-canvas]'));
    const ctx = canvas.getContext('2d');
    const frame = /** @type {HTMLElement} */ (modal.querySelector('[data-crop-frame]'));
    const previewCanvas = /** @type {HTMLCanvasElement} */ (modal.querySelector('[data-crop-preview]'));
    const previewCtx = previewCanvas.getContext('2d');
    const zoomInput = /** @type {HTMLInputElement} */ (modal.querySelector('[data-crop-zoom]'));

    /** @type {{input: HTMLInputElement, img: HTMLImageElement, kind: string, offsetX: number, offsetY: number}|null} */
    let state = null;

    function frameRect() {
        const s = stage.getBoundingClientRect(), f = frame.getBoundingClientRect();
        return { x: f.left - s.left, y: f.top - s.top, w: f.width, h: f.height };
    }

    function draw() {
        if (!state) return;
        const rect = stage.getBoundingClientRect();
        canvas.width = Math.max(1, Math.round(rect.width));
        canvas.height = Math.max(1, Math.round(rect.height));
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        ctx.fillStyle = '#09090b';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        const cover = Math.max(canvas.width / state.img.width, canvas.height / state.img.height);
        const scale = cover * Number(zoomInput.value || 1);
        const w = state.img.width * scale, h = state.img.height * scale;
        const x = (canvas.width - w) / 2 + (state.offsetX || 0), y = (canvas.height - h) / 2 + (state.offsetY || 0);
        ctx.imageSmoothingQuality = 'high';
        // Теперь state.img уже имеет тип HTMLImageElement из JSDoc
        ctx.drawImage(state.img, x, y, w, h);
        drawPreview();
    }

    function drawPreview() {
        if (!state) return;
        const fr = frameRect();
        const outW = state.kind === 'event' ? 320 : 220, outH = state.kind === 'event' ? 180 : 220;
        previewCanvas.width = outW; previewCanvas.height = outH;
        previewCtx.clearRect(0, 0, outW, outH);
        // canvas теперь имеет тип HTMLCanvasElement, ошибка исчезнет
        previewCtx.drawImage(canvas, fr.x, fr.y, fr.w, fr.h, 0, 0, outW, outH);
    }

    function openCropper(input, file) {
        /** @type {HTMLImageElement} */
        const img = new Image();
        img.onload = () => {
            state = { input, img, kind: input.dataset.cropKind || 'event', offsetX: 0, offsetY: 0 };
            frame.className = `cropper-frame ${state.kind}`;
            modal.querySelector('[data-crop-preview-wrap]').classList.toggle('avatar', state.kind === 'avatar');
            zoomInput.value = '1';
            modal.classList.add('open');
            requestAnimationFrame(draw);
        };
        img.src = URL.createObjectURL(file);
    }

    function closeCropper() { modal.classList.remove('open'); state = null; }

    function applyCrop() {
        if (!state) return;
        const fr = frameRect(), out = document.createElement('canvas');
        out.width = state.kind === 'event' ? 1280 : 512;
        out.height = state.kind === 'event' ? 720 : 512;
        // canvas теперь имеет тип HTMLCanvasElement
        out.getContext('2d').drawImage(canvas, fr.x, fr.y, fr.w, fr.h, 0, 0, out.width, out.height);
        const dataURL = out.toDataURL('image/png');
        const hidden = state.input.dataset.cropHidden ? document.querySelector(state.input.dataset.cropHidden) : null;
        const preview = state.input.dataset.cropPreview ? document.querySelector(state.input.dataset.cropPreview) : null;
        if (hidden) hidden.value = dataURL;
        if (preview) { preview.src = dataURL; preview.style.display = 'block'; }
        closeCropper();
    }

    cropInputs.forEach(inp => inp.addEventListener('change', () => { if (inp.files?.[0]) openCropper(inp, inp.files[0]); }));
    zoomInput.addEventListener('input', draw);
    modal.querySelector('[data-crop-apply]').addEventListener('click', applyCrop);
    modal.querySelector('[data-crop-close]').addEventListener('click', closeCropper);
    modal.querySelector('[data-crop-cancel]').addEventListener('click', closeCropper);
    modal.addEventListener('click', e => { if (e.target === modal) closeCropper(); });

    let drag = null;
    stage.addEventListener('pointerdown', e => {
        if (!state) return;
        drag = { x: e.clientX, y: e.clientY, ox: state.offsetX || 0, oy: state.offsetY || 0 };
        stage.classList.add('dragging');
        stage.setPointerCapture(e.pointerId);
    });
    stage.addEventListener('pointermove', e => {
        if (!drag) return;
        state.offsetX = drag.ox + e.clientX - drag.x;
        state.offsetY = drag.oy + e.clientY - drag.y;
        draw();
    });
    stage.addEventListener('pointerup', () => { drag = null; stage.classList.remove('dragging'); });
    stage.addEventListener('pointercancel', () => { drag = null; stage.classList.remove('dragging'); });
    window.addEventListener('resize', () => { if (state) draw(); });
})();