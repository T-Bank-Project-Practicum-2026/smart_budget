# Smart Budget

## Запуск

Нужны запущенный Docker и Make. При первом запуске скопируйте `.env.example` в `.env` и задайте `POSTGRES_PASSWORD`.

```sh
make up
```

Приложение: http://localhost:3000. Порт frontend задаётся через `FRONTEND_PORT` в `.env`. Backend доступен внутри Docker-сети, запросы `/api/` проходят через nginx.

| Команда | Действие |
| --- | --- |
| `make down` | Остановить контейнеры, сохранить данные |
| `make logs` | Смотреть логи |
| `make lint` | Проверить Go, TypeScript и форматирование |
| `make lint-fix` | Исправить форматирование |
| `make test` | Go-тесты и подключения к Kafka/PostgreSQL |
| `make migrate` | Применить SQL-миграции |
| `make migrate-status` | Показать состояние миграций |

В GoLand открывайте корень проекта: `go.work` подключает модуль `backend/`. В Settings → Go → Go Modules должна быть включена `Enable Go modules integration`. В Go SDK выберите Go 1.27.1 или новее, как требует `backend/go.mod`.

## PostgreSQL

`make up` запускает PostgreSQL внутри Docker на `postgres:5432`, база — `smart_budget`. Порт на хост не публикуется; volume `postgres_data` сохраняется при `make down`. Пароль берётся из `.env`; его изменение не меняет пароль уже созданной БД.

`backend/internal/postgres.NewPool(ctx, dsn)` возвращает проверенный `*pgxpool.Pool`; пустой DSN использует переменные `PGHOST`, `PGPORT`, `PGDATABASE`, `PGUSER`, `PGPASSWORD`, `PGSSLMODE`. Пул ограничен 10 соединениями, при завершении вызывается `Close()`. API и notifications пока не используют БД.

Миграции goose — `backend/migrations/*.sql`, встроены в бинарник `cmd/migrate`. После `make up` выполните `make migrate`; повторный запуск применяет только новые миграции. `make migrate-status` показывает их состояние. Первая миграция создаёт схему `app` без бизнес-таблиц. Запуски мигратора защищены PostgreSQL advisory lock.

Сейчас мигратор и интеграционные тесты используют локального администратора БД. Отдельные роли приложения и notifications с ограниченными правами добавим перед подключением бизнес-модулей; схема notifications пока не создана.

## Kafka

`make up` запускает один брокер Apache Kafka в KRaft-режиме. Адрес внутри Docker — `kafka:9092`; порт на хост не публикуется. Данные лежат в volume `kafka_data` и сохраняются при `make down`.

Основа клиента — `backend/internal/kafka.NewClient(brokers, opts...)`, возвращает `*kgo.Client`. Worker и notifications могут передавать адреса брокеров и свои опции franz-go, проверяют `Ping(ctx)` и вызывают `Close()` при завершении. API пока Kafka не использует. Топики и обработчики ещё не добавлены; автоматическое создание топиков выключено.

`make test` запускает Kafka и проверяет соединение через franz-go. При прямом `go test` без `KAFKA_BROKERS` эта проверка пропускается.

## Notifications

`make up` собирает и запускает отдельный контейнер `notifications` из `backend/cmd/notifications`. Он проверяет подключение к Kafka при старте и корректно закрывает HTTP-сервер и Kafka-клиент при завершении.

Конфигурация: `HTTP_ADDR` (по умолчанию `:8080`), обязательный `KAFKA_BROKERS` (адреса через запятую; в Compose — `kafka:9092`). Внутренний `http://notifications:8080/api/health` показывает состояние процесса, а не готовность доставки push; порт на хост не публикуется.

Это только каркас: подписки, чтение событий, inbox и отправка Web Push ещё не реализованы.

## Frontend

Из `frontend/`, с Node.js 22.12+ и npm:

```sh
npm ci
npm run dev
```

Для разработки frontend запустите `make up`. Vite открывает страницу на http://localhost:5173 и отправляет `/api/` через nginx на порту 3000. Если порт Docker-приложения изменён, задайте его через `API_PROXY_TARGET=http://localhost:3002 npm run dev`.

`npm run build` проверяет типы и собирает `dist/`. Docker раздаёт эту сборку через nginx. Исходники страницы — `frontend/src/`.

## Документы

[ТЗ](docs/requirements.md), [Архитектура](docs/architecture.md).