-- up
-- свайное поле
CREATE TABLE IF NOT EXISTS pile_field (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,             -- ID проекта (связь с таблицей project)
    name TEXT NOT NULL,                     -- Наименование свайного поля
    drawing_number TEXT,                    -- Номер чертежа
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    created_by TEXT,                        -- Кто создал запись
    notes TEXT,                             -- Примечания
    
    FOREIGN KEY (project_id) REFERENCES project(id)
);

CREATE TABLE IF NOT EXISTS pile_in_field (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pile_field_id INTEGER NOT NULL,         -- ID свайного поля
    pile_number TEXT NOT NULL,              -- Номер сваи (может быть буквенно-цифровым)
    pile_type TEXT,                         -- Тип сваи по проекту
    x_coord TEXT,                           -- Координата X
    y_coord TEXT,                           -- Координата Y
    z_coord TEXT,                           -- Координата Z (если нужна)
    design_pile_head INTEGER,               -- Абс. отметка головы сваи по проекту, мм
    design_pile_tip INTEGER,                -- Абс. отметка острия сваи по проекту, мм

    FOREIGN KEY (pile_field_id) REFERENCES pile_field(id),
    UNIQUE (pile_field_id, pile_number)     -- Уникальный номер сваи в пределах поля
);

CREATE INDEX IF NOT EXISTS idx_pile_field_project ON pile_field(project_id);
CREATE INDEX IF NOT EXISTS idx_pile_in_field_pile_field ON pile_in_field(pile_field_id);
CREATE INDEX IF NOT EXISTS idx_pile_in_field_number ON pile_in_field(pile_number);

CREATE TRIGGER IF NOT EXISTS update_pile_field_timestamp
AFTER UPDATE ON pile_field
FOR EACH ROW
BEGIN
    UPDATE pile_field SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE TRIGGER IF NOT EXISTS update_pile_in_field_timestamp
AFTER UPDATE ON pile_in_field
FOR EACH ROW
BEGIN
    UPDATE pile_in_field SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;