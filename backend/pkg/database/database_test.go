package database

import (
	"context"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	// Skip if no test database available
	t.Skip("Requires PostgreSQL test instance")

	cfg := &Config{
		Host:            "localhost",
		Port:            5432,
		User:            "wardflow",
		Password:        "wardflow_dev_password",
		DBName:          "wardflow",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		LogLevel:        "silent",
	}

	db, err := Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		t.Fatalf("failed to ping: %v", err)
	}
}

func TestHealthCheck(t *testing.T) {
	t.Skip("Requires PostgreSQL test instance")

	cfg := &Config{
		Host:            "localhost",
		Port:            5432,
		User:            "wardflow",
		Password:        "wardflow_dev_password",
		DBName:          "wardflow",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		LogLevel:        "silent",
	}

	db, err := Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	health, err := db.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}

	if health["status"] != "healthy" {
		t.Errorf("expected healthy status, got %v", health["status"])
	}

	if _, ok := health["open_connections"]; !ok {
		t.Error("expected open_connections in health check")
	}
}

func TestTransaction(t *testing.T) {
	t.Skip("Requires PostgreSQL test instance and schema")

	cfg := &Config{
		Host:            "localhost",
		Port:            5432,
		User:            "wardflow",
		Password:        "wardflow_dev_password",
		DBName:          "wardflow",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		LogLevel:        "silent",
	}

	db, err := Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test successful transaction
	err = db.Transaction(ctx, func(tx *DB) error {
		// Transaction logic here
		return nil
	})
	if err != nil {
		t.Errorf("transaction should succeed: %v", err)
	}
}
