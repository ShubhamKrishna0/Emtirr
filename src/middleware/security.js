// Input validation
const validateUsername = (username) => {
  if (!username || typeof username !== 'string') {
    return { valid: false, error: 'Username is required' };
  }
  
  const trimmed = username.trim();
  if (trimmed.length < 2 || trimmed.length > 20) {
    return { valid: false, error: 'Username must be 2-20 characters' };
  }
  
  if (!/^[a-zA-Z0-9_-]+$/.test(trimmed)) {
    return { valid: false, error: 'Username can only contain letters, numbers, _ and -' };
  }
  
  return { valid: true, username: trimmed };
};

const validateMove = (column) => {
  const col = parseInt(column);
  if (isNaN(col) || col < 0 || col > 6) {
    return { valid: false, error: 'Invalid column' };
  }
  return { valid: true, column: col };
};

module.exports = {
  validateUsername,
  validateMove
};