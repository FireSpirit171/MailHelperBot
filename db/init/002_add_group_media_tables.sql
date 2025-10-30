-- Таблица для хранения информации о группах
CREATE TABLE IF NOT EXISTS group_sessions (
                                              group_id BIGINT PRIMARY KEY,
                                              group_title TEXT NOT NULL,
                                              owner_id BIGINT NOT NULL,
                                              created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    media_type TEXT CHECK (media_type IN ('photos', 'videos', 'all')) DEFAULT 'photos'
    );

-- Таблица для отслеживания обработанных медиа (опционально, для избежания дубликатов)
CREATE TABLE IF NOT EXISTS processed_media (
                                               media_id TEXT PRIMARY KEY,
                                               group_id BIGINT NOT NULL,
                                               file_id TEXT NOT NULL,
                                               file_name TEXT NOT NULL,
                                               media_type TEXT NOT NULL,
                                               created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    FOREIGN KEY (group_id) REFERENCES group_sessions(group_id) ON DELETE CASCADE
    );

-- Индексы для улучшения производительности
CREATE INDEX IF NOT EXISTS idx_processed_media_group_id ON processed_media(group_id);
CREATE INDEX IF NOT EXISTS idx_processed_media_created_at ON processed_media(created_at);

-- Добавляем поле для отслеживания выгрузки истории
ALTER TABLE group_sessions ADD COLUMN IF NOT EXISTS history_processed BOOLEAN DEFAULT FALSE;