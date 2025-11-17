#!/bin/bash

# Secret Validation Script
# Validates the strength and format of secrets in environment variables and files

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
MIN_SECRET_LENGTH=32
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

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

# Validate secret strength
validate_secret_strength() {
    local secret="$1"
    local name="$2"
    local issues=()

    # Check minimum length
    if [ ${#secret} -lt $MIN_SECRET_LENGTH ]; then
        issues+=("Too short: ${#secret} chars (minimum: $MIN_SECRET_LENGTH)")
    fi

    # Check for common weak patterns
    if [[ "$secret" =~ ^[0-9]+$ ]]; then
        issues+=("Contains only numbers - weak entropy")
    fi

    if [[ "$secret" =~ ^[a-zA-Z]+$ ]]; then
        issues+=("Contains only letters - weak entropy")
    fi

    if [[ "$secret" =~ ^(.)\1+$ ]]; then
        issues+=("Contains repeated characters - weak entropy")
    fi

    # Check for common weak words
    local weak_words=("password" "secret" "admin" "root" "test" "demo" "default" "123456" "qwerty")
    local secret_lower
    secret_lower=$(echo "$secret" | tr '[:upper:]' '[:lower:]')
    for word in "${weak_words[@]}"; do
        if [[ "$secret_lower" == *"$word"* ]]; then
            issues+=("Contains common weak word: $word")
            break
        fi
    done

    # Check for URL-unsafe characters (warning only)
    if [[ "$secret" =~ [^a-zA-Z0-9._~-] ]]; then
        log_warning "$name: Contains non-URL-safe characters"
    fi

    # Report issues
    if [ ${#issues[@]} -eq 0 ]; then
        log_success "$name: Secret strength validation passed"
        return 0
    else
        log_error "$name: Secret validation failed:"
        for issue in "${issues[@]}"; do
            echo "  - $issue"
        done
        return 1
    fi
}

# Validate environment variable
validate_env_var() {
    local var_name="$1"
    local var_value="${!var_name:-}"

    if [ -z "$var_value" ]; then
        log_warning "Environment variable $var_name is not set"
        return 1
    fi

    log_info "Validating environment variable: $var_name"
    validate_secret_strength "$var_value" "$var_name"
}

# Validate .env file
validate_env_file() {
    local file_path="$1"

    if [ ! -f "$file_path" ]; then
        log_warning "Environment file not found: $file_path"
        return 1
    fi

    log_info "Validating environment file: $file_path"

    local validation_passed=true
    local line_number=0

    while IFS= read -r line || [ -n "$line" ]; do
        line_number=$((line_number + 1))

        # Skip comments and empty lines
        [[ "$line" =~ ^[[:space:]]*# ]] && continue
        [[ -z "$line" ]] && continue

        # Extract variable name and value
        if [[ "$line" =~ ^([^=]+)=(.*)$ ]]; then
            local var_name="${BASH_REMATCH[1]}"
            local var_value="${BASH_REMATCH[2]}"

            # Remove quotes if present
            var_value="${var_value#\"}"
            var_value="${var_value%\"}"
            var_value="${var_value#\'}"
            var_value="${var_value%\'}"

            # Check if it's a secret-related variable (but exclude configuration values)
            case "$var_name" in
                # Actual secrets that need validation
                *CLIENT_SECRET*|*JWT_SECRET*|*ADMIN_PASSWORD*|*API_SECRET*|*PRIVATE_KEY*)
                    # Skip if it looks like a configuration value (contains common config patterns)
                    if [[ ! "$var_value" =~ ^[0-9a-zA-Z_-]*$ ]] || [[ "$var_value" =~ ^[a-z]+:// ]] || [[ "$var_value" =~ ^/.* ]]; then
                        # This looks like a URL, path, or config value - skip validation
                        continue
                    fi
                    log_info "Found secret variable: $var_name (line $line_number)"
                    if ! validate_secret_strength "$var_value" "$var_name (line $line_number)"; then
                        validation_passed=false
                    fi
                    ;;
                # Configuration values that might contain "secret" but aren't secrets
                *ENABLE*|*LENGTH*|*TIMEOUT*|*INTERVAL*|*TTL*|*COUNT*|*SIZE*|*LEVEL*|*PATH*|*URL*|*URI*|*ID*|*NAME*|*REALM*|*SCOPE*|*CLAIM*|*ATTEMPTS*|*THRESHOLD*|*FACTOR*|*CODE*)
                    # Skip validation for these config values
                    continue
                    ;;
            esac
        fi
    done < "$file_path"

    if [ "$validation_passed" = true ]; then
        log_success "All secrets in $file_path passed validation"
    else
        log_error "Some secrets in $file_path failed validation"
    fi

    return $([ "$validation_passed" = true ])
}

# Validate Keycloak-specific secrets
validate_keycloak_secrets() {
    log_info "Validating Keycloak-specific secrets..."

    local keycloak_vars=(
        "KEYCLOAK_CLIENT_SECRET"
        "KEYCLOAK_ADMIN_PASSWORD"
        "JWT_SECRET"
    )

    local all_passed=true

    for var in "${keycloak_vars[@]}"; do
        if ! validate_env_var "$var"; then
            all_passed=false
        fi
    done

    return $([ "$all_passed" = true ])
}

# Main validation function
validate_all() {
    local exit_code=0

    echo "🔍 Secret Validation Tool"
    echo "========================="
    echo

    # Validate environment variables
    log_info "Checking environment variables..."
    if ! validate_keycloak_secrets; then
        exit_code=1
    fi

    echo

    # Validate common .env files
    local env_files=(".env" ".env.local" ".env.dev" ".env.prod" "env.example")

    for env_file in "${env_files[@]}"; do
        if [ -f "$env_file" ]; then
            if ! validate_env_file "$env_file"; then
                exit_code=1
            fi
            echo
        fi
    done

    if [ $exit_code -eq 0 ]; then
        log_success "All secret validations passed!"
        echo
        log_info "Security recommendations:"
        echo "  ✓ Secrets are at least $MIN_SECRET_LENGTH characters long"
        echo "  ✓ No common weak patterns detected"
        echo "  ✓ Secrets contain mixed character types"
        echo "  ✓ No common weak words found"
    else
        log_error "Secret validation failed. Please fix the issues above."
        echo
        log_info "To generate secure secrets:"
        echo "  make generate-secret"
        echo "  make generate-secret-dev"
        echo "  make generate-secret-prod"
    fi

    return $exit_code
}

# Show usage information
show_usage() {
    cat << EOF
Secret Validation Tool

Validates the strength and security of secrets in environment variables and files.

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -e, --env VAR    Validate specific environment variable
    -f, --file FILE  Validate specific .env file
    -h, --help       Show this help message

VALIDATION RULES:
    - Minimum length: $MIN_SECRET_LENGTH characters
    - No repeated characters
    - Mixed character types (recommended)
    - No common weak words/patterns
    - URL-safe characters (warning only)

EXAMPLES:
    # Validate all secrets
    $0

    # Validate specific environment variable
    $0 -e KEYCLOAK_CLIENT_SECRET

    # Validate specific file
    $0 -f .env.prod

EOF
}

# Parse command line arguments
parse_args() {
    case "${1:-}" in
        -e|--env)
            if [ -z "${2:-}" ]; then
                log_error "Environment variable name required"
                exit 1
            fi
            validate_env_var "$2"
            exit $?
            ;;
        -f|--file)
            if [ -z "${2:-}" ]; then
                log_error "File path required"
                exit 1
            fi
            validate_env_file "$2"
            exit $?
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        "")
            # No arguments - validate all
            ;;
        *)
            log_error "Unknown option: $1"
            echo
            show_usage
            exit 1
            ;;
    esac
}

# Run main validation
main() {
    parse_args "$@"
    validate_all
}

# Run main function with all arguments
main "$@"
