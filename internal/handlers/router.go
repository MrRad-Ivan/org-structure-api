// internal/handlers/router.go
package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	deptHandler := NewDepartmentHandler(db)

	r.Route("/departments", func(r chi.Router) {
		r.Post("/", deptHandler.CreateDepartment)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", deptHandler.GetDepartment)
			r.Delete("/", deptHandler.DeleteDepartment)
			r.Patch("/", deptHandler.UpdateDepartment)   

			r.Route("/employees", func(r chi.Router) {
				r.Post("/", deptHandler.CreateEmployee)
			})
		})
	})

	return r
}