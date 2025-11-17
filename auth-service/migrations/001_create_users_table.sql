-- Migration: create_users_table
-- Description: Create users table with proper indexes

-- +++++ UP
-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(20) NOT NULL UNIQUE,
    email VARCHAR(254) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_at TIMESTAMP,
    login_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(50) NOT NULL,
    updated_by VARCHAR(50) NOT NULL
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_user_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_user_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_user_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_user_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_user_is_verified ON users(is_verified);

-- +++++ DOWN
-- Drop users table and indexes
DROP INDEX IF EXISTS idx_user_is_verified;
DROP INDEX IF EXISTS idx_user_is_active;
DROP INDEX IF EXISTS idx_user_created_at;
DROP INDEX IF EXISTS idx_user_username;
DROP INDEX IF EXISTS idx_user_email;
DROP TABLE IF EXISTS users;