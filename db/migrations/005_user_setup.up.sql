-- up
CREATE TABLE IF NOT EXISTS user_setup (
    code TEXT NOT NULL PRIMARY KEY,                 -- код сотрудника\шифр
    first_name TEXT,                                -- имя
    last_name TEXT,                                 -- фамилия
    surname TEXT,                                   -- отчество
    initials TEXT,                                  -- инициалы
    tg_user_id INTEGER,                             -- тг user_id
    email TEXT,                                     -- адрес эл. почты
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);


-- Триггер для автоматического обновления поля updated_at
CREATE TRIGGER IF NOT EXISTS update_user_setup_timestamp
AFTER UPDATE ON user_setup
FOR EACH ROW
BEGIN
    UPDATE user_setup SET updated_at = CURRENT_TIMESTAMP WHERE code = OLD.code;
END;
