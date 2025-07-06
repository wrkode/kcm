# AKS Cluster Adoption Guide

This guide shows you how to adopt an existing AKS cluster into KCM for lifecycle management.

## Prerequisites

1. **Existing AKS Cluster**: You need a running AKS cluster
2. **Azure CLI**: Must be installed and authenticated
3. **kubectl**: Must be configured to access your management cluster
4. **KCM**: Must be installed in your management cluster

## Quick Start (Automated)

Use the provided script to automate the entire process:

```bash
# Make the script executable
chmod +x scripts/setup-aks-adoption.sh

# Run the script
./scripts/setup-aks-adoption.sh <resource-group> <cluster-name> <region>

# Example
./scripts/setup-aks-adoption.sh my-resource-group my-aks-cluster eastus
```

## Manual Step-by-Step Process

### Step 1: Get Your AKS Kubeconfig

```bash
# Get the kubeconfig for your existing AKS cluster
az aks get-credentials --resource-group <your-resource-group> --name <your-cluster-name> --file ~/.kube/aks-cluster-config
```

### Step 2: Create the Kubeconfig Secret

```bash
# Create the secret with your kubeconfig
kubectl create secret generic existing-aks-kubeconfig \
  --from-file=value=~/.kube/aks-cluster-config \
  --namespace=default
```

### Step 3: Create the Credential Resource

```yaml
# Create adopted-aks-credential.yaml
apiVersion: k0rdent.mirantis.com/v1beta1
kind: Credential
metadata:
  name: adopted-aks-credential
  namespace: default
spec:
  description: "Credential for adopted AKS cluster"
  identityRef:
    apiVersion: v1
    kind: Secret
    name: existing-aks-kubeconfig
    namespace: default
```

Apply it:
```bash
kubectl apply -f adopted-aks-credential.yaml
```

### Step 4: Create Azure Credentials (if not already exists)

```yaml
# Create azure-aks-credentials.yaml
apiVersion: v1
kind: Secret
metadata:
  name: azure-aks-credentials
  namespace: default
stringData:
  AZURE_CLIENT_ID: "your-azure-client-id"
  AZURE_CLIENT_SECRET: "your-azure-client-secret"
  AZURE_SUBSCRIPTION_ID: "your-azure-subscription-id"
  AZURE_TENANT_ID: "your-azure-tenant-id"
type: Opaque
---
apiVersion: k0rdent.mirantis.com/v1beta1
kind: Credential
metadata:
  name: azure-aks-credentials
  namespace: default
spec:
  description: "Azure credentials for AKS infrastructure management"
  identityRef:
    apiVersion: v1
    kind: Secret
    name: azure-aks-credentials
    namespace: default
```

Apply it:
```bash
kubectl apply -f azure-aks-credentials.yaml
```

### Step 5: Create the ClusterDeployment

```yaml
# Create adopted-aks-cluster.yaml
apiVersion: k0rdent.mirantis.com/v1beta1
kind: ClusterDeployment
metadata:
  name: adopted-aks-cluster
  namespace: default
spec:
  template: adopted-cluster-1-0-1
  credential: adopted-aks-credential
  propagateCredentials: true
  
  adoptedCluster:
    enabled: true
    discoveryMode: "auto"
    infrastructureMapping:
      provider: "azure"
      region: "eastus"  # Replace with your cluster region
      credentials:
        secretName: "azure-aks-credentials"
        namespace: "default"
    
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
```

Apply it:
```bash
kubectl apply -f adopted-aks-cluster.yaml
```

## Verification

### Check Resource Status

```bash
# Check ClusterDeployment
kubectl get clusterdeployment adopted-aks-cluster -n default

# Check Credential
kubectl get credential adopted-aks-credential -n default

# Check Secret
kubectl get secret existing-aks-kubeconfig -n default
```

### Monitor Adoption Process

```bash
# Get detailed status
kubectl describe clusterdeployment adopted-aks-cluster -n default

# Check adopted cluster status
kubectl get clusterdeployment adopted-aks-cluster -n default -o jsonpath='{.status.adoptedClusterStatus}'

# View KCM controller logs
kubectl logs -n kcm-system deployment/kcm-controller -c manager
```

## Troubleshooting

### Common Issues

1. **Kubeconfig Issues**
   ```bash
   # Verify kubeconfig is valid
   kubectl --kubeconfig ~/.kube/aks-cluster-config get nodes
   
   # Check if cluster is reachable
   kubectl --kubeconfig ~/.kube/aks-cluster-config cluster-info
   ```

2. **Credential Issues**
   ```bash
   # Check credential status
   kubectl describe credential adopted-aks-credential -n default
   
   # Verify secret exists
   kubectl get secret existing-aks-kubeconfig -n default -o yaml
   ```

3. **Discovery Issues**
   ```bash
   # Check discovery status
   kubectl get clusterdeployment adopted-aks-cluster -n default -o jsonpath='{.status.adoptedClusterStatus.discoveryStatus}'
   
   # Verify Azure credentials
   kubectl get secret azure-aks-credentials -n default -o yaml
   ```

### Status Conditions

Monitor these conditions:
- `TemplateReady`: ClusterTemplate validation
- `CredentialReady`: Credential validation
- `HelmReleaseReady`: HelmRelease deployment
- `CAPIClusterSummary`: CAPI cluster health

## Next Steps

After successful adoption:

1. **Monitor Lifecycle Management**: Check scaling, monitoring, and backup operations
2. **Deploy Services**: Use KCM's service deployment capabilities
3. **Configure Alerts**: Set up monitoring alerts for the adopted cluster
4. **Scale Operations**: Apply the same process to other existing clusters

## Examples

See the complete examples in this directory:
- `aks-adopted-cluster-minimal.yaml`: Minimal adoption setup
- `aks-adopted-cluster-example.yaml`: Comprehensive adoption with all features 