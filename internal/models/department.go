// internal/models/department.go
package models

import (
	"time"
)

type Department struct {
	ID        int           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string        `gorm:"type:varchar(200);not null;uniqueIndex:idx_dept_name_parent" json:"name"`
	ParentID  *int          `gorm:"index;constraint:OnDelete:CASCADE" json:"parent_id,omitempty"`
	CreatedAt time.Time     `gorm:"autoCreateTime" json:"created_at"`

	// Связи
	Children   []Department `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Employees  []Employee   `gorm:"foreignKey:DepartmentID" json:"employees,omitempty"`
	Parent     *Department  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
}

func (Department) TableName() string {
	return "departments"
}