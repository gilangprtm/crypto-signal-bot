#!/bin/bash

# Production startup script for Crypto Signal Bot
# Prevents multiple instances and ensures clean startup

echo "🚀 Starting Crypto Signal Bot (Production Mode)..."

# Kill any existing instances
echo "🔍 Checking for existing bot instances..."
pkill -f "crypto-signal-bot" || true
sleep 2

# Set environment for IPv4 preference
export GODEBUG=netdns=go+1

# Start the bot
echo "✅ Starting new bot instance..."
./crypto-signal-bot

echo "🛑 Bot stopped"
