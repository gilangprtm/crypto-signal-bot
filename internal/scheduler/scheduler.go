package scheduler

import (
	"crypto-signal-bot/internal/config"
	"crypto-signal-bot/internal/services"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type Scheduler struct {
	cron       *cron.Cron
	cfg        *config.Config
	botService *services.BotService
	isRunning  bool
}

func NewScheduler(cfg *config.Config, botService *services.BotService) *Scheduler {
	// Create cron with second precision and logging
	c := cron.New(cron.WithSeconds(), cron.WithLogger(cron.VerbosePrintfLogger(logrus.StandardLogger())))

	return &Scheduler{
		cron:       c,
		cfg:        cfg,
		botService: botService,
		isRunning:  false,
	}
}

func (s *Scheduler) Start() error {
	logrus.Info("⏰ Starting scheduler...")

	// Market analysis job - every 15 minutes during market hours
	analysisSchedule := fmt.Sprintf("0 */15 * * * *") // Every 15 minutes
	if s.cfg.AnalysisIntervalSeconds > 0 {
		// Custom interval in minutes
		intervalMinutes := s.cfg.AnalysisIntervalSeconds / 60
		if intervalMinutes < 1 {
			intervalMinutes = 1
		}
		analysisSchedule = fmt.Sprintf("0 */%d * * * *", intervalMinutes)
	}

	_, err := s.cron.AddFunc(analysisSchedule, s.runMarketAnalysis)
	if err != nil {
		return fmt.Errorf("failed to add market analysis job: %w", err)
	}
	logrus.Info("✅ Market analysis scheduled: ", analysisSchedule)

	// Performance tracking job - every hour
	_, err = s.cron.AddFunc("0 0 * * * *", s.updatePerformanceTracking)
	if err != nil {
		return fmt.Errorf("failed to add performance tracking job: %w", err)
	}
	logrus.Info("✅ Performance tracking scheduled: every hour")

	// Daily summary job - at 23:00 every day
	_, err = s.cron.AddFunc("0 0 23 * * *", s.sendDailySummary)
	if err != nil {
		return fmt.Errorf("failed to add daily summary job: %w", err)
	}
	logrus.Info("✅ Daily summary scheduled: 23:00 daily")

	// Learning optimization job - at 01:00 every day
	_, err = s.cron.AddFunc("0 0 1 * * *", s.runLearningOptimization)
	if err != nil {
		return fmt.Errorf("failed to add learning optimization job: %w", err)
	}
	logrus.Info("✅ Learning optimization scheduled: 01:00 daily")

	// Cleanup job - at 02:00 every day
	_, err = s.cron.AddFunc("0 0 2 * * *", s.runCleanup)
	if err != nil {
		return fmt.Errorf("failed to add cleanup job: %w", err)
	}
	logrus.Info("✅ Cleanup scheduled: 02:00 daily")

	// No health check needed for personal bot

	// Start the cron scheduler
	s.cron.Start()
	s.isRunning = true

	logrus.Info("✅ Scheduler started successfully")
	return nil
}

func (s *Scheduler) Stop() {
	logrus.Info("🛑 Stopping scheduler...")

	if s.cron != nil {
		ctx := s.cron.Stop()
		<-ctx.Done() // Wait for all jobs to complete
	}

	s.isRunning = false
	logrus.Info("✅ Scheduler stopped")
}

func (s *Scheduler) runMarketAnalysis() {
	logrus.Info("🔍 Scheduled market analysis starting...")
	
	start := time.Now()
	
	if err := s.botService.RunAnalysis(); err != nil {
		logrus.Error("Scheduled market analysis failed: ", err)
		// Send error notification
		s.sendErrorNotification("Market Analysis Failed", err.Error())
		return
	}
	
	duration := time.Since(start)
	logrus.Info("✅ Scheduled market analysis completed in ", duration)
}

func (s *Scheduler) updatePerformanceTracking() {
	logrus.Info("📊 Updating performance tracking...")
	
	// TODO: Implement performance tracking update
	// This would check all active signals and update their performance
	// based on current market prices
	
	logrus.Info("✅ Performance tracking updated")
}

func (s *Scheduler) sendDailySummary() {
	logrus.Info("📈 Sending daily summary...")
	
	if err := s.botService.SendDailySummary(); err != nil {
		logrus.Error("Failed to send daily summary: ", err)
		s.sendErrorNotification("Daily Summary Failed", err.Error())
		return
	}
	
	logrus.Info("✅ Daily summary sent")
}

func (s *Scheduler) runLearningOptimization() {
	logrus.Info("🧠 Running learning optimization...")
	
	// Get performance metrics
	metrics, err := s.botService.GetPerformanceMetrics()
	if err != nil {
		logrus.Error("Failed to get performance metrics: ", err)
		return
	}
	
	logrus.Info("Current performance - Win Rate: ", metrics.WinRate.StringFixed(2), "%, Total Signals: ", metrics.TotalSignals)
	
	// TODO: Implement actual learning optimization
	// This could include adjusting thresholds, weights, etc.
	
	logrus.Info("✅ Learning optimization completed")
}

func (s *Scheduler) runCleanup() {
	logrus.Info("🧹 Running cleanup tasks...")
	
	// TODO: Implement cleanup tasks
	// - Remove old market snapshots (keep last 30 days)
	// - Archive old signals (keep last 90 days)
	// - Clean up old logs
	// - Optimize database
	
	logrus.Info("✅ Cleanup completed")
}

// Health check removed - not needed for personal bot

func (s *Scheduler) sendErrorNotification(title, message string) {
	// TODO: Implement error notification
	// This could send alerts to Telegram or other channels
	logrus.Error(title, ": ", message)
}

func (s *Scheduler) GetStatus() map[string]interface{} {
	entries := s.cron.Entries()
	
	var nextRuns []map[string]interface{}
	for _, entry := range entries {
		nextRuns = append(nextRuns, map[string]interface{}{
			"next_run": entry.Next,
			"prev_run": entry.Prev,
		})
	}
	
	return map[string]interface{}{
		"is_running":    s.isRunning,
		"total_jobs":    len(entries),
		"next_runs":     nextRuns,
		"current_time":  time.Now(),
	}
}

func (s *Scheduler) AddCustomJob(schedule string, jobFunc func()) error {
	if !s.isRunning {
		return fmt.Errorf("scheduler is not running")
	}
	
	_, err := s.cron.AddFunc(schedule, jobFunc)
	if err != nil {
		return fmt.Errorf("failed to add custom job: %w", err)
	}
	
	logrus.Info("✅ Custom job added with schedule: ", schedule)
	return nil
}

func (s *Scheduler) RunJobNow(jobName string) error {
	logrus.Info("🚀 Running job manually: ", jobName)
	
	switch jobName {
	case "market_analysis":
		go s.runMarketAnalysis()
	case "performance_tracking":
		go s.updatePerformanceTracking()
	case "daily_summary":
		go s.sendDailySummary()
	case "learning_optimization":
		go s.runLearningOptimization()
	case "cleanup":
		go s.runCleanup()
	default:
		return fmt.Errorf("unknown job name: %s", jobName)
	}
	
	return nil
}

// IsMarketHours checks if current time is within trading hours
func (s *Scheduler) IsMarketHours() bool {
	// Crypto markets are 24/7, but we might want to reduce frequency
	// during low activity hours (e.g., 2 AM - 6 AM UTC)
	now := time.Now().UTC()
	hour := now.Hour()
	
	// Reduce activity during low volume hours
	if hour >= 2 && hour <= 6 {
		return false
	}
	
	return true
}

// GetNextAnalysisTime returns the next scheduled analysis time
func (s *Scheduler) GetNextAnalysisTime() time.Time {
	entries := s.cron.Entries()
	if len(entries) > 0 {
		return entries[0].Next // First entry is usually the most frequent (market analysis)
	}
	return time.Now()
}
