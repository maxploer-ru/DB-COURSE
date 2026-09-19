# REST mock-сервер

REST mock запускается на Linux/WSL непосредственно из OpenAPI-контракта. Он не
содержит backend-логики и базы данных: Prism читает
[openapi.yaml](../openapi/openapi.yaml), поднимает описанные HTTP-маршруты и
генерирует ответы по `example` или схемам.

## Запуск

Из корня проекта:

```bash
make rest-mock
```

По умолчанию сервер доступен на:

```text
http://localhost:4010
```

Порт можно изменить:

```bash
PORT=4020 make rest-mock
```

В терминале Prism выводит список найденных операций. Например, для текущего
контракта доступны маршруты `/api/v1/channels`, `/api/v1/videos` и остальные
пути из OpenAPI.

## Демонстрация

Получение списка каналов:

```bash
curl -i http://localhost:4010/api/v1/channels
```

Получение конкретного канала:

```bash
curl -i http://localhost:4010/api/v1/channels/7
```

Создание канала с JSON-телом:

```bash
curl -i -X POST http://localhost:4010/api/v1/channels \
  -H 'Content-Type: application/json' \
  -d '{"name":"databases-101","description":"REST demo"}'
```

Для демонстрации ошибки валидации можно передать некорректный параметр:

```bash
curl -i 'http://localhost:4010/api/v1/channels?limit=0'
```

`mjs`-проверка и REST mock решают разные задачи: `make openapi-check`
проверяет контракт автоматически, а Prism позволяет показать преподавателю
реальные HTTP-запросы и ответы без backend-а.

Mock-сервер не является реализацией авторизации или бизнес-правил. Проверка
реального backend на соответствие контракту относится к WebLab#3.
