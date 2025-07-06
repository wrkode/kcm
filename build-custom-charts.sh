#!/bin/bash

# Script to build and push KCM Helm charts to custom GHCR registry
# Usage: ./build-custom-charts.sh

set -e

CUSTOM_USERNAME="wrkode"
CUSTOM_VERSION="1.1.2"
CUSTOM_REGISTRY="ghcr.io/${CUSTOM_USERNAME}"

echo "Building KCM Helm charts for custom GHCR registry..."
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
echo "✅ Charts built and pushed to $REGISTRY_REPO"
echo ""
echo "📋 Pushed charts:"
echo "  - kcm:1.1.2"
echo "  - kcm-templates:1.1.2"
echo ""
echo "🔍 To verify charts are available:"
echo "helm pull oci://$CUSTOM_REGISTRY/kcm/charts/kcm --version 1.1.2"
echo "helm pull oci://$CUSTOM_REGISTRY/kcm/charts/kcm-templates --version 1.1.2" 