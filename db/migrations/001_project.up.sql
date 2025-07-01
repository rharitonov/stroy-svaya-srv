-- проект
CREATE TABLE IF NOT EXISTS project (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL UNIQUE,                      -- Шифр проекта (уникальный)
    name TEXT NOT NULL,                             -- Название проекта
    address TEXT,                                   -- Адрес объекта
    parent_project_id INTEGER,                      -- ID родительского проекта (для иерархии)
    start_date DATE,                                -- Дата начала проекта
    end_date DATE,                                  -- Планируемая дата завершения
    status TEXT DEFAULT 'active' 
        CHECK(status IN ('active', 'completed', 'canceled', 'planned')), -- Статус проекта
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    
    FOREIGN KEY (parent_project_id) REFERENCES project(id) -- Рекурсивная связь для иерархии проектов
    CHECK (end_date IS NULL OR start_date IS NULL OR end_date >= start_date) -- Валидация дат
);

-- Создаем индекс для родительских проектов
CREATE INDEX IF NOT EXISTS idx_project_parent ON project(parent_project_id);
CREATE INDEX IF NOT EXISTS idx_project_code ON project(code);

CREATE TRIGGER IF NOT EXISTS update_project_timestamp
AFTER UPDATE ON project
FOR EACH ROW
BEGIN
    UPDATE project SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;