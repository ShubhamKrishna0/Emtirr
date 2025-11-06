import React from 'react';
import './GameBoard.css';

const GameBoard = ({ board, onColumnClick, currentPlayer, yourPlayer, gameStatus }) => {
  const handleColumnClick = (column) => {
    if (gameStatus === 'playing' && currentPlayer === yourPlayer) {
      onColumnClick(column);
    }
  };

  const getCellClass = (cell, rowIndex, colIndex) => {
    let className = 'cell';
    
    if (cell === 1) {
      className += ' player1';
    } else if (cell === 2) {
      className += ' player2';
    } else {
      className += ' empty';
    }
    
    return className;
  };

  const getColumnClass = (colIndex) => {
    let className = 'column';
    
    if (gameStatus === 'playing' && currentPlayer === yourPlayer) {
      // Check if column is full
      const isColumnFull = board[0][colIndex] !== 0;
      if (!isColumnFull) {
        className += ' clickable';
      }
    }
    
    return className;
  };

  return (
    <div className="game-board">
      <div className="board-container">
        {Array.from({ length: 7 }, (_, colIndex) => (
          <div
            key={colIndex}
            className={getColumnClass(colIndex)}
            onClick={() => handleColumnClick(colIndex)}
          >
            {Array.from({ length: 6 }, (_, rowIndex) => (
              <div
                key={`${rowIndex}-${colIndex}`}
                className={getCellClass(board[rowIndex][colIndex], rowIndex, colIndex)}
              >
                <div className="disc"></div>
              </div>
            ))}
          </div>
        ))}
      </div>
      
      <div className="game-status">
        <div className="players">
          <div className={`player-indicator ${yourPlayer === 1 ? 'you' : ''} ${currentPlayer === 1 ? 'active' : ''}`}>
            <div className="player-disc player1-disc"></div>
            <span>Player 1 {yourPlayer === 1 ? '(You)' : ''}</span>
          </div>
          <div className={`player-indicator ${yourPlayer === 2 ? 'you' : ''} ${currentPlayer === 2 ? 'active' : ''}`}>
            <div className="player-disc player2-disc"></div>
            <span>Player 2 {yourPlayer === 2 ? '(You)' : ''}</span>
          </div>
        </div>
        
        {gameStatus === 'playing' && (
          <div className="turn-indicator">
            {currentPlayer === yourPlayer ? (
              <span className="your-turn">Your Turn!</span>
            ) : (
              <span className="opponent-turn">Opponent's Turn</span>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default GameBoard;