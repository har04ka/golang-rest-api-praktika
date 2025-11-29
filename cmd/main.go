package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"rest-api/config"
	"rest-api/internal/db"
	"rest-api/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	router := chi.NewRouter()

	pool, err := db.InitDatabase(ctx, cfg)
	if err != nil {
		log.Fatalf("database : %v\n", err)
	}
	defer pool.Close()

	api := handlers.NewAPI(pool)

	api.RegisterAll(router)

	fmt.Println("Starting server on port", cfg.Port)
	log.Fatal(http.ListenAndServe(cfg.Port, router))
}
