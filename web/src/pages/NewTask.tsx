import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api, type Urgency } from '../api';

const SNOOZE_OPTIONS = [5, 10, 15, 30];

const URGENCY_OPTIONS: { value: Urgency; label: string }[] = [
  { value: 'low', label: 'Baja' },
  { value: 'medium', label: 'Media' },
  { value: 'high', label: 'Alta' },
];

export default function NewTask() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [description, setDescription] = useState('');
  const [urgency, setUrgency] = useState<Urgency>('medium');
  const [scheduledAt, setScheduledAt] = useState('');
  const [snoozeMinutes, setSnoozeMinutes] = useState(15);

  const { mutate, isPending, isError } = useMutation({
    mutationFn: () =>
      api.createTask({
        description,
        urgency,
        scheduled_at: new Date(scheduledAt).toISOString(),
        snooze_minutes: snoozeMinutes,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      navigate('/');
    },
  });

  return (
    <div>
      <div className="top-bar">
        <Link to="/" aria-label="Volver">←</Link>
        <h2 style={{ margin: 0 }}>Nueva tarea</h2>
        <Link to="/" aria-label="Cerrar">×</Link>
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          mutate();
        }}
      >
        <label htmlFor="description">Descripción</label>
        <textarea
          id="description"
          placeholder="Tenés que ir a comprar verduras"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={3}
          required
          style={{ marginBottom: 18 }}
        />

        <label htmlFor="scheduledAt">Fecha y hora</label>
        <input
          id="scheduledAt"
          type="datetime-local"
          value={scheduledAt}
          onChange={(e) => setScheduledAt(e.target.value)}
          required
          style={{ marginBottom: 18 }}
        />

        <label>Urgencia</label>
        <div style={{ display: 'flex', gap: 8, marginBottom: 18 }}>
          {URGENCY_OPTIONS.map((option) => (
            <button
              type="button"
              key={option.value}
              onClick={() => setUrgency(option.value)}
              style={
                urgency === option.value
                  ? { background: 'var(--accent-bg)', color: 'var(--accent)', borderColor: 'var(--accent)' }
                  : {}
              }
            >
              {option.label}
            </button>
          ))}
        </div>

        <label htmlFor="snooze">Recordame de nuevo en</label>
        <select
          id="snooze"
          value={snoozeMinutes}
          onChange={(e) => setSnoozeMinutes(Number(e.target.value))}
          style={{ marginBottom: 28 }}
        >
          {SNOOZE_OPTIONS.map((minutes) => (
            <option key={minutes} value={minutes}>
              {minutes} min
            </option>
          ))}
        </select>

        {isError && <p className="muted" style={{ marginBottom: 12 }}>No se pudo guardar la tarea.</p>}

        <button type="submit" className="primary" style={{ width: '100%' }} disabled={isPending}>
          {isPending ? 'Guardando...' : 'Guardar tarea'}
        </button>
      </form>
    </div>
  );
}
