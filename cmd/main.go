package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"rest-api/config"
	"rest-api/internal/db"
	"rest-api/internal/handlers"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on system env")
	}

	cfg := config.Load()
	ctx := context.Background()

	router := chi.NewRouter()

	pool, err := db.InitDatabase(ctx, cfg)
	if err != nil {
		log.Fatalf("database : %v\n", err)
	}
	defer pool.Close()

	api := handlers.NewAPI(pool)

	if err := api.SeedAdmin(cfg); err != nil {
		log.Fatalf("seed admin: %v\n", err)
	}

	api.RegisterAll(router)

	srv := &http.Server{
		Addr:         cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		fmt.Println("server started on", cfg.Port)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	fmt.Println("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err = srv.Shutdown(ctx)
	if err != nil {
		log.Printf("graceful shutdown failed: %v, forcing close", err)
		if err2 := srv.Close(); err2 != nil {
			log.Printf("forced close failed: %v", err2)
		}
	}

	fmt.Println("server stopped")
}
