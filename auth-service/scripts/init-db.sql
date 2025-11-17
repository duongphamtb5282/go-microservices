-- Initialize keycloak database
-- This script is automatically run when PostgreSQL container starts

-- Create extensions for Keycloak
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Log completion
DO $$
BEGIN
   RAISE NOTICE 'Keycloak database initialization completed successfully';
END $$;
