# Secure Secret Generation for Keycloak

This directory contains tools for generating and validating cryptographically secure client secrets for Keycloak integration.

## Quick Start

### Generate a Keycloak Client Secret

```bash
# Generate a secure secret for production
make generate-secret

# Generate secrets for specific environments
make generate-secret-dev    # Saves to .env.dev
make generate-secret-prod   # Saves to .env.prod

# Validate existing secrets
make validate-secrets
```

### Manual Generation

```bash
# Generate with custom options
./scripts/generate-keycloak-secret.sh -l 64 -F hex -e MY_SECRET

# Save to file
./scripts/generate-keycloak-secret.sh -f .env.production

# Quiet mode for scripts
./scripts/generate-keycloak-secret.sh -q -e KEYCLOAK_CLIENT_SECRET
```

## Security Features

### Cryptographic Security
- Uses Go's `crypto/rand` for cryptographically secure randomness
- 256+ bits of entropy (32+ bytes minimum)
- Multiple output formats: base64url (recommended), base64, hex

### Validation Checks
- Minimum length validation (32 bytes)
- Entropy analysis (mixed character types)
- Common weak pattern detection
- URL-safe character validation

## Files

- `generate-keycloak-secret.sh` - Main secret generation script
- `validate-secrets.sh` - Secret strength validation tool
- `../cmd/secret-generator/main.go` - Go program for secure secret generation

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `KEYCLOAK_SECRET_LENGTH` | Secret length in bytes | 32 |
| `KEYCLOAK_SECRET_FORMAT` | Output format (base64, base64url, hex) | base64url |

## Usage Examples

### Basic Generation
```bash
$ make generate-secret
🔐 Keycloak Client Secret Generator
===================================

[INFO] Generating Keycloak client secret...
[INFO] Length: 32 bytes
[INFO] Format: base64url
export KEYCLOAK_CLIENT_SECRET=_xmnlQRhD2NMU85eKKNlV2OsIvkvqiOcXDWjwtPUCjw=
[SUCCESS] Keycloak client secret generated successfully!
```

### Environment-Specific
```bash
$ make generate-secret-dev
[INFO] Generating development Keycloak client secret...
export KEYCLOAK_CLIENT_SECRET_DEV=generated_secret_here
[SUCCESS] Secret saved to .env.dev
```

### Validation
```bash
$ make validate-secrets
🔍 Secret Validation Tool
=========================

[INFO] Checking environment variables...
[SUCCESS] KEYCLOAK_CLIENT_SECRET: Secret strength validation passed
[SUCCESS] All secret validations passed!
```

## Security Best Practices

1. **Never commit secrets** to version control
2. **Use different secrets** for each environment (dev/staging/prod)
3. **Rotate secrets regularly** (every 90-180 days in production)
4. **Store securely** using:
   - Environment variables
   - Secret management systems (AWS Secrets Manager, HashiCorp Vault, etc.)
   - Kubernetes secrets
5. **Validate regularly** using the validation tools

## Integration with CI/CD

### GitHub Actions Example
```yaml
- name: Generate Keycloak Secret
  run: |
    make generate-secret
    # Export the generated secret to environment
    echo "KEYCLOAK_CLIENT_SECRET=$(make generate-secret | grep export | cut -d'=' -f2)" >> $GITHUB_ENV
```

### Docker Build
```dockerfile
# Generate secret during build
RUN make generate-secret > /tmp/secret.env
# Copy to final image
COPY --from=builder /tmp/secret.env /app/.env
```

## Troubleshooting

### Common Issues

**"Go not found"**
```bash
# Install Go or ensure it's in PATH
which go
go version
```

**"Build failed"**
```bash
# Clean and rebuild
make clean
make generate-secret
```

**"Secret validation failed"**
```bash
# Generate a new secure secret
make generate-secret
# Or validate specific file
make validate-secrets ARGS="-f .env.production"
```

## Advanced Usage

### Custom Secret Generator
```go
// Use the secret generator programmatically
generator := secret.NewSecretGenerator(64, "hex")
secret, err := generator.GenerateSecureSecret()
```

### Integration with Secret Managers
```bash
# AWS Secrets Manager
aws secretsmanager create-secret \
  --name "prod/keycloak/client-secret" \
  --secret-string "$(make generate-secret | grep export | cut -d'=' -f2)"

# HashiCorp Vault
vault kv put secret/keycloak client_secret="$(make generate-secret | grep export | cut -d'=' -f2)"
```
