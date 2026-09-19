# Тестирование

Проект использует стандартный пакет `testing` и `testify/suite`. Команды запускаются из корня репозитория через `make`.

## Тестовые стили

В проекте представлены оба подхода:

- London: тесты бизнес-логики в `internal/service/*_test.go`. Зависимости сервисов заменены mock-объектами из `internal/testing/mocks`.
- Classic: тесты PostgreSQL repositories в `internal/infrastructure/db/postgres/repository/*_test.go`. Используется реальный PostgreSQL Testcontainer, а каждый тест работает внутри транзакции с rollback в `TearDownTest`.

Builders находятся в `internal/testing/builder`, Object Mother — в `internal/testing/mother`. Mother-фабрики строят объекты через builders, поэтому тестовые данные не дублируются по suite.

## Нейминг и техники подготовки данных

Имена тестов имеют форму:

```text
Test<PublicMethod>_<ExpectedResult>_<TestDataTechnique>
```

Примеры:

- `TestCreatePost_Positive_StateTransition` — успешный переход состояния;
- `TestCreatePost_Negative_EmptyContent_BoundaryValueAnalysis` — пустая граница допустимого значения;
- `TestChangeUserRole_Negative_SelfChange_Combinatorial` — комбинация роли и идентификаторов;
- `TestListChannels_Positive_BoundaryValueAnalysis` — проверка pagination boundary;
- `TestGetStats_Negative_RatingRepositoryError_EquivalencePartitioning` — класс ошибки repository.

Каждый suite использует Arrange–Act–Assert: настройка fixture и mock expectations, один вызов публичного метода в Act, затем проверки результата и взаимодействий.

Ожидаемые ошибки Go проверяются через `Error`, `ErrorIs`, `Nil` и соответствующие assertions. Это покрывает исключительные сценарии без обращения к private methods.

## Команды

```bash
# Все unit- и integration-тесты, случайный порядок
make test

# Только unit-тесты бизнес-логики и инфраструктурных компонентов без PostgreSQL
make test-unit

# Тот же offline-набор без обращения Go toolchain к proxy/sumdb
make test-offline

# Строковое покрытие, HTML-отчёт и branch/condition coverage
make test-coverage

# JUnit XML и Allure results для всех тестов
make test-allure

# JUnit XML и Allure results только для offline unit-набора
make test-allure-unit

# Открыть уже сгенерированный HTML-отчёт в браузере
make allure-open
```

`make test` и `make test-unit` используют `-shuffle=on`, поэтому порядок тестов выбирается случайно. Для воспроизведения конкретного прогона Go печатает seed в выводе; его можно передать через `-shuffle=<seed>`.

`make test-offline` не запускает PostgreSQL Testcontainer и не скачивает зависимости: он выполняет только mock-based/unit packages с `GOPROXY=off GOSUMDB=off`. Для первого запуска зависимости должны уже находиться в локальном module cache.

## Coverage

`go test -coverprofile` и `go tool cover` формируют line/statement coverage в `coverage.out`, `coverage.html` и выводе `go tool cover -func`.

Go пока не предоставляет branch coverage в штатном `go test`, поэтому `make test-coverage` дополнительно запускает pinned `github.com/rillig/gobco@v1.3.4` по unit-пакетам. Его вывод — condition/branch coverage с указанием непокрытых условий. Это отдельный инструмент поверх стандартного Go coverage.

## Allure

`go test -json` сохраняется в `test-results.json`. Локальный converter `scripts/testreport` без дополнительных Go-зависимостей создаёт:

- `allure-results/junit.xml`;
- native Allure `*-result.json` files;
- `allure-report`, если установлен CLI `allure`.

Отчёт генерируется даже при падении тестов; исходный exit code сохраняется.
После генерации HTML-отчёт открывается командой `make allure-open`.

## Процессы и параллелизм

Команда `go test ./...` сначала собирает отдельный test binary для каждого Go package и запускает package tests отдельным процессом. Все тесты одного package выполняются внутри одного процесса. `TestMain` repository package запускается один раз на этот package и управляет общим PostgreSQL Testcontainer.

`-shuffle=on` меняет порядок тестов внутри test binary, но не превращает каждый тест в отдельный процесс. Отдельные процессы между package создаёт сам `go test`; степень одновременного запуска package ограничивается `-p` и `GOMAXPROCS`. Внутри suite тесты не используют `t.Parallel`, поэтому порядок и состояние fixture контролируются suite lifecycle.

Полезные команды для проверки модели запуска:

```bash
go test -json -shuffle=on ./internal/service
go test -p 1 -json ./...
go test -p 4 -json ./...
```

Источники по стандартному coverage: [Go coverage documentation](https://go.dev/doc/build-cover). Для condition/branch coverage используется [gobco](https://github.com/rillig/gobco), поскольку штатный `go test` формирует statement coverage.
