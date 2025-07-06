#!/bin/bash

# Script to build and push KCM images and charts to custom GHCR registry
# Usage: ./build-custom-all.sh

set -e

CUSTOM_USERNAME="wrkode"
CUSTOM_VERSION="1.1.2"
CUSTOM_REGISTRY="ghcr.io/${CUSTOM_USERNAME}"

echo "Building KCM images and charts for custom GHCR registry..."
echo "Registry: $CUSTOM_REGISTRY"
echo "Version: $CUSTOM_VERSION"
echo ""

# Check if GITHUB_TOKEN is set
if [ -z "$GITHUB_TOKEN" ]; then
    echo "❌ Error: GITHUB_TOKEN environment variable is not set"
    echo "Please set it with: export GITHUB_TOKEN=your_github_token"
    exit 1
fi

# Login to GHCR
echo "🔐 Logging in to GHCR..."
echo $GITHUB_TOKEN | docker login ghcr.io -u $CUSTOM_USERNAME --password-stdin

echo ""
echo "🏗️  Step 1: Building and pushing controller images..."
echo "=================================================="

# Set environment variables for goreleaser
export REGISTRY="$CUSTOM_REGISTRY"
export IMAGE_NAME="controller"
export VERSION="$CUSTOM_VERSION"
export SKIP_SCM_RELEASE="true"

# Build and push images
goreleaser release --clean --verbose --skip=validate

echo ""
echo "✅ Controller images built and pushed to $CUSTOM_REGISTRY"
echo ""

echo "📦 Step 2: Building and pushing Helm charts..."
echo "============================================="

# Set environment variables for chart building
export REGISTRY_REPO="oci://$CUSTOM_REGISTRY/kcm/charts"
export VERSION="$CUSTOM_VERSION"
export IMG="$CUSTOM_REGISTRY/controller:$CUSTOM_VERSION"

# Update KCM repository URL in Makefile
echo "📝 Updating KCM repository URL..."
sed -i "s|KCM_REPO_URL ?= oci://ghcr.io/k0rdent/kcm/charts|KCM_REPO_URL ?= $REGISTRY_REPO|g" Makefile

# Set KCM repository for chart building
echo "🔧 Setting KCM repository for chart building..."
make set-kcm-repo

# Generate KCM chart release
echo "📦 Generating KCM chart release..."
make kcm-chart-release

# Push charts to GHCR
echo "🚀 Pushing charts to GHCR..."
make helm-push

echo ""
echo "✅ All components built and pushed successfully!"
echo ""
echo "📋 Summary:"
echo "  - Controller images: $CUSTOM_REGISTRY/controller:$CUSTOM_VERSION"
echo "  - Helm charts: $REGISTRY_REPO"
echo "    - kcm:1.1.2"
echo "    - kcm-templates:1.1.2"
echo ""
echo "🔍 To verify everything is available:"
echo "  # Check images"
echo "  docker pull $CUSTOM_REGISTRY/controller:$CUSTOM_VERSION"
echo ""
echo "  # Check charts"
echo "  helm pull oci://$CUSTOM_REGISTRY/kcm/charts/kcm --version 1.1.2"
echo "  helm pull oci://$CUSTOM_REGISTRY/kcm/charts/kcm-templates --version 1.1.2"
echo ""
echo "🚀 To install KCM with custom components:"
echo "  ./install-kcm-custom-1.1.2.sh" 