#!/bin/bash

# Script to modify KCM configuration for custom GHCR registry
# Usage: ./scripts/setup-custom-ghcr.sh <your-github-username>

set -e

if [ $# -eq 0 ]; then
    echo "Usage: $0 <your-github-username>"
    echo "Example: $0 myusername"
    exit 1
fi

CUSTOM_USERNAME=$1
CUSTOM_VERSION="1.1.2"
ORIGINAL_REGISTRY="ghcr.io/k0rdent"
CUSTOM_REGISTRY="ghcr.io/${CUSTOM_USERNAME}"

echo "Setting up KCM for custom GHCR registry..."
echo "Username: $CUSTOM_USERNAME"
echo "Version: $CUSTOM_VERSION"
echo "Original Registry: $ORIGINAL_REGISTRY"
echo "Custom Registry: $CUSTOM_REGISTRY"
echo ""

# Backup original files
echo "Creating backups..."
cp .goreleaser.yaml .goreleaser.yaml.backup
cp templates/provider/kcm/values.yaml templates/provider/kcm/values.yaml.backup
cp .github/workflows/post-merge-push.yaml .github/workflows/post-merge-push.yaml.backup
cp .github/workflows/build_test.yml .github/workflows/build_test.yml.backup
cp Makefile Makefile.backup

# Update .goreleaser.yaml
echo "Updating .goreleaser.yaml..."
sed -i "s|ghcr.io/k0rdent/kcm|${CUSTOM_REGISTRY}/kcm|g" .goreleaser.yaml

# Update values.yaml
echo "Updating templates/provider/kcm/values.yaml..."
sed -i "s|ghcr.io/k0rdent/kcm/controller|${CUSTOM_REGISTRY}/kcm/controller|g" templates/provider/kcm/values.yaml
sed -i "s|oci://ghcr.io/k0rdent/kcm/charts|oci://${CUSTOM_REGISTRY}/kcm/charts|g" templates/provider/kcm/values.yaml

# Update GitHub workflows
echo "Updating GitHub workflows..."
sed -i "s|ghcr.io/k0rdent/kcm|${CUSTOM_REGISTRY}/kcm|g" .github/workflows/post-merge-push.yaml
sed -i "s|ghcr.io/k0rdent/kcm|${CUSTOM_REGISTRY}/kcm|g" .github/workflows/build_test.yml

# Update Makefile
echo "Updating Makefile..."
sed -i "s|ghcr.io/k0rdent/kcm|${CUSTOM_REGISTRY}/kcm|g" Makefile

# Create a script to revert changes
cat > scripts/revert-custom-ghcr.sh << 'EOF'
#!/bin/bash
echo "Reverting to original k0rdent registry..."

# Restore original files
cp .goreleaser.yaml.backup .goreleaser.yaml
cp templates/provider/kcm/values.yaml.backup templates/provider/kcm/values.yaml
cp .github/workflows/post-merge-push.yaml.backup .github/workflows/post-merge-push.yaml
cp .github/workflows/build_test.yml.backup .github/workflows/build_test.yml
cp Makefile.backup Makefile

echo "Reverted to original configuration."
echo "You can now build and push images to the original k0rdent registry."
EOF

chmod +x scripts/revert-custom-ghcr.sh

# Create build script for custom registry
cat > scripts/build-custom-images.sh << EOF
#!/bin/bash

# Script to build and push KCM images to custom GHCR registry
# Usage: ./scripts/build-custom-images.sh

set -e

CUSTOM_USERNAME="${CUSTOM_USERNAME}"
CUSTOM_VERSION="${CUSTOM_VERSION}"
CUSTOM_REGISTRY="${CUSTOM_REGISTRY}"

echo "Building KCM images for custom GHCR registry..."
echo "Registry: \$CUSTOM_REGISTRY"
echo "Version: \$CUSTOM_VERSION"
echo ""

# Login to GHCR
echo "Logging in to GHCR..."
echo \$GITHUB_TOKEN | docker login ghcr.io -u \$CUSTOM_USERNAME --password-stdin

# Set environment variables for goreleaser
export REGISTRY="\$CUSTOM_REGISTRY"
export IMAGE_NAME="controller"
export VERSION="\$CUSTOM_VERSION"
export SKIP_SCM_RELEASE="true"

# Build and push images
echo "Building and pushing images..."
goreleaser release --clean --verbose --skip=validate

echo "Images built and pushed to \$CUSTOM_REGISTRY"
echo ""
echo "To install KCM with custom images, use:"
echo "helm install kcm ./templates/provider/kcm --set image.repository=\$CUSTOM_REGISTRY/controller --set image.tag=\$CUSTOM_VERSION"
EOF

chmod +x scripts/build-custom-images.sh

# Create installation script
cat > install-kcm-custom-1.1.2.sh << EOF
#!/bin/bash

# Script to install KCM with custom GHCR images
# Usage: ./install-kcm-custom-1.1.2.sh

set -e

CUSTOM_USERNAME="${CUSTOM_USERNAME}"
CUSTOM_VERSION="${CUSTOM_VERSION}"
CUSTOM_REGISTRY="${CUSTOM_REGISTRY}"

echo "Installing KCM with custom GHCR images..."
echo "Registry: \$CUSTOM_REGISTRY"
echo "Version: \$CUSTOM_VERSION"
echo ""

# Create namespace
kubectl create namespace kcm-system --dry-run=client -o yaml | kubectl apply -f -

# Install KCM with custom images
helm install kcm ./templates/provider/kcm \\
  --namespace kcm-system \\
  --set image.repository=\$CUSTOM_REGISTRY/controller \\
  --set image.tag=\$CUSTOM_VERSION \\
  --set image.pullPolicy=Always \\
  --set controller.templatesRepoURL=oci://\$CUSTOM_REGISTRY/kcm/charts

echo "KCM installed with custom images!"
echo ""
echo "To verify the installation:"
echo "kubectl get pods -n kcm-system"
echo "kubectl logs -n kcm-system -l app.kubernetes.io/name=kcm"
EOF

chmod +x install-kcm-custom-1.1.2.sh

echo ""
echo "✅ KCM configuration updated for custom GHCR registry!"
echo ""
echo "📋 Summary of changes:"
echo "  - Updated registry from $ORIGINAL_REGISTRY to $CUSTOM_REGISTRY"
echo "  - Version set to $CUSTOM_VERSION"
echo "  - Created backup files with .backup extension"
echo ""
echo "🔧 Available scripts:"
echo "  - scripts/build-custom-images.sh: Build and push images to your GHCR"
echo "  - scripts/revert-custom-ghcr.sh: Revert to original k0rdent registry"
echo "  - install-kcm-custom-1.1.2.sh: Install KCM with custom images"
echo ""
echo "🚀 Next steps:"
echo "  1. Set your GitHub token: export GITHUB_TOKEN=your_token_here"
echo "  2. Build and push images: ./scripts/build-custom-images.sh"
echo "  3. Install KCM: ./install-kcm-custom-1.1.2.sh"
echo ""
echo "⚠️  Note: Make sure your GHCR repository is public or you have proper access configured." 