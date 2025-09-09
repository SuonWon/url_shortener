package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"github.com/url_shortener/internal/database"
	handler "github.com/url_shortener/internal/handlers"
	router "github.com/url_shortener/internal/http"
	repo "github.com/url_shortener/internal/repos"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("Environment variable not found")
	}

	db := database.ConnectionDatabase()

	if db == nil {
		log.Fatal("db is nil")
	}
	port := os.Getenv("PORT")

	userRepo := repo.NewUserRepository(db)
	userHandler := handler.NewUserHandler(userRepo)
	domainRepo := repo.NewDomainRepository(db)
	domainHandler := handler.NewDomainRepository(domainRepo)
	linkRepo := repo.NewShortLinkRepository(db)
	linkHandler := handler.NewLinkHandler(linkRepo, domainRepo)
	router := router.New(userHandler, domainHandler, linkHandler)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("Listening on %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %v", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	log.Println("bye 👋")
}
