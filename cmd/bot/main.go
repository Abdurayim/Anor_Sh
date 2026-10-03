package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	_ "time/tzdata" // embed the time zone database: Asia/Tashkent must load even without the OS tzdata package

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"parent-bot/internal/config"
	"parent-bot/internal/database"
	"parent-bot/internal/handlers"
	"parent-bot/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	err = database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	log.Println("✓ Connected to database")

	// Create temp_docs directory for document generation
	if err := os.MkdirAll("./temp_docs", 0755); err != nil {
		log.Fatalf("Failed to create temp_docs directory: %v", err)
	}
	log.Println("✓ Temporary documents directory created")

	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("✓ Database schema is up to date")

	// Initialize bot service
	botService, err := services.NewBotService(cfg, database.DB)
	if err != nil {
		log.Fatalf("Failed to create bot service: %v", err)
	}

	log.Printf("✓ Bot authorized: @%s", botService.Bot.Self.UserName)

	// Initialize admins
	err = botService.InitializeAdmins()
	if err != nil {
		log.Printf("Warning: Failed to initialize admins: %v", err)
	} else {
		log.Println("✓ Admins initialized")
	}

	// Determine mode: webhook or polling
	useWebhook := cfg.Bot.WebhookURL != ""

	if useWebhook {
		// WEBHOOK MODE (Production)
		log.Println("🌐 Starting in WEBHOOK mode")
		startWebhookMode(cfg, botService)
	} else {
		// POLLING MODE (Development/Testing)
		log.Println("🔄 Starting in POLLING mode (for local testing)")
		startPollingMode(botService)
	}
}

// startWebhookMode starts the bot with webhook (for production)
func startWebhookMode(cfg *config.Config, botService *services.BotService) {
	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Create Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		err := database.HealthCheck()
		if err != nil {
			c.JSON(500, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Webhook endpoint
	router.POST("/webhook", func(c *gin.Context) {
		if cfg.Bot.WebhookSecret != "" && c.GetHeader("X-Telegram-Bot-Api-Secret-Token") != cfg.Bot.WebhookSecret {
			c.JSON(403, gin.H{"error": "forbidden"})
			return
		}

		var update tgbotapi.Update

		if err := c.BindJSON(&update); err != nil {
			log.Printf("Error binding update: %v", err)
			c.JSON(400, gin.H{"error": "invalid update"})
			return
		}

		// Handle update in goroutine to not block webhook response
		go handlers.SafeHandleUpdate(botService, update)

		c.JSON(200, gin.H{"ok": true})
	})

	// Admin API endpoints
	api := router.Group("/api", func(c *gin.Context) {
		// Admin API is disabled unless API_TOKEN is set; the token must be sent as a Bearer header.
		if cfg.Server.APIToken == "" || c.GetHeader("Authorization") != "Bearer "+cfg.Server.APIToken {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	})
	{
		admin := api.Group("/admin")
		{
			// Optional ?branch=olmazor|sergeli|<id>; without it data of all branches is returned.
			admin.GET("/users", func(c *gin.Context) {
				users, err := botService.UserService.GetAllUsers(apiBranchID(c, botService), 100, 0)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, gin.H{"users": users})
			})

			admin.GET("/complaints", func(c *gin.Context) {
				complaints, err := botService.ComplaintService.GetAllComplaintsWithUser(apiBranchID(c, botService), 100, 0)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, gin.H{"complaints": complaints})
			})

			admin.GET("/stats", func(c *gin.Context) {
				branchID := apiBranchID(c, botService)
				userCount, _ := botService.UserService.CountUsers(branchID)
				complaintCount, _ := botService.ComplaintService.CountComplaints(branchID)
				pendingCount, _ := botService.ComplaintService.CountComplaintsByStatus(branchID, "pending")

				c.JSON(200, gin.H{
					"total_users":        userCount,
					"total_complaints":   complaintCount,
					"pending_complaints": pendingCount,
				})
			})
		}
	}

	// Setup webhook
	webhookURL := cfg.Bot.WebhookURL + "/webhook"
	err := botService.SetWebhook(webhookURL, cfg.Bot.WebhookSecret)
	if err != nil {
		log.Printf("Warning: Failed to set webhook: %v", err)
	} else {
		log.Printf("✓ Webhook set to: %s", webhookURL)
	}

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("🚀 Server starting on %s", serverAddr)
	log.Printf("📱 Bot is ready to receive messages via webhook!")

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// startPollingMode starts the bot with polling (for development/testing)
func startPollingMode(botService *services.BotService) {
	// Remove webhook if set
	err := botService.RemoveWebhook()
	if err != nil {
		log.Printf("Warning: Failed to remove webhook: %v", err)
	}

	log.Println("✓ Webhook removed (using polling)")

	// Configure updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := botService.Bot.GetUpdatesChan(u)

	log.Println("📱 Bot is ready to receive messages via polling!")
	log.Println("💡 Press Ctrl+C to stop")
	log.Println(strings.Repeat("─", 50))

	// Process updates
	for update := range updates {
		// Handle each update
		handlers.SafeHandleUpdate(botService, update)
	}
}

// apiBranchID resolves the ?branch= query parameter (code or ID) to a branch ID; 0 = all branches.
func apiBranchID(c *gin.Context, botService *services.BotService) int {
	value := c.Query("branch")
	if value == "" {
		return 0
	}
	if id, err := strconv.Atoi(value); err == nil {
		return id
	}
	if branch, err := botService.BranchRepo.GetByCode(value); err == nil && branch != nil {
		return branch.ID
	}
	return -1
}
