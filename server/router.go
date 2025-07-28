package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func GetRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	return r
}
