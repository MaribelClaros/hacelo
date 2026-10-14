import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { idbGet, idbSet } from '../idb';
import { enablePushNotifications } from '../push';

// One-time screen: paste the same API_TOKEN configured in the backend (see
// README) and enable notifications. With a single user, nothing fancier is
// needed.
export default function Settings() {
  const [token, setToken] = useState('');
  const [saved, setSaved] = useState(false);
  const [pushStatus, setPushStatus] = useState<'idle' | 'loading' | 'ok' | 'error'>('idle');
  const [pushError, setPushError] = useState('');

  useEffect(() => {
    idbGet('apiToken').then((t) => {
      if (t) setToken(t);
    });
  }, []);

  return (
    <div>
      <div className="top-bar">
        <Link to="/" aria-label="Volver">←</Link>
        <h2 style={{ margin: 0 }}>Ajustes</h2>
        <div style={{ width: 16 }} />
      </div>

      <label htmlFor="token">Token de la API</label>
      <input
        id="token"
        type="password"
        value={token}
        onChange={(e) => {
          setToken(e.target.value);
          setSaved(false);
        }}
        placeholder="El mismo valor que API_TOKEN en el backend"
        style={{ marginBottom: 8 }}
      />
      <button
        type="button"
        onClick={async () => {
          await idbSet('apiToken', token);
          setSaved(true);
        }}
        style={{ marginBottom: 28 }}
      >
        {saved ? 'Guardado' : 'Guardar token'}
      </button>

      <label>Notificaciones</label>
      <button
        type="button"
        className="primary"
        style={{ width: '100%' }}
        disabled={pushStatus === 'loading'}
        onClick={async () => {
          setPushStatus('loading');
          setPushError('');
          try {
            await enablePushNotifications();
            setPushStatus('ok');
          } catch (err) {
            setPushStatus('error');
            setPushError((err as Error).message);
          }
        }}
      >
        {pushStatus === 'ok' ? 'Notificaciones activadas' : 'Activar notificaciones'}
      </button>
      {pushStatus === 'error' && <p className="muted" style={{ marginTop: 8 }}>{pushError}</p>}
    </div>
  );
}
