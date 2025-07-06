# KCM Custom GHCR Registry Setup

This guide explains how to build and use KCM container images from your own GitHub Container Registry (GHCR) instead of the official k0rdent registry.

## Overview

The configuration has been modified to use your GHCR registry (`ghcr.io/wrkode`) for version 1.1.2. This allows you to:

1. Build and push KCM images to your own registry
2. Test your custom builds before reverting to the official registry
3. Have full control over the image lifecycle

## Modified Files

The following files have been updated to use your custom registry:

- `templates/provider/kcm/values.yaml` - Default image repository and templates URL
- `cmd/main.go` - Default templates repository URL
- `Makefile` - Default KCM repository URL

## Prerequisites

1. **GitHub Token**: You need a GitHub Personal Access Token with `write:packages` permission
2. **Docker**: Docker must be installed and running
3. **GoReleaser**: For building multi-architecture images

## Setup Instructions

### 1. Set Your GitHub Token

```bash
export GITHUB_TOKEN=your_github_token_here
```

### 2. Build and Push Images

```bash
./build-custom-images.sh
```

This script will:
- Login to GHCR using your token
- Build multi-architecture images (amd64, arm64, armv7)
- Push images to `ghcr.io/wrkode/kcm/controller:1.1.2`

### 3. Install KCM with Custom Images

```bash
./install-kcm-custom-1.1.2.sh
```

This script will:
- Create the `kcm-system` namespace
- Install KCM using your custom images
- Configure the source-controller to pull from your registry

## Verification

After installation, verify that KCM is running with your custom images:

```bash
# Check pod status
kubectl get pods -n kcm-system

# Check pod logs
kubectl logs -n kcm-system -l app.kubernetes.io/name=kcm

# Check all resources
kubectl get all -n kcm-system
```

## Reverting to Original Registry

When you're done testing, you can revert to the original k0rdent registry:

```bash
./revert-to-k0rdent.sh
```

This will restore all files to use the original `ghcr.io/k0rdent` registry.

## Manual Installation

If you prefer to install manually:

```bash
# Create namespace
kubectl create namespace kcm-system --dry-run=client -o yaml | kubectl apply -f -

# Install KCM with custom images
helm install kcm ./templates/provider/kcm \
  --namespace kcm-system \
  --set image.repository=ghcr.io/wrkode/kcm/controller \
  --set image.tag=1.1.2 \
  --set image.pullPolicy=Always \
  --set controller.templatesRepoURL=oci://ghcr.io/wrkode/kcm/charts
```

## Troubleshooting

### Image Pull Errors

If you encounter image pull errors:

1. **Check registry visibility**: Ensure your GHCR repository is public or you have proper access configured
2. **Verify token permissions**: Make sure your GitHub token has `write:packages` permission
3. **Check image tags**: Verify the images were pushed with the correct tags

### Source-Controller Issues

If the source-controller can't pull images:

1. **Check network connectivity**: Ensure your cluster can reach `ghcr.io`
2. **Verify image existence**: Check that images exist in your registry
3. **Check pull secrets**: If using private registry, ensure proper pull secrets are configured

## Registry Structure

Your custom registry will contain:

```
ghcr.io/wrkode/kcm/
├── controller:1.1.2 (multi-arch manifest)
├── controller:1.1.2-amd64
├── controller:1.1.2-arm64v8
└── controller:1.1.2-armv7
```

## Cleanup

To remove KCM installation:

```bash
helm uninstall kcm -n kcm-system
kubectl delete namespace kcm-system
```

## Notes

- The source-controller will automatically pull images from your registry based on the configured `templatesRepoURL`
- All KCM components (controller, webhooks, etc.) will use your custom images
- The configuration is version-specific (1.1.2) - update the version in scripts if needed
- Make sure your GHCR repository is public or configure proper authentication for your cluster 