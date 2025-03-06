-- Create table to store hashed passwords
CREATE TABLE IF NOT EXISTS passwords (
    sha1 TEXT PRIMARY KEY
);

-- Create table to store API keys
CREATE TABLE IF NOT EXISTS api_keys (
    key_id TEXT PRIMARY KEY,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
