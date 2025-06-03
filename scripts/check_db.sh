#!/bin/bash

DB_PATH="db/stoy-svaya.db"
MIGRATIONS_DIR="db/migrations"

echo "🔧 Запуск проверки SQLite базы данных..."

# 1. Удалить старую БД, если есть
if [ -f "$DB_PATH" ]; then
    echo "🗑 Удаляю старую БД..."
    rm "$DB_PATH"
fi

# 2. Применить миграции
echo "🔁 Применяю миграции..."
go run cmd/migrate/main.go up
if [ $? -ne 0 ]; then
    echo "❌ Ошибка применения миграций"
    exit 1
fi

# 3. Проверить, существует ли БД
if [ ! -f "$DB_PATH" ]; then
    echo "❌ Файл БД не создан: $DB_PATH"
    exit 1
fi

echo "✅ База данных создана: $DB_PATH"

# 4. Показать список таблиц
echo -e "\n📋 Список таблиц:"
sqlite3 "$DB_PATH" .tables

# 5. Показать структуру таблиц
echo -e "\n📄 Схемы таблиц:"
for table in $(sqlite3 "$DB_PATH" ".tables"); do
    echo -e "\n--- Таблица: $table ---"
    sqlite3 "$DB_PATH" ".schema $table"
done

# 6. Показать статус миграций
echo -e "\n🔢 Статус миграций:"
sqlite3 "$DB_PATH" "SELECT * FROM schema_migrations;" 2>/dev/null
if [ $? -eq 0 ]; then
    sqlite3 "$DB_PATH" "SELECT * FROM schema_migrations;"
else
    echo "⚠ Таблица schema_migrations не найдена"
fi

echo -e "\n✅ Проверка завершена успешно!"