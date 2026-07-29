// Точка входа SPA: hash-роутер, авторизация, шапка.
import { api } from './api.js';
import { S } from './strings.js';
import { renderLogin } from './views/login.js';
import { renderCourses } from './views/courses.js';
import { renderTree } from './views/tree.js';
import { renderLevel } from './views/level.js';
import { renderProfile } from './views/profile.js';
import { renderAdminCourses } from './views/admin/courses.js';
import { renderAdminCourse } from './views/admin/course.js';
import { renderAdminLevel } from './views/admin/level.js';

const app = document.getElementById('app');
const crumbs = document.getElementById('crumbs');

// Фактическая высота шапки — для sticky-вкладок уровня (.level-screen .tabs).
const topbar = document.querySelector('.topbar');
const syncTopbarH = () =>
  document.documentElement.style.setProperty('--topbar-h', `${topbar.offsetHeight}px`);
syncTopbarH();
new ResizeObserver(syncTopbarH).observe(topbar);

export let currentUser = null;

export function setCurrentUser(u) {
  currentUser = u;
  renderUserBox();
}

export async function refreshXP() {
  if (!currentUser) return;
  try {
    const courses = await api.courses();
    const xp = courses.reduce((sum, c) => sum + c.xpEarned, 0);
    document.getElementById('xp-pill').textContent = `☆ ${xp} XP`;
  } catch { /* не критично */ }
}

function renderUserBox() {
  const box = document.getElementById('user-box');
  box.innerHTML = '';
  document.getElementById('xp-pill').hidden = !currentUser;
  queueMicrotask(syncTopbarH); // шапка меняет высоту после дорисовки user-box
  if (!currentUser) return;

  if (currentUser.role === 'admin') {
    const adminLink = document.createElement('a');
    adminLink.className = 'topbar-link';
    adminLink.href = '#/admin';
    adminLink.textContent = S.adminLink;
    box.append(adminLink);
  }
  const profile = document.createElement('a');
  profile.className = 'user-badge';
  profile.href = '#/profile';
  profile.title = S.profileLink;
  profile.textContent = currentUser.username;
  const logout = document.createElement('button');
  logout.className = 'topbar-link as-btn';
  logout.textContent = S.logout;
  logout.addEventListener('click', async () => {
    await api.logout().catch(() => {});
    setCurrentUser(null);
    location.hash = '#/login';
  });
  box.append(profile, logout);
}

function setCrumbs(parts) {
  crumbs.innerHTML = '';
  parts.forEach((p, i) => {
    if (i > 0) crumbs.append(' / ');
    if (p.href && i < parts.length - 1) {
      const a = document.createElement('a');
      a.href = p.href;
      a.textContent = p.title;
      crumbs.append(a);
    } else {
      const b = document.createElement('b');
      b.textContent = p.title;
      crumbs.append(b);
    }
  });
}

async function route() {
  const hash = location.hash || '#/courses';
  const parts = hash.replace(/^#\//, '').split('/').filter(Boolean);

  // Гвард: без пользователя — только экран входа.
  if (!currentUser && parts[0] !== 'login') {
    location.hash = '#/login';
    return;
  }
  if (currentUser && parts[0] === 'login') {
    location.hash = '#/courses';
    return;
  }

  // Карта курса и страница уровня — во всю ширину окна (#app.wide снимает кэп 1080px).
  const isCourseMap = parts[0] === 'course' && parts.length === 2;
  const isLevel = parts[0] === 'course' && parts[2] === 'level' && parts.length === 4;
  app.classList.toggle('wide', isCourseMap || isLevel);
  // Workspace-режим ставит сам renderLevel (зависит от ширины экрана).
  app.classList.remove('workspace');

  app.innerHTML = '<div class="loading">Загрузка…</div>';
  window.scrollTo(0, 0);

  try {
    if (parts[0] === 'login') {
      setCrumbs([]);
      renderLogin(app);
    } else if (parts[0] === 'courses' || parts.length === 0) {
      setCrumbs([{ title: S.courseCrumb }]);
      await renderCourses(app);
    } else if (parts[0] === 'profile') {
      setCrumbs([{ title: S.profileLink }]);
      await renderProfile(app);
    } else if (parts[0] === 'admin' && parts.length === 1) {
      setCrumbs([{ title: S.adminLink }]);
      await renderAdminCourses(app);
    } else if (parts[0] === 'admin' && parts[1] === 'course' && parts.length === 3) {
      const info = await renderAdminCourse(app, parts[2]);
      setCrumbs([{ title: S.adminLink, href: '#/admin' }, { title: info.title }]);
    } else if (parts[0] === 'admin' && parts[1] === 'course' && parts[3] === 'level' && parts.length === 5) {
      const info = await renderAdminLevel(app, parts[2], parts[4]);
      setCrumbs([
        { title: S.adminLink, href: '#/admin' },
        { title: info.courseTitle, href: `#/admin/course/${parts[2]}` },
        { title: info.levelTitle },
      ]);
    } else if (parts[0] === 'course' && parts.length === 2) {
      const tree = await renderTree(app, parts[1]);
      setCrumbs([{ title: S.courseCrumb, href: '#/courses' }, { title: tree.course.title }]);
    } else if (parts[0] === 'course' && parts[2] === 'level' && parts.length === 4) {
      const info = await renderLevel(app, parts[1], parts[3]);
      if (info) {
        setCrumbs([
          { title: S.courseCrumb, href: '#/courses' },
          { title: info.courseTitle, href: `#/course/${parts[1]}` },
          { title: info.levelTitle },
        ]);
      }
    } else {
      app.innerHTML = `<div class="error-box"><h2>${S.notFound}</h2></div>`;
    }
  } catch (e) {
    if (e.status === 401) return; // api.js уже увёл на #/login
    if (!app.querySelector('.error-box')) {
      app.innerHTML = `<div class="error-box"><h2>Ошибка</h2><p></p></div>`;
      app.querySelector('p').textContent = e.message;
    }
  }
  refreshXP();
}

window.addEventListener('hashchange', route);

// Загрузка приложения: сначала выясняем, кто мы.
(async () => {
  try {
    setCurrentUser(await api.me());
  } catch {
    setCurrentUser(null);
  }
  route();
})();
