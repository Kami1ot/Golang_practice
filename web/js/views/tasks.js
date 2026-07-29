// Вкладка «Задачи»: CodeMirror, запуск, результаты, подсказки, автосейв черновиков.
import { api, showAchievements } from '../api.js';
import { S, plural } from '../strings.js';
import { renderMD } from '../md.js';

export function renderTasks(panel, level, ctx) {
  panel.innerHTML = '';
  if (!level.tasks.length) {
    const p = document.createElement('p');
    p.className = 'page-sub';
    p.textContent = S.tasksEmpty;
    panel.append(p);
    return;
  }
  for (const task of level.tasks) {
    panel.append(buildTask(task, level, ctx));
  }
}

function buildTask(task, level, ctx) {
  const card = document.createElement('div');
  card.className = 'task-card';

  // Шапка задачи.
  const head = document.createElement('div');
  head.className = 'task-head';
  const h3 = document.createElement('h3');
  h3.textContent = task.title;
  head.append(h3);

  const typeChip = document.createElement('span');
  typeChip.className = 'chip';
  typeChip.textContent = task.type === 'unittest' ? S.taskUnit : S.taskIO;
  head.append(typeChip);

  if (!task.required) {
    const bonus = document.createElement('span');
    bonus.className = 'chip bonus';
    bonus.textContent = S.taskBonus;
    head.append(bonus);
  }

  const doneChip = document.createElement('span');
  doneChip.className = 'chip done';
  doneChip.hidden = !task.completed;
  doneChip.textContent = S.taskDone(task.attempts || 1);
  head.append(doneChip);
  card.append(head);

  // IDE-сплит: слева условие/примеры/подсказки, справа редактор/кнопки/результаты
  // (на узких экранах CSS складывает колонки обратно в столбик).
  const split = document.createElement('div');
  split.className = 'task-split';
  const side = document.createElement('div');
  side.className = 'task-side';
  const work = document.createElement('div');
  work.className = 'task-work';
  split.append(side, work);
  card.append(split);

  side.append(buildTaskSide(task));

  // Редактор.
  const shell = document.createElement('div');
  shell.className = 'editor-shell';
  const bar = document.createElement('div');
  bar.className = 'editor-bar';
  const fname = document.createElement('span');
  fname.textContent = task.type === 'unittest' ? 'solution.go' : 'main.go';
  const autosave = document.createElement('span');
  autosave.className = 'autosave';
  bar.append(fname, autosave);
  const editorHost = document.createElement('div');
  shell.append(bar, editorHost);
  work.append(shell);

  const cm = CodeMirror(editorHost, {
    value: task.code || task.starterCode || '',
    mode: 'text/x-go',
    theme: 'material-darker',
    lineNumbers: true,
    indentWithTabs: true,
    indentUnit: 4,
    tabSize: 4,
    matchBrackets: true,
    viewportMargin: Infinity,
  });
  // Пересчёт габаритов при смене раскладки сплит↔столбик и ресайзе окна.
  new ResizeObserver(() => cm.refresh()).observe(editorHost);

  // Автосейв черновика: debounce 1 c после последней правки.
  let saveTimer;
  cm.on('change', () => {
    autosave.textContent = S.draftSaving;
    autosave.classList.remove('saved');
    clearTimeout(saveTimer);
    saveTimer = setTimeout(async () => {
      try {
        await api.saveDraft(ctx.courseId, ctx.levelId, task.id, cm.getValue());
        autosave.textContent = S.draftSaved;
        autosave.classList.add('saved');
      } catch {
        autosave.textContent = '';
      }
    }, 1000);
  });

  // Кнопки.
  const actions = document.createElement('div');
  actions.className = 'task-actions';
  const runBtn = document.createElement('button');
  runBtn.className = 'btn primary';
  runBtn.textContent = S.run;
  const resetBtn = document.createElement('button');
  resetBtn.className = 'btn ghost';
  resetBtn.textContent = S.reset;
  const status = document.createElement('span');
  status.className = 'run-status';
  actions.append(runBtn, resetBtn);
  if (ctx.askMentor) {
    const mentorBtn = document.createElement('button');
    mentorBtn.className = 'btn ghost';
    mentorBtn.textContent = `🎓 ${S.mentorAskTask}`;
    mentorBtn.addEventListener('click', () => ctx.askMentor(task.id));
    actions.append(mentorBtn);
  }
  actions.append(status);
  work.append(actions);

  const resultsHost = document.createElement('div');
  work.append(resultsHost);

  resetBtn.addEventListener('click', () => {
    if (confirm(S.resetConfirm)) cm.setValue(task.starterCode || '');
  });

  runBtn.addEventListener('click', async () => {
    runBtn.disabled = true;
    runBtn.innerHTML = `<span class="spin"></span> ${S.running}`;
    status.textContent = '';
    resultsHost.innerHTML = '';
    try {
      const res = await api.runTask(ctx.courseId, ctx.levelId, task.id, cm.getValue());
      renderResults(resultsHost, res);
      status.textContent = res.status === 'passed' ? S.statusPassed : '';
      if (res.taskCompleted && !task.completed) {
        task.completed = true;
        doneChip.hidden = false;
        doneChip.textContent = S.taskDone((task.attempts || 0) + 1);
        const done = level.tasks.filter(t => t.required && t.completed).length;
        const total = level.tasks.filter(t => t.required).length;
        ctx.tasksTick(done, total);
      }
      task.attempts = (task.attempts || 0) + 1;
      showAchievements(res.newAchievements);
      if (res.levelCompleted) ctx.levelCompleted();
    } finally {
      runBtn.disabled = false;
      runBtn.textContent = S.run;
    }
  });

  return card;
}

// buildTaskSide — условие, примеры и подсказки задачи.
// Используется и карточкой (узкие экраны), и workspace-режимом.
export function buildTaskSide(task) {
  const frag = document.createDocumentFragment();

  // Условие.
  const statement = document.createElement('div');
  statement.className = 'task-statement';
  statement.append(renderMD(task.statementMd));
  frag.append(statement);

  // Примеры тест-кейсов (io).
  if (task.examples?.length) {
    const details = document.createElement('details');
    details.className = 'examples';
    const summary = document.createElement('summary');
    summary.textContent = S.examples;
    const table = document.createElement('table');
    const thead = document.createElement('tr');
    for (const cap of [S.exName, S.exInput, S.exOutput]) {
      const th = document.createElement('th');
      th.textContent = cap;
      thead.append(th);
    }
    table.append(thead);
    for (const ex of task.examples) {
      const tr = document.createElement('tr');
      for (const val of [ex.name, ex.stdin, ex.stdout]) {
        const td = document.createElement('td');
        td.textContent = (val ?? '').replace(/\n$/, '');
        tr.append(td);
      }
      table.append(tr);
    }
    details.append(summary, table);
    frag.append(details);
  }

  // Подсказки: раскрываются по одной, от лёгкой к почти-решению.
  if (task.hints?.length) {
    const hintBox = document.createElement('div');
    hintBox.className = 'hints';
    const revealed = document.createElement('div');
    revealed.className = 'hints-revealed';
    const btn = document.createElement('button');
    btn.className = 'btn ghost hint-btn';
    let shown = 0;
    btn.textContent = `💡 ${S.hintButton(1, task.hints.length)}`;
    btn.addEventListener('click', () => {
      if (shown >= task.hints.length) return;
      const h = document.createElement('div');
      h.className = 'hint-item';
      h.append(renderMD(task.hints[shown]));
      revealed.append(h);
      shown++;
      if (shown >= task.hints.length) btn.remove();
      else btn.textContent = `💡 ${S.hintButton(shown + 1, task.hints.length)}`;
    });
    hintBox.append(revealed, btn);
    frag.append(hintBox);
  }
  return frag;
}

export function renderResults(host, res) {
  host.innerHTML = '';

  if (res.status === 'compile_error') {
    const box = document.createElement('div');
    box.className = 'compile-error';
    const cap = document.createElement('div');
    cap.className = 'cap';
    cap.textContent = S.compileError;
    const pre = document.createElement('pre');
    pre.textContent = res.compileOutput;
    box.append(cap, pre);
    host.append(box);
    return;
  }

  if (!res.tests?.length) return;
  const list = document.createElement('div');
  list.className = 'results';

  for (const t of res.tests) {
    const row = document.createElement('div');
    row.className = `res-row ${t.skipped ? 'skip' : t.passed ? 'pass' : 'fail'}`;

    const ico = document.createElement('span');
    ico.className = 'res-ico';
    ico.textContent = t.skipped ? '·' : t.passed ? '✓' : '✕';

    const name = document.createElement('span');
    name.className = 'res-name';
    name.textContent = t.name;

    row.append(ico, name);
    if (t.hidden) {
      const hid = document.createElement('span');
      hid.className = 'res-hidden';
      hid.textContent = S.hiddenTest;
      row.append(hid);
    }
    const time = document.createElement('span');
    time.className = 'res-time';
    time.textContent = t.skipped ? '' : `${t.durationMs} мс`;
    row.append(time);
    list.append(row);

    // Пояснение провала: stderr/сообщение тестов.
    if (!t.passed && !t.skipped && t.output) {
      const out = document.createElement('div');
      out.className = 'res-output';
      out.textContent = t.output;
      list.append(out);
    }
    // Диф для видимых io-тестов.
    if (!t.passed && !t.skipped && !t.hidden && t.expected !== undefined && t.actual !== undefined && !t.output) {
      const diff = document.createElement('div');
      diff.className = 'diff';
      diff.innerHTML =
        `<div class="exp"><div class="cap">${S.expected}</div><pre></pre></div>` +
        `<div class="act"><div class="cap">${S.actual}</div><pre></pre></div>`;
      diff.querySelector('.exp pre').textContent = t.expected;
      diff.querySelector('.act pre').textContent = t.actual;
      list.append(diff);
    }
  }
  host.append(list);
}
