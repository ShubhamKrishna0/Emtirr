const express = require('express');
const http = require('http');
const socketIo = require('socket.io');
const cors = require('cors');
const path = require('path');
require('dotenv').config();

const GameManager = require('./src/game/GameManager');
const DatabaseService = require('./src/services/DatabaseService');
const AnalyticsService = require('./src/services/AnalyticsService');
const { validateUsername, validateMove } = require('./src/middleware/security');

const app = express();
const server = http.createServer(app);
const io = socketIo(server, {
  cors: {
    origin: "*",
    methods: ["GET", "POST"]
  }
});

// Middleware
app.use(cors());
app.use(express.json());
app.use(express.static(path.join(__dirname, '../frontend/build')));

// Services
const dbService = new DatabaseService();
const analyticsService = new AnalyticsService();
const gameManager = new GameManager(io, dbService, analyticsService);

// Routes
app.get('/api/leaderboard', async (req, res) => {
  try {
    const leaderboard = await dbService.getLeaderboard();
    res.json(leaderboard);
  } catch (error) {
    // Return empty leaderboard if DB unavailable
    res.json([]);
  }
});

app.get('/api/analytics', async (req, res) => {
  try {
    const analytics = await dbService.getAnalytics();
    res.json(analytics);
  } catch (error) {
    res.status(500).json({ error: 'Failed to fetch analytics' });
  }
});

// Serve React app
app.get('*', (req, res) => {
  res.sendFile(path.join(__dirname, '../frontend/build', 'index.html'));
});

// Socket.IO connection handling
io.on('connection', (socket) => {
  console.log(`Player connected: ${socket.id}`);
  
  socket.on('join_game', (data) => {
    try {
      gameManager.handlePlayerJoin(socket, data);
    } catch (error) {
      console.error('Join game error:', error.message);
      socket.emit('error', { message: 'Failed to join game' });
    }
  });
  
  socket.on('make_move', (data) => {
    gameManager.handlePlayerMove(socket, data);
  });
  
  socket.on('rejoin_game', (data) => {
    gameManager.handlePlayerRejoin(socket, data);
  });
  
  socket.on('disconnect', () => {
    gameManager.handlePlayerDisconnect(socket);
  });
});

// Initialize services
async function initialize() {
  try {
    await analyticsService.initialize();
    console.log('Analytics initialized');
    
    // Try database, continue without it if it fails
    try {
      await dbService.initialize();
      console.log('Database initialized successfully');
    } catch (dbError) {
      console.log('Database unavailable, continuing without persistence');
    }
  } catch (error) {
    console.error('Service initialization error:', error);
  }
}

const PORT = process.env.PORT || 3001;
server.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
  initialize();
});