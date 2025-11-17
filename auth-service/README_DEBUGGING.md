# VSCode Debugging and Hot Reload Setup

This guide explains how to set up debugging and hot reload for the Auth Service in VSCode.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Initial Setup](#initial-setup)
3. [VSCode Configuration](#vscode-configuration)
4. [Debugging](#debugging)
5. [Hot Reload](#hot-reload)
6. [Development Workflow](#development-workflow)
7. [Troubleshooting](#troubleshooting)

## Prerequisites

- **Go 1.24+** installed
- **VSCode** with Go extension
- **Git** for version control

### Required VSCode Extensions

Install these extensions from the VSCode marketplace:

- **Go** (by Google) - Core Go language support
- **Go Nightly** (optional) - Latest Go features

## Initial Setup

### 1. Install Development Tools

```bash
# Install all development tools at once
make dev-tools

# Or install individually
go install github.com/air-verse/air@latest              # Hot reload
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest  # Linting
go install github.com/go-delve/delve/cmd/dlv@latest     # Debugging
```

### 2. Setup Development Environment

```bash
# Setup complete development environment
make dev-setup

# This will:
# - Install development tools
# - Create .env file from template
# - Provide setup instructions
```

### 3. Configure Environment

```bash
# Copy environment template
cp env.example .env

# Generate secure secrets
make generate-secret

# Edit .env with your local configuration
# Update database URLs, ports, etc.
```

## VSCode Configuration

The following VSCode configuration files are already set up:

- `.vscode/launch.json` - Debug configurations
- `.vscode/tasks.json` - Build and run tasks
- `.vscode/settings.json` - Go-specific settings

### Launch Configurations

Available debug configurations:

1. **Debug Auth Service** - Debug the main application
2. **Debug Auth Service (with env file)** - Debug with environment variables from `.env`
3. **Debug Migration** - Debug database migrations
4. **Debug Secret Generator** - Debug the secret generation tool
5. **Debug Tests** - Debug unit tests
6. **Attach to Process** - Attach debugger to running process

## Debugging

### Start Debugging

1. **Open the project in VSCode**
2. **Set breakpoints** by clicking in the gutter next to line numbers
3. **Press F5** or go to Run → Start Debugging
4. **Select configuration** from the dropdown (usually "Debug Auth Service")

### Debug with Environment File

For debugging with your `.env` file:

1. Select "Debug Auth Service (with env file)" configuration
2. Ensure your `.env` file exists and is configured
3. Press F5 to start debugging

### Debug Features

- **Step Over (F10)** - Execute current line and move to next
- **Step Into (F11)** - Step into function calls
- **Step Out (Shift+F11)** - Step out of current function
- **Continue (F5)** - Continue execution until next breakpoint
- **Stop (Shift+F5)** - Stop debugging

### Debug Console

Use the Debug Console to:
- Evaluate expressions: `variableName`
- Execute Go code: `fmt.Println("debug")`
- Inspect variables and call stack

## Hot Reload

Hot reload automatically rebuilds and restarts the application when files change.

### Start Hot Reload

```bash
# Basic hot reload
make dev-hot

# With debug logging
make dev-debug

# With verbose output
make dev-verbose
```

### Hot Reload Features

- **Automatic rebuild** on file changes
- **Fast restart** - only rebuilds changed files
- **Preserves state** - keeps database connections, etc.
- **Error handling** - stops on build errors

### Configuration

Hot reload is configured in `.air.toml`:

```toml
[build]
  cmd = "go build -o ./tmp/main ./cmd/http/main.go"
  bin = "./tmp/main"
  include_ext = ["go", "tpl", "tmpl", "html"]
  exclude_dir = ["tmp", "vendor", "testdata"]
  delay = 1000  # Delay before rebuild (ms)
```

## Development Workflow

### Recommended Development Setup

1. **Start dependencies** (database, Redis, etc.)
   ```bash
   make docker-up
   ```

2. **Start hot reload** in one terminal
   ```bash
   make dev-hot
   ```

3. **Debug in VSCode** in another terminal/session
   - Open VSCode
   - Set breakpoints
   - Press F5 to start debugging

4. **Make changes** - hot reload will automatically rebuild
5. **Debug issues** - use breakpoints and debug console

### Full Development Command

```bash
# Terminal 1: Start infrastructure
make docker-up

# Terminal 2: Start hot reload
make dev-hot

# VSCode: Debug with F5
# Make changes, see them reload automatically
```

## Testing

### Debug Tests

1. Open a test file
2. Click the "Debug Test" above a test function
3. Or use the "Debug Tests" launch configuration

### Run Tests with Coverage

```bash
make test-coverage
# Opens coverage report in browser
```

## Troubleshooting

### Common Issues

#### Hot Reload Not Working

```bash
# Check Air is installed
air -v

# Check configuration
air -c .air.toml

# Manual rebuild
make build
```

#### Debug Not Starting

```bash
# Check Delve is installed
dlv version

# Check Go version
go version

# Clean and rebuild
make clean
make build-debug
```

#### Breakpoints Not Hit

- Ensure you're running the debug configuration, not regular run
- Check the binary path in launch.json matches your build
- Try rebuilding: `make build-debug`

#### Environment Variables Not Loading

```bash
# Check .env file exists
ls -la .env

# Validate .env syntax
cat .env | grep -v '^#' | grep '='

# Use absolute paths if needed
```

#### Port Conflicts

```bash
# Check what's using the port
lsof -i :8085

# Change port in .env
echo "SERVER_PORT=8086" >> .env
```

### Debug Logs

Enable debug logging for troubleshooting:

```bash
# In .env
LOG_LEVEL=debug
LOG_FORMAT=text

# Or run with debug
make dev-debug
```

### Performance Issues

If hot reload is slow:

```bash
# Increase delay in .air.toml
delay = 2000

# Exclude more directories
exclude_dir = ["node_modules", "tmp", "vendor", "assets"]
```

## Advanced Configuration

### Custom Launch Configuration

Add to `.vscode/launch.json`:

```json
{
    "name": "Debug with Custom Config",
    "type": "go",
    "request": "launch",
    "mode": "debug",
    "program": "${workspaceFolder}/cmd/http/main.go",
    "env": {
        "APP_ENV": "development",
        "LOG_LEVEL": "trace"
    },
    "args": ["--verbose"],
    "showLog": true
}
```

### Custom Air Configuration

Modify `.air.toml` for your needs:

```toml
[build]
  # Custom build command
  cmd = "go build -tags=debug -o ./tmp/main ./cmd/http/main.go"

  # Watch additional file types
  include_ext = ["go", "yaml", "yml", "json"]

  # Custom delay
  delay = 500
```

### Remote Debugging

For debugging in containers or remote servers:

```json
{
    "name": "Remote Debug",
    "type": "go",
    "request": "attach",
    "mode": "remote",
    "host": "localhost",
    "port": 40000
}
```

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| F5 | Start/Continue debugging |
| F10 | Step over |
| F11 | Step into |
| Shift+F11 | Step out |
| Shift+F5 | Stop debugging |
| Ctrl+Shift+D | Open debug panel |
| Ctrl+Shift+Y | Open debug console |

## Resources

- [Go VSCode Extension](https://marketplace.visualstudio.com/items?itemName=golang.Go)
- [Air Documentation](https://github.com/air-verse/air)
- [Delve Debugger](https://github.com/go-delve/delve)
- [Go Debugging Guide](https://golang.org/doc/gdb)

---

Happy debugging! 🐛✨
