import React, { useState, useEffect } from 'react';
import GameBoard from './components/GameBoard';
import Leaderboard from './components/Leaderboard';

import './App.css';

let socket = null;

function App() {
  const [gameState, setGameState] = useState(null);
  const [username, setUsername] = useState('');
  const [isConnected, setIsConnected] = useState(false);
  const [gameStatus, setGameStatus] = useState('menu'); // menu, waiting, playing, finished
  const [message, setMessage] = useState('');
  const [yourPlayer, setYourPlayer] = useState(null);
  const [showLeaderboard, setShowLeaderboard] = useState(false);

  useEffect(() => {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//${window.location.host}/ws`;
    
    console.log('Connecting to WebSocket:', wsUrl);
    
    socket = new WebSocket(wsUrl);
    
    socket.onopen = () => {
      setIsConnected(true);
      console.log('WebSocket connected successfully');
    };
    
    socket.onclose = (event) => {
      setIsConnected(false);
      console.log('WebSocket disconnected:', event.code, event.reason);
    };
    
    socket.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
    
    socket.onmessage = (event) => {
      const message = JSON.parse(event.data);
      const { type, data } = message;
      console.log('Received WebSocket message:', { type, data });
      
      switch (type) {
        case 'waiting_for_opponent':
          setGameStatus('waiting');
          setMessage('Waiting for opponent... (Bot will join in 10 seconds)');
          break;
        case 'game_started':
          setGameState(data.gameState);
          setGameStatus('playing');
          setYourPlayer(data.yourPlayer);
          setMessage(`Game started! You are Player ${data.yourPlayer}`);
          break;
        case 'your_turn':
          setYourPlayer(data.player);
          break;
        case 'move_made':
          setGameState(data.gameState);
          const currentPlayer = data.gameState.currentPlayer;
          if (currentPlayer === yourPlayer) {
            setMessage('Your turn!');
          } else {
            setMessage(`Player ${data.player} played. ${data.gameState.isBot && currentPlayer === 2 ? 'Bot is thinking...' : 'Waiting for opponent...'}`);
          }
          break;
        case 'game_ended':
          setGameState(data.gameState);
          setGameStatus('finished');
          if (data.winner === null) {
            setMessage('Game ended in a draw!');
          } else if (data.winner === yourPlayer) {
            setMessage('🎉 You won!');
          } else {
            setMessage(`😔 You lost. Player ${data.winner} won!`);
          }
          break;
        case 'player_disconnected':
          setMessage(`${data.player} disconnected. They have ${data.reconnectTime} seconds to reconnect.`);
          break;
        case 'game_rejoined':
          setGameState(data.gameState);
          setGameStatus('playing');
          setYourPlayer(data.yourPlayer);
          setMessage('Reconnected to game!');
          break;
        case 'error':
          setMessage(`Error: ${data.message}`);
          break;
        default:
          console.log('Unknown message type:', type);
      }
    };
    
    return () => {
      if (socket) {
        socket.close();
      }
    };
  }, [yourPlayer]);

  const joinGame = () => {
    if (username.trim() && socket && socket.readyState === WebSocket.OPEN) {
      console.log('Sending join_game message:', username.trim());
      socket.send(JSON.stringify({
        type: 'join_game',
        data: { username: username.trim() }
      }));
    } else {
      console.log('Cannot join game:', { username: username.trim(), socket, readyState: socket?.readyState });
    }
  };

  const makeMove = (column) => {
    console.log('makeMove called:', { column, gameState, gameStatus, currentPlayer: gameState?.currentPlayer, yourPlayer });
    if (gameState && gameStatus === 'playing' && gameState.currentPlayer === yourPlayer && socket && socket.readyState === WebSocket.OPEN) {
      console.log('Sending move:', { gameId: gameState.id, column });
      socket.send(JSON.stringify({
        type: 'make_move',
        data: { gameId: gameState.id, column }
      }));
    } else {
      console.log('Cannot make move - conditions not met');
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