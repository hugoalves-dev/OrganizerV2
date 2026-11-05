// Generated bindings are checked at build/test time before embedding.
// Keep each notification inside its own helper process to isolate event data.
import { Details, Ready, StartDrag, Action, Close } from './bindings/github.com/vitorhugo-java/organizerv2-popup/popupservice.js';

window.popupDiagnostic?.('main-module', 'main.js and imports loaded');
const el = id => document.getElementById(id);
let timer;
let dragging = false;
const dismiss = () => Close().catch(console.error);

function schedule(seconds) {
  clearTimeout(timer);
  timer = setTimeout(() => {
    if (!dragging) dismiss();
  }, seconds * 1000);
}

async function initialize() {
  window.popupDiagnostic?.('initialize', 'requesting Details binding');
  const details = await Details();
  window.popupDiagnostic?.('initialize', 'Details binding resolved');
  if (!details?.filename || !details?.destination ||
      !Array.isArray(details.choices) || details.choices.length === 0) {
    throw new Error('Incomplete popup payload');
  }

  el('filename').textContent = details.filename;
  el('filename').title = details.filename;
  el('folder').textContent = `Moved to ${details.category} · ${details.destination}`;
  el('folder').title = details.destination;

  const extension = details.filename.split('.').pop()?.toLowerCase();
  el('file-icon').textContent = ['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg'].includes(extension)
    ? '🖼️' : ['zip', 'rar', '7z', 'tar'].includes(extension) ? '🗜️' : '📄';

  const destination = el('destination');
  for (const choice of details.choices) {
    const option = document.createElement('option');
    option.value = choice.id;
    option.textContent = choice.name;
    destination.append(option);
  }
  destination.value = details.choices[0].id;
  destination.disabled = details.choices.length < 2;

  const actions = [
    [details.openLocation, 'open_location', 'Open Location'],
    [details.openFile, 'open_file', 'Open File'],
    [details.confirm, 'confirm', 'Confirm'],
  ];
  for (const [enabled, code, title] of actions) {
    if (!enabled) continue;
    const button = document.createElement('button');
    button.type = 'button';
    button.textContent = title;
    button.onclick = async () => {
      if (dragging) return;
      button.disabled = true;
      try {
        await Action(code, destination.value);
        dismiss();
      } catch (error) {
        el('status').textContent = 'Action failed: ' + String(error);
        button.disabled = false;
      }
    };
    el('actions').append(button);
  }

  el('close').onclick = dismiss;
  if (!details.dragEnabled) {
    el('drag').draggable = false;
    el('status').textContent = 'Drag-out disabled';
  } else {
    el('drag').addEventListener('dragstart', async event => {
      event.preventDefault();
      event.stopPropagation();
      if (dragging) return;
      dragging = true;
      clearTimeout(timer);
      el('status').textContent = 'Native drag in progress…';
      try {
        const result = await StartDrag();
        el('status').textContent = result === 'dropped'
          ? 'Drop accepted. Verify the copy in Explorer.'
          : 'Drag cancelled. Original preserved.';
      } catch (error) {
        el('status').textContent = 'Drag failed: ' + String(error);
      } finally {
        dragging = false;
        schedule(details.duration);
      }
    });
  }

  const panel = document.querySelector('main');
  panel.addEventListener('mouseenter', () => clearTimeout(timer));
  panel.addEventListener('mouseleave', () => {
    if (!dragging) schedule(details.duration);
  });

  // A successful READY means the *complete* UI is already present and visible.
  window.popupDiagnostic?.('initialize', 'requesting Ready binding');
  await Ready();
  window.popupDiagnostic?.('initialize', 'Ready binding resolved');
  schedule(details.duration);
}
initialize().catch(error => {
  console.error('OrganizerV2 popup initialization failed', error);
  window.popupDiagnostic?.('initialize-error', error?.stack || String(error));
  el('status').textContent = 'Popup error: ' + String(error);
  // Do not send READY; the watcher will fall back to native after its deadline.
});
