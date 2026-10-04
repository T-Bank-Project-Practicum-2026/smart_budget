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

База `smart_budget` запускается в Docker (`postgres:5432`), данные сохраняются в volume. Задайте `POSTGRES_PASSWORD` в `.env`; миграции применяются через `make migrate`.

## Kafka

Локальный Kafka доступен сервисам по адресу `kafka:9092`. `make test` проверяет подключение.

## Notifications

Отдельный контейнер `notifications` запускается вместе с проектом и проверяет подключение к Kafka. Отправка push пока не реализована.

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
