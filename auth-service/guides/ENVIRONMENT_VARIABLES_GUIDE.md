# Environment Variables Guide

## Overview

This guide explains how to use `.env` files to manage environment variables instead of manually exporting them.

---

## Why Use .env Files?

### Problems with Manual Export

**Manual approach** (what you were doing):

```bash
export OTEL_ENABLED=true
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318
export DATABASE_USERNAME=auth_user
export DATABASE_PASSWORD=auth_password
# ... many more variables
```

**Problems**:

- ❌ Tedious to type every time
- ❌ Easy to forget variables
- ❌ Hard to switch between environments
- ❌ Variables lost when terminal closes
- ❌ Not reproducible across team members

### Benefits of .env Files

**.env approach**:

```bash
# One-time setup
cp .env.example .env

# Every time you run
source scripts/load-env.sh
./auth-service
```

Or even simpler:

```bash
./scripts/run.sh  # Automatically loads .env
```

**Benefits**:

- ✅ All variables in one file
- ✅ Easy to manage and update
- ✅ Consistent across sessions
- ✅ Easy environment switching (dev/staging/prod)
- ✅ Team can share `.env.example` template
- ✅ Secrets not in version control

---

## Quick Start

### Step 1: Create Your .env File

```bash
# Copy the template
cp .env.example .env

# Edit with your values (optional, defaults work for local dev)
nano .env
# or
code .env
```

### Step 2: Use It

#### Option A: Load Manually (if you want to run commands yourself)

```bash
# Load environment variables
source scripts/load-env.sh

# Now run any command with env vars loaded
./auth-service
go test ./...
```

#### Option B: Use Auto-Loading Scripts (Recommended)

```bash
# Scripts automatically load .env
./scripts/run.sh                    # Run service
./scripts/start-with-telemetry.sh   # Run with telemetry
./scripts/quick-setup.sh            # Setup and run
```

---

## File Structure

```
auth-service/
├── .env.example          # Template with all variables (committed to git)
├── .env                  # Your local config (gitignored, never commit!)
├── .env.development      # Development defaults (optional)
├── .env.production       # Production config (optional)
└── scripts/
    └── load-env.sh       # Script to load .env files
```

### .env.example

**What**: Template file with all available variables and documentation

**Purpose**:

- Documents all configuration options
- Provides sensible defaults
- Shared with team via git

**Usage**:

```bash
cp .env.example .env
```

### .env

**What**: Your personal local configuration

**Purpose**:

- Override defaults for your local setup
- Store sensitive values (passwords, secrets)
- **Never committed to git** (in .gitignore)

**Usage**: Edit directly as needed

### .env.development / .env.production

**What**: Environment-specific configurations

**Purpose**:

- Different settings for dev vs prod
- Can be committed (if no secrets) or templated

**Usage**:

```bash
source scripts/load-env.sh development
source scripts/load-env.sh production
```

---

## Loading Environment Variables

### Method 1: Auto-Load with Scripts (Easiest)

Scripts in `scripts/` directory automatically load `.env`:

```bash
./scripts/run.sh
./scripts/start-with-telemetry.sh
./scripts/quick-setup.sh
```

These scripts now include:

```bash
if [ -f .env ]; then
    source .env
fi
```

### Method 2: Manual Load

Load `.env` before running commands:

```bash
# Load default .env
source scripts/load-env.sh

# Load specific environment
source scripts/load-env.sh development
source scripts/load-env.sh production

# Then run your commands
./auth-service
go test ./...
```

### Method 3: One-Line Command

```bash
# Load and run in one command
env $(cat .env | grep -v '^#' | xargs) ./auth-service
```

### Method 4: Use direnv (Advanced)

Install `direnv`:

```bash
# macOS
brew install direnv

# Add to ~/.zshrc or ~/.bashrc
eval "$(direnv hook zsh)"
```

Create `.envrc`:

```bash
echo "dotenv" > .envrc
direnv allow
```

Now `.env` loads automatically when you `cd` into the directory!

---

## Available Variables

### Application

| Variable   | Default       | Description                                  |
| ---------- | ------------- | -------------------------------------------- |
| `APP_ENV`  | `development` | Environment (development/staging/production) |
| `APP_PORT` | `8085`        | HTTP server port                             |

### Database

| Variable            | Default         | Description                |
| ------------------- | --------------- | -------------------------- |
| `DATABASE_HOST`     | `localhost`     | PostgreSQL host            |
| `DATABASE_PORT`     | `5432`          | PostgreSQL port            |
| `DATABASE_NAME`     | `auth_service`  | Database name              |
| `DATABASE_USERNAME` | `auth_user`     | Database user              |
| `DATABASE_PASSWORD` | `auth_password` | Database password          |
| `DATABASE_SSLMODE`  | `disable`       | SSL mode (disable/require) |

### Redis

| Variable         | Default     | Description    |
| ---------------- | ----------- | -------------- |
| `REDIS_HOST`     | `localhost` | Redis host     |
| `REDIS_PORT`     | `6379`      | Redis port     |
| `REDIS_PASSWORD` | (empty)     | Redis password |

### Kafka

| Variable         | Default              | Description            |
| ---------------- | -------------------- | ---------------------- |
| `KAFKA_BROKERS`  | `localhost:9092`     | Kafka broker addresses |
| `KAFKA_TOPIC`    | `auth-events`        | Default topic          |
| `KAFKA_GROUP_ID` | `auth-service-group` | Consumer group ID      |

### JWT

| Variable       | Default              | Description        |
| -------------- | -------------------- | ------------------ |
| `JWT_SECRET`   | (from config)        | JWT signing secret |
| `JWT_EXPIRY`   | `24h`                | Token expiration   |
| `JWT_ISSUER`   | `auth-service`       | Token issuer       |
| `JWT_AUDIENCE` | `auth-service-users` | Token audience     |

### Authorization

| Variable             | Default | Description             |
| -------------------- | ------- | ----------------------- |
| `AUTHORIZATION_MODE` | `jwt`   | Mode: `jwt` or `pingam` |

### PingAM

| Variable               | Default                          | Description         |
| ---------------------- | -------------------------------- | ------------------- |
| `PINGAM_BASE_URL`      | `http://localhost:1080`          | PingAM server URL   |
| `PINGAM_CLIENT_ID`     | `auth-service-client`            | OAuth client ID     |
| `PINGAM_CLIENT_SECRET` | (required)                       | OAuth client secret |
| `PINGAM_REDIRECT_URI`  | `http://localhost:8085/callback` | OAuth redirect URI  |

### OpenTelemetry

| Variable                      | Default          | Description                |
| ----------------------------- | ---------------- | -------------------------- |
| `OTEL_ENABLED`                | `true`           | Enable telemetry           |
| `OTEL_SERVICE_NAME`           | `auth-service`   | Service name in traces     |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `localhost:4318` | OTLP endpoint (no http://) |
| `OTEL_ENVIRONMENT`            | `development`    | Environment tag            |

### Logging

| Variable     | Default | Description                       |
| ------------ | ------- | --------------------------------- |
| `LOG_LEVEL`  | `debug` | Log level (debug/info/warn/error) |
| `LOG_FORMAT` | `json`  | Log format (json/text)            |

---

## Environment-Specific Configurations

### Development

```bash
# Use development settings
source scripts/load-env.sh development

# Or create .env from development template
cp .env.example .env
# (already has development defaults)
```

**Typical dev settings**:

- `LOG_LEVEL=debug`
- `OTEL_ENABLED=true`
- `ENABLE_SWAGGER=true`
- `DATABASE_SSLMODE=disable`

### Staging

Create `.env.staging`:

```bash
APP_ENV=staging
DATABASE_HOST=staging-db.example.com
DATABASE_SSLMODE=require
LOG_LEVEL=info
OTEL_ENABLED=true
```

Load it:

```bash
source scripts/load-env.sh staging
```

### Production

Create `.env.production`:

```bash
APP_ENV=production
DATABASE_HOST=${DB_HOST}  # Reference external secrets
DATABASE_SSLMODE=require
LOG_LEVEL=warn
OTEL_ENABLED=true
ENABLE_SWAGGER=false
ENABLE_PPROF=false
```

Load it:

```bash
source scripts/load-env.sh production
```

---

## Best Practices

### 1. Never Commit .env

**Always gitignored**:

```gitignore
# .gitignore
.env
.env.local
.env.*.local
```

**DO commit**:

- ✅ `.env.example` (template)
- ✅ `.env.development` (if no secrets)
- ✅ `.env.production` (as template, use ${VAR} for secrets)

### 2. Document All Variables

In `.env.example`, document every variable:

```bash
# JWT Configuration
# The secret key used to sign JWT tokens
# IMPORTANT: Change this in production!
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Token expiration duration (e.g., 1h, 24h, 7d)
JWT_EXPIRY=24h
```

### 3. Use Sensible Defaults

Provide defaults that work for local development:

```bash
# Good: Works out of the box for local dev
DATABASE_HOST=localhost

# Bad: Requires user to know and set
# DATABASE_HOST=
```

### 4. Separate Secrets

**Development**: Okay to have fake secrets in `.env.example`

```bash
JWT_SECRET=dev-secret-not-for-production
```

**Production**: Use secret management

```bash
# Option 1: Reference from secret manager
JWT_SECRET=${SECRET_MANAGER_JWT_SECRET}

# Option 2: Load from separate file
JWT_SECRET=$(cat /run/secrets/jwt_secret)
```

### 5. Validate Required Variables

In your scripts, check for required variables:

```bash
if [ -z "$DATABASE_PASSWORD" ]; then
    echo "❌ Error: DATABASE_PASSWORD not set!"
    exit 1
fi
```

### 6. Use Environment-Specific Files

```bash
# Local development
cp .env.example .env

# CI/CD pipeline
cp .env.staging .env

# Production deployment
cp .env.production .env
```

---

## Comparison: Manual vs .env

### Before (Manual Export)

```bash
# Every terminal session:
export OTEL_ENABLED=true
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318
export DATABASE_USERNAME=auth_user
export DATABASE_PASSWORD=auth_password
export AUTHORIZATION_MODE=jwt
export LOG_LEVEL=debug
export JWT_SECRET=my-secret
export KAFKA_BROKERS=localhost:9092
# ... 20+ more variables

# Run service
./auth-service
```

**Problems**:

- 😫 Tedious every time
- 🤔 Easy to forget variables
- 😱 Lost when terminal closes
- 🐛 Hard to debug inconsistencies

### After (.env File)

```bash
# One-time setup:
cp .env.example .env

# Every session (automatic):
./scripts/run.sh
```

**Benefits**:

- 😊 Simple and fast
- ✅ Consistent every time
- 🎯 Easy to switch environments
- 🔒 Secure (not in shell history)

---

## Troubleshooting

### Variables Not Loading

**Problem**: Variables from `.env` not being set

**Solutions**:

1. **Check file exists**:

   ```bash
   ls -la .env
   ```

2. **Check file format** (no spaces around `=`):

   ```bash
   # Good
   DATABASE_HOST=localhost

   # Bad
   DATABASE_HOST = localhost
   ```

3. **Use `source` not `sh`**:

   ```bash
   # Correct
   source scripts/load-env.sh

   # Wrong (creates subshell, variables not exported)
   sh scripts/load-env.sh
   ```

4. **Check export**:
   ```bash
   source scripts/load-env.sh
   echo $DATABASE_HOST  # Should print value
   ```

### Variables Not Persisting

**Problem**: Variables disappear after closing terminal

**Solution**: This is expected! `.env` is session-specific. Either:

1. **Load in each session**:

   ```bash
   source scripts/load-env.sh
   ```

2. **Use direnv** (auto-loads):

   ```bash
   brew install direnv
   echo 'eval "$(direnv hook zsh)"' >> ~/.zshrc
   echo "dotenv" > .envrc
   direnv allow
   ```

3. **Use auto-loading scripts**:
   ```bash
   ./scripts/run.sh  # Loads .env automatically
   ```

### Wrong Environment Loaded

**Problem**: Production variables in development

**Solution**:

1. **Check which file is loaded**:

   ```bash
   echo $APP_ENV  # Should show current environment
   ```

2. **Explicitly load correct env**:

   ```bash
   source scripts/load-env.sh development
   ```

3. **Use separate .env files**:

   ```bash
   # Development
   ln -sf .env.development .env

   # Production
   ln -sf .env.production .env
   ```

---

## Integration with Scripts

### Update Existing Scripts

Add this to the beginning of your scripts:

```bash
#!/bin/bash

# Load .env if it exists
if [ -f .env ]; then
    set -a
    source .env
    set +a
fi

# Rest of your script
echo "Database: $DATABASE_HOST"
```

### Already Updated Scripts

These scripts now auto-load `.env`:

- ✅ `scripts/run.sh`
- ✅ `scripts/start-with-telemetry.sh`
- ✅ `scripts/restart-with-telemetry.sh`
- ✅ `scripts/quick-setup.sh`

---

## Examples

### Example 1: Switch to PingAM Mode

**Old way**:

```bash
export AUTHORIZATION_MODE=pingam
export PINGAM_BASE_URL=http://localhost:1080
export PINGAM_CLIENT_ID=my-client
export PINGAM_CLIENT_SECRET=my-secret
./auth-service
```

**New way**:

```bash
# Edit .env
echo "AUTHORIZATION_MODE=pingam" >> .env

# Run
./scripts/run.sh
```

### Example 2: Change Database

**Old way**:

```bash
export DATABASE_HOST=prod-db.example.com
export DATABASE_USERNAME=prod_user
export DATABASE_PASSWORD=prod_password
export DATABASE_SSLMODE=require
./auth-service
```

**New way**:

```bash
# Edit .env
cat >> .env << EOF
DATABASE_HOST=prod-db.example.com
DATABASE_USERNAME=prod_user
DATABASE_PASSWORD=prod_password
DATABASE_SSLMODE=require
EOF

# Run
./scripts/run.sh
```

### Example 3: Multiple Environments

**Setup**:

```bash
# Create environment files
cp .env.example .env.development
cp .env.example .env.staging
cp .env.example .env.production

# Edit each with appropriate values
nano .env.development
nano .env.staging
nano .env.production
```

**Use**:

```bash
# Development
source scripts/load-env.sh development
./auth-service

# Staging
source scripts/load-env.sh staging
./auth-service

# Production
source scripts/load-env.sh production
./auth-service
```

---

## Advanced: Secret Management

For production, don't store secrets in `.env`. Instead:

### Option 1: External Secret Manager

```bash
# .env.production
DATABASE_PASSWORD=$(aws secretsmanager get-secret-value --secret-id db-password --query SecretString --output text)
JWT_SECRET=$(vault kv get -field=value secret/jwt)
```

### Option 2: Docker Secrets

```yaml
# docker-compose.yml
services:
  auth-service:
    environment:
      - DATABASE_PASSWORD_FILE=/run/secrets/db_password
    secrets:
      - db_password

secrets:
  db_password:
    external: true
```

### Option 3: Kubernetes Secrets

```yaml
# deployment.yaml
env:
  - name: DATABASE_PASSWORD
    valueFrom:
      secretKeyRef:
        name: db-credentials
        key: password
```

---

## Quick Reference

### Commands

```bash
# Create .env from template
cp .env.example .env

# Load environment
source scripts/load-env.sh

# Load specific environment
source scripts/load-env.sh development
source scripts/load-env.sh production

# Run with auto-load
./scripts/run.sh

# Check loaded variables
env | grep -E "OTEL|DATABASE|JWT"
```

### File Locations

```
.env.example      # Template (commit to git)
.env              # Local config (gitignored)
.env.development  # Dev settings (optional)
.env.production   # Prod settings (optional)
scripts/load-env.sh  # Load script
```

---

## Summary

✅ **What Changed**:

- Created `.env.example` with all variables
- Created `.env` for local configuration
- Created `scripts/load-env.sh` to load environments
- Updated scripts to auto-load `.env`

✅ **How to Use**:

```bash
# One-time setup
cp .env.example .env

# Every time (automatic)
./scripts/run.sh
```

✅ **Benefits**:

- No more manual `export` commands
- Consistent configuration
- Easy environment switching
- Secure (not in shell history)
- Team-friendly (shared template)

🎯 **Result**: Simple, secure, and maintainable environment configuration!
