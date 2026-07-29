// Админка: список курсов + создание.
import { api } from '../../api.js';
import { S } from '../../strings.js';
import { showErrors, clearErrors, field, textInput, parseTags } from './errors.js';

export async function renderAdminCourses(app) {
  const courses = await api.admin.courses();
  app.innerHTML = '';

  const h1 = document.createElement('h1');
  h1.className = 'page-title';
  h1.textContent = S.adminCoursesTitle;
  const sub = document.createElement('p');
  sub.className = 'page-sub';
  sub.textContent = 'Курсы — это файлы в content/. Каждое сохранение проверяется валидатором; при ошибке изменения откатываются.';
  app.append(h1, sub);

  // Форма создания.
  const form = document.createElement('form');
  form.className = 'admin-form';
  const idIn = textInput('', 'go-flow');
  const titleIn = textInput('', 'Циклы и коллекции');
  const descIn = textInput('', 'Короткое описание для карточки');
  const tagsIn = textInput('', 'go, основы, backend');
  const row = document.createElement('div');
  row.className = 'row';
  row.append(field('id (латиница, дефисы)', idIn), field('Название', titleIn), field('Описание', descIn), field('Теги (через запятую)', tagsIn));
  const submit = document.createElement('button');
  submit.className = 'btn primary';
  submit.textContent = `+ ${S.adminNewCourse}`;
  form.append(row, submit);
  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    clearErrors(app);
    try {
      await api.admin.courseCreate({
        id: idIn.value.trim(),
        title: titleIn.value.trim(),
        description: descIn.value.trim(),
        tags: parseTags(tagsIn.value),
      });
      location.hash = `#/admin/course/${idIn.value.trim()}`;
    } catch (err) {
      showErrors(app, err);
    }
  });
  app.append(form);

  const list = document.createElement('div');
  list.className = 'admin-list';
  for (const c of courses) {
    const row = document.createElement('div');
    row.className = 'admin-row';
    const grow = document.createElement('div');
    grow.className = 'grow';
    const title = document.createElement('div');
    title.className = 'title';
    title.textContent = c.title;
    const meta = document.createElement('div');
    meta.className = 'meta';
    meta.textContent = `${c.id} · уровней: ${c.levels.length}`;
    grow.append(title, meta);

    const open = document.createElement('a');
    open.className = 'btn primary';
    open.href = `#/admin/course/${c.id}`;
    open.textContent = 'Открыть';
    const del = document.createElement('button');
    del.className = 'btn ghost';
    del.textContent = S.adminDelete;
    del.addEventListener('click', async () => {
      if (!confirm(S.adminDeleteConfirm(`курс «${c.title}»`))) return;
      clearErrors(app);
      try {
        await api.admin.courseDelete(c.id);
        renderAdminCourses(app);
      } catch (err) {
        showErrors(app, err);
      }
    });
    row.append(grow, open, del);
    list.append(row);
  }
  app.append(list);
}
