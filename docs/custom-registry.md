# Using Your Own GitHub Container Registry for KCM

This guide explains how to build and publish KCM container images to your own GitHub Container Registry (GHCR) instead of using the official k0rdent registry.

## Overview

By default, KCM uses the `ghcr.io/k0rdent/kcm` registry for all container images and Helm charts. This guide helps you:

1. Build all required KCM images
2. Push them to your own GHCR registry
3. Configure KCM to use your custom registry
4. Install KCM from your custom registry

## Prerequisites

Before you begin, ensure you have:

- Docker installed and running
- Git installed
- A GitHub account with a Personal Access Token (PAT)
- The KCM source code cloned locally

### GitHub Personal Access Token

You'll need a GitHub Personal Access Token with the following permissions:

- `write:packages` - To push images to GHCR
- `read:packages` - To read images from GHCR
- `repo` - To access your repository

To create a token:

1. Go to [GitHub Settings > Developer settings > Personal access tokens](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Select the required permissions
4. Copy the token and keep it secure

## Quick Start

### 1. Set up your environment

```bash
# Clone the repository (if you haven't already)
git clone https://github.com/wrkode/kcm.git
cd kcm

# Set your GitHub token
export GITHUB_TOKEN=your_github_token_here
```

### 2. Run the build script

```bash
# Make the script executable
chmod +x scripts/build-and-push-images.sh

# Run the build script
./scripts/build-and-push-images.sh
```

The script will:
- Check prerequisites
- Log in to GHCR
- Build and push the controller image
- Build and push Helm charts
- Build and push CI images
- Generate installation scripts

### 3. Install KCM from your registry

```bash
# Use the generated installation script
./install-kcm-custom.sh
```

## Customization

### Using a different registry

If you want to use a different registry name, you can:

1. **Modify the build script directly:**
   ```bash
   # Edit the script
   vim scripts/build-and-push-images.sh
   
   # Change the REGISTRY variable
   REGISTRY="ghcr.io/your-username/your-repo"
   ```

2. **Use environment variables:**
   ```bash
   export REGISTRY="ghcr.io/your-username/your-repo"
   export VERSION="v1.0.0"
   ./scripts/build-and-push-images.sh
   ```

3. **Use the configuration file:**
   ```bash
   # Copy the config file
   cp scripts/registry-config.env my-config.env
   
   # Edit the configuration
   vim my-config.env
   
   # Source the config and run
   source my-config.env
   ./scripts/build-and-push-images.sh
   ```

### Using a different version

You can specify a custom version:

```bash
export VERSION="v1.0.0"
export CHART_VERSION="v1.0.0"
./scripts/build-and-push-images.sh
```

## Manual Build Process

If you prefer to build manually instead of using the script:

### 1. Login to GHCR

```bash
echo "$GITHUB_TOKEN" | docker login ghcr.io -u your-username --password-stdin
```

### 2. Build and push controller image

```bash
# Set environment variables
export REGISTRY="ghcr.io/your-username/your-repo"
export VERSION="v1.0.0"

# Build the image
make docker-build

# Push the image
make docker-push
```

### 3. Build and push Helm charts

```bash
# Set environment variables
export REGISTRY_REPO="oci://ghcr.io/your-username/your-repo/charts"
export VERSION="v1.0.0"

# Build and push charts
make helm-push
```

### 4. Build and push CI images

```bash
# Set environment variables
export REGISTRY="ghcr.io/your-username/your-repo"
export VERSION="v1.0.0"
export REGISTRY_REPO="oci://ghcr.io/your-username/your-repo/charts-ci"

# Build and push CI images
make docker-build-ci
make docker-push-ci
```

## Installation

### Standard Installation

```bash
# Install from your custom registry
helm install kcm oci://ghcr.io/your-username/your-repo/charts/kcm --version v1.0.0 -n kcm-system --create-namespace
```

### Air-gapped Installation

For environments without internet access:

```bash
# Download the chart
helm pull oci://ghcr.io/your-username/your-repo/charts/kcm --version v1.0.0 --untar

# Install from local chart
helm install kcm ./kcm -n kcm-system --create-namespace
```

## Configuration

### Updating KCM to use your registry

If you want to update an existing KCM installation to use your registry:

```bash
# Update the Helm release
helm upgrade kcm oci://ghcr.io/your-username/your-repo/charts/kcm --version v1.0.0 -n kcm-system
```

### Updating templates repository URL

You can also configure KCM to use your custom templates repository:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kcm-config
  namespace: kcm-system
data:
  templatesRepoURL: "oci://ghcr.io/your-username/your-repo/charts"
```

## Troubleshooting

### Common Issues

1. **Authentication failed**
   ```
   Error: failed to login to ghcr.io
   ```
   
   **Solution:** Ensure your `GITHUB_TOKEN` is set correctly and has the required permissions.

2. **Permission denied**
   ```
   Error: denied: permission_denied
   ```
   
   **Solution:** Check that your GitHub token has `write:packages` permission.

3. **Image not found**
   ```
   Error: manifest for ghcr.io/your-username/your-repo/controller:v1.0.0 not found
   ```
   
   **Solution:** Ensure the image was built and pushed successfully. Check the build logs.

4. **Chart not found**
   ```
   Error: chart "kcm" not found in oci://ghcr.io/your-username/your-repo/charts
   ```
   
   **Solution:** Ensure the Helm charts were built and pushed successfully.

### Debugging

Enable verbose output:

```bash
# For Docker builds
DOCKER_BUILDKIT=1 make docker-build

# For Helm operations
helm install kcm oci://ghcr.io/your-username/your-repo/charts/kcm --version v1.0.0 -n kcm-system --create-namespace --debug
```

### Checking registry contents

```bash
# List images in your registry
docker images ghcr.io/your-username/your-repo/*

# Check Helm chart availability
helm search repo oci://ghcr.io/your-username/your-repo/charts
```

## Security Considerations

1. **Token Security**: Keep your GitHub token secure and don't commit it to version control
2. **Registry Access**: Consider using repository-specific tokens for better security
3. **Image Signing**: Consider signing your images for additional security
4. **Vulnerability Scanning**: Regularly scan your images for vulnerabilities

## Best Practices

1. **Versioning**: Use semantic versioning for your releases
2. **Tagging**: Tag your images with both version and `latest` tags
3. **Documentation**: Document any customizations you make
4. **Testing**: Test your custom builds in a staging environment first
5. **Backup**: Keep backups of your custom images and charts

## Advanced Configuration

### Multi-platform builds

To build for multiple platforms:

```bash
export BUILD_PLATFORMS="linux/amd64,linux/arm64"
export PUSH_PLATFORMS="linux/amd64,linux/arm64"
./scripts/build-and-push-images.sh
```

### Custom Dockerfile

If you need to customize the Docker build:

```bash
# Create a custom Dockerfile
cp Dockerfile Dockerfile.custom

# Edit the custom Dockerfile
vim Dockerfile.custom

# Build with custom Dockerfile
docker build -f Dockerfile.custom -t ghcr.io/your-username/your-repo/controller:v1.0.0 .
```

### CI/CD Integration

You can integrate this into your CI/CD pipeline:

```yaml
# Example GitHub Actions workflow
name: Build and Push KCM Images

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2
      - name: Login to GHCR
        uses: docker/login-action@v2
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - name: Build and push
        run: |
          export REGISTRY="ghcr.io/${{ github.repository_owner }}/kcm"
          export VERSION="${{ github.ref_name }}"
          ./scripts/build-and-push-images.sh
```

## Support

If you encounter issues:

1. Check the [KCM documentation](https://github.com/wrkode/kcm/docs)
2. Review the build logs for error messages
3. Ensure all prerequisites are met
4. Verify your GitHub token has the correct permissions

For additional help, you can:
- Open an issue in the KCM repository
- Check the troubleshooting section above
- Review the KCM source code for build configurations 