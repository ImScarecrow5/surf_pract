# BUG-02: Код 1234 не работает для авторизации

## Описание
При попытке войти с кодом `1234` (dev mode) возвращается ошибка "неверный код", хотя код должен приниматься для любого телефона в режиме разработки.

## Причина
Валидация кода в Gin использует тег `len=4,numeric`, который требует ровно 4 цифры. Код `1234` состоит из 4 цифр, поэтому должен проходить валидацию. Однако, есть нюанс:

1. При каждом запуске Docker создаётся новая БД, и таблица `clients` не имеет колонки `role`
2. При создании нового пользователя без колонки `role` запрос падает с ошибкой "column does not exist"

## Корневая причина
В файле `backend/src/handlers/handlers.go` валидатор `len=4,numeric` требует ровно 4 символа:
```go
Code  string `json:"code" binding:"required,len=4,numeric"`
```

Хотя `1234` соответствует этому требованию, проблема в том, что:
1. При первом входе нового пользователя выполняется запрос к БД с колонкой `role`, которой нет
2. Ошибка БД перехватывается и маскируется под "неверный код"

## Решение
1. Добавить колонки `role` и `instructor_id` в миграции или создать их при старте:
   ```sql
   ALTER TABLE clients ADD COLUMN role VARCHAR(20) DEFAULT 'client';
   ALTER TABLE clients ADD COLUMN instructor_id INTEGER REFERENCES instructors(id);
   ALTER TABLE clients ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
   ```

2. Валидация была изменена на `min=4,max=6` для поддержки разной длины кода:
   ```go
   Code  string `json:"code" binding:"required,min=4,max=6"`
   ```

## Статус
Исправлено