package api

import (
	"crypto-signal-bot/internal/config"
	"crypto-signal-bot/internal/database"
	"crypto-signal-bot/internal/models"
	"crypto-signal-bot/internal/scheduler"
	"crypto-signal-bot/internal/services"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Server struct {
	cfg        *config.Config
	db         *database.SupabaseClient
	botService *services.BotService
	scheduler  *scheduler.Scheduler
	router     *mux.Router
	server     *http.Server
}

func NewServer(cfg *config.Config, db *database.SupabaseClient, botService *services.BotService, scheduler *scheduler.Scheduler) *Server {
	s := &Server{
		cfg:        cfg,
		db:         db,
		botService: botService,
		scheduler:  scheduler,
		router:     mux.NewRouter(),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Root endpoint
	s.router.HandleFunc("/", s.handleRoot).Methods("GET")

	// API prefix
	api := s.router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/", s.handleRoot).Methods("GET")

	// Bot status and control
	api.HandleFunc("/bot/status", s.handleBotStatus).Methods("GET")
	api.HandleFunc("/bot/start", s.handleBotStart).Methods("POST")
	api.HandleFunc("/bot/stop", s.handleBotStop).Methods("POST")

	// Manual operations
	api.HandleFunc("/bot/analyze", s.handleManualAnalysis).Methods("POST")
	api.HandleFunc("/bot/summary", s.handleDailySummary).Methods("POST")

	// Signals
	api.HandleFunc("/signals", s.handleGetSignals).Methods("GET")
	api.HandleFunc("/signals/{id}", s.handleGetSignal).Methods("GET")
	api.HandleFunc("/signals/analytics", s.handleSignalAnalytics).Methods("GET")

	// Performance
	api.HandleFunc("/performance/metrics", s.handlePerformanceMetrics).Methods("GET")
	api.HandleFunc("/performance/learning", s.handleLearningInsights).Methods("GET")

	// Scheduler
	api.HandleFunc("/scheduler/status", s.handleSchedulerStatus).Methods("GET")
	api.HandleFunc("/scheduler/jobs/{job}/run", s.handleRunJob).Methods("POST")

	// Market data
	api.HandleFunc("/market/{symbol}", s.handleGetMarketData).Methods("GET")
	api.HandleFunc("/cryptocurrencies", s.handleGetCryptocurrencies).Methods("GET")

	// Static files (for simple dashboard)
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/static/")))

	// Middleware
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.corsMiddleware)
}

func (s *Server) Start() error {
	port := s.cfg.APIPort
	if port == 0 {
		port = 8080
	}

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logrus.Info("🌐 Starting API server on port ", port)
	return s.server.ListenAndServe()
}

func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// Root endpoint
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "🤖 Crypto Signal Bot API",
		Data: map[string]interface{}{
			"version":   "1.0.0",
			"status":    "running",
			"endpoints": []string{"/api/v1/bot/status", "/api/v1/bot/start", "/api/v1/bot/stop"},
			"timestamp": time.Now().Format(time.RFC3339),
		},
	})
}



// Bot status endpoint
func (s *Server) handleBotStatus(w http.ResponseWriter, r *http.Request) {
	status := s.botService.GetStatus()
	s.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    status,
	})
}

// Start bot endpoint
func (s *Server) handleBotStart(w http.ResponseWriter, r *http.Request) {
	if err := s.botService.Start(); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Bot started successfully",
	})
}

// Stop bot endpoint
func (s *Server) handleBotStop(w http.ResponseWriter, r *http.Request) {
	if err := s.botService.Stop(); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Bot stopped successfully",
	})
}

// Manual analysis endpoint
func (s *Server) handleManualAnalysis(w http.ResponseWriter, r *http.Request) {
	if err := s.botService.RunAnalysis(); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Manual analysis completed",
	})
}

// Daily summary endpoint
func (s *Server) handleDailySummary(w http.ResponseWriter, r *http.Request) {
	if err := s.botService.SendDailySummary(); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Daily summary sent",
	})
}

// Get signals endpoint
func (s *Server) handleGetSignals(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	signals, err := s.db.GetRecentSignals(limit)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    signals,
	})
}

// Get single signal endpoint
func (s *Server) handleGetSignal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	signalID := vars["id"]

	signal, err := s.db.GetSignalByID(signalID)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "Signal not found",
		})
		return
	}

	s.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    signal,
	})
}

// Signal analytics endpoint
func (s *Server) handleSignalAnalytics(w http.ResponseWriter, r *http.Request) {
	analytics, err := s.db.GetSignalAnalytics()
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    analytics,
	})
}

// Performance metrics endpoint
func (s *Server) handlePerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := s.botService.GetPerformanceMetrics()
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    metrics,
	})
}

// Learning insights endpoint
func (s *Server) handleLearningInsights(w http.ResponseWriter, r *http.Request) {
	insights, err := s.db.GetLearningInsights()
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    insights,
	})
}

// Scheduler status endpoint
func (s *Server) handleSchedulerStatus(w http.ResponseWriter, r *http.Request) {
	status := s.scheduler.GetStatus()
	s.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    status,
	})
}

// Run job endpoint
func (s *Server) handleRunJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobName := vars["job"]

	if err := s.scheduler.RunJobNow(jobName); err != nil {
		s.writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("Job '%s' started", jobName),
	})
}

// Get market data endpoint
func (s *Server) handleGetMarketData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]

	// TODO: Implement market data retrieval
	s.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"symbol": symbol, "status": "not_implemented"},
	})
}

// Get cryptocurrencies endpoint
func (s *Server) handleGetCryptocurrencies(w http.ResponseWriter, r *http.Request) {
	cryptos, err := s.db.GetCryptocurrencies()
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    cryptos,
	})
}

// Helper methods
func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logrus.Debug(
			"API Request: ",
			r.Method, " ",
			r.RequestURI, " ",
			time.Since(start),
		)
	})
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
