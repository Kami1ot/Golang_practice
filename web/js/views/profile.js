// Личный кабинет: статистика + сетка ачивок + заметки к уровням.
import { api } from '../api.js';
import { S } from '../strings.js';

export async function renderProfile(app) {
  const [p, notes] = await Promise.all([api.profile(), api.notes().catch(() => [])]);
  app.innerHTML = '';

  const screen = document.createElement('div');
  screen.className = 'screen';

  const head = document.createElement('div');
  head.className = 'screen-head';
  const headText = document.createElement('div');
  const h1 = document.createElement('h1');
  h1.textContent = p.user.username;
  const sub = document.createElement('p');
  const since = new Date(p.registeredAt).toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' });
  sub.textContent = `${p.user.role === 'admin' ? 'администратор · ' : ''}${S.registeredAt(since)}`;
  headText.append(h1, sub);
  head.append(headText);

  const stats = document.createElement('div');
  stats.className = 'profile-stats';
  const statDefs = [
    [p.stats.xp, S.statXP, 'gold'],
    [p.stats.levelsCompleted, S.statLevels, ''],
    [p.stats.tasksDone, S.statTasks, ''],
    [p.stats.quizzesPassed, S.statQuizzes, ''],
  ];
  for (const [num, label, cls] of statDefs) {
    const cell = document.createElement('div');
    cell.className = 'stat-cell';
    const n = document.createElement('div');
    n.className = 'stat-num ' + cls;
    n.textContent = num;
    const l = document.createElement('div');
    l.className = 'stat-label';
    l.textContent = label;
    cell.append(n, l);
    stats.append(cell);
  }

  const achTitle = document.createElement('h2');
  achTitle.className = 'profile-section-title';
  const unlocked = p.achievements.filter(a => a.grantedAt).length;
  achTitle.textContent = `${S.achievementsTitle} · ${unlocked}/${p.achievements.length}`;

  const grid = document.createElement('div');
  grid.className = 'achievement-grid';
  for (const a of p.achievements) {
    const card = document.createElement('div');
    card.className = 'achievement-card' + (a.grantedAt ? ' unlocked' : '');
    const icon = document.createElement('div');
    icon.className = 'ach-icon';
    icon.textContent = a.icon;
    const title = document.createElement('div');
    title.className = 'ach-title';
    title.textContent = a.title;
    const desc = document.createElement('div');
    desc.className = 'ach-desc';
    desc.textContent = a.description;
    card.append(icon, title, desc);
    if (a.grantedAt) {
      const date = document.createElement('div');
      date.className = 'ach-date';
      date.textContent = new Date(a.grantedAt).toLocaleDateString('ru-RU');
      card.append(date);
    }
    grid.append(card);
  }

  const body = document.createElement('div');
  body.className = 'profile-body';
  body.append(stats, achTitle, grid);

  // Заметки к уровням: имя = {курс} — {модуль} — {уровень}, ссылка на уровень.
  if (notes.length) {
    const notesTitle = document.createElement('h2');
    notesTitle.className = 'profile-section-title';
    notesTitle.textContent = `${S.notesTitle} · ${notes.length}`;

    const list = document.createElement('div');
    list.className = 'note-list';
    for (const n of notes) {
      const row = document.createElement('a');
      row.className = 'note-row';
      row.href = `#/course/${n.courseId}/level/${n.levelId}`;
      const name = document.createElement('div');
      name.className = 'note-name';
      name.textContent = [n.courseTitle, n.group, n.levelTitle].filter(Boolean).join(' — ');
      const date = document.createElement('div');
      date.className = 'note-date';
      date.textContent = new Date(n.updatedAt).toLocaleDateString('ru-RU');
      const snippet = document.createElement('div');
      snippet.className = 'note-snippet';
      snippet.textContent = n.body.split('\n').find(l => l.trim()) || '';
      row.append(name, date, snippet);
      list.append(row);
    }
    body.append(notesTitle, list);
  }

  screen.append(head, body);
  app.append(screen);
}
