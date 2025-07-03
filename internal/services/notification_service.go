package services

import (
	"crypto-signal-bot/internal/config"
	"crypto-signal-bot/internal/models"
	"fmt"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type NotificationService struct {
	cfg         *config.Config
	telegramBot *tgbotapi.BotAPI
	botService  *BotService // Add reference to bot service for menu actions
}

func NewNotificationService(cfg *config.Config) *NotificationService {
	ns := &NotificationService{
		cfg: cfg,
	}

	// Initialize Telegram bot if token is provided
	if cfg.TelegramBotToken != "" {
		bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
		if err != nil {
			logrus.Error("Failed to initialize Telegram bot: ", err)
		} else {
			ns.telegramBot = bot
			logrus.Info("✅ Telegram bot initialized successfully")
		}
	}

	return ns
}

// SetBotService sets the bot service reference for menu actions
func (ns *NotificationService) SetBotService(botService *BotService) {
	ns.botService = botService
}

// StartTelegramBot starts the Telegram bot with command handlers
func (ns *NotificationService) StartTelegramBot() error {
	if ns.telegramBot == nil {
		return fmt.Errorf("telegram bot not initialized")
	}

	logrus.Info("🤖 Starting Telegram bot with interactive menu...")

	// Set up update configuration
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := ns.telegramBot.GetUpdatesChan(u)

	// Handle updates in a goroutine
	go func() {
		for update := range updates {
			if update.Message != nil {
				ns.handleMessage(update.Message)
			} else if update.CallbackQuery != nil {
				ns.handleCallbackQuery(update.CallbackQuery)
			}
		}
	}()

	logrus.Info("✅ Telegram bot started with interactive menu")
	return nil
}

// handleMessage handles incoming text messages and commands
func (ns *NotificationService) handleMessage(message *tgbotapi.Message) {
	if message.IsCommand() {
		ns.handleCommand(message)
	} else {
		// Handle regular text messages if needed
		ns.sendHelpMessage(message.Chat.ID)
	}
}

// handleCommand handles bot commands
func (ns *NotificationService) handleCommand(message *tgbotapi.Message) {
	command := message.Command()
	chatID := message.Chat.ID

	logrus.Infof("Received command: /%s from chat %d", command, chatID)

	switch command {
	case "start":
		ns.sendWelcomeMessage(chatID)
	case "menu":
		ns.sendMainMenu(chatID)
	case "status":
		ns.sendBotStatus(chatID)
	case "coins":
		ns.sendCoinsList(chatID)
	case "performance":
		ns.sendPerformanceReport(chatID)
	case "help":
		ns.sendHelpMessage(chatID)
	default:
		ns.sendUnknownCommandMessage(chatID)
	}
}

// handleCallbackQuery handles button presses
func (ns *NotificationService) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	chatID := callbackQuery.Message.Chat.ID
	data := callbackQuery.Data

	logrus.Infof("Received callback: %s from chat %d", data, chatID)

	// Acknowledge the callback query
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	ns.telegramBot.Request(callback)

	switch data {
	case "main_menu":
		ns.sendMainMenu(chatID)
	case "bot_status":
		ns.sendBotStatus(chatID)
	case "manual_analysis":
		ns.runManualAnalysis(chatID)
	case "coins_list":
		ns.sendCoinsList(chatID)
	case "add_coin":
		ns.sendAddCoinMenu(chatID)
	case "performance":
		ns.sendPerformanceReport(chatID)
	case "settings":
		ns.sendSettingsMenu(chatID)
	case "daily_summary":
		ns.sendDailySummaryNow(chatID)
	case "learning_stats":
		ns.sendLearningStats(chatID)
	default:
		if len(data) > 9 && data[:9] == "add_coin_" {
			symbol := data[9:]
			ns.addCoinToWatch(chatID, symbol)
		} else if len(data) > 12 && data[:12] == "remove_coin_" {
			symbol := data[12:]
			ns.removeCoinFromWatch(chatID, symbol)
		} else {
			ns.sendMainMenu(chatID)
		}
	}
}

func (ns *NotificationService) SendSignalNotification(signal *models.TradingSignal) error {
	logrus.Info("Sending signal notification for: ", signal.Crypto.Symbol)

	// Format message
	message := ns.formatSignalMessage(signal)

	// Send to Telegram
	if ns.telegramBot != nil && ns.cfg.TelegramChatID != "" {
		if err := ns.sendTelegramMessage(message); err != nil {
			logrus.Error("Failed to send Telegram message: ", err)
			return err
		}
	}

	// Send to WhatsApp (if enabled)
	if ns.cfg.WhatsAppEnabled {
		if err := ns.sendWhatsAppMessage(message); err != nil {
			logrus.Error("Failed to send WhatsApp message: ", err)
			// Don't return error for WhatsApp failure, continue with other notifications
		}
	}

	logrus.Info("✅ Signal notification sent successfully")
	return nil
}

func (ns *NotificationService) formatSignalMessage(signal *models.TradingSignal) string {
	// Get action emoji
	var actionEmoji string
	switch signal.Action {
	case "BUY":
		actionEmoji = "🟢"
	case "SELL":
		actionEmoji = "🔴"
	default:
		actionEmoji = "🟡"
	}

	// Format confidence as percentage
	confidence := signal.ConfidenceScore.Mul(decimal.NewFromInt(100))

	// Format prices
	entryPrice := signal.EntryPrice.StringFixed(8)
	stopLoss := ""
	takeProfit1 := ""
	takeProfit2 := ""

	if signal.StopLoss != nil {
		stopLoss = signal.StopLoss.StringFixed(8)
	}
	if signal.TakeProfit1 != nil {
		takeProfit1 = signal.TakeProfit1.StringFixed(8)
	}
	if signal.TakeProfit2 != nil {
		takeProfit2 = signal.TakeProfit2.StringFixed(8)
	}

	// Build message
	message := fmt.Sprintf(`🚨 *CRYPTO SIGNAL* 🚨

%s *%s/USDT*
📈 *Action:* %s
💵 *Entry Price:* $%s
🎯 *Confidence:* %.1f%%

📊 *Analysis:*`,
		actionEmoji,
		signal.Crypto.Symbol,
		signal.Action,
		entryPrice,
		confidence.InexactFloat64(),
	)

	// Add technical indicators
	if signal.RSI != nil {
		message += fmt.Sprintf("\n• RSI: %.2f", signal.RSI.InexactFloat64())
	}

	if signal.MACDHistogram != nil {
		macdStatus := "Bullish"
		if signal.MACDHistogram.LessThan(decimal.Zero) {
			macdStatus = "Bearish"
		}
		message += fmt.Sprintf("\n• MACD: %s", macdStatus)
	}

	if signal.FearGreedIndex != nil {
		fgiText := ns.getFearGreedText(*signal.FearGreedIndex)
		message += fmt.Sprintf("\n• Fear & Greed: %d (%s)", *signal.FearGreedIndex, fgiText)
	}

	// Add price targets
	if signal.Action != "HOLD" {
		message += "\n\n🎯 *Targets:*"
		if stopLoss != "" {
			message += fmt.Sprintf("\n• Stop Loss: $%s", stopLoss)
		}
		if takeProfit1 != "" {
			message += fmt.Sprintf("\n• Take Profit 1: $%s", takeProfit1)
		}
		if takeProfit2 != "" {
			message += fmt.Sprintf("\n• Take Profit 2: $%s", takeProfit2)
		}
	}

	// Add reasoning
	if signal.Reasoning != "" {
		message += fmt.Sprintf("\n\n💡 *Reasoning:*\n%s", signal.Reasoning)
	}

	// Add timestamp
	message += fmt.Sprintf("\n\n⏰ %s WIB", signal.CreatedAt.Format("15:04 02/01/2006"))

	// Add disclaimer
	message += "\n\n⚠️ *DYOR - Not Financial Advice*"

	return message
}

func (ns *NotificationService) sendTelegramMessage(message string) error {
	return ns.sendTelegramMessageToChat(ns.cfg.TelegramChatID, message)
}

func (ns *NotificationService) sendTelegramMessageToChat(chatIDStr string, message string) error {
	var msg tgbotapi.MessageConfig

	// Try to parse as numeric chat ID first
	if chatID, err := strconv.ParseInt(chatIDStr, 10, 64); err == nil {
		msg = tgbotapi.NewMessage(chatID, message)
	} else {
		// Use as username/channel (e.g., @username)
		msg = tgbotapi.NewMessageToChannel(chatIDStr, message)
	}

	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	_, err := ns.telegramBot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send Telegram message to %s: %w", chatIDStr, err)
	}

	logrus.Info("✅ Telegram message sent successfully to ", chatIDStr)
	return nil
}

func (ns *NotificationService) sendWhatsAppMessage(message string) error {
	// TODO: Implement WhatsApp Business API integration
	// For now, just log that WhatsApp is not implemented
	logrus.Info("WhatsApp notification would be sent: ", message[:50], "...")
	return nil
}

func (ns *NotificationService) getFearGreedText(index int) string {
	switch {
	case index <= 20:
		return "Extreme Fear"
	case index <= 40:
		return "Fear"
	case index <= 60:
		return "Neutral"
	case index <= 80:
		return "Greed"
	default:
		return "Extreme Greed"
	}
}

func (ns *NotificationService) SendSystemNotification(level, message string) error {
	if ns.telegramBot == nil || ns.cfg.TelegramChatID == "" {
		return nil
	}

	var emoji string
	switch level {
	case "error":
		emoji = "🚨"
	case "warning":
		emoji = "⚠️"
	case "info":
		emoji = "ℹ️"
	default:
		emoji = "📢"
	}

	systemMessage := fmt.Sprintf("%s *System %s*\n\n%s\n\n⏰ %s",
		emoji,
		level,
		message,
		time.Now().Format("15:04 02/01/2006"),
	)

	return ns.sendTelegramMessage(systemMessage)
}

func (ns *NotificationService) SendDailySummary(analytics []*models.SignalAnalytics) error {
	if len(analytics) == 0 {
		return nil
	}

	message := "📊 *Daily Signal Summary*\n\n"

	totalSignals := 0
	totalWinRate := decimal.Zero
	totalPnL := decimal.Zero

	for _, analytic := range analytics {
		if analytic.TotalSignals > 0 {
			message += fmt.Sprintf("*%s:* %d signals, %.1f%% win rate, %.2f%% avg PnL\n",
				analytic.Symbol,
				analytic.TotalSignals,
				analytic.WinRatePercentage.InexactFloat64(),
				analytic.AvgPnLPercentage.InexactFloat64(),
			)
			totalSignals += analytic.TotalSignals
			totalWinRate = totalWinRate.Add(analytic.WinRatePercentage)
			totalPnL = totalPnL.Add(analytic.AvgPnLPercentage)
		}
	}

	if totalSignals > 0 {
		avgWinRate := totalWinRate.Div(decimal.NewFromInt(int64(len(analytics))))
		avgPnL := totalPnL.Div(decimal.NewFromInt(int64(len(analytics))))

		message += fmt.Sprintf("\n*Overall:* %d signals, %.1f%% avg win rate, %.2f%% avg PnL",
			totalSignals,
			avgWinRate.InexactFloat64(),
			avgPnL.InexactFloat64(),
		)
	}

	message += fmt.Sprintf("\n\n⏰ %s", time.Now().Format("15:04 02/01/2006"))

	return ns.sendTelegramMessage(message)
}

func (ns *NotificationService) SendPerformanceUpdate(signal *models.TradingSignal, performance *models.SignalPerformance) error {
	if performance.Outcome == "pending" {
		return nil // Don't send updates for pending signals
	}

	var emoji string
	switch performance.Outcome {
	case "profit":
		emoji = "✅"
	case "loss":
		emoji = "❌"
	default:
		emoji = "⚖️"
	}

	pnlPercent := decimal.Zero
	if performance.PnLPercentage != nil {
		pnlPercent = *performance.PnLPercentage
	}

	message := fmt.Sprintf(`%s *Signal Update*

*%s/USDT* %s signal closed
💰 *PnL:* %.2f%%
⏱️ *Duration:* %d minutes
📈 *Entry:* $%s
📉 *Exit:* $%s

⏰ %s`,
		emoji,
		signal.Crypto.Symbol,
		signal.Action,
		pnlPercent.InexactFloat64(),
		*performance.DurationMinutes,
		signal.EntryPrice.StringFixed(8),
		performance.ExitPrice.StringFixed(8),
		time.Now().Format("15:04 02/01/2006"),
	)

	return ns.sendTelegramMessage(message)
}

func (ns *NotificationService) TestConnection() error {
	if ns.telegramBot == nil {
		return fmt.Errorf("telegram bot not initialized")
	}

	testMessage := "🤖 *Crypto Signal Bot Test*\n\nConnection successful!\n\n⏰ " + time.Now().Format("15:04 02/01/2006")
	return ns.sendTelegramMessage(testMessage)
}
