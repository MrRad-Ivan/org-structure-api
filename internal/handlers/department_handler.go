// internal/handlers/department_handler.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"org-structure-api/internal/models"
	"org-structure-api/internal/repository"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type DepartmentHandler struct {
	deptRepo *repository.DepartmentRepository
	empRepo  *repository.EmployeeRepository
}

func NewDepartmentHandler(db *gorm.DB) *DepartmentHandler {
	return &DepartmentHandler{
		deptRepo: repository.NewDepartmentRepository(db),
		empRepo:  repository.NewEmployeeRepository(db),
	}
}

// POST /departments/
func (h *DepartmentHandler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		ParentID *int   `json:"parent_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Неверный формат JSON"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" || len(req.Name) > 200 {
		http.Error(w, `{"error":"Название обязательно и должно быть от 1 до 200 символов"}`, http.StatusBadRequest)
		return
	}

	if req.ParentID != nil {
		if !h.deptRepo.Exists(*req.ParentID) {
			http.Error(w, `{"error":"Родительское подразделение не найдено"}`, http.StatusNotFound)
			return
		}
	}

	dept := &models.Department{
		Name:     req.Name,
		ParentID: req.ParentID,
	}

	if err := h.deptRepo.Create(dept); err != nil {
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "idx_dept_name_parent") ||
			err.Error() == "подразделение с таким названием уже существует в этом родителе" {
			http.Error(w, `{"error":"Подразделение с таким названием уже существует в этом родителе"}`, http.StatusConflict)
			return
		}

		http.Error(w, `{"error":"Не удалось создать подразделение"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dept)
}
// GET /departments/{id}?depth=2&include_employees=true
func (h *DepartmentHandler) GetDepartment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"Неверный формат ID"}`, http.StatusBadRequest)
		return
	}

	// Получаем параметры
	depth := 1
	if d := r.URL.Query().Get("depth"); d != "" {
		if val, err := strconv.Atoi(d); err == nil && val >= 1 && val <= 5 {
			depth = val
		}
	}

	includeEmployees := true
	if inc := r.URL.Query().Get("include_employees"); inc == "false" {
		includeEmployees = false
	}

	dept, err := h.deptRepo.GetByIDWithDepth(id, depth, includeEmployees)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"Подразделение не найдено"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"Ошибка базы данных"}`, http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	response := map[string]interface{}{
		"department": dept,
		"employees":  dept.Employees,
		"children":   dept.Children,
	}

	// Чтобы не дублировать данные внутри department
	dept.Children = nil   // очищаем, чтобы не было дублирования
	if !includeEmployees {
		dept.Employees = nil
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// POST /departments/{id}/employees/
func (h *DepartmentHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	departmentID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"Неверный формат ID подразделения"}`, http.StatusBadRequest)
		return
	}

	if !h.deptRepo.Exists(departmentID) {
		http.Error(w, `{"error":"Подразделение не найдено"}`, http.StatusNotFound)
		return
	}

	var req struct {
		FullName string `json:"full_name"`
		Position string `json:"position"`
		HiredAt  string `json:"hired_at"` 
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Неверный формат JSON"}`, http.StatusBadRequest)
		return
	}

	if req.FullName == "" || len(req.FullName) > 200 || req.Position == "" || len(req.Position) > 200 {
		http.Error(w, `{"error":"full_name и position обязательны (1-200 символов)"}`, http.StatusBadRequest)
		return
	}

	employee := &models.Employee{
		DepartmentID: departmentID,
		FullName:     req.FullName,
		Position:     req.Position,
	}

	// Парсим дату, если она пришла
	if req.HiredAt != "" {
		t, parseErr := time.Parse("2006-01-02", req.HiredAt)
		if parseErr != nil {
			http.Error(w, `{"error":"Неверный формат hired_at. Используйте YYYY-MM-DD"}`, http.StatusBadRequest)
			return
		}
		employee.HiredAt = &t
	}

	if err := h.empRepo.Create(employee); err != nil {
		http.Error(w, `{"error":"Не удалось создать сотрудника"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(employee)
}

// DELETE /departments/{id}
func (h *DepartmentHandler) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"Неверный формат ID"}`, http.StatusBadRequest)
		return
	}

	mode := r.URL.Query().Get("mode")
	if mode != "cascade" && mode != "reassign" {
		http.Error(w, `{"error":"mode должен быть cascade или reassign"}`, http.StatusBadRequest)
		return
	}

	var reassignTo *int
	if mode == "reassign" {
		if reassignStr := r.URL.Query().Get("reassign_to_department_id"); reassignStr != "" {
			if val, err := strconv.Atoi(reassignStr); err == nil {
				reassignTo = &val
			}
		}
		if reassignTo == nil {
			http.Error(w, `{"error":"reassign_to_department_id обязателен при mode=reassign"}`, http.StatusBadRequest)
			return
		}
	}

	if err := h.deptRepo.Delete(id, mode, reassignTo); err != nil {
		if err.Error() == "подразделение не найдено" {
			http.Error(w, `{"error":"Подразделение не найдено"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PATCH /departments/{id}
func (h *DepartmentHandler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"Неверный формат ID"}`, http.StatusBadRequest)
		return
	}

	// Используем map, чтобы точно понять, какие поля были в JSON
	var raw map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, `{"error":"Неверный формат JSON"}`, http.StatusBadRequest)
		return
	}

	if len(raw) == 0 {
		http.Error(w, `{"error":"Должно быть указано хотя бы одно поле (name или parent_id)"}`, http.StatusBadRequest)
		return
	}

	// Преобразуем в нормальную структуру
	var req struct {
		Name     *string `json:"name,omitempty"`
		ParentID *int    `json:"parent_id"`
	}
	json.NewDecoder(r.Body).Decode(&req) // повторно парсим (body уже прочитан)

	if err := h.deptRepo.Update(id, req.Name, req.ParentID); err != nil {
		http.Error(w, `{"error":"Не удалось обновить подразделение"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}