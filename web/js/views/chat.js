// Чат-наставник: выдвижная панель на странице уровня.
import { api } from '../api.js';
import { S } from '../strings.js';
import { renderMD } from '../md.js';

// mountChat вызывается из level.js; возвращает {open(taskId?)} или null, если чат выключен.
export function mountChat(container, courseId, levelId, tasks) {
  let enabled = null; // выясняется лениво
  let taskContext = ''; // id задачи, о которой идёт разговор

  const drawer = document.createElement('aside');
  drawer.className = 'chat-drawer';
  drawer.hidden = true;

  const head = document.createElement('div');
  head.className = 'chat-head';
  const title = document.createElement('b');
  title.textContent = `🎓 ${S.mentor}`;
  const clearBtn = document.createElement('button');
  clearBtn.className = 'chat-clear';
  clearBtn.textContent = S.mentorClear;
  const closeBtn = document.createElement('button');
  closeBtn.className = 'chat-close';
  closeBtn.textContent = '✕';
  closeBtn.setAttribute('aria-label', 'закрыть');
  head.append(title, clearBtn, closeBtn);

  const ctxBar = document.createElement('div');
  ctxBar.className = 'chat-context';
  ctxBar.hidden = true;

  const log = document.createElement('div');
  log.className = 'chat-log';

  const form = document.createElement('form');
  form.className = 'chat-form';
  const input = document.createElement('textarea');
  input.rows = 2;
  input.placeholder = S.mentorPlaceholder;
  const send = document.createElement('button');
  send.className = 'btn primary';
  send.type = 'submit';
  send.textContent = S.mentorSend;
  form.append(input, send);

  drawer.append(head, ctxBar, log, form);
  container.append(drawer);

  const fab = document.createElement('button');
  fab.className = 'chat-fab';
  fab.innerHTML = `🎓 <span>${S.mentor}</span>`;
  fab.hidden = true;
  container.append(fab);

  function addBubble(role, content) {
    const empty = log.querySelector('.chat-empty');
    if (empty) empty.remove();
    const b = document.createElement('div');
    b.className = `chat-bubble ${role}`;
    if (role === 'assistant') b.append(renderMD(content));
    else b.textContent = content;
    log.append(b);
    log.scrollTop = log.scrollHeight;
    return b;
  }

  function setContext(taskId) {
    taskContext = taskId || '';
    const task = tasks.find(t => t.id === taskContext);
    ctxBar.hidden = !task;
    if (task) {
      ctxBar.innerHTML = '';
      ctxBar.append(S.mentorTaskContext(task.title) + ' ');
      const drop = document.createElement('button');
      drop.textContent = '✕';
      drop.addEventListener('click', () => setContext(''));
      ctxBar.append(drop);
    }
  }

  async function load() {
    const hist = await api.chatHistory(courseId, levelId);
    enabled = hist.enabled;
    if (!enabled) return false;
    log.innerHTML = '';
    if (!hist.messages.length) {
      const empty = document.createElement('div');
      empty.className = 'chat-empty';
      empty.textContent = S.mentorEmpty;
      log.append(empty);
    }
    for (const m of hist.messages) addBubble(m.role, m.content);
    return true;
  }

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const msg = input.value.trim();
    if (!msg) return;
    input.value = '';
    addBubble('user', msg);
    send.disabled = true;
    input.disabled = true;
    const typing = addBubble('assistant', S.mentorThinking);
    typing.classList.add('typing');
    try {
      const res = await api.chatSend(courseId, levelId, msg, taskContext || undefined);
      typing.remove();
      addBubble('assistant', res.reply);
    } catch (err) {
      typing.remove();
      addBubble('assistant', `⚠ ${err.message}`);
    } finally {
      send.disabled = false;
      input.disabled = false;
      input.focus();
    }
  });

  clearBtn.addEventListener('click', async () => {
    await api.chatClear(courseId, levelId);
    log.innerHTML = '';
    const empty = document.createElement('div');
    empty.className = 'chat-empty';
    empty.textContent = S.mentorEmpty;
    log.append(empty);
  });
  closeBtn.addEventListener('click', () => { drawer.hidden = true; });
  fab.addEventListener('click', () => open());

  async function open(taskId) {
    if (enabled === null) await load();
    if (!enabled) return;
    setContext(taskId);
    drawer.hidden = false;
    input.focus();
  }

  // Ленивая инициализация: узнаём enabled и показываем кнопку.
  load().then(ok => { fab.hidden = !ok; }).catch(() => {});

  return { open };
}
