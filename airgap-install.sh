#!/bin/bash

# Air-gapped KCM Installation Script
# Generated on Sun Jul  6 09:58:41 AM UTC 2025
# Registry: ghcr.io/wrkode/kcm
# Version: v1.0.0-112-g27cf2cf
# Chart Version: 1.1.2

set -e

echo "Air-gapped KCM Installation"
echo "Registry: ghcr.io/wrkode/kcm"
echo "Version: v1.0.0-112-g27cf2cf"
echo "Chart Version: 1.1.2"

# Create namespace
kubectl create namespace kcm-system --dry-run=client -o yaml | kubectl apply -f -

# Download and install KCM
echo "Downloading KCM Helm chart..."
helm pull oci://ghcr.io/wrkode/kcm/charts/kcm --version 1.1.2 --untar

echo "Installing KCM..."
helm install kcm ./kcm -n kcm-system --create-namespace

echo "KCM air-gapped installation completed successfully!"
echo "You can check the status with: kubectl get pods -n kcm-system"
