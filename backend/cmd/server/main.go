package main

import (
	"log"
	"net/http"
	"xpired/internal/api"
	"xpired/internal/auth"
	"xpired/internal/config"
	"xpired/internal/db"
	database "xpired/internal/db"
	worker "xpired/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	db, err := db.NewConnection(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.RunMigrations("./migrations"); err != nil {
		log.Fatal("Failed to run database migrations:", err)
	}

	auth.Init(cfg)
	worker.InitQueue(cfg)

	server := worker.NewServer(cfg)
	repo := database.NewRepository(db)
	mux := worker.NewMux(repo)

	r := api.SetupRoutes(db)

	log.Println("Server listening on PORT :8080")
	http.ListenAndServe(":8080", r)

	log.Println("Starting Asynq worker...")
	if err := server.Run(mux); err != nil {
		log.Fatalf("could not run asynq server: %v", err)
	}
}
