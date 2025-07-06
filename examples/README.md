# KCM Examples

This directory contains examples demonstrating how to use KCM (Kubernetes Cluster Manager) to create and manage Kubernetes clusters across different cloud providers.

## AKS (Azure Kubernetes Service) Examples

### Overview

KCM provides a declarative way to create and manage AKS clusters using Kubernetes-native resources. The workflow involves:

1. **Credentials Management**: Azure service principal credentials stored as Kubernetes secrets
2. **Template-Based Deployment**: Pre-defined Helm charts that generate the necessary CAPI (Cluster API) resources
3. **Infrastructure Provisioning**: Azure Service Operator (ASO) creates the actual AKS resources
4. **Lifecycle Management**: KCM monitors and manages the cluster lifecycle

### Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   ClusterDeploy │    │   KCM Controller │    │  Azure Service  │
│   ment          │───▶│                  │───▶│   Operator      │
│                 │    │                  │    │   (ASO)         │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌──────────────────┐
                       │   CAPI Resources │
                       │                  │
                       │ • Cluster        │
                       │ • AzureASOManaged│
                       │   Cluster        │
                       │ • MachinePool   │
                       └──────────────────┘
```

### Key Components

#### 1. ClusterTemplate
- References a Helm chart that contains the AKS cluster templates
- Defines the structure and configuration options for AKS clusters
- Example: `azure-aks-1-0-1` template

#### 2. Credential
- Standardized way to reference cloud provider credentials
- Points to a Kubernetes Secret containing Azure service principal details
- Used by KCM and propagated to the cluster for CCM (Cloud Controller Manager)

#### 3. ClusterDeployment
- The main resource that triggers cluster creation
- References a ClusterTemplate and Credential
- Contains configuration that gets passed to the Helm chart
- KCM creates a HelmRelease that manages the cluster lifecycle

#### 4. Generated Resources
When a ClusterDeployment is applied, KCM generates:

- **Cluster** (`cluster.x-k8s.io/v1beta1`): The main CAPI cluster resource
- **AzureASOManagedCluster** (`infrastructure.cluster.x-k8s.io/v1alpha1`): Azure-specific cluster infrastructure
- **AzureASOManagedControlPlane** (`infrastructure.cluster.x-k8s.io/v1alpha1`): AKS control plane configuration
- **MachinePool** resources for each node pool (system and user)

### Examples

#### 1. Minimal AKS Cluster (`aks-cluster-minimal.yaml`)
Shows the essential components needed to create a basic AKS cluster:
- Azure credentials (Secret + Credential)
- Basic ClusterDeployment with minimal configuration
- System and user node pools

#### 2. Comprehensive AKS Cluster (`aks-cluster-example.yaml`)
Demonstrates advanced features:
- Complete AKS configuration with all options
- Service deployment after cluster creation
- IPAM (IP Address Management) configuration
- Detailed workflow explanation

### Workflow Steps

1. **Validation Phase**
   - KCM validates the ClusterTemplate exists and is valid
   - Validates the Credential is ready and accessible
   - Validates the configuration against the template schema

2. **Helm Release Creation**
   - Creates a HelmRelease that references the AKS chart
   - The chart contains templates for AzureASOManagedCluster, AzureASOManagedControlPlane, and MachinePool resources

3. **Infrastructure Provisioning**
   - Azure Resource Group is created
   - AKS Managed Cluster is provisioned
   - System and User node pools are created
   - Network configuration is applied

4. **CAPI Integration**
   - CAPI Cluster resource is created
   - CAPI MachinePool resources are created for each node pool
   - Infrastructure resources are linked to CAPI resources

5. **Monitoring and Status**
   - KCM monitors the cluster creation process
   - Status conditions are updated (Ready, InfrastructureReady, etc.)
   - Events are generated for important milestones

6. **Service Deployment** (if specified)
   - Services are deployed to the cluster after it's ready
   - Sveltos Profile manages the service lifecycle
   - Health checks ensure services are running correctly

### Configuration Options

#### Basic Configuration
```yaml
config:
  location: "eastus"
  kubernetes:
    version: "v1.31.1"
    networkPlugin: "azure"
    networkPolicy: "azure"
  machinePools:
    system:
      count: 3
      vmSize: "Standard_D2s_v3"
    user:
      count: 2
      vmSize: "Standard_D4s_v3"
```

#### Advanced Configuration
```yaml
config:
  # Network configuration
  clusterNetwork:
    pods:
      cidrBlocks: ["10.244.0.0/16"]
    services:
      cidrBlocks: ["10.96.0.0/12"]
  
  # API Server access
  apiServerAccessProfile:
    authorizedIPRanges: []
    enablePrivateCluster: false
  
  # Auto-upgrade settings
  autoUpgradeProfile:
    nodeOSUpgradeChannel: "None"
    upgradeChannel: "none"
  
  # Security features
  securityProfile:
    azureKeyVaultKms:
      enabled: false
    defender:
      securityMonitoring:
        enabled: false
    workloadIdentity:
      enabled: false
```

### Monitoring and Troubleshooting

#### Check Cluster Status
```bash
# Check ClusterDeployment status
kubectl get clusterdeployment my-aks-cluster -o yaml

# Check generated CAPI resources
kubectl get cluster,azureasomanagedcluster,machinepool

# Check HelmRelease status
kubectl get helmrelease my-aks-cluster
```

#### Common Status Conditions
- `TemplateReady`: ClusterTemplate validation status
- `CredentialReady`: Credential validation status
- `HelmReleaseReady`: HelmRelease deployment status
- `CAPIClusterSummary`: CAPI cluster health status
- `SveltosProfileReady`: Service deployment status

#### Troubleshooting
1. **Template Not Found**: Ensure the ClusterTemplate exists and is valid
2. **Credential Issues**: Verify Azure credentials are correct and have proper permissions
3. **Infrastructure Errors**: Check Azure subscription and resource group permissions
4. **Network Issues**: Verify network configuration and CIDR blocks don't conflict

### Best Practices

1. **Credentials Management**
   - Use dedicated service principals for each environment
   - Rotate credentials regularly
   - Use Azure Key Vault for credential storage in production

2. **Network Planning**
   - Plan CIDR blocks carefully to avoid conflicts
   - Use private clusters for production workloads
   - Configure authorized IP ranges for API server access

3. **Node Pool Design**
   - Use system node pool for system workloads only
   - Design user node pools based on workload requirements
   - Enable autoscaling for cost optimization

4. **Security**
   - Enable Azure Defender for production clusters
   - Use workload identity for secure service-to-service communication
   - Configure RBAC and network policies

5. **Monitoring**
   - Enable Azure Monitor for comprehensive monitoring
   - Set up alerts for cluster health and performance
   - Use KCM's service deployment for monitoring stacks

### Related Resources

- [KCM Documentation](../docs/)
- [Azure AKS Templates](../templates/cluster/azure-aks/)
- [Cluster API Documentation](https://cluster-api.sigs.k8s.io/)
- [Azure Service Operator](https://azure.github.io/azure-service-operator/) 