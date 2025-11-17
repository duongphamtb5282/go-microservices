-- Migration: create_roles_table
-- Description: Create roles table for role-based access control

-- +++++ UP
-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(50) NOT NULL,
    updated_by VARCHAR(50) NOT NULL
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_role_name ON roles(name);
CREATE INDEX IF NOT EXISTS idx_role_is_active ON roles(is_active);

-- +++++ DOWN
-- Drop roles table and indexes
DROP INDEX IF EXISTS idx_role_is_active;
DROP INDEX IF EXISTS idx_role_name;
DROP TABLE IF EXISTS roles;
