// internal/models/employee.go

package models

import (
	"time"
)

type Employee struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`
	DepartmentID int      `gorm:"not null;index;constraint:OnDelete:CASCADE" json:"department_id"`
	FullName    string    `gorm:"type:varchar(200);not null" json:"full_name"`
	Position    string    `gorm:"type:varchar(200);not null" json:"position"`
	HiredAt     *time.Time `gorm:"type:date" json:"hired_at,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`

	Department *Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
}

func (Employee) TableName() string {
	return "employees"
}