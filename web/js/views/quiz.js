// Вкладка «Тест»: рендер вопросов по типам, проверка, объяснения, повтор.
import { api, showAchievements } from '../api.js';
import { S } from '../strings.js';
import { renderMD } from '../md.js';

const TYPE_LABEL = {
  single: S.qTypeSingle,
  multi: S.qTypeMulti,
  output: S.qTypeOutput,
  blank: S.qTypeBlank,
};

export function renderQuiz(panel, level, ctx) {
  panel.innerHTML = '';
  if (!level.quiz) {
    const p = document.createElement('p');
    p.className = 'page-sub';
    p.textContent = S.quizEmpty;
    panel.append(p);
    return;
  }

  const quiz = document.createElement('div');
  quiz.className = 'quiz';
  panel.append(quiz);

  if (level.quizPassed) {
    const note = document.createElement('div');
    note.className = 'quiz-passed-note';
    note.textContent = `✓ ${S.quizPassed}`;
    quiz.append(note);
  }

  const cards = level.quiz.questions.map((q, i) => buildQuestion(q, i));
  cards.forEach(c => quiz.append(c.el));

  const footer = document.createElement('div');
  footer.className = 'quiz-footer';
  const btn = document.createElement('button');
  btn.className = 'btn primary';
  btn.textContent = S.quizCheck;
  const score = document.createElement('span');
  score.className = 'quiz-score';
  footer.append(btn, score);
  quiz.append(footer);

  btn.addEventListener('click', async () => {
    if (btn.dataset.mode === 'retry') {
      renderQuiz(panel, level, ctx); // свежий бланк
      return;
    }
    const answers = {};
    for (const c of cards) answers[c.q.id] = c.value();

    btn.disabled = true;
    let res;
    try {
      res = await api.submitQuiz(ctx.courseId, ctx.levelId, answers);
    } finally {
      btn.disabled = false;
    }

    let correct = 0;
    for (const c of cards) {
      const r = res.results[c.q.id];
      if (!r) continue;
      if (r.correct) correct++;
      c.showResult(r);
    }
    const total = cards.length;
    score.innerHTML = '';
    score.append(S.quizScore('', '').slice(0, 0)); // сбрасываем
    const b = document.createElement('b');
    b.className = res.passed ? 'good' : 'bad';
    b.textContent = `${correct}/${total}`;
    score.append(`результат: `, b);
    if (!res.passed) {
      const need = Math.ceil(level.quiz.passScore * total);
      score.append(S.quizNeed(`${need}/${total}`));
      btn.textContent = S.quizRetry;
      btn.dataset.mode = 'retry';
    } else {
      btn.remove();
      level.quizPassed = true;
      ctx.quizTick(true);
      showAchievements(res.newAchievements);
      if (res.levelCompleted) ctx.levelCompleted();
    }
  });
}

function buildQuestion(q, index) {
  const card = document.createElement('div');
  card.className = 'q-card';

  const num = document.createElement('div');
  num.className = 'q-num';
  num.textContent = `${S.questionNum(index + 1)} · ${TYPE_LABEL[q.type] || q.type}`;
  card.append(num);

  const prompt = document.createElement('div');
  prompt.className = 'q-prompt';
  prompt.append(renderMD(q.prompt));
  card.append(prompt);

  if (q.code) {
    const pre = document.createElement('pre');
    const code = document.createElement('code');
    code.innerHTML = hljs.highlight(q.code, { language: 'go' }).value;
    pre.append(code);
    card.append(pre);
  }

  let value, showPicked;
  if (q.type === 'single' || q.type === 'multi') {
    const inputs = [];
    (q.options || []).forEach((opt, i) => {
      const label = document.createElement('label');
      label.className = 'opt';
      const input = document.createElement('input');
      input.type = q.type === 'single' ? 'radio' : 'checkbox';
      input.name = `q-${q.id}`;
      input.value = i;
      const span = document.createElement('span');
      span.append(renderMD(opt));
      label.append(input, span);
      card.append(label);
      inputs.push({ input, label });
    });
    value = () => {
      if (q.type === 'single') {
        const picked = inputs.find(x => x.input.checked);
        return picked ? Number(picked.input.value) : -1;
      }
      return inputs.filter(x => x.input.checked).map(x => Number(x.input.value));
    };
    showPicked = (correct) => {
      inputs.forEach(({ input, label }) => {
        input.disabled = true;
        if (input.checked) label.classList.add(correct ? 'picked-right' : 'picked-wrong');
      });
    };
  } else {
    // output-вопросы бывают многострочными — textarea; blank — однострочный input.
    const input = document.createElement(q.type === 'output' ? 'textarea' : 'input');
    input.className = 'answer-input';
    if (q.type === 'output') input.rows = 2;
    input.placeholder = q.type === 'output' ? S.outputPlaceholder : S.blankPlaceholder;
    input.autocomplete = 'off';
    input.spellcheck = false;
    card.append(input);
    value = () => input.value;
    showPicked = (correct) => {
      input.disabled = true;
      input.style.borderColor = correct ? 'var(--success)' : 'var(--danger)';
    };
  }

  const showResult = (r) => {
    card.classList.remove('correct', 'wrong');
    card.classList.add(r.correct ? 'correct' : 'wrong');
    showPicked(r.correct);
    card.querySelector('.explain')?.remove();
    if (r.explanation) {
      const ex = document.createElement('div');
      ex.className = `explain ${r.correct ? 'good' : 'bad'}`;
      const mark = document.createElement('span');
      mark.className = 'mark';
      mark.textContent = r.correct ? '✓' : '✕';
      const body = document.createElement('div');
      body.className = 'body';
      body.append(renderMD(r.explanation));
      ex.append(mark, body);
      card.append(ex);
    }
  };

  return { q, el: card, value, showResult };
}
