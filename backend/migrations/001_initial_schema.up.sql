-- users
CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text UNIQUE NOT NULL,
    password text NOT NULL,
    phone_number text,
    name text NOT NULL,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

-- documents
CREATE TABLE IF NOT EXISTS documents (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL,
    description text,
    identifier text,
    expiration_date date NOT NULL,
    timezone text DEFAULT 'UTC',
    attachment_url text,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

-- reminder_intervals (global list of supported intervals; used for presets)
CREATE TABLE IF NOT EXISTS reminder_intervals (
    id serial PRIMARY KEY,
    label text NOT NULL, -- e.g. '1 week before'
    days_before int NOT NULL, -- e.g. 7, 3, 1, 0
    id_label text NOT NULL -- e.g. '7d', '3d', '1d', '0d'
);

-- document_reminders (what reminders are enabled for this document)
CREATE TABLE IF NOT EXISTS document_reminders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id uuid REFERENCES documents(id) ON DELETE CASCADE,
    reminder_interval_id int REFERENCES reminder_intervals(id) ON DELETE CASCADE,
    enabled boolean DEFAULT true,
    sent_at timestamptz NULL -- last sent time for this reminder occurrence (optional)
);

-- notification_logs
CREATE TABLE IF NOT EXISTS notification_logs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid REFERENCES users(id),
    document_id uuid REFERENCES documents(id),
    reminder_interval_id int,
    channel text, -- 'email' | 'push' | 'sms'
    status text,  -- 'sent' | 'failed'
    response jsonb,
    created_at timestamptz DEFAULT now()
);


-- Unique index for quick user lookup
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_id ON users(id);

-- Unique index for quick email lookup
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Documents: often queried by user
CREATE INDEX IF NOT EXISTS idx_documents_user_id ON documents(user_id);

-- Document reminders: queried by document & interval
CREATE INDEX IF NOT EXISTS idx_document_reminders_document_id ON document_reminders(document_id);
CREATE INDEX IF NOT EXISTS idx_document_reminders_interval_id ON document_reminders(reminder_interval_id);

-- Notification logs: filter by user, document, interval, or status
CREATE INDEX IF NOT EXISTS idx_notification_logs_user_id ON notification_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_notification_logs_document_id ON notification_logs(document_id);
CREATE INDEX IF NOT EXISTS idx_notification_logs_interval_id ON notification_logs(reminder_interval_id);
CREATE INDEX IF NOT EXISTS idx_notification_logs_status ON notification_logs(status);