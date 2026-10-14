import { useState } from 'react';
import { useNavigate, useParams, Link } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '../api';

// Screen opened by the "Ahora no puedo" notification button
// (see public/sw.js, notificationclick event).
export default function Reschedule() {
  const { id } = useParams<{ id: string }>();
  const taskId = Number(id);
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [customDate, setCustomDate] = useState('');

  const { data: task } = useQuery({
    queryKey: ['tasks', taskId],
    queryFn: () => api.getTask(taskId),
    enabled: Number.isFinite(taskId),
  });

  const { mutate } = useMutation({
    mutationFn: (nextReminderAt: string) =>
      api.updateTask(taskId, { next_reminder_at: nextReminderAt }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      navigate('/');
    },
  });

  const presets = [
    { label: 'En 1 hora', at: () => new Date(Date.now() + 60 * 60 * 1000) },
    { label: 'Esta noche, 20:00', at: () => atToday(20, 0) },
    { label: 'Mañana, misma hora', at: () => addDays(task ? new Date(task.original_time) : new Date(), 1) },
  ];

  return (
    <div>
      <div className="top-bar">
        <Link to="/" aria-label="Volver">←</Link>
        <h2 style={{ margin: 0 }}>Reprogramar</h2>
        <div style={{ width: 16 }} />
      </div>
      {task && <p className="muted" style={{ marginBottom: 24 }}>{task.description}</p>}

      <label>Elegí cuándo recordarte</label>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 20 }}>
        {presets.map((preset) => (
          <button
            key={preset.label}
            type="button"
            style={{ textAlign: 'left' }}
            onClick={() => mutate(preset.at().toISOString())}
          >
            {preset.label}
          </button>
        ))}
      </div>

      <label htmlFor="customDate">O elegí fecha y hora</label>
      <input
        id="customDate"
        type="datetime-local"
        value={customDate}
        onChange={(e) => setCustomDate(e.target.value)}
        style={{ marginBottom: 24 }}
      />

      <button
        type="button"
        className="primary"
        style={{ width: '100%' }}
        disabled={!customDate}
        onClick={() => mutate(new Date(customDate).toISOString())}
      >
        Guardar
      </button>
    </div>
  );
}

// Today at the given time, or tomorrow if that time already passed.
function atToday(hours: number, minutes: number): Date {
  const d = new Date();
  d.setHours(hours, minutes, 0, 0);
  if (d.getTime() < Date.now()) d.setDate(d.getDate() + 1);
  return d;
}

function addDays(date: Date, days: number): Date {
  const d = new Date(date);
  d.setDate(d.getDate() + days);
  return d;
}
