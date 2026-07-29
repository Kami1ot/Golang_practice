// Рендер Markdown: marked + highlight.js + GitHub-callouts [!NOTE]/[!WARNING].

marked.setOptions({
  highlight(code, lang) {
    if (lang && hljs.getLanguage(lang)) {
      return hljs.highlight(code, { language: lang }).value;
    }
    return hljs.highlightAuto(code).value;
  },
  langPrefix: 'hljs language-',
  mangle: false,
  headerIds: false,
});

export function renderMD(md) {
  const html = marked.parse(md || '');
  const box = document.createElement('div');
  box.innerHTML = html;
  transformCallouts(box);
  return box;
}

// > [!NOTE] / > [!WARNING] в начале цитаты -> цветной callout.
function transformCallouts(root) {
  for (const bq of [...root.querySelectorAll('blockquote')]) {
    const first = bq.querySelector('p');
    if (!first) continue;
    const m = first.textContent.match(/^\s*\[!(NOTE|WARNING)\]\s*/);
    if (!m) continue;
    const kind = m[1];
    first.innerHTML = first.innerHTML.replace(/^\s*\[!(NOTE|WARNING)\]\s*(<br\s*\/?>)?/, '');
    const callout = document.createElement('div');
    callout.className = kind === 'WARNING' ? 'callout warn' : 'callout';
    const ico = document.createElement('span');
    ico.className = 'ico';
    ico.textContent = kind === 'WARNING' ? '!' : 'i';
    const body = document.createElement('div');
    body.append(...bq.childNodes);
    callout.append(ico, body);
    bq.replaceWith(callout);
  }
}
