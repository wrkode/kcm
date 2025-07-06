#!/bin/bash

# Script to set up AKS cluster adoption
# Usage: ./scripts/setup-aks-adoption.sh <resource-group> <cluster-name> <region>

set -e

if [ $# -ne 3 ]; then
    echo "Usage: $0 <resource-group> <cluster-name> <region>"
    echo "Example: $0 my-resource-group my-aks-cluster eastus"
    exit 1
fi

RESOURCE_GROUP=$1
CLUSTER_NAME=$2
REGION=$3
NAMESPACE=${NAMESPACE:-default}

echo "Setting up AKS cluster adoption..."
echo "Resource Group: $RESOURCE_GROUP"
echo "Cluster Name: $CLUSTER_NAME"
echo "Region: $REGION"
echo "Namespace: $NAMESPACE"
echo ""

# Step 1: Get the kubeconfig
echo "Step 1: Getting AKS cluster kubeconfig..."
TEMP_KUBECONFIG=$(mktemp)
az aks get-credentials --resource-group "$RESOURCE_GROUP" --name "$CLUSTER_NAME" --file "$TEMP_KUBECONFIG"

if [ ! -f "$TEMP_KUBECONFIG" ]; then
    echo "❌ Error: Failed to get kubeconfig"
    exit 1
fi

echo "✅ Kubeconfig retrieved successfully"
echo ""

# Step 2: Create the kubeconfig secret
echo "Step 2: Creating kubeconfig secret..."
kubectl create secret generic existing-aks-kubeconfig \
  --from-file=value="$TEMP_KUBECONFIG" \
  --namespace="$NAMESPACE" \
  --dry-run=client -o yaml > kubeconfig-secret.yaml

echo "✅ Kubeconfig secret YAML created: kubeconfig-secret.yaml"
echo ""

# Step 3: Create the credential
echo "Step 3: Creating credential resource..."
cat > adopted-aks-credential.yaml <<EOF
apiVersion: k0rdent.mirantis.com/v1beta1
kind: Credential
metadata:
  name: adopted-aks-credential
  namespace: $NAMESPACE
spec:
  description: "Credential for adopted AKS cluster $CLUSTER_NAME"
  identityRef:
    apiVersion: v1
    kind: Secret
    name: existing-aks-kubeconfig
    namespace: $NAMESPACE
EOF

echo "✅ Credential YAML created: adopted-aks-credential.yaml"
echo ""

# Step 4: Create the ClusterDeployment
echo "Step 4: Creating ClusterDeployment..."
cat > adopted-aks-cluster.yaml <<EOF
apiVersion: k0rdent.mirantis.com/v1beta1
kind: ClusterDeployment
metadata:
  name: adopted-aks-$CLUSTER_NAME
  namespace: $NAMESPACE
  labels:
    environment: production
    managed-by: kcm
    adopted-cluster: "true"
spec:
  template: adopted-cluster-1-0-1
  credential: adopted-aks-credential
  propagateCredentials: true
  
  adoptedCluster:
    enabled: true
    discoveryMode: "auto"
    infrastructureMapping:
      provider: "azure"
      region: "$REGION"
      credentials:
        secretName: "azure-aks-credentials"
        namespace: "$NAMESPACE"
    
    lifecycleManagement:
      enabled: true
      scaling:
        enabled: true
        autoScaling: true
        minNodes: 2
        maxNodes: 10
      monitoring:
        enabled: true
        metrics:
          - "cpu_usage_percent"
          - "memory_usage_percent"
      backup:
        enabled: true
        schedule: "0 1 * * *"
EOF

echo "✅ ClusterDeployment YAML created: adopted-aks-cluster.yaml"
echo ""

# Step 5: Apply the resources
echo "Step 5: Applying resources..."
echo "Applying kubeconfig secret..."
kubectl apply -f kubeconfig-secret.yaml

echo "Applying credential..."
kubectl apply -f adopted-aks-credential.yaml

echo "Applying ClusterDeployment..."
kubectl apply -f adopted-aks-cluster.yaml

echo ""
echo "✅ All resources applied successfully!"
echo ""

# Step 6: Show status
echo "Step 6: Checking status..."
echo "ClusterDeployment status:"
kubectl get clusterdeployment "adopted-aks-$CLUSTER_NAME" -n "$NAMESPACE"

echo ""
echo "Credential status:"
kubectl get credential "adopted-aks-credential" -n "$NAMESPACE"

echo ""
echo "Secret status:"
kubectl get secret "existing-aks-kubeconfig" -n "$NAMESPACE"

echo ""
echo "🎉 AKS cluster adoption setup complete!"
echo ""
echo "Next steps:"
echo "1. Monitor the adoption process:"
echo "   kubectl describe clusterdeployment adopted-aks-$CLUSTER_NAME -n $NAMESPACE"
echo ""
echo "2. Check adopted cluster status:"
echo "   kubectl get clusterdeployment adopted-aks-$CLUSTER_NAME -n $NAMESPACE -o jsonpath='{.status.adoptedClusterStatus}'"
echo ""
echo "3. View logs:"
echo "   kubectl logs -n kcm-system deployment/kcm-controller -c manager"
echo ""

# Cleanup temp file
rm -f "$TEMP_KUBECONFIG" 