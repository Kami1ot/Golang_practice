// Панель ошибок валидации: сервер откатил изменения и вернул точные file:line:col.
import { S } from '../../strings.js';

export function showErrors(host, err) {
  clearErrors(host);
  const box = document.createElement('div');
  box.className = 'validation-errors';
  const title = document.createElement('b');
  const pre = document.createElement('pre');
  if (err.payload?.errors) {
    title.textContent = S.adminValidationErrors;
    pre.textContent = err.payload.errors.join('\n');
  } else {
    title.textContent = 'Ошибка';
    pre.textContent = err.message;
  }
  box.append(title, pre);
  host.prepend(box);
  box.scrollIntoView({ block: 'nearest' });
}

export function clearErrors(host) {
  host.querySelectorAll('.validation-errors').forEach(el => el.remove());
}

// Мигающая пометка «Сохранено ✓» рядом с кнопкой.
export function flashSaved(btn) {
  const note = document.createElement('span');
  note.className = 'admin-save-note';
  note.textContent = S.adminSaved;
  btn.after(note);
  setTimeout(() => note.remove(), 2000);
}

// Утилита: form-поле с подписью.
export function field(labelText, inputEl) {
  const label = document.createElement('label');
  label.append(labelText, inputEl);
  return label;
}

export function textInput(value = '', placeholder = '') {
  const el = document.createElement('input');
  el.value = value;
  el.placeholder = placeholder;
  return el;
}

export function numInput(value = 0) {
  const el = document.createElement('input');
  el.type = 'number';
  el.value = value;
  return el;
}

export function textArea(value = '', rows = 3) {
  const el = document.createElement('textarea');
  el.value = value;
  el.rows = rows;
  return el;
}

// «go, основы, backend» → ["go", "основы", "backend"] (пустые куски отбрасываются).
export function parseTags(value) {
  return value.split(',').map(s => s.trim()).filter(Boolean);
}

export function checkbox(labelText, checked = false) {
  const label = document.createElement('label');
  label.className = 'checkbox';
  const el = document.createElement('input');
  el.type = 'checkbox';
  el.checked = checked;
  label.append(el, labelText);
  return { label, input: el };
}
