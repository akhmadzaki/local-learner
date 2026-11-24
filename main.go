package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/akhmadzaki/go-backend/api/routes"
	"github.com/akhmadzaki/go-backend/config"
	"github.com/akhmadzaki/go-backend/db"
)

// @title           Local AI Knowledge Base API
// @version         1.0
// @description     RAG-based knowledge base with document ingestion, hybrid search, and agentic chat.
// @termsOfService  http://swagger.io/terms/
//
// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@swagger.io
//
// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host            localhost:8080
// @BasePath        /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in               header
// @name             Authorization
//
// @externalDocs.description  OpenAPI
// @externalDocs.url         https://swagger.io/resources/open-api/

func main() {
	// Optional: run migrations and exit without starting the HTTP server.
	migrateOnly := flag.Bool("migrate-only", false, "run database migrations and exit")
	flag.Parse()

	// 1. Load configuration
	cfg := config.Load()
	log.Printf("Starting %s server on port %d (env: %s)", cfg.App.Version, cfg.App.Port, cfg.App.Env)

	// 2. Connect to PostgreSQL (pgvector)
	database, err := db.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// 3. Run schema migrations (golang-migrate)
	if *migrateOnly || cfg.Database.AutoMigrate {
		log.Println("Running database migrations...")
		if err := db.MigrateUp(&cfg.Database); err != nil {
			log.Fatalf("Failed to run database migrations: %v", err)
		}
		log.Println("Database migrations applied")
	}

	// 3a. Exit early when only migrations were requested.
	if *migrateOnly {
		return
	}

	// 4. Register API routes with all dependencies wired
	router := routes.RegisterRoutes(&cfg, database)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 5. Start HTTP server
	addr := ":" + strconv.Itoa(cfg.App.Port)
	log.Printf("Server listening on %s", addr)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
