-- up
CREATE TABLE IF NOT EXISTS pile_driving_record (
    entry_no INTEGER PRIMARY KEY AUTOINCREMENT,
    pile_field_id INTEGER NOT NULL,
    pile_number TEXT NOT NULL,                      -- Номер сваи
    project_id INTEGER NOT NULL,                    -- ID проекта
    start_date DATE NOT NULL,                       -- Дата начала забивки
    end_date DATE,                                  -- Дата окончания забивки
    fact_pile_head INTEGER                          -- Абс. отметка верха головы сваи, факт, мм
    blows_count INTEGER,                            -- Количество ударов
    equip_code TEXT,                                -- молот
    recorded_by INTEGER NOT NULL,                   -- ID оператора/инженера
    notes TEXT,                                     -- Дополнительные заметки
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,  -- Дата создания записи
    updated_at DATETIME,                            -- Дата обновления записи
    
    UNIQUE (project_id, pile_field_id, pile_number),        -- Уникальный номер сваи в пределах поля проекта
    FOREIGN KEY (project_id) REFERENCES project(id),        -- Связь с таблицей проектов
    FOREIGN KEY (pile_field_id) REFERENCES pile_field(id),
    FOREIGN KEY (equip_code) REFERENCES equip(code)      
);

CREATE INDEX IF NOT EXISTS idx_pile_driving_pile_number ON pile_driving_record(pile_number);
CREATE INDEX IF NOT EXISTS idx_pile_driving_project ON pile_driving_record(project_id);
CREATE INDEX IF NOT EXISTS idx_pile_driving_times ON pile_driving_record(start_date, end_date);

-- Триггер для автоматического обновления поля updated_at
CREATE TRIGGER IF NOT EXISTS update_pile_driving_timestamp
AFTER UPDATE ON pile_driving_record
FOR EACH ROW
BEGIN
    UPDATE pile_driving_record SET updated_at = CURRENT_TIMESTAMP WHERE entry_no = OLD.entry_no;
END;
