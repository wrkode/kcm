#!/bin/bash

# Script to build and push KCM images to your own GHCR registry
# This script helps you create all the container images required for KCM to work
# in your own GitHub Container Registry instead of k0rdent's

set -e

# Configuration
REGISTRY="ghcr.io/wrkode/kcm"
VERSION=${VERSION:-$(git describe --tags --always --dirty)}
CHART_VERSION=${CHART_VERSION:-$VERSION}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if docker is available
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    # Check if make is available
    if ! command -v make &> /dev/null; then
        print_error "Make is not installed or not in PATH"
        exit 1
    fi
    
    # Check if git is available
    if ! command -v git &> /dev/null; then
        print_error "Git is not installed or not in PATH"
        exit 1
    fi
    
    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Not in a git repository"
        exit 1
    fi
    
    # Check if curl is available (for token validation)
    if ! command -v curl &> /dev/null; then
        print_warning "curl is not available, skipping token validation"
    fi
    
    print_success "Prerequisites check passed"
}

# Clean up existing local images
cleanup_local_images() {
    print_status "Cleaning up existing local images..."
    
    # Remove any existing localhost/kcm/controller images
    if docker images | grep -q "localhost/kcm/controller"; then
        print_status "Removing existing localhost/kcm/controller images..."
        docker rmi $(docker images | grep "localhost/kcm/controller" | awk '{print $3}') 2>/dev/null || true
    fi
    
    # Remove any existing images with our registry name
    if docker images | grep -q "$REGISTRY/controller"; then
        print_status "Removing existing $REGISTRY/controller images..."
        docker rmi $(docker images | grep "$REGISTRY/controller" | awk '{print $3}') 2>/dev/null || true
    fi
    
    print_success "Cleanup completed"
}

# Validate GitHub token
validate_github_token() {
    if [ -z "$GITHUB_TOKEN" ]; then
        return 0  # Skip validation if no token set
    fi
    
    if ! command -v curl &> /dev/null; then
        print_warning "curl not available, skipping token validation"
        return 0
    fi
    
    print_status "Validating GitHub token..."
    
    # Test the token by making a request to GitHub API
    if curl -s -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user > /dev/null 2>&1; then
        print_success "GitHub token is valid"
    else
        print_error "Invalid GitHub token or network error"
        print_status "Please check your GITHUB_TOKEN and try again"
        exit 1
    fi
}

# Login to GHCR
login_to_ghcr() {
    print_status "Logging in to GitHub Container Registry..."
    
    # Check if GITHUB_TOKEN is set
    if [ -z "$GITHUB_TOKEN" ]; then
        print_error "GITHUB_TOKEN environment variable is not set"
        print_status "Please set it with: export GITHUB_TOKEN=your_github_token"
        print_status "You can create a token at: https://github.com/settings/tokens"
        exit 1
    fi
    
    # Check if GITHUB_USERNAME is set
    if [ -z "$GITHUB_USERNAME" ]; then
        print_error "GITHUB_USERNAME environment variable is not set"
        print_status "Please set it with: export GITHUB_USERNAME=your_github_username"
        print_status "Example: export GITHUB_USERNAME=wrkode"
        exit 1
    fi
    
    print_status "Using GitHub username: $GITHUB_USERNAME"
    print_status "Registry: $REGISTRY"
    
    # Login to GHCR
    if echo "$GITHUB_TOKEN" | docker login ghcr.io -u "$GITHUB_USERNAME" --password-stdin; then
        print_success "Successfully logged in to GHCR"
    else
        print_error "Failed to login to GHCR"
        print_status "Please check:"
        print_status "1. Your GITHUB_TOKEN is valid and has the correct permissions"
        print_status "2. Your GitHub username is correct"
        print_status "3. You have write access to the registry"
        print_status ""
        print_status "You can also try setting GITHUB_USERNAME explicitly:"
        print_status "export GITHUB_USERNAME=your_github_username"
        exit 1
    fi
}

# Check registry access
check_registry_access() {
    print_status "Checking registry access..."
    
    # Extract registry name from REGISTRY variable
    REGISTRY_NAME=$(echo "$REGISTRY" | sed 's|ghcr.io/||')
    
    print_status "Registry: $REGISTRY"
    print_status "Registry name: $REGISTRY_NAME"
    
    # Check if we can access the registry
    if docker pull "$REGISTRY/controller:latest" > /dev/null 2>&1; then
        print_success "Registry access confirmed"
    else
        print_warning "Cannot pull from registry (this is normal for new registries)"
        print_status "Will attempt to push to create the registry"
    fi
}

# Check available chart versions
check_available_versions() {
    print_status "Checking available chart versions..."
    
    # Try to pull the chart and get available versions
    if helm pull oci://$REGISTRY/charts/kcm --untar > /dev/null 2>&1; then
        if [ -f "kcm/Chart.yaml" ]; then
            CHART_VERSION_AVAILABLE=$(grep "^version:" kcm/Chart.yaml | awk '{print $2}')
            print_success "Available chart version: $CHART_VERSION_AVAILABLE"
            rm -rf kcm/
        else
            print_warning "Could not determine chart version from Chart.yaml"
        fi
    else
        print_warning "Could not pull chart to check version"
    fi
}

# Build and push controller image
build_controller_image() {
    print_status "Building and pushing controller image..."
    
    # Set environment variables for the build
    export REGISTRY=$REGISTRY
    export VERSION=$VERSION
    export IMG="$REGISTRY/controller:$VERSION"
    
    print_status "Building image: $IMG"
    
    # Build the image
    make docker-build
    
    if [ $? -eq 0 ]; then
        print_success "Controller image built successfully"
    else
        print_error "Failed to build controller image"
        exit 1
    fi
    
    print_status "Pushing image: $IMG"
    
    # Push the image
    make docker-push
    
    if [ $? -eq 0 ]; then
        print_success "Controller image pushed successfully"
    else
        print_error "Failed to push controller image"
        exit 1
    fi
}

# Build and push Helm charts
build_helm_charts() {
    print_status "Building and pushing Helm charts..."
    
    # Set environment variables for the build
    export REGISTRY_REPO="oci://$REGISTRY/charts"
    export VERSION=$CHART_VERSION
    
    # Build and push charts
    make helm-push
    
    if [ $? -eq 0 ]; then
        print_success "Helm charts built and pushed successfully"
    else
        print_error "Failed to build and push Helm charts"
        exit 1
    fi
}

# Build and push CI images (for testing)
build_ci_images() {
    print_status "Building and pushing CI images..."
    
    # Set environment variables for the build
    export REGISTRY=$REGISTRY
    export VERSION=$VERSION
    export REGISTRY_REPO="oci://$REGISTRY/charts-ci"
    export IMG="$REGISTRY/controller-ci:$VERSION"
    
    print_status "Building CI image: $IMG"
    
    # Build and push CI images using the same targets but with different IMG
    make docker-build
    make docker-push
    
    if [ $? -eq 0 ]; then
        print_success "CI images built and pushed successfully"
    else
        print_error "Failed to build and push CI images"
        exit 1
    fi
}

# Generate installation script
generate_installation_script() {
    print_status "Generating installation script..."
    
    # Detect the actual chart version that was pushed
    print_status "Detecting available chart version..."
    CHART_VERSION_AVAILABLE=""
    
    # Try to pull the chart and get its version
    if helm pull oci://$REGISTRY/charts/kcm --untar > /dev/null 2>&1; then
        if [ -f "kcm/Chart.yaml" ]; then
            CHART_VERSION_AVAILABLE=$(grep "^version:" kcm/Chart.yaml | awk '{print $2}')
            rm -rf kcm/
            print_success "Detected chart version: $CHART_VERSION_AVAILABLE"
        fi
    fi
    
    # Use detected version or fall back to CHART_VERSION
    INSTALL_VERSION=${CHART_VERSION_AVAILABLE:-$CHART_VERSION}
    
    cat > install-kcm-custom.sh << EOF
#!/bin/bash

# KCM Installation Script for Custom Registry
# Generated on $(date)
# Registry: $REGISTRY
# Version: $VERSION
# Chart Version: $INSTALL_VERSION

set -e

echo "Installing KCM from custom registry: $REGISTRY"
echo "Version: $VERSION"
echo "Chart Version: $INSTALL_VERSION"

# Create namespace
kubectl create namespace kcm-system --dry-run=client -o yaml | kubectl apply -f -

# Install KCM using Helm
helm install kcm oci://$REGISTRY/charts/kcm --version $INSTALL_VERSION -n kcm-system --create-namespace

echo "KCM installation completed successfully!"
echo "You can check the status with: kubectl get pods -n kcm-system"
EOF
    
    chmod +x install-kcm-custom.sh
    print_success "Installation script generated: install-kcm-custom.sh"
}

# Generate air-gapped installation script
generate_airgap_script() {
    print_status "Generating air-gapped installation script..."
    
    # Use the same version detection logic
    CHART_VERSION_AVAILABLE=""
    
    # Try to pull the chart and get its version
    if helm pull oci://$REGISTRY/charts/kcm --untar > /dev/null 2>&1; then
        if [ -f "kcm/Chart.yaml" ]; then
            CHART_VERSION_AVAILABLE=$(grep "^version:" kcm/Chart.yaml | awk '{print $2}')
            rm -rf kcm/
            print_success "Detected chart version for air-gap: $CHART_VERSION_AVAILABLE"
        fi
    fi
    
    # Use detected version or fall back to CHART_VERSION
    INSTALL_VERSION=${CHART_VERSION_AVAILABLE:-$CHART_VERSION}
    
    cat > airgap-install.sh << EOF
#!/bin/bash

# Air-gapped KCM Installation Script
# Generated on $(date)
# Registry: $REGISTRY
# Version: $VERSION
# Chart Version: $INSTALL_VERSION

set -e

echo "Air-gapped KCM Installation"
echo "Registry: $REGISTRY"
echo "Version: $VERSION"
echo "Chart Version: $INSTALL_VERSION"

# Create namespace
kubectl create namespace kcm-system --dry-run=client -o yaml | kubectl apply -f -

# Download and install KCM
echo "Downloading KCM Helm chart..."
helm pull oci://$REGISTRY/charts/kcm --version $INSTALL_VERSION --untar

echo "Installing KCM..."
helm install kcm ./kcm -n kcm-system --create-namespace

echo "KCM air-gapped installation completed successfully!"
echo "You can check the status with: kubectl get pods -n kcm-system"
EOF
    
    chmod +x airgap-install.sh
    print_success "Air-gapped installation script generated: airgap-install.sh"
}

# Main execution
main() {
    echo "=========================================="
    echo "KCM Custom Registry Build and Push Script"
    echo "=========================================="
    echo "Registry: $REGISTRY"
    echo "Version: $VERSION"
    echo "Chart Version: $CHART_VERSION"
    echo "=========================================="
    echo
    
    check_prerequisites
    validate_github_token
    login_to_ghcr
    cleanup_local_images
    check_available_versions
    build_controller_image
    build_helm_charts
    build_ci_images
    generate_installation_script
    generate_airgap_script
    
    echo
    echo "=========================================="
    print_success "All images and charts have been built and pushed successfully!"
    echo "=========================================="
    echo
    echo "Next steps:"
    echo "1. Run './install-kcm-custom.sh' to install KCM from your registry"
    echo "2. For air-gapped environments, use './airgap-install.sh'"
    echo "3. Update your Kubernetes manifests to use the new registry"
    echo
    echo "Registry URLs:"
    echo "- Controller image: $REGISTRY/controller:$VERSION"
    echo "- Helm charts: oci://$REGISTRY/charts"
    echo "- CI charts: oci://$REGISTRY/charts-ci"
    echo
}

# Run main function
main "$@" 