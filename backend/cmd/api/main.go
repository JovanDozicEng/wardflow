package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wardflow/backend/internal/config"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/internal/router"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
	"github.com/wardflow/backend/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize logger
	logger.Init(cfg.LogLevel)

	// Initialize database connection
	dbCfg := &database.Config{
		Host:            cfg.DBHost,
		Port:            cfg.DBPort,
		User:            cfg.DBUser,
		Password:        cfg.DBPassword,
		DBName:          cfg.DBName,
		SSLMode:         cfg.DBSSLMode,
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: time.Duration(cfg.DBConnMaxLifetime) * time.Minute,
		LogLevel:        cfg.LogLevel,
	}

	db, err := database.Connect(dbCfg)
	if err != nil {
		logger.Fatal("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Verify database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		logger.Fatal("failed to ping database: %v", err)
	}
	logger.Info("database connection verified")

	// Run database migrations
	if err := runMigrations(db); err != nil {
		logger.Fatal("failed to run migrations: %v", err)
	}

	// Initialize JWT service
	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTExpiration)
	logger.Info("JWT service initialized")

	// Initialize auth service
	authService := auth.NewService(db, jwtService)
	logger.Info("auth service initialized")

	// Initialize router with all dependencies
	r := router.New(db, jwtService, authService)
	logger.Info("router initialized")

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("starting server on port %d", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown: %v", err)
	}

	logger.Info("server stopped")
}

// runMigrations runs database migrations
func runMigrations(db *database.DB) error {
	logger.Info("running database migrations...")
	
	// Auto-migrate User model
	if err := db.AutoMigrate(&models.User{}); err != nil {
		return fmt.Errorf("failed to migrate User: %w", err)
	}

	logger.Info("database migrations completed")
	return nil
}
