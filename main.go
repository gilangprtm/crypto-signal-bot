package main

import (
	"crypto-signal-bot/internal/api"
	"crypto-signal-bot/internal/config"
	"crypto-signal-bot/internal/database"
	"crypto-signal-bot/internal/scheduler"
	"crypto-signal-bot/internal/services"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.Load()

	// Setup logging
	setupLogging(cfg.LogLevel)

	logrus.Info("🚀 Starting Personal Crypto Signal Bot...")

	// Initialize database connection
	db, err := database.NewSupabaseClient(cfg)
	if err != nil {
		logrus.Fatal("Failed to connect to Supabase: ", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logrus.Fatal("Failed to ping database: ", err)
	}
	logrus.Info("✅ Connected to Supabase successfully")

	// Initialize bot service (orchestrator)
	botService := services.NewBotService(db, cfg)

	// Initialize scheduler
	schedulerService := scheduler.NewScheduler(cfg, botService)

	// Initialize API server
	apiServer := api.NewServer(cfg, db, botService, schedulerService)

	// Start bot service
	if err := botService.Start(); err != nil {
		logrus.Fatal("Failed to start bot service: ", err)
	}

	// Start scheduler
	logrus.Info("🔄 Starting scheduler...")
	if err := schedulerService.Start(); err != nil {
		logrus.Fatal("Failed to start scheduler: ", err)
	}

	// Start API server
	logrus.Info("🌐 Starting API server on port ", cfg.APIPort)
	go func() {
		if err := apiServer.Start(); err != nil {
			logrus.Error("API server error: ", err)
		}
	}()

	// Initial market analysis
	logrus.Info("📊 Running initial market analysis...")
	go func() {
		time.Sleep(5 * time.Second) // Wait for services to initialize
		if err := botService.RunAnalysis(); err != nil {
			logrus.Error("Initial market analysis failed: ", err)
		}
	}()

	logrus.Info("✅ Personal Crypto Signal Bot is running!")
	logrus.Info("📱 Signals will be sent to your Telegram/WhatsApp")
	logrus.Info("📊 Analytics available at http://localhost:", cfg.APIPort, "/api/v1")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logrus.Info("🛑 Shutting down Personal Crypto Signal Bot...")

	// Graceful shutdown
	if err := botService.Stop(); err != nil {
		logrus.Error("Error stopping bot service: ", err)
	}

	schedulerService.Stop()

	if err := apiServer.Stop(); err != nil {
		logrus.Error("Error stopping API server: ", err)
	}

	logrus.Info("👋 Goodbye!")
}

func setupLogging(level string) {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}
