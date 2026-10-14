import { api } from './api';

function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
  const rawData = atob(base64);
  return Uint8Array.from([...rawData].map((c) => c.charCodeAt(0)));
}

// Asks for notification permission and registers the push subscription with
// the backend. Called once from the settings screen. Errors are shown to the
// user, so their messages are in Spanish.
export async function enablePushNotifications(): Promise<void> {
  if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
    throw new Error('Este navegador no soporta notificaciones push');
  }

  const permission = await Notification.requestPermission();
  if (permission !== 'granted') {
    throw new Error('Permiso de notificaciones denegado');
  }

  const registration = await navigator.serviceWorker.ready;
  const { public_key } = await api.vapidPublicKey();

  const subscription = await registration.pushManager.subscribe({
    userVisibleOnly: true,
    applicationServerKey: urlBase64ToUint8Array(public_key) as BufferSource,
  });

  await api.subscribePush(subscription.toJSON());
}
