package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"codnect.io/chrono"
	"github.com/ODawah/Trading-Insights/database"
	"github.com/ODawah/Trading-Insights/models"
	"github.com/ODawah/Trading-Insights/repository"
	"github.com/ODawah/Trading-Insights/server"
	"github.com/ODawah/Trading-Insights/services"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables early
	_ = godotenv.Load(".env")

	pg, err := database.ConnectPostgreSQL()
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	err = pg.AutoMigrate(
		models.User{},
		models.WatchItem{},
		models.Currency{},
		models.UserLedgerEntry{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	if err := database.EnsureTimescale(pg); err != nil {
		log.Printf("TimescaleDB not enabled (continuing without hypertables): %v", err)
	}

	redis, err := database.ConnectRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	//Initialize Repositories
	currencyRepo := repository.NewCurrencyRepository(redis, pg)
	userRepo := repository.NewUserRepository(pg)
	ledgerRepo := repository.NewLedgerRepository(pg)
	ingestionService := services.NewIngestionAPIClient(currencyRepo)
	currencyService := services.NewCurrencyService(currencyRepo)
	authService := services.NewAuthService(userRepo)
	watchListService := services.NewWatchListService(repository.NewWatchListRepository(pg))
	ledgerService := services.NewLedgerService(ledgerRepo)
	analyticsService := services.NewAnalyticsService(currencyRepo, ledgerRepo)

	serverServices := &server.Services{
		Ingestion: ingestionService,
		Currency:  currencyService,
		Auth:      authService, // ← New
		WatchList: watchListService, // ← New
		Ledger:    ledgerService,
		Analytics: analyticsService,
	}

	r := server.Routes(serverServices)

	taskScheduler := chrono.NewDefaultTaskScheduler()
	// Check for Scheduler Error
	_, err = taskScheduler.ScheduleAtFixedRate(func(ctx context.Context) {
		log.Println("Running scheduled task to fetch all currencies")
		snapshot, err := ingestionService.FetchRates(ctx)
		if err != nil {
			log.Printf("Error fetching currencies: %v", err)
			return
		}
		log.Printf("Fetched %v currencies", snapshot)
	}, 30*time.Second)

	if err != nil { // ← Check the scheduler scheduling error here
		log.Fatalf("Failed to schedule task: %v", err)
	}

	err = http.ListenAndServe(":8000", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
