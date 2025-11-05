# 🔴 4 in a Row - Real-time Multiplayer Game

A professional-grade implementation of the classic Connect Four game with real-time multiplayer capabilities, competitive AI bot, and comprehensive analytics using Kafka.

## 🚀 Features

### Core Gameplay
- **Real-time multiplayer** using WebSockets (Socket.IO)
- **Competitive AI bot** with minimax algorithm and alpha-beta pruning
- **Automatic matchmaking** with 10-second bot fallback
- **Reconnection system** with 30-second grace period
- **Game state persistence** using PostgreSQL

### Advanced Features
- **Live leaderboard** with player statistics
- **Analytics pipeline** using Kafka for event streaming
- **Responsive web interface** built with React
- **Professional code architecture** with separation of concerns

## 🛠 Tech Stack

### Backend
- **Node.js** with Express.js
- **Socket.IO** for real-time communication
- **PostgreSQL** for data persistence
- **Kafka** for analytics and event streaming
- **Minimax AI** with alpha-beta pruning

### Frontend
- **React** with modern hooks
- **Socket.IO Client** for real-time updates
- **CSS3** with animations and responsive design
- **Modern UI/UX** with glassmorphism effects

## 📋 Prerequisites

Before running the application, ensure you have:

1. **Node.js** (v14 or higher)
2. **PostgreSQL** (v12 or higher)
3. **Apache Kafka** (v2.8 or higher)
4. **npm** or **yarn** package manager

## 🔧 Installation & Setup

### 1. Clone and Install Dependencies

```bash
# Clone the repository
git clone <repository-url>
cd 4-in-a-row-game

# Install backend dependencies
npm install

# Install frontend dependencies
cd client
npm install
cd ..
```

### 2. Database Setup

```bash
# Create PostgreSQL database
createdb four_in_a_row

# Or using psql
psql -U postgres
CREATE DATABASE four_in_a_row;
\q
```

### 3. Kafka Setup

**Windows:**
```bash
# Terminal 1 - Start Zookeeper
cd C:\kafka
.\bin\windows\zookeeper-server-start.bat .\config\zookeeper.properties

# Terminal 2 - Start Kafka
cd C:\kafka
.\bin\windows\kafka-server-start.bat .\config\server.properties

# Terminal 3 - Create topic
cd C:\kafka
.\bin\windows\kafka-topics.bat --create --topic game-events --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
```

**Linux/Mac:**
```bash
# Start Zookeeper (in separate terminal)
bin/zookeeper-server-start.sh config/zookeeper.properties

# Start Kafka Server (in separate terminal)
bin/kafka-server-start.sh config/server.properties

# Create required topic
bin/kafka-topics.sh --create --topic game-events --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
```

### 4. Environment Configuration

Create a `.env` file in the root directory:

```env
PORT=3001
DB_HOST=localhost
DB_PORT=5432
DB_NAME=four_in_a_row
DB_USER=postgres
DB_PASSWORD=your_password
KAFKA_BROKER=localhost:9092
NODE_ENV=development
```

### 5. Build and Start

```bash
# Build React frontend
npm run build

# Start the server
npm start

# For development (with auto-reload)
npm run dev
```

The application will be available at `http://localhost:3001`

## 🎮 How to Play

1. **Enter Username**: Start by entering your username
2. **Matchmaking**: Wait for an opponent or play against the AI bot
3. **Gameplay**: Click columns to drop your discs
4. **Objective**: Connect 4 discs vertically, horizontally, or diagonally
5. **Reconnection**: If disconnected, rejoin within 30 seconds to continue

## 🏗 Architecture Overview

### Backend Structure
```
src/
├── game/
│   ├── GameManager.js     # Game lifecycle and matchmaking
│   └── Bot.js            # AI bot with minimax algorithm
├── models/
│   └── Game.js           # Game state and logic
├── services/
│   ├── DatabaseService.js # PostgreSQL operations
│   └── AnalyticsService.js # Kafka event processing
└── utils/               # Utility functions
```

### Frontend Structure
```
client/src/
├── components/
│   ├── GameBoard.js      # Interactive game grid
│   └── Leaderboard.js    # Player rankings
├── App.js               # Main application component
└── *.css               # Styling and animations
```

## 🤖 AI Bot Implementation

The competitive bot uses advanced algorithms:

- **Minimax Algorithm** with 6-move lookahead
- **Alpha-Beta Pruning** for optimization
- **Strategic Evaluation** including:
  - Center column preference
  - Threat detection and blocking
  - Winning opportunity recognition
  - Position scoring system

## 📊 Analytics & Kafka Integration

### Event Types Tracked
- `game_started` - New game initialization
- `move_made` - Player moves
- `game_ended` - Game completion
- `player_disconnected` - Connection issues
- `player_rejoined` - Successful reconnections
- `bot_move` - AI decisions

### Analytics Capabilities
- Real-time game metrics
- Player behavior analysis
- Bot performance evaluation
- System health monitoring

## 🔌 API Endpoints

### REST API
- `GET /api/leaderboard` - Fetch player rankings
- `GET /api/analytics` - Get game analytics

### WebSocket Events
- `join_game` - Enter matchmaking
- `make_move` - Submit game move
- `rejoin_game` - Reconnect to existing game

## 🧪 Testing

```bash
# Run backend tests
npm test

# Run frontend tests
cd client && npm test
```

## 🚀 Deployment

### Production Build
```bash
# Build optimized frontend
cd client && npm run build

# Set production environment
export NODE_ENV=production

# Start production server
npm start
```

### Docker Deployment
```bash
# Build Docker image
docker build -t 4-in-a-row .

# Run with docker-compose
docker-compose up -d
```

## 📈 Performance Considerations

- **Connection Pooling** for database efficiency
- **Event Batching** for Kafka optimization
- **Memory Management** for active games
- **Graceful Cleanup** of disconnected sessions

## 🔒 Security Features

- **Input Validation** for all user inputs
- **Rate Limiting** for API endpoints
- **CORS Configuration** for cross-origin requests
- **Environment Variables** for sensitive data

## 🐛 Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Verify PostgreSQL is running
   - Check database credentials in `.env`

2. **Kafka Connection Error**
   - Ensure Kafka and Zookeeper are running
   - Verify broker address in configuration

3. **WebSocket Connection Issues**
   - Check firewall settings
   - Verify CORS configuration

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👨‍💻 Author

**Emitrr Backend Engineering Intern Assignment**

Built with ❤️ using modern web technologies and best practices for scalable, maintainable code.

---

## 🎯 Interview Preparation Notes

This codebase demonstrates:

- **System Design**: Microservices architecture with clear separation
- **Real-time Systems**: WebSocket implementation with reconnection logic
- **Algorithm Implementation**: Minimax AI with optimization
- **Database Design**: Efficient schema and query optimization
- **Event-Driven Architecture**: Kafka integration for analytics
- **Error Handling**: Comprehensive error management
- **Code Quality**: Clean, documented, and maintainable code
- **Scalability**: Designed for horizontal scaling
- **Testing**: Unit and integration test structure
- **DevOps**: Docker and deployment configurations