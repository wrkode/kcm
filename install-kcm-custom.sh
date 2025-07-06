#!/bin/bash

# KCM Installation Script for Custom Registry
# Generated on Sun Jul  6 09:58:40 AM UTC 2025
# Registry: ghcr.io/wrkode/kcm
# Version: v1.0.0-112-g27cf2cf
# Chart Version: 1.1.2

set -e

echo "Installing KCM from custom registry: ghcr.io/wrkode/kcm"
echo "Version: v1.0.0-112-g27cf2cf"
echo "Chart Version: 1.1.2"

# Create namespace
kubectl create namespace kcm-system --dry-run=client -o yaml | kubectl apply -f -

# Install KCM using Helm
helm install kcm oci://ghcr.io/wrkode/kcm/charts/kcm --version 1.1.2 -n kcm-system --create-namespace

echo "KCM installation completed successfully!"
echo "You can check the status with: kubectl get pods -n kcm-system"
