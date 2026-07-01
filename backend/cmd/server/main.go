package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/contrib/websocket"
	"optistock/internal/auth"
	"optistock/internal/config"
	"optistock/internal/database"
	"optistock/internal/domain/admin"
	"optistock/internal/domain/alert"
	"optistock/internal/domain/approval"
	"optistock/internal/domain/item"
	"optistock/internal/domain/ledger"
	"optistock/internal/domain/lot"
	"optistock/internal/domain/receiving"
	"optistock/internal/domain/report"
	"optistock/internal/domain/stock"
	"optistock/internal/domain/workorder"
	"optistock/internal/middleware"
	"optistock/internal/ws"
	"optistock/pkg/response"
	"optistock/pkg/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := database.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	tokenManager, err := auth.NewTokenManager(cfg.JWTPrivateKeyPath, cfg.JWTPublicKeyPath, cfg.AccessTokenTTL)
	if err != nil {
		log.Fatalf("init token manager: %v", err)
	}

	hub := ws.NewHub()
	alertRepo := alert.NewRepository(pool)
	alertService := alert.NewService(alertRepo, hub)
	alertService.StartNearExpiryScheduler(ctx)

	authRepo := auth.NewRepository(pool)
	authService := auth.NewService(authRepo, tokenManager, cfg)
	authHandler := auth.NewHandler(authService, cfg.AppEnv)

	itemRepo := item.NewRepository(pool)
	itemService := item.NewService(itemRepo)
	itemHandler := item.NewHandler(itemService)

	lotRepo := lot.NewRepository(pool)
	lotService := lot.NewService(lotRepo, alertService)
	lotHandler := lot.NewHandler(lotService)

	ledgerRepo := ledger.NewRepository(pool)
	ledgerService := ledger.NewService(ledgerRepo)
	ledgerHandler := ledger.NewHandler(ledgerService)

	receivingRepo := receiving.NewRepository(pool)
	receivingService := receiving.NewService(receivingRepo, alertService)
	receivingHandler := receiving.NewHandler(receivingService)

	stockRepo := stock.NewRepository(pool)
	stockService := stock.NewService(stockRepo)
	stockHandler := stock.NewHandler(stockService)

	woRepo := workorder.NewRepository(pool)
	approvalRepo := approval.NewRepository(pool)
	woService := workorder.NewService(woRepo, approvalRepo, alertService)
	woHandler := workorder.NewHandler(woService, alertService)

	approvalService := approval.NewService(approvalRepo, woRepo, woService, alertService)
	approvalHandler := approval.NewHandler(approvalService)

	reportRepo := report.NewRepository(pool)
	reportUploader := storage.NewFromEnv()
	reportService := report.NewService(reportRepo, reportUploader)
	reportHandler := report.NewHandler(reportService)

	adminRepo := admin.NewRepository(pool)
	adminService := admin.NewService(adminRepo)
	adminHandler := admin.NewHandler(adminService)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return response.Fail(c, err)
		},
	})
	middleware.Register(app, cfg.CORSOrigins)

	api := app.Group("/api/v1")
	api.Get("/health", healthHandler(pool, cfg.AppVersion))

	authRoutes := api.Group("/auth")
	authHandler.RegisterRoutes(authRoutes)

	api.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	api.Get("/ws", ws.Handler(tokenManager, hub))

	protected := api.Group("", middleware.JWTAuth(tokenManager))
	protected.Get("/me", authHandler.Me)

	itemsRoutes := protected.Group("/items")
	itemHandler.RegisterItemRoutes(itemsRoutes)
	bomRoutes := protected.Group("/bom")
	itemHandler.RegisterBOMRoutes(bomRoutes)

	lotsRoutes := protected.Group("/lots")
	lotsRoutes.Patch("/:lotInternalID/status", middleware.RequireRole(auth.RoleSupervisor, auth.RoleAdmin), lotHandler.UpdateLotStatus)
	lotsRoutes.Get("/", lotHandler.ListLots)
	lotsRoutes.Post("/", lotHandler.CreateLot)
	lotsRoutes.Get("/:lotInternalID", lotHandler.GetLot)

	ledgerRoutes := protected.Group("/ledger")
	ledgerHandler.RegisterRoutes(ledgerRoutes)

	receivingRoutes := protected.Group("/receiving")
	receivingHandler.RegisterRoutes(receivingRoutes)

	stockRoutes := protected.Group("/stock")
	stockHandler.RegisterRoutes(stockRoutes)

	woRoutes := protected.Group("/workorders")
	woRoutes.Get("/:woNumber/approvals", approvalHandler.ListByWO)
	woHandler.RegisterRoutes(woRoutes)

	approvalRoutes := protected.Group("/approvals", middleware.RequireRole(auth.RoleSupervisor, auth.RoleAdmin))
	approvalRoutes.Get("/", approvalHandler.ListPending)
	approvalRoutes.Post("/:approvalID/approve", approvalHandler.Approve)
	approvalRoutes.Post("/:approvalID/reject", approvalHandler.Reject)

	protected.Post("/approvals/:approvalID/withdraw", approvalHandler.Withdraw)

	reportRoutes := protected.Group("/reports")
	reportHandler.RegisterRoutes(reportRoutes)

	adminRoutes := protected.Group("/admin", middleware.RequireRole(auth.RoleAdmin))
	adminHandler.RegisterRoutes(adminRoutes)

	go func() {
		log.Printf("api listening on :%s", cfg.HTTPPort)
		if err := app.Listen(":" + cfg.HTTPPort); err != nil {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func healthHandler(pool interface{ Ping(context.Context) error }, version string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := pool.Ping(c.Context()); err != nil {
			return response.Fail(c, err)
		}
		return response.OK(c, fiber.Map{
			"status":  "ok",
			"service": "optistock-api",
			"version": version,
		})
	}
}
