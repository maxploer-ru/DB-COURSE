# GraphQL API (Bonus #1)

Дополнительный контракт находится в [schema.graphql](../graphql/schema.graphql).
REST API остаётся основным публичным API проекта, а GraphQL предоставляет
альтернативную точку доступа к тем же ресурсам и операциям.

## Подключение

Контракт предполагает одну HTTP-точку:

```text
POST /graphql
```

Запрос содержит стандартные GraphQL-поля `query`, `variables` и, при
необходимости, `operationName`. Для защищённых query и mutation используется
тот же заголовок, что и в REST:

```http
Authorization: Bearer <JWT>
Content-Type: application/json
```

Аутентификация выполняется middleware до вызова resolver-а. Публичными являются
регистрация, вход, обновление токена, получение публичных каналов/видео и
публичные чтения комментариев, плейлистов и community.

## Соответствие REST и GraphQL

GraphQL заменяет набор URL одним типизированным контрактом:

- `Query.channels`, `Query.channel` и `Mutation.createChannel` /
  `updateChannel` / `deleteChannel` соответствуют CRUD канала;
- optional-поля `ChannelUpdateInput`, `VideoUpdateInput` и `PlaylistUpdateInput`
  дают частичный UPDATE. Resolver отклоняет input без единого поля;
- поля `Video.comments`, `Playlist.videos`, `Channel.playlists` и
  `CommunityPost.comments` показывают преимущество GraphQL — связанные данные
  можно выбрать одним запросом;
- административные операции помечены в SDL комментариями `Requires the admin
  role`; resolver проверяет роль и возвращает `FORBIDDEN` при её отсутствии;
- удаление видео и комментариев дополнительно проверяет владельца,
  модератора или администратора по правилам предметной области.

Пагинация использует `limit`, `offset` и объект `PageInfo`. Идентификаторы в
GraphQL имеют тип `ID`, хотя в REST они передаются как положительные целые
числа.

## Ошибки

GraphQL использует стандартное поле `errors` в ответе. Для согласованности с
REST в `errors[].extensions.code` используются коды вроде `UNAUTHORIZED`,
`FORBIDDEN`, `NOT_FOUND`, `INVALID_REQUEST` и `INTERNAL_ERROR`. Ошибка
авторизации не должна маскироваться nullable-данными: resolver возвращает
ошибку с соответствующим кодом.

Например:

```json
{
  "data": null,
  "errors": [
    {
      "message": "insufficient permissions",
      "extensions": { "code": "FORBIDDEN" }
    }
  ]
}
```

Ошибки разбора HTTP/JSON и GraphQL-документа могут возвращаться как HTTP
`400`, а ошибки resolver-ов обычно представлены в GraphQL-ответе с HTTP
`200`, согласно стандартной модели GraphQL.

## Загрузка видео

GraphQL не проксирует большой бинарный файл через resolver. Mutation
`initializeVideoUpload` возвращает `uploadUrl`, после чего клиент выполняет
`PUT` файла в object storage с `Content-Type: application/octet-stream` и
завершает публикацию mutation `updateVideoPublication`.
