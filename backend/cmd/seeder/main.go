package main

import (
	"context"
	"log"

	"github.com/RianAsmara/cargo-tracker/internal/config"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, cfg.DBUrl)
	if err != nil {
		log.Fatalf("Seeder failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("Seeder database ping failed: %v", err)
	}

	log.Println("Seeder running: connection established")

	// 1. Seed Tenant
	tenantID := uuid.New()
	tenantName := "Demo Express"
	tenantDomain := "demo.cargotracker.local"
	webhook := "http://localhost:8080/webhook" // Example mock

	// Upsert Tenant
	_, err = dbPool.Exec(ctx, `
		INSERT INTO tenants (id, name, domain, webhook_url, delay_threshold_hours)
		VALUES ($1, $2, $3, $4, 24)
		ON CONFLICT (domain) DO UPDATE SET name = EXCLUDED.name, webhook_url = EXCLUDED.webhook_url
	`, tenantID, tenantName, tenantDomain, webhook)

	if err != nil {
		log.Fatalf("Failed to seed tenant: %v", err)
	}

	// Fetch real Tenant ID (in case of conflict update)
	err = dbPool.QueryRow(ctx, "SELECT id FROM tenants WHERE domain = $1", tenantDomain).Scan(&tenantID)
	if err != nil {
		log.Fatalf("Failed to retrieve injected tenant ID: %v", err)
	}
	log.Printf("Tenant seeded successfully: %v (ID: %s)", tenantDomain, tenantID)

	// 2. Seed User
	email := "admin@demo.cargotracker.local"
	rawPassword := "password"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Upsert User
	userID := uuid.New()
	_, err = dbPool.Exec(ctx, `
		INSERT INTO users (id, tenant_id, email, password_hash, role)
		VALUES ($1, $2, $3, $4, 'operator')
		ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, role = EXCLUDED.role
	`, userID, tenantID, email, string(hashedPassword))

	if err != nil {
		log.Fatalf("Failed to seed user %v: %v", email, err)
	}

	log.Printf("User seeded successfully: %v (Password: %v)", email, rawPassword)
	log.Println("Seeding completed successfully!")
}
