-- Таблица для сессий пользователей
CREATE TABLE IF NOT EXISTS user_sessions (
    chat_id BIGINT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Таблица для хранения OAuth state
CREATE TABLE IF NOT EXISTS oauth_states (
    state TEXT PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    expires_at TIMESTAMP NOT NULL
);
