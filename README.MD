# Person Service API: Тестовое задание

## Описание проекта

Данный проект представляет собой REST API, реализованное на **Golang** с использованием фреймворка **Gin**, для управления записями о людях. API поддерживает CRUD-операции, обогащение данных (возраст, пол, национальность) с помощью внешних API, пагинацию, фильтрацию, структурированное логирование, конфигурацию через `.env`, миграции базы данных и Swagger-документацию. Все требования тестового задания выполнены.

### Выполненные требования

1. **REST API**:
    - Реализованы эндпоинты:
        - `POST /api/person`: Создание персоны.
        - `GET /api/person/:id`: Получение персоны по ID.
        - `GET /api/people`: Получение списка персон с пагинацией и фильтрами.
        - `PUT /api/person/:id`: Обновление персоны.
        - `DELETE /api/person/:id`: Удаление персоны.
    - Формат создания персоны:
      ```json
      {
          "name": "Dmitriy",
          "surname": "Ushakov",
          "patronymic": "Vasilevich" // необязательно
      }
      ```
    - Поддержка пагинации (`page`, `page_size`) и фильтров (`name`, `surname`, `age`, `gender`, `nationality`).

2. **Обогащение данных**:
    - Интеграция с внешними API:
        - Возраст: `https://api.agify.io/?name=NAME`
        - Пол: `https://api.genderize.io/?name=NAME`
        - Национальность: `https://api.nationalize.io/?name=NAME`
    - Обогащённые данные сохраняются в БД и возвращаются в ответах.

3. **База данных**:
    - Используется **PostgreSQL**.
    - Структура таблицы `people` создаётся через миграции с помощью `golang-migrate`.
    - Поля: `id`, `name`, `surname`, `patronymic`, `age`, `gender`, `nationality`, `created_at`.

4. **Логирование**:
    - Реализовано с помощью `logrus` (уровни `debug`, `info`).
    - Логи включают:
        - Детали запросов (эндпоинт, параметры, ошибки).
        - Вызовы внешних API (успех/ошибка).
        - Операции с БД (создание, чтение, обновление, удаление).
        - Запуск/остановка сервера.

5. **Конфигурация**:
    - Все конфигурационные параметры вынесены в `.env`

6. **Swagger-документация**:
    - Сгенерирована с помощью `swaggo/swag`.
    - Доступна по `http://localhost:8080/swagger/index.html`.
    - Включает описание всех эндпоинтов, параметров, ответов и ошибок.

## Технологический стек

- **Язык**: Golang
- **Фреймворк**: Gin
- **База данных**: PostgreSQL
- **Логирование**: Logrus
- **Миграции**: golang-migrate
- **Документация**: Swagger (swaggo)
- **Внешние API**: agify.io, genderize.io, nationalize.io
- **Зависимости**: Управляются через `go.mod`

## Установка и запуск

1. **Клонирование репозитория**:

   ```bash
   git clone https://github.com/SharpDenin/person-service
   cd person-service
   ```

2. **Установка зависимостей**:

   ```bash
   go mod tidy
   ```

3. **Настройка PostgreSQL**:

    - Cоздание БД происходит с помощью docker-compose.yaml
    - Миграции создадут таблицу автоматически при запуске.


4. **Настройка `.env`**:

   Создайте файл `.env` в корне проекта:

   ```env
   # Database
   DATABASE_URL=postgres://postgres:postgres@localhost:5444/person-service-db
   DATABASE_DSN=postgres://postgres:postgres@localhost:5444/person-service-db?sslmode=disable
   # Server
   GIN_MODE=release
   SERVER_PORT=8080
   ```

5. **Установка Swagger CLI**:

   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   ```

6. **Генерация Swagger-документации**:

   ```bash
   swag init
   ```

7. **Запуск сервера**:

   ```bash
   go run cmd/main.go
   ```

   API доступно по `http://localhost:8080`.

8. **Проверка Swagger UI**:

   Откройте `http://localhost:8080/swagger/index.html`.