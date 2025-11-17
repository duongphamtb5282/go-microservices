-- Migration: remove_modified_columns
-- Description: Remove modified_by and modified_at columns from users table

-- +++++ UP
-- Drop the modified_by column
ALTER TABLE users DROP COLUMN IF EXISTS modified_by;

-- Drop the modified_at column
ALTER TABLE users DROP COLUMN IF EXISTS modified_at;

-- +++++ DOWN
-- Add back the modified_by column
ALTER TABLE users ADD COLUMN IF NOT EXISTS modified_by VARCHAR(50);

-- Add back the modified_at column
ALTER TABLE users ADD COLUMN IF NOT EXISTS modified_at TIMESTAMP;
