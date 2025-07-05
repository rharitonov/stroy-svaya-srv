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