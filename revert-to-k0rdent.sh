#!/bin/bash

# Script to revert KCM configuration back to k0rdent registry
# Usage: ./revert-to-k0rdent.sh

set -e

echo "Reverting KCM configuration to k0rdent registry..."

# Revert values.yaml
echo "📝 Reverting templates/provider/kcm/values.yaml..."
sed -i 's|ghcr.io/wrkode/kcm/controller|ghcr.io/k0rdent/kcm/controller|g' templates/provider/kcm/values.yaml
sed -i 's|oci://ghcr.io/wrkode/kcm/charts|oci://ghcr.io/k0rdent/kcm/charts|g' templates/provider/kcm/values.yaml
sed -i 's|tag: 1.1.2|tag: latest|g' templates/provider/kcm/values.yaml

# Revert main.go
echo "📝 Reverting cmd/main.go..."
sed -i 's|oci://ghcr.io/wrkode/kcm/charts|oci://ghcr.io/k0rdent/kcm/charts|g' cmd/main.go

# Revert Makefile
echo "📝 Reverting Makefile..."
sed -i 's|oci://ghcr.io/wrkode/kcm/charts|oci://ghcr.io/k0rdent/kcm/charts|g' Makefile

echo ""
echo "✅ Reverted to original k0rdent registry configuration!"
echo ""
echo "📦 To install KCM with original images, use:"
echo "helm install kcm ./templates/provider/kcm \\"
echo "  --namespace kcm-system \\"
echo "  --set image.repository=ghcr.io/k0rdent/kcm/controller \\"
echo "  --set image.tag=latest \\"
echo "  --set controller.templatesRepoURL=oci://ghcr.io/k0rdent/kcm/charts" 