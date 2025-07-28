package server

import (
	"github.com/ODawah/Trading-Insights/handlers"
	"github.com/ODawah/Trading-Insights/persistence"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
	"net/http"
)

func Routes(redis persistence.RedisRepository, postgres *gorm.DB) *chi.Mux {

	r := GetRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Get("/currencies", handlers.GetAllCurrenciesHandler(redis))

	return r
}
