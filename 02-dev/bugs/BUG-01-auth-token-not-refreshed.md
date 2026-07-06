# Bug: Выкидывает из аккаунта при обновлении страницы

## Описание
При обновлении страницы (F5 или Ctrl+R) пользователя выкидывало из аккаунта на экран авторизации.

## Причина
Токен доступа сохранялся в localStorage, но при загрузке страницы:
1. ApiService считывал токен один раз в конструкторе при инициализации
2. При первом запросе (например, getProfile) использовался токен, который был сохранён в переменную экземпляра `this.accessToken`
3. Если токен не был установлен при создании экземпляра (что происходило при первом рендере), запросы отправлялись без авторизации

Также были дополнительные проблемы:
- Бэкенд возвращал 404 на `/v1/profile` из-за некорректных SQL-запросов (неправильные имена колонок)
- При ошибке в БД вызывался `c.Abort()` в middleware, что приводило к неправильной обработке

## Решение

### 1. Исправление чтения токена (frontend/src/services/api.js):

**Было:**
```javascript
class ApiService {
  constructor() {
    this.accessToken = localStorage.getItem('accessToken');
  }

  getHeaders() {
    const headers = {
      'Content-Type': 'application/json',
    };
    if (this.accessToken) {
      headers['Authorization'] = `Bearer ${this.accessToken}`;
    }
    return headers;
  }
}
```

**Стало:**
```javascript
getHeaders() {
  const token = localStorage.getItem('accessToken');
  const headers = {
    'Content-Type': 'application/json',
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}
```

Теперь токен читается из localStorage при каждом запросе, а не один раз при создании экземпляра.

### 2. Исправление SQL-запросов (backend/src/handlers/handlers.go):
- Исправлены имена колонок: `total_participants` → `total_places`, `free_participants` → `free_places`
- Исправлен статус: `scheduled` → `available`
- Исправлены запросы в GetBookings: убраны несуществующие колонки `cancellation_fee`, `cancelled_at`

### 3. Исправление поведения при ошибках (backend/src/middleware/auth.go):
- Убран `c.Abort()` при ошибках авторизации, чтобы запросы продолжали обрабатываться

## Дата
2026-07-06

## Статус
Исправлен