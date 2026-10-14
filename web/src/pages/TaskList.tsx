import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { api, type Status } from '../api';

const STATUS_LABEL: Record<Status, string> = {
  pending: 'Pendiente',
  verifying: 'Verificando',
  overdue: 'Atrasada',
  done: 'Hecha',
};

function formatTime(iso: string) {
  return new Date(iso).toLocaleTimeString('es-AR', { hour: '2-digit', minute: '2-digit' });
}

export default function TaskList() {
  const { data: tasks, isLoading, error } = useQuery({
    queryKey: ['tasks', 'today'],
    queryFn: () => api.listTasks(),
  });

  return (
    <div>
      <div className="top-bar">
        <div>
          <p className="muted">
            {new Date().toLocaleDateString('es-AR', { weekday: 'long', day: 'numeric', month: 'long' })}
          </p>
          <h1>Hoy</h1>
        </div>
        <Link to="/settings">Ajustes</Link>
      </div>

      {isLoading && <p className="muted">Cargando...</p>}
      {error && <p className="muted">No se pudo conectar con el backend.</p>}

      {tasks?.length === 0 && <p className="muted">No tenés tareas todavía.</p>}

      {tasks?.map((task) => (
        <div className="card" key={task.id}>
          <div>
            <p style={{ fontWeight: 500, fontSize: 14, textDecoration: task.status === 'done' ? 'line-through' : 'none' }}>
              {task.description}
            </p>
            <p className="muted">{formatTime(task.original_time)}</p>
          </div>
          <span className={`badge badge-${task.status}`}>{STATUS_LABEL[task.status]}</span>
        </div>
      ))}

      <div style={{ position: 'relative', height: 44, marginTop: 12 }}>
        <Link
          to="/tasks/new"
          className="primary"
          style={{
            position: 'absolute',
            right: 0,
            bottom: 0,
            width: 44,
            height: 44,
            borderRadius: '50%',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            textDecoration: 'none',
            fontSize: 22,
            background: 'var(--accent)',
            color: 'white',
          }}
          aria-label="Nueva tarea"
        >
          +
        </Link>
      </div>
    </div>
  );
}
