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
  const [disconnectedGameId, setDisconnectedGameId] = useState(null);
  const [reconnectAttempted, setReconnectAttempted] = useState(false);

  useEffect(() => {
    let reconnectTimer = null;
    
    const connectWebSocket = () => {
      if (socket && socket.readyState === WebSocket.OPEN) return;
      
      const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${wsProtocol}//${window.location.host}/ws`;
      
      socket = new WebSocket(wsUrl);
    
    socket.onopen = () => {
      setIsConnected(true);
      console.log('WebSocket connected');
      if (reconnectTimer) {
        clearTimeout(reconnectTimer);
        reconnectTimer = null;
      }
      
      // Auto-reconnect if we have stored game info and no existing disconnected game
      const storedUsername = localStorage.getItem('gameUsername');
      const storedGameId = localStorage.getItem('gameId');
      if (storedUsername && storedGameId && !reconnectAttempted && !disconnectedGameId) {
        setReconnectAttempted(true);
        setUsername(storedUsername);
        setDisconnectedGameId(storedGameId);
        console.log('Attempting auto-reconnect:', { username: storedUsername, gameId: storedGameId });
        socket.send(JSON.stringify({
          type: 'rejoin_game',
          data: { username: storedUsername, gameId: storedGameId }
        }));
      }
    };
    
    socket.onclose = () => {
      setIsConnected(false);
      console.log('WebSocket closed');
      
      // Store game info for reconnection if we're in a game
      if (gameState && gameStatus === 'playing') {
        localStorage.setItem('gameUsername', username);
        localStorage.setItem('gameId', gameState.id);
        setMessage('Connection lost. Attempting to reconnect...');
      }
      
      // Always try to reconnect
      if (!reconnectTimer) {
        reconnectTimer = setTimeout(() => {
          reconnectTimer = null;
          connectWebSocket();
        }, 1000);
      }
    };
    
    socket.onerror = () => {
      console.log('WebSocket error');
    };
    };
    
    connectWebSocket();
    
    return () => {
      if (reconnectTimer) {
        clearTimeout(reconnectTimer);
      }
      if (socket) {
        socket.close();
      }
    };
  }, []);

  useEffect(() => {
    if (!socket) return;
    
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
          // Store game info for reconnection
          localStorage.setItem('gameUsername', username);
          localStorage.setItem('gameId', data.gameState.id);
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
          // Clear stored game info
          localStorage.removeItem('gameUsername');
          localStorage.removeItem('gameId');
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
        case 'player_reconnected':
          setMessage(`${data.player} has reconnected!`);
          break;
        case 'game_rejoined':
          setGameState(data.gameState);
          setGameStatus('playing');
          setYourPlayer(data.yourPlayer);
          setMessage('Successfully reconnected to your game!');
          setDisconnectedGameId(null);
          // Update stored game info
          localStorage.setItem('gameUsername', username);
          localStorage.setItem('gameId', data.gameState.id);
          break;
        case 'error':
          setMessage(`Error: ${data.message}`);
          // Clear reconnection data on any error during rejoin
          localStorage.removeItem('gameUsername');
          localStorage.removeItem('gameId');
          setDisconnectedGameId(null);
          setReconnectAttempted(false);
          break;
        default:
          console.log('Unknown message type:', type);
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

  const rejoinGame = () => {
    if (username.trim() && disconnectedGameId && socket && socket.readyState === WebSocket.OPEN) {
      console.log('Attempting to rejoin game:', { username: username.trim(), gameId: disconnectedGameId });
      socket.send(JSON.stringify({
        type: 'rejoin_game',
        data: { username: username.trim(), gameId: disconnectedGameId }
      }));
    }
  };

  const makeMove = (column) => {
    console.log('makeMove called:', { column, gameState, gameStatus, currentPlayer: gameState?.currentPlayer, yourPlayer });
    
    // Reconnect if disconnected
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      console.log('WebSocket not ready');
      return;
    }
    
    if (gameState && gameStatus === 'playing' && gameState.currentPlayer === yourPlayer) {
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
    setDisconnectedGameId(null);
    setReconnectAttempted(false);
    // Clear stored game info
    localStorage.removeItem('gameUsername');
    localStorage.removeItem('gameId');
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
            <h2>🚀 ENTER THE ARENA 🎮</h2>
            <p className="subtitle">⚡ CONNECT FOUR TO DOMINATE ⚡</p>
            {disconnectedGameId && (
              <div className="reconnect-section">
                <h3>🔄 Reconnect to Game</h3>
                <p>You have an active game waiting for you!</p>
                <button onClick={rejoinGame} disabled={!username.trim()} className="rejoin-btn">
                  🔄 Rejoin Game
                </button>
                <button onClick={() => setDisconnectedGameId(null)} className="new-game-btn">
                  🆕 Start New Game Instead
                </button>
              </div>
            )}
            <div className="username-input">
              <input
                type="text"
                placeholder="⚡ ENTER GAMER TAG ⚡"
                className="gamer-input"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && (disconnectedGameId ? rejoinGame() : joinGame())}
                maxLength={20}
              />
              <button onClick={disconnectedGameId ? rejoinGame : joinGame} disabled={!username.trim()}>
                {disconnectedGameId ? '🔄 REJOIN BATTLE' : '🚀 LAUNCH GAME'}
              </button>
            </div>
            <button onClick={toggleLeaderboard} className="leaderboard-btn">
              {showLeaderboard ? '🙈 HIDE RANKINGS' : '🏆 VIEW RANKINGS'}
            </button>
          </div>
        )}

        {gameStatus === 'waiting' && (
          <div className="waiting">
            <h2>⚡ SCANNING FOR OPPONENTS ⚡</h2>
            <div className="spinner-container">
              <div className="spinner"></div>
              <div className="scanner-line"></div>
            </div>
            <p className="waiting-message">🤖 {message}</p>
            <div className="loading-dots">
              <span></span><span></span><span></span>
            </div>
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
                  🆕 NEXT BATTLE
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