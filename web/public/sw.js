// Hacelo service worker. Plain JS (it bypasses the Vite build) because the
// browser must be able to load it directly from /sw.js.
//
// Two jobs: display the push sent by the backend with its action buttons,
// and handle button clicks even when the app is closed.

const DB_NAME = 'hacelo';
const STORE = 'kv';

function idbGet(key) {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, 1);
    req.onupgradeneeded = () => {
      req.result.createObjectStore(STORE);
    };
    req.onsuccess = () => {
      const tx = req.result.transaction(STORE, 'readonly');
      const getReq = tx.objectStore(STORE).get(key);
      getReq.onsuccess = () => resolve(getReq.result);
      getReq.onerror = () => reject(getReq.error);
    };
    req.onerror = () => reject(req.error);
  });
}

self.addEventListener('install', () => {
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  event.waitUntil(self.clients.claim());
});

// The payload (title, body, button labels) is built by the backend.
self.addEventListener('push', (event) => {
  if (!event.data) return;
  const payload = event.data.json();

  const options = {
    body: payload.body,
    data: { taskId: payload.task_id },
    actions: (payload.actions || []).map((a) => ({ action: a.action, title: a.title })),
    tag: 'hacelo-task-' + payload.task_id,
    renotify: true,
    requireInteraction: true,
  };

  event.waitUntil(self.registration.showNotification(payload.title, options));
});

self.addEventListener('notificationclick', (event) => {
  const taskId = event.notification.data && event.notification.data.taskId;
  const action = event.action;
  event.notification.close();

  if (!taskId) return;

  if (action === 'reschedule') {
    event.waitUntil(self.clients.openWindow('/tasks/' + taskId + '/reschedule'));
    return;
  }

  if (action === 'snooze' || action === 'done') {
    event.waitUntil(callAction(taskId, action));
    return;
  }

  // The notification body was tapped, not a button: open the task list.
  event.waitUntil(self.clients.openWindow('/'));
});

async function callAction(taskId, action) {
  const token = await idbGet('apiToken');
  await fetch('/api/tasks/' + taskId + '/actions/' + action, {
    method: 'POST',
    headers: token ? { Authorization: 'Bearer ' + token } : {},
  });
}
