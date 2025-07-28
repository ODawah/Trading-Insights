package main

import (
	"codnect.io/chrono"
	"context"
	"github.com/ODawah/Trading-Insights/models"
	"github.com/ODawah/Trading-Insights/persistence"
	"github.com/ODawah/Trading-Insights/server"
	"github.com/ODawah/Trading-Insights/services"
	"log"
	"net/http"
	"time"
)

func main() {
	pg, err := persistence.InitPG()
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	err = pg.AutoMigrate(models.User{}, models.WatchItem{}, models.Purchase{}, models.Currency{})
	if err != nil {
		panic(err)
	}

	redis, err := persistence.ConnectRedis()
	if err != nil {
		panic(err)
	}

	r := server.Routes(redis, pg)

	taskScheduler := chrono.NewDefaultTaskScheduler()
	_, err = taskScheduler.ScheduleAtFixedRate(func(ctx context.Context) {
		log.Println("Running scheduled task to fetch all currencies")
		snapshot, err := services.FetchAllCurrencies(redis, *pg)
		if err != nil {
			log.Printf("Error fetching all currencies: %v", err)
			return
		}
		log.Printf("Fetched %v currencies and cached them", snapshot)
	}, 30*time.Second)

	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
