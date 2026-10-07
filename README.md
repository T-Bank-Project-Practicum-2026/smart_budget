# Smart Budget

## Запуск

Нужны запущенный Docker и Make. `make up` создаст недостающие секреты и VAPID-ключи в `.env`, применит миграции и запустит сервисы.

```sh
make up
```

Приложение: http://localhost:3000. Порт frontend задаётся через `FRONTEND_PORT` в `.env`; при его смене обновите `APP_ORIGINS`. Backend доступен внутри Docker-сети, запросы `/api/` проходят через nginx.

| Команда | Действие |
| --- | --- |
| `make down` | Остановить контейнеры, сохранить данные |
| `make logs` | Смотреть логи |
| `make lint` | Проверить Go, TypeScript и форматирование |
| `make lint-fix` | Исправить форматирование |
| `make test` | Go-тесты с Redpanda/PostgreSQL |
| `make generate` | Пересобрать Go-методы из SQL через sqlc |
| `make check-generated` | Проверить актуальность сгенерированного кода |
| `make migrate` | Применить SQL-миграции |
| `make migrate-status` | Показать состояние миграций |

В GoLand открывайте корень проекта: `go.work` подключает модуль `backend/`. В Settings → Go → Go Modules должна быть включена `Enable Go modules integration`. В Go SDK выберите Go 1.27.1 или новее, как требует `backend/go.mod`.

## PostgreSQL

База — `smart_budget`, адрес внутри Docker — `postgres:5432`. Данные сохраняются в volume.

Миграции лежат в `backend/migrations/`, SQL-запросы — в `backend/queries/`. После изменения SQL запускайте `make generate`. Применённые миграции не редактируйте.

## Redpanda

Брокер работает на `kafka:9092` по протоколу Kafka. Worker забирает события из outbox и отправляет их в топик `notifications.requested.v1`.

Для разработки выделены 1 CPU и 512 MiB памяти, лимит контейнера — 768 MiB. Данные лежат в `redpanda_data`; старые топики из Kafka автоматически не переносятся.

## Авторизация

Регистрация и вход по email и паролю. JWT живёт 15 минут, сессия — до 30 дней. Токены хранятся в HttpOnly-cookie. Выход завершает сессию только на текущем устройстве.

Для POST, PUT и DELETE нужны разрешённый `Origin` и `X-CSRF-Protection: 1`. При запуске по HTTPS задайте `AUTH_COOKIE_SECURE=true` и свой домен в `APP_ORIGINS`.

## Уведомления

Сервис notifications читает события из Redpanda и отправляет Web Push. Перед отправкой проверяет сессию; при временных ошибках повторяет попытку. Закрытая вкладка доставке не мешает.

## Frontend

Пока это экран бюджета с демо-данными. Авторизация и push на backend готовы, но к интерфейсу ещё не подключены.

## Диагностика

`/api/health` — процесс работает, `/api/ready` — база доступна. Логи — `make logs`, тесты — `make test`. Тесты используют отдельные базы и топики.

## Документы

[ТЗ](docs/requirements.md) · [Архитектура](docs/architecture.md) · [API](api/openapi.yaml)
