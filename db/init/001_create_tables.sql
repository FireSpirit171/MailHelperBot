-- =====================================================
-- ПОЛЬЗОВАТЕЛИ И АВТОРИЗАЦИЯ
-- =====================================================

-- Таблица для авторизованных пользователей
CREATE TABLE IF NOT EXISTS user_sessions (
                                             chat_id BIGINT PRIMARY KEY,  -- Telegram chat ID
                                             name TEXT NOT NULL,
                                             email TEXT NOT NULL UNIQUE,
                                             access_token TEXT,
                                             refresh_token TEXT,
                                             token_expires_at TIMESTAMP WITH TIME ZONE,
                                             is_logged_in BOOLEAN DEFAULT TRUE,
                                             created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
    );

-- Временные OAuth состояния для процесса авторизации
CREATE TABLE IF NOT EXISTS oauth_states (
    state TEXT PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() + INTERVAL '10 minutes'),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
    );

-- =====================================================
-- ГРУППЫ И МЕДИА
-- =====================================================

-- Информация о группах где работает бот
CREATE TABLE IF NOT EXISTS group_sessions (
                                              group_id BIGINT PRIMARY KEY,  -- Telegram group ID
                                              group_title TEXT NOT NULL,
                                              owner_chat_id BIGINT NOT NULL,  -- chat_id создателя/админа
                                              media_type TEXT CHECK (media_type IN ('photos', 'videos', 'all')) DEFAULT 'photos',
    cloud_folder_path TEXT,
    public_url TEXT,
    history_processed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    FOREIGN KEY (owner_chat_id) REFERENCES user_sessions(chat_id) ON DELETE CASCADE
    );

-- Учет загруженных медиафайлов
CREATE TABLE IF NOT EXISTS processed_media (
                                               id SERIAL PRIMARY KEY,
                                               group_id BIGINT NOT NULL,
                                               file_unique_id TEXT NOT NULL UNIQUE,  -- Уникальный ID файла в Telegram
                                               file_name TEXT,
                                               media_type TEXT NOT NULL CHECK (media_type IN ('photo', 'video')),
    file_size_bytes BIGINT,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    FOREIGN KEY (group_id) REFERENCES group_sessions(group_id) ON DELETE CASCADE
    );

-- =====================================================
-- ПУБЛИЧНЫЕ ПАПКИ
-- =====================================================

-- История расшаренных папок пользователя
CREATE TABLE IF NOT EXISTS shared_folders (
                                              id SERIAL PRIMARY KEY,
                                              chat_id BIGINT NOT NULL,
                                              folder_name TEXT NOT NULL,
                                              folder_path TEXT NOT NULL,
                                              public_url TEXT NOT NULL UNIQUE,
                                              created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    FOREIGN KEY (chat_id) REFERENCES user_sessions(chat_id) ON DELETE CASCADE
    );

-- =====================================================
-- ИНДЕКСЫ ДЛЯ ПРОИЗВОДИТЕЛЬНОСТИ
-- =====================================================

-- Индексы для user_sessions
CREATE INDEX idx_user_sessions_email ON user_sessions(email);
CREATE INDEX idx_user_sessions_logged_in ON user_sessions(is_logged_in);

-- Индексы для oauth_states
CREATE INDEX idx_oauth_states_chat_id ON oauth_states(chat_id);
CREATE INDEX idx_oauth_states_expires_at ON oauth_states(expires_at);

-- Индексы для group_sessions
CREATE INDEX idx_group_sessions_owner_chat_id ON group_sessions(owner_chat_id);

-- Индексы для processed_media
CREATE INDEX idx_processed_media_group_id ON processed_media(group_id);
CREATE INDEX idx_processed_media_media_type ON processed_media(media_type);
CREATE INDEX idx_processed_media_group_type ON processed_media(group_id, media_type);

-- Индексы для shared_folders
CREATE INDEX idx_shared_folders_chat_id ON shared_folders(chat_id);

-- =====================================================
-- ТРИГГЕРЫ ДЛЯ АВТООБНОВЛЕНИЯ updated_at
-- =====================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_user_sessions_updated_at
    BEFORE UPDATE ON user_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_group_sessions_updated_at
    BEFORE UPDATE ON group_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
-- =====================================================

-- Функция для подсчета статистики группы
CREATE OR REPLACE FUNCTION get_group_stats(p_group_id BIGINT)
RETURNS TABLE (
    photos_count BIGINT,
    videos_count BIGINT,
    total_size_mb NUMERIC
) AS $$
BEGIN
RETURN QUERY
SELECT
    COUNT(*) FILTER (WHERE media_type = 'photo') as photos_count,
        COUNT(*) FILTER (WHERE media_type = 'video') as videos_count,
        ROUND(COALESCE(SUM(file_size_bytes), 0) / 1048576.0, 2) as total_size_mb
FROM processed_media
WHERE group_id = p_group_id;
END;
$$ LANGUAGE plpgsql;

-- Функция для очистки истекших OAuth состояний
CREATE OR REPLACE FUNCTION cleanup_expired_oauth_states()
RETURNS void AS $$
BEGIN
DELETE FROM oauth_states WHERE expires_at < now();
END;
$$ LANGUAGE plpgsql;