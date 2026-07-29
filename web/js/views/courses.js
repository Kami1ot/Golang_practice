// Экран списка курсов: витрина с поиском, фильтром по тегам и статистикой.
import { api } from '../api.js';
import { S } from '../strings.js';

export async function renderCourses(app) {
  const courses = await api.courses();
  app.innerHTML = '';

  const h1 = document.createElement('h1');
  h1.className = 'page-title';
  h1.textContent = S.coursesTitle;
  const sub = document.createElement('p');
  sub.className = 'page-sub';
  sub.textContent = S.coursesSub;

  // Тулбар: поиск + чипы тегов (собираются со всех курсов, в порядке появления).
  const toolbar = document.createElement('div');
  toolbar.className = 'courses-toolbar';
  const search = document.createElement('input');
  search.type = 'search';
  search.className = 'search-input';
  search.placeholder = S.coursesSearchPlaceholder;
  toolbar.append(search);

  const allTags = [];
  for (const c of courses) {
    for (const tag of c.tags || []) {
      if (!allTags.includes(tag)) allTags.push(tag);
    }
  }
  const tagRow = document.createElement('div');
  tagRow.className = 'tag-row';
  const activeTags = new Set();
  for (const tag of allTags) {
    const chip = document.createElement('button');
    chip.type = 'button';
    chip.className = 'tag-chip';
    chip.textContent = tag;
    chip.setAttribute('aria-pressed', 'false');
    chip.addEventListener('click', () => {
      if (activeTags.has(tag)) activeTags.delete(tag);
      else activeTags.add(tag);
      chip.setAttribute('aria-pressed', activeTags.has(tag) ? 'true' : 'false');
      applyFilter();
    });
    tagRow.append(chip);
  }
  if (allTags.length) toolbar.append(tagRow);

  const grid = document.createElement('div');
  grid.className = 'course-grid';

  const empty = document.createElement('p');
  empty.className = 'courses-empty';
  empty.textContent = S.coursesEmpty;
  empty.hidden = true;

  app.append(h1, sub, toolbar, grid, empty);

  // Карточки создаются один раз; фильтр только переключает hidden.
  const cards = [];
  courses.forEach((c, i) => {
    const card = document.createElement('a');
    card.className = 'course-card';
    card.href = `#/course/${c.id}`;

    const eyebrow = document.createElement('div');
    eyebrow.className = 'eyebrow';
    eyebrow.textContent = `курс ${String(i + 1).padStart(2, '0')}`;

    const title = document.createElement('h2');
    title.textContent = c.title;
    const desc = document.createElement('p');
    desc.textContent = c.description;

    const tags = document.createElement('div');
    tags.className = 'course-tags';
    for (const tag of c.tags || []) {
      const t = document.createElement('span');
      t.className = 'chip';
      t.textContent = tag;
      tags.append(t);
    }

    const bar = document.createElement('div');
    bar.className = 'bar';
    const fill = document.createElement('i');
    fill.style.width = c.levelsTotal ? `${(100 * c.levelsCompleted / c.levelsTotal).toFixed(0)}%` : '0%';
    bar.append(fill);

    const meta = document.createElement('div');
    meta.className = 'course-meta';
    const lv = document.createElement('span');
    lv.textContent = S.levelsOf(c.levelsCompleted, c.levelsTotal);
    const xp = document.createElement('span');
    xp.className = 'xp';
    xp.textContent = `${c.xpEarned}/${c.xpTotal} XP`;
    meta.append(lv, xp);

    const stats = document.createElement('div');
    stats.className = 'course-stats';
    const done = document.createElement('span');
    done.className = 'done';
    done.textContent = S.statsCompleted(c.usersCompleted || 0);
    const dot = document.createElement('span');
    dot.className = 'sep';
    dot.textContent = '·';
    const active = document.createElement('span');
    active.className = 'active';
    active.textContent = S.statsInProgress(c.usersInProgress || 0);
    stats.append(done, dot, active);

    card.append(eyebrow, title, desc, ...(c.tags?.length ? [tags] : []), bar, meta, stats);
    grid.append(card);

    const haystack = [c.title, c.description, ...(c.tags || [])].join('\n').toLowerCase();
    cards.push({ card, haystack, tags: c.tags || [] });
  });

  function applyFilter() {
    const q = search.value.trim().toLowerCase();
    let visible = 0;
    for (const { card, haystack, tags } of cards) {
      const matchQuery = !q || haystack.includes(q);
      const matchTags = [...activeTags].every(t => tags.includes(t));
      card.hidden = !(matchQuery && matchTags);
      if (!card.hidden) visible++;
    }
    empty.hidden = visible > 0;
  }
  search.addEventListener('input', applyFilter);
}
