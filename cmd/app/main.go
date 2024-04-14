package main

import (
	"banner-api/internal/cache"
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

	connStr := "user=postgres dbname=banners sslmode=disable password=123456 host=localhost port=5432"
	db, err := psql.NewPostgresDB(connStr)
	if err != nil {
		panic("failed to connect to database")
	}
	defer db.Close()
	psql.PGDBInit(db)
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
