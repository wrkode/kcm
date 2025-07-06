#!/bin/bash

# Script to build and push KCM images to custom GHCR registry
# Usage: ./build-custom-images.sh

set -e

CUSTOM_USERNAME="wrkode"
CUSTOM_VERSION="1.1.2"
CUSTOM_REGISTRY="ghcr.io/${CUSTOM_USERNAME}"

echo "Building KCM images for custom GHCR registry..."
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

# Set environment variables for goreleaser
export REGISTRY="$CUSTOM_REGISTRY"
export IMAGE_NAME="controller"
export VERSION="$CUSTOM_VERSION"
export SKIP_SCM_RELEASE="true"

# Build and push images
echo "🏗️  Building and pushing images..."
goreleaser release --clean --verbose --skip=validate

echo ""
echo "✅ Images built and pushed to $CUSTOM_REGISTRY"
echo ""
echo "📦 To install KCM with custom images, use:"
echo "helm install kcm ./templates/provider/kcm \\"
echo "  --namespace kcm-system \\"
echo "  --set image.repository=$CUSTOM_REGISTRY/controller \\"
echo "  --set image.tag=$CUSTOM_VERSION \\"
echo "  --set image.pullPolicy=Always \\"
echo "  --set controller.templatesRepoURL=oci://$CUSTOM_REGISTRY/kcm/charts" 