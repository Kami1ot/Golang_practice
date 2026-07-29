// Обёртки над REST API. 401 — тихий редирект на вход; остальные ошибки — тост.

async function request(method, url, body, opts = {}) {
  const fetchOpts = { method, headers: {} };
  if (body !== undefined && !(body instanceof FormData)) {
    fetchOpts.headers['Content-Type'] = 'application/json';
    fetchOpts.body = JSON.stringify(body);
  } else if (body instanceof FormData) {
    fetchOpts.body = body;
  }
  let resp;
  try {
    resp = await fetch(url, fetchOpts);
  } catch (e) {
    toast('Сервер недоступен — запущен ли go run ./cmd/server?');
    throw e;
  }
  if (resp.status === 401 && !opts.allow401) {
    // Сессии нет или протухла: молча уводим на вход.
    if (!location.hash.startsWith('#/login')) location.hash = '#/login';
    const err = new Error('требуется вход');
    err.status = 401;
    throw err;
  }
  if (resp.status === 204) return null;
  const data = await resp.json().catch(() => ({}));
  if (!resp.ok) {
    const err = new Error(data.error || `HTTP ${resp.status}`);
    err.status = resp.status;
    err.payload = data;
    if (![401, 403, 422].includes(resp.status) && !opts.quiet) toast(err.message);
    throw err;
  }
  return data;
}

export const api = {
  // auth
  me: () => request('GET', '/api/auth/me', undefined, { allow401: true }),
  register: (username, password) => request('POST', '/api/auth/register', { username, password }, { quiet: true }),
  login: (username, password) => request('POST', '/api/auth/login', { username, password }, { quiet: true }),
  logout: () => request('POST', '/api/auth/logout'),

  // курсы и прохождение
  courses: () => request('GET', '/api/courses'),
  tree: (course) => request('GET', `/api/courses/${course}/tree`),
  level: (course, level) => request('GET', `/api/courses/${course}/levels/${level}`),
  submitQuiz: (course, level, answers) =>
    request('POST', `/api/courses/${course}/levels/${level}/quiz`, { answers }),
  runTask: (course, level, task, code) =>
    request('POST', `/api/courses/${course}/levels/${level}/tasks/${task}/run`, { code }),
  saveDraft: (course, level, task, code) =>
    request('PUT', `/api/courses/${course}/levels/${level}/tasks/${task}/draft`, { code }),

  // заметки к уровням
  note: (course, level) => request('GET', `/api/courses/${course}/levels/${level}/note`),
  saveNote: (course, level, body) =>
    request('PUT', `/api/courses/${course}/levels/${level}/note`, { body }),
  notes: () => request('GET', '/api/notes'),

  // кабинет
  profile: () => request('GET', '/api/profile'),

  // чат-наставник
  chatHistory: (course, level) => request('GET', `/api/courses/${course}/levels/${level}/chat`),
  chatSend: (course, level, message, taskId) =>
    request('POST', `/api/courses/${course}/levels/${level}/chat`, { message, taskId }, { quiet: true }),
  chatClear: (course, level) => request('DELETE', `/api/courses/${course}/levels/${level}/chat`),

  // админка (все ошибки валидации обрабатываются на месте, без тостов)
  admin: {
    courses: () => request('GET', '/api/admin/courses'),
    courseCreate: (c) => request('POST', '/api/admin/courses', c, { quiet: true }),
    course: (id) => request('GET', `/api/admin/courses/${id}`),
    courseUpdate: (id, c) => request('PUT', `/api/admin/courses/${id}`, c, { quiet: true }),
    courseDelete: (id) => request('DELETE', `/api/admin/courses/${id}`, undefined, { quiet: true }),
    levelCreate: (course, meta) => request('POST', `/api/admin/courses/${course}/levels`, meta, { quiet: true }),
    level: (course, level) => request('GET', `/api/admin/courses/${course}/levels/${level}`),
    levelUpdate: (course, level, meta) =>
      request('PUT', `/api/admin/courses/${course}/levels/${level}`, meta, { quiet: true }),
    levelDelete: (course, level) =>
      request('DELETE', `/api/admin/courses/${course}/levels/${level}`, undefined, { quiet: true }),
    lessonPut: (course, level, markdown) =>
      request('PUT', `/api/admin/courses/${course}/levels/${level}/lesson`, { markdown }, { quiet: true }),
    quizPut: (course, level, quiz) =>
      request('PUT', `/api/admin/courses/${course}/levels/${level}/quiz`, quiz, { quiet: true }),
    quizDelete: (course, level) =>
      request('DELETE', `/api/admin/courses/${course}/levels/${level}/quiz`, undefined, { quiet: true }),
    taskSave: (course, level, task) =>
      request('POST', `/api/admin/courses/${course}/levels/${level}/tasks`, task, { quiet: true }),
    taskDelete: (course, level, id) =>
      request('DELETE', `/api/admin/courses/${course}/levels/${level}/tasks/${id}`, undefined, { quiet: true }),
    imageUpload: (course, level, file) => {
      const fd = new FormData();
      fd.append('file', file);
      return request('POST', `/api/admin/courses/${course}/levels/${level}/images`, fd, { quiet: true });
    },
    imageDelete: (course, level, name) =>
      request('DELETE', `/api/admin/courses/${course}/levels/${level}/images/${name}`, undefined, { quiet: true }),
    reload: () => request('POST', '/api/reload', undefined, { quiet: true }),
  },
};

let toastTimer;
export function toast(msg) {
  const el = document.getElementById('toast');
  el.textContent = msg;
  el.hidden = false;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => { el.hidden = true; }, 4000);
}

// Всплывашка «Новая ачивка!» — вызывается из quiz.js / tasks.js.
export function showAchievements(list) {
  if (!list?.length) return;
  for (const [i, a] of list.entries()) {
    setTimeout(() => {
      const el = document.createElement('div');
      el.className = 'achievement-pop';
      el.innerHTML = `<span class="icon"></span><div><b>Новая ачивка!</b><div class="t"></div></div>`;
      el.querySelector('.icon').textContent = a.icon;
      el.querySelector('.t').textContent = a.title;
      document.body.append(el);
      setTimeout(() => el.classList.add('gone'), 4200);
      setTimeout(() => el.remove(), 4800);
    }, i * 700);
  }
}
