# SSO Example (Keycloak + Django + Go)

Проект для демонстрации концепции Single Sign-On (SSO) на базе Keycloak с двумя независимыми сервисами:

- Django + DRF + PostgreSQL
- Go (Gin)

Каждый сервис имеет **свой docker-compose** и работает как отдельное приложение, а общая авторизация реализована через Keycloak.

## Структура проекта

- `Makefile` — общий запуск/остановка всех стеков.
- `keycloak/`
  - `docker-compose.yml` — Keycloak + его PostgreSQL.
  - `Dockerfile` — базовый образ Keycloak (место для кастомизаций).
  - `realm-export.json` — автоконфигурация realm, клиентов и тестового пользователя.
  - `.env.example` — пример переменных (не обязателен для запуска в текущей конфигурации).
- `django_app/`
  - `docker-compose.yml` — Django + его PostgreSQL.
  - `Dockerfile`, `requirements.txt`, `.env.example`.
  - Django‑проект с кастомной моделью пользователя и DRF.
- `go_app/`
  - `docker-compose.yml` — Go‑сервис.
  - `Dockerfile`, `go.mod`, `go.sum`, `.env.example`.
  - Gin-сервер с OIDC‑авторизацией через Keycloak.

Все стеки используют общую Docker‑сеть `sso-example-net`.

## Что поднимается

- **Keycloak**
  - Realm: `sso-example`.
  - Клиенты:
    - `django-app` (confidential, redirect: `http://localhost:8211/*`).
    - `go-app` (confidential, redirect: `http://localhost:8212/callback`).
  - Тестовый пользователь:
    - Логин: `testuser`
    - Пароль: `testpassword`
- **Django-сервис**
  - Главная страница `/`: текст `Привет, <username>, это Django сервис` после входа через Keycloak.
  - DRF эндпоинт `/api/me/` — возвращает текущего пользователя.
- **Go-сервис**
  - Главная страница `/`: текст `Привет, <username>, это Go сервис` после входа через Keycloak.

## Первый запуск (максимально автоматизирован)

Требования:

- Docker + Docker Compose V2 (`docker compose`).

Шаги:

1. Клонировать/создать проект и перейти в корень:

   ```bash
   cd sso_example
   ```

2. Выполнить **первый запуск**:

   ```bash
   make first-run
   ```

   Что делает `make first-run`:

   - создаёт общую Docker‑сеть `sso-example-net` (если ещё нет);
   - копирует `.env.example` → `.env` для:
     - `django_app/.env`
     - `go_app/.env`
   - собирает образы Keycloak, Django и Go:
     - `docker compose -f keycloak/docker-compose.yml build`
     - `docker compose -f django_app/docker-compose.yml build`
     - `docker compose -f go_app/docker-compose.yml build`
   - поднимает все стеки:
     - Keycloak стэк
     - Django стэк
     - Go стэк
   - запускает Keycloak с автоимпортом `realm-export.json`, создавая:
     - realm `sso-example`
     - клиентов `django-app`, `go-app`
     - пользователя `testuser / testpassword`

3. После успешного запуска сервисы будут доступны по адресам:

   - Keycloak: `http://localhost:8210`
   - Django: `http://localhost:8211`
   - Go: `http://localhost:8212`

4. Протестировать SSO:

   - Перейти на `http://localhost:8211/` — тебя перенаправит на логин в Keycloak (`realm: sso-example`), залогиниться `testuser / testpassword`, после чего увидишь:
     - `Привет, testuser, это Django сервис`
   - Перейти на `http://localhost:8212/` — аналогично, после логина/уже существующей сессии Keycloak увидишь:
     - `Привет, testuser, это Go сервис`

## Повторные запуски

После того как первый запуск был выполнен и контейнеры/тома уже созданы:

- **Запуск всех стеков**:

  ```bash
  make up-all
  ```

- **Остановка всех стеков**:

  ```bash
  make down-all
  ```

- **Запуск/остановка отдельных стеков**:

  ```bash
  make up-keycloak
  make down-keycloak

  make up-django
  make down-django

  make up-go
  make down-go
  ```

## Изменение настроек

- **Django**:
  - править `django_app/.env` (секрет, БД, OIDC‑параметры);
  - при изменениях кода/зависимостей — пересобрать образ:

    ```bash
    docker compose -f django_app/docker-compose.yml build
    ```

- **Go**:
  - править `go_app/.env` (порт, OIDC‑параметры, секрет сессии);
  - при изменениях кода:

    ```bash
    docker compose -f go_app/docker-compose.yml build
    ```

- **Keycloak**:
  - дефолтная конфигурация берётся из `keycloak/realm-export.json`;
  - для изменения realm/клиентов/пользователей:
    - можно изменить `realm-export.json` и пересобрать/перезапустить Keycloak:

      ```bash
      docker compose -f keycloak/docker-compose.yml build
      make down-keycloak
      make up-keycloak
      ```

## Резюме

- Первый запуск сводится к одной команде: **`make first-run`**.
- Keycloak, Django и Go автоматически поднимаются и настраиваются для работы с общим SSO.
- Тестовый пользователь и клиенты для обоих сервисов создаются автоматически через импорт realm.

