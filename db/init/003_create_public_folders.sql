-- Таблица для хранения информации о публичных папках
CREATE TABLE IF NOT EXISTS public_folders (
                                              id SERIAL PRIMARY KEY,
                                              chat_id BIGINT NOT NULL,
                                              folder_name TEXT NOT NULL,
                                              public_url TEXT NOT NULL,
                                              files_count INT DEFAULT 0,
                                              created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    FOREIGN KEY (chat_id) REFERENCES user_sessions(chat_id) ON DELETE CASCADE
    );

-- Индекс для быстрого поиска по chat_id
CREATE INDEX idx_public_folders_chat_id ON public_folders(chat_id);