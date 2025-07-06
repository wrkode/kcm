#!/bin/bash

# KCM Installation Script for wrkode's custom registry (Fixed Versions)
# Generated on $(date)

set -e

echo "=========================================="
echo "KCM Installation Script (Fixed Versions)"
echo "=========================================="
echo "Registry: ghcr.io/wrkode/kcm"
echo "Chart: kcm"
echo "Version: 1.1.2"
echo "=========================================="

# Check if Helm is installed
if ! command -v helm &> /dev/null; then
    echo "[ERROR] Helm is not installed"
    echo "[INFO] Please install Helm first: https://helm.sh/docs/intro/install/"
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo "[ERROR] kubectl is not installed"
    echo "[INFO] Please install kubectl first: https://kubernetes.io/docs/tasks/tools/"
    exit 1
fi

# Check if we have a kubeconfig
if ! kubectl cluster-info &> /dev/null; then
    echo "[ERROR] No valid kubeconfig found"
    echo "[INFO] Please configure kubectl to point to your cluster"
    exit 1
fi

echo "[INFO] Installing KCM from custom registry with version 1.1.2..."

# Install KCM with specific version
helm install kcm oci://ghcr.io/wrkode/kcm/kcm \
    --namespace kcm-system \
    --create-namespace \
    --version 1.1.2 \
    --wait \
    --timeout 10m

if [ $? -eq 0 ]; then
    echo "[SUCCESS] KCM installed successfully!"
    echo "[INFO] You can check the status with: kubectl get pods -n kcm-system"
else
    echo "[ERROR] Failed to install KCM"
    echo "[INFO] You may need to fix CRD conflicts first. Run: ./scripts/fix-crd-conflicts.sh"
    exit 1
fi
