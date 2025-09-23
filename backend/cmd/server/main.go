package main

import (
	"log"
	"net/http"
	"xpired/internal/api"
	"xpired/internal/auth"
	"xpired/internal/config"
	"xpired/internal/db"
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

	r := api.SetupRoutes(db)

	log.Println("Server listening on PORT :8080")
	http.ListenAndServe(":8080", r)
}
