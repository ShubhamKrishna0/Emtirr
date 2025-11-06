# ⚡ 4 in a Row - Real-time Multiplayer Game 🎯

A professional Connect Four game with real-time multiplayer, competitive AI bot, and Kafka analytics.

## 🚀 Quick Start

### 1. Install Dependencies
```bash
npm run setup
```

### 2. Setup Kafka (Optional - for analytics)

**Download Kafka:**
- Download from: https://kafka.apache.org/downloads
- Extract to `C:\kafka`

**Start Kafka (3 terminals):**
```bash
# Terminal 1 - Start Zookeeper
cd C:\kafka
.\bin\windows\zookeeper-server-start.bat .\config\zookeeper.properties

# Terminal 2 - Start Kafka
cd C:\kafka
.\bin\windows\kafka-server-start.bat .\config\server.properties

# Terminal 3 - Create Topic
cd C:\kafka
.\bin\windows\kafka-topics.bat --create --topic game-events --bootstrap-server localhost:9092
```

### 3. Setup Database
```bash
# Create PostgreSQL database
createdb four_in_a_row

# Or using psql
psql -U postgres -c "CREATE DATABASE four_in_a_row;"
```

### 4. Configure Environment
```bash
# Copy environment template
copy .env.example .env

# Edit .env file with your database password
```

### 5. Run Application
```bash
npm start
```

**Game available at:** `http://localhost:3001`

## 🎮 How to Play

1. Enter username
2. Wait for opponent (bot joins after 10 seconds)
3. Click columns to drop discs
4. Connect 4 discs to win!
5. View leaderboard

## 🏗️ Architecture

- **Backend:** Node.js + Express + Socket.IO + PostgreSQL + Kafka
- **Frontend:** React + Socket.IO Client
- **AI Bot:** Minimax algorithm with alpha-beta pruning
- **Analytics:** Real-time Kafka event streaming

## 📋 Prerequisites

- Node.js (v14+)
- PostgreSQL (v12+)
- Kafka (optional)

## 🔧 Development Commands

```bash
npm run install-all    # Install all dependencies
npm run build          # Build frontend
npm start              # Start application
npm run dev            # Development mode
npm run start:backend  # Backend only
npm run start:frontend # Frontend only
```

## 🎯 Features

✅ Real-time multiplayer with WebSockets  
✅ Competitive AI bot (10-second fallback)  
✅ 30-second reconnection system  
✅ PostgreSQL game persistence  
✅ Live leaderboard  
✅ Kafka analytics pipeline  
✅ Responsive React UI  

## 🚀 Deployment

**Heroku:**
```bash
heroku create your-app-name
heroku addons:create heroku-postgresql:hobby-dev
git push heroku main
```

**Docker:**
```bash
docker-compose up -d
```

Built with ❤️ for Emitrr Backend Engineering Assignment