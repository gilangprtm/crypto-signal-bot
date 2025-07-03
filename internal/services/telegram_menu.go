package services

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// sendWelcomeMessage sends welcome message with main menu
func (ns *NotificationService) sendWelcomeMessage(chatID int64) {
	message := `🤖 *Selamat datang di Crypto Signal Bot!*

Bot ini akan membantu Anda mendapatkan sinyal trading cryptocurrency yang akurat dengan analisis teknikal real-time.

🎯 *Fitur Utama:*
• 📊 Analisis 10+ cryptocurrency
• 🔍 Technical indicators (RSI, MACD, BB)
• 🧠 AI learning untuk meningkatkan akurasi
• 📱 Notifikasi real-time
• 📈 Performance tracking

Gunakan /menu untuk melihat semua fitur yang tersedia.

⚡ *Quick Commands:*
/menu - Menu utama
/status - Status bot
/coins - Daftar coin yang dipantau
/performance - Laporan performa`

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	
	// Add main menu button
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Menu Utama", "main_menu"),
		),
	)
	msg.ReplyMarkup = keyboard

	ns.telegramBot.Send(msg)
}

// sendMainMenu sends the main interactive menu
func (ns *NotificationService) sendMainMenu(chatID int64) {
	message := `🤖 *Crypto Signal Bot - Menu Utama*

Pilih opsi yang ingin Anda akses:`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Status Bot", "bot_status"),
			tgbotapi.NewInlineKeyboardButtonData("🔍 Analisis Manual", "manual_analysis"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💰 Daftar Coins", "coins_list"),
			tgbotapi.NewInlineKeyboardButtonData("➕ Tambah Coin", "add_coin"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 Performance", "performance"),
			tgbotapi.NewInlineKeyboardButtonData("🧠 Learning Stats", "learning_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Daily Summary", "daily_summary"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Settings", "settings"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	ns.telegramBot.Send(msg)
}

// sendBotStatus sends current bot status
func (ns *NotificationService) sendBotStatus(chatID int64) {
	if ns.botService == nil {
		ns.sendErrorMessage(chatID, "Bot service tidak tersedia")
		return
	}

	status := "🔴 Stopped"
	if ns.botService.isRunning {
		status = "🟢 Running"
	}

	lastAnalysis := "Belum pernah"
	if !ns.botService.lastAnalysisTime.IsZero() {
		lastAnalysis = ns.botService.lastAnalysisTime.Format("15:04 02/01/2006")
	}

	message := fmt.Sprintf(`📊 *Status Bot*

🤖 *Status:* %s
📊 *Coins Dipantau:* %d
📈 *Sinyal Hari Ini:* %d
🕐 *Analisis Terakhir:* %s
⏰ *Waktu Sekarang:* %s`,
		status,
		len(ns.botService.cryptoList),
		ns.botService.totalSignalsToday,
		lastAnalysis,
		time.Now().Format("15:04 02/01/2006"),
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", "bot_status"),
			tgbotapi.NewInlineKeyboardButtonData("🔍 Analisis Manual", "manual_analysis"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	ns.telegramBot.Send(msg)
}

// sendCoinsList sends list of monitored coins
func (ns *NotificationService) sendCoinsList(chatID int64) {
	if ns.botService == nil {
		ns.sendErrorMessage(chatID, "Bot service tidak tersedia")
		return
	}

	var coinsList strings.Builder
	coinsList.WriteString("💰 *Daftar Cryptocurrency yang Dipantau:*\n\n")

	for i, crypto := range ns.botService.cryptoList {
		status := "🟢"
		if !crypto.IsActive {
			status = "🔴"
		}
		coinsList.WriteString(fmt.Sprintf("%d. %s *%s* - %s\n", 
			i+1, status, crypto.Symbol, crypto.Name))
	}

	if len(ns.botService.cryptoList) == 0 {
		coinsList.WriteString("Tidak ada cryptocurrency yang dipantau.")
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Tambah Coin", "add_coin"),
			tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", "coins_list"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, coinsList.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	ns.telegramBot.Send(msg)
}

// sendAddCoinMenu sends menu to add new coins
func (ns *NotificationService) sendAddCoinMenu(chatID int64) {
	message := `➕ *Tambah Cryptocurrency Baru*

Pilih cryptocurrency yang ingin ditambahkan ke watchlist:`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🐕 DOGE", "add_coin_DOGE"),
			tgbotapi.NewInlineKeyboardButtonData("🐕 SHIB", "add_coin_SHIB"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🐸 PEPE", "add_coin_PEPE"),
			tgbotapi.NewInlineKeyboardButtonData("🔥 FLOKI", "add_coin_FLOKI"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚡ TRX", "add_coin_TRX"),
			tgbotapi.NewInlineKeyboardButtonData("🌊 XRP", "add_coin_XRP"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Kembali", "coins_list"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	ns.telegramBot.Send(msg)
}

// sendPerformanceReport sends performance statistics
func (ns *NotificationService) sendPerformanceReport(chatID int64) {
	message := `📈 *Laporan Performance*

🎯 *Statistik Sinyal:*
• Total Sinyal: 0
• Win Rate: 0%
• Profit/Loss: 0%

📊 *Analisis Teknikal:*
• RSI Accuracy: 0%
• MACD Accuracy: 0%
• BB Accuracy: 0%

🧠 *Learning Engine:*
• Data Points: 0
• Model Accuracy: 0%
• Last Update: Never

_Data akan tersedia setelah bot berjalan beberapa waktu_`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", "performance"),
			tgbotapi.NewInlineKeyboardButtonData("🧠 Learning Stats", "learning_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	ns.telegramBot.Send(msg)
}

// sendHelpMessage sends help information
func (ns *NotificationService) sendHelpMessage(chatID int64) {
	message := `❓ *Bantuan - Crypto Signal Bot*

🤖 *Commands:*
/start - Mulai bot dan tampilkan welcome
/menu - Tampilkan menu utama
/status - Cek status bot
/coins - Lihat daftar coins
/performance - Laporan performa
/help - Tampilkan bantuan ini

📱 *Cara Menggunakan:*
1. Bot akan otomatis menganalisis market setiap 15 menit
2. Sinyal akan dikirim jika confidence > 70%
3. Gunakan menu interaktif untuk kontrol manual
4. Bot belajar dari hasil sinyal untuk meningkatkan akurasi

🔧 *Fitur:*
• Real-time market analysis
• Technical indicators (RSI, MACD, Bollinger Bands)
• AI learning system
• Risk management (Stop Loss & Take Profit)
• Performance tracking

📞 *Support:* Hubungi developer jika ada masalah`

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

// sendUnknownCommandMessage sends message for unknown commands
func (ns *NotificationService) sendUnknownCommandMessage(chatID int64) {
	message := `❓ *Command tidak dikenal*

Gunakan /menu untuk melihat semua fitur yang tersedia.

*Available Commands:*
/start, /menu, /status, /coins, /performance, /help`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Menu Utama", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	ns.telegramBot.Send(msg)
}

// sendErrorMessage sends error message
func (ns *NotificationService) sendErrorMessage(chatID int64, errorMsg string) {
	message := fmt.Sprintf("🚨 *Error*\n\n%s", errorMsg)
	
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	
	ns.telegramBot.Send(msg)
}
