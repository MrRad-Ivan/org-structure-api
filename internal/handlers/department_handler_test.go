package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"org-structure-api/internal/models"
)

//Проверяем успешное создание нового подразделения
func TestCreateDepartment(t *testing.T) {
	db := setupTestDB(t)
	router := NewRouter(db)

	reqBody := `{"name": "Тестовый Отдел"}`
	req := httptest.NewRequest(http.MethodPost, "/departments", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "должен возвращаться статус 201 Created")
}

// Проверяем создание сотрудника в существующем подразделении
func TestCreateEmployee(t *testing.T) {
	db := setupTestDB(t)
	router := NewRouter(db)

	dept := &models.Department{Name: "Отдел Для Теста"}
	db.Create(dept)

	reqBody := `{"full_name": "Петр Петров", "position": "Разработчик"}`
	url := "/departments/" + strconv.Itoa(dept.ID) + "/employees"

	req := httptest.NewRequest(http.MethodPost, url, bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "должен возвращаться статус 201 Created")
}

// Проверяем получение подразделения с деревом 
func TestGetDepartmentWithDepth(t *testing.T) {
	// Проверяем получение подразделения с деревом 
	db := setupTestDB(t)
	router := NewRouter(db)

	// Cоздаём департамент
	db.Create(&models.Department{Name: "Компания"})

	req := httptest.NewRequest(http.MethodGet, "/departments/1?depth=2&include_employees=true", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "должен возвращаться статус 200 OK")
}

// Проверяем запрет создания дубликата названия в одном родителе
func TestDuplicateDepartmentNameReturns409(t *testing.T) {
	db := setupTestDB(t)
	router := NewRouter(db)

	reqBody := `{"name": "Тестовый Отдел"}`
	req := httptest.NewRequest(http.MethodPost, "/departments", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Повторная попытка
	req = httptest.NewRequest(http.MethodPost, "/departments", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code, "должен возвращаться 409 Conflict")
}

// Проверяем, что нельзя добавить сотрудника в несуществующее подразделение
func TestCreateEmployeeInNonExistentDepartmentReturns404(t *testing.T) {

	db := setupTestDB(t)
	router := NewRouter(db)

	reqBody := `{"full_name": "Тест", "position": "Тест"}`
	req := httptest.NewRequest(http.MethodPost, "/departments/999/employees", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code, "должен возвращаться 404 Not Found")
}

// ВСПОМОГАТЕЛЬНАЯ ФУНКЦИЯ
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.Department{}, &models.Employee{})
	require.NoError(t, err)

	return db
}