package main

import (
	"log"
	"net/http"
	"os"

	"github.com/bullshit-wtf/server/internal/config"
	"github.com/bullshit-wtf/server/internal/db"
	"github.com/bullshit-wtf/server/internal/game"
	"github.com/bullshit-wtf/server/internal/handlers"
	"github.com/bullshit-wtf/server/internal/hub"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	questionStore := game.NewDBQuestionStore(database)
	pinGen := game.NewPinGenerator()
	h := hub.NewHub(questionStore, pinGen)
	go h.Run()

	mux := handlers.NewRouter(h, database)

	addr := ":" + cfg.Port
	log.Printf("server listening on %s", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Serve static frontend files if the directory exists
	if _, err := os.Stat("./static"); err == nil {
		log.Println("serving static files from ./static")
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
