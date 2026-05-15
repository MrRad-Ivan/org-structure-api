// internal/repository/department_repository.go

package repository

import (
	"errors"
	"org-structure-api/internal/models"

	"gorm.io/gorm"
)

type DepartmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

func (r *DepartmentRepository) Create(dept *models.Department) error {
	// Проверка уникальности имени внутри parent
	if err := r.checkNameUnique(dept.Name, dept.ParentID); err != nil {
		return err
	}
	return r.db.Create(dept).Error
}

// Проверка на цикл (простая реализация)
func (r *DepartmentRepository) wouldCreateCycle(deptID int, newParentID int) bool {
	if deptID == newParentID {
		return true
	}

	visited := make(map[int]bool)
	current := newParentID

	for current != 0 && !visited[current] {
		visited[current] = true

		var p models.Department
		if err := r.db.Select("parent_id").First(&p, current).Error; err != nil {
			break
		}
		if p.ParentID == nil {
			break
		}
		current = *p.ParentID

		if current == deptID {
			return true
		}
	}
	return false
}

// Update — обновление подразделения
func (r *DepartmentRepository) Update(id int, name *string, parentID *int) error {
	var dept models.Department
	if err := r.db.First(&dept, id).Error; err != nil {
		return err
	}

	// Обновляем имя
	if name != nil {
		if err := r.checkNameUnique(*name, dept.ParentID); err != nil {
			return err
		}
		dept.Name = *name
	}

	// Обновляем parent_id
	if parentID != nil {
		newParent := *parentID

		if newParent == id {
			return errors.New("нельзя сделать подразделение родителем самого себя")
		}

		if r.wouldCreateCycle(id, newParent) {
			return errors.New("перемещение создаст цикл в дереве")
		}

		dept.ParentID = &newParent
	} 

	return r.db.Save(&dept).Error
}

func (r *DepartmentRepository) checkNameUnique(name string, parentID *int) error {
	var count int64
	query := r.db.Model(&models.Department{}).Where("name = ?", name)

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	query.Count(&count)

	if count > 0 {
		return errors.New("подразделение с таким названием уже существует в этом родителе")
	}
	return nil
}

func (r *DepartmentRepository) GetByID(id int) (*models.Department, error) {
	var dept models.Department
	err := r.db.Preload("Employees").Preload("Children").First(&dept, id).Error
	return &dept, err
}

func (r *DepartmentRepository) Exists(id int) bool {
	var count int64
	r.db.Model(&models.Department{}).Where("id = ?", id).Count(&count)
	return count > 0
}

// GetByIDWithDepth — рекурсивная загрузка дерева до указанной глубины
func (r *DepartmentRepository) GetByIDWithDepth(id int, depth int, includeEmployees bool) (*models.Department, error) {
	var dept models.Department

	query := r.db.Preload("Parent")

	if includeEmployees {
		query = query.Preload("Employees")
	}

	err := query.First(&dept, id).Error
	if err != nil {
		return nil, err
	}

	if depth > 1 {
		if err := r.loadChildrenRecursive(&dept, depth-1, includeEmployees); err != nil {
			return nil, err
		}
	}

	return &dept, nil
}

func (r *DepartmentRepository) loadChildrenRecursive(dept *models.Department, remainingDepth int, includeEmployees bool) error {
	if remainingDepth <= 0 {
		return nil
	}

	var children []models.Department
	query := r.db.Where("parent_id = ?", dept.ID)

	if includeEmployees {
		query = query.Preload("Employees")
	}

	if err := query.Find(&children).Error; err != nil {
		return err
	}

	dept.Children = children

	for i := range dept.Children {
		if err := r.loadChildrenRecursive(&dept.Children[i], remainingDepth-1, includeEmployees); err != nil {
			return err
		}
	}
	return nil
}

// Delete — удаление подразделения
func (r *DepartmentRepository) Delete(id int, mode string, reassignTo *int) error {
	// Проверяем существование
	var count int64
	r.db.Model(&models.Department{}).Where("id = ?", id).Count(&count)

	if count == 0 {
		return errors.New("подразделение не найдено")
	}

	if mode == "cascade" {
		return r.deleteCascade(id)
	}
	if mode == "reassign" {
		if reassignTo == nil {
			return errors.New("reassign_to_department_id обязателен")
		}
		return r.deleteWithReassign(id, *reassignTo)
	}

	return errors.New("неизвестный mode удаления")
}

func (r *DepartmentRepository) deleteCascade(id int) error {
	return r.db.Select("Employees").Delete(&models.Department{}, id).Error
}

func (r *DepartmentRepository) deleteWithReassign(id int, reassignTo int) error {
	// Проверяем, существует ли целевой департамент
	if !r.Exists(reassignTo) {
		return errors.New("подразделение для переназначения не найдено")
	}

	// Переносим всех сотрудников
	if err := r.db.Model(&models.Employee{}).
		Where("department_id = ?", id).
		Update("department_id", reassignTo).Error; err != nil {
		return err
	}

	// Удаляем подразделение
	return r.db.Delete(&models.Department{}, id).Error
}

