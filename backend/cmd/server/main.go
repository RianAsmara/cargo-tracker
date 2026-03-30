package main

import (
	"context"
	"log"

	"time"

	"github.com/RianAsmara/cargo-tracker/internal/config"
	"github.com/RianAsmara/cargo-tracker/internal/delivery/http/handlers"
	"github.com/RianAsmara/cargo-tracker/internal/delivery/http/middlewares"
	"github.com/RianAsmara/cargo-tracker/internal/infrastructure"
	"github.com/RianAsmara/cargo-tracker/internal/repository"
	"github.com/RianAsmara/cargo-tracker/internal/usecase"
	"github.com/RianAsmara/cargo-tracker/internal/worker"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg := config.Load()

	// Setup Database Connection Pool
	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, cfg.DBUrl)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatal("Database ping failed:", err)
	}

	log.Println("Database connection established")

	// Repositories
	tenantRepo := repository.NewPostgresTenantRepository(dbPool)
	userRepo := repository.NewPostgresUserRepository(dbPool)
	shipmentRepo := repository.NewPostgresShipmentRepository(dbPool)

	// Infrastructure
	webhookDispatcher := infrastructure.NewWebhookDispatcher()

	// UseCases
	shipmentUseCase := usecase.NewShipmentUseCase(shipmentRepo, tenantRepo, webhookDispatcher)

	// Handlers
	h := handlers.NewHandlers(shipmentUseCase, cfg.JWTSecret)

	// Echo App
	e := echo.New()
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())

	// ============================================
	// Delay Detector Background Worker
	// ============================================
	delayRepo := repository.NewPostgresDelayWorkerRepository(dbPool)
	delayWorker := worker.NewDelayDetector(delayRepo, 5*time.Minute)

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker() // Graceful cancellation upon app exit
	go delayWorker.Start(workerCtx)

	// ============================================
	// Public Routes (Require Domain Middleware)
	// ============================================
	// GET /track/:resi (public tracking endpoint)
	e.GET("/track/:resi", h.GetTracking, middlewares.TenantFromHeader(tenantRepo))

	// ============================================
	// Auth Routes (Require Domain Middleware)
	// ============================================
	auth := e.Group("/auth")
	auth.Use(middlewares.TenantFromHeader(tenantRepo))
	// POST /auth/login
	auth.POST("/login", func(c echo.Context) error {
		return h.AdminLogin(c, userRepo)
	})

	// ============================================
	// Internal Admin Routes (Require JWT Auth)
	// ============================================
	admin := e.Group("")
	admin.Use(middlewares.RequireAdminAuth(cfg.JWTSecret))

	// Operations (Staff & Admin & Courier)
	opsGroup := admin.Group("/shipments")
	opsGroup.Use(middlewares.RequireRole("admin", "staff", "courier"))

	opsGroup.POST("", h.CreateShipment)
	opsGroup.GET("/:resi", h.GetShipment)
	opsGroup.POST("/:id/events", h.AddEvent)

	// Start Server
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
