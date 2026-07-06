# Test Layout Generate — Проверки макета и верстки

> **Роль:** Senior QA Engineer  
> **Назначение:** Шаблон для генерации проверок верстки: grid, flexbox, отступы, размеры  
> **Применение:** Используется LLM для создания проверок соответствия HTML/CSS макету

---

## Переменные

| Переменная | Тип | Описание | Пример |
|------------|-----|----------|--------|
| `$COMPONENT` | string | Имя компонента/секции | Header, Card, Modal |
| `$LAYOUT_TYPE` | string | Тип компоновки | flex, grid, block |
| `$CONTAINER_SPEC` | object | Спецификация контейнера | max-width, padding |
| `$CHILD_ELEMENTS` | array | Дочерние элементы | [title, subtitle, actions] |
| `$SPACING_SYSTEM` | object | Система отступов | xs: 4px, sm: 8px |
| `$RESPONSIVE` | object | Правила адаптивности | breakpoints, orders |

---

## Правила использования

### Типы компоновки

| Тип | Проверяем | Методы |
|-----|-----------|--------|
| **Flexbox** | direction, justify-content, align-items, gap, wrap | Computed styles |
| **Grid** | grid-template-*, grid-gap, grid-area | Computed styles |
| **Block** | width, margin, padding, float | Computed styles |
| **Position** | position, top, left, right, bottom | Computed styles |

### Проверяемые свойства

1. **Spacing (отступы)**
   - margin (top, right, bottom, left)
   - padding (top, right, bottom, left)
   - gap (row-gap, column-gap)

2. **Sizing (размеры)**
   - width, height
   - min-width, min-height
   - max-width, max-height

3. **Positioning (позиционирование)**
   - position (static, relative, absolute, fixed, sticky)
   - top, right, bottom, left
   - z-index

4. **Alignment (выравнивание)**
   - flex-direction
   - justify-content
   - align-items / align-self
   - text-align

---

## Задачи

### Задача 1: Проверка контейнера

**Фокус:** Wrapper, section, card container

| Проверка | Ожидаемое | Приоритет |
|----------|-----------|-----------|
| Max-width | 1200px или согласно макету | high |
| Padding | Согласно spacing system | high |
| Margin auto | Центрирование при max-width | medium |
| Background | Цвет согласно схеме | high |

### Задача 2: Проверка сетки (Grid/Flex)

**Фокус:** Расположение элементов в контейнере

| Проверка | Ожидаемое | Приоритет |
|----------|-----------|-----------|
| Direction | row / column | high |
| Gap | Значение из spacing system | high |
| Justify-content | flex-start / center / space-between | high |
| Align-items | stretch / flex-start / center | medium |
| Grid columns | Количество колонок | high |

### Задача 3: Проверка отступов между элементами

**Фокус:** Расстояние между child elements

| Проверка | Ожидаемое | Приоритет |
|----------|-----------|-----------|
| Первый элемент от границы | 0px (или согласно макету) | medium |
| Расстояние между элементами | gap или margin-bottom | high |
| Последний элемент от границы | 0px | low |

### Задача 4: Проверка адаптивности

**Фокус:** Изменение layout на разных экранах

| Проверка | Условие | Ожидаемое |
|----------|---------|-----------|
| Grid columns | 320px | 1 колонка |
| Grid columns | 768px | 2 колонки |
| Grid columns | 1024px | 3-4 колонки |
| Padding | 320px | 16px |
| Padding | 1024px | 24px |

---

## Шаблон вывода

```markdown
## [COMPONENT] — Проверки верстки

### Контейнер

| № | Свойство | CSS-свойство | Ожидаемое | Метод | Приоритет |
|---|----------|--------------|-----------|-------|-----------|
| 1 | Ширина контейнера | max-width | 400px | computed | high |

### Grid / Flex

| № | Свойство | CSS-свойство | Ожидаемое | Метод | Приоритет |
|---|----------|--------------|-----------|-------|-----------|
| 1 | Тип компоновки | display | flex | computed | high |

### Отступы

| № | Элемент | CSS-свойство | Ожидаемое | Метод | Приоритет |
|---|---------|--------------|-----------|-------|-----------|
| 1 | .container | padding | 16px | computed | high |

### Адаптивность

| Breakpoint | Проверка | Ожидаемое | Фактическое | Статус |
|------------|----------|-----------|-------------|--------|
| 320px | .grid | grid-template-columns | 1fr | - |
```

---

## Пример использования

**Входные данные:**
- COMPONENT: Карточка слота тренировки
- LAYOUT_TYPE: flex, column
- CONTAINER_SPEC: { max-width: 100%, padding: 16px, border-radius: 12px }
- SPACING_SYSTEM: { xs: 4px, sm: 8px, md: 16px, lg: 24px }

---

## Пример вывода

## Card Slot — Проверки верстки

### Контейнер

| № | Свойство | CSS-свойство | Ожидаемое | Метод | Приоритет |
|---|----------|--------------|-----------|-------|-----------|
| 1 | Тип дисплея | display | flex | computed | high |
| 2 | Направление | flex-direction | column | computed | high |
| 3 | Max-width | max-width | 100% | computed | high |
| 4 | Padding | padding | 16px | computed | high |
| 5 | Border-radius | border-radius | 12px | computed | high |
| 6 | Background | background-color | #FFFFFF | computed | high |
| 7 | Box-shadow | box-shadow | 0 2px 8px rgba(0,0,0,0.08) | computed | medium |

### Grid / Flex

| № | Свойство | CSS-свойство | Ожидаемое | Метод | Приоритет |
|---|----------|--------------|-----------|-------|-----------|
| 1 | Выравнивание | justify-content | space-between | computed | high |
| 2 | Выравнивание по оси | align-items | center | computed | medium |
| 3 | Gap между элементами | gap | 12px | computed | high |

### Отступы внутри

| № | Элемент | CSS-свойство | Ожидаемое | Метод | Приоритет |
|---|---------|--------------|-----------|-------|-----------|
| 1 | .card-header | margin-bottom | 8px | computed | medium |
| 2 | .card-info | margin-bottom | 12px | computed | medium |
| 3 | .card-actions | margin-top | 16px | computed | medium |
| 4 | .zone-badge | margin-right | 8px | computed | low |

### Размеры

| № | Элемент | Свойство | Ожидаемое | Метод | Приоритет |
|---|---------|----------|-----------|-------|-----------|
| 1 | .card | min-height | 120px | computed | low |
| 2 | .zone-badge | height | 24px | computed | low |

### Адаптивность

| Breakpoint | Проверка | CSS-свойство | Ожидаемое | Статус |
|------------|----------|--------------|-----------|--------|
| 320px | Card padding | padding | 12px | ⏳ |
| 320px | Font size title | font-size | 16px | ⏳ |
| 768px | Card padding | padding | 16px | ⏳ |
| 768px | Card | max-width | 100% | ⏳ |

### Визуальные состояния

| Состояние | Проверка | Ожидаемое | Приоритет |
|-----------|----------|-----------|-----------|
| Default | Background | #FFFFFF | high |
| Hover | Box-shadow | 0 4px 12px rgba(0,0,0,0.12) | medium |
| Hover | Transform | translateY(-2px) | medium |
| Active | Transform | translateY(0) | low |
| Disabled | Opacity | 0.5 | low |

---

## CSS селекторы для проверок

```css
/* Основные селекторы */
.component { }
.component > .child { }
.component .descendant { }

/* Псевдоклассы */
.component:hover { }
.component:focus { }
.component:active { }
.component:disabled { }

/* Псевдоэлементы */
.component::before { }
.component::after { }
```

---

## Инструменты проверки

| Инструмент | Назначение | Тип проверки |
|------------|------------|--------------|
| Chrome DevTools > Computed | Значения CSS | Ручная |
| Cypress `.should('have.css')` | Автоматизация | Авто |
| Playwright `.toHaveCSS()` | Автоматизация | Авто |
| Jest + jest-styled-components | Стили в unit-тестах | Авто |
| Chromatic | Визуальная регрессия | Скриншоты |

---

## Требования к LLM

1. Использовать CSS-свойства для проверок
2. Применять computed styles method
3. Учитывать spacing system из дизайн-системы
4. Проверять все состояния: default, hover, active, focus, disabled
5. Генерировать адаптивные проверки для всех breakpoints
6. Использовать конкретные селекторы элементов