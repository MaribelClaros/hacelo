-- Initial schema. Executed on every startup, so every statement is idempotent.
-- Do not use semicolons inside comments: statements are split on them.

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS push_subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    endpoint TEXT NOT NULL UNIQUE,
    p256dh_key TEXT NOT NULL,
    auth_key TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    description TEXT NOT NULL,
    urgency TEXT NOT NULL CHECK (urgency IN ('low', 'medium', 'high')) DEFAULT 'medium',
    status TEXT NOT NULL CHECK (status IN ('pending', 'verifying', 'overdue', 'done')) DEFAULT 'pending',
    original_time TEXT NOT NULL,
    snooze_minutes INTEGER NOT NULL DEFAULT 15,
    next_reminder_at TEXT,
    verification_round INTEGER NOT NULL DEFAULT 0,
    last_sent_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    completed_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_tasks_due ON tasks(status, next_reminder_at);

-- The app has a single user for now. Every repository query uses id = 1.
INSERT OR IGNORE INTO users (id, email) VALUES (1, 'me@local');
