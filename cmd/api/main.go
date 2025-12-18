package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/arabella/ai-studio-backend/config"
	"github.com/arabella/ai-studio-backend/internal/infrastructure/auth"
	"github.com/arabella/ai-studio-backend/internal/infrastructure/cache"
	"github.com/arabella/ai-studio-backend/internal/infrastructure/database"
	"github.com/arabella/ai-studio-backend/internal/infrastructure/provider"
	"github.com/arabella/ai-studio-backend/internal/infrastructure/queue"
	infraRepo "github.com/arabella/ai-studio-backend/internal/infrastructure/repository"
	"github.com/arabella/ai-studio-backend/internal/infrastructure/worker"
	"github.com/arabella/ai-studio-backend/internal/interface/http/handler"
	"github.com/arabella/ai-studio-backend/internal/interface/http/middleware"
	"github.com/arabella/ai-studio-backend/internal/interface/websocket"
	"github.com/arabella/ai-studio-backend/internal/usecase"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	_ "github.com/arabella/ai-studio-backend/docs"
)

// Version and BuildTime are set during build
var (
	Version   = "dev"
	BuildTime = "unknown"
)

// @title Arabella API
// @version 1.0
// @description AI Video Generation Platform API
// @termsOfService https://arabella.app/terms

// @contact.name API Support
// @contact.url https://arabella.app/support
// @contact.email support@arabella.app

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host api.arabella.uz
// @BasePath /api/v1
// @schemes https http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := initLogger(cfg)
	defer logger.Sync()

	logger.Info("Starting Arabella API",
		zap.String("version", Version),
		zap.String("build_time", BuildTime),
		zap.String("environment", string(cfg.App.Environment)),
	)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database
	dbConfig := database.PostgresConfig{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.Database,
		SSLMode:         cfg.Database.SSLMode,
		MaxConnections:  cfg.Database.MaxConnections,
		MinConnections:  cfg.Database.MinConnections,
		MaxConnLifetime: cfg.Database.MaxConnLifetime,
		MaxConnIdleTime: cfg.Database.MaxConnIdleTime,
	}

	db, err := database.NewPostgresDB(ctx, dbConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Redis
	redisConfig := cache.RedisConfig{
		Host:         cfg.Redis.Host,
		Port:         cfg.Redis.Port,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	}

	redisCache, err := cache.NewRedisCache(ctx, redisConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisCache.Close()

	// Initialize repositories
	userRepo := infraRepo.NewUserRepositoryPostgres(db.Pool())
	templateRepo := infraRepo.NewTemplateRepositoryPostgres(db.Pool())
	videoJobRepo := infraRepo.NewVideoJobRepositoryPostgres(db.Pool())

	// Initialize rate limiter
	rateLimiter := cache.NewRateLimiter(redisCache.Client())

	// Initialize job queue
	jobQueue := queue.NewRedisQueue(redisCache.Client(), logger)

	// Initialize AI providers
	providerRegistry := provider.NewProviderRegistry(logger)

	if cfg.AI.UseMockProvider {
		mockProvider := provider.NewMockProvider(logger, false)
		providerRegistry.Register(mockProvider)
	}

	if cfg.AI.GeminiAPIKey != "" {
		geminiProvider := provider.NewGeminiProvider(cfg.AI.GeminiAPIKey, logger)
		providerRegistry.Register(geminiProvider)
	}

	if cfg.AI.WanAIAPIKey != "" {
		wanaiProvider := provider.NewWanAIProvider(cfg.AI.WanAIAPIKey, cfg.AI.WanAIVersion, cfg.AI.WanAIBaseURL, cfg.Server.BaseURL, logger)
		providerRegistry.Register(wanaiProvider)
		logger.Info("Wan AI provider registered",
			zap.String("version", cfg.AI.WanAIVersion),
			zap.String("base_url", cfg.AI.WanAIBaseURL),
		)
	}

	providerSelector := provider.NewProviderSelector(providerRegistry, logger)

	// Initialize WebSocket hub
	wsHub := websocket.NewHub(logger)
	go wsHub.Run()

	// Initialize auth components
	jwtConfig := auth.JWTConfig{
		SecretKey:            cfg.JWT.SecretKey,
		AccessTokenDuration:  cfg.JWT.AccessTokenDuration,
		RefreshTokenDuration: cfg.JWT.RefreshTokenDuration,
		Issuer:               cfg.JWT.Issuer,
	}
	tokenGenerator := auth.NewJWTTokenGenerator(jwtConfig)

	googleConfig := auth.GoogleAuthConfig{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
	}
	googleVerifier := auth.NewGoogleAuthVerifier(googleConfig, logger)

	// Initialize use cases
	authUseCase := usecase.NewAuthUseCase(userRepo, tokenGenerator, googleVerifier)
	templateUseCase := usecase.NewTemplateUseCase(templateRepo, redisCache)
	userUseCase := usecase.NewUserUseCase(userRepo, videoJobRepo)
	videoUseCase := usecase.NewVideoUseCase(
		videoJobRepo,
		templateRepo,
		userRepo,
		providerSelector,
		jobQueue,
		wsHub,
	)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUseCase)
	templateHandler := handler.NewTemplateHandler(templateUseCase)
	userHandler := handler.NewUserHandler(userUseCase)
	videoHandler := handler.NewVideoHandler(videoUseCase)
	uploadHandler := handler.NewUploadHandler()

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authUseCase)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(rateLimiter)
	loggingMiddleware := middleware.NewLoggingMiddleware(logger)

	// Initialize WebSocket handler
	wsHandler := websocket.NewHandler(wsHub, authUseCase, logger)

	// Initialize video worker to process queued jobs
	videoWorker := worker.NewVideoWorker(
		videoJobRepo,
		templateRepo,
		userRepo,
		providerSelector,
		jobQueue,
		wsHub,
		logger,
	)
	videoWorker.Start(ctx)
	logger.Info("Video worker started")

	// Setup router
	router := setupRouter(cfg, logger, authHandler, templateHandler, userHandler, videoHandler, uploadHandler,
		authMiddleware, rateLimitMiddleware, loggingMiddleware, wsHandler)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Server starting",
			zap.String("address", server.Addr),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped gracefully")
}

// initLogger initializes the zap logger
func initLogger(cfg *config.Config) *zap.Logger {
	var zapConfig zap.Config

	if cfg.IsDevelopment() {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	logger, err := zapConfig.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	return logger
}

// setupRouter configures the Gin router with all routes and middleware
func setupRouter(
	cfg *config.Config,
	logger *zap.Logger,
	authHandler *handler.AuthHandler,
	templateHandler *handler.TemplateHandler,
	userHandler *handler.UserHandler,
	videoHandler *handler.VideoHandler,
	uploadHandler *handler.UploadHandler,
	authMiddleware *middleware.AuthMiddleware,
	rateLimitMiddleware *middleware.RateLimitMiddleware,
	loggingMiddleware *middleware.LoggingMiddleware,
	wsHandler *websocket.Handler,
) *gin.Engine {
	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.RequestID())
	router.Use(loggingMiddleware.Logger())
	router.Use(loggingMiddleware.Recovery())

	// CORS configuration from config
	corsConfig := middleware.CORSConfig{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     cfg.CORS.AllowedMethods,
		AllowHeaders:     cfg.CORS.AllowedHeaders,
		ExposeHeaders:    []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "Retry-After"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
	router.Use(middleware.CORS(corsConfig))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"version": Version,
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Swagger documentation (development only)
	if cfg.IsDevelopment() {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Static file serving for thumbnails
	router.Static("/thumbnails", "./static/thumbnails")

	// Static file serving for cached temp images (for DashScope)
	router.Static("/temp-images", "./static/temp-images")

	// Static file serving for uploaded images (at root level for direct backend access)
	router.Static("/uploads", "./static/uploads")

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Static file serving for uploaded images under /api/v1/uploads (for Nginx proxy)
		v1.StaticFS("/uploads", http.Dir("./static/uploads"))

		// Image proxy endpoint (for external images that DashScope can't access)
		// Also used by frontend for nanobanana.uz images
		// Caches images locally to avoid repeated downloads
		v1.GET("/proxy/image", func(c *gin.Context) {
			imageURL := c.Query("url")
			if imageURL == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "url parameter is required"})
				return
			}

			// Validate URL
			parsedURL, err := url.Parse(imageURL)
			if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid URL"})
				return
			}

			// For nanobanana.uz images, check cache first
			if parsedURL.Host == "nanobanana.uz" {
				cacheDir := "./static/temp-images"
				os.MkdirAll(cacheDir, 0755)

				// Generate cache filename from URL hash
				hash := md5.Sum([]byte(imageURL))
				filename := hex.EncodeToString(hash[:]) + ".png"
				cachePath := filepath.Join(cacheDir, filename)

				// Check if cached and file is reasonable size (>100KB suggests complete image)
				if fileInfo, err := os.Stat(cachePath); err == nil && fileInfo.Size() > 100*1024 {
					// Serve from cache
					c.File(cachePath)
					return
				}
			}

			// For nanobanana.uz, always use HTTP (HTTPS has SSL issues)
			originalScheme := parsedURL.Scheme
			if parsedURL.Host == "nanobanana.uz" && parsedURL.Scheme == "https" {
				parsedURL.Scheme = "http"
				imageURL = parsedURL.String()
			}

			// Fetch the image with retry logic
			// Create HTTP client that ignores SSL certificate errors
			// This is needed for nanobanana.uz which has SSL issues
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // Accept any certificate
				},
				DisableKeepAlives:     false,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				ResponseHeaderTimeout: 30 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			}
			client := &http.Client{
				Timeout:   180 * time.Second, // Longer timeout for large images
				Transport: tr,
			}

			var resp *http.Response
			var fetchErr error
			maxFetchRetries := 3

			for i := 0; i < maxFetchRetries; i++ {
				req, err := http.NewRequest("GET", imageURL, nil)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create request", "details": err.Error()})
					return
				}

				// Set headers to mimic a browser request
				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
				req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
				req.Header.Set("Accept-Language", "en-US,en;q=0.9")
				req.Header.Set("Referer", "https://arabella.uz/")
				req.Header.Set("Cache-Control", "no-cache")
				req.Header.Set("Connection", "keep-alive")
				req.Header.Set("Accept-Encoding", "identity") // Disable compression to avoid issues

				// Try both HTTP and HTTPS if original is HTTP
				if parsedURL.Scheme == "http" {
					// Some servers redirect HTTP to HTTPS
					req.Header.Set("Upgrade-Insecure-Requests", "1")
				}

				resp, fetchErr = client.Do(req)

				if fetchErr != nil {
					// If HTTPS fails and we haven't tried HTTP yet, fallback to HTTP
					if originalScheme == "https" && parsedURL.Scheme == "https" && i == 0 {
						// Try HTTP as fallback
						parsedURL.Scheme = "http"
						imageURL = parsedURL.String()
						continue
					}
					// Continue to retry or fail after max retries
					if i < maxFetchRetries-1 {
						time.Sleep(time.Duration(i+1) * 2 * time.Second) // Exponential backoff
					}
					continue
				}

				if resp != nil {
					// Follow redirects
					if resp.StatusCode >= 300 && resp.StatusCode < 400 {
						location := resp.Header.Get("Location")
						if location != "" {
							resp.Body.Close()
							imageURL = location
							parsedURL, _ = url.Parse(imageURL)
							// For nanobanana.uz redirects, ensure HTTP
							if parsedURL.Host == "nanobanana.uz" && parsedURL.Scheme == "https" {
								parsedURL.Scheme = "http"
								imageURL = parsedURL.String()
							}
							continue // Retry with new URL
						}
					}
					if resp.StatusCode == http.StatusOK {
						break // Success!
					}
					// Non-200 status, close and retry
					resp.Body.Close()
					if i < maxFetchRetries-1 {
						time.Sleep(time.Duration(i+1) * 2 * time.Second)
					}
				}
			}

			if fetchErr != nil || resp == nil {
				// Log error but don't expose internal details to client
				c.JSON(http.StatusBadGateway, gin.H{
					"error":   "failed to fetch image",
					"details": "image source unavailable",
					"url":     imageURL,
				})
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				c.JSON(http.StatusBadGateway, gin.H{
					"error":  "failed to fetch image",
					"status": resp.StatusCode,
					"url":    imageURL,
				})
				return
			}

			// Read the entire image into memory to ensure we get the complete file
			// nanobanana.uz closes connections prematurely, so we need to read it all first
			// Set appropriate headers
			contentType := resp.Header.Get("Content-Type")
			if contentType == "" {
				contentType = "image/png" // Default content type
			}

			// Read the complete image with retry logic
			// nanobanana.uz closes connections prematurely, so we need to retry until we get the full image
			expectedSize := resp.ContentLength
			var imageData []byte
			maxRetries := 5

			for retry := 0; retry < maxRetries; retry++ {
				// Close previous response if retrying
				if retry > 0 {
					resp.Body.Close()
					// Re-fetch the image
					req, _ := http.NewRequest("GET", imageURL, nil)
					req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
					req.Header.Set("Accept", "image/*")
					req.Header.Set("Connection", "keep-alive")
					req.Header.Set("Accept-Encoding", "identity")

					newResp, err := client.Do(req)
					if err != nil || newResp == nil || newResp.StatusCode != http.StatusOK {
						if newResp != nil {
							newResp.Body.Close()
						}
						if retry < maxRetries-1 {
							time.Sleep(time.Duration(retry+1) * time.Second)
							continue
						}
						break
					}
					resp = newResp
				}

				// Read with timeout
				ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
				defer cancel()
				var buf bytes.Buffer
				maxSize := int64(50 * 1024 * 1024) // 50MB max
				limitedReader := io.LimitReader(resp.Body, maxSize)

				done := make(chan error, 1)
				go func() {
					_, err := io.Copy(&buf, limitedReader)
					done <- err
				}()

				var readErr error
				select {
				case <-ctx.Done():
					readErr = ctx.Err()
				case readErr = <-done:
				}
				cancel()

				imageData = buf.Bytes()

				// Check if we got the complete image
				if expectedSize > 0 {
					if int64(len(imageData)) >= expectedSize {
						// Got the full image!
						break
					}
				} else {
					// No Content-Length, check if read completed without error
					if readErr == nil {
						// Assume we got it all if no error
						break
					}
				}

				// If we didn't get the full image and there are retries left, try again
				if retry < maxRetries-1 {
					time.Sleep(time.Duration(retry+1) * 2 * time.Second)
				}
			}

			if len(imageData) == 0 {
				c.JSON(http.StatusBadGateway, gin.H{
					"error":   "failed to read image data",
					"details": "could not fetch complete image after retries",
				})
				return
			}

			// For nanobanana.uz images, cache the complete image locally
			if parsedURL != nil && parsedURL.Host == "nanobanana.uz" && len(imageData) > 0 {
				cacheDir := "./static/temp-images"
				os.MkdirAll(cacheDir, 0755)
				hash := md5.Sum([]byte(imageURL))
				filename := hex.EncodeToString(hash[:]) + ".png"
				cachePath := filepath.Join(cacheDir, filename)

				// Only cache if we got the complete image (or at least a reasonable amount)
				if expectedSize == 0 || int64(len(imageData)) >= expectedSize || len(imageData) > 100*1024 {
					if err := os.WriteFile(cachePath, imageData, 0644); err == nil {
						// Successfully cached, serve from cache next time
					}
				}
			}

			// Send the image with Content-Length matching what we actually have
			// This prevents ERR_CONTENT_LENGTH_MISMATCH
			c.Header("Content-Type", contentType)
			c.Header("Content-Length", fmt.Sprintf("%d", len(imageData)))
			c.Header("Cache-Control", "public, max-age=3600")
			c.Header("Access-Control-Allow-Origin", "*")

			// Send the image data
			c.Data(http.StatusOK, contentType, imageData)
		})

		// Rate limiting for all API routes
		v1.Use(rateLimitMiddleware.Limit(100, time.Minute))

		// Auth routes (public)
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/google", authHandler.GoogleAuth)
			authRoutes.POST("/test", authHandler.TestLogin) // Test login (skip Google)
			authRoutes.POST("/refresh", authHandler.RefreshToken)
			authRoutes.POST("/logout", authMiddleware.RequireAuth(), authHandler.Logout)
		}

		// Template routes (public, optional auth)
		templateRoutes := v1.Group("/templates")
		{
			templateRoutes.GET("", templateHandler.ListTemplates)
			templateRoutes.GET("/popular", templateHandler.GetPopularTemplates)
			templateRoutes.GET("/categories", templateHandler.GetCategories)
			templateRoutes.GET("/category/:category", templateHandler.GetTemplatesByCategory)
			templateRoutes.GET("/:id", templateHandler.GetTemplate)
		}

		// Admin routes (authenticated, admin only - TODO: add admin role check)
		adminRoutes := v1.Group("/admin")
		adminRoutes.Use(authMiddleware.RequireAuth())
		{
			// Admin template management
			adminTemplateRoutes := adminRoutes.Group("/templates")
			{
				adminTemplateRoutes.POST("", templateHandler.CreateTemplate)
				adminTemplateRoutes.PUT("/:id", templateHandler.UpdateTemplate)
				adminTemplateRoutes.DELETE("/:id", templateHandler.DeleteTemplate)
			}

			// Admin upload endpoints
			adminRoutes.POST("/upload/image", uploadHandler.UploadImage)
		}

		// Video routes (authenticated)
		videoRoutes := v1.Group("/videos")
		videoRoutes.Use(authMiddleware.RequireAuth())
		{
			videoRoutes.POST("/generate", rateLimitMiddleware.LimitGeneration(), videoHandler.GenerateVideo)
			videoRoutes.GET("", videoHandler.ListUserVideos)
			videoRoutes.GET("/recent", videoHandler.GetRecentVideos)
			videoRoutes.GET("/:id", videoHandler.GetVideo)
			videoRoutes.GET("/:id/status", videoHandler.GetJobStatus)
			videoRoutes.POST("/:id/cancel", videoHandler.CancelJob)
		}

		// User routes (authenticated)
		userRoutes := v1.Group("/user")
		userRoutes.Use(authMiddleware.RequireAuth())
		{
			userRoutes.GET("/profile", userHandler.GetProfile)
			userRoutes.PUT("/profile", userHandler.UpdateProfile)
			userRoutes.GET("/credits", userHandler.GetCredits)
			userRoutes.DELETE("/account", userHandler.DeleteAccount)
		}

		// Subscription routes (authenticated)
		subscriptionRoutes := v1.Group("/subscriptions")
		subscriptionRoutes.Use(authMiddleware.RequireAuth())
		{
			subscriptionRoutes.POST("", userHandler.UpgradeSubscription)
		}

		// WebSocket routes
		wsRoutes := v1.Group("/ws")
		{
			wsRoutes.GET("/videos/:id", wsHandler.HandleJobConnection)
		}
	}

	return router
}
