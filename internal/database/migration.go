// internal/database/migration.go
package database

import (
	"errors"
	"fmt"
	"os"

	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

// RunMigrations запускает goose-миграции
func RunMigrations(db *gorm.DB) error {
	// Получаем sql.DB из GORM
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("Не удалось получить доступ к базе данных: %w", err)
	}

	// Путь к папке с миграциями
	migrationsDir := "./migrations"

	// Проверяем, существует ли папка
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return errors.New("Каталог миграций не найден: " + migrationsDir)
	}

	// Настраиваем goose
	goose.SetBaseFS(nil) // используем файловую систему

	// Запускаем миграции
	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		return fmt.Errorf("goose миграция не удалась: %w", err)
	}

	fmt.Println("Миграции успешно завершены.")
	return nil
}