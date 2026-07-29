// Workspace-режим уровня (широкие экраны): полноэкранный двухпанельный layout
// с draggable-сплиттером. Слева — сайдбар разделов (Теория / Тест / Задачи)
// и контент, справа — рабочие зоны: IDE (один редактор на все задачи, контент
// меняется при переключении) и Заметки (Markdown, одна на уровень).
import { api, showAchievements } from '../api.js';
import { S } from '../strings.js';
import { renderMD } from '../md.js';
import { renderQuiz } from './quiz.js';
import { buildTaskSide, renderResults } from './tasks.js';
import { mountChat } from './chat.js';

const SPLIT_KEY = 'gopractice.ws.split';

export function renderLevelWorkspace(app, courseId, levelId, level, tree) {
  app.classList.add('workspace');
  app.innerHTML = '';

  const ws = document.createElement('div');
  ws.className = 'ws-grid'; // не 'workspace' — этот класс носит #app (режим страницы)
  const savedSplit = parseFloat(localStorage.getItem(SPLIT_KEY));
  if (savedSplit >= 24 && savedSplit <= 72) ws.style.setProperty('--ws-split', savedSplit + '%');

  // ── Левый сайдбар: шапка уровня + разделы ────────────────────────────────
  const nav = document.createElement('nav');
  nav.className = 'ws-nav';

  const head = document.createElement('div');
  head.className = 'ws-level-head';
  const eyebrow = document.createElement('div');
  eyebrow.className = 'eyebrow';
  eyebrow.textContent = level.optional ? S.bonusEyebrow(level.xp) : S.levelEyebrow(level.xp);
  const title = document.createElement('h2');
  title.textContent = level.title;
  head.append(eyebrow, title);
  nav.append(head);

  const left = document.createElement('div');
  left.className = 'ws-left';

  const sections = {}; // key -> панель контента слева
  const navItems = {}; // key -> кнопка сайдбара

  function addNavItem(key, label) {
    const btn = document.createElement('button');
    btn.className = 'ws-item';
    const lab = document.createElement('span');
    lab.className = 'ws-item-label';
    lab.textContent = label;
    btn.append(lab);
    btn.addEventListener('click', () => selectSection(key));
    nav.append(btn);
    navItems[key] = btn;

    const panel = document.createElement('section');
    panel.className = 'ws-section';
    panel.hidden = true;
    sections[key] = panel;
    left.append(panel);
    return btn;
  }

  // Теория.
  addNavItem('lesson', S.wsTheory);
  const lesson = document.createElement('article');
  lesson.className = 'lesson';
  lesson.append(renderMD(level.lessonMd));
  sections.lesson.append(lesson);

  // Тест.
  if (level.quiz) {
    addNavItem('quiz', S.tabQuiz);
    const quizBody = document.createElement('div');
    sections.quiz.append(quizBody);
    renderQuiz(quizBody, level, makeCtx());
  }

  // Задачи: разделитель со счётчиком + пункт на задачу.
  let tasksDivider = null;
  if (level.tasks.length) {
    tasksDivider = document.createElement('div');
    tasksDivider.className = 'ws-divider';
    const cap = document.createElement('span');
    cap.textContent = S.tabTasks;
    const cnt = document.createElement('span');
    cnt.className = 'ws-divider-count';
    tasksDivider.append(cap, cnt);
    nav.append(tasksDivider);

    for (const task of level.tasks) {
      const btn = addNavItem('task:' + task.id, task.title);
      if (!task.required) btn.classList.add('bonus');

      const panel = sections['task:' + task.id];
      const thead = document.createElement('div');
      thead.className = 'ws-task-head';
      const h = document.createElement('h2');
      h.textContent = task.title;
      thead.append(h, chip(task.type === 'unittest' ? S.taskUnit : S.taskIO));
      if (!task.required) thead.append(chip(S.taskBonus, 'bonus'));
      const doneChip = chip(S.taskDone(task.attempts || 1), 'done');
      doneChip.hidden = !task.completed;
      thead.append(doneChip);
      task._doneChip = doneChip;
      panel.append(thead, buildTaskSide(task));
    }
  }

  // ── Статусы в сайдбаре ───────────────────────────────────────────────────
  function decorateQuizNav() {
    const btn = navItems.quiz;
    if (!btn) return;
    btn.querySelectorAll('.tick, .count').forEach(el => el.remove());
    if (level.quizPassed) btn.append(tick());
    else {
      const c = document.createElement('span');
      c.className = 'count';
      c.textContent = `${level.quiz.questions.length}`;
      btn.append(c);
    }
  }
  function decorateTaskNav(task) {
    const btn = navItems['task:' + task.id];
    btn.querySelectorAll('.tick').forEach(el => el.remove());
    if (task.completed) btn.append(tick());
  }
  function decorateTasksCount() {
    if (!tasksDivider) return;
    const done = level.tasks.filter(t => t.required && t.completed).length;
    const total = level.tasks.filter(t => t.required).length;
    const cnt = tasksDivider.querySelector('.ws-divider-count');
    cnt.textContent = total ? (done >= total ? '✓' : `${done}/${total}`) : '';
    cnt.classList.toggle('full', total > 0 && done >= total);
  }

  // ── Контекст для quiz/задач (наставник, статусы, баннер) ────────────────
  const chatCtl = mountChat(ws, courseId, levelId, level.tasks);
  function makeCtx() {
    return {
      courseId,
      levelId,
      askMentor: (taskId) => chatCtl.open(taskId),
      quizTick() { decorateQuizNav(); },
      tasksTick() { decorateTasksCount(); },
      levelCompleted() { showWinBanner(); },
    };
  }
  const ctx = makeCtx();

  function showWinBanner() {
    if (ws.querySelector('.win-banner')) return;
    const banner = document.createElement('div');
    banner.className = 'win-banner floating';
    banner.innerHTML = `<span class="big">🎉</span><div><h3></h3><p></p></div>`;
    banner.querySelector('h3').textContent = S.levelDone(level.xp);
    banner.querySelector('p').textContent = S.levelDoneUnlocked;
    const toTree = document.createElement('a');
    toTree.className = 'btn primary';
    toTree.href = `#/course/${courseId}`;
    toTree.textContent = S.toTree;
    const close = document.createElement('button');
    close.className = 'win-close';
    close.textContent = '✕';
    close.setAttribute('aria-label', 'закрыть');
    close.addEventListener('click', () => banner.remove());
    banner.append(toTree, close);
    ws.append(banner);
  }

  // ── Сплиттер ─────────────────────────────────────────────────────────────
  const splitter = document.createElement('div');
  splitter.className = 'ws-splitter';
  splitter.title = S.splitterTitle;
  splitter.addEventListener('pointerdown', (e) => {
    e.preventDefault();
    // capture может бросить InvalidPointerId (синтетика, планшеты) — драг важнее.
    try { splitter.setPointerCapture(e.pointerId); } catch { /* не критично */ }
    splitter.classList.add('dragging');
    document.body.classList.add('ws-resizing');
  });
  splitter.addEventListener('pointermove', (e) => {
    if (!splitter.classList.contains('dragging')) return;
    const rect = ws.getBoundingClientRect();
    const pct = Math.min(72, Math.max(24, (e.clientX - rect.left - nav.offsetWidth) / rect.width * 100));
    ws.style.setProperty('--ws-split', pct + '%');
  });
  splitter.addEventListener('pointerup', (e) => {
    splitter.classList.remove('dragging');
    document.body.classList.remove('ws-resizing');
    try { splitter.releasePointerCapture(e.pointerId); } catch { /* не критично */ }
    const cur = parseFloat(ws.style.getPropertyValue('--ws-split'));
    if (cur) localStorage.setItem(SPLIT_KEY, String(cur));
  });
  splitter.addEventListener('dblclick', () => {
    ws.style.removeProperty('--ws-split');
    localStorage.removeItem(SPLIT_KEY);
  });

  // ── Правая панель: зоны IDE и Заметки ────────────────────────────────────
  const right = document.createElement('div');
  right.className = 'ws-right';

  // IDE: один редактор; плейсхолдер, пока задача не выбрана.
  const zoneIDE = document.createElement('div');
  zoneIDE.className = 'ws-zone ws-ide';
  zoneIDE.hidden = true;

  const idePlaceholder = document.createElement('div');
  idePlaceholder.className = 'ws-placeholder';
  idePlaceholder.textContent = level.tasks.length ? S.ideNoTask : S.ideNoTasks;

  const shell = document.createElement('div');
  shell.className = 'editor-shell';
  shell.hidden = true;
  const bar = document.createElement('div');
  bar.className = 'editor-bar';
  const fname = document.createElement('span');
  const taskName = document.createElement('span');
  taskName.className = 'editor-task';
  const autosave = document.createElement('span');
  autosave.className = 'autosave';
  bar.append(fname, taskName, autosave);
  const editorHost = document.createElement('div');
  editorHost.className = 'ws-editor-host';
  shell.append(bar, editorHost);

  const actions = document.createElement('div');
  actions.className = 'task-actions';
  actions.hidden = true;
  const runBtn = document.createElement('button');
  runBtn.className = 'btn primary';
  runBtn.textContent = S.run;
  runBtn.title = 'Ctrl+Enter';
  const resetBtn = document.createElement('button');
  resetBtn.className = 'btn ghost';
  resetBtn.textContent = S.reset;
  const mentorBtn = document.createElement('button');
  mentorBtn.className = 'btn ghost';
  mentorBtn.textContent = `🎓 ${S.mentorAskTask}`;
  const status = document.createElement('span');
  status.className = 'run-status';
  actions.append(runBtn, resetBtn, mentorBtn, status);

  const resultsHost = document.createElement('div');
  resultsHost.className = 'ws-results';

  zoneIDE.append(idePlaceholder, shell, actions, resultsHost);

  const cm = CodeMirror(editorHost, {
    value: '',
    mode: 'text/x-go',
    theme: 'material-darker',
    lineNumbers: true,
    indentWithTabs: true,
    indentUnit: 4,
    tabSize: 4,
    matchBrackets: true,
    extraKeys: { 'Ctrl-Enter': () => runCurrent(), 'Cmd-Enter': () => runCurrent() },
  });
  new ResizeObserver(() => cm.refresh()).observe(editorHost);

  // Автосейв черновика активной задачи: debounce 1 c, flush при переключении.
  let curTask = null;
  let loadingCode = false;
  let saveTimer = null;
  let dirtyTask = null;

  cm.on('change', () => {
    if (loadingCode || !curTask) return;
    curTask.code = cm.getValue();
    dirtyTask = curTask;
    autosave.textContent = S.draftSaving;
    autosave.classList.remove('saved');
    clearTimeout(saveTimer);
    saveTimer = setTimeout(flushDraft, 1000);
  });

  async function flushDraft() {
    clearTimeout(saveTimer);
    const task = dirtyTask;
    if (!task) return;
    dirtyTask = null;
    try {
      await api.saveDraft(courseId, levelId, task.id, task.code);
      if (curTask === task) {
        autosave.textContent = S.draftSaved;
        autosave.classList.add('saved');
      }
    } catch {
      if (curTask === task) autosave.textContent = '';
    }
  }

  function loadTask(task) {
    if (dirtyTask && dirtyTask !== task) flushDraft();
    curTask = task;
    idePlaceholder.hidden = true;
    shell.hidden = false;
    actions.hidden = false;
    fname.textContent = task.type === 'unittest' ? 'solution.go' : 'main.go';
    taskName.textContent = task.title;
    autosave.textContent = '';
    loadingCode = true;
    cm.setValue(task.code || task.starterCode || '');
    loadingCode = false;
    status.textContent = task._statusText || '';
    resultsHost.innerHTML = '';
    if (task._lastRes) renderResults(resultsHost, task._lastRes);
    cm.refresh();
    cm.focus();
  }

  async function runCurrent() {
    const task = curTask;
    if (!task || runBtn.disabled) return;
    runBtn.disabled = true;
    runBtn.innerHTML = `<span class="spin"></span> ${S.running}`;
    status.textContent = '';
    resultsHost.innerHTML = '';
    try {
      const res = await api.runTask(courseId, levelId, task.id, cm.getValue());
      task._lastRes = res;
      task._statusText = res.status === 'passed' ? S.statusPassed : '';
      task.attempts = (task.attempts || 0) + 1;
      if (res.taskCompleted && !task.completed) task.completed = true;
      if (task.completed) {
        task._doneChip.hidden = false;
        task._doneChip.textContent = S.taskDone(task.attempts || 1);
        decorateTaskNav(task);
        decorateTasksCount();
      }
      if (curTask === task) {
        renderResults(resultsHost, res);
        status.textContent = task._statusText;
      }
      showAchievements(res.newAchievements);
      if (res.levelCompleted) ctx.levelCompleted();
    } finally {
      runBtn.disabled = false;
      runBtn.textContent = S.run;
    }
  }

  runBtn.addEventListener('click', runCurrent);
  resetBtn.addEventListener('click', () => {
    if (curTask && confirm(S.resetConfirm)) cm.setValue(curTask.starterCode || '');
  });
  mentorBtn.addEventListener('click', () => ctx.askMentor(curTask?.id));

  // Заметки: Markdown, одна на уровень, автосейв как у черновиков.
  const zoneNotes = document.createElement('div');
  zoneNotes.className = 'ws-zone ws-notes';
  zoneNotes.hidden = true;

  const group = tree.nodes.find(n => n.id === levelId)?.group || '';
  const noteName = [tree.course.title, group, level.title].filter(Boolean).join(' — ');

  const notesBar = document.createElement('div');
  notesBar.className = 'notes-bar';
  const notesTitle = document.createElement('span');
  notesTitle.className = 'notes-name';
  notesTitle.textContent = noteName;
  notesTitle.title = noteName;
  const previewBtn = document.createElement('button');
  previewBtn.className = 'btn ghost notes-toggle';
  previewBtn.textContent = S.notesPreview;
  const notesSave = document.createElement('span');
  notesSave.className = 'autosave';
  notesBar.append(notesTitle, previewBtn, notesSave);

  const notesArea = document.createElement('textarea');
  notesArea.className = 'notes-area';
  notesArea.placeholder = S.notesPlaceholder;
  notesArea.spellcheck = false;

  const notesPreview = document.createElement('div');
  notesPreview.className = 'notes-preview lesson';
  notesPreview.hidden = true;

  zoneNotes.append(notesBar, notesArea, notesPreview);

  let noteLoaded = false;
  let noteDirty = false;
  let notesTimer = null;

  async function loadNote() {
    if (noteLoaded) return;
    noteLoaded = true;
    try {
      const n = await api.note(courseId, levelId);
      if (!noteDirty && n.body) notesArea.value = n.body;
    } catch {
      noteLoaded = false; // попробуем ещё раз при следующем открытии
    }
  }

  notesArea.addEventListener('input', () => {
    noteDirty = true;
    notesSave.textContent = S.draftSaving;
    notesSave.classList.remove('saved');
    clearTimeout(notesTimer);
    notesTimer = setTimeout(async () => {
      try {
        await api.saveNote(courseId, levelId, notesArea.value);
        notesSave.textContent = S.noteSaved;
        notesSave.classList.add('saved');
      } catch {
        notesSave.textContent = '';
      }
    }, 1000);
  });

  previewBtn.addEventListener('click', () => {
    const showPreview = notesPreview.hidden;
    if (showPreview) {
      notesPreview.innerHTML = '';
      notesPreview.append(renderMD(notesArea.value || ''));
    }
    notesPreview.hidden = !showPreview;
    notesArea.hidden = showPreview;
    previewBtn.textContent = showPreview ? S.notesEdit : S.notesPreview;
  });

  right.append(zoneIDE, zoneNotes);

  // ── Правый рейл: переключатель зон (у правого края экрана) ──────────────
  const rail = document.createElement('div');
  rail.className = 'ws-rail';
  const zoneBtns = {};
  function addZoneBtn(key, icon, label) {
    const b = document.createElement('button');
    b.className = 'ws-rail-btn';
    b.title = label;
    b.setAttribute('aria-label', label);
    const i = document.createElement('span');
    i.className = 'ico';
    i.textContent = icon;
    b.append(i);
    b.addEventListener('click', () => setZone(key));
    rail.append(b);
    zoneBtns[key] = b;
  }
  addZoneBtn('ide', '</>', S.zoneIDE);
  addZoneBtn('notes', '✎', S.zoneNotes);

  // ── Переключение разделов и зон ──────────────────────────────────────────
  let zone = null;

  function setZone(key) {
    if (key === 'ide' && !curTask && level.tasks.length) {
      // IDE без активной задачи — перекидываем левую панель на задачи.
      const next = level.tasks.find(t => !t.completed) || level.tasks[0];
      selectSection('task:' + next.id); // сам вызовет setZone('ide') с задачей
      return;
    }
    zone = key;
    for (const [k, b] of Object.entries(zoneBtns)) b.classList.toggle('active', k === key);
    zoneIDE.hidden = key !== 'ide';
    zoneNotes.hidden = key !== 'notes';
    if (key === 'ide') cm.refresh();
    if (key === 'notes') loadNote();
  }

  function selectSection(key) {
    for (const [k, btn] of Object.entries(navItems)) {
      btn.classList.toggle('active', k === key);
      sections[k].hidden = k !== key;
    }
    left.scrollTop = 0;
    if (key.startsWith('task:')) {
      const task = level.tasks.find(t => 'task:' + t.id === key);
      loadTask(task);
      if (zone !== 'ide') setZone('ide');
    }
  }

  // ── Сборка ───────────────────────────────────────────────────────────────
  ws.append(nav, left, splitter, right, rail);
  app.append(ws);

  decorateQuizNav();
  level.tasks.forEach(decorateTaskNav);
  decorateTasksCount();
  selectSection('lesson');
  setZone('notes'); // при чтении теории справа — заметки

  return { courseTitle: tree.course.title, levelTitle: level.title };
}

function chip(text, cls = '') {
  const c = document.createElement('span');
  c.className = 'chip' + (cls ? ' ' + cls : '');
  c.textContent = text;
  return c;
}

function tick() {
  const s = document.createElement('span');
  s.className = 'tick';
  s.textContent = '✓';
  return s;
}
