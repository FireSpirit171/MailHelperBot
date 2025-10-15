# MailHelperBot

### Разработка

1. Локальный запуск

```sh
docker compose up --build
```

2. Параллельно (в другом терминале) можно смотреть, что лежит в БД

```sh
docker exec -it mail_helper_db psql -U mail_bot -d mail_helper
```

Затем ввести пароль из .env

3. При необходимости (например, осталась прошлая таблица в БД) можно почстить образы

```sh
docker compose down -v
```

### Зависимости

- PostgreSQL 18
- golang 1.23.3
