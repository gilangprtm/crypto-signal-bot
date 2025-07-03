package main

import (
	"crypto-signal-bot/internal/api"
	"crypto-signal-bot/internal/config"
	"crypto-signal-bot/internal/database"
	"crypto-signal-bot/internal/scheduler"
	"crypto-signal-bot/internal/services"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
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

	// Check for existing instance and create PID file
	if err := createPIDFile(); err != nil {
		logrus.Fatalf("Failed to create PID file: %v", err)
	}
	defer removePIDFile()

	logrus.Info("🚀 Starting Personal Crypto Signal Bot (Production Mode)...")

	// Initialize database with retry mechanism and graceful degradation
	var db *database.SupabaseClient
	var err error

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		db, err = database.NewSupabaseClient(cfg)
		if err != nil {
			logrus.Warnf("Database connection attempt %d/%d failed: %v", i+1, maxRetries, err)
			if i == maxRetries-1 {
				logrus.Warn("⚠️ Running in degraded mode without database")
				db = nil
				break
			}
			time.Sleep(time.Duration(i+1) * 2 * time.Second)
			continue
		}
		logrus.Info("✅ Database connected successfully")
		break
	}

	if db != nil {
		defer db.Close()
	}

	// Initialize services
	botService := services.NewBotService(db, cfg)

	// Initialize scheduler
	schedulerService := scheduler.NewScheduler(cfg, botService)

	// Initialize API server
	apiServer := api.NewServer(cfg, db, botService, schedulerService)

	// Start API server
	logrus.Info("🌐 Starting API server on port ", cfg.APIPort)
	go func() {
		if err := apiServer.Start(); err != nil {
			logrus.Error("API server error: ", err)
		}
	}()

	// Start scheduler
	logrus.Info("🔄 Starting scheduler...")
	go func() {
		if err := schedulerService.Start(); err != nil {
			logrus.Error("Scheduler error: ", err)
		}
	}()

	// Start bot service with retry mechanism
	go func() {
		maxBotRetries := 3
		for i := 0; i < maxBotRetries; i++ {
			if err := botService.Start(); err != nil {
				logrus.Errorf("Bot service start attempt %d/%d failed: %v", i+1, maxBotRetries, err)
				if i < maxBotRetries-1 {
					time.Sleep(time.Duration(i+1) * 5 * time.Second)
					continue
				}
				logrus.Error("❌ Failed to start bot service after all retries")
				return
			}
			logrus.Info("✅ Bot service started successfully")
			break
		}
	}()

	// Run initial market analysis in background
	go func() {
		time.Sleep(10 * time.Second) // Wait for services to start
		logrus.Info("📊 Running initial market analysis...")
		if err := botService.RunAnalysis(); err != nil {
			logrus.Error("Initial market analysis failed: ", err)
		}
	}()

	logrus.Info("✅ Personal Crypto Signal Bot is running in production mode!")
	logrus.Info("📱 Telegram bot is ready for commands")
	logrus.Info("📊 API available at: http://localhost:", cfg.APIPort, "/api/v1")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logrus.Info("🛑 Shutting down...")

	// Graceful shutdown
	// Stop scheduler
	schedulerService.Stop()

	// Stop bot service
	if err := botService.Stop(); err != nil {
		logrus.Error("Bot service shutdown error: ", err)
	}

	// Stop API server
	if err := apiServer.Stop(); err != nil {
		logrus.Error("API server shutdown error: ", err)
	}

	logrus.Info("👋 Goodbye!")
}

const pidFile = "/tmp/crypto-signal-bot.pid"

func createPIDFile() error {
	// Check if PID file exists
	if _, err := os.Stat(pidFile); err == nil {
		// Read existing PID
		data, err := ioutil.ReadFile(pidFile)
		if err == nil {
			if pid, err := strconv.Atoi(string(data)); err == nil {
				// Check if process is still running
				if process, err := os.FindProcess(pid); err == nil {
					if err := process.Signal(syscall.Signal(0)); err == nil {
						return fmt.Errorf("bot is already running with PID %d", pid)
					}
				}
			}
		}
		// Remove stale PID file
		os.Remove(pidFile)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(pidFile), 0755); err != nil {
		return fmt.Errorf("failed to create PID directory: %w", err)
	}

	// Write current PID
	pid := os.Getpid()
	return ioutil.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
}

func removePIDFile() {
	os.Remove(pidFile)
}

func setupLogging(level string) {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   false, // Disable colors for production logs
	})

	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Info("📝 Logging level set to: debug")
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
		logrus.Info("📝 Logging level set to: info")
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
		logrus.Info("📝 Logging level set to: warn")
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
		logrus.Info("📝 Logging level set to: error")
	default:
		logrus.SetLevel(logrus.InfoLevel)
		logrus.Info("📝 Logging level set to: info (default)")
	}
}
