# ⚡ 4 in a Row - Real-time Multiplayer Game 🎯

A professional Connect Four game with real-time multiplayer, competitive AI bot, and Redis analytics system built with **Go backend** and React frontend.

## 🚀 Live Demo

- **🎮 Play Game**: [https://emitrr-4-in-a-row.onrender.com](https://emitrr-4-in-a-row.onrender.com)
- **📊 Live Analytics**: [https://emitrr-4-in-a-row.onrender.com/api/analytics](https://emitrr-4-in-a-row.onrender.com/api/analytics)
- **📁 GitHub Repo**: [https://github.com/ShubhamKrishna0/Emtirr.git](https://github.com/ShubhamKrishna0/Emtirr.git)

## 🎯 Features

✅ **Real-time Multiplayer** - WebSocket-based gameplay  
✅ **AI Bot Integration** - Smart bot joins after 10 seconds  
✅ **Reconnection System** - 30-second grace period  
✅ **PostgreSQL Persistence** - Game history & leaderboard  
✅ **Redis Analytics** - Real-time event streaming  
✅ **Live Metrics** - Game duration, win rates, player stats  
✅ **Production Ready** - Deployed on Render with full scaling  

## 🏗️ Architecture

```
Frontend (React)     Backend (Go)          Database & Analytics
     │                      │                       │
     ├─ WebSocket ──────────┼─ Gin Server           ├─ PostgreSQL
     ├─ Game Board          ├─ Game Manager         ├─ Redis (Analytics)
     ├─ Leaderboard         ├─ AI Bot Logic         └─ Real-time Metrics
     └─ Real-time UI        └─ Analytics Service
```

## 🚀 Quick Start

```bash
# Clone and setup
git clone https://github.com/ShubhamKrishna0/Emtirr.git
cd Emtirr
go mod tidy
cd frontend && npm install && npm run build
cd .. && go run .
```

**Game available at**: `http://localhost:3001`

## 🎮 How to Play

1. **Enter Username** - Type your name and click "Join Game"
2. **Wait for Opponent** - Another player or bot (after 10 seconds)
3. **Make Moves** - Click columns to drop your discs
4. **Win Condition** - Connect 4 discs horizontally, vertically, or diagonally
5. **View Stats** - Check leaderboard for rankings

## 📊 Analytics System

### Real-Time Event Tracking
- **Game Events**: Start, moves, end, duration
- **Player Metrics**: Win rates, activity patterns
- **Bot Performance**: Decision patterns, effectiveness
- **System Health**: Connection stability, response times

### View Analytics
- **API Endpoint**: `/api/analytics`
- **Live Logs**: Check console for real-time events

## 🚀 Production Deployment

### Deploy to Render

1. **Push to GitHub**:
```bash
git add .
git commit -m "Deploy to production"
git push origin main
```

2. **Connect to Render**:
   - Go to [render.com](https://render.com)
   - Connect GitHub account
   - Select this repository
   - Render auto-detects `render.yaml`

3. **Services Created**:
   - Web Service (Main app)
   - PostgreSQL Database
   - Redis (Analytics)

### Environment Variables (Auto-Set)
```env
NODE_ENV=production
PORT=3001
DATABASE_URL=postgresql://... (auto-generated)
REDIS_URL=redis://... (auto-generated)
```

## 📁 Project Structure

```
Emtirr/
├── internal/
│   ├── config/          # Configuration management
│   ├── game/            # Game logic & AI bot
│   ├── handlers/        # HTTP & WebSocket handlers
│   ├── models/          # Data models
│   └── services/        # Database & Analytics
├── frontend/
│   ├── src/
│   │   ├── components/  # React components
│   │   └── App.js       # Main application
│   └── package.json
├── main.go              # Application entry point
├── render.yaml          # Production deployment
└── README.md
```

## 🔧 Development Commands

```bash
go mod tidy              # Install Go dependencies
go run .                 # Start application
go build -o main .       # Build binary

# Frontend development
cd frontend && npm install && npm run build
```

## 🧪 Testing

1. **Single Player**: Join game, wait for bot
2. **Multiplayer**: Open two browser tabs
3. **Reconnection**: Refresh page during game
4. **Analytics**: Check `/api/analytics`

## 🔍 Troubleshooting

**Database Issues**: App works without database (empty leaderboard)
**Port Issues**: Change PORT in environment
**WebSocket Issues**: Check firewall settings

## 👨💻 Author

**Shubham Krishna**
- GitHub: [@ShubhamKrishna0](https://github.com/ShubhamKrishna0)
- Project: [Emtirr](https://github.com/ShubhamKrishna0/Emtirr)

---

## 🎯 Assignment Requirements Met

✅ **Real-time Multiplayer Game** - WebSocket implementation  
✅ **AI Bot Integration** - Minimax algorithm  
✅ **Database Integration** - PostgreSQL persistence  
✅ **Analytics System** - Redis event streaming  
✅ **Production Deployment** - Live on Render  
✅ **Complete Documentation** - Setup & usage guide  

**Built with ❤️ for Emitrr Backend Engineering Assignment**