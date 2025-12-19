package server

import (
	"net/http"
	"time"

	"github.com/ODawah/Trading-Insights/handlers"
	"github.com/ODawah/Trading-Insights/middleware"
	"github.com/ODawah/Trading-Insights/services"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type Services struct {
	Currency  services.CurrencyService
	Ingestion services.IngestionService
	Auth      services.AuthService
	WatchList services.WatchListService
	Ledger    services.LedgerService
	Analytics services.AnalyticsService
}

func Routes(services *Services) *chi.Mux {
	r := GetRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	authHandler := handlers.NewAuthHandler(services.Auth)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.Signup)
		r.Post("/login", authHandler.Login)
	})

	currenciesHandler := handlers.NewCurrenciesHandler(services.Currency)
	r.Get("/currencies/latest", currenciesHandler.GetAllCurrenciesHandler())
	r.Get("/currencies/{ticker}/history", currenciesHandler.GetHistoryHandler())
	r.Get("/currencies/{ticker}/candles", currenciesHandler.GetCandlesHandler())
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware) // Apply JWT middleware

	})
	watchListHandler := handlers.NewWatchListHandler(services.WatchList)
	r.Route("/watchlist", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Post("/add", watchListHandler.AddToWatchlist)
		r.Post("/remove", watchListHandler.RemoveFromWatchlist)
		r.Get("/", watchListHandler.GetWatchlist)
	})

	ledgerHandler := handlers.NewLedgerHandler(services.Ledger)
	r.Route("/ledger", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Post("/exchange", ledgerHandler.RecordExchange)
		r.Get("/", ledgerHandler.ListEntries)
		r.Get("/trade/{tradeID}", ledgerHandler.GetTrade)
	})

	analyticsHandler := handlers.NewAnalyticsHandler(services.Analytics)
	r.Route("/analytics", func(r chi.Router) {
		r.Get("/cross", analyticsHandler.CrossRate)
		r.Get("/chart", analyticsHandler.ChartCross)
		r.Get("/convert", analyticsHandler.ConvertAt)
		r.Get("/correlation", analyticsHandler.Correlation)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Get("/portfolio/value", analyticsHandler.PortfolioValue)
			r.Get("/portfolio/history", analyticsHandler.PortfolioHistory)
		})
	})

	return r
}
