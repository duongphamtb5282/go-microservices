-- Migration: create_user_events_table
-- Description: Create user_events table for tracking user-related events

-- +++++ UP
-- Create user_events table
CREATE TABLE IF NOT EXISTS user_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    email VARCHAR(255),
    username VARCHAR(100),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    service_name VARCHAR(100) NOT NULL,
    performed_by VARCHAR(255) NOT NULL,
    event_time TIMESTAMP NOT NULL DEFAULT NOW(),
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT user_events_event_type_check CHECK (event_type IN ('user_created', 'user_updated', 'user_deleted'))
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_user_events_user_id ON user_events(user_id);
CREATE INDEX IF NOT EXISTS idx_user_events_event_type ON user_events(event_type);
CREATE INDEX IF NOT EXISTS idx_user_events_service_name ON user_events(service_name);
CREATE INDEX IF NOT EXISTS idx_user_events_event_time ON user_events(event_time DESC);
CREATE INDEX IF NOT EXISTS idx_user_events_email ON user_events(email);
CREATE INDEX IF NOT EXISTS idx_user_events_composite ON user_events(user_id, event_type, event_time DESC);

-- Create index on metadata JSONB column for faster queries
CREATE INDEX IF NOT EXISTS idx_user_events_metadata ON user_events USING GIN (metadata);

-- +++++ DOWN
-- Drop user_events table and indexes
DROP INDEX IF EXISTS idx_user_events_metadata;
DROP INDEX IF EXISTS idx_user_events_composite;
DROP INDEX IF EXISTS idx_user_events_email;
DROP INDEX IF EXISTS idx_user_events_event_time;
DROP INDEX IF EXISTS idx_user_events_service_name;
DROP INDEX IF EXISTS idx_user_events_event_type;
DROP INDEX IF EXISTS idx_user_events_user_id;
DROP TABLE IF EXISTS user_events;

