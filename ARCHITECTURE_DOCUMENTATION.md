# CRYPTO SIGNAL BOT - ARCHITECTURE DOCUMENTATION

## 🎯 **ALUR SISTEMATIS YANG TELAH DITERAPKAN**

### **1. ANALISIS FITUR → 2. DESAIN DATABASE → 3. IMPLEMENTASI SERVICE**

---

## 📊 **DATABASE SCHEMA (FINAL & KONSISTEN)**

### **Tabel Utama:**

#### **1. `cryptocurrencies`**
- **Purpose**: Menyimpan daftar cryptocurrency yang akan dianalisis
- **Key Fields**: `id`, `symbol`, `name`, `cmc_id`, `contract_address`, `platform`
- **Data Source**: CoinMarketCap API
- **Status**: ✅ HARUS ADA DATA AWAL

#### **2. `market_snapshots`**
- **Purpose**: Menyimpan data market dan technical indicators
- **Key Fields**: `cryptocurrency_id`, `price`, `volume_24h`, `rsi`, `macd_*`, `bb_*`
- **Data Source**: Bot otomatis setiap 15 menit
- **Status**: 🤖 DIISI OTOMATIS

#### **3. `trading_signals`**
- **Purpose**: Menyimpan signal BUY/SELL/HOLD
- **Key Fields**: `cryptocurrency_id`, `action`, `confidence_score`, `entry_price`
- **Data Source**: Analysis engine
- **Status**: 🤖 DIISI OTOMATIS

#### **4. `signal_performance`**
- **Purpose**: Track performance dari setiap signal
- **Key Fields**: `signal_id`, `profit_loss_percentage`, `execution_type`
- **Data Source**: Performance tracker
- **Status**: 🤖 DIISI OTOMATIS

#### **5. `learning_data`**
- **Purpose**: Data untuk machine learning improvement
- **Key Fields**: `signal_id`, `features`, `actual_outcome`
- **Data Source**: Learning engine
- **Status**: 🤖 DIISI OTOMATIS

#### **6. `notification_logs`**
- **Purpose**: Log semua notifikasi Telegram
- **Key Fields**: `type`, `recipient`, `message`, `status`
- **Data Source**: Telegram service
- **Status**: 🤖 DIISI OTOMATIS

#### **7. `system_logs`**
- **Purpose**: Log sistem untuk debugging
- **Key Fields**: `level`, `component`, `message`, `context`
- **Data Source**: Semua komponen
- **Status**: 🤖 DIISI OTOMATIS

#### **8. `bot_settings`**
- **Purpose**: Konfigurasi bot yang bisa diubah runtime
- **Key Fields**: `key`, `value`, `description`, `data_type`
- **Data Source**: Manual/Admin
- **Status**: ✅ HARUS ADA DATA AWAL

---

## 🔧 **MODEL CONSISTENCY**

### **Field Naming Convention:**
- `cryptocurrency_id` (bukan `crypto_id`) - Konsisten di semua tabel
- `created_at`, `updated_at` - Timestamp standard
- `decimal.Decimal` untuk semua nilai finansial
- `uuid.UUID` untuk semua ID

### **Key Models Updated:**
```go
type MarketSnapshot struct {
    ID                 uuid.UUID       `json:"id" db:"id"`
    CryptocurrencyID   uuid.UUID       `json:"cryptocurrency_id" db:"cryptocurrency_id"`
    Price              decimal.Decimal `json:"price" db:"price"`
    // ... technical indicators
}

type Cryptocurrency struct {
    ID              uuid.UUID  `json:"id" db:"id"`
    Symbol          string     `json:"symbol" db:"symbol"`
    Name            string     `json:"name" db:"name"`
    CmcID           *int       `json:"cmc_id" db:"cmc_id"`
    ContractAddress *string    `json:"contract_address" db:"contract_address"`
    Platform        *string    `json:"platform" db:"platform"`
    // ...
}
```

---

## 🚀 **SERVICE ARCHITECTURE**

### **Data Flow:**
1. **Data Collection** → `market_snapshots`
2. **Technical Analysis** → `market_snapshots` (indicators)
3. **Signal Generation** → `trading_signals`
4. **Performance Tracking** → `signal_performance`
5. **Learning** → `learning_data`
6. **Notifications** → `notification_logs`
7. **System Events** → `system_logs`

### **Service Dependencies:**
```
CoinMarketCapService → MarketDataCollector → TechnicalAnalysis
                                          ↓
                                    SignalGenerator
                                          ↓
                                  PerformanceTracker
                                          ↓
                                   LearningEngine
```

---

## 📋 **DATA INITIALIZATION CHECKLIST**

### **✅ REQUIRED INITIAL DATA:**

#### **1. Cryptocurrencies Table:**
```sql
INSERT INTO cryptocurrencies (symbol, name, cmc_id, is_active) VALUES
('BTC', 'Bitcoin', 1, true),
('ETH', 'Ethereum', 1027, true),
('BNB', 'BNB', 1839, true),
('SOL', 'Solana', 5426, true),
('ADA', 'Cardano', 2010, true),
('DOT', 'Polkadot', 6636, true),
('MATIC', 'Polygon', 3890, true),
('AVAX', 'Avalanche', 5805, true),
('LINK', 'Chainlink', 1975, true),
('ATOM', 'Cosmos', 3794, true);
```

#### **2. Bot Settings Table:**
```sql
INSERT INTO bot_settings (key, value, description, data_type) VALUES
('min_confidence_threshold', '0.70', 'Minimum confidence score', 'float'),
('max_signals_per_day', '10', 'Maximum signals per day', 'integer'),
('analysis_interval_minutes', '15', 'Analysis interval', 'integer'),
('rsi_oversold_threshold', '30', 'RSI oversold level', 'float'),
('rsi_overbought_threshold', '70', 'RSI overbought level', 'float');
```

### **🤖 AUTO-POPULATED TABLES:**
- `market_snapshots` - Setiap 15 menit
- `trading_signals` - Saat confidence > threshold
- `signal_performance` - Saat signal triggered/expired
- `learning_data` - Setelah signal selesai
- `notification_logs` - Setiap notifikasi
- `system_logs` - Setiap event

---

## ✅ **CONSISTENCY ACHIEVED**

1. **Database Schema** ✅ Konsisten dan lengkap
2. **Model Definitions** ✅ Sesuai dengan schema
3. **Service Implementation** ✅ Menggunakan field yang benar
4. **Error Handling** ✅ Graceful degradation
5. **Data Flow** ✅ Jelas dan terstruktur

**Tidak ada lagi bolak-balik field names atau schema mismatch!**
