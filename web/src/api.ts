import { idbGet } from './idb';

export type Urgency = 'low' | 'medium' | 'high';
export type Status = 'pending' | 'verifying' | 'overdue' | 'done';

export type Task = {
  id: number;
  description: string;
  urgency: Urgency;
  status: Status;
  original_time: string;
  snooze_minutes: number;
  next_reminder_at?: string;
  verification_round: number;
  completed_at?: string;
};

export type CreateTaskInput = {
  description: string;
  urgency: Urgency;
  scheduled_at: string;
  snooze_minutes?: number;
};

export type UpdateTaskInput = Partial<{
  description: string;
  urgency: Urgency;
  next_reminder_at: string;
  snooze_minutes: number;
}>;

async function authHeaders(): Promise<Record<string, string>> {
  const token = await idbGet('apiToken');
  return token ? { Authorization: `Bearer ${token}` } : {};
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`/api${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(await authHeaders()),
      ...(init?.headers as Record<string, string> | undefined),
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Request failed with status ${res.status}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const api = {
  listTasks: (params: { date?: string; status?: string } = {}) => {
    const qs = new URLSearchParams(params as Record<string, string>).toString();
    return request<Task[]>(`/tasks${qs ? `?${qs}` : ''}`);
  },
  createTask: (input: CreateTaskInput) =>
    request<Task>('/tasks', { method: 'POST', body: JSON.stringify(input) }),
  getTask: (id: number) => request<Task>(`/tasks/${id}`),
  updateTask: (id: number, input: UpdateTaskInput) =>
    request<Task>(`/tasks/${id}`, { method: 'PATCH', body: JSON.stringify(input) }),
  deleteTask: (id: number) => request<void>(`/tasks/${id}`, { method: 'DELETE' }),
  vapidPublicKey: () => request<{ public_key: string }>('/push/vapid-public-key'),
  subscribePush: (sub: unknown) =>
    request<void>('/push/subscribe', { method: 'POST', body: JSON.stringify(sub) }),
};
