import React, { useState, useEffect } from 'react';
import io from 'socket.io-client';
import GameBoard from './components/GameBoard';
import Leaderboard from './components/Leaderboard';

import './App.css';

const socket = io(process.env.REACT_APP_SERVER_URL || 'http://localhost:3001');

function App() {
  const [gameState, setGameState] = useState(null);
  const [username, setUsername] = useState('');
  const [isConnected, setIsConnected] = useState(false);
  const [gameStatus, setGameStatus] = useState('menu'); // menu, waiting, playing, finished
  const [message, setMessage] = useState('');
  const [yourPlayer, setYourPlayer] = useState(null);
  const [showLeaderboard, setShowLeaderboard] = useState(false);

  useEffect(() => {
    // Socket event listeners
    socket.on('connect', () => {
      setIsConnected(true);
      console.log('Connected to server');
    });

    socket.on('disconnect', () => {
      setIsConnected(false);
      console.log('Disconnected from server');
    });

    socket.on('waiting_for_opponent', () => {
      setGameStatus('waiting');
      setMessage('Waiting for opponent... (Bot will join in 10 seconds)');
    });

    socket.on('game_started', (data) => {
      setGameState(data.gameState);
      setGameStatus('playing');
      setYourPlayer(data.yourPlayer);
      setMessage(`Game started! You are Player ${data.yourPlayer}`);
    });

    socket.on('your_turn', (data) => {
      setYourPlayer(data.player);
    });

    socket.on('move_made', (data) => {
      setGameState(data.gameState);
      const currentPlayer = data.gameState.currentPlayer;
      if (currentPlayer === yourPlayer) {
        setMessage('Your turn!');
      } else {
        setMessage(`Player ${data.player} played. ${data.gameState.isBot && currentPlayer === 2 ? 'Bot is thinking...' : 'Waiting for opponent...'}`);
      }
    });

    socket.on('game_ended', (data) => {
      setGameState(data.gameState);
      setGameStatus('finished');
      
      if (data.winner === null) {
        setMessage('Game ended in a draw!');
      } else if (data.winner === yourPlayer) {
        setMessage('🎉 You won!');
      } else {
        setMessage(`😔 You lost. Player ${data.winner} won!`);
      }
    });

    socket.on('player_disconnected', (data) => {
      setMessage(`${data.player} disconnected. They have ${data.reconnectTime} seconds to reconnect.`);
    });

    socket.on('game_rejoined', (data) => {
      setGameState(data.gameState);
      setGameStatus('playing');
      setYourPlayer(data.yourPlayer);
      setMessage('Reconnected to game!');
    });

    socket.on('error', (data) => {
      setMessage(`Error: ${data.message}`);
    });

    return () => {
      socket.off('connect');
      socket.off('disconnect');
      socket.off('waiting_for_opponent');
      socket.off('game_started');
      socket.off('your_turn');
      socket.off('move_made');
      socket.off('game_ended');
      socket.off('player_disconnected');
      socket.off('game_rejoined');
      socket.off('error');
    };
  }, [yourPlayer]);

  const joinGame = () => {
    if (username.trim()) {
      socket.emit('join_game', { username: username.trim() });
    }
  };

  const makeMove = (column) => {
    if (gameState && gameStatus === 'playing' && gameState.currentPlayer === yourPlayer) {
      socket.emit('make_move', { gameId: gameState.id, column });
    }
  };

  const startNewGame = () => {
    setGameState(null);
    setGameStatus('menu');
    setMessage('');
    setYourPlayer(null);
  };

  const toggleLeaderboard = () => {
    setShowLeaderboard(!showLeaderboard);
  };

  return (
    <div className="App">

      <header className="App-header">
        <h1>⚡ 4 in a Row 🎯</h1>
        <div className="connection-status">
          Status: {isConnected ? '✅ Connected' : '❌ Disconnected'}
        </div>
      </header>

      <main className="App-main">
        {gameStatus === 'menu' && (
          <div className="menu">
            <h2>🚀 Welcome to 4 in a Row! 🎮</h2>
            <div className="username-input">
              <input
                type="text"
                placeholder="Enter username... 👤"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && joinGame()}
                maxLength={20}
              />
              <button onClick={joinGame} disabled={!username.trim()}>
                🚀 Join Game
              </button>
            </div>
            <button onClick={toggleLeaderboard} className="leaderboard-btn">
              {showLeaderboard ? '🙈 Hide Leaderboard' : '🏆 Show Leaderboard'}
            </button>
          </div>
        )}

        {gameStatus === 'waiting' && (
          <div className="waiting">
            <h2>🔍 Finding Opponent...</h2>
            <div className="spinner"></div>
            <p>🤖 {message}</p>
          </div>
        )}

        {(gameStatus === 'playing' || gameStatus === 'finished') && gameState && (
          <div className="game-container">
            <div className="game-info">
              <h3>
                {gameState.player1.username} vs {gameState.player2.username}
                {gameState.isBot && ' 🤖'}
              </h3>
              <p className="game-message">{message}</p>
              {gameStatus === 'finished' && (
                <button onClick={startNewGame} className="new-game-btn">
                  🆕 New Game
                </button>
              )}
            </div>
            <GameBoard 
              board={gameState.board}
              onColumnClick={makeMove}
              currentPlayer={gameState.currentPlayer}
              yourPlayer={yourPlayer}
              gameStatus={gameStatus}
            />
          </div>
        )}

        {showLeaderboard && (
          <div className="leaderboard-container">
            <Leaderboard />
          </div>
        )}
      </main>
    </div>
  );
}

export default App;