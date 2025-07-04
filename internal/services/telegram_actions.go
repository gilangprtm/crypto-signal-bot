package services

import (
	"crypto-signal-bot/internal/models"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// runManualAnalysis triggers manual market analysis
func (ns *NotificationService) runManualAnalysis(chatID int64) {
	if ns.botService == nil {
		ns.sendErrorMessage(chatID, "Bot service tidak tersedia")
		return
	}

	// Send "analyzing" message
	msg := tgbotapi.NewMessage(chatID, "🔍 *Memulai analisis manual...*\n\nMohon tunggu, sedang menganalisis market...")
	msg.ParseMode = "Markdown"
	ns.telegramBot.Send(msg)

	// Run analysis
	go func() {
		err := ns.botService.RunAnalysis()
		
		var resultMessage string
		if err != nil {
			resultMessage = fmt.Sprintf("🚨 *Analisis Gagal*\n\nError: %s", err.Error())
		} else {
			resultMessage = fmt.Sprintf(`✅ *Analisis Manual Selesai*

🕐 *Waktu:* %s
📊 *Coins Dianalisis:* %d
📈 *Sinyal Baru:* Cek notifikasi di atas

Analisis berikutnya akan berjalan otomatis sesuai jadwal.`,
				time.Now().Format("15:04 02/01/2006"),
				len(ns.botService.cryptoList),
			)
		}

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📊 Status Bot", "bot_status"),
				tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "main_menu"),
			),
		)

		msg := tgbotapi.NewMessage(chatID, resultMessage)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = keyboard
		ns.telegramBot.Send(msg)
	}()
}

// addCoinToWatch adds a new cryptocurrency to watchlist
func (ns *NotificationService) addCoinToWatch(chatID int64, symbol string) {
	if ns.botService == nil {
		ns.sendErrorMessage(chatID, "Bot service tidak tersedia")
		return
	}

	// Check if coin already exists
	for _, crypto := range ns.botService.cryptoList {
		if crypto.Symbol == symbol {
			message := fmt.Sprintf("⚠️ *%s sudah ada dalam watchlist*", symbol)
			
			keyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("💰 Lihat Coins", "coins_list"),
					tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "main_menu"),
				),
			)

			msg := tgbotapi.NewMessage(chatID, message)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = keyboard
			ns.telegramBot.Send(msg)
			return
		}
	}

	// Add new cryptocurrency
	newCrypto := &models.Cryptocurrency{
		ID:        uuid.New(),
		Symbol:    symbol,
		Name:      getCoinName(symbol),
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	// Add to database
	if err := ns.botService.db.CreateCryptocurrency(newCrypto); err != nil {
		ns.sendErrorMessage(chatID, fmt.Sprintf("Gagal menambahkan %s: %s", symbol, err.Error()))
		return
	}

	// Add to bot's crypto list
	ns.botService.cryptoList = append(ns.botService.cryptoList, newCrypto)

	message := fmt.Sprintf(`✅ *%s berhasil ditambahkan!*

🪙 *Coin:* %s (%s)
📊 *Status:* Aktif
🔍 *Analisis:* Akan dimulai pada siklus berikutnya

Bot sekarang memantau %d cryptocurrency.`,
		symbol,
		newCrypto.Name,
		symbol,
		len(ns.botService.cryptoList),
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💰 Lihat Coins", "coins_list"),
			tgbotapi.NewInlineKeyboardButtonData("➕ Tambah Lagi", "add_coin"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	ns.telegramBot.Send(msg)

	logrus.Infof("Added new cryptocurrency to watchlist: %s", symbol)
}

// removeCoinFromWatch removes a cryptocurrency from watchlist
func (ns *NotificationService) removeCoinFromWatch(chatID int64, symbol string) {
	if ns.botService == nil {
		ns.sendErrorMessage(chatID, "Bot service tidak tersedia")
		return
	}

	// Find and remove from crypto list
	found := false
	for i, crypto := range ns.botService.cryptoList {
		if crypto.Symbol == symbol {
			// Remove from slice
			ns.botService.cryptoList = append(ns.botService.cryptoList[:i], ns.botService.cryptoList[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		message := fmt.Sprintf("⚠️ *%s tidak ditemukan dalam watchlist*", symbol)
		msg := tgbotapi.NewMessage(chatID, message)
		msg.ParseMode = "Markdown"
		ns.telegramBot.Send(msg)
		return
	}

	message := fmt.Sprintf(`✅ *%s berhasil dihapus dari watchlist*

Bot sekarang memantau %d cryptocurrency.`,
		symbol,
		len(ns.botService.cryptoList),
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💰 Lihat Coins", "coins_list"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	ns.telegramBot.Send(msg)

	logrus.Infof("Removed cryptocurrency from watchlist: %s", symbol)
}

// sendSettingsMenu sends settings configuration menu
func (ns *NotificationService) sendSettingsMenu(chatID int64) {
	message := `⚙️ *Pengaturan Bot*

🔧 *Konfigurasi Saat Ini:*
• Min Confidence: 70%
• Max Signals/Day: 10
• Analysis Interval: 15 menit
• Stop Loss: 5%
• Take Profit 1: 3%
• Take Profit 2: 6%

_Pengaturan lanjutan akan tersedia di versi mendatang_`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	ns.telegramBot.Send(msg)
}

// sendDailySummaryNow sends daily summary immediately
func (ns *NotificationService) sendDailySummaryNow(chatID int64) {
	message := fmt.Sprintf(`📋 *Daily Summary - %s*

📊 *Statistik Hari Ini:*
• Sinyal Dikirim: %d
• Analisis Dilakukan: Auto + Manual
• Coins Dipantau: %d
• Bot Status: %s

📈 *Market Overview:*
• Fear & Greed Index: Updating...
• Top Performer: Updating...
• Market Trend: Updating...

🧠 *Learning Progress:*
• Data Points Collected: Updating...
• Model Accuracy: Updating...

_Summary lengkap dikirim otomatis setiap hari pukul 23:00_`,
		time.Now().Format("02/01/2006"),
		ns.botService.totalSignalsToday,
		len(ns.botService.cryptoList),
		func() string {
			if ns.botService.isRunning {
				return "🟢 Running"
			}
			return "🔴 Stopped"
		}(),
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 Performance", "performance"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	ns.telegramBot.Send(msg)
}

// sendLearningStats sends AI learning statistics
func (ns *NotificationService) sendLearningStats(chatID int64) {
	message := `🧠 *AI Learning Statistics*

📚 *Learning Data:*
• Total Data Points: 0
• Training Samples: 0
• Validation Accuracy: 0%

🎯 *Pattern Recognition:*
• Bullish Patterns: 0
• Bearish Patterns: 0
• Neutral Patterns: 0

📊 *Indicator Performance:*
• RSI Accuracy: 0%
• MACD Accuracy: 0%
• Bollinger Bands: 0%

🔄 *Model Updates:*
• Last Training: Never
• Next Update: Auto
• Improvement Rate: 0%

_Data akan tersedia setelah bot mengumpulkan cukup data trading_`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 Performance", "performance"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	ns.telegramBot.Send(msg)
}

// Helper functions
func getCoinName(symbol string) string {
	coinNames := map[string]string{
		"DOGE":  "Dogecoin",
		"SHIB":  "Shiba Inu",
		"PEPE":  "Pepe",
		"FLOKI": "Floki Inu",
		"TRX":   "TRON",
		"XRP":   "Ripple",
	}
	
	if name, exists := coinNames[symbol]; exists {
		return name
	}
	return symbol
}

func getCoinGeckoID(symbol string) string {
	geckoIDs := map[string]string{
		"DOGE":  "dogecoin",
		"SHIB":  "shiba-inu",
		"PEPE":  "pepe",
		"FLOKI": "floki",
		"TRX":   "tron",
		"XRP":   "ripple",
	}
	
	if id, exists := geckoIDs[symbol]; exists {
		return id
	}
	return strings.ToLower(symbol)
}
