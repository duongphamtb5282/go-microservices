-- Migration: create_permissions_table
-- Description: Create permissions table for fine-grained access control

-- +++++ UP
-- Create permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(50) NOT NULL,
    updated_by VARCHAR(50) NOT NULL
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_permission_name ON permissions(name);
CREATE INDEX IF NOT EXISTS idx_permission_resource ON permissions(resource);
CREATE INDEX IF NOT EXISTS idx_permission_action ON permissions(action);
CREATE INDEX IF NOT EXISTS idx_permission_is_active ON permissions(is_active);

-- +++++ DOWN
-- Drop permissions table and indexes
DROP INDEX IF EXISTS idx_permission_is_active;
DROP INDEX IF EXISTS idx_permission_action;
DROP INDEX IF EXISTS idx_permission_resource;
DROP INDEX IF EXISTS idx_permission_name;
DROP TABLE IF EXISTS permissions;
