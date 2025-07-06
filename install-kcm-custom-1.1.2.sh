#!/bin/bash

# Script to install KCM with custom GHCR images
# Usage: ./install-kcm-custom-1.1.2.sh

set -e

CUSTOM_USERNAME="wrkode"
CUSTOM_VERSION="1.1.2"
CUSTOM_REGISTRY="ghcr.io/${CUSTOM_USERNAME}"

echo "Installing KCM with custom GHCR images..."
echo "Registry: $CUSTOM_REGISTRY"
echo "Version: $CUSTOM_VERSION"
echo ""

# Create namespace
echo "📦 Creating namespace..."
kubectl create namespace kcm-system --dry-run=client -o yaml | kubectl apply -f -

# Install KCM with custom images
echo "🚀 Installing KCM with custom images..."
helm install kcm ./templates/provider/kcm \
  --namespace kcm-system \
  --set image.repository=$CUSTOM_REGISTRY/controller \
  --set image.tag=$CUSTOM_VERSION \
  --set image.pullPolicy=Always \
  --set controller.templatesRepoURL=oci://$CUSTOM_REGISTRY/kcm/charts

echo ""
echo "✅ KCM installed with custom images!"
echo ""
echo "🔍 To verify the installation:"
echo "kubectl get pods -n kcm-system"
echo "kubectl logs -n kcm-system -l app.kubernetes.io/name=kcm"
echo ""
echo "📊 To check the status:"
echo "kubectl get all -n kcm-system" 