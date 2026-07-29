// Экран входа / регистрации.
import { api } from '../api.js';
import { S } from '../strings.js';
import { setCurrentUser } from '../app.js';

export function renderLogin(app) {
  app.innerHTML = '';
  let mode = 'login';

  const card = document.createElement('div');
  card.className = 'login-card';

  const logo = document.createElement('div');
  logo.className = 'login-logo';
  logo.innerHTML = `go<span>practice</span>`;
  const h1 = document.createElement('h1');
  h1.textContent = S.loginTitle;
  const sub = document.createElement('p');
  sub.className = 'login-sub';
  sub.textContent = S.loginSub;

  const form = document.createElement('form');
  const nameInput = document.createElement('input');
  nameInput.placeholder = S.username;
  nameInput.autocomplete = 'username';
  nameInput.required = true;
  const passInput = document.createElement('input');
  passInput.type = 'password';
  passInput.placeholder = S.password;
  passInput.autocomplete = 'current-password';
  passInput.required = true;

  const errBox = document.createElement('div');
  errBox.className = 'login-error';
  errBox.hidden = true;

  const submit = document.createElement('button');
  submit.className = 'btn primary';
  submit.type = 'submit';

  const firstNote = document.createElement('p');
  firstNote.className = 'login-note';
  firstNote.textContent = S.firstUserNote;

  const switcher = document.createElement('button');
  switcher.type = 'button';
  switcher.className = 'login-switch';

  function applyMode() {
    submit.textContent = mode === 'login' ? S.loginBtn : S.registerBtn;
    switcher.textContent = mode === 'login' ? S.switchToRegister : S.switchToLogin;
    passInput.autocomplete = mode === 'login' ? 'current-password' : 'new-password';
    firstNote.hidden = mode === 'login';
    errBox.hidden = true;
  }
  switcher.addEventListener('click', () => {
    mode = mode === 'login' ? 'register' : 'login';
    applyMode();
  });
  applyMode();

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    submit.disabled = true;
    errBox.hidden = true;
    try {
      const call = mode === 'login' ? api.login : api.register;
      const user = await call(nameInput.value.trim(), passInput.value);
      setCurrentUser(user);
      location.hash = '#/courses';
    } catch (err) {
      errBox.textContent = err.message;
      errBox.hidden = false;
    } finally {
      submit.disabled = false;
    }
  });

  form.append(nameInput, passInput, errBox, submit);
  card.append(logo, h1, sub, form, firstNote, switcher);
  app.append(card);
  nameInput.focus();
}
