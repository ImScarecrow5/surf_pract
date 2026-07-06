# Bug: Не отображались слоты (тренировки)

## Описание
На главной странице не отображались доступные слоты (тренировки). Пользователь видел пустой экран или сообщение "Нет доступных слотов".

## Причина
Несоответствие между схемой БД и SQL-запросами в бэкенде:

1. **Неправильные имена колонок** в запросах GetSlots:
   - `total_participants` → должно быть `total_places`
   - `free_participants` → должно быть `free_places`
   - `max_participants` → должно быть `max_capacity`

2. **Неправильный статус слота**:
   - В запросах использовался `status = 'scheduled'`
   - В БД слоты имели статус `available`

3. **Синтаксическая ошибка в миграции** (db/migrations/001_init.sql):
   - Был написан некорректный SQL с `INSTRUCTOR i ON ...` вместо `JOIN instructors i ON ...`
   - Из-за этого слоты не генерировались при создании БД

4. **Фильтр для новичков**:
   - При фильтрации для `client_type = 'novice'` использовалось английское значение `'bouldering'`
   - В БД зона называется `'Болдеринг'`

5. **Отсутствовало поле `confirmation_deadline`** в таблице bookings:
   - При создании бронирования запрос падал из-за отсутствующей колонки

## Решение

### 1. Исправление SQL-запросов (backend/src/handlers/handlers.go:168-175):

**Было:**
```go
query := `
    SELECT s.id, s.start_time, s.total_participants, s.free_participants, s.price, s.status,
        z.id, z.name, z.description, z.max_participants, z.duration_minutes,
        ...
    WHERE s.status = 'scheduled'
`
```

**Стало:**
```go
query := `
    SELECT s.id, s.start_time, s.total_places, s.free_places, s.price, s.status,
        z.id, z.name, z.description, z.max_capacity, z.duration_minutes,
        ...
    WHERE s.status = 'available'
`
```

### 2. Удаление фильтра для новичков (handlers.go:221-223):

Ранее новичкам показывались только слоты с болдерингом. Теперь все пользователи видят все зоны.

### 3. Добавление поля level в слоты (models.go + handlers.go):

Добавлено поле `level` в модель Slot и заполняется на основе зоны:
- Болдеринг → "novice" (8 мест)
- Трассы с верёвкой → "experienced" (16 мест)

### 4. Добавление колонки в БД:

```sql
ALTER TABLE bookings ADD COLUMN confirmation_deadline TIMESTAMP;
```

### 5. Генерация слотов в БД:

Создано 28 слотов (2 зоны × 7 дней × 2 слота в день):

```sql
INSERT INTO slots (zone_id, instructor_id, start_time, end_time, total_places, free_places, price, status)
SELECT z.id, i.id, 
       CURRENT_DATE + (g.n || ' days')::INTERVAL + ((h.n * 3) || ' hours')::INTERVAL,
       CURRENT_DATE + (g.n || ' days')::INTERVAL + ((h.n * 3 + z.duration_minutes) || ' minutes')::INTERVAL,
       z.max_capacity, z.max_capacity, 
       CASE WHEN z.name = 'Болдеринг' THEN 1000.00 ELSE 1500.00 END, 
       'available' 
FROM zones z 
CROSS JOIN (SELECT generate_series(0, 6) n) g
CROSS JOIN (SELECT generate_series(0, 1) n) h
JOIN instructors i ON i.id = (g.n + h.n) % 3 + 1
WHERE h.n < 2;
```

## Дата
2026-07-06

## Статус
Исправлен