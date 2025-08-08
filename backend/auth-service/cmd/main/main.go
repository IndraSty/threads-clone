package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"github.com/IndraSty/threads-clone/backend/auth-service/configs"
	"github.com/IndraSty/threads-clone/backend/auth-service/internal/database"
	"github.com/IndraSty/threads-clone/backend/auth-service/internal/handlers"
	"github.com/IndraSty/threads-clone/backend/auth-service/internal/middleware"
	"github.com/IndraSty/threads-clone/backend/auth-service/internal/models"
	"github.com/IndraSty/threads-clone/backend/auth-service/internal/services"
	"github.com/IndraSty/threads-clone/backend/auth-service/internal/utils"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := configs.LoadConfig()

	// Initialize database connection
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := database.NewUserRepository(db)

	// Initialize utilities
	jwtManager := utils.NewJWTManager(cfg)
	oauthManager := utils.NewOAuthManager(cfg)

	// Initialize services
	authService := services.NewAuthService(userRepo, jwtManager, oauthManager)
	userService := services.NewUserService(userRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			return c.Status(code).JSON(models.ErrorResponse(
				"SERVER_ERROR",
				err.Error(),
				nil,
			))
		},
		DisableStartupMessage: cfg.Server.Env == "production",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://localhost:3001",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "auth-service",
			"version": "1.0.0",
		})
	})

	// Setup routes
	setupRoutes(app, authHandler, userHandler, authMiddleware)

	// Start server
	port := ":" + cfg.Server.Port
	log.Printf("🚀 Auth service starting on port %s", cfg.Server.Port)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := app.Listen(port); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	<-c
	log.Println("🔥 Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Fatal("Failed to shutdown server:", err)
	}
	log.Println("✅ Server shutdown complete")
}

func setupRoutes(
	app *fiber.App,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	authMiddleware *middleware.AuthMiddleware,
) {
	// API v1 group
	api := app.Group("/api/v1")

	// Auth routes
	auth := api.Group("/auth")
	{
		// Traditional auth
		auth.Post("/register", authHandler.Register)
		auth.Post("/login", authHandler.Login)

		// OAuth routes
		auth.Get("/google", authHandler.GoogleLogin)
		auth.Get("/google/callback", authHandler.GoogleCallback)
		auth.Get("/facebook", authHandler.FacebookLogin)
		auth.Get("/facebook/callback", authHandler.FacebookCallback)

		// Token validation (for other services)
		auth.Post("/validate", authHandler.ValidateToken)
	}

	// User routes (protected)
	users := api.Group("/users")
	users.Use(authMiddleware.JWTMiddleware())
	{
		users.Get("/profile", userHandler.GetProfile)
		users.Put("/profile", userHandler.UpdateProfile)
	}

	// Public user routes
	publicUsers := api.Group("/users")
	{
		publicUsers.Get("/:username", userHandler.GetUserByUsername)
	}

	// Legacy routes (without /api/v1 prefix for backward compatibility)
	legacyAuth := app.Group("/auth")
	{
		legacyAuth.Post("/register", authHandler.Register)
		legacyAuth.Post("/login", authHandler.Login)
		legacyAuth.Get("/google", authHandler.GoogleLogin)
		legacyAuth.Get("/google/callback", authHandler.GoogleCallback)
		legacyAuth.Get("/facebook", authHandler.FacebookLogin)
		legacyAuth.Get("/facebook/callback", authHandler.FacebookCallback)
		legacyAuth.Post("/validate", authHandler.ValidateToken)
	}

	legacyUsers := app.Group("/users")
	legacyUsers.Use(authMiddleware.JWTMiddleware())
	{
		legacyUsers.Get("/profile", userHandler.GetProfile)
		legacyUsers.Put("/profile", userHandler.UpdateProfile)
	}

	legacyPublicUsers := app.Group("/users")
	{
		legacyPublicUsers.Get("/:username", userHandler.GetUserByUsername)
	}
}
