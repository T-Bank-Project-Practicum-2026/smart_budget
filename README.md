# Smart Budget
## Запуск

Нужны запущенный Docker и Make. `make up` создаст недостающие секреты и VAPID-ключи в `.env`, применит миграции и запустит сервисы.

```sh
make up
```

Приложение: http://localhost:3000. Порт frontend задаётся через `FRONTEND_PORT` в `.env`; при его смене обновите `APP_ORIGINS`. Backend доступен внутри Docker-сети, запросы `/api/` проходят через nginx.

Swagger UI: http://localhost:3000/swagger/ — документация и OpenAPI доступны через тот же nginx.

| Команда | Действие |
| --- | --- |
| `make down` | Остановить контейнеры, сохранить данные |
| `make logs` | Смотреть логи |
| `make lint` | Проверить Go, TypeScript и форматирование |
| `make lint-fix` | Исправить форматирование |
| `make test` | Go-тесты с Redpanda/PostgreSQL |
| `make generate` | Пересобрать SQL-методы, Protobuf и OpenAPI |
| `make proto` | Сгенерировать Go-сообщения, RPC-клиенты и обработчики из `.proto` |
| `make check-proto` | Проверить схему и актуальность сгенерированных контрактов |
| `make swagger` | Сгенерировать `api/openapi.yaml` из Go-кода |
| `make check-swagger` | Проверить актуальность OpenAPI |
| `make check-generated` | Проверить актуальность SQL-методов, Protobuf и OpenAPI |
| `make migrate` | Применить SQL-миграции |
| `make migrate-status` | Показать состояние миграций |

В GoLand открывайте корень проекта: `go.work` подключает модуль `backend/`. В Settings → Go → Go Modules должна быть включена `Enable Go modules integration`. В Go SDK выберите Go 1.27.1 или новее, как требует `backend/go.mod`.

## Swagger

После изменения API запустите `make swagger`. YAML генерируется через swaggo в Docker, вручную его править не нужно. `make generate` тоже обновляет Swagger, а CI проверяет его актуальность.

## Межсервисные контракты

Источник контрактов — [`proto/smartbudget/services/v1/contracts.proto`](proto/smartbudget/services/v1/contracts.proto). API вызывает `SubscriptionsService` в notifications; notifications проверяет сессию через `AuthService` в API. Клиенты и HTTP-обработчики генерируются Connect Go, вызовы используют бинарный Protobuf поверх HTTP на существующем порту 8080. Адреса задаются `NOTIFICATIONS_URL` и `AUTH_API_URL`, доступ защищён `NOTIFICATIONS_TOKEN`. Внешний API для браузера остаётся REST/JSON.

После изменения `.proto` запускайте `make proto` или `make generate`. Генераторы Buf, protoc-gen-go и protoc-gen-connect-go с закреплёнными версиями запускаются в Docker. Файлы в `backend/internal/gen/` сохраняются в Git и не редактируются вручную; сборка сервисов использует готовый код. `make check-generated` в CI проверяет также Protobuf-код, lint и форматирование схемы.

## PostgreSQL

База — `smart_budget`, адрес внутри Docker — `postgres:5432`. В текущем релизе данные хранятся в `postgres_baseline_data`.

### Первый запуск после перехода на начальную схему

Сохраните предыдущую версию проекта и её volumes. Остановите предыдущую версию через `make down`, затем запустите текущую через `make up` с тем же именем Compose-проекта. Compose создаст пустые `postgres_baseline_data` и `redpanda_baseline_data`. Мигратор отклоняет базы со старой историей Goose. В новой установке пользователи регистрируются заново и повторно загружают выписки. Push-подписки нужно создать заново, накопленные JSON-сообщения остаются в прежнем брокере.

Для отката остановите текущую версию через `make down`, верните предыдущую версию проекта и выполните `make up` с исходным именем Compose-проекта. Она подключит прежние `postgres_data` и `redpanda_data`. Сохраните volumes обеих версий.

Миграции лежат в `backend/migrations/`, SQL-запросы — в `backend/queries/`. После изменения SQL запускайте `make generate`. Применённые миграции не редактируйте.

## Redpanda

Брокер работает на `kafka:9092` по протоколу Kafka. Worker забирает события из outbox и отправляет их в топик `notifications.requested.v2` в бинарном формате Protobuf. `kafka-init` создаёт этот топик, notifications читает его в группе `notifications-v2`.

Для разработки выделены 1 CPU и 512 MiB памяти, лимит контейнера — 768 MiB. Данные лежат в `redpanda_baseline_data`.

## Авторизация

Регистрация и вход по email и паролю. JWT живёт 15 минут, сессия — до 30 дней. Токены хранятся в HttpOnly-cookie. Выход завершает сессию только на текущем устройстве.

Для POST, PUT и DELETE нужны разрешённый `Origin` и `X-CSRF-Protection: 1`. При запуске по HTTPS задайте `AUTH_COOKIE_SECURE=true` и свой домен в `APP_ORIGINS`.

`403 forbidden_origin` означает отказ CSRF-проверки: проверьте оба условия выше. Общий клиент `frontend/src/shared/api/client.ts` добавляет CSRF-заголовок к изменяющим запросам и отправляет cookie.

При доступе по локальной сети в `.env` на машине с Docker добавьте адрес, по которому открываете сайт. Например (замените IP своим):

```dotenv
APP_ORIGINS=http://localhost:3000,http://localhost:5173,http://192.168.1.10:3000
```

Указывайте схему, хост и порт из браузера, без пути и завершающего `/`. Адреса разделяются запятыми; пробелы вокруг них допускаются. `localhost`, `127.0.0.1` и LAN-IP — разные origin. После изменения `.env` пересоздайте API: `docker compose up -d --no-deps backend`; обычный restart не обновляет окружение. При деплое `.env` не копируется через rsync, поэтому `APP_ORIGINS` нужно настроить на самом сервере.

## Уведомления

Сервис notifications читает события из Redpanda и отправляет Web Push. Перед отправкой проверяет сессию; при временных ошибках повторяет попытку. Закрытая вкладка доставке не мешает.

## Frontend

Интерфейс поддерживает регистрацию, вход, выход и ручной импорт CSV T-Bank. Бюджет использует демоданные. Демо Web Push доступен через API.

## Диагностика

`/api/health` — процесс работает, `/api/ready` — база доступна. Логи — `make logs`, тесты — `make test`. Тесты используют отдельные базы и топики.

## Документы

[ТЗ](docs/requirements.md) · [Архитектура](docs/architecture.md) · [Целевой backend (UML)](docs/architecture.md#целевой-backend-uml-диаграмма-компонентов) · [API](api/openapi.yaml)
