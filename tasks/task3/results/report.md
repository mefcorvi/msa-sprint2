# Задание 3. Отчёт

## Что сделано

GraphQL API личного кабинета на Apollo Federation (Node.js, TypeScript). Сервисы подключены к сети `hotelio-net` и используют сервисы задания 2.

| Сервис | Порт | Что делает |
|---|---|---|
| `apollo-gateway` | 4000 | Объединяет схемы субграфов и передаёт им заголовок `userid` |
| `booking-subgraph` | 4001 | Проверяет доступ и возвращает бронирования пользователя из `booking-service` по gRPC |
| `hotel-subgraph` | 4002 | Возвращает отели по `id` из REST API монолита |

Поле `Booking.hotel` возвращает `id` отеля, остальные поля отеля шлюз запрашивает у `hotel-subgraph`.

## Изменения относительно заготовки

- Заглушки заменены вызовами `booking-service` по gRPC и монолита по REST.
- ACL: пользователь видит только свои бронирования. Без заголовка `userid` субграф возвращает `UNAUTHENTICATED`, с чужим `userid` - `FORBIDDEN`.
- Шлюз передаёт `userid` субграфам через `RemoteGraphQLDataSource.willSendRequest`, так как сам Apollo Gateway этого не делает.
- В `Booking` добавлены `hotel`, `price` и `createdAt`, а `discountPercent` стал `Float`. В `Hotel` поле `name` возвращает описание отеля, так как названия в монолите нет, а поле `stars` убрано.

## Запуск

```bash
cd tasks/task2 && docker compose up -d --build
cd ../task3 && docker compose up -d --build
```

Запрос на `http://localhost:4000/` с заголовком `userid: test-user-2`:

```graphql
query {
  bookingsByUser(userId: "test-user-2") {
    id
    hotel { name city }
    discountPercent
  }
}
```

## Файлы результата

| Файл | Содержание |
|---|---|
| [docker-ps.txt](docker-ps.txt) | `docker ps` |
| [acl-success.png](acl-success.png) | Успешный вызов: `userid: test-user-2` получает свои бронирования с данными отелей |
| [acl-deny.png](acl-deny.png) | Отказ без заголовка `userid`: `UNAUTHENTICATED` |
| [acl-deny-2.png](acl-deny-2.png) | Отказ при запросе чужих бронирований (`userid: test-user-3`): `FORBIDDEN` |
| [booking-subgraph-logs.txt](booking-subgraph-logs.txt) | Логи `booking-subgraph` после двух запросов |

## Замечания

- Заголовок `userid` присылает клиент, поэтому его можно подделать. В целевом состоянии (ADR-001) `userid` подставляет API Gateway из проверенного JWT.
