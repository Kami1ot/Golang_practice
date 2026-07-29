// Панорама и зум SVG-карты через viewBox. Без внешних библиотек.
//
// Управляет атрибутом viewBox готового SVG внутри окна wrap фиксированного
// размера: drag-пан (мышь и тач через Pointer Events), колесо — прокрутка
// (Shift — горизонтальная), Ctrl+колесо — зум к точке курсора (pinch тачпада
// шлёт wheel с ctrlKey — попадает сюда же), pinch двумя пальцами на таче,
// программные fit/centerOn/ensureVisible.
// Клик по узлам карты отличается от драга порогом смещения: после жеста
// длиннее CLICK_SLOP событие click гасится в capture-фазе.

const CLICK_SLOP = 5;      // px: меньше — клик, больше — драг
const TWEEN_MS = 180;      // длительность анимации fit/centerOn
const WHEEL_SPEED = 0.0015;

export function initPanZoom(wrap, svg, opts) {
  const { contentW, contentH } = opts;
  const margin = opts.margin ?? 80;
  const maxScale = opts.maxScale ?? 1.5;
  const baseMinScale = opts.minScale ?? 0.4;
  const reduceMotion = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches;

  const state = { x: 0, y: 0, scale: 1 };
  let raf = null;            // текущий твин
  let tweenGuard = null;     // страховка твина для вкладок с заторможенным rAF
  let moved = 0;             // накопленное смещение жеста (для клик-против-драга)
  let captured = false;      // указатель захвачен = жест признан драгом
  const pointers = new Map(); // активные указатели (пан + pinch)
  let pinchDist = 0;

  const viewW = () => wrap.clientWidth / state.scale;
  const viewH = () => wrap.clientHeight / state.scale;

  // «Вся карта» с полями по краям; и нижний предел зума, чтобы fit был достижим.
  function fitScale() {
    const w = wrap.clientWidth, h = wrap.clientHeight;
    if (!w || !h) return 1;
    return Math.min(w / (contentW + margin * 2), h / (contentH + margin * 2), maxScale);
  }
  const minScale = () => Math.min(baseMinScale, fitScale());

  function clampAxis(v, view, content) {
    // Контент меньше окна — принудительно центрируем (пан по оси заблокирован).
    if (view >= content + margin * 2) return (content - view) / 2;
    return Math.max(-margin, Math.min(content + margin - view, v));
  }

  function apply() {
    if (!wrap.clientWidth || !wrap.clientHeight) return;
    state.scale = Math.max(minScale(), Math.min(maxScale, state.scale));
    state.x = clampAxis(state.x, viewW(), contentW);
    state.y = clampAxis(state.y, viewH(), contentH);
    svg.setAttribute('viewBox', `${state.x} ${state.y} ${viewW()} ${viewH()}`);
  }

  function stopTween() {
    if (raf) { cancelAnimationFrame(raf); raf = null; }
    if (tweenGuard) { clearTimeout(tweenGuard); tweenGuard = null; }
  }

  function animateTo(target) {
    stopTween();
    if (reduceMotion) {
      Object.assign(state, target);
      apply();
      return;
    }
    const from = { ...state };
    const start = performance.now();
    const step = (now) => {
      const t = Math.min(1, (now - start) / TWEEN_MS);
      const e = 1 - (1 - t) * (1 - t); // easeOutQuad
      state.x = from.x + (target.x - from.x) * e;
      state.y = from.y + (target.y - from.y) * e;
      state.scale = from.scale + (target.scale - from.scale) * e;
      apply();
      raf = t < 1 ? requestAnimationFrame(step) : null;
      if (!raf) { clearTimeout(tweenGuard); tweenGuard = null; }
    };
    raf = requestAnimationFrame(step);
    // rAF в фоновой/заторможенной вкладке может не тикать — доводим прыжком.
    tweenGuard = setTimeout(() => {
      if (raf) { stopTween(); Object.assign(state, target); apply(); }
    }, TWEEN_MS + 120);
  }

  // Зум к точке окна (px, py — координаты относительно wrap).
  function zoomAt(px, py, factor) {
    const sx = state.x + px / state.scale;
    const sy = state.y + py / state.scale;
    const next = Math.max(minScale(), Math.min(maxScale, state.scale * factor));
    state.x = sx - px / next;
    state.y = sy - py / next;
    state.scale = next;
    apply();
  }

  // ── Ввод ──────────────────────────────────────────────────────────────────

  // ВАЖНО: захватывать указатель сразу на pointerdown нельзя — captured pointer
  // переадресует последующий click контейнеру, и клики по узлам карты умирают.
  // Поэтому мёртвая зона: захват и пан начинаются только после CLICK_SLOP.
  function engageDrag(pointerId) {
    if (captured) return;
    captured = true;
    try { wrap.setPointerCapture(pointerId); } catch { /* указатель мог уже пропасть */ }
    wrap.classList.add('dragging');
  }

  function onPointerDown(e) {
    stopTween();
    pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
    if (pointers.size === 1) moved = 0;
    if (pointers.size === 2) {
      const [a, b] = [...pointers.values()];
      pinchDist = Math.hypot(a.x - b.x, a.y - b.y);
      // Два пальца — это точно жест, а не клик: захватываем оба указателя.
      for (const id of pointers.keys()) {
        try { wrap.setPointerCapture(id); } catch { /* ок */ }
      }
      captured = true;
      wrap.classList.add('dragging');
    }
  }

  function onPointerMove(e) {
    const prev = pointers.get(e.pointerId);
    if (!prev) return;
    const cur = { x: e.clientX, y: e.clientY };
    pointers.set(e.pointerId, cur);

    if (pointers.size === 1) {
      moved += Math.hypot(cur.x - prev.x, cur.y - prev.y);
      if (moved <= CLICK_SLOP) return; // мёртвая зона: дрожание мыши — ещё не драг
      engageDrag(e.pointerId);
      state.x -= (cur.x - prev.x) / state.scale;
      state.y -= (cur.y - prev.y) / state.scale;
      apply();
    } else if (pointers.size === 2) {
      // Pinch: масштаб — по изменению дистанции, зум — к середине жеста.
      moved = CLICK_SLOP + 1;
      const [a, b] = [...pointers.values()];
      const dist = Math.hypot(a.x - b.x, a.y - b.y);
      const rect = wrap.getBoundingClientRect();
      const midX = (a.x + b.x) / 2 - rect.left;
      const midY = (a.y + b.y) / 2 - rect.top;
      if (pinchDist > 0 && dist > 0) zoomAt(midX, midY, dist / pinchDist);
      pinchDist = dist;
    }
  }

  function onPointerEnd(e) {
    pointers.delete(e.pointerId);
    pinchDist = 0;
    if (!pointers.size) {
      captured = false;
      wrap.classList.remove('dragging');
    }
  }

  function onWheel(e) {
    e.preventDefault();
    stopTween();
    if (e.ctrlKey || e.metaKey) {
      // Ctrl+колесо — зум к курсору (сюда же попадает pinch тачпада).
      const rect = wrap.getBoundingClientRect();
      zoomAt(e.clientX - rect.left, e.clientY - rect.top, Math.exp(-e.deltaY * WHEEL_SPEED));
      return;
    }
    // Обычное колесо — прокрутка карты; Shift превращает вертикаль в горизонталь.
    let dx = e.deltaX, dy = e.deltaY;
    if (e.shiftKey && !dx) { dx = dy; dy = 0; }
    const k = e.deltaMode === 1 ? 16 : 1; // deltaMode 1 — «строки» (Firefox)
    state.x += (dx * k) / state.scale;
    state.y += (dy * k) / state.scale;
    apply();
  }

  // Гасим клик после драга — иначе отпускание кнопки на узле открывает уровень.
  function onClickCapture(e) {
    if (moved > CLICK_SLOP) {
      e.stopPropagation();
      e.preventDefault();
      moved = 0;
    }
  }

  wrap.addEventListener('pointerdown', onPointerDown);
  wrap.addEventListener('pointermove', onPointerMove);
  wrap.addEventListener('pointerup', onPointerEnd);
  wrap.addEventListener('pointercancel', onPointerEnd);
  wrap.addEventListener('wheel', onWheel, { passive: false });
  wrap.addEventListener('click', onClickCapture, true);

  const ro = new ResizeObserver(() => {
    // Сохраняем центр обзора при изменении размеров окна карты.
    const cx = state.x + viewW() / 2;
    const cy = state.y + viewH() / 2;
    state.x = cx - viewW() / 2;
    state.y = cy - viewH() / 2;
    apply();
  });
  ro.observe(wrap);

  // ── Публичное API ─────────────────────────────────────────────────────────

  const api = {
    fit(animate = true) {
      const scale = fitScale();
      const target = {
        scale,
        x: (contentW - wrap.clientWidth / scale) / 2,
        y: (contentH - wrap.clientHeight / scale) / 2,
      };
      animate ? animateTo(target) : (Object.assign(state, target), apply());
    },
    centerOn(cx, cy, scale = 1, animate = true) {
      const target = {
        scale,
        x: cx - wrap.clientWidth / scale / 2,
        y: cy - wrap.clientHeight / scale / 2,
      };
      animate ? animateTo(target) : (Object.assign(state, target), apply());
    },
    // Минимальный сдвиг, чтобы прямоугольник (x, y, w, h) целиком попал в окно.
    ensureVisible(x, y, w, h, pad = 24) {
      let { x: nx, y: ny } = state;
      if (x - pad < nx) nx = x - pad;
      if (x + w + pad > nx + viewW()) nx = x + w + pad - viewW();
      if (y - pad < ny) ny = y - pad;
      if (y + h + pad > ny + viewH()) ny = y + h + pad - viewH();
      if (nx !== state.x || ny !== state.y) animateTo({ x: nx, y: ny, scale: state.scale });
    },
    zoomBy(factor) {
      stopTween();
      zoomAt(wrap.clientWidth / 2, wrap.clientHeight / 2, factor);
    },
    destroy() {
      stopTween();
      ro.disconnect();
      wrap.removeEventListener('pointerdown', onPointerDown);
      wrap.removeEventListener('pointermove', onPointerMove);
      wrap.removeEventListener('pointerup', onPointerEnd);
      wrap.removeEventListener('pointercancel', onPointerEnd);
      wrap.removeEventListener('wheel', onWheel);
      wrap.removeEventListener('click', onClickCapture, true);
    },
  };
  return api;
}
