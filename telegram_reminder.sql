CREATE SCHEMA IF NOT EXISTS telegram;

CREATE SEQUENCE IF NOT EXISTS telegram.seq_reminder;

CREATE TABLE IF NOT EXISTS telegram.reminder
(
    id INT NOT NULL default nextval('telegram.seq_reminder'::regClass),
    txt TEXT NOT NULL,
    reminder_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    chat_id INT NOT NULL
);
