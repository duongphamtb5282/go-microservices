#!/bin/bash

# Production Deployment Script for Auth Service
# Usage: ./deploy.sh [environment] [action]

set -euo pipefail

# Configuration
ENVIRONMENT="${1:-production}"
ACTION="${2:-deploy}"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-your-registry.com}"
IMAGE_NAME="${DOCKER_REGISTRY}/auth-service"
NAMESPACE="${NAMESPACE:-auth-system}"

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

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Prerequisites check
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check if required tools are installed
    local tools=("docker" "kubectl" "helm")
    for tool in "${tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            log_error "Required tool '$tool' not found. Please install it first."
            exit 1
        fi
    done

    # Check if Docker daemon is running
    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running"
        exit 1
    fi

    # Check Kubernetes connection
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi

    log_success "Prerequisites check passed"
}

# Build and push Docker image
build_and_push_image() {
    local tag="${IMAGE_NAME}:$(date +%Y%m%d-%H%M%S)"
    local latest_tag="${IMAGE_NAME}:latest"

    log_info "Building Docker image..."
    docker build -t "$tag" -t "$latest_tag" .

    log_info "Scanning image for vulnerabilities..."
    if command -v trivy &> /dev/null; then
        trivy image --exit-code 1 --severity HIGH,CRITICAL "$tag" || {
            log_warn "Security vulnerabilities found, but continuing deployment"
        }
    fi

    log_info "Pushing Docker image..."
    docker push "$tag"
    docker push "$latest_tag"

    echo "$tag"
}

# Deploy to Kubernetes
deploy_to_kubernetes() {
    local image_tag="$1"

    log_info "Deploying to Kubernetes..."

    # Create namespace if it doesn't exist
    kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

    # Update image tag in deployment
    sed -i.bak "s|image: auth-service:latest|image: $image_tag|g" k8s/deployment.yaml

    # Apply Kubernetes manifests
    kubectl apply -f k8s/ -n "$NAMESPACE"

    # Wait for deployment to be ready
    log_info "Waiting for deployment to be ready..."
    kubectl rollout status deployment/auth-service -n "$NAMESPACE" --timeout=300s

    # Verify deployment
    local replicas
    replicas=$(kubectl get deployment auth-service -n "$NAMESPACE" -o jsonpath='{.status.readyReplicas}')
    if [ "$replicas" -gt 0 ]; then
        log_success "Deployment completed successfully"
        log_info "Ready replicas: $replicas"
    else
        log_error "Deployment failed - no ready replicas"
        exit 1
    fi

    # Restore original deployment file
    mv k8s/deployment.yaml.bak k8s/deployment.yaml
}

# Run health checks
run_health_checks() {
    log_info "Running health checks..."

    # Check service health
    local health_url="https://auth.yourcompany.com/health"
    if curl -f -k --max-time 10 "$health_url" &> /dev/null; then
        log_success "Service health check passed"
    else
        log_error "Service health check failed"
        exit 1
    fi

    # Check metrics endpoint
    local metrics_url="https://auth.yourcompany.com/metrics"
    if curl -f -k --max-time 10 "$metrics_url" &> /dev/null; then
        log_success "Metrics endpoint accessible"
    else
        log_warn "Metrics endpoint not accessible (this may be expected if behind VPN)"
    fi
}

# Rollback deployment
rollback_deployment() {
    log_info "Rolling back deployment..."

    kubectl rollout undo deployment/auth-service -n "$NAMESPACE"

    # Wait for rollback to complete
    kubectl rollout status deployment/auth-service -n "$NAMESPACE" --timeout=300s

    log_success "Rollback completed"
}

# Scale deployment
scale_deployment() {
    local replicas="${1:-3}"

    log_info "Scaling deployment to $replicas replicas..."

    kubectl scale deployment auth-service --replicas="$replicas" -n "$NAMESPACE"

    # Wait for scaling to complete
    kubectl rollout status deployment/auth-service -n "$NAMESPACE" --timeout=300s

    log_success "Scaling completed"
}

# Get deployment status
get_status() {
    log_info "Getting deployment status..."

    echo "=== Kubernetes Resources ==="
    kubectl get all,ingress,svc,hpa -n "$NAMESPACE"

    echo ""
    echo "=== Pod Status ==="
    kubectl get pods -n "$NAMESPACE" -o wide

    echo ""
    echo "=== Recent Events ==="
    kubectl get events -n "$NAMESPACE" --sort-by=.metadata.creationTimestamp | tail -10

    echo ""
    echo "=== Resource Usage ==="
    kubectl top pods -n "$NAMESPACE" 2>/dev/null || echo "Metrics server not available"
}

# Main deployment function
main() {
    log_info "Starting deployment for environment: $ENVIRONMENT"
    log_info "Action: $ACTION"

    case "$ACTION" in
        "deploy")
            check_prerequisites
            image_tag=$(build_and_push_image)
            deploy_to_kubernetes "$image_tag"
            run_health_checks
            log_success "Deployment completed successfully!"
            ;;
        "rollback")
            check_prerequisites
            rollback_deployment
            ;;
        "scale")
            check_prerequisites
            scale_deployment "${3:-3}"
            ;;
        "status")
            check_prerequisites
            get_status
            ;;
        "build")
            check_prerequisites
            image_tag=$(build_and_push_image)
            log_success "Build completed: $image_tag"
            ;;
        *)
            log_error "Invalid action: $ACTION"
            echo "Usage: $0 [environment] [action]"
            echo "Actions: deploy, rollback, scale, status, build"
            echo "Example: $0 production deploy"
            exit 1
            ;;
    esac
}

# Handle script interruption
trap 'log_error "Deployment interrupted by user"; exit 1' INT TERM

# Run main function with all arguments
main "$@"
