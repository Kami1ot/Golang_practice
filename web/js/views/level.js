// Экран уровня. Широкий экран — workspace-режим (две панели, IDE, заметки),
// узкий — прежние вкладки Урок / Тест / Задачи (fallback).
import { api } from '../api.js';
import { S } from '../strings.js';
import { renderMD } from '../md.js';
import { renderQuiz } from './quiz.js';
import { renderTasks } from './tasks.js';
import { mountChat } from './chat.js';
import { renderLevelWorkspace } from './workspace.js';

const wideMQ = window.matchMedia('(min-width: 1100px)');

export async function renderLevel(app, courseId, levelId) {
  let level;
  try {
    level = await api.level(courseId, levelId);
  } catch (e) {
    if (e.status === 403) {
      app.innerHTML = '';
      const box = document.createElement('div');
      box.className = 'error-box';
      box.innerHTML = `<h2></h2><p></p><p style="margin-top:12px"><a></a></p>`;
      box.querySelector('h2').textContent = S.lockedTitle;
      box.querySelector('p').textContent = S.lockedBody;
      const a = box.querySelector('a');
      a.href = `#/course/${courseId}`;
      a.textContent = S.backToTree;
      app.append(box);
      return null;
    }
    throw e;
  }

  // Название курса для хлебных крошек берём из дерева (кэш не критичен).
  const tree = await api.tree(courseId);

  app.classList.remove('workspace');
  const info = wideMQ.matches
    ? renderLevelWorkspace(app, courseId, levelId, level, tree)
    : renderLevelTabs(app, courseId, levelId, level, tree);

  // Пересечение брейкпоинта — перерисовываем страницу в другом режиме.
  // Слушаем resize окна, а не событие matchMedia: последнее не срабатывает
  // при эмуляции вьюпорта devtools. Слушатель снимает себя сам.
  const marker = app.firstElementChild;
  const renderedWide = wideMQ.matches;
  const onResize = () => {
    if (!marker || !app.contains(marker)) {
      window.removeEventListener('resize', onResize);
      return;
    }
    if (wideMQ.matches !== renderedWide) {
      window.removeEventListener('resize', onResize);
      renderLevel(app, courseId, levelId);
    }
  };
  window.addEventListener('resize', onResize);

  return info;
}

// ── Fallback: вкладки Урок / Тест / Задачи (узкие экраны) ───────────────────

function renderLevelTabs(app, courseId, levelId, level, tree) {
  app.innerHTML = '';
  const screen = document.createElement('div');
  screen.className = 'screen level-screen';

  const head = document.createElement('div');
  head.className = 'level-head';
  const eyebrow = document.createElement('div');
  eyebrow.className = 'eyebrow';
  eyebrow.textContent = level.optional ? S.bonusEyebrow(level.xp) : S.levelEyebrow(level.xp);
  const h1 = document.createElement('h1');
  h1.textContent = level.title;
  const sub = document.createElement('p');
  sub.className = 'sub';
  sub.textContent = level.summary;
  head.append(eyebrow, h1, sub);

  // Вкладки.
  const tabs = document.createElement('div');
  tabs.className = 'tabs';
  tabs.setAttribute('role', 'tablist');
  const panels = {};
  const tabBtns = {};
  const tabDefs = [['lesson', S.tabLesson], ['quiz', S.tabQuiz], ['tasks', S.tabTasks]];

  for (const [key, label] of tabDefs) {
    const btn = document.createElement('button');
    btn.className = 'tab';
    btn.setAttribute('role', 'tab');
    btn.textContent = label;
    btn.addEventListener('click', () => activate(key, true));
    tabs.append(btn);
    tabBtns[key] = btn;

    const panel = document.createElement('div');
    panel.className = `tab-panel tab-panel-${key}`;
    panel.hidden = true;
    panels[key] = panel;
  }

  function activate(key, scroll = false) {
    for (const [k, btn] of Object.entries(tabBtns)) {
      btn.classList.toggle('active', k === key);
      panels[k].hidden = k !== key;
    }
    // CodeMirror требует refresh после показа скрытой панели.
    if (key === 'tasks') {
      panels.tasks.querySelectorAll('.CodeMirror').forEach(el => el.CodeMirror?.refresh());
    }
    if (scroll) window.scrollTo({ top: 0 });
  }

  const chatCtl = mountChat(screen, courseId, levelId, level.tasks);

  const ctx = {
    courseId,
    levelId,
    askMentor: (taskId) => chatCtl.open(taskId),
    quizTick(passed) {
      decorateQuizTab(tabBtns.quiz, level, passed);
    },
    tasksTick(done, total) {
      decorateTasksTab(tabBtns.tasks, done, total);
    },
    levelCompleted() {
      if (screen.querySelector('.win-banner')) return;
      const banner = document.createElement('div');
      banner.className = 'win-banner';
      banner.innerHTML = `<span class="big">🎉</span><div><h3></h3><p></p></div>`;
      banner.querySelector('h3').textContent = S.levelDone(level.xp);
      banner.querySelector('p').textContent = S.levelDoneUnlocked;
      const btn = document.createElement('a');
      btn.className = 'btn primary';
      btn.href = `#/course/${courseId}`;
      btn.textContent = S.toTree;
      banner.append(btn);
      screen.append(banner);
      banner.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    },
  };

  // Наполнение вкладок. Тесту и задачам — свой контейнер: renderQuiz при
  // повторе очищает его целиком, кнопки навигации ниже должны уцелеть.
  const lesson = document.createElement('article');
  lesson.className = 'lesson';
  lesson.append(renderMD(level.lessonMd));
  panels.lesson.append(lesson);

  const quizBody = document.createElement('div');
  panels.quiz.append(quizBody);
  renderQuiz(quizBody, level, ctx);

  const tasksBody = document.createElement('div');
  panels.tasks.append(tasksBody);
  renderTasks(tasksBody, level, ctx);

  // Кнопки «дальше/назад» внизу вкладок: урок → тест → задачи (пустые шаги пропускаем).
  const steps = ['lesson'];
  if (level.quiz) steps.push('quiz');
  if (level.tasks.length) steps.push('tasks');
  const prevLabel = { lesson: S.navToLesson, quiz: S.navToQuizBack };
  const nextLabel = { quiz: S.navToQuiz, tasks: S.navToTasks };
  if (steps.length > 1) {
    steps.forEach((key, i) => {
      const nav = document.createElement('div');
      nav.className = 'panel-nav';
      if (i > 0) nav.append(navBtn('btn ghost', prevLabel[steps[i - 1]], () => activate(steps[i - 1], true)));
      if (i < steps.length - 1) nav.append(navBtn('btn primary next', nextLabel[steps[i + 1]], () => activate(steps[i + 1], true)));
      panels[key].append(nav);
    });
  }

  decorateQuizTab(tabBtns.quiz, level, level.quizPassed);
  const doneTasks = level.tasks.filter(t => t.required && t.completed).length;
  const totalTasks = level.tasks.filter(t => t.required).length;
  decorateTasksTab(tabBtns.tasks, doneTasks, totalTasks);

  screen.append(head, tabs, panels.lesson, panels.quiz, panels.tasks);
  app.append(screen);
  activate('lesson');

  // Уровень уже был завершён ранее — баннер не показываем, всё и так зелёное.
  return { courseTitle: tree.course.title, levelTitle: level.title };
}

function navBtn(cls, label, onClick) {
  const b = document.createElement('button');
  b.className = cls;
  b.textContent = label;
  b.addEventListener('click', onClick);
  return b;
}

function decorateQuizTab(btn, level, passed) {
  btn.querySelectorAll('.tick, .count').forEach(el => el.remove());
  if (!level.quiz) return;
  if (passed) {
    const tick = document.createElement('span');
    tick.className = 'tick';
    tick.textContent = '✓';
    btn.append(tick);
  } else {
    const count = document.createElement('span');
    count.className = 'count';
    count.textContent = `${level.quiz.questions.length}`;
    btn.append(count);
  }
}

function decorateTasksTab(btn, done, total) {
  btn.querySelectorAll('.tick, .count').forEach(el => el.remove());
  if (total === 0) return;
  if (done >= total) {
    const tick = document.createElement('span');
    tick.className = 'tick';
    tick.textContent = '✓';
    btn.append(tick);
  } else {
    const count = document.createElement('span');
    count.className = 'count';
    count.textContent = `${done}/${total}`;
    btn.append(count);
  }
}
