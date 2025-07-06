#!/bin/bash

# Script to fix GHCR package visibility and make packages public
# This script helps make your KCM packages publicly accessible

set -e

# Configuration
REGISTRY="ghcr.io/wrkode/kcm"
PACKAGES=("charts" "controller" "controller-ci")

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
    
    # Check if GITHUB_TOKEN is set
    if [ -z "$GITHUB_TOKEN" ]; then
        print_error "GITHUB_TOKEN environment variable is not set"
        print_status "Please set it with: export GITHUB_TOKEN=your_github_token"
        exit 1
    fi
    
    # Check if GITHUB_USERNAME is set
    if [ -z "$GITHUB_USERNAME" ]; then
        print_error "GITHUB_USERNAME environment variable is not set"
        print_status "Please set it with: export GITHUB_USERNAME=your_github_username"
        exit 1
    fi
    
    # Check if curl is available
    if ! command -v curl &> /dev/null; then
        print_error "curl is not installed or not in PATH"
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Make package public
make_package_public() {
    local package_name=$1
    print_status "Making package $package_name public..."
    
    # Get the package ID
    local package_id=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
        "https://api.github.com/user/packages/container/$package_name" | \
        jq -r '.id // empty')
    
    if [ -z "$package_id" ] || [ "$package_id" = "null" ]; then
        print_warning "Package $package_name not found or not accessible"
        return 1
    fi
    
    print_status "Package ID: $package_id"
    
    # Make the package public
    local response=$(curl -s -X PATCH \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Accept: application/vnd.github.v3+json" \
        -d '{"visibility":"public"}' \
        "https://api.github.com/user/packages/container/$package_name")
    
    if echo "$response" | jq -e '.message' > /dev/null 2>&1; then
        print_error "Failed to make package $package_name public: $(echo "$response" | jq -r '.message')"
        return 1
    else
        print_success "Package $package_name is now public"
        return 0
    fi
}

# List all packages
list_packages() {
    print_status "Listing all packages..."
    
    local response=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
        "https://api.github.com/user/packages")
    
    if echo "$response" | jq -e '.message' > /dev/null 2>&1; then
        print_error "Failed to list packages: $(echo "$response" | jq -r '.message')"
        return 1
    fi
    
    echo "$response" | jq -r '.[] | select(.package_type == "container") | "\(.name) (\(.visibility))"'
}

# Main execution
main() {
    echo "=========================================="
    echo "GHCR Package Visibility Fix Script"
    echo "=========================================="
    echo "Registry: $REGISTRY"
    echo "Username: $GITHUB_USERNAME"
    echo "=========================================="
    echo
    
    check_prerequisites
    
    print_status "Current packages:"
    list_packages
    
    echo
    print_status "Making packages public..."
    
    success_count=0
    total_count=0
    
    for package in "${PACKAGES[@]}"; do
        total_count=$((total_count + 1))
        if make_package_public "$package"; then
            success_count=$((success_count + 1))
        fi
    done
    
    echo
    echo "=========================================="
    if [ $success_count -eq $total_count ]; then
        print_success "All packages made public successfully!"
    else
        print_warning "Made $success_count out of $total_count packages public"
    fi
    echo "=========================================="
    echo
    echo "Next steps:"
    echo "1. Try installing KCM again: ./install-kcm-custom.sh"
    echo "2. If you still get CRD errors, run: ./scripts/fix-crd-conflicts.sh"
    echo
}

# Run main function
main "$@" 