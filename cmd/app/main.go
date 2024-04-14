package main

import (
	"banner-api/internal/cache"
	"banner-api/internal/config"
	"banner-api/internal/handlers"
	"banner-api/internal/repository/psql"
	"banner-api/internal/service"
	"fmt"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"net/http"
	"time"
)

func main() {

	connectionStr := config.DBConnectionString()

	db := psql.NewPostgresDB(connectionStr)
	defer db.Close()
	err := psql.PGDBInit(db)
	if err != nil {
		panic(err)
	}

	bannerCache := cache.NewBannerCache()

	go bannerCache.StartCleanup()

	bannerRepo := psql.NewBannerRepo(db)

	bannerService := service.NewBannerService(bannerRepo, bannerCache)

	bannerHandler := handlers.NewBannerHandler(bannerService)

	router := chi.NewRouter()

	bannerHandler.RegisterHandlers(router)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	fmt.Println("Listening on port 8080")

	server.ListenAndServe()
}
