# 🚀 Production Deployment Guide

## 📋 Prerequisites

1. ✅ Bot tested locally and working
2. ✅ GitHub repository ready
3. ✅ Supabase database configured
4. ✅ Telegram bot token obtained
5. ✅ CoinMarketCap API key ready

## 🚂 Option 1: Deploy to Railway (Recommended)

### Step 1: Prepare Repository
```bash
# Initialize git if not already done
git init
git add .
git commit -m "Initial commit: Crypto Signal Bot"

# Push to GitHub
git remote add origin https://github.com/yourusername/crypto-signal-bot.git
git push -u origin main
```

### Step 2: Deploy to Railway
1. Go to [railway.app](https://railway.app)
2. Sign up/Login with GitHub
3. Click "New Project" → "Deploy from GitHub repo"
4. Select your repository
5. Railway will auto-detect Dockerfile and deploy

### Step 3: Configure Environment Variables
In Railway dashboard, go to Variables tab and add:

```env
COINMARKETCAP_API_KEY=983f33a6-b19d-49fd-80d7-8603890f094b
SUPABASE_URL=https://syojcjdcpufgyojnxhqa.supabase.co
SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InN5b2pjamRjcHVmZ3lvam54aHFhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NTE1MjczODMsImV4cCI6MjA2NzEwMzM4M30.NVxtRoDYpDozaNtXnGvn4jN2ToqgWbpXhX02uuWNTko
SUPABASE_SERVICE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InN5b2pjamRjcHVmZ3lvam54aHFhIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc1MTUyNzM4MywiZXhwIjoyMDY3MTAzMzgzfQ.k4Wqq3S1Ogk9q1BfMPG11SRhSC-Yk7XQAgYxIYhPJcU
TELEGRAM_BOT_TOKEN=7685238155:AAFUWTRiERicLs1R4t8B1EIz6aLunNeQkRw
TELEGRAM_CHAT_ID=@dejavusinyal_bot
LOG_LEVEL=info
MIN_CONFIDENCE_THRESHOLD=0.70
MAX_SIGNALS_PER_DAY=10
ANALYSIS_INTERVAL_SECONDS=900
```

### Step 4: Monitor Deployment
- Check logs in Railway dashboard
- Test health endpoint: `https://your-app.railway.app/health`
- Test Telegram bot with `/start` command

## 🎨 Option 2: Deploy to Render

### Step 1: Prepare Repository (same as Railway)

### Step 2: Deploy to Render
1. Go to [render.com](https://render.com)
2. Sign up/Login with GitHub
3. Click "New" → "Web Service"
4. Connect your GitHub repository
5. Use these settings:
   - **Build Command**: `go build -o crypto-signal-bot .`
   - **Start Command**: `./crypto-signal-bot`
   - **Environment**: `Go`

### Step 3: Configure Environment Variables
Add the same environment variables as Railway

## 🔧 Post-Deployment Checklist

### ✅ Verify Deployment
1. **Health Check**: Visit `https://your-app.domain/health`
2. **API Status**: Visit `https://your-app.domain/api/v1/bot/status`
3. **Telegram Test**: Send `/start` to your bot
4. **Menu Test**: Send `/menu` and test interactive buttons

### ✅ Monitor Performance
1. **Check Logs**: Monitor application logs for errors
2. **Database Connection**: Verify Supabase connectivity
3. **API Limits**: Monitor CoinMarketCap API usage
4. **Signal Generation**: Confirm signals are being generated

### ✅ Set Up Monitoring
1. **Uptime Monitoring**: Use UptimeRobot or similar
2. **Error Alerts**: Configure log-based alerts
3. **Performance Metrics**: Monitor response times

## 🚨 Troubleshooting

### Common Issues:

#### 1. Database Connection Failed
```
Error: dial tcp: lookup db.syojcjdcpufgyojnxhqa.supabase.co: no such host
```
**Solution**: Check Supabase URL and service key

#### 2. Telegram Bot Not Responding
```
Error: chat not found
```
**Solution**: 
- Start conversation with bot first (`/start`)
- Verify bot token and chat ID

#### 3. API Rate Limits
```
Error: API rate limit exceeded
```
**Solution**: 
- Check CoinMarketCap API usage
- Increase analysis interval if needed

#### 4. Memory/CPU Limits
**Solution**:
- Upgrade to paid tier if needed
- Optimize analysis frequency

## 📊 Production Monitoring

### Key Metrics to Monitor:
- ✅ Bot uptime
- ✅ Signal generation rate
- ✅ API response times
- ✅ Database connection status
- ✅ Memory/CPU usage
- ✅ Error rates

### Recommended Tools:
- **Uptime**: UptimeRobot, Pingdom
- **Logs**: Railway/Render built-in logs
- **Performance**: New Relic, DataDog (free tiers)

## 🔄 Updates and Maintenance

### Automatic Deployments:
- Push to main branch triggers auto-deploy
- Monitor deployment logs
- Test after each deployment

### Manual Maintenance:
- Weekly: Check API usage and limits
- Monthly: Review performance metrics
- Quarterly: Update dependencies

## 🎯 Success Indicators

Your deployment is successful when:
- ✅ Health endpoint returns 200 OK
- ✅ Telegram bot responds to `/start`
- ✅ Interactive menu works properly
- ✅ Signals are generated and sent
- ✅ No errors in application logs
- ✅ Database connections are stable

## 📞 Support

If you encounter issues:
1. Check application logs first
2. Verify all environment variables
3. Test individual components (DB, Telegram, API)
4. Review this troubleshooting guide
