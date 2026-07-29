// Админка: редактор уровня — Мета / Урок / Тест / Задачи.
import { api } from '../../api.js';
import { S } from '../../strings.js';
import { renderMD } from '../../md.js';
import { showErrors, clearErrors, flashSaved, field, textInput, numInput, textArea, checkbox } from './errors.js';

export async function renderAdminLevel(app, courseId, levelId) {
  const [data, courseData] = await Promise.all([
    api.admin.level(courseId, levelId),
    api.admin.course(courseId),
  ]);
  app.innerHTML = '';

  const screen = document.createElement('div');
  screen.className = 'screen';
  const head = document.createElement('div');
  head.className = 'level-head';
  const eyebrow = document.createElement('div');
  eyebrow.className = 'eyebrow';
  eyebrow.textContent = `админка · ${courseId} / ${levelId}`;
  const h1 = document.createElement('h1');
  h1.textContent = data.meta.title;
  head.append(eyebrow, h1);

  const tabs = document.createElement('div');
  tabs.className = 'tabs';
  const panels = {};
  const tabBtns = {};
  for (const [key, label] of [['meta', S.adminTabMeta], ['lesson', S.adminTabLesson], ['quiz', S.adminTabQuiz], ['tasks', S.adminTabTasks]]) {
    const btn = document.createElement('button');
    btn.className = 'tab';
    btn.textContent = label;
    btn.addEventListener('click', () => activate(key));
    tabs.append(btn);
    tabBtns[key] = btn;
    const panel = document.createElement('div');
    panel.className = 'tab-panel';
    panel.hidden = true;
    panels[key] = panel;
  }
  function activate(key) {
    for (const [k, btn] of Object.entries(tabBtns)) {
      btn.classList.toggle('active', k === key);
      panels[k].hidden = k !== key;
    }
    panels[key].querySelectorAll('.CodeMirror').forEach(el => el.CodeMirror?.refresh());
  }

  buildMetaTab(panels.meta, app, courseId, levelId, data.meta, courseData.levels);
  buildLessonTab(panels.lesson, app, courseId, levelId, data);
  buildQuizTab(panels.quiz, app, courseId, levelId, data.quiz);
  buildTasksTab(panels.tasks, app, courseId, levelId, data.tasks);

  screen.append(head, tabs, panels.meta, panels.lesson, panels.quiz, panels.tasks);
  app.append(screen);
  activate('meta');

  return { courseTitle: courseData.course.title, levelTitle: data.meta.title };
}

function cmEditor(host, value, mode, minHeight = 160) {
  const cm = CodeMirror(host, {
    value: value || '',
    mode,
    theme: 'material-darker',
    lineNumbers: true,
    indentWithTabs: mode === 'text/x-go',
    indentUnit: mode === 'text/x-go' ? 4 : 2,
    tabSize: 4,
    lineWrapping: mode === 'markdown',
    viewportMargin: Infinity,
  });
  cm.getWrapperElement().style.border = '1px solid var(--border-strong)';
  cm.getWrapperElement().style.borderRadius = '8px';
  cm.getScrollerElement().style.minHeight = minHeight + 'px';
  return cm;
}

// ── Мета ────────────────────────────────────────────────────────────────────

function buildMetaTab(panel, app, courseId, levelId, meta, allLevels) {
  const form = document.createElement('form');
  form.className = 'admin-form';
  const titleIn = textInput(meta.title);
  const summaryIn = textInput(meta.summary);
  const xpIn = numInput(meta.xp);
  const groupIn = textInput(meta.group || '', 'Основы');
  const optionalCb = checkbox('бонусный уровень', meta.optional);

  const reqBox = document.createElement('div');
  reqBox.className = 'row';
  const reqChecks = [];
  for (const lv of allLevels) {
    if (lv.id === levelId) continue;
    const cb = checkbox(`${lv.title} (${lv.id})`, meta.requires?.includes(lv.id));
    reqChecks.push({ id: lv.id, input: cb.input });
    reqBox.append(cb.label);
  }
  const reqLabel = document.createElement('label');
  reqLabel.append('Открывается после (requires):', reqBox);

  const row = document.createElement('div');
  row.className = 'row';
  row.append(field('Название', titleIn), field('Подзаголовок', summaryIn),
    field('Секция на карте (group)', groupIn), field('XP', xpIn));
  const save = document.createElement('button');
  save.className = 'btn primary';
  save.textContent = S.adminSave;
  form.append(row, reqLabel, optionalCb.label, save);
  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    clearErrors(panel);
    try {
      await api.admin.levelUpdate(courseId, levelId, {
        id: levelId,
        title: titleIn.value.trim(),
        summary: summaryIn.value.trim(),
        requires: reqChecks.filter(c => c.input.checked).map(c => c.id),
        optional: optionalCb.input.checked,
        xp: Number(xpIn.value) || 0,
        group: groupIn.value.trim(),
      });
      flashSaved(save);
    } catch (err) {
      showErrors(panel, err);
    }
  });
  panel.append(form);
}

// ── Урок ────────────────────────────────────────────────────────────────────

function buildLessonTab(panel, app, courseId, levelId, data) {
  const split = document.createElement('div');
  split.className = 'admin-split';
  const editorHost = document.createElement('div');
  const preview = document.createElement('div');
  preview.className = 'admin-preview lesson';
  split.append(editorHost, preview);

  const cm = cmEditor(editorHost, data.lessonMd, 'markdown', 320);
  const renderPreview = () => {
    preview.innerHTML = '';
    preview.append(renderMD(cm.getValue()));
  };
  let t;
  cm.on('change', () => { clearTimeout(t); t = setTimeout(renderPreview, 400); });
  renderPreview();

  const bar = document.createElement('div');
  bar.className = 'admin-toolbar';
  bar.style.marginTop = '12px';
  const save = document.createElement('button');
  save.className = 'btn primary';
  save.textContent = S.adminSave;
  save.addEventListener('click', async () => {
    clearErrors(panel);
    try {
      await api.admin.lessonPut(courseId, levelId, cm.getValue());
      flashSaved(save);
    } catch (err) {
      showErrors(panel, err);
    }
  });

  // Картинки.
  const upload = document.createElement('label');
  upload.className = 'btn ghost';
  upload.textContent = `📎 ${S.adminUploadImage}`;
  const fileIn = document.createElement('input');
  fileIn.type = 'file';
  fileIn.accept = 'image/*';
  fileIn.hidden = true;
  upload.append(fileIn);
  const imgList = document.createElement('div');
  imgList.className = 'img-list';

  const renderImages = (names) => {
    imgList.innerHTML = '';
    for (const name of names) {
      const chip = document.createElement('span');
      chip.className = 'img-chip';
      chip.textContent = `img/${name}`;
      const del = document.createElement('button');
      del.textContent = '✕';
      del.addEventListener('click', async () => {
        await api.admin.imageDelete(courseId, levelId, name).catch(err => showErrors(panel, err));
        chip.remove();
      });
      chip.append(del);
      imgList.append(chip);
    }
  };
  renderImages(data.images || []);

  fileIn.addEventListener('change', async () => {
    if (!fileIn.files.length) return;
    clearErrors(panel);
    try {
      const res = await api.admin.imageUpload(courseId, levelId, fileIn.files[0]);
      cm.replaceSelection(res.markdown + '\n');
      const fresh = await api.admin.level(courseId, levelId);
      renderImages(fresh.images || []);
    } catch (err) {
      showErrors(panel, err);
    } finally {
      fileIn.value = '';
    }
  });

  bar.append(save, upload);
  panel.append(split, bar, imgList);
}

// ── Тест (конструктор вопросов) ────────────────────────────────────────────

function buildQuizTab(panel, app, courseId, levelId, rawQuiz) {
  const quiz = rawQuiz || { passScore: 1, questions: [] };
  const questions = quiz.questions || [];

  const info = document.createElement('p');
  info.className = 'page-sub';
  info.textContent = 'Ответы хранятся только на сервере и никогда не отдаются ученикам.';
  panel.append(info);

  const listHost = document.createElement('div');
  panel.append(listHost);

  const builders = [];

  function addQuestion(q) {
    const box = document.createElement('div');
    box.className = 'q-builder';
    const headRow = document.createElement('div');
    headRow.className = 'q-head';
    const typeSel = document.createElement('select');
    for (const [v, label] of [['single', 'один вариант'], ['multi', 'несколько'], ['output', 'что выведет код'], ['blank', 'пропуск']]) {
      const opt = document.createElement('option');
      opt.value = v;
      opt.textContent = label;
      typeSel.append(opt);
    }
    typeSel.value = q.type || 'single';
    const removeBtn = document.createElement('button');
    removeBtn.type = 'button';
    removeBtn.className = 'btn ghost';
    removeBtn.textContent = '✕ вопрос';
    headRow.append(typeSel, removeBtn);

    const promptIn = textArea(q.prompt || '', 2);
    const codeIn = textArea(q.code || '', 4);
    const optionsIn = textArea((q.options || []).join('\n'), 4);
    const answerIn = textInput(
      q.type === 'multi' ? (q.answers || []).join(',')
        : q.type === 'blank' ? (q.answers || []).join('|')
        : q.answer !== undefined ? String(q.answer) : '');
    const explainIn = textArea(q.explanation || '', 2);

    const fields = document.createElement('div');
    const rebuild = () => {
      fields.innerHTML = '';
      fields.append(field('Вопрос (Markdown)', promptIn));
      const type = typeSel.value;
      if (type === 'single' || type === 'multi') {
        fields.append(field('Варианты (по одному на строку, Markdown)', optionsIn));
        fields.append(field(type === 'single' ? 'Номер правильного (с 0)' : 'Номера правильных через запятую (с 0)', answerIn));
      } else if (type === 'output') {
        fields.append(field('Код программы', codeIn));
        fields.append(field('Точный вывод', answerIn));
      } else {
        fields.append(field('Допустимые ответы через |', answerIn));
      }
      fields.append(field('Объяснение (показывается после проверки)', explainIn));
    };
    typeSel.addEventListener('change', rebuild);
    rebuild();

    box.append(headRow, fields);
    listHost.append(box);

    const builder = {
      box,
      toJSON(index) {
        const type = typeSel.value;
        const out = { id: `q${index + 1}`, type, prompt: promptIn.value, explanation: explainIn.value };
        if (type === 'single' || type === 'multi') {
          out.options = optionsIn.value.split('\n').map(s => s.trim()).filter(Boolean);
          if (type === 'single') out.answer = Number(answerIn.value);
          else out.answers = answerIn.value.split(',').map(s => Number(s.trim())).filter(n => !Number.isNaN(n));
        } else if (type === 'output') {
          out.code = codeIn.value;
          out.answer = answerIn.value;
        } else {
          out.answers = answerIn.value.split('|').map(s => s.trim()).filter(Boolean);
        }
        return out;
      },
    };
    builders.push(builder);
    removeBtn.addEventListener('click', () => {
      builders.splice(builders.indexOf(builder), 1);
      box.remove();
    });
  }

  for (const q of questions) addQuestion(q);

  const bar = document.createElement('div');
  bar.className = 'admin-toolbar';
  const addBtn = document.createElement('button');
  addBtn.className = 'btn ghost';
  addBtn.textContent = '+ вопрос';
  addBtn.addEventListener('click', () => addQuestion({ type: 'single' }));
  const save = document.createElement('button');
  save.className = 'btn primary';
  save.textContent = S.adminSave;
  save.addEventListener('click', async () => {
    clearErrors(panel);
    try {
      if (builders.length === 0) {
        await api.admin.quizDelete(courseId, levelId);
      } else {
        await api.admin.quizPut(courseId, levelId, {
          passScore: 1,
          questions: builders.map((b, i) => b.toJSON(i)),
        });
      }
      flashSaved(save);
    } catch (err) {
      showErrors(panel, err);
    }
  });
  bar.append(addBtn, save);
  panel.append(bar);
}

// ── Задачи ──────────────────────────────────────────────────────────────────

function buildTasksTab(panel, app, courseId, levelId, tasks) {
  const listHost = document.createElement('div');
  panel.append(listHost);

  function addTaskEditor(task, isNew) {
    const box = document.createElement('div');
    box.className = 'admin-form';
    listHost.append(box);

    const idIn = textInput(task.id || '', 'sum-func');
    idIn.disabled = !isNew;
    const titleIn = textInput(task.title || '');
    const typeSel = document.createElement('select');
    for (const [v, label] of [['io', 'io: ввод/вывод'], ['unittest', 'unittest: скрытые тесты']]) {
      const opt = document.createElement('option');
      opt.value = v;
      opt.textContent = label;
      typeSel.append(opt);
    }
    typeSel.value = task.type || 'io';
    const orderIn = numInput(task.order ?? 1);
    const timeoutIn = numInput(task.timeoutSec || 10);
    const requiredCb = checkbox('обязательная', task.required !== false);

    const row1 = document.createElement('div');
    row1.className = 'row';
    row1.append(field('id', idIn), field('Название', titleIn), field('Тип', typeSel),
      field('Порядок', orderIn), field('Таймаут, сек', timeoutIn));
    box.append(row1, requiredCb.label);

    const stmtHost = document.createElement('div');
    box.append(field('Условие (Markdown)', stmtHost));
    const stmtCM = cmEditor(stmtHost, task.statementMd, 'markdown', 100);

    const starterHost = document.createElement('div');
    box.append(field('starter.go (заготовка в редакторе ученика)', starterHost));
    const starterCM = cmEditor(starterHost, task.starterCode || 'package main\n\nfunc main() {\n\t\n}\n', 'text/x-go', 100);

    // io: таблица тестов.
    const testsHost = document.createElement('div');
    const testRows = [];
    const testsLabel = field('Тест-кейсы (io)', testsHost);
    const addTestRow = (tc = { name: '', stdin: '', stdout: '', hidden: true }) => {
      const row = document.createElement('div');
      row.className = 'row';
      const name = textInput(tc.name, 'название');
      const stdin = textInput(tc.stdin, 'stdin');
      const stdout = textInput(tc.stdout, 'ожидаемый stdout (\\n = перевод строки)');
      const hidden = checkbox('скрытый', tc.hidden);
      const rm = document.createElement('button');
      rm.type = 'button';
      rm.className = 'btn ghost';
      rm.textContent = '✕';
      row.append(field('имя', name), field('stdin', stdin), field('stdout', stdout), hidden.label, rm);
      testsHost.append(row);
      const entry = { name, stdin, stdout, hidden: hidden.input, row };
      testRows.push(entry);
      rm.addEventListener('click', () => { testRows.splice(testRows.indexOf(entry), 1); row.remove(); });
    };
    for (const tc of task.tests || []) {
      addTestRow({ ...tc, stdin: JSON.stringify(tc.stdin).slice(1, -1), stdout: JSON.stringify(tc.stdout).slice(1, -1) });
    }
    const addTest = document.createElement('button');
    addTest.type = 'button';
    addTest.className = 'btn ghost';
    addTest.textContent = '+ тест-кейс';
    addTest.addEventListener('click', () => addTestRow());
    testsHost.append(addTest);

    // unittest: main_test.go.
    const testFileHost = document.createElement('div');
    const testFileLabel = field('main_test.go (скрытые unit-тесты, package main)', testFileHost);
    const testFileCM = cmEditor(testFileHost, task.testFile || 'package main\n\nimport "testing"\n\nfunc TestX(t *testing.T) {\n\t\n}\n', 'text/x-go', 100);

    const hintsIn = textArea((task.hints || []).join('\n'), 3);
    box.append(field('Подсказки (по одной на строку, от лёгкой к почти-решению)', hintsIn));
    box.append(testsLabel, testFileLabel);

    const syncTypeUI = () => {
      testsLabel.hidden = typeSel.value !== 'io';
      testFileLabel.hidden = typeSel.value !== 'unittest';
      if (typeSel.value === 'unittest') testFileCM.refresh();
    };
    typeSel.addEventListener('change', syncTypeUI);
    syncTypeUI();

    const bar = document.createElement('div');
    bar.className = 'admin-toolbar';
    const save = document.createElement('button');
    save.className = 'btn primary';
    save.textContent = S.adminSave;
    save.addEventListener('click', async () => {
      clearErrors(box);
      const unescape = (s) => s.replaceAll('\\n', '\n').replaceAll('\\t', '\t');
      const payload = {
        id: idIn.value.trim(),
        type: typeSel.value,
        title: titleIn.value.trim(),
        required: requiredCb.input.checked,
        order: Number(orderIn.value) || 1,
        timeoutSec: Number(timeoutIn.value) || 10,
        statementMd: stmtCM.getValue(),
        starterCode: starterCM.getValue(),
        hints: hintsIn.value.split('\n').map(s => s.trim()).filter(Boolean),
      };
      if (typeSel.value === 'io') {
        payload.tests = testRows.map(tr => ({
          name: tr.name.value, stdin: unescape(tr.stdin.value),
          stdout: unescape(tr.stdout.value), hidden: tr.hidden.checked,
        }));
      } else {
        payload.testFile = testFileCM.getValue();
      }
      try {
        await api.admin.taskSave(courseId, levelId, payload);
        idIn.disabled = true;
        flashSaved(save);
      } catch (err) {
        showErrors(box, err);
      }
    });
    const del = document.createElement('button');
    del.className = 'btn ghost';
    del.textContent = S.adminDelete;
    del.addEventListener('click', async () => {
      if (isNew && idIn.disabled === false) { box.remove(); return; }
      if (!confirm(S.adminDeleteConfirm(`задачу «${titleIn.value}»`))) return;
      clearErrors(box);
      try {
        await api.admin.taskDelete(courseId, levelId, idIn.value.trim());
        box.remove();
      } catch (err) {
        showErrors(box, err);
      }
    });
    bar.append(save, del);
    box.append(bar);
  }

  for (const t of tasks || []) addTaskEditor(t, false);

  const addBtn = document.createElement('button');
  addBtn.className = 'btn primary';
  addBtn.textContent = `+ ${S.adminNewTask}`;
  addBtn.addEventListener('click', () => addTaskEditor({}, true));
  panel.append(addBtn);
}
