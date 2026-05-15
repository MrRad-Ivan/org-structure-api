API Организационной структуры

REST API для управления иерархией подразделений и сотрудниками компании.

Стек технологий
- Go 1.25
- Router: chi/v5
- ORM: GORM
- База данных: PostgreSQL
- Миграции: Goose
- Контейнеризация: Docker + docker-compose
- Тестирование: testify

Быстрый запуск

1. Клонировать репозиторий
   git clone https://github.com/MrRad-Ivan/org-structure-api.git
   cd org-structure-api

2. Запустить приложение
   docker-compose up --build

После запуска API доступно по адресу:
http://localhost:8080

Основные эндпоинты

Метод     Эндпоинт                                              Описание
POST      /departments                                          Создать подразделение
POST      /departments/{id}/employees                           Добавить сотрудника в подразделение
GET       /departments/{id}?depth={n}&include_employees={bool} Получить подразделение + дерево + сотрудников
PATCH     /departments/{id}                                     Обновить название или parent_id
DELETE    /departments/{id}?mode=cascade                       Удалить подразделение и все дочерние (каскадно)
DELETE    /departments/{id}?mode=reassign&reassign_to_department_id={id} Удалить подразделение и переназначить сотрудников

Полный список тестов для проверки

1. Создать корневой департамент
   POST http://localhost:8080/departments
   {"name": "Компания"}

2. Создать IT Отдел
   POST http://localhost:8080/departments
   {"name": "IT Отдел", "parent_id": 1}

3. Создать Backend Team
   POST http://localhost:8080/departments
   {"name": "Backend Team", "parent_id": 2}

4-5. Создать сотрудников
   POST http://localhost:8080/departments/2/employees
   {"full_name": "Иван Иванов", "position": "Senior Go Developer", "hired_at": "2025-01-15"}
   {"full_name": "Анна Смирнова", "position": "Team Lead"}

6. Получить дерево
   GET http://localhost:8080/departments/1?depth=3&include_employees=true

7-9. Обновление (PATCH)
   PATCH http://localhost:8080/departments/2 → {"name": "IT Департамент"}
   PATCH http://localhost:8080/departments/3 → {"parent_id": 1}
   PATCH http://localhost:8080/departments/3 → {"parent_id": null}

10-11. Удаление
   DELETE http://localhost:8080/departments/3?mode=cascade
   DELETE http://localhost:8080/departments/2?mode=reassign&reassign_to_department_id=1

Negative тесты (должны возвращать ошибку)
12. Дубликат названия → 409
    POST http://localhost:8080/departments
    {"name": "IT Департамент", "parent_id": 1}

13. Сотрудник в несуществующем департаменте → 404
    POST http://localhost:8080/departments/999/employees

14. Создать цикл → 409
    PATCH http://localhost:8080/departments/2
    {"parent_id": 2}

15. PATCH без полей → 400
    PATCH http://localhost:8080/departments/1
    {}

16. DELETE reassign без reassign_to → 400
    DELETE http://localhost:8080/departments/1?mode=reassign

Особенности реализации
- Уникальность названия подразделения внутри одного родителя
- Защита от создания циклов в дереве подразделений
- Поддержка параметра depth (максимум 5 уровней)
- Все сообщения об ошибках возвращаются на русском языке
- Чистая архитектура (handlers → repository → models)

      org-structure-api/
         cmd/
            app/
               main.go                  # Точка входа приложения
   
         internal/                        # Основной код проекта
            config/
               config.go                # Конфигурация (.env)

            database/
               db.go                    # Подключение к PostgreSQL
               migration.go             # Запуск миграций Goose
      
            handlers/
               router.go                # Настройка всех маршрутов (chi)
               department_handler.go    # Все HTTP-обработчики

            models/
               department.go            # Модель Подразделения
               employee.go              # Модель Сотрудника

           repository/
               department_repository.go # Логика работы с БД по подразделениям
               employee_repository.go   # Логика работы с БД по сотрудникам
         
           migrations/                      # Миграции базы данных (Goose)
               00001_create_departments.sql
               00002_create_employees.sql

      .env                             # Настройки окружения (локальные)
      .env.example                     # Пример настроек для GitHub
      docker-compose.yml               # Запуск PostgreSQL + приложения
      Dockerfile                       # Сборка Docker-образа
      go.mod
      go.sum
      README.md

Запуск тестов

go test ./internal/handlers -v
