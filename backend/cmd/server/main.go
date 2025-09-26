package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"xpired/internal/api"
	"xpired/internal/auth"
	"xpired/internal/config"
	database "xpired/internal/db"
	worker "xpired/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.RunMigrations("./migrations"); err != nil {
		log.Fatal("Failed to run database migrations:", err)
	}

	auth.Init(cfg)
	worker.InitQueue(cfg)

	repo := database.NewRepository(db)

	r := api.SetupRoutes(db)
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	workerServer := worker.NewServer(cfg)
	workerMux := worker.NewMux(repo)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting HTTP server on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting Asynq worker...")
		if err := workerServer.Run(workerMux); err != nil {
			log.Fatalf("Asynq worker failed: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Application started successfully")
	<-sigChan
	log.Println("Shutting down gracefully...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	workerServer.Shutdown()

	wg.Wait()
	log.Println("Application shutdown complete")
}
