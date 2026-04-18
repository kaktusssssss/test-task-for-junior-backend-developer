# Task Service

Сервис для управления задачами с HTTP API на Go.

## Требования

- Go `1.23+`
- Docker и Docker Compose

## Быстрый запуск через Docker Compose

```bash
docker compose up --build
```

После запуска сервис будет доступен по адресу `http://localhost:8080`.

Если `postgres` уже запускался ранее со старой схемой, пересоздай volume:

```bash
docker compose down -v
docker compose up --build
```

Причина в том, что SQL-файл из `migrations/0001_create_tasks.up.sql` монтируется в `docker-entrypoint-initdb.d` и применяется только при инициализации пустого data volume.

## Swagger

Swagger UI:

```text
http://localhost:8080/swagger/
```

OpenAPI JSON:

```text
http://localhost:8080/swagger/openapi.json
```

## API

Базовый префикс API:

```text
/api/v1
```

Основные маршруты:

- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`

В сервис добавлена поддержка периодических задач. При создании задачи с настройками периодичности автоматически создаются дочерние задачи на указанные даты.

Типы периодичности
1. Daily (ежедневно)
Задача повторяется каждые N дней.

json
{
  "title": "Prepare release",
  "description": "Collect release notes and check migrations",
  "status": "new",
  "recurrence_type": "daily",
  "recurrence_value": "2"
}
recurrence_value — интервал в днях (например, "2" = каждые 2 дня)

Создаются задачи на 30 дней вперёд

2. Monthly (ежемесячно)
Задача повторяется каждый месяц в указанное число.

json
{
  "title": "Prepare release",
  "description": "Collect release notes and check migrations",
  "status": "new",
  "recurrence_type": "monthly",
  "recurrence_value": "15"
}
recurrence_value — число месяца от 1 до 31

Создаются задачи на 30 дней вперёд

3. Parity (чётные/нечётные дни)
Задача создаётся только на чётные или только на нечётные дни месяца.

json
{
  "title": "Prepare release",
  "description": "Collect release notes and check migrations",
  "status": "new",
  "recurrence_type": "parity",
  "recurrence_value": "even"
}
recurrence_value: "even" — только чётные дни

recurrence_value: "odd" — только нечётные дни

Создаются задачи на 30 дней вперёд

4. Specific (конкретные даты)
Задача создаётся только на указанные даты.

json
{
  "title": "Prepare release",
  "description": "Collect release notes and check migrations",
  "status": "new",
  "recurrence_type": "specific",
  "recurrence_value": "2026-05-15,2026-06-01,2026-06-15,2026-07-01"
}
recurrence_value — список дат в формате YYYY-MM-DD, разделённых запятыми

Создаются задачи только на указанные даты

Некорректные даты пропускаются без ошибки

Структура ответа
Родительская задача 
json
{
"id": 1,
"title": "Prepare release",
"description": "Collect release notes and check migrations",
"status": "new",
"created_at": "2026-04-18T07:32:15.111823Z",
"updated_at": "2026-04-18T07:32:15.111823Z",
"recurrence_type": "specific",
"recurrence_value": "2026-05-15,2026-06-01,2026-06-15,2026-07-01"
}
Дочерняя задача 
json
{
"id": 2,
"title": "Prepare release",
"description": "Collect release notes and check migrations",
"status": "new",
"created_at": "2026-05-15T00:00:00Z",
"updated_at": "2026-05-15T00:00:00Z",
"parent_task_id": "1"
}

Принятые решения
Создание дочерних задач: При создании задачи с периодичностью автоматически создаются отдельные записи в БД для каждого вхождения (на 30 дней вперёд).

Независимость экземпляров: Каждая дочерняя задача независима — можно завершить одну, не влияя на остальные.

Связь с родителем: Дочерние задачи содержат parent_task_id для отслеживания происхождения.

Отсутствие периодичности у дочерних задач: Дочерние задачи не содержат полей recurrence_type и recurrence_value, так как являются конкретными экземплярами.

Ограничение по времени: В текущей версии периодические задачи создаются на 30 дней вперёд. При необходимости лимит можно увеличить.

Обработка ошибок: При создании дочерних задач ошибки логируются, но не отменяют создание родительской задачи.

Известные ограничения
- Периодические задачи создаются только на 30 дней вперёд
- При обновлении родительской задачи дочерние не обновляются автоматически
- Удаление родительской задачи не удаляет дочерние (сохраняется история)
- Тип `specific` ожидает даты в формате UTC (без времени)