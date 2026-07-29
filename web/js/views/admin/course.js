// Админка: страница курса — мета, список уровней с порядком, создание уровня.
import { api } from '../../api.js';
import { S } from '../../strings.js';
import { showErrors, clearErrors, flashSaved, field, textInput, numInput, checkbox, parseTags } from './errors.js';

export async function renderAdminCourse(app, courseId) {
  const data = await api.admin.course(courseId);
  const { course, levels } = data;
  app.innerHTML = '';

  const h1 = document.createElement('h1');
  h1.className = 'page-title';
  h1.textContent = course.title;
  app.append(h1);

  // Мета курса.
  const metaForm = document.createElement('form');
  metaForm.className = 'admin-form';
  const titleIn = textInput(course.title);
  const descIn = textInput(course.description);
  const tagsIn = textInput((course.tags || []).join(', '), 'go, основы, backend');
  const row = document.createElement('div');
  row.className = 'row';
  row.append(field('Название', titleIn), field('Описание', descIn), field('Теги (через запятую)', tagsIn));
  const saveBtn = document.createElement('button');
  saveBtn.className = 'btn primary';
  saveBtn.textContent = S.adminSave;
  metaForm.append(row, saveBtn);
  metaForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    clearErrors(app);
    try {
      await api.admin.courseUpdate(courseId, {
        title: titleIn.value.trim(),
        description: descIn.value.trim(),
        tags: parseTags(tagsIn.value),
        levels: course.levels,
      });
      flashSaved(saveBtn);
    } catch (err) {
      showErrors(app, err);
    }
  });
  app.append(metaForm);

  // Уровни: порядок = course.levels.
  const listTitle = document.createElement('h2');
  listTitle.className = 'profile-section-title';
  listTitle.textContent = 'Уровни';
  app.append(listTitle);

  const list = document.createElement('div');
  list.className = 'admin-list';

  const byId = new Map(levels.map(l => [l.id, l]));
  for (const [i, id] of course.levels.entries()) {
    const lv = byId.get(id);
    const row = document.createElement('div');
    row.className = 'admin-row';
    const grow = document.createElement('div');
    grow.className = 'grow';
    const t = document.createElement('div');
    t.className = 'title';
    t.textContent = lv ? lv.title : id + ' (нет level.json)';
    const m = document.createElement('div');
    m.className = 'meta';
    if (lv) {
      const parts = [id, `xp ${lv.xp}`];
      if (lv.requires?.length) parts.push(`после: ${lv.requires.join(', ')}`);
      if (lv.optional) parts.push('бонус');
      if (lv.hasQuiz) parts.push('тест ✓');
      parts.push(`задач: ${lv.tasks}`);
      m.textContent = parts.join(' · ');
    } else {
      m.textContent = id;
    }
    grow.append(t, m);
    row.append(grow);

    // Порядок в course.json (влияет только на раскладку дерева).
    const up = document.createElement('button');
    up.className = 'btn ghost';
    up.textContent = '↑';
    up.disabled = i === 0;
    const down = document.createElement('button');
    down.className = 'btn ghost';
    down.textContent = '↓';
    down.disabled = i === course.levels.length - 1;
    const move = async (delta) => {
      const order = [...course.levels];
      [order[i], order[i + delta]] = [order[i + delta], order[i]];
      clearErrors(app);
      try {
        await api.admin.courseUpdate(courseId, { title: course.title, description: course.description, levels: order });
        renderAdminCourse(app, courseId);
      } catch (err) {
        showErrors(app, err);
      }
    };
    up.addEventListener('click', () => move(-1));
    down.addEventListener('click', () => move(1));

    const open = document.createElement('a');
    open.className = 'btn primary';
    open.href = `#/admin/course/${courseId}/level/${id}`;
    open.textContent = 'Редактировать';
    const del = document.createElement('button');
    del.className = 'btn ghost';
    del.textContent = S.adminDelete;
    del.addEventListener('click', async () => {
      if (!confirm(S.adminDeleteConfirm(`уровень «${lv ? lv.title : id}»`))) return;
      clearErrors(app);
      try {
        await api.admin.levelDelete(courseId, id);
        renderAdminCourse(app, courseId);
      } catch (err) {
        showErrors(app, err);
      }
    });
    row.append(up, down, open, del);
    list.append(row);
  }
  app.append(list);

  // Создание уровня.
  const newTitle = document.createElement('h2');
  newTitle.className = 'profile-section-title';
  newTitle.style.marginTop = '24px';
  newTitle.textContent = S.adminNewLevel;
  app.append(newTitle);

  const form = document.createElement('form');
  form.className = 'admin-form';
  const idIn = textInput('', '05-loops');
  const ltitleIn = textInput('', 'Циклы');
  const summaryIn = textInput('', 'for, range, break/continue');
  const groupIn = textInput('', 'Основы');
  const xpIn = numInput(20);
  const optionalCb = checkbox('бонусный (не блокирует прогресс)');

  const reqBox = document.createElement('div');
  reqBox.className = 'row';
  const reqChecks = [];
  for (const lv of levels) {
    const cb = checkbox(lv.title + ` (${lv.id})`);
    reqChecks.push({ id: lv.id, input: cb.input });
    reqBox.append(cb.label);
  }
  const reqLabel = document.createElement('label');
  reqLabel.append('Открывается после (requires):', reqBox);

  const row1 = document.createElement('div');
  row1.className = 'row';
  row1.append(field('id', idIn), field('Название', ltitleIn), field('Подзаголовок', summaryIn),
    field('Секция (group)', groupIn), field('XP', xpIn));
  const submit = document.createElement('button');
  submit.className = 'btn primary';
  submit.textContent = `+ ${S.adminNewLevel}`;
  form.append(row1, reqLabel, optionalCb.label, submit);
  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    clearErrors(app);
    try {
      await api.admin.levelCreate(courseId, {
        id: idIn.value.trim(),
        title: ltitleIn.value.trim(),
        summary: summaryIn.value.trim(),
        requires: reqChecks.filter(c => c.input.checked).map(c => c.id),
        optional: optionalCb.input.checked,
        xp: Number(xpIn.value) || 0,
        group: groupIn.value.trim(),
      });
      location.hash = `#/admin/course/${courseId}/level/${idIn.value.trim()}`;
    } catch (err) {
      showErrors(app, err);
    }
  });
  app.append(form);

  return { title: course.title };
}
