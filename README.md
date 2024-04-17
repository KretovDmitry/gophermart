# Go RESTful API loyalty service for Gophermart ⭐⭐⭐⭐⭐ 


## Начало


```shell
# запуск постгреса в докере
make db-start

# наполнить базу тестовыми данными
make testdata

# запуск
make run

# запуск с рестартом при любом изменение файлов проекта
# требуется fswatch
make run-live
```

Адрес `http://127.0.0.1:8080`. Эндпойнты:

* `POST /api/user/register` — регистрация пользователя;
* `POST /api/user/login` — аутентификация пользователя;
* `POST /api/user/orders` — загрузка пользователем номера заказа для расчёта;
* `GET /api/user/orders` — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
* `GET /api/user/balance` — получение текущего баланса счёта баллов лояльности пользователя;
* `POST /api/user/balance/withdraw` — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
* `GET /api/user/withdrawals` — получение информации о выводе средств с накопительного счёта пользователем.

## Запросы в Постмане


## Структура проекта

 
```
.
├── cmd                  исполняемые файлы
│   ├── gophermart       сервис накопления баллов
│   └── accural          сервис расчёта баллов          
├── config               файлы конфигураций для разных сред
├── internal             приватные пакеты
│   ├── banner           сервис баннеров
│   ├── auth             аутентификация
│   ├── config           для загрузки конфига
│   ├── jwt              для работы с токеном
│   ├── user             пользователи
├── migrations           миграции бд
├── pkg                  публичные пакеты
│   ├── accesslog        логирование каждого запроса
│   ├── logger           логгер
│   └── test             ... не успел
└── testdata             скрипт для наполнения бд тестовыми данными
```

### Все доступные команды make

```shell
build                          build the API server binary
build-docker                   build the API server as a docker image
clean                          remove temporary files
db-start                       start the database server
db-stop                        stop the database server
fmt                            run "go fmt" on all Go packages
help                           help information about make commands
lint                           run golangchi lint on all Go package
migrate-down                   revert database to the last migration step
migrate-new                    create a new database migration
migrate-reset                  reset database and re-run all migrations
migrate                        run all new database migrations
redis-start                    start the redis server
redis-stop                     stop the redis server
run-live                       run the API server with live reload support (requires fswatch)
run-restart                    restart the API server
run                            run the API server
sqlc-generate                  generate Go code that presents type-safe interfaces to service queries
sqlc-verify                    verify schema changes
sqlc-vet                       run query analyzer on cloud hosted database
test-cover                     run unit tests and show test coverage information
testdata                       populate the database with test data
test                           run unit tests
version                        display the version of the API server
```

