#!/bin/bash

NAME=$1
if [ -z "$NAME" ]; then
  echo "Использование: create_migration.sh имя_миграции"
  exit 1
fi

COUNT=$(ls db/migrations/*.up.sql | wc -l | awk '{print $1}')
NEXT=$(printf "%03d" $((COUNT + 1)))

echo "-- up" > "db/migrations/${NEXT}_${NAME}.up.sql"
echo "-- down" > "db/migrations/${NEXT}_${NAME}.down.sql"

echo "Созданы файлы:"
echo "- db/migrations/${NEXT}_${NAME}.up.sql"
echo "- db/migrations/${NEXT}_${NAME}.down.sql"