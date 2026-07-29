// Экран дерева курса: послойная раскладка DAG + отрисовка в SVG.
// Карта живёт в окне фиксированной высоты с зумом и панорамой (panzoom.js).
import { api } from '../api.js';
import { S } from '../strings.js';
import { initPanZoom } from '../panzoom.js';

const NODE_W = 248;
const NODE_H = 72;
const STEP_X = 278;
const STEP_Y = 150;
const TOP = 56;
const GROUP_GAP = 64; // дополнительный зазор перед новой секцией
const SVG_NS = 'http://www.w3.org/2000/svg';

export async function renderTree(app, courseId) {
  const tree = await api.tree(courseId);
  app.innerHTML = '';

  const screen = document.createElement('div');
  screen.className = 'screen';

  // Шапка экрана с прогрессом.
  const head = document.createElement('div');
  head.className = 'screen-head';
  const headText = document.createElement('div');
  const h1 = document.createElement('h1');
  h1.textContent = tree.course.title;
  const hint = document.createElement('p');
  hint.textContent = S.treeHint;
  headText.append(h1, hint);

  const completedCount = tree.nodes.filter(n => n.state === 'completed').length;
  const prog = document.createElement('div');
  prog.className = 'progress-block';
  const row = document.createElement('div');
  row.className = 'progress-row';
  row.innerHTML = `<span></span><b></b>`;
  row.querySelector('span').textContent = S.progress;
  row.querySelector('b').textContent = `${completedCount}/${tree.nodes.length} · ${tree.xpEarned}/${tree.xpTotal} XP`;
  const bar = document.createElement('div');
  bar.className = 'bar';
  const fill = document.createElement('i');
  fill.style.width = tree.nodes.length ? `${(100 * completedCount / tree.nodes.length).toFixed(0)}%` : '0%';
  bar.append(fill);
  prog.append(row, bar);
  head.append(headText, prog);

  const wrap = document.createElement('div');
  wrap.className = 'tree-wrap map-wrap';
  const { svg, width, height, pos, hereId } = buildSVG(tree, courseId);
  wrap.append(svg);

  // Контролы поверх окна карты.
  const controls = document.createElement('div');
  controls.className = 'map-controls';
  const ctlBtn = (label, text) => {
    const b = document.createElement('button');
    b.type = 'button';
    b.setAttribute('aria-label', label);
    b.title = label;
    b.textContent = text;
    controls.append(b);
    return b;
  };
  const zoomInBtn = ctlBtn(S.mapZoomIn, '+');
  const zoomOutBtn = ctlBtn(S.mapZoomOut, '−');
  const fitBtn = ctlBtn(S.mapFit, '⛶');
  const hereBtn = ctlBtn(S.mapToCurrent, '◉');
  hereBtn.disabled = !hereId;
  wrap.append(controls);

  // Легенда — оверлей в нижнем левом углу окна.
  const legend = document.createElement('div');
  legend.className = 'legend map-legend';
  legend.innerHTML =
    `<span><i class="dot c"></i>${S.legendDone}</span>` +
    `<span><i class="dot a"></i>${S.legendAvailable}</span>` +
    `<span><i class="dot"></i>${S.legendLocked}</span>` +
    `<span><i class="dot b"></i>${S.legendBonus}</span>`;
  wrap.append(legend);

  screen.append(head, wrap);
  app.append(screen);

  // Пан/зум поднимается после вставки в DOM: ему нужны размеры окна.
  const pz = initPanZoom(wrap, svg, { contentW: width, contentH: height });
  const centerOnHere = (animate) => {
    const p = pos.get(hereId);
    pz.centerOn(p.x + NODE_W / 2, p.y + NODE_H / 2, 1, animate);
  };
  if (hereId) centerOnHere(false);
  else pz.fit(false);

  zoomInBtn.addEventListener('click', () => pz.zoomBy(1.25));
  zoomOutBtn.addEventListener('click', () => pz.zoomBy(1 / 1.25));
  fitBtn.addEventListener('click', () => pz.fit());
  if (hereId) hereBtn.addEventListener('click', () => centerOnHere(true));

  // Tab-навигация: сфокусированный узел подкручивается в видимую область.
  svg.addEventListener('focusin', (e) => {
    const id = e.target.closest?.('.node')?.dataset.id;
    const p = id && pos.get(id);
    if (p) pz.ensureVisible(p.x, p.y, NODE_W, NODE_H);
  });

  return tree;
}

// Послойная раскладка: глубина по requires, порядок в слое — барицентр родителей.
function layout(nodes) {
  const byId = new Map(nodes.map(n => [n.id, n]));
  const depth = new Map();
  const calcDepth = (id) => {
    if (depth.has(id)) return depth.get(id);
    const n = byId.get(id);
    const d = n.requires.length ? 1 + Math.max(...n.requires.map(calcDepth)) : 0;
    depth.set(id, d);
    return d;
  };
  nodes.forEach(n => calcDepth(n.id));

  const layers = [];
  nodes.forEach(n => {
    const d = depth.get(n.id);
    (layers[d] ??= []).push(n);
  });

  // Один проход сверху вниз: сортируем слой по среднему X родителей.
  const xIndex = new Map();
  layers.forEach((layer, d) => {
    if (d > 0) {
      const orderPos = new Map(nodes.map((n, i) => [n.id, i]));
      layer.sort((a, b) => {
        const bary = (n) => {
          const xs = n.requires.map(r => xIndex.get(r)).filter(x => x !== undefined);
          return xs.length ? xs.reduce((s, x) => s + x, 0) / xs.length : 0;
        };
        return (bary(a) - bary(b)) || (orderPos.get(a.id) - orderPos.get(b.id));
      });
    }
    layer.forEach((n, i) => xIndex.set(n.id, i - (layer.length - 1) / 2));
  });

  const maxAbsX = Math.max(0.5, ...nodes.map(n => Math.abs(xIndex.get(n.id))));
  const width = Math.max(720, maxAbsX * 2 * STEP_X + NODE_W + 60);
  const centerX = width / 2;

  // Секции (group): перед слоем, где начинается новая группа, — зазор и подпись.
  const layerY = [];
  const groupLabels = [];
  let offset = 0;
  let prevGroup = null;
  layers.forEach((layer, d) => {
    const g = layer[0]?.group || null;
    if (g && g !== prevGroup) {
      offset += GROUP_GAP;
      groupLabels.push({ y: TOP + d * STEP_Y + offset - GROUP_GAP / 2 - 4, title: g });
      prevGroup = g;
    } else if (g === null && prevGroup !== null) {
      // безгрупповые уровни продолжают текущую секцию
    }
    layerY[d] = TOP + d * STEP_Y + offset;
  });

  const pos = new Map();
  nodes.forEach(n => {
    pos.set(n.id, {
      x: centerX + xIndex.get(n.id) * STEP_X - NODE_W / 2,
      y: layerY[depth.get(n.id)],
    });
  });
  const height = layerY[layers.length - 1] + NODE_H + 40;
  return { pos, width, height, byId, groupLabels };
}

function el(name, attrs = {}, cls = '') {
  const node = document.createElementNS(SVG_NS, name);
  for (const [k, v] of Object.entries(attrs)) node.setAttribute(k, v);
  if (cls) node.setAttribute('class', cls);
  return node;
}

function buildSVG(tree, courseId) {
  const { pos, width, height, byId, groupLabels } = layout(tree.nodes);
  // width/height не ставим: элемент растягивается CSS, обзором управляет
  // panzoom через viewBox (начальный — вся карта).
  const svg = el('svg', {
    viewBox: `0 0 ${width} ${height}`,
    role: 'img',
    'aria-label': `Дерево уровней курса «${tree.course.title}»`,
  }, 'tree');

  // Подписи секций: линия — подпись — линия.
  for (const g of groupLabels) {
    const textW = g.title.length * 8.4 + 28;
    svg.append(el('line', { x1: 36, y1: g.y, x2: (width - textW) / 2, y2: g.y }, 'group-sep'));
    svg.append(el('line', { x1: (width + textW) / 2, y1: g.y, x2: width - 36, y2: g.y }, 'group-sep'));
    const label = el('text', { x: width / 2, y: g.y + 4, 'text-anchor': 'middle' }, 'group-label');
    label.textContent = g.title.toUpperCase();
    svg.append(label);
  }

  // Рёбра — под узлами.
  for (const n of tree.nodes) {
    const to = pos.get(n.id);
    for (const req of n.requires) {
      const from = pos.get(req);
      if (!from) continue;
      const x1 = from.x + NODE_W / 2, y1 = from.y + NODE_H;
      const x2 = to.x + NODE_W / 2, y2 = to.y;
      const midY = (y1 + y2) / 2;
      let cls = 'edge';
      if (n.state === 'locked') cls += ' to-locked';
      else if (byId.get(req).state === 'completed' && n.state === 'completed') cls += ' done';
      else cls += ' next';
      svg.append(el('path', { d: `M ${x1} ${y1} C ${x1} ${midY} ${x2} ${midY} ${x2} ${y2}` }, cls));
    }
  }

  // «Вы здесь» — первый доступный необязательный уровень (в порядке курса).
  const hereId = (tree.nodes.find(n => n.state === 'available' && !n.optional)
    || tree.nodes.find(n => n.state === 'available'))?.id;

  for (const n of tree.nodes) {
    const p = pos.get(n.id);
    const g = el('g', { tabindex: n.state === 'locked' ? -1 : 0, role: 'link', 'data-id': n.id },
      `node ${n.state}${n.optional ? ' bonus' : ''}`);

    g.append(el('rect', { x: p.x, y: p.y, width: NODE_W, height: NODE_H, rx: 12 }, 'node-box'));

    // Бейдж состояния слева.
    const bx = p.x + 30, by = p.y + NODE_H / 2;
    if (n.state === 'completed') {
      g.append(el('circle', { cx: bx, cy: by, r: 12 }, 'badge-done'));
      g.append(el('path', {
        d: `M ${bx - 5.5} ${by} l 4 4 l 7 -7.5`,
        stroke: 'var(--accent-ink)', 'stroke-width': 2.4, fill: 'none', 'stroke-linecap': 'round',
      }));
    } else if (n.state === 'available' && !n.optional) {
      g.append(el('circle', { cx: bx, cy: by, r: 12 }, 'badge-play'));
      g.append(el('path', { d: `M ${bx - 3.5} ${by - 6.5} l 10 6.5 l -10 6.5 z`, fill: 'var(--accent-ink)' }));
    } else if (n.optional) {
      const star = el('text', { x: bx - 7, y: by + 5.5, 'font-size': 15, fill: 'var(--gold)' });
      star.textContent = '★';
      g.append(star);
    } else {
      g.append(el('rect', { x: bx - 7, y: by - 2, width: 14, height: 11, rx: 2.5, fill: 'var(--locked)' }));
      g.append(el('path', {
        d: `M ${bx - 4} ${by - 1} v -3 a 4 4 0 0 1 8 0 v 3`,
        stroke: 'var(--locked)', 'stroke-width': 2, fill: 'none',
      }));
    }

    const title = el('text', { x: p.x + 54, y: p.y + 30 }, 'node-title');
    title.textContent = truncate(n.title, 20);
    g.append(title);

    const sub = el('text', { x: p.x + 54, y: p.y + 50 },
      n.id === hereId ? 'node-here' : 'node-sub');
    sub.textContent = n.id === hereId ? S.youAreHere : subLine(n);
    g.append(sub);

    const xp = el('text', { x: p.x + NODE_W - 14, y: p.y + 22, 'text-anchor': 'end' }, 'node-xp');
    xp.textContent = `+${n.xp}`;
    g.append(xp);

    // Подсказка: для закрытых — какие уровни открыть сначала.
    const tip = el('title');
    tip.textContent = n.state === 'locked'
      ? S.lockedAfter(n.requires.map(r => byId.get(r)?.title || r).join(', '))
      : `${n.title} — ${n.summary}`;
    g.append(tip);

    if (n.state !== 'locked') {
      const go = () => { location.hash = `#/course/${courseId}/level/${n.id}`; };
      g.addEventListener('click', go);
      g.addEventListener('keydown', (e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); go(); } });
    }
    svg.append(g);
  }
  return { svg, width, height, pos, hereId };
}

function subLine(n) {
  const parts = [];
  if (n.progress.hasQuiz) parts.push(`тест ${n.progress.quizPassed ? '✓' : '—'}`);
  if (n.progress.tasksRequired > 0) parts.push(`задачи ${n.progress.tasksDone}/${n.progress.tasksRequired}`);
  if (!parts.length && n.optional) parts.push(S.optionalLevel);
  return parts.join(' · ');
}

function truncate(s, max) {
  return s.length > max ? s.slice(0, max - 1) + '…' : s;
}
