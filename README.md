# Search Service

Пример сервисa поиска пользователей по XML-датасету. Проект реализован в рамках Курса  "Разработка веб-сервисов на Golang" от VK.

## Структура

- `cmd/server` - точка входа HTTP-сервера, обрабатывающего запросы поиска.
- `cmd/client` - пример клиента.
- `internal/handlers` - HTTP-обработчики, логика фильтрации и сортировки.
- `internal/storage` - загрузка пользователей из датасета.
- `internal/models` - общие структуры и константы.
- `pkg` - пакет с клиентом, который отправляет запросы к поисковому API.
- `tests` - интеграционные тесты.
- `dataset.xml` - тестовый набор данных.

## Запуск сервера

```bash
go run ./cmd/server
```

Сервер по умолчанию слушает `http://localhost:8080` и ожидает заголовок `AccessToken`.

## Пример клиента

```bash
go run ./cmd/client
```

## Тестирование

```bash
go test ./...
```

### Покрытие тестов


```bash
go test -coverpkg=./... -coverprofile=cover.out ./...
go tool cover -html=cover.out -o cover.html
```




