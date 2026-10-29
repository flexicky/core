# core

Сервис аутентификации и профилей: gRPC API (login / register / me) поверх PostgreSQL и Redis.

## Возможности

- регистрация и вход по email + пароль (bcrypt);
- access-токены JWT на ed25519, сессии в PostgreSQL;
- кэш сессий в Redis (ключ `session-<id>`) с откатом на БД при промахе;
- фоновая очистка истёкших сессий;
- миграции из обычных `.sql` файлов.

## Структура

```
cmd/core            вход в gRPC-сервис
cmd/migration       применение миграций и создание новых файлов
internal
  action            обработчики RPC (transport-facing use cases)
  app               сборка приложения: grpc / postgres / redis / validator
  apperr            доменные ошибки + маппинг в сообщения для клиента
  config            конфигурация (yaml + env)
  const             константы домена
  ctxkeys           типизированные ключи контекста (client info, session id)
  dto               структуры передачи данных + validate-теги
  grpc              gRPC-сервер аутентификации
  logger            настройка slog
  middleware/jwt    интерцептор проверки access-токена
  migrator          раннер миграций и разбор SQL-скриптов
  redis             подключение к Redis
  repository        доступ к БД (users, sessions)
  service           доменные сервисы (auth, token, user, session, redis)
  storage           пул соединений PostgreSQL
  utils             вспомогательные функции
  worker            фоновые процессы
migrations          SQL-миграции
```

Слои идут сверху вниз: `grpc → action → service → repository`, зависимости
описаны интерфейсами, поэтому сервисы тестируются без БД и Redis.

## Быстрый старт

```bash
cp config/config.example.yaml config/config.yaml   # при необходимости поправьте параметры
make up                                            # PostgreSQL + Redis
make migrate                                       # применить миграции
make run                                           # запустить gRPC-сервис
```

Конфигурация читается из `-config` или `CONFIG_PATH`; все поля можно
переопределить переменными окружения (`DB_HOST`, `REDIS_PASSWORD`,
`ACCESS_TOKEN_TTL`, `JWT_SEED`, ...).

## Переменные окружения (основные)

| Переменная | Назначение | По умолчанию |
| --- | --- | --- |
| `CONFIG_PATH` | путь к yaml-конфигу | — |
| `GRPC_PORT` | порт gRPC | `9090` |
| `DB_HOST` / `DB_PORT` / `DB_NAME` / `DB_USER` / `DB_PASSWORD` / `DB_SSL_MODE` | PostgreSQL | `localhost` / `5432` / — / — / — / `disable` |
| `REDIS_HOST` / `REDIS_PORT` / `REDIS_PASSWORD` / `REDIS_PREFIX` | Redis | `localhost` / `6379` / — / пусто |
| `ACCESS_TOKEN_TTL` | время жизни access-токена | `1h` |
| `SESSION_TTL` | время жизни сессии | `720h` |
| `SESSION_CACHE_TTL` | TTL сессии в Redis | `1h` |
| `JWT_SEED` | base64 ed25519 seed для подписи токенов | пусто (ключ генерируется на старте) |

## Тесты и качество

```bash
make test      # go test ./...
make cover     # покрытие
make vet       # go vet ./...
make fmt       # gofmt
```

## Ограничения, которые стоит знать

- refresh-токен создаётся, но клиенту не выдаётся и в БД хранится в открытом
  виде — эндпоинта обновления токена пока нет;
- при пустом `JWT_SEED` ключ подписи генерируется на старте, поэтому после
  перезапуска ранее выданные access-токены недействительны;
- роли и права (RBAC) не реализованы: любой валидный токен даёт доступ ко всем
  защищённым методам.
