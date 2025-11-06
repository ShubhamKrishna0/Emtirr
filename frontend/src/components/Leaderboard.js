import React, { useState, useEffect } from 'react';
import './Leaderboard.css';

const Leaderboard = () => {
  const [leaderboard, setLeaderboard] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetchLeaderboard();
  }, []);

  const fetchLeaderboard = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/leaderboard');
      
      if (!response.ok) {
        throw new Error('Failed to fetch leaderboard');
      }
      
      const data = await response.json();
      setLeaderboard(Array.isArray(data) ? data : []);
      setError(null);
    } catch (err) {
      setError(err.message);
      console.error('Error fetching leaderboard:', err);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString();
  };

  if (loading) {
    return (
      <div className="leaderboard">
        <h3>🏆 Leaderboard</h3>
        <div className="loading">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="leaderboard">
        <h3>🏆 Leaderboard</h3>
        <div className="error">
          Error: {error}
          <button onClick={fetchLeaderboard} className="retry-btn">
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="leaderboard">
      <div className="leaderboard-header">
        <h3>🏆 Leaderboard</h3>
        <button onClick={fetchLeaderboard} className="refresh-btn">
          🔄 Refresh
        </button>
      </div>
      
      {!leaderboard || leaderboard.length === 0 ? (
        <div className="no-data">
          No games played yet. Be the first to play!
        </div>
      ) : (
        <div className="leaderboard-table">
          <div className="table-header">
            <div className="rank">Rank</div>
            <div className="username">Player</div>
            <div className="wins">Wins</div>
            <div className="games">Games</div>
            <div className="winrate">Win Rate</div>
            <div className="last-played">Last Played</div>
          </div>
          
          {leaderboard && leaderboard.map((player, index) => (
            <div key={player.username} className="table-row">
              <div className="rank">
                {index === 0 && '🥇'}
                {index === 1 && '🥈'}
                {index === 2 && '🥉'}
                {index > 2 && `#${index + 1}`}
              </div>
              <div className="username">{player.username}</div>
              <div className="wins">{player.games_won}</div>
              <div className="games">{player.games_played}</div>
              <div className="winrate">{player.win_rate}%</div>
              <div className="last-played">{formatDate(player.last_played)}</div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default Leaderboard;