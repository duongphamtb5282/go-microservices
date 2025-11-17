#!/bin/bash

# Keycloak Client Secret Generator Script
# Generates cryptographically secure client secrets for Keycloak integration

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SECRET_GENERATOR="$PROJECT_ROOT/cmd/secret-generator"
BINARY_NAME="secret-generator"
BUILD_DIR="$PROJECT_ROOT/bin"
SECRET_LENGTH=${KEYCLOAK_SECRET_LENGTH:-32}
SECRET_FORMAT=${KEYCLOAK_SECRET_FORMAT:-base64url}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.19+ to use this script."
        exit 1
    fi

    # Check Go version
    GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    if ! awk -v ver="$GO_VERSION" 'BEGIN { if (ver < 1.19) exit 1; }'; then
        log_warning "Go version $GO_VERSION detected. Recommended version is 1.19+"
    fi
}

# Build the secret generator
build_secret_generator() {
    log_info "Building secret generator..."

    cd "$PROJECT_ROOT"
    mkdir -p "$BUILD_DIR"

    if ! go build -o "$BUILD_DIR/$BINARY_NAME" "$SECRET_GENERATOR"; then
        log_error "Failed to build secret generator"
        exit 1
    fi

    log_success "Secret generator built successfully"
}

# Generate the secret
generate_secret() {
    local env_var="${1:-KEYCLOAK_CLIENT_SECRET}"
    local output_file="${2:-}"

    log_info "Generating Keycloak client secret..."
    log_info "Length: $SECRET_LENGTH bytes"
    log_info "Format: $SECRET_FORMAT"

    # Run the secret generator
    local secret_output
    secret_output=$("$BUILD_DIR/$BINARY_NAME" -length "$SECRET_LENGTH" -format "$SECRET_FORMAT" -env "$env_var" -quiet)

    if [ $? -ne 0 ]; then
        log_error "Failed to generate secret"
        exit 1
    fi

    # Output the secret
    echo "$secret_output"

    # Save to file if specified
    if [ -n "$output_file" ]; then
        echo "$secret_output" > "$output_file"
        log_success "Secret saved to $output_file"
        log_info "Add the following line to your .env file:"
        echo "  $secret_output"
    fi

    log_success "Keycloak client secret generated successfully!"
    echo
    log_info "Security recommendations:"
    echo "  - Store this secret securely (environment variables or secret management)"
    echo "  - Never commit secrets to version control"
    echo "  - Rotate secrets regularly (every 90-180 days in production)"
    echo "  - Use different secrets for different environments"
}

# Show usage information
show_usage() {
    cat << EOF
Keycloak Client Secret Generator

This script generates cryptographically secure client secrets for Keycloak integration.

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -e, --env VAR       Environment variable name (default: KEYCLOAK_CLIENT_SECRET)
    -f, --file FILE     Save secret to file
    -l, --length BYTES  Secret length in bytes (default: 32, min: 32)
    -F, --format FORMAT Output format: base64, base64url, hex (default: base64url)
    -h, --help          Show this help message
    --no-build         Skip building the secret generator (use existing binary)

ENVIRONMENT VARIABLES:
    KEYCLOAK_SECRET_LENGTH    Secret length in bytes (default: 32)
    KEYCLOAK_SECRET_FORMAT    Output format (default: base64url)

EXAMPLES:
    # Generate default secret
    $0

    # Generate secret for specific environment variable
    $0 -e MY_CUSTOM_SECRET

    # Generate longer secret and save to file
    $0 -l 64 -f .env.local

    # Generate hex format secret
    $0 -F hex

SECURITY NOTES:
    - Uses crypto/rand for cryptographically secure randomness
    - Minimum 32 bytes (256 bits) recommended for OAuth2 clients
    - Base64url format is recommended for URL safety

EOF
}

# Parse command line arguments
parse_args() {
    ENV_VAR="KEYCLOAK_CLIENT_SECRET"
    OUTPUT_FILE=""
    SKIP_BUILD=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--env)
                ENV_VAR="$2"
                shift 2
                ;;
            -f|--file)
                OUTPUT_FILE="$2"
                shift 2
                ;;
            -l|--length)
                SECRET_LENGTH="$2"
                shift 2
                ;;
            -F|--format)
                SECRET_FORMAT="$2"
                shift 2
                ;;
            --no-build)
                SKIP_BUILD=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                echo
                show_usage
                exit 1
                ;;
        esac
    done

    # Validate inputs
    if [ "$SECRET_LENGTH" -lt 32 ]; then
        log_error "Secret length must be at least 32 bytes for security"
        exit 1
    fi

    case "$SECRET_FORMAT" in
        base64|base64url|hex) ;;
        *)
            log_error "Invalid format. Must be: base64, base64url, or hex"
            exit 1
            ;;
    esac
}

# Main function
main() {
    echo "🔐 Keycloak Client Secret Generator"
    echo "==================================="
    echo

    parse_args "$@"
    check_go

    if [ "$SKIP_BUILD" = false ]; then
        build_secret_generator
    else
        if [ ! -f "$BUILD_DIR/$BINARY_NAME" ]; then
            log_error "Secret generator binary not found. Run without --no-build first."
            exit 1
        fi
    fi

    generate_secret "$ENV_VAR" "$OUTPUT_FILE"
}

# Run main function with all arguments
main "$@"
